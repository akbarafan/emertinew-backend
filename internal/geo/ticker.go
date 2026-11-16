package geo

import (
	"fmt"
	"log"
	"math"
	"time"

	"github.com/Alugara-Inovasi-Utama/emerti/model"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// GeofenceTicker periodically checks vessel positions against company geofence zones
type GeofenceTicker struct {
	db       *gorm.DB
	repo     Repository
	stopChan chan struct{}
}

func NewGeofenceTicker(db *gorm.DB, repo Repository) *GeofenceTicker {
	return &GeofenceTicker{
		db:       db,
		repo:     repo,
		stopChan: make(chan struct{}),
	}
}

func (t *GeofenceTicker) Start() {
	go func() {
		t.process()

		ticker := time.NewTicker(1 * time.Minute)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				t.process()
			case <-t.stopChan:
				log.Println("[GEO-TICKER] Stopped")
				return
			}
		}
	}()
	log.Println("[GEO-TICKER] Started (runs every 1 minute for all fleet vessels)")
}

func (t *GeofenceTicker) Stop() {
	close(t.stopChan)
}

func (t *GeofenceTicker) process() {
	// 1. Fetch all active vessels
	var vessels []model.MVessel
	if err := t.db.Find(&vessels).Error; err != nil {
		log.Printf("[GEO-TICKER] Error fetching vessels: %v", err)
		return
	}

	if len(vessels) == 0 {
		return
	}

	// 2. Fetch all active geofences
	var geofences []model.MGeofence
	if err := t.db.Where("is_active = ?", true).Find(&geofences).Error; err != nil {
		log.Printf("[GEO-TICKER] Error fetching geofences: %v", err)
		return
	}

	if len(geofences) == 0 {
		return
	}

	for _, v := range vessels {
		t.processVessel(&v, geofences)
	}
}

func (t *GeofenceTicker) processVessel(vessel *model.MVessel, allGeofences []model.MGeofence) {
	// Get latest tracking position (silenced: "record not found" is normal for vessels without telemetry yet)
	var tracking model.Tracking
	err := t.db.Session(&gorm.Session{Logger: logger.Discard}).
		Where("vessel_id = ?", vessel.ID).
		Order("time DESC").First(&tracking).Error
	if err != nil || tracking.Lat == nil || tracking.Lon == nil {
		return
	}

	lat := *tracking.Lat
	lon := *tracking.Lon
	var speed float64
	if tracking.Speed != nil {
		speed = *tracking.Speed
	}

	// Filter geofences matching vessel company or public (company_id is null)
	for _, g := range allGeofences {
		if g.CompanyID != nil && vessel.CompanyID != nil && *g.CompanyID != *vessel.CompanyID {
			continue // Different company, skip
		}
		if len(g.Points) < 2 {
			continue
		}

		t.evaluateZone(vessel, &g, lat, lon, speed)
	}
}

func (t *GeofenceTicker) evaluateZone(vessel *model.MVessel, geofence *model.MGeofence, lat, lon, speed float64) {
	radiusMeters := geofence.RadiusMiles * 1609.34
	distToLine := calculateMinDistance(lat, lon, geofence.Points)

	// Check if coordinate is physically inside polygon zone interior
	isInsidePoly := len(geofence.Points) >= 3 && pointInPolygon(lat, lon, geofence.Points)

	var isOutside bool
	var distMiles float64

	if isInsidePoly {
		isOutside = false
		distMiles = 0
	} else if distToLine <= radiusMeters {
		// Outside core polygon, but within buffer radius
		isOutside = false
		distMiles = 0
	} else {
		// Strictly outside polygon and buffer
		isOutside = true
		distMiles = (distToLine - radiusMeters) / 1609.34
	}

	// Check last log for this vessel and geofence (silenced: "record not found" is normal for initial checks)
	var lastLog model.GeofenceLog
	hasLastLog := t.db.Session(&gorm.Session{Logger: logger.Discard}).
		Where("vessel_id = ? AND geofence_id = ?", vessel.ID, geofence.ID).
		Order("created_at DESC").First(&lastLog).Error == nil

	// Check state transition or periodic alert
	zoneType := geofence.ZoneType
	if zoneType == "" {
		zoneType = "work_zone"
	}

	now := time.Now()

	// 1. Work Zone (Keluar area kerja)
	if zoneType == "work_zone" && geofence.AlertOnExit {
		if isOutside {
			// Trigger exit log if state changed or if outside and no log in last 30 minutes
			if !hasLastLog || !lastLog.IsOutsideGeofence || now.Sub(lastLog.CreatedAt) > 30*time.Minute {
				msg := fmt.Sprintf("Kapal %s berada di luar zona kerja %s (jarak: %.2f nm)", vessel.Name, geofence.Name, distMiles)
				t.createLog(vessel.ID, geofence.ID, "exited", lat, lon, speed, true, distMiles, msg)
			}
		} else if hasLastLog && lastLog.IsOutsideGeofence {
			// Kembali masuk ke zona kerja
			msg := fmt.Sprintf("Kapal %s telah kembali masuk ke dalam zona kerja %s", vessel.Name, geofence.Name)
			t.createLog(vessel.ID, geofence.ID, "entered", lat, lon, speed, false, 0, msg)
		}
	}

	// 2. Restricted Zone (Masuk area bahaya / terlarang)
	if zoneType == "restricted" && geofence.AlertOnEntry {
		if !isOutside {
			// Trigger entered violation log if newly entered or no log in last 30 minutes
			if !hasLastLog || lastLog.IsOutsideGeofence || now.Sub(lastLog.CreatedAt) > 30*time.Minute {
				msg := fmt.Sprintf("PERINGATAN: Kapal %s memasuki zona terlarang/safety zone %s!", vessel.Name, geofence.Name)
				t.createLog(vessel.ID, geofence.ID, "entered", lat, lon, speed, false, 0, msg)
			}
		} else if hasLastLog && !lastLog.IsOutsideGeofence {
			msg := fmt.Sprintf("Kapal %s telah keluar dari zona terlarang %s", vessel.Name, geofence.Name)
			t.createLog(vessel.ID, geofence.ID, "exited", lat, lon, speed, true, distMiles, msg)
		}
	}

	// 3. Over Speed Violation in Zone
	if geofence.MaxSpeedKnots != nil && *geofence.MaxSpeedKnots > 0 && !isOutside && speed > *geofence.MaxSpeedKnots {
		if !hasLastLog || lastLog.EventType != "speed_violation" || now.Sub(lastLog.CreatedAt) > 15*time.Minute {
			msg := fmt.Sprintf("PELANGGARAN KECEPATAN: Kapal %s melaju %.1f kn (batas zona %s: %.1f kn)", vessel.Name, speed, geofence.Name, *geofence.MaxSpeedKnots)
			t.createLog(vessel.ID, geofence.ID, "speed_violation", lat, lon, speed, isOutside, distMiles, msg)
		}
	}
}

