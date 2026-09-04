// Copyright 2026 Specter Ops, Inc.
//
// Licensed under the Apache License, Version 2.0
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//	http://www.apache.org/licenses/LICENSE-2.0
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
	"path"
	"strconv"
	"strings"

	"github.com/gorilla/mux"
	"github.com/specterops/bloodhound/cmd/api/src/api"
	"github.com/specterops/bloodhound/cmd/api/src/database"
	"github.com/specterops/bloodhound/cmd/api/src/model"
	"github.com/specterops/bloodhound/cmd/api/src/model/ingest"
	"github.com/specterops/bloodhound/cmd/api/src/utils"
	"github.com/specterops/bloodhound/packages/go/bhlog/attr"
	"github.com/specterops/bloodhound/packages/go/bomenc"
	"github.com/specterops/bloodhound/packages/go/headers"
	"github.com/specterops/bloodhound/packages/go/mediatypes"
	"github.com/specterops/dawgs/graph"
)

//go:generate go run go.uber.org/mock/mockgen -copyright_file ../../../../../LICENSE.header -destination=./mocks/graphschemaextensions.go -package=mocks . OpenGraphSchemaService
type OpenGraphSchemaService interface {
	UpsertOpenGraphExtension(ctx context.Context, openGraphExtension model.GraphExtensionInput) (bool, error)
	ListExtensions(ctx context.Context) (model.GraphSchemaExtensions, error)
	DeleteExtension(ctx context.Context, extensionID int32) error
	GetEnvironmentKindsAndSchemaEnvironmentData(ctx context.Context, onlyBuiltin bool) (graph.Kinds, model.EnvironmentKindsToEnvironment, error)
	GetSchemaFindings(ctx context.Context, filters model.Filters, sort model.Sort, skip, limit int) ([]model.SchemaFinding, int, error)
}

// OpenGraphSchemaIngest - handles incoming graph extension upsert requests
func (s Resources) OpenGraphSchemaIngest(response http.ResponseWriter, request *http.Request) {
	var (
		ctx = request.Context()
		err error

		updated bool

		extractExtensionData func(file io.Reader) (model.GraphExtensionPayload, error)
		payload              model.GraphExtensionPayload
	)

	if request.Body == nil {
		var errMessage = "open graph extension payload cannot be empty"
		api.WriteErrorResponse(ctx, api.BuildErrorResponse(http.StatusBadRequest, errMessage, request), response)
		return
	}
	request.Body = http.MaxBytesReader(response, request.Body, api.DefaultAPIPayloadReadLimitBytes)
	defer request.Body.Close()
	switch {
	case utils.HeaderMatches(request.Header, headers.ContentType.String(), mediatypes.ApplicationJson.String()):
		extractExtensionData = extractExtensionDataFromJSON
	case utils.HeaderMatches(request.Header, headers.ContentType.String(), ingest.AllowedZipFileUploadTypes...):
		extractExtensionData = extractBundleFromZip
	default:
		var errMessage = fmt.Sprintf("%s; Content type must be application/json or application/zip",
			fmt.Sprintf("invalid content-type: %s", request.Header[headers.ContentType.String()]))
		api.WriteErrorResponse(ctx, api.BuildErrorResponse(http.StatusUnsupportedMediaType, errMessage, request), response)
		return
	}

	var graphExtensionInput model.GraphExtensionInput

	if payload, err = extractExtensionData(request.Body); err != nil {
		api.WriteErrorResponse(ctx, api.BuildErrorResponse(http.StatusBadRequest, err.Error(), request), response)
		return
	} else if graphExtensionInput, err = payload.ToGraphExtensionInput(); err != nil {
		api.WriteErrorResponse(ctx, api.BuildErrorResponse(http.StatusBadRequest, err.Error(), request), response)
		return
	} else if updated, err = s.OpenGraphSchemaService.UpsertOpenGraphExtension(ctx, graphExtensionInput); err != nil {
		switch {
		case strings.Contains(err.Error(), model.ErrGraphExtensionValidation.Error()) ||
			strings.Contains(err.Error(), model.ErrGraphExtensionBuiltIn.Error()):
			api.WriteErrorResponse(ctx, api.BuildErrorResponse(http.StatusBadRequest, err.Error(), request), response)
		case strings.Contains(err.Error(), model.ErrGraphDBRefreshKinds.Error()):
			fallthrough
		default:
			slog.WarnContext(
				ctx,
				"Error updating open graph schema",
				attr.Error(err),
			)
			api.WriteErrorResponse(ctx, api.BuildErrorResponse(http.StatusInternalServerError, api.ErrorResponseDetailsInternalServerError, request), response)
		}
	} else if updated {
		response.WriteHeader(http.StatusOK)
	} else {
		response.WriteHeader(http.StatusCreated)
	}
}

