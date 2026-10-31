package utils

import (
	"fmt"
	"math/rand"
	"time"
)

// GenerateRandomString generates a random string with specified length
func GenerateRandomString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyz0123456789"
	seededRand := rand.New(rand.NewSource(time.Now().UnixNano()))
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[seededRand.Intn(len(charset))]
	}
	return string(b)
}

// GenerateDeletedSuffix generates a suffix for soft-deleted unique fields
// Format: _del_{date}_{random8}
// Example: _del_20260119_abc123xy (date format: YYYYMMDD)
func GenerateDeletedSuffix() string {
	date := time.Now().Format("20060102") // YYYYMMDD
	random := GenerateRandomString(8)
	return fmt.Sprintf("_del_%s_%s", date, random)
}

// GenerateDeletedUniqueID generates a unique ID for soft-deleted records
// This is used to release unique constraints on uint fields
// Format: {date}{random6} - Example: 20260119123456
// Returns: date as YYYYMMDD + 6 digit random to ensure uniqueness
func GenerateDeletedUniqueID() uint {
	date := time.Now().Format("20060102") // YYYYMMDD as string
	seededRand := rand.New(rand.NewSource(time.Now().UnixNano()))
	random := seededRand.Intn(999999) // 6 digit random

	// Convert date string to uint and combine with random
	// Example: 20260119 * 1000000 + 123456 = 20260119123456
	dateUint := uint(0)
	fmt.Sscanf(date, "%d", &dateUint)
	return dateUint*1000000 + uint(random)
}
