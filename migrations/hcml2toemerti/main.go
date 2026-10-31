package main

import (
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/Alugara-Inovasi-Utama/emerti/config"
	"github.com/Alugara-Inovasi-Utama/emerti/model"
	"github.com/xuri/excelize/v2"
	"gorm.io/gorm"
)

// VesselMapping maps vessel names from Excel to vessel IDs in DB
var VesselMapping = map[string]uint{
	"AHT KYLIE S":      324,
	"CB.ALDWIN":        322,
	"CB ALDWIN":        322,
	"CB. ALDWIN":       322,
	"HT.LIMIN ASD 43":  321,
	"HT LIMIN ASD 43":  321,
	"HT. LIMIN ASD 43": 321,
	"HT.LIMIN":         321,
	"AHT. LIMIN":       321,
	"NMS ACHIEVE":      320,
	"OPS ALHENA":       323,
}

// Track unknown vessels for debugging
var unknownVessels = make(map[string]int)

func main() {
	log.Println("[MIGRATION] Starting HCML2 to Emerti Migration...")
	log.Println("=========================================")

	// Load config
	cfg := config.LoadConfig()

	// Connect to database
	config.ConnectDatabase(*cfg)
	db := config.DB

	// Step 1: Verify vessels exist
	log.Println("\n📦 Step 1: Verifying master data...")
	if !verifyMasterData(db) {
		log.Fatal("[MIGRATION] Master data not found. Please run 'make run-seed' first!")
	}

	// Step 2: Read Excel file
	log.Println("[MIGRATION] Step 2: Reading Excel file...")
	excelPath := "assets/file/emerti_octavms.xlsx"

	if _, err := os.Stat(excelPath); os.IsNotExist(err) {
		log.Fatalf("[MIGRATION] Excel file not found: %s", excelPath)
	}

	f, err := excelize.OpenFile(excelPath)
	if err != nil {
		log.Fatalf("[MIGRATION] Failed to open Excel file: %v", err)
	}
	defer f.Close()

	// Get sheet names
	sheets := f.GetSheetList()
	log.Printf("📋 Found %d sheets: %v", len(sheets), sheets)

	// Step 3: Process each sheet (should be tracking data)
	for _, sheet := range sheets {
		log.Printf("\n📄 Processing sheet: %s", sheet)
		processSheet(db, f, sheet)
	}

	log.Println("\n=========================================")
	log.Println("[MIGRATION] Completed!")
}

// verifyMasterData checks if vessels exist in database
func verifyMasterData(db *gorm.DB) bool {
	var count int64
	db.Model(&model.MVessel{}).Count(&count)
	if count == 0 {
		log.Println("[MIGRATION] No vessels found in database")
		return false
	}
	log.Printf("[MIGRATION] Found %d vessels in database", count)

	// Log vessel mapping info
	for name, id := range VesselMapping {
		var vessel model.MVessel
		if err := db.First(&vessel, id).Error; err == nil {
			log.Printf("[MIGRATION] Vessel mapping: %s -> ID %d (%s)", name, id, vessel.Name)
		}
	}

	return true
}

// processSheet processes a single sheet from Excel
func processSheet(db *gorm.DB, f *excelize.File, sheet string) {
	rows, err := f.GetRows(sheet)
	if err != nil {
		log.Printf("[MIGRATION] Failed to get rows from sheet %s: %v", sheet, err)
		return
	}

	if len(rows) < 2 {
		log.Printf("[MIGRATION] Sheet %s has no data rows", sheet)
		return
	}

	// Get headers (first row)
	headers := rows[0]
	log.Printf("[MIGRATION] Headers: %v", headers)
	log.Printf("[MIGRATION] Total rows: %d", len(rows)-1)

	// Create column index map
	colIndex := make(map[string]int)
	for i, h := range headers {
		colIndex[strings.ToLower(strings.TrimSpace(h))] = i
	}

	// Log column mapping
	log.Printf("📍 Column mapping: %v", colIndex)

	// Process data rows in batches
	batchSize := 1000
	var trackings []model.Tracking
	totalProcessed := 0
	totalCreated := 0
	skipped := 0

	for i := 1; i < len(rows); i++ {
		row := rows[i]
		if len(row) == 0 {
			continue
		}

		tracking := parseRowToTracking(row, colIndex)
		if tracking == nil {
			skipped++
			continue
		}

		trackings = append(trackings, *tracking)

		// Save batch
		if len(trackings) >= batchSize {
			created := saveBatch(db, trackings)
			totalCreated += created
			totalProcessed += len(trackings)
			log.Printf("[MIGRATION] Progress: %d/%d rows (created: %d)", totalProcessed, len(rows)-1, totalCreated)
			trackings = nil
		}
	}

	// Save remaining
	if len(trackings) > 0 {
		created := saveBatch(db, trackings)
		totalCreated += created
		totalProcessed += len(trackings)
	}

	log.Printf("[MIGRATION] Sheet %s completed: %d rows processed, %d created, %d skipped", sheet, totalProcessed, totalCreated, skipped)

	// Log unknown vessels
	if len(unknownVessels) > 0 {
		log.Println("[MIGRATION] Unknown vessels found:")
		for name, count := range unknownVessels {
			log.Printf("   - '%s': %d rows", name, count)
		}
	}
}

