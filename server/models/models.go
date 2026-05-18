package models

import (
	"time"
)

type RequestedAnalysisType string

const (
	RequestedAnalysisTypeAnalysis RequestedAnalysisType = "analysis"
	RequestedAnalysisTypeDeletion RequestedAnalysisType = "deletion"
)

type RequestedAnalysis struct {
	RequestedBy string
	RequestType RequestedAnalysisType
	RequestedAt time.Time
	// Deletes all nodes and edges in the graph
	DeleteAllGraph bool
	// Deletes all nodes and edges in the graph that have a type not registered in the source_kinds table
	DeleteSourcelessGraph bool
	// Deletes all nodes and edges per kind provided.
	DeleteSourceKinds []string
	// Deletes all relationships by name
	DeleteRelationships []string
}
