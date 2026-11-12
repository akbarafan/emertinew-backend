package tracking

import (
	"encoding/json"
	"fmt"
	"math"
	"net/http"
	"sync"
	"time"
)

// BMKGWeatherSnapshot represents maritime weather data for a given point
type BMKGWeatherSnapshot struct {
	WaveHeight   *float64
	WaveCat      *string
	CurrentSpeed *float64
	CurrentTo    *string
	WindSpeed    *float64
	WindGust     *float64
	WindFrom     *string
	Weather      *string
}

type bmkgPortMeta struct {
	Code     string
	Name     string
	Province string
	Lat      float64
	Lon      float64
}

type rawPortFeatureCollection struct {
	Features []struct {
		Properties struct {
			Code     string `json:"code"`
			Name     string `json:"name"`
			Province string `json:"province"`
		} `json:"properties"`
		Geometry struct {
			Coordinates []float64 `json:"coordinates"` // [lon, lat]
		} `json:"geometry"`
	} `json:"features"`
}

type rawPortForecast struct {
	Code         string             `json:"code"`
	Name         string             `json:"name"`
	ForecastDay1 []rawForecastPoint `json:"forecast_day1"`
	ForecastDay2 []rawForecastPoint `json:"forecast_day2-4"`
}

type rawForecastPoint struct {
	Time         string   `json:"time"`
	Weather      string   `json:"weather"`
	WaveCat      *string  `json:"wave_cat"`
	WaveHeight   *float64 `json:"wave_height"`
	CurrentTo    *string  `json:"current_to"`
	CurrentSpeed *float64 `json:"current_speed"`
	WindFrom     *string  `json:"wind_from"`
	WindSpeed    *float64 `json:"wind_speed"`
	WindGust     *float64 `json:"wind_gust"`
}

var (
	bmkgClient = &http.Client{
		Timeout: 10 * time.Second,
	}

	weatherCacheMutex sync.RWMutex
	cachedPorts       []bmkgPortMeta
	portsLastFetched  time.Time

	// portForecastCache maps portCode -> { snapshot, expiresAt }
	portForecastCache = make(map[string]cachedForecastEntry)
)

type cachedForecastEntry struct {
	snapshot  *BMKGWeatherSnapshot
	expiresAt time.Time
}

func haversineKm(lat1, lon1, lat2, lon2 float64) float64 {
	const R = 6371.0
	dLat := (lat2 - lat1) * math.Pi / 180.0
	dLon := (lon2 - lon1) * math.Pi / 180.0
	a := math.Sin(dLat/2)*math.Sin(dLat/2) +
		math.Cos(lat1*math.Pi/180.0)*math.Cos(lat2*math.Pi/180.0)*
			math.Sin(dLon/2)*math.Sin(dLon/2)
	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))
	return R * c
}

func fetchBMKGPorts() ([]bmkgPortMeta, error) {
	weatherCacheMutex.RLock()
	if len(cachedPorts) > 0 && time.Since(portsLastFetched) < 24*time.Hour {
		ports := cachedPorts
		weatherCacheMutex.RUnlock()
		return ports, nil
	}
	weatherCacheMutex.RUnlock()

	years := []int{time.Now().UTC().Year(), time.Now().UTC().Year() - 1}
	var lastErr error

	for _, y := range years {
		url := fmt.Sprintf("https://maritim.bmkg.go.id/marine%d-data/meta/pelabuhan.json", y)
		req, err := http.NewRequest(http.MethodGet, url, nil)
		if err != nil {
			lastErr = err
			continue
		}
		req.Header.Set("User-Agent", "EMERTI-Fleet-Telemetry/1.0")

		resp, err := bmkgClient.Do(req)
		if err != nil {
			lastErr = err
			continue
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			lastErr = fmt.Errorf("status %d", resp.StatusCode)
			continue
		}

		var fc rawPortFeatureCollection
		if err := json.NewDecoder(resp.Body).Decode(&fc); err != nil {
			lastErr = err
			continue
		}

		var ports []bmkgPortMeta
		for _, f := range fc.Features {
			if len(f.Geometry.Coordinates) >= 2 && f.Properties.Code != "" {
				ports = append(ports, bmkgPortMeta{
					Code:     f.Properties.Code,
					Name:     f.Properties.Name,
					Province: f.Properties.Province,
					Lon:      f.Geometry.Coordinates[0],
					Lat:      f.Geometry.Coordinates[1],
				})
			}
		}

		if len(ports) > 0 {
			weatherCacheMutex.Lock()
			cachedPorts = ports
			portsLastFetched = time.Now()
			weatherCacheMutex.Unlock()
			return ports, nil
		}
	}

	return nil, lastErr
}

