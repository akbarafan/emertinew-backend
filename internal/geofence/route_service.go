package geofence

import (
	"errors"
	"fmt"
	"math"

	"github.com/Alugara-Inovasi-Utama/emerti/model"
)

type RouteService interface {
	CreateRoute(req CreateRouteRequest) (*RouteResponse, error)
	GetRouteByID(id uint) (*RouteResponse, error)
	GetAllRoutes(page, limit int, companyID *uint, voyageID *uint, vesselID *uint) ([]RouteResponse, int64, error)
	UpdateRoute(id uint, req UpdateRouteRequest) (*RouteResponse, error)
	DeleteRoute(id uint) error
	GenerateFromVoyage(req GenerateRouteFromVoyageRequest, companyID *uint) (*RouteResponse, error)
}

type routeService struct {
	repo RouteRepository
}

func NewRouteService(repo RouteRepository) RouteService {
	return &routeService{repo: repo}
}

// haversineNM calculates great-circle distance between two points in Nautical Miles
func haversineNM(lat1, lon1, lat2, lon2 float64) float64 {
	const earthRadiusNM = 3440.065 // Earth radius in nautical miles
	dLat := (lat2 - lat1) * math.Pi / 180.0
	dLon := (lon2 - lon1) * math.Pi / 180.0

	rLat1 := lat1 * math.Pi / 180.0
	rLat2 := lat2 * math.Pi / 180.0

	a := math.Sin(dLat/2)*math.Sin(dLat/2) +
		math.Sin(dLon/2)*math.Sin(dLon/2)*math.Cos(rLat1)*math.Cos(rLat2)
	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))

	return earthRadiusNM * c
}

func toRouteResponse(route *model.MGeofenceRoute) RouteResponse {
	resp := RouteResponse{
		ID:                     route.ID,
		CompanyID:              route.CompanyID,
		Name:                   route.Name,
		Description:            route.Description,
		TotalDistanceNM:        route.TotalDistanceNM,
		EstimatedDurationHours: route.EstimatedDurationHours,
		DefaultCorridorWidthM:  route.DefaultCorridorWidthM,
		Status:                 route.Status,
		IsActive:               route.IsActive,
		AlertOnDeviation:       route.AlertOnDeviation,
		Waypoints:              route.Waypoints,
		CreatedAt:              route.CreatedAt,
		UpdatedAt:              route.UpdatedAt,
	}

	if route.Company != nil {
		resp.CompanyName = route.Company.Name
	}
	if route.VoyageID != nil {
		resp.VoyageID = route.VoyageID
		if route.Voyage != nil {
			resp.VoyageNumber = route.Voyage.VoyageNumber
		}
	}
	if route.VesselID != nil {
		resp.VesselID = route.VesselID
		if route.Vessel != nil {
			resp.VesselName = route.Vessel.Name
		}
	}

	return resp
}

func (s *routeService) CreateRoute(req CreateRouteRequest) (*RouteResponse, error) {
	if len(req.Waypoints) < 2 {
		return nil, errors.New("rute membutuhkan minimal 2 waypoint (asal dan tujuan)")
	}

	defaultCorridor := req.DefaultCorridorWidthM
	if defaultCorridor <= 0 {
		defaultCorridor = 926 // 0.5 NM
	}

	status := req.Status
	if status == "" {
		status = "draft"
	}

	isActive := true
	if req.IsActive != nil {
		isActive = *req.IsActive
	}

	alertOnDev := true
	if req.AlertOnDeviation != nil {
		alertOnDev = *req.AlertOnDeviation
	}

	var totalDist float64
	var waypoints []model.MGeofenceRouteWaypoint

	for i, wp := range req.Waypoints {
		seq := i + 1
		if wp.Sequence > 0 {
			seq = wp.Sequence
		}

		wpType := wp.WaypointType
		if wpType == "" {
			if i == 0 || i == len(req.Waypoints)-1 {
				wpType = "port_call"
			} else {
				wpType = "waypoint"
			}
		}

		corridorW := wp.CorridorWidthM
		if corridorW <= 0 {
			corridorW = defaultCorridor
		}

		radiusM := wp.RadiusM
		if radiusM <= 0 {
			radiusM = 185 // 0.1 NM
		}

		name := wp.Name
		if name == "" {
			if i == 0 {
				name = "Titik Keberangkatan"
			} else if i == len(req.Waypoints)-1 {
				name = "Titik Tujuan"
			} else {
				name = fmt.Sprintf("WP-%d", seq)
			}
		}

		if i > 0 {
			prev := req.Waypoints[i-1]
			totalDist += haversineNM(prev.Latitude, prev.Longitude, wp.Latitude, wp.Longitude)
		}

		waypoints = append(waypoints, model.MGeofenceRouteWaypoint{
			Sequence:       seq,
			Name:           name,
			Latitude:       wp.Latitude,
			Longitude:      wp.Longitude,
			WaypointType:   wpType,
			CorridorWidthM: corridorW,
			RadiusM:        radiusM,
			Notes:          wp.Notes,
		})
	}

	if req.TotalDistanceNM > 0 {
		totalDist = req.TotalDistanceNM
	}

	var estDuration *float64
	if req.EstimatedDurationHours != nil {
		estDuration = req.EstimatedDurationHours
	} else if totalDist > 0 {
		hours := totalDist / 8.0 // default estimate at 8 knots
		estDuration = &hours
	}

	route := model.MGeofenceRoute{
		CompanyID:              req.CompanyID,
		VoyageID:               req.VoyageID,
		VesselID:               req.VesselID,
		Name:                   req.Name,
		Description:            req.Description,
		TotalDistanceNM:        math.Round(totalDist*100) / 100,
		EstimatedDurationHours: estDuration,
		DefaultCorridorWidthM:  defaultCorridor,
		Status:                 status,
		IsActive:               isActive,
		AlertOnDeviation:       alertOnDev,
		Waypoints:              waypoints,
	}

	if err := s.repo.Create(&route); err != nil {
		return nil, err
	}

	// Fetch with preloads
	created, err := s.repo.FindByID(route.ID)
	if err != nil {
		resp := toRouteResponse(&route)
		return &resp, nil
	}

	resp := toRouteResponse(created)
	return &resp, nil
}

