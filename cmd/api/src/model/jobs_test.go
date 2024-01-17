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
	"github.com/stretchr/testify/require"
	"testing"
)

func TestFileUploadJobs_IsSortable(t *testing.T) {
	fuj := FileUploadJobs{}
	require.True(t, fuj.IsSortable("user_id"))
	require.True(t, fuj.IsSortable("status"))
	require.True(t, fuj.IsSortable("user_email_address"))
	require.True(t, fuj.IsSortable("status_message"))
	require.True(t, fuj.IsSortable("start_time"))
	require.True(t, fuj.IsSortable("end_time"))
	require.True(t, fuj.IsSortable("last_ingest"))
	require.True(t, fuj.IsSortable("id"))
	require.True(t, fuj.IsSortable("created_at"))
	require.True(t, fuj.IsSortable("updated_at"))
	require.True(t, fuj.IsSortable("deleted_at"))
	require.False(t, fuj.IsSortable("foobar"))
}

func TestFileUploadJobs_ValidFilters(t *testing.T) {
	fuj := FileUploadJobs{}
	columns := fuj.ValidFilters()
	require.Equal(t, 11, len(columns))
}
