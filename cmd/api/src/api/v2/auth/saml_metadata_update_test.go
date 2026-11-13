// Copyright 2026 Specter Ops, Inc.
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

package auth_test

import (
	"bytes"
	"context"
	"encoding/xml"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/crewjam/saml/samlsp"
	"github.com/gorilla/mux"
	"github.com/specterops/bloodhound/cmd/api/src/api"
	v2auth "github.com/specterops/bloodhound/cmd/api/src/api/v2/auth"
	"github.com/specterops/bloodhound/cmd/api/src/auth"
	"github.com/specterops/bloodhound/cmd/api/src/config"
	dbMocks "github.com/specterops/bloodhound/cmd/api/src/database/mocks"
	"github.com/specterops/bloodhound/cmd/api/src/database/types/null"
	"github.com/specterops/bloodhound/cmd/api/src/model"
	"github.com/specterops/bloodhound/cmd/api/src/test/fixtures"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

// samlMetadataRequest builds the multipart request used by both public SAML management handlers.
func samlMetadataRequest(t *testing.T, method, target string, metadata []byte, fields map[string]string) *http.Request {
	t.Helper()
	var (
		body   bytes.Buffer
		writer = multipart.NewWriter(&body)
	)

	for name, value := range fields {
		require.NoError(t, writer.WriteField(name, value))
	}
	file, err := writer.CreateFormFile("metadata", "idp_metadata.xml")
	require.NoError(t, err)
	_, err = file.Write(metadata)
	require.NoError(t, err)
	require.NoError(t, writer.Close())

	request := httptest.NewRequest(method, target, &body)
	request.Header.Set("Content-Type", writer.FormDataContentType())
	return request
}

// TestManagementResource_SAMLMetadataCreateAndUpdate verifies the same federation metadata is accepted for creation and replacement without changing BloodHound's callback version.
func TestManagementResource_SAMLMetadataCreateAndUpdate(t *testing.T) {
	var (
		metadata       = fixtures.NewLoader(fixtures.NewTestErrorHandler(t)).Get("saml/idp_metadata.xml")
		providerConfig = model.SSOProviderConfig{
			AutoProvision: model.SSOProviderAutoProvisionConfig{
				Enabled:       true,
				DefaultRoleId: 3,
				RoleProvision: true,
			},
		}
	)

	t.Run("create accepts federation metadata", func(t *testing.T) {
		var (
			controller = gomock.NewController(t)
			database   = dbMocks.NewMockDatabase(controller)
			resource   = v2auth.NewManagementResource(config.Configuration{}, database, auth.Authorizer{}, nil, nil, nil, nil)
			response   = httptest.NewRecorder()
			request    = samlMetadataRequest(t, http.MethodPost, "/api/v2/sso-providers/saml", metadata, map[string]string{
				"name":                                  "existing-provider",
				"config.auto_provision.enabled":         "true",
				"config.auto_provision.default_role_id": "3",
				"config.auto_provision.role_provision":  "true",
			})
		)
		database.EXPECT().GetRole(gomock.Any(), int32(3)).Return(model.Role{Serial: model.Serial{ID: 3}}, nil)
		database.EXPECT().CreateSAMLIdentityProvider(gomock.Any(), gomock.Any(), gomock.Any()).DoAndReturn(
			func(_ context.Context, provider model.SAMLProvider, receivedConfig model.SSOProviderConfig) (model.SAMLProvider, error) {
				require.Equal(t, metadata, provider.MetadataXML)
				require.Equal(t, "http://auth.specterops.io/adfs/services/trust", provider.IssuerURI)
				require.Equal(t, "https://auth.specterops.io/adfs/ls/", provider.SingleSignOnURI)
				require.Equal(t, providerConfig, receivedConfig)
				return provider, nil
			})

		resource.CreateSAMLProviderMultipart(response, request)
		require.Equal(t, http.StatusOK, response.Code, response.Body.String())
	})

	for _, testCase := range []struct {
		name        string
		rootVersion model.SAMLRootURIVersion
		spCallback  string
	}{
		{name: "replace preserves legacy callback", rootVersion: model.SAMLRootURIVersion1},
		{name: "replace preserves current callback", rootVersion: model.SAMLRootURIVersion2},
		{name: "foreign current callback cannot change legacy version", rootVersion: model.SAMLRootURIVersion1, spCallback: "https://other.example.test/api/v2/sso/foreign/callback"},
		{name: "foreign legacy callback cannot change current version", rootVersion: model.SAMLRootURIVersion2, spCallback: "https://other.example.test/api/v1/login/saml/foreign/acs"},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			metadataXML := metadata
			if testCase.spCallback != "" {
				descriptor, err := samlsp.ParseMetadata(metadata)
				require.NoError(t, err)
				for descriptorIndex := range descriptor.SPSSODescriptors {
					for serviceIndex := range descriptor.SPSSODescriptors[descriptorIndex].AssertionConsumerServices {
						descriptor.SPSSODescriptors[descriptorIndex].AssertionConsumerServices[serviceIndex].Location = testCase.spCallback
					}
				}
				metadataXML, err = xml.Marshal(descriptor)
				require.NoError(t, err)
			}

			var (
				controller = gomock.NewController(t)
				database   = dbMocks.NewMockDatabase(controller)
				resource   = v2auth.NewManagementResource(config.Configuration{}, database, auth.Authorizer{}, nil, nil, nil, nil)
				response   = httptest.NewRecorder()
				request    = samlMetadataRequest(t, http.MethodPatch, "/api/v2/sso-providers/77", metadataXML, nil)
				provider   = model.SSOProvider{
					Serial: model.Serial{ID: 77},
					Name:   "existing-provider",
					Type:   model.SessionAuthProviderSAML,
					Config: providerConfig,
					SAMLProvider: &model.SAMLProvider{
						Serial:                     model.Serial{ID: 42},
						SSOProviderID:              null.Int32From(77),
						Name:                       "existing-provider",
						DisplayName:                "existing-provider",
						IssuerURI:                  "https://old.example.test/issuer",
						SingleSignOnURI:            "https://old.example.test/sso",
						MetadataXML:                []byte("<old/>"),
						RootURIVersion:             testCase.rootVersion,
						PrincipalAttributeMappings: []string{"custom-principal"},
					},
				}
			)
			database.EXPECT().GetSSOProviderById(gomock.Any(), int32(77)).Return(provider, nil)
			database.EXPECT().UpdateSAMLIdentityProvider(gomock.Any(), gomock.Any()).DoAndReturn(
				func(_ context.Context, updated model.SSOProvider) (model.SAMLProvider, error) {
					require.Equal(t, int32(77), updated.ID)
					require.Equal(t, int32(42), updated.SAMLProvider.ID)
					require.Equal(t, null.Int32From(77), updated.SAMLProvider.SSOProviderID)
					require.Equal(t, "existing-provider", updated.Name)
					require.Equal(t, providerConfig, updated.Config)
					require.Equal(t, []string{"custom-principal"}, updated.SAMLProvider.PrincipalAttributeMappings)
					require.Equal(t, testCase.rootVersion, updated.SAMLProvider.RootURIVersion)
					require.Equal(t, metadataXML, updated.SAMLProvider.MetadataXML)
					require.Equal(t, "http://auth.specterops.io/adfs/services/trust", updated.SAMLProvider.IssuerURI)
					require.Equal(t, "https://auth.specterops.io/adfs/ls/", updated.SAMLProvider.SingleSignOnURI)
					return *updated.SAMLProvider, nil
				})

			router := mux.NewRouter()
			router.HandleFunc("/api/v2/sso-providers/{"+api.URIPathVariableSSOProviderID+"}", resource.UpdateSSOProvider).Methods(http.MethodPatch)
			router.ServeHTTP(response, request)
			require.Equal(t, http.StatusOK, response.Code, response.Body.String())
		})
	}
}
