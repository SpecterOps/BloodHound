// Copyright 2024 Specter Ops, Inc.
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
	"testing"

	"github.com/gofrs/uuid"
	"github.com/stretchr/testify/assert"
)

type MockAuditable struct {
	Data AuditData
}

func (m MockAuditable) AuditData() AuditData {
	return m.Data
}

func TestMatches_Success(t *testing.T) {
	validEntry1 := AuditEntry{
		CommitID: uuid.Must(uuid.NewV4()),
		Action:   "action",
		Model:    MockAuditable{AuditData{"key": "value"}},
		Status:   "ok",
		ErrorMsg: "err",
	}

	validEntry2 := AuditEntry{
		CommitID: uuid.Must(uuid.NewV4()),
		Action:   "action",
		Model:    MockAuditable{AuditData{"key": "value"}},
		Status:   "ok",
		ErrorMsg: "err",
	}

	assert.True(t, validEntry1.Matches(validEntry2), "Expected validEntry1 to match validEntry2")
}

func TestMatches_FailureAction(t *testing.T) {
	actionEntry1 := AuditEntry{
		CommitID: uuid.Must(uuid.NewV4()),
		Action:   "action",
		Model:    MockAuditable{AuditData{"key": "value"}},
		Status:   "ok",
		ErrorMsg: "err",
	}

	actionEntry2 := AuditEntry{
		CommitID: uuid.Must(uuid.NewV4()),
		Action:   "invalid action",
		Model:    MockAuditable{AuditData{"key": "value"}},
		Status:   "ok",
		ErrorMsg: "err",
	}

	assert.False(t, actionEntry1.Matches(actionEntry2), "Expected actionEntry1 not to match actionEntry2")
}

func TestMatches_FailureModel(t *testing.T) {
	modelEntry1 := AuditEntry{
		CommitID: uuid.Must(uuid.NewV4()),
		Action:   "action",
		Model:    MockAuditable{AuditData{"key": "value"}},
		Status:   "ok",
		ErrorMsg: "err",
	}

	modelEntry2 := AuditEntry{
		CommitID: uuid.Must(uuid.NewV4()),
		Action:   "action",
		Model:    MockAuditable{AuditData{"key2": "value2"}},
		Status:   "ok",
		ErrorMsg: "err",
	}

	assert.False(t, modelEntry1.Matches(modelEntry2), "Expected modelEntry1 not to match modelEntry2")
}

func TestMatches_FailureStatus(t *testing.T) {
	statusEntry1 := AuditEntry{
		CommitID: uuid.Must(uuid.NewV4()),
		Action:   "action",
		Model:    MockAuditable{AuditData{"key": "value"}},
		Status:   "ok",
		ErrorMsg: "err",
	}

	statusEntry2 := AuditEntry{
		CommitID: uuid.Must(uuid.NewV4()),
		Action:   "action",
		Model:    MockAuditable{AuditData{"key": "value"}},
		Status:   "not ok",
		ErrorMsg: "err",
	}

	assert.False(t, statusEntry1.Matches(statusEntry2), "Expected statusEntry1 not to match statusEntry2")
}

func TestMatches_FailureErrMsg(t *testing.T) {
	errMsgEntry1 := AuditEntry{
		CommitID: uuid.Must(uuid.NewV4()),
		Action:   "action",
		Model:    MockAuditable{AuditData{"key": "value"}},
		Status:   "ok",
		ErrorMsg: "err",
	}

	errMsgEntry2 := AuditEntry{
		CommitID: uuid.Must(uuid.NewV4()),
		Action:   "action",
		Model:    MockAuditable{AuditData{"key": "value"}},
		Status:   "ok",
		ErrorMsg: "errors",
	}

	assert.False(t, errMsgEntry1.Matches(errMsgEntry2), "Expected errMsgsEntry1 not to match errMsgEntry2")
}
