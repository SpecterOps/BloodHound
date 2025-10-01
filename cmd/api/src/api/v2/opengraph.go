package v2

import (
	"encoding/json"
	"net/http"

	"github.com/specterops/bloodhound/cmd/api/src/api"
	"github.com/specterops/bloodhound/packages/go/graphify/graph"
)

func (s *Resources) ExportGraph(response http.ResponseWriter, request *http.Request) {
	if nodes, edges, err := graph.GetNodesAndEdges(request.Context(), s.Graph); err != nil {
		api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusInternalServerError, api.ErrorResponseDetailsInternalServerError, request), response)
	} else if ogGraph, err := graph.TransformGraph(nodes, edges); err != nil {
		api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusInternalServerError, api.ErrorResponseDetailsInternalServerError, request), response)
	} else if b, err := json.MarshalIndent(graph.GenerateGenericGraphFile(&ogGraph), "", "  "); err != nil {
		api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusInternalServerError, api.ErrorResponseDetailsInternalServerError, request), response)
	} else {
		api.WriteBinaryResponse(request.Context(), b, "og_graph.json", http.StatusOK, response)
	}
}