// Recognized component file names within an extension bundle.
const (
	bundleFileNameSchema       = "schema.json"
	bundleFileNamePzRules      = "pz_rules.json"
	bundleFileNameSavedQueries = "saved_queries.json"
)

func validateZipBundle(payload model.GraphExtensionPayload) error {
	if len(payload.GraphRelationshipFindings) > 0 && payload.PZRules == nil {
		return fmt.Errorf("extension declares relationship findings and requires a %q component", bundleFileNamePzRules)
	} else if len(payload.GraphRelationshipFindings) > 0 && len(payload.PZRules.Rules) == 0 {
		return fmt.Errorf("extension declares relationship findings and %q must contain at least one rule", bundleFileNamePzRules)
	}
	return nil
}

// extractExtensionDataFromJSON - extracts a model.GraphExtensionPayload from the incoming payload. Will return an error
// if the decoder fails to decode the payload.
func extractExtensionDataFromJSON(payload io.Reader) (model.GraphExtensionPayload, error) {
	var graphExtension model.GraphExtensionPayload

	if normFile, err := bomenc.NormalizeToUTF8(payload); err != nil {
		return graphExtension, fmt.Errorf("failed to normalize json payload: %w", err)
	} else if err = json.NewDecoder(normFile).Decode(&graphExtension); err != nil {
		return graphExtension, fmt.Errorf("unable to decode graph extension payload: %w", err)
	}

	return graphExtension, nil
}

// extractPZRulesFromJSON - extracts a model.PZRulesPayload from the incoming payload. Will return an error if the
// decoder fails to decode the payload.
func extractPZRulesFromJSON(payload io.Reader) (model.PZRulesPayload, error) {
	var pzRules model.PZRulesPayload

	if normFile, err := bomenc.NormalizeToUTF8(payload); err != nil {
		return pzRules, fmt.Errorf("failed to normalize %s: %w", bundleFileNamePzRules, err)
	} else if err = json.NewDecoder(normFile).Decode(&pzRules); err != nil {
		return pzRules, fmt.Errorf("unable to decode %s: %w", bundleFileNamePzRules, err)
	}

	return pzRules, nil
}

// extractSavedQueriesFromJSON - extracts a model.SavedQueriesPayload from the incoming payload. Will return an error
// if the decoder fails to decode the payload.
func extractSavedQueriesFromJSON(payload io.Reader) (model.SavedQueriesPayload, error) {
	var savedQueries model.SavedQueriesPayload

	if normFile, err := bomenc.NormalizeToUTF8(payload); err != nil {
		return savedQueries, fmt.Errorf("failed to normalize %s: %w", bundleFileNameSavedQueries, err)
	} else if err = json.NewDecoder(normFile).Decode(&savedQueries); err != nil {
		return savedQueries, fmt.Errorf("unable to decode %s: %w", bundleFileNameSavedQueries, err)
	}

	return savedQueries, nil
}

func validateSchemaComponent(payload model.GraphExtensionPayload) error {
	var invalidComponents []string

	if payload.PZRules != nil {
		invalidComponents = append(invalidComponents, bundleFileNamePzRules)
	}
	if payload.SavedQueries != nil {
		invalidComponents = append(invalidComponents, bundleFileNameSavedQueries)
	}

	if len(invalidComponents) > 0 {
		return fmt.Errorf("%s must not embed optional components: %s", bundleFileNameSchema, strings.Join(invalidComponents, ", "))
	}

	return nil
}

// extractBundleFromZip - returns a model.GraphExtensionPayload assembled from a ZIP archive. The archive must contain a
// schema.json component and may contain the optional pz_rules.json and saved_queries.json components
func extractBundleFromZip(payload io.Reader) (model.GraphExtensionPayload, error) {
	var (
		extension    model.GraphExtensionPayload
		archiveBytes []byte
		zipReader    *zip.Reader
		schemaFound  bool
		errs         []error
		err          error
	)

	if archiveBytes, err = io.ReadAll(payload); err != nil {
		return extension, fmt.Errorf("failed to read zip payload: %w", err)
	} else if zipReader, err = zip.NewReader(bytes.NewReader(archiveBytes), int64(len(archiveBytes))); err != nil {
		return extension, fmt.Errorf("unable to open zip archive: %w", err)
	}

	for _, file := range zipReader.File {
		if file.FileInfo().IsDir() {
			// ignore directories
			continue
		}
		if err = decodeFileIntoPayload(&extension, &schemaFound, file); err != nil {
			errs = append(errs, err)
		}
	}

	if !schemaFound {
		errs = append(errs, fmt.Errorf("required component %q not found in extension bundle", bundleFileNameSchema))
	} else if err = validateZipBundle(extension); err != nil {
		errs = append(errs, err)
	}

	return extension, errors.Join(errs...)
}