func (t *GeofenceTicker) createLog(vesselID, geofenceID uint, eventType string, lat, lon, speed float64, isOutside bool, distance float64, message string) {
	entry := &model.GeofenceLog{
		VesselID:          vesselID,
		GeofenceID:        geofenceID,
		EventType:         eventType,
		Latitude:          lat,
		Longitude:         lon,
		Speed:             speed,
		IsOutsideGeofence: isOutside,
		Distance:          distance,
		Message:           message,
	}
	if err := t.db.Create(entry).Error; err != nil {
		log.Printf("[GEO-TICKER] Error creating geofence log: %v", err)
	}
}

// pointInPolygon determines if coordinate (lat, lon) is inside a polygon using ray casting algorithm.
func pointInPolygon(lat, lon float64, points model.Points) bool {
	if len(points) < 3 {
		return false
	}
	inside := false
	j := len(points) - 1
	for i := 0; i < len(points); i++ {
		xi, yi := points[i].Latitude, points[i].Longitude
		xj, yj := points[j].Latitude, points[j].Longitude

		intersect := ((yi > lon) != (yj > lon)) &&
			(lat < (xj-xi)*(lon-yi)/(yj-yi)+xi)
		if intersect {
			inside = !inside
		}
		j = i
	}
	return inside
}

// ── Haversine Distance Helpers ─────────────────────────────────────

func haversine(lat1, lon1, lat2, lon2 float64) float64 {
	const R = 6371000
	dLat := (lat2 - lat1) * math.Pi / 180
	dLon := (lon2 - lon1) * math.Pi / 180
	a := math.Sin(dLat/2)*math.Sin(dLat/2) +
		math.Cos(lat1*math.Pi/180)*math.Cos(lat2*math.Pi/180)*
			math.Sin(dLon/2)*math.Sin(dLon/2)
	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))
	return R * c
}

func calculateMinDistance(lat, lon float64, points model.Points) float64 {
	if len(points) < 2 {
		return 0
	}

	minDist := math.MaxFloat64

	for i := 0; i < len(points)-1; i++ {
		dist := distanceToLineSegment(
			lat, lon,
			points[i].Latitude, points[i].Longitude,
			points[i+1].Latitude, points[i+1].Longitude,
		)
		if dist < minDist {
			minDist = dist
		}
	}

	// For polygons (3+ points), also check the closing segment
	if len(points) >= 3 {
		lastIdx := len(points) - 1
		if points[0].Latitude != points[lastIdx].Latitude || points[0].Longitude != points[lastIdx].Longitude {
			closingDist := distanceToLineSegment(
				lat, lon,
				points[lastIdx].Latitude, points[lastIdx].Longitude,
				points[0].Latitude, points[0].Longitude,
			)
			if closingDist < minDist {
				minDist = closingDist
			}
		}
	}

	// Distance to start and end points
	startDist := haversine(lat, lon, points[0].Latitude, points[0].Longitude)
	if startDist < minDist {
		minDist = startDist
	}

	endDist := haversine(lat, lon, points[len(points)-1].Latitude, points[len(points)-1].Longitude)
	if endDist < minDist {
		minDist = endDist
	}

	return minDist
}

func distanceToLineSegment(px, py, x1, y1, x2, y2 float64) float64 {
	dx := x2 - x1
	dy := y2 - y1

	if dx == 0 && dy == 0 {
		return haversine(px, py, x1, y1)
	}

	t := ((px-x1)*dx + (py-y1)*dy) / (dx*dx + dy*dy)

	if t < 0 {
		t = 0
	} else if t > 1 {
		t = 1
	}

	closestX := x1 + t*dx
	closestY := y1 + t*dy

	return haversine(px, py, closestX, closestY)
}