func (s *routeService) GetRouteByID(id uint) (*RouteResponse, error) {
	route, err := s.repo.FindByID(id)
	if err != nil {
		return nil, err
	}
	resp := toRouteResponse(route)
	return &resp, nil
}

func (s *routeService) GetAllRoutes(page, limit int, companyID *uint, voyageID *uint, vesselID *uint) ([]RouteResponse, int64, error) {
	routes, total, err := s.repo.FindAll(page, limit, companyID, voyageID, vesselID)
	if err != nil {
		return nil, 0, err
	}

	var resps []RouteResponse
	for i := range routes {
		resps = append(resps, toRouteResponse(&routes[i]))
	}
	return resps, total, nil
}

func (s *routeService) UpdateRoute(id uint, req UpdateRouteRequest) (*RouteResponse, error) {
	route, err := s.repo.FindByID(id)
	if err != nil {
		return nil, err
	}

	if req.CompanyID != nil {
		route.CompanyID = req.CompanyID
	}
	if req.VoyageID != nil {
		route.VoyageID = req.VoyageID
	}
	if req.VesselID != nil {
		route.VesselID = req.VesselID
	}
	if req.Name != nil && *req.Name != "" {
		route.Name = *req.Name
	}
	if req.Description != nil {
		route.Description = *req.Description
	}
	if req.DefaultCorridorWidthM != nil && *req.DefaultCorridorWidthM > 0 {
		route.DefaultCorridorWidthM = *req.DefaultCorridorWidthM
	}
	if req.Status != nil && *req.Status != "" {
		route.Status = *req.Status
	}
	if req.IsActive != nil {
		route.IsActive = *req.IsActive
	}
	if req.AlertOnDeviation != nil {
		route.AlertOnDeviation = *req.AlertOnDeviation
	}
	if req.EstimatedDurationHours != nil {
		route.EstimatedDurationHours = req.EstimatedDurationHours
	}

	var newWaypoints []model.MGeofenceRouteWaypoint
	if len(req.Waypoints) > 0 {
		var totalDist float64
		for i, wp := range req.Waypoints {
			seq := i + 1
			if wp.Sequence > 0 {
				seq = wp.Sequence
			}
			wpType := wp.WaypointType
			if wpType == "" {
				if i == 0 || i == len(req.Waypoints)-1 {
					wpType = "port_call"
				} else {
					wpType = "waypoint"
				}
			}
			corridorW := wp.CorridorWidthM
			if corridorW <= 0 {
				corridorW = route.DefaultCorridorWidthM
			}
			radiusM := wp.RadiusM
			if radiusM <= 0 {
				radiusM = 185
			}
			name := wp.Name
			if name == "" {
				name = fmt.Sprintf("WP-%d", seq)
			}

			if i > 0 {
				prev := req.Waypoints[i-1]
				totalDist += haversineNM(prev.Latitude, prev.Longitude, wp.Latitude, wp.Longitude)
			}

			newWaypoints = append(newWaypoints, model.MGeofenceRouteWaypoint{
				RouteID:        route.ID,
				Sequence:       seq,
				Name:           name,
				Latitude:       wp.Latitude,
				Longitude:      wp.Longitude,
				WaypointType:   wpType,
				CorridorWidthM: corridorW,
				RadiusM:        radiusM,
				Notes:          wp.Notes,
			})
		}
		route.TotalDistanceNM = math.Round(totalDist*100) / 100
		if req.TotalDistanceNM != nil && *req.TotalDistanceNM > 0 {
			route.TotalDistanceNM = *req.TotalDistanceNM
		}
		if route.EstimatedDurationHours == nil && route.TotalDistanceNM > 0 {
			h := route.TotalDistanceNM / 8.0
			route.EstimatedDurationHours = &h
		}
	}

	if err := s.repo.Update(route, newWaypoints); err != nil {
		return nil, err
	}

	updated, err := s.repo.FindByID(route.ID)
	if err != nil {
		resp := toRouteResponse(route)
		return &resp, nil
	}
	resp := toRouteResponse(updated)
	return &resp, nil
}

