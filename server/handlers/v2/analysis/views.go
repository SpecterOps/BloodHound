package analysis

import (
	"time"

	"github.com/specterops/bloodhound/server/models"
)

type RequestedAnalysisView struct {
	RequestedBy string                       `json:"requested_by"`
	RequestType models.RequestedAnalysisType `json:"request_type"`
	RequestedAt time.Time                    `json:"requested_at"`
	// Deletes all nodes and edges in the graph
	DeleteAllGraph bool `json:"delete_all_graph"`
	// Deletes all nodes and edges in the graph that have a type not registered in the source_kinds table
	DeleteSourcelessGraph bool     `json:"delete_sourceless_graph"`
	DeleteSourceKinds     []string `json:"delete_source_kinds"`
	DeleteRelationships   []string `json:"delete_relationships"`
}

func BuildRequestedAnalysisView(ra models.RequestedAnalysis) RequestedAnalysisView {
	return RequestedAnalysisView{
		RequestedBy:           ra.RequestedBy,
		RequestType:           ra.RequestType,
		RequestedAt:           ra.RequestedAt,
		DeleteAllGraph:        ra.DeleteAllGraph,
		DeleteSourcelessGraph: ra.DeleteSourcelessGraph,
		DeleteSourceKinds:     ra.DeleteSourceKinds,
		DeleteRelationships:   ra.DeleteRelationships,
	}
}
