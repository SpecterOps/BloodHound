package model

import (
	"encoding/json"
	"errors"
)

type CustomNodeKind struct {
	ID     int32                `json:"id"`
	KindID int16                `json:"kind_id"`
	Config CustomNodeKindConfig `json:"config" gorm:"type:jsonb column:config"`
}

type CustomNodeKindConfig struct {
	Icon CustomNodeIcon `json:"icon"`
}

type CustomNodeIcon struct {
	Type string `json:"type"`
	Name string `json:"name"`
}

func (s *CustomNodeKind) Scan(value interface{}) error {
	if value == nil {
		*s = CustomNodeKind{}
	}

	if bytes, ok := value.([]byte); !ok {
		return errors.New("type assertion to []byte failed for SSOProviderConfig")
	} else {
		return json.Unmarshal(bytes, &s.Config)
	}
}
