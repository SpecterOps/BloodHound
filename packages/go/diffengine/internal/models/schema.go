package models

import (
	"time"
	"gorm.io/gorm"
)

type Schema struct {
	ID        uint           `gorm:"primarykey" json:"id,omitempty"`
	CreatedAt time.Time      `json:"created_at,omitempty"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
	Name      string         `gorm:"uniqueIndex;not null" json:"name"`
	Version   string         `gorm:"not null" json:"version"`
	Fields    []Field        `gorm:"foreignKey:SchemaID;constraint:OnDelete:CASCADE" json:"fields"`
	Metadata  *Metadata      `gorm:"foreignKey:SchemaID;constraint:OnDelete:CASCADE" json:"metadata,omitempty"`
	Tags      []Tag          `gorm:"foreignKey:SchemaID;constraint:OnDelete:CASCADE" json:"tags,omitempty"`
}

type Field struct {
	ID         uint           `gorm:"primarykey" json:"id,omitempty"`
	CreatedAt  time.Time      `json:"-"`
	UpdatedAt  time.Time      `json:"-"`
	DeletedAt  gorm.DeletedAt `gorm:"index" json:"-"`
	SchemaID   uint           `gorm:"not null;index" json:"schema_id,omitempty"`
	Name       string         `gorm:"column:field_name;not null" json:"name"`
	Type       string         `gorm:"column:field_type;not null" json:"type"`
	IsRequired bool           `gorm:"default:false" json:"is_required"`
}

type Metadata struct {
	ID          uint           `gorm:"primarykey" json:"id,omitempty"`
	CreatedAt   time.Time      `json:"-"`
	UpdatedAt   time.Time      `json:"-"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`
	SchemaID    uint           `gorm:"uniqueIndex;not null" json:"schema_id,omitempty"`
	Description string         `json:"description"`
	Author      string         `json:"author"`
}

type Tag struct {
	ID        uint           `gorm:"primarykey" json:"id,omitempty"`
	CreatedAt time.Time      `json:"-"`
	UpdatedAt time.Time      `json:"-"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
	SchemaID  uint           `gorm:"not null;index" json:"schema_id,omitempty"`
	Key       string         `gorm:"column:tag_key;not null" json:"key"`
	Value     string         `gorm:"column:tag_value;not null" json:"value"`
}
