package v2

import (
	"net/http"

	"github.com/specterops/bloodhound/cmd/api/src/api"
	"github.com/specterops/bloodhound/packages/go/graphify/graph"
)

func (s *Resources) ExportGraph(response http.ResponseWriter, request *http.Request) {
	if nodes, edges, err := graph.GetNodesAndEdges(request.Context(), s.Graph); err != nil {
		api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusInternalServerError, api.ErrorResponseDetailsInternalServerError, request), response)
	} else if ogGraph, err := graph.TransformGraph(nodes, edges); err != nil {
		api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusInternalServerError, api.ErrorResponseDetailsInternalServerError, request), response)
	} else {
		api.WriteJSONResponse(request.Context(), ogGraph, http.StatusOK, response)
	}
}
