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

package apitest

import (
	"fmt"
	"log/slog"
	"os"

	"github.com/specterops/bloodhound/cmd/api/src/api"
	"github.com/specterops/bloodhound/cmd/api/src/api/v2/auth"
	authPkg "github.com/specterops/bloodhound/cmd/api/src/auth"
	"github.com/specterops/bloodhound/cmd/api/src/config"
	"github.com/specterops/bloodhound/cmd/api/src/database/mocks"
	"go.uber.org/mock/gomock"
)

func NewAuthManagementResource(mockCtrl *gomock.Controller) (auth.ManagementResource, *mocks.MockDatabase) {
	cfg, err := config.NewDefaultConfiguration()
	if err != nil {
		slog.Error(fmt.Sprintf("Failed to create default configuration: %v", err))
		os.Exit(1)
	}

	cfg.Crypto.Argon2.NumIterations = 1
	cfg.Crypto.Argon2.NumThreads = 1

	mockDB := mocks.NewMockDatabase(mockCtrl)
	resources := auth.NewManagementResource(cfg, mockDB, authPkg.NewAuthorizer(mockDB), api.NewAuthenticator(cfg, mockDB, mocks.NewMockAuthContextInitializer(mockCtrl)))

	return resources, mockDB
}