func getNearestPort(lat, lon float64) (*bmkgPortMeta, error) {
	ports, err := fetchBMKGPorts()
	if err != nil || len(ports) == 0 {
		return nil, err
	}

	var best *bmkgPortMeta
	bestDist := math.MaxFloat64
	for i := range ports {
		p := &ports[i]
		d := haversineKm(lat, lon, p.Lat, p.Lon)
		if d < bestDist {
			bestDist = d
			best = p
		}
	}
	return best, nil
}

// GetBMKGWeather retrieves or fetches the maritime weather snapshot for the coordinates
func GetBMKGWeather(lat, lon float64) *BMKGWeatherSnapshot {
	port, err := getNearestPort(lat, lon)
	if err != nil || port == nil {
		return nil
	}

	weatherCacheMutex.RLock()
	if entry, ok := portForecastCache[port.Code]; ok {
		if time.Now().Before(entry.expiresAt) {
			weatherCacheMutex.RUnlock()
			return entry.snapshot
		}
	}
	weatherCacheMutex.RUnlock()

	years := []int{time.Now().UTC().Year(), time.Now().UTC().Year() - 1}
	for _, y := range years {
		url := fmt.Sprintf("https://maritim.bmkg.go.id/marine%d-data/pelabuhan/%s.json", y, port.Code)
		req, err := http.NewRequest(http.MethodGet, url, nil)
		if err != nil {
			continue
		}
		req.Header.Set("User-Agent", "EMERTI-Fleet-Telemetry/1.0")

		resp, err := bmkgClient.Do(req)
		if err != nil {
			continue
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			continue
		}

		var data rawPortForecast
		if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
			continue
		}

		all := append(data.ForecastDay1, data.ForecastDay2...)
		if len(all) == 0 {
			continue
		}

		// Find point closest to current time
		now := time.Now().UTC()
		var bestPt *rawForecastPoint
		minDiff := time.Duration(1<<63 - 1)

		for i := range all {
			pt := &all[i]
			t, err := time.Parse("2006-01-02 15:04 UTC", pt.Time)
			if err != nil {
				t, err = time.Parse("2006-01-02 15:04:05", pt.Time)
			}
			if err == nil {
				diff := now.Sub(t)
				if diff < 0 {
					diff = -diff
				}
				if diff < minDiff {
					minDiff = diff
					bestPt = pt
				}
			}
		}

		if bestPt == nil {
			bestPt = &all[0]
		}

		snap := &BMKGWeatherSnapshot{
			WaveHeight:   bestPt.WaveHeight,
			WaveCat:      bestPt.WaveCat,
			CurrentSpeed: bestPt.CurrentSpeed,
			CurrentTo:    bestPt.CurrentTo,
			WindSpeed:    bestPt.WindSpeed,
			WindGust:     bestPt.WindGust,
			WindFrom:     bestPt.WindFrom,
		}
		if bestPt.Weather != "" {
			w := bestPt.Weather
			snap.Weather = &w
		}

		weatherCacheMutex.Lock()
		portForecastCache[port.Code] = cachedForecastEntry{
			snapshot:  snap,
			expiresAt: time.Now().Add(1 * time.Hour),
		}
		weatherCacheMutex.Unlock()

		return snap
	}

	return nil
}
