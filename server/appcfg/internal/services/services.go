package services

import (
	"context"
	"errors"
	"time"

	"github.com/specterops/bloodhound/cmd/api/src/database/types/null"
)

//go:generate go tool mockery

// DatapipeStatusType represents the current status of the datapipe.
type DatapipeStatusType string

const (
	DatapipeStatusIdle      DatapipeStatusType = "idle"
	DatapipeStatusIngesting DatapipeStatusType = "ingesting"
	DatapipeStatusAnalyzing DatapipeStatusType = "analyzing"
	DatapipeStatusPurging   DatapipeStatusType = "purging"
	DatapipeStatusPruning   DatapipeStatusType = "pruning"
	DatapipeStatusStarting  DatapipeStatusType = "starting"
)

// DatapipeStatus represents the current state of the datapipe.
type DatapipeStatus struct {
	Status                  DatapipeStatusType
	UpdatedAt               time.Time
	LastCompleteAnalysisAt  time.Time
	LastAnalysisRunAt       time.Time
	NextScheduledAnalysisAt null.Time
}

var (
	ErrNotFound = errors.New("not found")
)

type Database interface {
	GetDatapipeStatus(ctx context.Context) (DatapipeStatus, error)
}

type Service struct {
	db Database
}

func NewService(db Database) *Service {
	return &Service{db: db}
}

// GetDatapipeService returns the current datapipe status
func (s *Service) GetDatapipeStatus(ctx context.Context) (DatapipeStatus, error) {
	return s.db.GetDatapipeStatus(ctx)
}