func decodeFileIntoPayload(extension *model.GraphExtensionPayload, schemaFound *bool, file *zip.File) error {
	// instead of switching on the filename, we could open each file and inspect for known keys. This is nice and simple though.
	switch file.Name {
	case bundleFileNameSchema:
		if *schemaFound {
			return fmt.Errorf("duplicate component %q in zip archive", bundleFileNameSchema)
		} else if reader, err := file.Open(); err != nil {
			return fmt.Errorf("unable to open %s in zip archive: %w", bundleFileNameSchema, err)
		} else {
			defer reader.Close()
			if schema, err := extractExtensionDataFromJSON(reader); err != nil {
				return err
			} else if err = validateSchemaComponent(schema); err != nil {
				return err
			} else {
				// We cannot guarantee the order of files in the zip archive, so we need to merge PZRules and SavedQueries into the schema if they were already extracted.
				schema.PZRules = extension.PZRules
				schema.SavedQueries = extension.SavedQueries
				*extension = schema
				*schemaFound = true
			}
		}
	case bundleFileNamePzRules:
		if extension.PZRules != nil {
			return fmt.Errorf("duplicate component %q in zip archive", bundleFileNamePzRules)
		} else if reader, err := file.Open(); err != nil {
			return fmt.Errorf("unable to open %s in zip archive: %w", bundleFileNamePzRules, err)
		} else {
			defer reader.Close()
			if rules, err := extractPZRulesFromJSON(reader); err != nil {
				return err
			} else {
				extension.PZRules = &rules
			}
		}
	case bundleFileNameSavedQueries:
		if extension.SavedQueries != nil {
			return fmt.Errorf("duplicate component %q in zip archive", bundleFileNameSavedQueries)
		} else if reader, err := file.Open(); err != nil {
			return fmt.Errorf("unable to open %s in zip archive: %w", bundleFileNameSavedQueries, err)
		} else {
			defer reader.Close()
			if queries, err := extractSavedQueriesFromJSON(reader); err != nil {
				return err
			} else {
				extension.SavedQueries = &queries
			}
		}
	default:
		return fmt.Errorf("unexpected file %q in zip archive", path.Base(file.Name))
	}

	return nil
}

type ExtensionsResponse struct {
	Extensions []ExtensionInfo `json:"extensions"`
}

type ExtensionInfo struct {
	ID        int32  `json:"id"`
	Name      string `json:"name"`
	Namespace string `json:"namespace"`
	Version   string `json:"version"`
	IsBuiltIn bool   `json:"is_builtin"`
}

func (s Resources) ListExtensions(response http.ResponseWriter, request *http.Request) {
	var (
		ctx = request.Context()
	)

	if extensions, err := s.OpenGraphSchemaService.ListExtensions(ctx); err != nil {
		api.WriteErrorResponse(ctx, api.BuildErrorResponse(http.StatusInternalServerError, api.ErrorResponseDetailsInternalServerError, request), response)
		return
	} else {
		var extensionsResponse = make([]ExtensionInfo, len(extensions))
		for i, extension := range extensions {
			extensionsResponse[i] = ExtensionInfo{
				ID:        extension.ID,
				Name:      extension.DisplayName,
				Version:   extension.Version,
				IsBuiltIn: extension.IsBuiltin,
				Namespace: extension.Namespace,
			}
		}

		api.WriteBasicResponse(ctx, ExtensionsResponse{Extensions: extensionsResponse}, http.StatusOK, response)
	}
}

func (s Resources) DeleteExtension(response http.ResponseWriter, request *http.Request) {
	var (
		ctx         = request.Context()
		extensionID = mux.Vars(request)[api.URIPathVariableExtensionID]
	)

	if extID, err := strconv.ParseInt(extensionID, 10, 32); err != nil {
		api.WriteErrorResponse(ctx, api.BuildErrorResponse(http.StatusBadRequest, api.ErrorResponseDetailsIDMalformed, request), response)
	} else if err := s.OpenGraphSchemaService.DeleteExtension(ctx, int32(extID)); err != nil {
		if errors.Is(err, database.ErrNotFound) {
			api.WriteErrorResponse(ctx, api.BuildErrorResponse(http.StatusNotFound, fmt.Sprintf("no extension found matching extension id: %s", extensionID), request), response)
		} else if errors.Is(err, model.ErrGraphExtensionBuiltIn) {
			api.WriteErrorResponse(ctx, api.BuildErrorResponse(http.StatusBadRequest, "built-in extensions cannot be deleted", request), response)
		} else {
			api.WriteErrorResponse(ctx, api.BuildErrorResponse(http.StatusInternalServerError, api.ErrorResponseDetailsInternalServerError, request), response)
		}
	} else {
		response.WriteHeader(http.StatusNoContent)
	}
}
