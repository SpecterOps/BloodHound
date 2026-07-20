package services

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/gofrs/uuid"
)

type Status string

const (
	StatusIntent  Status = "intent"
	StatusSuccess Status = "success"
	StatusFailure Status = "failure"
)

type Source string

const SourceMiddleware Source = "middleware"

// AuditRecord is the persistence-facing representation the Store writes.
type AuditRecord struct {
	Action          string
	ActorID         string
	ActorName       string
	ActorEmail      string
	RequestID       string
	SourceIPAddress string
	Status          Status
	CommitID        uuid.UUID
	Fields          map[string]any
	Source          Source
}

// Entry is the domain input the middleware/public API hands to the service.
type Entry struct {
	Action          string
	ActorID         string
	ActorName       string
	ActorEmail      string
	RequestID       string
	SourceIPAddress string
	Fields          map[string]any
}
type Service struct {
	db Database
}

func NewService(db Database) *Service {
	return &Service{db: db}
}

// Database is the port the audit service requires; appdb.Store implements it.
type Database interface {
	InsertAuditLog(ctx context.Context, record AuditRecord) error
}

// Maintainer is the port the GC daemon requires to manage audit partitions.
// appdb.Store implements it.
type Maintainer interface {
	PreCreateNextPartition(ctx context.Context, asOf time.Time) error
	DropExpiredPartitions(ctx context.Context, asOf time.Time, retentionMonths int) error
}

var ErrInvalidAuditRecord = errors.New("invalid audit record")

var sensitivePatternsLower = []string{
	"password", "secret", "token", "api_key", "apikey", "private_key", "privatekey",
}

func redactSensitiveFields(fields map[string]any) map[string]any {
	if len(fields) == 0 {
		return fields
	}
	redacted := make(map[string]any, len(fields))
	for key, value := range fields {
		keyLower := strings.ToLower(key)
		isSensitive := false
		for _, pattern := range sensitivePatternsLower {
			if strings.Contains(keyLower, pattern) {
				isSensitive = true
				break
			}
		}
		if isSensitive {
			redacted[key] = "[REDACTED]"
		} else {
			redacted[key] = value
		}
	}
	return redacted
}

// Intent writes the pre-execution row synchronously and returns the commit id
// that links it to the eventual result.
func (s *Service) Intent(ctx context.Context, entry Entry) (uuid.UUID, error) {
	var (
		commitID uuid.UUID
		err      error
	)
	commitID, err = uuid.NewV4()
	if err != nil {
		return uuid.UUID{}, fmt.Errorf("generating commit id: %w", err)
	}
	if err = s.db.InsertAuditLog(ctx, s.toRecord(entry, commitID, StatusIntent)); err != nil {
		return commitID, err
	}
	return commitID, nil
}

// Success writes the post-execution success row synchronously.
func (s *Service) Success(ctx context.Context, commitID uuid.UUID, entry Entry) error {
	return s.db.InsertAuditLog(ctx, s.toRecord(entry, commitID, StatusSuccess))
}

// Failure writes the post-execution failure row synchronously.
func (s *Service) Failure(ctx context.Context, commitID uuid.UUID, entry Entry) error {
	return s.db.InsertAuditLog(ctx, s.toRecord(entry, commitID, StatusFailure))
}

func (s *Service) toRecord(entry Entry, commitID uuid.UUID, status Status) AuditRecord {
	return AuditRecord{
		Action:          entry.Action,
		ActorID:         entry.ActorID,
		ActorName:       entry.ActorName,
		ActorEmail:      entry.ActorEmail,
		RequestID:       entry.RequestID,
		SourceIPAddress: entry.SourceIPAddress,
		Status:          status,
		CommitID:        commitID,
		Fields:          redactSensitiveFields(entry.Fields),
		Source:          SourceMiddleware,
	}
}
