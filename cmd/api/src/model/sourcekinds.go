package model

import "github.com/specterops/dawgs/graph"

type SourceKind struct {
	ID     int        `json:"id"`
	Name   graph.Kind `json:"name"`
	Active bool       `json:"active"`
}

func (s SourceKind) TableName() string {
	return "source_kinds"
}
