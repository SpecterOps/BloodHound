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
	"context"
	"slices"
	"testing"

	"github.com/gofrs/uuid"
	"github.com/specterops/bloodhound/cmd/api/src/database/types/null"
	"github.com/specterops/bloodhound/cmd/api/src/model"
)

func TestCreateAndGetAllIngestJobs(t *testing.T) {
	var (
		testCtx       = context.Background()
		dbInst, roles = initAndGetRoles(t)
		seedUsers     = []model.User{
			{
				Roles:         roles,
				FirstName:     null.StringFrom("First"),
				LastName:      null.StringFrom("Last"),
				EmailAddress:  null.StringFrom(userPrincipal),
				PrincipalName: userPrincipal,
			}, {
				Roles:         roles,
				FirstName:     null.StringFrom("First2"),
				LastName:      null.StringFrom("Last2"),
				EmailAddress:  null.StringFrom(user2Principal),
				PrincipalName: user2Principal,
			}}
		createdUsers []model.User
	)

	for _, user := range seedUsers {
		if newUser, err := dbInst.CreateUser(testCtx, user); err != nil {
			t.Fatalf("Error seeding user: %v", err)
		} else {
			createdUsers = append(createdUsers, newUser)
		}
	}

	// insert 3 jobs,
	//   first 2 with valid user_ids
	//   second with bogus use_email_address
	//   third with just user_email_address
	seedJobs := []model.IngestJob{
		{UserID: uuid.NullUUID{UUID: createdUsers[0].ID, Valid: true}},
		{UserID: uuid.NullUUID{UUID: createdUsers[1].ID, Valid: true}, UserEmailAddress: null.StringFrom("super@fake.email")},
		{UserEmailAddress: null.StringFrom(user2Principal)},
	}

	for _, seed := range seedJobs {
		if _, err := dbInst.CreateIngestJob(testCtx, seed); err != nil {
			t.Fatalf("Error seeding ingest jobs")
		}
	}

	if jobs, _, err := dbInst.GetAllIngestJobs(testCtx, 0, 3, "start_time", model.SQLFilter{}); err != nil {
		t.Fatalf("Failed to get users ingest jobs: %v", err)
	} else {
		if len(jobs) != len(seedJobs) {
			t.Fatalf("expected to retrieve %d ingest jobs, got %d", len(seedJobs), len(jobs))
		}

		for _, job := range jobs {
			// if the job is missing a userID, ensure the user_email_address is intact
			if isValidUuid := job.UserID.Valid; isValidUuid == false {
				if job.UserEmailAddress != seedJobs[2].UserEmailAddress {
					t.Fatalf("Missing valid user_id and user_email_address: %v", job)
				}
			}

			// if the job has a userID, ensure the user_email_address matches the association to the user
			if jobUserId := job.UserID; jobUserId.Valid {
				matchingUserIdx := slices.IndexFunc(createdUsers, func(u model.User) bool {
					return uuid.NullUUID{UUID: u.ID, Valid: true} == jobUserId
				})

				if matchingUserIdx == -1 {
					t.Fatalf("ingest job references unexpected user_id: %v", jobUserId.UUID)
				}

				if userEmail := createdUsers[matchingUserIdx].EmailAddress; userEmail.Valid == false {
					t.Fatalf("Failed to retrieve user email_address: %v", createdUsers[matchingUserIdx])
				} else if jobUserEmailAddress := job.UserEmailAddress; jobUserEmailAddress != userEmail {
					t.Fatalf("Failed to associate user email_address to job user_email_address")
				}
			}

		}
	}
}

// ensure that updating a given IngestJob doesnt insert an empty uuid
