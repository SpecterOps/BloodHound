package handlers

import (
	"encoding/json"
	"time"

	"github.com/specterops/bloodhound/cmd/api/src/database/types/null"
	"github.com/specterops/bloodhound/server/appcfg/internal/services"
)

type DatapipeStatusView struct {
	Status                  services.DatapipeStatusType `json:"status"`
	UpdatedAt               time.Time                   `json:"updated_at"`
	LastCompleteAnalysisAt  time.Time                   `json:"last_complete_analysis_at"`
	LastAnalysisRunAt       time.Time                   `json:"last_analysis_run_at"`
	NextScheduledAnalysisAt null.Time                   `json:"next_scheduled_analysis_at"`
}

func BuildDatapipeStatusView(status services.DatapipeStatus) DatapipeStatusView {
	return DatapipeStatusView{
		Status:                  status.Status,
		UpdatedAt:               status.UpdatedAt,
		LastCompleteAnalysisAt:  status.LastCompleteAnalysisAt,
		LastAnalysisRunAt:       status.LastAnalysisRunAt,
		NextScheduledAnalysisAt: status.NextScheduledAnalysisAt,
	}
}

// JSONView marshals the view to the byte slice expected by responses.WriteBasic,
// satisfying the responses.JSONViewer contract.
func (s DatapipeStatusView) JSONView() ([]byte, error) {
	return json.Marshal(s)
}
