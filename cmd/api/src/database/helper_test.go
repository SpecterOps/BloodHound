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

package database_test

import (
	"testing"

	"github.com/specterops/bloodhound/src/database"
	"gorm.io/gorm"

	"github.com/gofrs/uuid"
	"github.com/stretchr/testify/require"
)

func TestDatabase_checkError_NotFound(t *testing.T) {
	errType := gorm.ErrRecordNotFound
	tx := gorm.DB{
		Error: errType,
	}

	err := database.CheckError(&tx)
	require.NotNil(t, err)

	// gorm.errRecordNotFound should be converted to ErrNotFound
	require.Equal(t, database.ErrNotFound, err)
}

func TestDatabase_checkError_OtherError(t *testing.T) {
	errType := gorm.ErrInvalidData
	tx := gorm.DB{
		Error: errType,
	}

	err := database.CheckError(&tx)
	require.NotNil(t, err)

	// error should be returned as received
	require.Equal(t, errType, err)
}

func TestDatabase_NullUUID(t *testing.T) {
	value, err := uuid.NewV4()
	require.Nil(t, err)

	output := database.NullUUID(value)
	require.NotNil(t, output.UUID)
	require.True(t, output.Valid)
}