func (s *routeService) DeleteRoute(id uint) error {
	return s.repo.Delete(id)
}

func (s *routeService) GenerateFromVoyage(req GenerateRouteFromVoyageRequest, companyID *uint) (*RouteResponse, error) {
	voyage, err := s.repo.GetVoyageByID(req.VoyageID)
	if err != nil {
		return nil, fmt.Errorf("voyage dengan ID %d tidak ditemukan: %w", req.VoyageID, err)
	}

	if voyage.OriginLat == nil || voyage.OriginLong == nil {
		return nil, errors.New("voyage belum memiliki koordinat pelabuhan asal (OriginLat/Long)")
	}
	if voyage.DestinationLat == nil || voyage.DestinationLong == nil {
		return nil, errors.New("voyage belum memiliki koordinat tujuan (DestinationLat/Long)")
	}

	origLat, origLong := *voyage.OriginLat, *voyage.OriginLong
	destLat, destLong := *voyage.DestinationLat, *voyage.DestinationLong

	totalDist := haversineNM(origLat, origLong, destLat, destLong)

	corridorWidth := req.DefaultCorridorWidthM
	if corridorWidth <= 0 {
		corridorWidth = 926 // 0.5 NM
	}

	routeName := req.Name
	if routeName == "" {
		routeName = fmt.Sprintf("%s → %s", voyage.OriginName, voyage.DestinationName)
	}

	interval := req.WaypointIntervalNM
	if interval <= 0 {
		interval = 15.0 // default generate waypoint every 15 NM
	}

	var wps []WaypointDTO
	// Origin waypoint
	wps = append(wps, WaypointDTO{
		Sequence:       1,
		Name:           fmt.Sprintf("%s (Asal)", voyage.OriginName),
		Latitude:       origLat,
		Longitude:      origLong,
		WaypointType:   "port_call",
		CorridorWidthM: corridorWidth,
		RadiusM:        300,
	})

	// Interpolate intermediate waypoints if distance > interval
	numIntermediate := int(totalDist / interval)
	if numIntermediate > 0 && numIntermediate < 20 { // sensible cap
		for i := 1; i <= numIntermediate; i++ {
			fraction := float64(i) / float64(numIntermediate+1)
			interLat := origLat + fraction*(destLat-origLat)
			interLong := origLong + fraction*(destLong-origLong)

			wps = append(wps, WaypointDTO{
				Sequence:       i + 1,
				Name:           fmt.Sprintf("WP-%d", i+1),
				Latitude:       math.Round(interLat*10000) / 10000,
				Longitude:      math.Round(interLong*10000) / 10000,
				WaypointType:   "waypoint",
				CorridorWidthM: corridorWidth,
				RadiusM:        185,
			})
		}
	}

	// Destination waypoint
	wps = append(wps, WaypointDTO{
		Sequence:       len(wps) + 1,
		Name:           fmt.Sprintf("%s (Tujuan)", voyage.DestinationName),
		Latitude:       destLat,
		Longitude:      destLong,
		WaypointType:   "port_call",
		CorridorWidthM: corridorWidth,
		RadiusM:        300,
	})

	createReq := CreateRouteRequest{
		CompanyID:             companyID,
		VoyageID:              &voyage.ID,
		VesselID:              &voyage.VesselID,
		Name:                  routeName,
		Description:           fmt.Sprintf("Rute otomatis dibuat dari Voyage %s", voyage.VoyageNumber),
		TotalDistanceNM:       math.Round(totalDist*100) / 100,
		DefaultCorridorWidthM: corridorWidth,
		Status:                "draft",
		Waypoints:             wps,
	}

	return s.CreateRoute(createReq)
}
