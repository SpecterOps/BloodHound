package v2

import (
	"net/http"

	"github.com/specterops/bloodhound/cmd/api/src/api"
)

type DogtagsResponse struct {
	Data map[string]interface{} `json:"data"`
}

type DogtagEvaluation struct {
	Value    interface{} `json:"value"`
	Reason   string      `json:"reason,omitempty"`
	Variant  string      `json:"variant,omitempty"`
	Metadata interface{} `json:"metadata,omitempty"`
}

func (s Resources) GetDogtags(response http.ResponseWriter, request *http.Request) {
	if s.DogtagsService == nil {
		api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(
			http.StatusServiceUnavailable,
			"Dogtags service not available",
			request,
		), response)
		return
	}

	flags := s.DogtagsService.GetAllFlags(request.Context())

	api.WriteBasicResponse(request.Context(), flags, http.StatusOK, response)
}