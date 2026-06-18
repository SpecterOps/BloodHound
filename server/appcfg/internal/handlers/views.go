package handlers

import (
	"time"

	"github.com/specterops/bloodhound/cmd/api/src/database/types/null"
	"github.com/specterops/bloodhound/server/appcfg/internal/services"
)

type DatapipeStatus struct {
	Status                  services.DatapipeStatusType `json:"status"`
	UpdatedAt               time.Time                   `json:"updated_at"`
	LastCompleteAnalysisAt  time.Time                   `json:"last_complete_analysis_at"`
	LastAnalysisRunAt       time.Time                   `json:"last_analysis_run_at"`
	NextScheduledAnalysisAt null.Time                   `json:"next_scheduled_analysis_at"`
}
