// Copyright 2023 Specter Ops, Inc.
//
// Licensed under the Apache License, Version 2.0
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//
// SPDX-License-Identifier: Apache-2.0

package model

import (
	"fmt"
	"strings"
	"time"

	"github.com/gofrs/uuid"
	"github.com/specterops/bloodhound/src/database/types/null"
)

type FileUploadJob struct {
	UserID           uuid.UUID   `json:"user_id"`
	UserEmailAddress null.String `json:"user_email_address"`
	User             User        `json:"-"`
	Status           JobStatus   `json:"status"`
	StatusMessage    string      `json:"status_message"`
	StartTime        time.Time   `json:"start_time"`
	EndTime          time.Time   `json:"end_time"`
	LastIngest       time.Time   `json:"last_ingest"`
	//DomainResults []DomainCollectionResult `json:"domain_results" gorm:"-"`

	BigSerial
}

type FileUploadJobs []FileUploadJob

func (s FileUploadJobs) IsSortable(column string) bool {
	switch column {
	case "user_id",
		"user_email_address",
		"status",
		"status_message",
		"start_time",
		"end_time",
		"last_ingest",
		"id",
		"created_at",
		"updated_at",
		"deleted_at":
		return true
	default:
		return false
	}
}

func (s FileUploadJobs) ValidFilters() map[string][]FilterOperator {
	return map[string][]FilterOperator{
		"user_id":            {Equals, NotEquals},
		"user_email_address": {Equals, NotEquals},
		"status":             {Equals, NotEquals},
		"status_message":     {Equals, NotEquals},
		"start_time":         {Equals, GreaterThan, GreaterThanOrEquals, LessThan, LessThanOrEquals, NotEquals},
		"end_time":           {Equals, GreaterThan, GreaterThanOrEquals, LessThan, LessThanOrEquals, NotEquals},
		"last_ingest":        {Equals, GreaterThan, GreaterThanOrEquals, LessThan, LessThanOrEquals, NotEquals},
		"id":                 {Equals, GreaterThan, GreaterThanOrEquals, LessThan, LessThanOrEquals, NotEquals},
		"created_at":         {Equals, GreaterThan, GreaterThanOrEquals, LessThan, LessThanOrEquals, NotEquals},
		"updated_at":         {Equals, GreaterThan, GreaterThanOrEquals, LessThan, LessThanOrEquals, NotEquals},
		"deleted_at":         {Equals, GreaterThan, GreaterThanOrEquals, LessThan, LessThanOrEquals, NotEquals},
	}
}

func (s FileUploadJobs) IsString(column string) bool {
	switch column {
	case "status_message", "user_id", "user_email_address":
		return true
	default:
		return false
	}
}

func (s FileUploadJobs) GetFilterableColumns() []string {
	var columns = make([]string, 0)
	for column := range s.ValidFilters() {
		columns = append(columns, column)
	}
	return columns
}

func (s FileUploadJobs) GetValidFilterPredicatesAsStrings(column string) ([]string, error) {
	if predicates, validColumn := s.ValidFilters()[column]; !validColumn {
		return []string{}, fmt.Errorf("the specified column cannot be filtered")
	} else {
		var stringPredicates = make([]string, 0)
		for _, predicate := range predicates {
			stringPredicates = append(stringPredicates, string(predicate))
		}
		return stringPredicates, nil
	}
}

type JobStatus int

const (
	JobStatusInvalid           JobStatus = -1
	JobStatusReady             JobStatus = 0
	JobStatusRunning           JobStatus = 1
	JobStatusComplete          JobStatus = 2
	JobStatusCanceled          JobStatus = 3
	JobStatusTimedOut          JobStatus = 4
	JobStatusFailed            JobStatus = 5
	JobStatusIngesting         JobStatus = 6
	JobStatusAnalyzing         JobStatus = 7
	JobStatusPartiallyComplete JobStatus = 8
)

func allJobStatuses() []JobStatus {
	return []JobStatus{
		JobStatusInvalid,
		JobStatusReady,
		JobStatusRunning,
		JobStatusComplete,
		JobStatusCanceled,
		JobStatusTimedOut,
		JobStatusFailed,
		JobStatusIngesting,
		JobStatusAnalyzing,
		JobStatusPartiallyComplete,
	}
}

func ParseJobStatus(jobStatusStr string) (JobStatus, error) {
	sanitized := strings.ToUpper(jobStatusStr)

	for _, jobStatus := range allJobStatuses() {
		if jobStatus.String() == sanitized {
			return jobStatus, nil
		}
	}

	return JobStatusInvalid, fmt.Errorf("no matching job status for: %s", jobStatusStr)
}

func GetVisibleJobStatuses() []JobStatus {
	return []JobStatus{JobStatusComplete, JobStatusCanceled, JobStatusTimedOut, JobStatusFailed, JobStatusIngesting, JobStatusAnalyzing, JobStatusPartiallyComplete}
}

func (s JobStatus) String() string {
	switch s {
	case JobStatusReady:
		return "READY"

	case JobStatusRunning:
		return "RUNNING"

	case JobStatusComplete:
		return "COMPLETE"

	case JobStatusCanceled:
		return "CANCELED"

	case JobStatusTimedOut:
		return "TIMEDOUT"

	case JobStatusFailed:
		return "FAILED"

	case JobStatusIngesting:
		return "INGESTING"

	case JobStatusAnalyzing:
		return "ANALYZING"

	case JobStatusPartiallyComplete:
		return "PARTIALLYCOMPLETE"

	default:
		return "INVALIDSTATUS"
	}
}

func (s JobStatus) IsValidEndState() error {
	switch s {
	case JobStatusFailed, JobStatusComplete:
		return nil
	default:
		return fmt.Errorf("invalid job end state (%s|%s): %s", JobStatusComplete, JobStatusFailed, s)
	}
}

type DomainCollectionResult struct {
	JobID             int64  `json:"job_id"` // TODO remove this field to enable moving this model to FOSS
	DomainName        string `json:"domain_name"`
	Success           bool   `json:"success"`
	Message           string `json:"message"`
	UserCount         int    `json:"user_count"`
	GroupCount        int    `json:"group_count"`
	ComputerCount     int    `json:"computer_count"`
	GPOCount          int    `json:"gpo_count"`
	OUCount           int    `json:"ou_count"`
	ContainerCount    int    `json:"container_count"`
	AIACACount        int    `json:"aiaca_count" gorm:"column:aiaca_count"`
	RootCACount       int    `json:"rootca_count" gorm:"column:rootca_count"`
	EnterpriseCACount int    `json:"enterpriseca_count" gorm:"column:enterpriseca_count"`
	NTAuthStoreCount  int    `json:"ntauthstore_count" gorm:"column:ntauthstore_count"`
	CertTemplateCount int    `json:"certtemplate_count" gorm:"column:certtemplate_count"`
	DeletedCount      int    `json:"deleted_count"`

	BigSerial
}
