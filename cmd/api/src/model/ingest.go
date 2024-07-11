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
	"github.com/specterops/bloodhound/src/database/types/null"
)

type IngestTask struct {
	FileName    string     `json:"file_name"`
	RequestGUID string     `json:"request_guid"`
	TaskID      null.Int64 `json:"task_id"`
	FileType    FileType   `json:"file_type"`

	BigSerial
}

type IngestTasks []IngestTask

type FileType int

const (
	FileTypeInvalid FileType = iota
	FileTypeJson
	FileTypeZip
)
