//
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

//go:build integration
// +build integration

package database_test

import (
	"testing"
)

func TestBloodhoundDB_GetSavedQueriesSharedWithUser(t *testing.T) {
	//var (
	//	testCtx = context.Background()
	//	dbInst  = integration.SetupDB(t)
	//)
	//
	//userOneUUID, err := uuid.NewV4()
	//require.NoError(t, err)
	//
	//userTwoUUID, err := uuid.NewV4()
	//require.NoError(t, err)
	//
	//_, err := dbInst.CreateSavedQuery(testCtx, userOneUUID, "test-query", "", "hello world!")
	//require.NoError(t, err)
	//
	//_, err := dbInst.CreateSavedQu
}

func TestBloodhoundDB_CheckUserHasPermissionToSavedQuery(t *testing.T) {

}
