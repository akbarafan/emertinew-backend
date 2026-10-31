package geofencelog

import (
	"errors"
	"math"

	"github.com/Alugara-Inovasi-Utama/emerti/model"
	"gorm.io/gorm"
)

var (
	ErrGeofenceNotFound = errors.New("geofence tidak ditemukan")
	ErrGeofenceInactive = errors.New("geofence tidak aktif")
)

type Service interface {
	CheckGeofence(req CheckGeofenceRequest) (*CheckGeofenceResponse, error)
	CreateGeofenceLog(req CreateGeofenceLogRequest) (*GeofenceLogResponse, error)
	GetLogsByVesselID(vesselID uint, page, limit int) ([]GeofenceLogResponse, int64, error)
	GetLogsByGeofenceID(geofenceID uint, vesselID *uint, isOutside *bool, page, limit int) ([]GeofenceLogResponse, int64, error)
	ListLogs(companyID *uint, vesselID *uint, geofenceID *uint, eventType string, page, limit int) ([]GeofenceLogResponse, int64, error)
}

type service struct {
	logRepo      Repository
	geofenceRepo GeofenceRepository
}

func NewService(logRepo Repository, geofenceRepo GeofenceRepository) Service {
	return &service{
		logRepo:      logRepo,
		geofenceRepo: geofenceRepo,
	}
}

func (s *service) CheckGeofence(req CheckGeofenceRequest) (*CheckGeofenceResponse, error) {
	// Get geofence
	geofence, err := s.geofenceRepo.GetByID(req.GeofenceID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrGeofenceNotFound
		}
		return nil, err
	}

	// Check if geofence is active
	if !geofence.IsActive {
		return nil, ErrGeofenceInactive
	}

	// Calculate distance to geofence line
	distanceToLine := s.calculateDistanceToGeofence(req.Latitude, req.Longitude, geofence)

	// Convert radius from miles to meters
	radiusMeters := geofence.RadiusMiles * 1609.34

	// Check if coordinate is physically inside polygon zone interior
	isInsidePoly := len(geofence.Points) >= 3 && pointInPolygon(req.Latitude, req.Longitude, geofence.Points)

	// Check if outside geofence (both polygon interior and buffer)
	var isOutside bool
	var distance float64

	if isInsidePoly {
		isOutside = false
		distance = 0
	} else if distanceToLine <= radiusMeters {
		isOutside = false
		distance = 0
	} else {
		isOutside = true
		distance = (distanceToLine - radiusMeters) / 1609.34 // Convert to miles
	}

	return &CheckGeofenceResponse{
		VesselID:          req.VesselID,
		GeofenceID:        geofence.ID,
		GeofenceName:      geofence.Name,
		Latitude:          req.Latitude,
		Longitude:         req.Longitude,
		IsOutsideGeofence: isOutside,
		Distance:          distance,
	}, nil
}

func (s *service) CreateGeofenceLog(req CreateGeofenceLogRequest) (*GeofenceLogResponse, error) {
	// Get geofence for validation and name
	geofence, err := s.geofenceRepo.GetByID(req.GeofenceID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrGeofenceNotFound
		}
		return nil, err
	}

	// Check if geofence is active
	if !geofence.IsActive {
		return nil, ErrGeofenceInactive
	}

	// Create log using data from frontend
	log := &model.GeofenceLog{
		VesselID:          req.VesselID,
		GeofenceID:        req.GeofenceID,
		Latitude:          req.Latitude,
		Longitude:         req.Longitude,
		IsOutsideGeofence: req.IsOutsideGeofence,
		Distance:          req.Distance,
	}

	if err := s.logRepo.Create(log); err != nil {
		return nil, err
	}

	return &GeofenceLogResponse{
		ID:                log.ID,
		VesselID:          log.VesselID,
		GeofenceID:        log.GeofenceID,
		GeofenceName:      geofence.Name,
		Latitude:          log.Latitude,
		Longitude:         log.Longitude,
		IsOutsideGeofence: log.IsOutsideGeofence,
		Distance:          log.Distance,
		CreatedAt:         log.CreatedAt,
	}, nil
}

func (s *service) GetLogsByVesselID(vesselID uint, page, limit int) ([]GeofenceLogResponse, int64, error) {
	logs, total, err := s.logRepo.GetByVesselID(vesselID, page, limit)
	if err != nil {
		return nil, 0, err
	}

	responses := make([]GeofenceLogResponse, len(logs))
	for i, log := range logs {
		// Get geofence name
		geofence, _ := s.geofenceRepo.GetByID(log.GeofenceID)
		geofenceName := ""
		if geofence != nil {
			geofenceName = geofence.Name
		}

		responses[i] = GeofenceLogResponse{
			ID:                log.ID,
			VesselID:          log.VesselID,
			GeofenceID:        log.GeofenceID,
			GeofenceName:      geofenceName,
			Latitude:          log.Latitude,
			Longitude:         log.Longitude,
			IsOutsideGeofence: log.IsOutsideGeofence,
			Distance:          log.Distance,
			CreatedAt:         log.CreatedAt,
		}
	}

	return responses, total, nil
}

