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

package v2

import (
	"archive/zip"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/gofrs/uuid"
	"github.com/gorilla/mux"
	"gorm.io/gorm/utils"

	"github.com/specterops/bloodhound/headers"
	"github.com/specterops/bloodhound/mediatypes"
	"github.com/specterops/bloodhound/src/api"
	"github.com/specterops/bloodhound/src/auth"
	ctx2 "github.com/specterops/bloodhound/src/ctx"
	"github.com/specterops/bloodhound/src/database"
	"github.com/specterops/bloodhound/src/model"
	"github.com/specterops/bloodhound/src/model/ingest"
	"github.com/specterops/bloodhound/src/services/upload"
	bhUtils "github.com/specterops/bloodhound/src/utils"
)

//gocyclo:ignore
func (s Resources) ListSavedQueries(response http.ResponseWriter, request *http.Request) {
	var (
		order         []string
		queryParams   = request.URL.Query()
		sortByColumns = queryParams[api.QueryParameterSortBy]
		savedQueries  model.SavedQueries
		scopes        = queryParams[api.QueryParameterScope]
	)

	for _, column := range sortByColumns {
		var descending bool
		if string(column[0]) == "-" {
			descending = true
			column = column[1:]
		}

		if !savedQueries.IsSortable(column) {
			api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusBadRequest, api.ErrorResponseDetailsNotSortable, request), response)
			return
		}

		if descending {
			order = append(order, column+" desc")
		} else {
			order = append(order, column)
		}

	}

	queryParameterFilterParser := model.NewQueryParameterFilterParser()
	if queryFilters, err := queryParameterFilterParser.ParseQueryParameterFilters(request); err != nil {
		api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusBadRequest, api.ErrorResponseDetailsBadQueryParameterFilters, request), response)
		return
	} else {
		for name, filters := range queryFilters {
			if validPredicates, err := savedQueries.GetValidFilterPredicatesAsStrings(name); err != nil {
				api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusBadRequest, fmt.Sprintf("%s: %s", api.ErrorResponseDetailsColumnNotFilterable, name), request), response)
				return
			} else {
				for i, filter := range filters {
					if !utils.Contains(validPredicates, string(filter.Operator)) {
						api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusBadRequest, fmt.Sprintf("%s: %s %s", api.ErrorResponseDetailsFilterPredicateNotSupported, filter.Name, filter.Operator), request), response)
						return
					}
					queryFilters[name][i].IsStringData = savedQueries.IsString(filter.Name)
				}
			}
		}

		if user, isUser := auth.GetUserFromAuthCtx(ctx2.FromRequest(request).AuthCtx); !isUser {
			api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusBadRequest, "No associated user found", request), response)
		} else if sqlFilter, err := queryFilters.BuildSQLFilter(); err != nil {
			api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusBadRequest, "error building SQL for filter", request), response)
		} else if skip, err := ParseSkipQueryParameter(queryParams, 0); err != nil {
			api.WriteErrorResponse(request.Context(), ErrBadQueryParameter(request, model.PaginationQueryParameterSkip, err), response)
		} else if limit, err := ParseLimitQueryParameter(queryParams, 10000); err != nil {
			api.WriteErrorResponse(request.Context(), ErrBadQueryParameter(request, model.PaginationQueryParameterLimit, err), response)
		} else if len(scopes) == 0 {
			if queries, count, err := s.DB.ListSavedQueries(request.Context(), user.ID, strings.Join(order, ", "), sqlFilter, skip, limit); err != nil {
				api.HandleDatabaseError(request, response, err)
			} else {
				api.WriteResponseWrapperWithPagination(request.Context(), queries, limit, skip, count, http.StatusOK, response)
			}
		} else {
			var queries []model.SavedQueryResponse
			var count int
			for _, scope := range strings.Split(scopes[0], ",") {
				var scopedQueries model.SavedQueries
				var scopedCount int

				switch strings.ToLower(scope) {
				case string(model.SavedQueryScopePublic):
					scopedQueries, err = s.DB.GetPublicSavedQueries(request.Context())
					scopedCount = len(scopedQueries)
				case string(model.SavedQueryScopeShared):
					scopedQueries, err = s.DB.GetSharedSavedQueries(request.Context(), user.ID)
					scopedCount = len(scopedQueries)
				case string(model.SavedQueryScopeOwned):
					scopedQueries, scopedCount, err = s.DB.ListSavedQueries(request.Context(), user.ID, strings.Join(order, ", "), sqlFilter, skip, limit)
				default:
					api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusBadRequest, "invalid scope param", request), response)
					return
				}

				if err != nil {
					api.HandleDatabaseError(request, response, err)
					return
				}

				for _, query := range scopedQueries {
					queries = append(queries, model.SavedQueryResponse{
						SavedQuery: query,
						Scope:      scope,
					})
				}
				count += scopedCount

			}
			api.WriteResponseWrapperWithPagination(request.Context(), queries, limit, skip, count, http.StatusOK, response)
		}
	}

}

