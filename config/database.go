package config

import (
	"database/sql"
	"fmt"
	"log"

	"github.com/Alugara-Inovasi-Utama/emerti/model"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"github.com/lib/pq" // untuk QuoteIdentifier
)

var DB *gorm.DB

func ConnectDatabase(cfg Config) {
	// Langkah 1: Connect dulu ke default DB "postgres"
	defaultDSN := fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s port=%s sslmode=disable TimeZone=Asia/Jakarta",
		cfg.DBHost, cfg.DBUser, cfg.DBPassword, cfg.DBName, cfg.DBPort,
	)

	sqlDB, err := sql.Open("postgres", defaultDSN)
	if err != nil {
		log.Fatal("Failed to connect to default DB:", err)
	}
	defer sqlDB.Close()

	// Langkah 2: Cek apakah database sudah ada
	var exists bool
	err = sqlDB.QueryRow("SELECT EXISTS(SELECT datname FROM pg_database WHERE datname = $1)", cfg.DBName).Scan(&exists)
	if err != nil {
		log.Fatal("Failed to check if DB exists:", err)
	}

	// Langkah 3: Jika belum ada, buat
	if !exists {
		createQuery := fmt.Sprintf("CREATE DATABASE %s", pq.QuoteIdentifier(cfg.DBName))
		_, err = sqlDB.Exec(createQuery)
		if err != nil {
			log.Fatal("Failed to create database:", err)
		}
		fmt.Println("Database", cfg.DBName, "created successfully.")
	}

	// Langkah 4: Connect ke DB tujuan
	targetDSN := fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s port=%s sslmode=disable TimeZone=Asia/Jakarta",
		cfg.DBHost, cfg.DBUser, cfg.DBPassword, cfg.DBName, cfg.DBPort,
	)

	DB, err = gorm.Open(postgres.Open(targetDSN), &gorm.Config{})
	if err != nil {
		log.Fatal("Failed to connect to target DB:", err)
	}

	// Langkah 5: Auto migrate
	// Urutan penting: parent tables dulu sebelum child tables dengan foreign keys
	err = DB.AutoMigrate(
		&model.MCompany{},
		&model.User{},
		&model.AuditLog{},
		&model.MClassification{},
		&model.MVesselType{},
		&model.MOperationStatus{},
		&model.MMaterialHull{},
		&model.MFleet{},
		&model.MSquad{},
		&model.MFuelType{},
		&model.MVessel{},
		&model.VesselDimensions{},
		&model.VesselCapacity{},
		&model.MMCU{},
		&model.MSensor{},
		&model.SensorData{},
		&model.Tracking{},
		&model.MGeofence{},
		&model.GeofenceLog{},
		&model.MGeofenceRoute{},
		&model.MGeofenceRouteWaypoint{},
		&model.MOffshore{},
		&model.MPort{},
		&model.MCrew{},
		&model.Manifest{},
		&model.VesselGeofenceAssignment{},
		&model.VesselCpLog{},
		&model.Voyage{},
		&model.SupportTicket{},
		&model.TicketReply{},
		&model.VesselEngineRunningHours{},
		&model.TripRunningHoursResetLog{},
	)
	if err != nil {
		log.Fatal("Database migration failed:", err)
	}

	// Backfill topic_key for existing sensors: normalize / to : separator
	DB.Exec(`
		UPDATE m_sensors
		SET topic_key = REPLACE(topic, '/', ':')
		WHERE topic_key IS NULL
	`)

	// Create unique index for topic_key (IF NOT EXISTS = safe on every startup)
	DB.Exec("CREATE UNIQUE INDEX IF NOT EXISTS idx_m_sensors_topic_key ON m_sensors(topic_key)")

	// Auto-heal PostgreSQL sequences that may fall out of sync after pg_dump restores
	syncTableSequences(DB)

	// Langkah 6: Setup TimescaleDB extension dan hypertables
	setupTimescaleDB(DB)

	// Langkah 7: Buat partial index untuk tracking GPS (otomatis tiap startup, skip jika sudah ada)
	createTrackingIndexes(DB)

	fmt.Println("Connected to database and migration successful.")
}

