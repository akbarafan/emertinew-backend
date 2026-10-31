package fuel

import (
	"sync"
	"testing"

	"gorm.io/gorm/schema"
)

// The raw query that reads the engine label of every flow meter fills the
// struct with GORM, which matches a field to the column of the same name.
//
// The field used to be called `FlowEngine` while the column is `engine_name`,
// so the label never arrived: every engine came out as "Unassigned" and the
// per-engine breakdown of the command center had a single bucket. This test
// pins the field to the column, because the failure is silent — the query
// succeeds and simply returns nothing.
func TestSensorRowMapsEngineNameToItsColumn(t *testing.T) {
	parsed, err := schema.Parse(&sensorRow{}, &sync.Map{}, schema.NamingStrategy{})
	if err != nil {
		t.Fatalf("parse sensorRow: %v", err)
	}

	var column string
	for _, field := range parsed.Fields {
		if field.Name == "EngineName" {
			column = field.DBName
			break
		}
	}
	if column == "" {
		t.Fatal("sensorRow has no EngineName field")
	}
	if column != "engine_name" {
		t.Fatalf("EngineName is mapped to column %q, want %q", column, "engine_name")
	}
}