// TransferableSavedQuery - Used for importing/exporting saved queries
type TransferableSavedQuery struct {
	Query       string `json:"query"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

// ExportSavedQuery - Returns the saved query as a json file using the saved query's name as the filename.
// Admins can share any public query regardless of user ownership.
func (s Resources) ExportSavedQuery(response http.ResponseWriter, request *http.Request) {
	var (
		rawSavedQueryID    = mux.Vars(request)[api.URIPathVariableSavedQueryID]
		savedQuery         model.SavedQuery
		auditLogEntry      model.AuditEntry
		err                error
		savedQueryID       int64
		isAccessibleToUser bool
		data               []byte
	)

	// defer audit function
	defer func() {
		if auditLogEntry.Status == model.AuditLogStatusFailure || auditLogEntry.Status == model.AuditLogStatusSuccess {
			if err != nil {
				auditLogEntry.ErrorMsg = err.Error()
			}
			if err = s.DB.AppendAuditLog(request.Context(), auditLogEntry); err != nil {
				if errors.Is(err, database.ErrNotFound) {
					slog.ErrorContext(request.Context(), fmt.Sprintf("resource not found: %v", err))
				} else if errors.Is(err, context.DeadlineExceeded) {
					slog.ErrorContext(request.Context(), fmt.Sprintf("context deadline exceeded: %v", err))
				} else {
					slog.ErrorContext(request.Context(), fmt.Sprintf("unexpected database error: %v", err))
				}
			}
		}
		// did not make it far enough in the api request for an audit log event
	}()

	if user, isUser := auth.GetUserFromAuthCtx(ctx2.FromRequest(request).AuthCtx); !isUser {
		api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusBadRequest, "No associated user found", request), response)
	} else if savedQueryID, err = strconv.ParseInt(rawSavedQueryID, 10, 64); err != nil {
		api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusBadRequest, api.ErrorResponseDetailsIDMalformed, request), response)
	} else if auditLogEntry, err = model.NewAuditEntry(model.AuditLogActionExportSavedQuery, model.AuditLogStatusIntent, model.AuditData{"target_query_id": savedQueryID, "user_id": user.ID.String()}); err != nil {
		api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusInternalServerError, api.ErrorResponseDetailsInternalServerError, request), response)
	} else if err = s.DB.AppendAuditLog(request.Context(), auditLogEntry); err != nil {
		api.HandleDatabaseError(request, response, err)
	} else if savedQuery, err = s.DB.GetSavedQuery(request.Context(), savedQueryID); err != nil {
		auditLogEntry.Status = model.AuditLogStatusFailure
		api.HandleDatabaseError(request, response, err)
	} else if isAccessibleToUser, err = s.canUserAccessQuery(request.Context(), savedQuery, user.ID); err != nil {
		auditLogEntry.Status = model.AuditLogStatusFailure
		api.HandleDatabaseError(request, response, err)
	} else if !isAccessibleToUser {
		err = fmt.Errorf("query does not exist")
		auditLogEntry.Status = model.AuditLogStatusFailure
		api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusNotFound, err.Error(), request), response)
	} else if data, err = api.ToJSONRawMessage(TransferableSavedQuery{Query: savedQuery.Query, Name: savedQuery.Name, Description: savedQuery.Description}); err != nil {
		auditLogEntry.Status = model.AuditLogStatusFailure
		api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusInternalServerError, api.ErrorResponseDetailsInternalServerError, request), response)
	} else {
		auditLogEntry.Status = model.AuditLogStatusSuccess
		api.WriteBinaryResponse(request.Context(), data, fmt.Sprintf("%s.json", filepath.Base(savedQuery.Name)), http.StatusOK, response)
	}
}

func (s Resources) canUserAccessQuery(ctx context.Context, query model.SavedQuery, userId uuid.UUID) (bool, error) {
	if userId.String() == query.UserID {
		return true, nil
	}
	return s.DB.IsSavedQuerySharedToUserOrPublic(ctx, query.ID, userId)
}

// ExportSavedQueries - Exports one or more saved queries in a ZIP file. The scope query parameter determines which queries are exported.
// Only the first scope query parameter will be considered.
func (s Resources) ExportSavedQueries(response http.ResponseWriter, request *http.Request) {
	var (
		auditLogEntry = model.AuditEntry{}
		err           error
		queryParams   = request.URL.Query()
		scope         = queryParams.Get(api.QueryParameterScope)
		savedQueries  model.SavedQueries
		zipBytes      []byte
	)

	defer func() {
		if auditLogEntry.Status == model.AuditLogStatusFailure || auditLogEntry.Status == model.AuditLogStatusSuccess {
			if err != nil {
				auditLogEntry.ErrorMsg = err.Error()
			}
			if err = s.DB.AppendAuditLog(request.Context(), auditLogEntry); err != nil {
				if errors.Is(err, database.ErrNotFound) {
					slog.ErrorContext(request.Context(), fmt.Sprintf("resource not found: %v", err))
				} else if errors.Is(err, context.DeadlineExceeded) {
					slog.ErrorContext(request.Context(), fmt.Sprintf("context deadline exceeded: %v", err))
				} else {
					slog.ErrorContext(request.Context(), fmt.Sprintf("unexpected database error: %v", err))
				}
			}
		}
		// did not make it far enough in the api request for an audit log event
	}()

	if user, isUser := auth.GetUserFromAuthCtx(ctx2.FromRequest(request).AuthCtx); !isUser {
		api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusBadRequest, "No associated user found", request), response)
	} else if auditLogEntry, err = model.NewAuditEntry(model.AuditLogActionExportSavedQueries, model.AuditLogStatusIntent, model.AuditData{"export_saved_queries_scope": scope, "user_id": user.ID.String()}); err != nil {
		api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusInternalServerError, api.ErrorResponseDetailsInternalServerError, request), response)
	} else if err = s.DB.AppendAuditLog(request.Context(), auditLogEntry); err != nil {
		api.HandleDatabaseError(request, response, err)
	} else if scope == "" {
		auditLogEntry.Status = model.AuditLogStatusFailure
		err = fmt.Errorf("scope query parameter cannot be empty")
		api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusBadRequest, err.Error(), request), response)
	} else if savedQueries, err = s.getSavedQueriesByUserAndScope(request.Context(), user.ID, scope); err != nil {
		auditLogEntry.Status = model.AuditLogStatusFailure
		if strings.Contains(err.Error(), "invalid scope param") {
			api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusBadRequest, err.Error(), request), response)
		} else {
			api.HandleDatabaseError(request, response, err)
		}
	} else if zipBytes, err = createSavedQueriesZipFile(savedQueries); err != nil {
		auditLogEntry.Status = model.AuditLogStatusFailure
		api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusInternalServerError, api.ErrorResponseDetailsInternalServerError, request), response)
	} else {
		auditLogEntry.Status = model.AuditLogStatusSuccess
		api.WriteBinaryResponse(request.Context(), zipBytes, "exported_queries.zip", http.StatusOK, response)
	}
}

func (s Resources) getSavedQueriesByUserAndScope(ctx context.Context, userId uuid.UUID, scope string) (model.SavedQueries, error) {
	var (
		savedQueries model.SavedQueries
		err          error
	)

	switch strings.ToLower(scope) {
	case string(model.SavedQueryScopeAll):
		savedQueries, err = s.DB.GetAllSavedQueriesByUser(ctx, userId)
	case string(model.SavedQueryScopePublic):
		savedQueries, err = s.DB.GetPublicSavedQueries(ctx)
	case string(model.SavedQueryScopeShared):
		savedQueries, err = s.DB.GetSharedSavedQueries(ctx, userId)
	case string(model.SavedQueryScopeOwned):
		savedQueries, _, err = s.DB.ListSavedQueries(ctx, userId, "id", model.SQLFilter{}, 0, 0)
	default:
		return nil, fmt.Errorf("invalid scope param: %s", scope)
	}
	return savedQueries, err
}

func createSavedQueriesZipFile(savedQueries []model.SavedQuery) ([]byte, error) {
	var err error
	zipBuffer := new(bytes.Buffer)
	zipWriter := zip.NewWriter(zipBuffer)
	for _, query := range savedQueries {
		var (
			file        io.Writer
			jsonBytes   []byte
			exportQuery = TransferableSavedQuery{
				Query:       query.Query,
				Name:        query.Name,
				Description: query.Description,
			}
		)

		if file, err = zipWriter.Create(fmt.Sprintf("%s.json", filepath.Base(exportQuery.Name))); err != nil {
			return nil, err
		} else if jsonBytes, err = json.Marshal(exportQuery); err != nil {
			return nil, err
		} else if _, err = io.Copy(file, bytes.NewReader(jsonBytes)); err != nil {
			return nil, err
		}
	}
	if err = zipWriter.Close(); err != nil {
		return nil, err
	}
	return zipBuffer.Bytes(), nil
}

// ImportSavedQueries - Used to import custom cypher queries.
// Can import a single query in a json file, or a zip file containing multiple json files.
func (s Resources) ImportSavedQueries(response http.ResponseWriter, request *http.Request) {

	var (
		extractQueriesFromFileFunc func(userId uuid.UUID, file io.Reader) (model.SavedQueries, error)
		auditLogEntry              model.AuditEntry
		err                        error
		savedQueries               model.SavedQueries
	)

	// defer audit function
	defer func() {
		if auditLogEntry.Status == model.AuditLogStatusFailure || auditLogEntry.Status == model.AuditLogStatusSuccess {
			if err != nil {
				auditLogEntry.ErrorMsg = err.Error()
			}
			if err = s.DB.AppendAuditLog(request.Context(), auditLogEntry); err != nil {
				if errors.Is(err, database.ErrNotFound) {
					slog.ErrorContext(request.Context(), fmt.Sprintf("resource not found: %v", err))
				} else if errors.Is(err, context.DeadlineExceeded) {
					slog.ErrorContext(request.Context(), fmt.Sprintf("context deadline exceeded: %v", err))
				} else {
					slog.ErrorContext(request.Context(), fmt.Sprintf("unexpected database error: %v", err))
				}
			}
		}
		// did not make it far enough in the api request for an audit log event
	}()

	if user, isUser := auth.GetUserFromAuthCtx(ctx2.FromRequest(request).AuthCtx); !isUser {
		api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusBadRequest, "No associated user found", request), response)
	} else if auditLogEntry, err = model.NewAuditEntry(model.AuditLogActionImportSavedQuery, model.AuditLogStatusIntent, model.AuditData{"user_id": user.ID}); err != nil {
		api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusInternalServerError, api.ErrorResponseDetailsInternalServerError, request), response)
	} else if err = s.DB.AppendAuditLog(request.Context(), auditLogEntry); err != nil {
		api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusInternalServerError, api.ErrorResponseDetailsInternalServerError, request), response)
	} else if request.Body == nil {
		err = fmt.Errorf("import cypher query request body cannot be empty")
		auditLogEntry.Status = model.AuditLogStatusFailure
		api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusBadRequest, err.Error(), request), response)
	} else {
		request.Body = http.MaxBytesReader(response, request.Body, api.DefaultAPIPayloadReadLimitBytes)
		defer request.Body.Close()
		switch {
		case bhUtils.HeaderMatches(request.Header, headers.ContentType.String(), mediatypes.ApplicationJson.String()):
			extractQueriesFromFileFunc = extractImportQueriesFromJsonFile
		case bhUtils.HeaderMatches(request.Header, headers.ContentType.String(), ingest.AllowedZipFileUploadTypes...):
			extractQueriesFromFileFunc = extractImportQueriesFromZipFile
		default:
			err = fmt.Errorf("invalid content-type: %s", request.Header[headers.ContentType.String()])
			auditLogEntry.Status = model.AuditLogStatusFailure
			api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusUnsupportedMediaType, fmt.Sprintf("%s; Content type must be application/json or application/zip", err.Error()), request), response)
			return
		}
		if savedQueries, err = extractQueriesFromFileFunc(user.ID, request.Body); err != nil {
			auditLogEntry.Status = model.AuditLogStatusFailure
			switch {
			case strings.Contains(err.Error(), "failed to unmarshal json file"):
				api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusBadRequest, err.Error(), request), response)
			case strings.Contains(err.Error(), "error during zip validation") || strings.Contains(err.Error(), "not a valid zip file"):
				api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusBadRequest, err.Error(), request), response)
			default:
				api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusInternalServerError, api.ErrorResponseDetailsInternalServerError, request), response)
			}
			return
		} else if err = s.DB.CreateSavedQueries(request.Context(), savedQueries); err != nil {
			auditLogEntry.Status = model.AuditLogStatusFailure
			if strings.Contains(err.Error(), "duplicate key value violates unique constraint") {
				api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusBadRequest, "duplicate name for saved query: please choose a different name", request), response)
			} else {
				api.HandleDatabaseError(request, response, err)
			}
		} else {
			api.WriteBasicResponse(request.Context(), fmt.Sprintf("imported %d queries", len(savedQueries)), http.StatusCreated, response)
			auditLogEntry.Status = model.AuditLogStatusSuccess
		}
	}
}

func extractImportQueriesFromJsonFile(userId uuid.UUID, file io.Reader) (model.SavedQueries, error) {
	var (
		savedQueries = make(model.SavedQueries, 0)
		query        TransferableSavedQuery
	)
	if jsonQueryFile, err := io.ReadAll(file); err != nil {
		return savedQueries, err
	} else if err = json.Unmarshal(jsonQueryFile, &query); err != nil {
		return savedQueries, fmt.Errorf("failed to unmarshal json file: %w", err)
	} else {
		savedQueries = append(savedQueries, model.SavedQuery{
			UserID:      userId.String(),
			Name:        query.Name,
			Query:       query.Query,
			Description: query.Description,
		})
	}
	return savedQueries, nil
}

func extractImportQueriesFromZipFile(userId uuid.UUID, zipFile io.Reader) (model.SavedQueries, error) {
	if zipFileBytes, err := io.ReadAll(zipFile); err != nil {
		return model.SavedQueries{}, err
	} else if zipReader, err := zip.NewReader(bytes.NewReader(zipFileBytes), int64(len(zipFileBytes))); err != nil {
		return model.SavedQueries{}, err
	} else {
		queries := make(model.SavedQueries, 0)
		for _, zipQueryFile := range zipReader.File {
			// OSX will zip hidden files which we don't want to process
			if strings.Contains(zipQueryFile.Name, "__MACOSX") || strings.HasPrefix(zipQueryFile.Name, ".") {
				continue
			}
			if jsonQueryFile, err := upload.ReadZippedFile(zipQueryFile); err != nil {
				return queries, err
			} else {
				var importQuery TransferableSavedQuery
				if err = json.Unmarshal(jsonQueryFile, &importQuery); err != nil {
					return queries, fmt.Errorf("failed to unmarshal json file: %w", err)
				}
				queries = append(queries, model.SavedQuery{
					Query:       importQuery.Query,
					Name:        importQuery.Name,
					UserID:      userId.String(),
					Description: importQuery.Description,
				})
			}
		}
		return queries, nil
	}
}

type CreateSavedQueryRequest struct {
	Query       string `json:"query"`
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
}

func (s Resources) CreateSavedQuery(response http.ResponseWriter, request *http.Request) {
	var (
		createRequest CreateSavedQueryRequest
	)

	if user, isUser := auth.GetUserFromAuthCtx(ctx2.FromRequest(request).AuthCtx); !isUser {
		api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusBadRequest, "No associated user found", request), response)
	} else if err := api.ReadJSONRequestPayloadLimited(&createRequest, request); err != nil {
		api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusBadRequest, err.Error(), request), response)
	} else if createRequest.Name == "" || createRequest.Query == "" {
		api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusBadRequest, "the name and/or query field is empty", request), response)
	} else if savedQuery, err := s.DB.CreateSavedQuery(request.Context(), user.ID, createRequest.Name, createRequest.Query, createRequest.Description); err != nil {
		if strings.Contains(err.Error(), "duplicate key value violates unique constraint") {
			api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusBadRequest, "duplicate name for saved query: please choose a different name", request), response)
		} else {
			api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusInternalServerError, err.Error(), request), response)
		}
	} else {
		api.WriteBasicResponse(request.Context(), savedQuery, http.StatusCreated, response)
	}
}

func (s Resources) UpdateSavedQuery(response http.ResponseWriter, request *http.Request) {
	var (
		rawSavedQueryID = mux.Vars(request)[api.URIPathVariableSavedQueryID]
		updateRequest   CreateSavedQueryRequest
		savedQuery      model.SavedQuery
		err             error
	)

	if user, isUser := auth.GetUserFromAuthCtx(ctx2.FromRequest(request).AuthCtx); !isUser {
		api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusBadRequest, "No associated user found", request), response)
		return
	} else if err := api.ReadJSONRequestPayloadLimited(&updateRequest, request); err != nil {
		api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusBadRequest, err.Error(), request), response)
		return
	} else if savedQueryID, err := strconv.ParseInt(rawSavedQueryID, 10, 64); err != nil {
		api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusBadRequest, api.ErrorResponseDetailsIDMalformed, request), response)
		return
	} else if savedQuery, err = s.DB.GetSavedQuery(request.Context(), savedQueryID); err != nil {
		api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusInternalServerError, err.Error(), request), response)
		return
	} else if savedQuery.UserID != user.ID.String() {
		if !user.Roles.Has(model.Role{Name: auth.RoleAdministrator}) {
			api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusNotFound, "query does not exist", request), response)
			return
		} else {
			if isPublic, err := s.DB.IsSavedQueryPublic(request.Context(), savedQuery.ID); err != nil {
				api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusInternalServerError, err.Error(), request), response)
				return
			} else if !isPublic {
				api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusNotFound, "query does not exist", request), response)
				return
			}
		}
	}

	if updateRequest.Query != "" {
		savedQuery.Query = updateRequest.Query
	}
	if updateRequest.Name != "" {
		savedQuery.Name = updateRequest.Name
	}
	if updateRequest.Description != "" {
		savedQuery.Description = updateRequest.Description
	}

	if savedQuery, err = s.DB.UpdateSavedQuery(request.Context(), savedQuery); err != nil {
		api.HandleDatabaseError(request, response, err)
	} else {
		api.WriteBasicResponse(request.Context(), savedQuery, http.StatusOK, response)
	}
}

func (s Resources) DeleteSavedQuery(response http.ResponseWriter, request *http.Request) {
	var (
		rawSavedQueryID = mux.Vars(request)[api.URIPathVariableSavedQueryID]
	)

	if user, isUser := auth.GetUserFromAuthCtx(ctx2.FromRequest(request).AuthCtx); !isUser {
		api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusBadRequest, "No associated user found", request), response)
	} else if savedQueryID, err := strconv.ParseInt(rawSavedQueryID, 10, 64); err != nil {
		api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusBadRequest, api.ErrorResponseDetailsIDMalformed, request), response)
	} else if savedQueryBelongsToUser, err := s.DB.SavedQueryBelongsToUser(request.Context(), user.ID, savedQueryID); errors.Is(err, database.ErrNotFound) {
		api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusNotFound, "query does not exist", request), response)
	} else if err != nil {
		api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusInternalServerError, api.ErrorResponseDetailsInternalServerError, request), response)
	} else {
		if !savedQueryBelongsToUser {
			if _, isAdmin := user.Roles.FindByName(auth.RoleAdministrator); !isAdmin {
				api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusForbidden, "User does not have permission to delete this query", request), response)
				return
			} else if isPublicQuery, err := s.DB.IsSavedQueryPublic(request.Context(), savedQueryID); err != nil {
				api.HandleDatabaseError(request, response, err)
				return
			} else if !isPublicQuery {
				api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusForbidden, "User does not have permission to delete this query", request), response)
				return
			}
		}

		if err := s.DB.DeleteSavedQuery(request.Context(), savedQueryID); errors.Is(err, database.ErrNotFound) {
			// This is an edge case and can only occur if the database has a concurrent operation that deletes the saved query
			// after the check at s.DB.SavedQueryBelongsToUser but before getting here.
			// Still, adding in the same check for good measure.
			api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusNotFound, "query does not exist", request), response)
		} else if err != nil {
			api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusInternalServerError, api.ErrorResponseDetailsInternalServerError, request), response)
		} else {
			response.WriteHeader(http.StatusNoContent)
		}

	}
}
