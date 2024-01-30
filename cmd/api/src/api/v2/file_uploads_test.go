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

package v2_test

import (
	"context"
	"database/sql"
	"net/http"
	"testing"

	"github.com/specterops/bloodhound/errors"
	v2 "github.com/specterops/bloodhound/src/api/v2"
	"github.com/specterops/bloodhound/src/api/v2/apitest"
	"github.com/specterops/bloodhound/src/auth"
	"github.com/specterops/bloodhound/src/ctx"
	taskerMocks "github.com/specterops/bloodhound/src/daemons/datapipe/mocks"
	dbMocks "github.com/specterops/bloodhound/src/database/mocks"
	"github.com/specterops/bloodhound/src/database/types/null"
	"github.com/specterops/bloodhound/src/model"
	"go.uber.org/mock/gomock"
)

func setupUser() model.User {
	return model.User{
		FirstName:     null.String{NullString: sql.NullString{String: "John", Valid: true}},
		LastName:      null.String{NullString: sql.NullString{String: "Doe", Valid: true}},
		EmailAddress:  null.String{NullString: sql.NullString{String: "johndoe@gmail.com", Valid: true}},
		PrincipalName: "John",
		AuthTokens:    model.AuthTokens{},
	}
}

func setupUserCtx(user model.User) context.Context {
	return context.WithValue(context.Background(), ctx.ValueKey, &ctx.Context{
		AuthCtx: auth.Context{
			PermissionOverrides: auth.PermissionOverrides{},
			Owner:               user,
			Session:             model.UserSession{},
		},
	})
}

func TestResources_ListFileUploadJobs(t *testing.T) {
	var (
		mockCtrl   = gomock.NewController(t)
		mockDB     = dbMocks.NewMockDatabase(mockCtrl)
		mockTasker = taskerMocks.NewMockTasker(mockCtrl)
		resources  = v2.Resources{DB: mockDB, TaskNotifier: mockTasker}
	)
	defer mockCtrl.Finish()

	apitest.
		NewHarness(t, resources.ListFileUploadJobs).
		Run([]apitest.Case{
			apitest.NewSortingErrorCase(),
			apitest.NewColumnNotFilterableCase(),
			apitest.NewInvalidFilterPredicateCase("id"),
			{
				Name: "GetAllFileUploadJobsDatabaseError",
				Setup: func() {
					mockDB.EXPECT().GetAllFileUploadJobs(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, 0, errors.New("database error"))
				},
				Test: func(output apitest.Output) {
					apitest.StatusCode(output, http.StatusInternalServerError)
				},
			},
			{
				Name: "Success",
				Input: func(input *apitest.Input) {
					apitest.AddQueryParam(input, "skip", "1")
					apitest.AddQueryParam(input, "limit", "2")
					apitest.AddQueryParam(input, "sort_by", "start_time")
					apitest.AddQueryParam(input, "user_id", "eq:123")
				},
				Setup: func() {
					mockDB.EXPECT().GetAllFileUploadJobs(1, 2, "start_time", model.SQLFilter{SQLString: "user_id = ?", Params: []any{"123"}}).Return([]model.FileUploadJob{}, 0, nil)
				},
				Test: func(output apitest.Output) {
					apitest.StatusCode(output, http.StatusOK)
				},
			},
		})

}