// setupTimescaleDB enables TimescaleDB extension and creates hypertables
// Runs in background to not block server startup
func setupTimescaleDB(db *gorm.DB) {
	go func() {
		log.Println("Setting up TimescaleDB in background...")

		// Enable TimescaleDB extension
		err := db.Exec("CREATE EXTENSION IF NOT EXISTS timescaledb CASCADE").Error
		if err != nil {
			log.Printf("Warning: Failed to create TimescaleDB extension: %v", err)
			return
		}

		// Check if sensor_data hypertable already exists
		var sensorHypertableExists bool
		db.Raw(`
			SELECT EXISTS(
				SELECT 1 FROM timescaledb_information.hypertables 
				WHERE hypertable_name = 'sensor_data'
			)
		`).Scan(&sensorHypertableExists)

		if !sensorHypertableExists {
			// Create hypertable for sensor_data
			err = db.Exec(`
				SELECT create_hypertable(
					'sensor_data', 
					'time',
					chunk_time_interval => INTERVAL '1 day',
					if_not_exists => TRUE
				)
			`).Error
			if err != nil {
				log.Printf("Warning: Failed to create hypertable for sensor_data: %v", err)
			} else {
				log.Println("Hypertable created for sensor_data")

				// Enable compression on the hypertable first
				db.Exec(`ALTER TABLE sensor_data SET (timescaledb.compress)`)

				// Add compression policy
				err = db.Exec(`
					SELECT add_compression_policy('sensor_data', INTERVAL '7 days', if_not_exists => TRUE)
				`).Error
				if err != nil {
					log.Printf("Warning: Failed to add compression policy for sensor_data: %v", err)
				} else {
					log.Println("Compression policy added for sensor_data")
				}
			}
		} else {
			log.Println("Hypertable for sensor_data already exists, skipping...")
		}

		// Check if trackings hypertable already exists
		var trackingHypertableExists bool
		db.Raw(`
			SELECT EXISTS(
				SELECT 1 FROM timescaledb_information.hypertables 
				WHERE hypertable_name = 'trackings'
			)
		`).Scan(&trackingHypertableExists)

		if !trackingHypertableExists {
			// Create hypertable for tracking
			err = db.Exec(`
				SELECT create_hypertable(
					'trackings', 
					'time',
					chunk_time_interval => INTERVAL '1 day',
					if_not_exists => TRUE
				)
			`).Error
			if err != nil {
				log.Printf("Warning: Failed to create hypertable for tracking: %v", err)
			} else {
				log.Println("Hypertable created for tracking")

				// Enable compression on the hypertable first
				db.Exec(`ALTER TABLE trackings SET (timescaledb.compress)`)

				// Add compression policy
				err = db.Exec(`
					SELECT add_compression_policy('trackings', INTERVAL '7 days', if_not_exists => TRUE)
				`).Error
				if err != nil {
					log.Printf("Warning: Failed to add compression policy for tracking: %v", err)
				} else {
					log.Println("Compression policy added for tracking")
				}
			}
		} else {
			log.Println("Hypertable for trackings already exists, skipping...")
		}

		log.Println("TimescaleDB setup completed in background")
	}()
}

// createTrackingIndexes membuat partial index untuk query tracking GPS tercepat
// Partial index ini memungkinkan query WHERE lat IS NOT NULL AND lon IS NOT NULL
// langsung lompat ke record GPS terakhir tanpa scan data sensor doang.
func createTrackingIndexes(db *gorm.DB) {
	err := db.Exec(`
		CREATE INDEX IF NOT EXISTS idx_tracking_vessel_coords
		ON trackings (vessel_id, time DESC)
		WHERE lat IS NOT NULL AND lon IS NOT NULL
	`).Error
	if err != nil {
		log.Printf("Warning: Failed to create tracking coordinates index: %v", err)
	} else {
		log.Println("Index idx_tracking_vessel_coords created/verified successfully")
	}
}

// syncTableSequences checks and advances PostgreSQL serial sequences to MAX(id).
// This auto-heals "duplicate key value violates unique constraint *_pkey" errors
// that occur after restoring database dumps where sequences were not advanced.
func syncTableSequences(db *gorm.DB) {
	tables := []string{
		"m_vessels",
		"m_fleets",
		"m_squads",
		"m_company",
		"m_mcu",
		"m_sensors",
		"users",
		"vessel_capacities",
		"vessel_dimensions",
		"m_alert_events",
		"m_classification",
		"m_crew",
		"m_fuel_types",
		"m_geofences",
		"m_material_hull",
		"m_offshore",
		"m_operation_status",
		"m_port",
		"m_vessel_type",
		"t_voyages",
		"t_support_tickets",
		"t_ticket_replies",
	}

	for _, tbl := range tables {
		query := fmt.Sprintf(`
			DO $$
			DECLARE
				seq text;
			BEGIN
				seq := pg_get_serial_sequence('%s', 'id');
				IF seq IS NOT NULL THEN
					EXECUTE format('SELECT setval(%%L, COALESCE((SELECT MAX(id) FROM %s), 1))', seq);
				END IF;
			END $$;
		`, tbl, tbl)
		_ = db.Exec(query).Error
	}
}
