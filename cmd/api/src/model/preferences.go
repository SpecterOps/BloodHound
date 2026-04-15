package model

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"

	"gorm.io/gorm"
	"gorm.io/gorm/schema"
)

// PreferenceItem represents a single user preference with its value and an enterprise flag.
type PreferenceItem struct {
	Value      any  `json:"value"`
	Enterprise bool `json:"enterprise"`
}

// Preferences represents a map of preference names to their PreferenceItem values, stored as JSONB in the database.
type Preferences map[string]PreferenceItem

// Scan verifies that the input value is []byte (the Go representation of JSONB) and unmarshals it into the receiver.
// sql.Scanner implementation
func (s *Preferences) Scan(value any) error {
	var (
		bytes  []byte
		typeOK bool
	)
	if bytes, typeOK = value.([]byte); !typeOK {
		return fmt.Errorf("expected []byte representation of JSONB, received %T", value)
	}

	return json.Unmarshal(bytes, s)
}

// Value marshals the Preferences map into []byte for storage as JSONB.
// driver.Valuer implementation
func (s *Preferences) Value() (driver.Value, error) {
	return json.Marshal(s)
}

// GormDBDataType returns JSONB if postgres, otherwise panics due to lack of DB type support.
// GORM schema.DataTypeInterface implementation
func (s *Preferences) GormDBDataType(db *gorm.DB, _ *schema.Field) string {
	switch dbDialect := db.Name(); dbDialect {
	case "postgres":
		return "JSONB"

	default:
		panic(fmt.Sprintf("Unsupported database dialect for JSON datatype: %s", dbDialect))
	}
}
