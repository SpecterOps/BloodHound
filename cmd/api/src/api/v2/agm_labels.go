package v2

import (
	"net/http"
	"slices"

	"github.com/specterops/bloodhound/src/api"
)

func (s Resources) GetAssetGroupLabels(response http.ResponseWriter, request *http.Request) {
	const (
		pnameLabelType     = "type"
		pnameIncludeCounts = "includeCounts"
	)
	var (
		pvalsLabelType     = []string{"label", "tier"}
		pvalsIncludeCounts = []string{"false", "true"}
	)

	var params = request.URL.Query()
	// set defaults. This works since .Get() only retrives the first value
	params.Add(pnameLabelType, pvalsLabelType[0])
	params.Add(pnameIncludeCounts, pvalsIncludeCounts[0])

	if paramLabelType := params.Get(pnameLabelType); !slices.Contains(pvalsLabelType, paramLabelType) {
		api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusBadRequest, "Invalid value specifed for label type", request), response)
	} else if paramIncludeCounts := params.Get(pnameIncludeCounts); !slices.Contains(pvalsIncludeCounts, paramIncludeCounts) {
		api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusBadRequest, "Invalid value specifed for include counts", request), response)
	} else if labels, err := s.DB.GetAssetGroupLabels(request.Context(), paramLabelType); err != nil {
		api.HandleDatabaseError(request, response, err)
	} else {
		resp := map[string]any{"asset_group_labels": labels}
		if paramIncludeCounts == "true" {
		}
		api.WriteBasicResponse(request.Context(), resp, http.StatusOK, response)
	}
}
