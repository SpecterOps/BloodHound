package model

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
)

type CustomNodeKind struct {
	ID       int32                `json:"id"`
	KindName string               `json:"kind"`
	Config   CustomNodeKindConfig `json:"config"`
}

type CustomNodeKindConfig struct {
	Icon CustomNodeIcon `json:"icon"`
}

type CustomNodeIcon struct {
	Type string `json:"type"`
	Name string `json:"name"`
}

func (s *CustomNodeKindConfig) Scan(value interface{}) error {
	if value == nil {
		*s = CustomNodeKindConfig{}
	}

	if bytes, ok := value.([]byte); !ok {
		return errors.New("type assertion to []byte failed for SSOProviderConfig")
	} else {
		return json.Unmarshal(bytes, &s)
	}
}

func (s CustomNodeKindConfig) Value() (driver.Value, error) {
	return json.Marshal(s)
}
