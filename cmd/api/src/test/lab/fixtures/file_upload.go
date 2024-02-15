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

package fixtures

import (
	"fmt"
	"log"

	"github.com/specterops/bloodhound/lab"
	"github.com/specterops/bloodhound/src/database"
	"github.com/specterops/bloodhound/src/model"
)

func NewFileUploadFixture(db *lab.Fixture[*database.BloodhoundDB], userFixture *lab.Fixture[*model.User]) *lab.Fixture[*model.FileUploadJobs] {

	fixture := lab.NewFixture(
		func(h *lab.Harness) (*model.FileUploadJobs, error) {
			if user, ok := lab.Unpack(h, userFixture); !ok {
				return nil, fmt.Errorf("unable to unpack UserFixture")
			} else if db, ok := lab.Unpack(h, PostgresFixture); !ok {
				return nil, fmt.Errorf("unable to unpack BloodhoundDB")
			} else {
				var fileUploadJobs model.FileUploadJobs

				// create 3 file upload jobs
				for i := 0; i < 3; i++ {
					if job, err := db.CreateFileUploadJob(model.FileUploadJob{User: *user, UserID: user.ID}); err != nil {
						return nil, fmt.Errorf("unable to create file upload: %w", err)
					} else {
						fileUploadJobs = append(fileUploadJobs, job)
					}
				}

				return &fileUploadJobs, nil
			}
		},
		func(h *lab.Harness, fileUploadJobs *model.FileUploadJobs) error {
			if db, ok := lab.Unpack(h, PostgresFixture); !ok {
				return fmt.Errorf("unable to unpack BloodhoundDB")
			} else {
				if err := db.DeleteAllFileUploads(); err != nil {
					return fmt.Errorf("failure deleting file uploads: %w", err)
				}
			}
			return nil
		},
	)

	if err := lab.SetDependency(fixture, userFixture); err != nil {
		log.Fatalf("FileUploadFixture dependency error: %v", err)
	} else if err := lab.SetDependency(fixture, db); err != nil {
		log.Fatalf("FileUploadFixture dependency error: %v", err)
	} else if err := lab.SetDependency(fixture, BHAdminApiClientFixture); err != nil {
		log.Fatalf("FileUploadFixture dependency error: %v", err)
	}

	return fixture
}