func TestResources_StartFileUploadJob(t *testing.T) {
	var (
		mockCtrl  = gomock.NewController(t)
		mockDB    = dbMocks.NewMockDatabase(mockCtrl)
		resources = v2.Resources{DB: mockDB}
		user      = setupUser()
		userCtx   = setupUserCtx(user)
	)
	defer mockCtrl.Finish()

	apitest.
		NewHarness(t, resources.StartFileUploadJob).
		Run([]apitest.Case{
			{
				Name: "Unauthorized",
				Test: func(output apitest.Output) {
					apitest.StatusCode(output, http.StatusUnauthorized)
				},
			},
			{
				Name: "DatabaseError",
				Input: func(input *apitest.Input) {
					apitest.SetContext(input, userCtx)
				},
				Setup: func() {
					mockDB.EXPECT().CreateFileUploadJob(gomock.Any()).Return(model.FileUploadJob{}, errors.New("db error"))
				},
				Test: func(output apitest.Output) {
					apitest.StatusCode(output, http.StatusInternalServerError)
				},
			},
			{
				Name: "Success",
				Input: func(input *apitest.Input) {
					apitest.SetContext(input, userCtx)
				},
				Setup: func() {
					mockDB.EXPECT().CreateFileUploadJob(gomock.Any()).Return(model.FileUploadJob{}, nil)
				},
				Test: func(output apitest.Output) {
					apitest.StatusCode(output, http.StatusCreated)
				},
			},
		})
}

func TestResources_EndFileUploadJob(t *testing.T) {
	var (
		mockCtrl   = gomock.NewController(t)
		mockDB     = dbMocks.NewMockDatabase(mockCtrl)
		mockTasker = taskerMocks.NewMockTasker(mockCtrl)
		resources  = v2.Resources{DB: mockDB, TaskNotifier: mockTasker}
	)
	defer mockCtrl.Finish()

	apitest.
		NewHarness(t, resources.EndFileUploadJob).
		Run([]apitest.Case{
			{
				Name: "InvalidJobID",
				Input: func(input *apitest.Input) {
					apitest.SetURLVar(input, v2.FileUploadJobIdPathParameterName, "invalid")
				},
				Test: func(output apitest.Output) {
					apitest.StatusCode(output, http.StatusBadRequest)
				},
			},
			{
				Name: "GetFileUploadJobDatabaseError",
				Input: func(input *apitest.Input) {
					apitest.SetURLVar(input, v2.FileUploadJobIdPathParameterName, "123")
				},
				Setup: func() {
					mockDB.EXPECT().GetFileUploadJob(gomock.Any()).Return(model.FileUploadJob{}, errors.New("db error"))
				},
				Test: func(output apitest.Output) {
					apitest.StatusCode(output, http.StatusInternalServerError)
				},
			},
			{
				Name: "InvalidJobStatus",
				Input: func(input *apitest.Input) {
					apitest.SetURLVar(input, v2.FileUploadJobIdPathParameterName, "123")
				},
				Setup: func() {
					mockDB.EXPECT().GetFileUploadJob(gomock.Any()).Return(model.FileUploadJob{
						Status: model.JobStatusComplete,
					}, nil)
				},
				Test: func(output apitest.Output) {
					apitest.StatusCode(output, http.StatusBadRequest)
					apitest.BodyContains(output, "job must be in running status")
				},
			},
			{
				Name: "UpdateFileUploadJobDatabaseError",
				Input: func(input *apitest.Input) {
					apitest.SetURLVar(input, v2.FileUploadJobIdPathParameterName, "123")
				},
				Setup: func() {
					mockDB.EXPECT().GetFileUploadJob(gomock.Any()).Return(model.FileUploadJob{
						Status: model.JobStatusRunning,
					}, nil)
					mockDB.EXPECT().UpdateFileUploadJob(gomock.Any()).Return(errors.New("database error"))
				},
				Test: func(output apitest.Output) {
					apitest.StatusCode(output, http.StatusInternalServerError)
				},
			},
			{
				Name: "Success",
				Input: func(input *apitest.Input) {
					apitest.SetURLVar(input, v2.FileUploadJobIdPathParameterName, "123")
				},
				Setup: func() {
					mockDB.EXPECT().GetFileUploadJob(gomock.Any()).Return(model.FileUploadJob{
						Status: model.JobStatusRunning,
					}, nil)
					mockDB.EXPECT().UpdateFileUploadJob(gomock.Any()).Return(nil)
				},
				Test: func(output apitest.Output) {
					apitest.StatusCode(output, http.StatusOK)
				},
			},
		})
}