// parseRowToTracking parses a row to Tracking model
func parseRowToTracking(row []string, colIndex map[string]int) *model.Tracking {
	getValue := func(key string) string {
		if idx, ok := colIndex[key]; ok && idx < len(row) {
			return strings.TrimSpace(row[idx])
		}
		return ""
	}

	// Parse date/time
	dateStr := getValue("date")
	if dateStr == "" {
		return nil
	}

	// Try multiple date formats
	var trackingTime time.Time
	var err error
	formats := []string{
		"2006-01-02 15:04:05",
		"2006-01-02 15:04",
		"02/01/2006 15:04:05",
		"02/01/2006 15:04",
		"01/02/2006 15:04:05",
		"01/02/2006 15:04",
		time.RFC3339,
	}

	for _, format := range formats {
		trackingTime, err = time.Parse(format, dateStr)
		if err == nil {
			break
		}
	}

	if err != nil {
		// Try parsing Excel serial date
		if serial, err := strconv.ParseFloat(dateStr, 64); err == nil {
			trackingTime = excelSerialToTime(serial)
		} else {
			return nil
		}
	}

	// Set timezone to Jakarta
	loc, _ := time.LoadLocation("Asia/Jakarta")
	trackingTime = trackingTime.In(loc)

	// Parse vessel
	vesselName := getValue("vessel")
	vesselID, ok := VesselMapping[vesselName]
	if !ok {
		// Try case-insensitive match
		for k, v := range VesselMapping {
			if strings.EqualFold(k, vesselName) {
				vesselID = v
				ok = true
				break
			}
		}
	}
	if !ok {
		unknownVessels[vesselName]++
		return nil
	}

	tracking := &model.Tracking{
		Time:     trackingTime,
		VesselID: vesselID,
	}

	// Parse latitude
	if lat := getValue("latitude"); lat != "" {
		if v, err := strconv.ParseFloat(lat, 64); err == nil {
			tracking.Lat = &v
		}
	}

	// Parse longitude
	if lon := getValue("longitude"); lon != "" {
		if v, err := strconv.ParseFloat(lon, 64); err == nil {
			tracking.Lon = &v
		}
	}

	// Parse speed
	if speed := getValue("speed"); speed != "" {
		if v, err := strconv.ParseFloat(speed, 64); err == nil {
			tracking.Speed = &v
		}
	}

	// Parse heading
	if heading := getValue("heading"); heading != "" {
		if v, err := strconv.ParseFloat(heading, 64); err == nil {
			tracking.Heading = &v
		}
	}

	// Parse RPM 1
	if rpm1 := getValue("rpm 1"); rpm1 != "" {
		if v, err := strconv.ParseFloat(rpm1, 64); err == nil {
			intVal := int(v)
			tracking.RPM1 = &intVal
		}
	}

	// Parse RPM 2
	if rpm2 := getValue("rpm 2"); rpm2 != "" {
		if v, err := strconv.ParseFloat(rpm2, 64); err == nil {
			intVal := int(v)
			tracking.RPM2 = &intVal
		}
	}

	return tracking
}

// excelSerialToTime converts Excel serial date to time.Time
func excelSerialToTime(serial float64) time.Time {
	// Excel's epoch starts at January 1, 1900
	// But Excel incorrectly considers 1900 as a leap year
	// So we need to subtract 2 days (1 for epoch diff, 1 for leap year bug)
	excelEpoch := time.Date(1899, 12, 30, 0, 0, 0, 0, time.UTC)
	days := int(serial)
	fraction := serial - float64(days)

	result := excelEpoch.AddDate(0, 0, days)
	result = result.Add(time.Duration(fraction * 24 * float64(time.Hour)))

	return result
}

// saveBatch saves a batch of trackings to database
func saveBatch(db *gorm.DB, trackings []model.Tracking) int {
	if len(trackings) == 0 {
		return 0
	}

	// Use ON CONFLICT to skip duplicates (based on time + vessel_id)
	result := db.CreateInBatches(trackings, 100)
	if result.Error != nil {
		log.Printf("[MIGRATION] Batch save error: %v", result.Error)
		return 0
	}

	return int(result.RowsAffected)
}