func (s *service) GetLogsByGeofenceID(geofenceID uint, vesselID *uint, isOutside *bool, page, limit int) ([]GeofenceLogResponse, int64, error) {
	logs, total, err := s.logRepo.GetByGeofenceID(geofenceID, vesselID, isOutside, page, limit)
	if err != nil {
		return nil, 0, err
	}

	// Get geofence name once
	geofence, _ := s.geofenceRepo.GetByID(geofenceID)
	geofenceName := ""
	if geofence != nil {
		geofenceName = geofence.Name
	}

	responses := make([]GeofenceLogResponse, len(logs))
	for i, log := range logs {
		responses[i] = GeofenceLogResponse{
			ID:                log.ID,
			VesselID:          log.VesselID,
			GeofenceID:        log.GeofenceID,
			GeofenceName:      geofenceName,
			EventType:         log.EventType,
			Latitude:          log.Latitude,
			Longitude:         log.Longitude,
			Speed:             log.Speed,
			IsOutsideGeofence: log.IsOutsideGeofence,
			Distance:          log.Distance,
			Message:           log.Message,
			CreatedAt:         log.CreatedAt,
		}
	}

	return responses, total, nil
}

func (s *service) ListLogs(companyID *uint, vesselID *uint, geofenceID *uint, eventType string, page, limit int) ([]GeofenceLogResponse, int64, error) {
	logs, total, err := s.logRepo.ListLogs(companyID, vesselID, geofenceID, eventType, page, limit)
	if err != nil {
		return nil, 0, err
	}

	responses := make([]GeofenceLogResponse, len(logs))
	for i, log := range logs {
		vName := ""
		if log.Vessel != nil {
			vName = log.Vessel.Name
		}
		gName := ""
		if log.Geofence != nil {
			gName = log.Geofence.Name
		}
		responses[i] = GeofenceLogResponse{
			ID:                log.ID,
			VesselID:          log.VesselID,
			VesselName:        vName,
			GeofenceID:        log.GeofenceID,
			GeofenceName:      gName,
			EventType:         log.EventType,
			Latitude:          log.Latitude,
			Longitude:         log.Longitude,
			Speed:             log.Speed,
			IsOutsideGeofence: log.IsOutsideGeofence,
			Distance:          log.Distance,
			Message:           log.Message,
			CreatedAt:         log.CreatedAt,
		}
	}

	return responses, total, nil
}

// calculateDistanceToGeofence calculates the minimum distance from a point to the geofence
// Returns distance in meters (positive if outside, negative if inside the buffer zone)
func (s *service) calculateDistanceToGeofence(lat, lon float64, geofence *model.MGeofence) float64 {
	if len(geofence.Points) < 2 {
		return 0
	}

	minDistance := math.MaxFloat64

	// Calculate distance to all line segments
	for i := 0; i < len(geofence.Points)-1; i++ {
		distance := calculateDistanceToLineSegment(
			lat, lon,
			geofence.Points[i].Latitude, geofence.Points[i].Longitude,
			geofence.Points[i+1].Latitude, geofence.Points[i+1].Longitude,
		)
		if distance < minDistance {
			minDistance = distance
		}
	}

	// For polygons (3+ points), also check the closing segment
	if len(geofence.Points) >= 3 {
		lastIdx := len(geofence.Points) - 1
		if geofence.Points[0].Latitude != geofence.Points[lastIdx].Latitude || geofence.Points[0].Longitude != geofence.Points[lastIdx].Longitude {
			closingDist := calculateDistanceToLineSegment(
				lat, lon,
				geofence.Points[lastIdx].Latitude, geofence.Points[lastIdx].Longitude,
				geofence.Points[0].Latitude, geofence.Points[0].Longitude,
			)
			if closingDist < minDistance {
				minDistance = closingDist
			}
		}
	}

	// Calculate distance to start point (rounded cap)
	distanceToStart := calculateDistanceToPoint(
		lat, lon,
		geofence.Points[0].Latitude, geofence.Points[0].Longitude,
	)
	if distanceToStart < minDistance {
		minDistance = distanceToStart
	}

	// Calculate distance to end point (rounded cap)
	lastIdx := len(geofence.Points) - 1
	distanceToEnd := calculateDistanceToPoint(
		lat, lon,
		geofence.Points[lastIdx].Latitude, geofence.Points[lastIdx].Longitude,
	)
	if distanceToEnd < minDistance {
		minDistance = distanceToEnd
	}

	return minDistance
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

// calculateDistanceToPoint menggunakan Haversine formula
func calculateDistanceToPoint(lat1, lon1, lat2, lon2 float64) float64 {
	const earthRadius = 6371000 // Earth radius in meters

	lat1Rad := lat1 * math.Pi / 180
	lat2Rad := lat2 * math.Pi / 180
	deltaLat := (lat2 - lat1) * math.Pi / 180
	deltaLon := (lon2 - lon1) * math.Pi / 180

	a := math.Sin(deltaLat/2)*math.Sin(deltaLat/2) +
		math.Cos(lat1Rad)*math.Cos(lat2Rad)*
			math.Sin(deltaLon/2)*math.Sin(deltaLon/2)

	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))

	return earthRadius * c
}

// calculateDistanceToLineSegment menghitung jarak dari titik ke line segment
func calculateDistanceToLineSegment(px, py, x1, y1, x2, y2 float64) float64 {
	// Vector dari p1 ke p2
	dx := x2 - x1
	dy := y2 - y1

	if dx == 0 && dy == 0 {
		// p1 dan p2 adalah titik yang sama
		return calculateDistanceToPoint(px, py, x1, y1)
	}

	// Parameter t untuk proyeksi point ke line
	t := ((px-x1)*dx + (py-y1)*dy) / (dx*dx + dy*dy)

	// Clamp t ke range [0, 1] untuk line segment
	if t < 0 {
		t = 0
	} else if t > 1 {
		t = 1
	}

	// Titik terdekat di line segment
	closestX := x1 + t*dx
	closestY := y1 + t*dy

	return calculateDistanceToPoint(px, py, closestX, closestY)
}
