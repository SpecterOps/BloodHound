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

package apiclient

import (
	"bytes"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"strconv"

	"github.com/gofrs/uuid"
	"github.com/specterops/bloodhound/headers"
	"github.com/specterops/bloodhound/src/api"
	v2 "github.com/specterops/bloodhound/src/api/v2"
	authapi "github.com/specterops/bloodhound/src/api/v2/auth"
	"github.com/specterops/bloodhound/src/auth"
	"github.com/specterops/bloodhound/src/model"
)

func (s Client) ListSAMLSignOnEndpoints() (v2.ListSAMLSignOnEndpointsResponse, error) {
	var providersResponse v2.ListSAMLSignOnEndpointsResponse

	if response, err := s.Request(http.MethodGet, "api/v2/saml/sso", nil, nil); err != nil {
		return providersResponse, err
	} else {
		defer response.Body.Close()

		if api.IsErrorResponse(response) {
			return providersResponse, ReadAPIError(response)
		}

		return providersResponse, api.ReadAPIV2ResponsePayload(&providersResponse, response)
	}
}

func (s Client) ListSAMLIdentityProviders() (v2.ListSAMLProvidersResponse, error) {
	var providersResponse v2.ListSAMLProvidersResponse

	if response, err := s.Request(http.MethodGet, "api/v2/saml", nil, nil); err != nil {
		return providersResponse, err
	} else {
		defer response.Body.Close()

		if api.IsErrorResponse(response) {
			return providersResponse, ReadAPIError(response)
		}

		return providersResponse, api.ReadAPIV2ResponsePayload(&providersResponse, response)
	}
}

func (s Client) CreateSAMLIdentityProvider(request v2.CreateSAMLAuthProviderRequest) (model.SAMLProvider, error) {
	var newProvider model.SAMLProvider

	if response, err := s.Request(http.MethodPost, "api/v2/saml", nil, request); err != nil {
		return newProvider, err
	} else {
		defer response.Body.Close()

		if api.IsErrorResponse(response) {
			return newProvider, ReadAPIError(response)
		}

		return newProvider, api.ReadAPIV2ResponsePayload(&newProvider, response)
	}
}

func (s Client) CreateSAMLIdentityProviderMultipart(name, metadata string) (model.SAMLProvider, error) {
	var (
		newProvider model.SAMLProvider

		buffer          = &bytes.Buffer{}
		header          = make(http.Header)
		multipartWriter = multipart.NewWriter(buffer)
	)

	if err := multipartWriter.WriteField("name", name); err != nil {
		return newProvider, err
	} else if fileWriter, err := multipartWriter.CreateFormFile("metadata", "metadata.xml"); err != nil {
		return newProvider, err
	} else {
		if _, err := io.Copy(fileWriter, bytes.NewBufferString(metadata)); err != nil {
			return newProvider, fmt.Errorf("failed to copy metadata to file: %w", err)
		}
		multipartWriter.Close()

		header.Set(headers.ContentType.String(), multipartWriter.FormDataContentType())

		if response, err := s.Request(http.MethodPost, "api/v2/saml/providers", nil, buffer.Bytes(), header); err != nil {
			return newProvider, err
		} else {
			defer response.Body.Close()

			if api.IsErrorResponse(response) {
				return newProvider, ReadAPIError(response)
			}

			return newProvider, api.ReadJsonResponsePayload(&newProvider, response)
		}
	}
}

func (s Client) ListAuthTokens() (v2.ListTokensResponse, error) {
	var tokens v2.ListTokensResponse

	if response, err := s.Request(http.MethodGet, "api/v2/tokens", nil, nil); err != nil {
		return tokens, err
	} else {
		defer response.Body.Close()

		if api.IsErrorResponse(response) {
			return tokens, ReadAPIError(response)
		}

		return tokens, api.ReadAPIV2ResponsePayload(&tokens, response)
	}
}

func (s Client) ListUserTokens(id uuid.UUID) (v2.ListTokensResponse, error) {
	var tokens v2.ListTokensResponse

	params := url.Values{}
	params.Set("user_id", fmt.Sprintf("eq:%s", id.String()))
	if response, err := s.Request(http.MethodGet, "api/v2/tokens", params, nil); err != nil {
		return tokens, err
	} else {
		defer response.Body.Close()

		if api.IsErrorResponse(response) {
			return tokens, ReadAPIError(response)
		}

		return tokens, api.ReadAPIV2ResponsePayload(&tokens, response)
	}
}

func (s Client) EnrollMFA(id uuid.UUID, secret string) (authapi.MFAEnrollmentReponse, error) {
	var (
		enrollmentResponse authapi.MFAEnrollmentReponse
		payload            = authapi.MFAEnrollmentRequest{
			Secret: secret,
		}
	)

	if response, err := s.Request(http.MethodPost, fmt.Sprintf("api/v2/bloodhound-users/%s/mfa", id), nil, payload); err != nil {
		return enrollmentResponse, err
	} else {
		defer response.Body.Close()

		if api.IsErrorResponse(response) {
			return enrollmentResponse, ReadAPIError(response)
		}

		return enrollmentResponse, api.ReadAPIV2ResponsePayload(&enrollmentResponse, response)
	}
}

func (s Client) ActivateMFA(id uuid.UUID, otp string) (authapi.MFAStatusResponse, error) {
	var (
		mfaStatusResponse authapi.MFAStatusResponse
		payload           = authapi.MFAActivationRequest{
			OTP: otp,
		}
	)

	if response, err := s.Request(http.MethodPost, fmt.Sprintf("api/v2/bloodhound-users/%s/mfa-activation", id), nil, payload); err != nil {
		return mfaStatusResponse, err
	} else {
		defer response.Body.Close()

		if api.IsErrorResponse(response) {
			return mfaStatusResponse, ReadAPIError(response)
		}

		return mfaStatusResponse, api.ReadAPIV2ResponsePayload(&mfaStatusResponse, response)
	}
}

func (s Client) GetMFAActivationStatus(id uuid.UUID) (authapi.MFAStatusResponse, error) {
	var mfaStatusResponse authapi.MFAStatusResponse

	if response, err := s.Request(http.MethodGet, fmt.Sprintf("api/v2/bloodhound-users/%s/mfa-activation", id), nil, nil); err != nil {
		return mfaStatusResponse, err
	} else {
		defer response.Body.Close()

		if api.IsErrorResponse(response) {
			return mfaStatusResponse, ReadAPIError(response)
		}

		return mfaStatusResponse, api.ReadAPIV2ResponsePayload(&mfaStatusResponse, response)
	}
}

func (s Client) LookupSelf() (model.User, error) {
	var self model.User
	if response, err := s.Request(http.MethodGet, "api/v2/auth/self", nil, nil); err != nil {
		return self, err
	} else {
		defer response.Body.Close()

		if api.IsErrorResponse(response) {
			return self, ReadAPIError(response)
		}

		return self, api.ReadAPIV2ResponsePayload(&self, response)
	}
}

func (s Client) CreateSAMLUser(userPrincipal, userEmailAddress string, samlProviderID int32, roles []int32) (model.User, error) {
	var newUser model.User

	payload := v2.CreateUserRequest{
		UpdateUserRequest: v2.UpdateUserRequest{
			Principal:      userPrincipal,
			EmailAddress:   userEmailAddress,
			Roles:          roles,
			SAMLProviderID: strconv.FormatInt(int64(samlProviderID), 10),
		},
	}

	if response, err := s.Request(http.MethodPost, "api/v2/bloodhound-users", nil, payload); err != nil {
		return newUser, err
	} else {
		defer response.Body.Close()

		if api.IsErrorResponse(response) {
			return newUser, ReadAPIError(response)
		}

		return newUser, api.ReadAPIV2ResponsePayload(&newUser, response)
	}
}

func (s Client) UpdateUser(userID uuid.UUID, updateUserRequest v2.UpdateUserRequest) error {
	if response, err := s.Request(http.MethodPut, fmt.Sprintf("api/v2/bloodhound-users/%s", userID), nil, updateUserRequest); err != nil {
		return err
	} else {
		defer response.Body.Close()

		if api.IsErrorResponse(response) {
			return ReadAPIError(response)
		}

		return nil
	}
}

func (s Client) CreateUser(userPrincipal, userEmailAddress string, roles []int32) (model.User, error) {
	var newUser model.User

	payload := v2.CreateUserRequest{
		UpdateUserRequest: v2.UpdateUserRequest{
			Principal:    userPrincipal,
			EmailAddress: userEmailAddress,
			Roles:        roles,
		},
	}

	if response, err := s.Request(http.MethodPost, "api/v2/bloodhound-users", nil, payload); err != nil {
		return newUser, err
	} else {
		defer response.Body.Close()

		if api.IsErrorResponse(response) {
			return newUser, ReadAPIError(response)
		}

		return newUser, api.ReadAPIV2ResponsePayload(&newUser, response)
	}
}

func (s Client) DeleteUser(userID uuid.UUID) error {
	if response, err := s.Request(http.MethodDelete, fmt.Sprintf("api/v2/bloodhound-users/%s", userID), nil, nil); err != nil {
		return err
	} else {
		defer response.Body.Close()

		if api.IsErrorResponse(response) {
			return ReadAPIError(response)
		}
	}

	return nil
}

func (s Client) GetUser(id uuid.UUID) (model.User, error) {
	var user model.User

	if response, err := s.Request(http.MethodGet, fmt.Sprintf("api/v2/bloodhound-users/%s", id), nil, nil); err != nil {
		return user, err
	} else {
		defer response.Body.Close()

		if api.IsErrorResponse(response) {
			return user, ReadAPIError(response)
		}

		return user, api.ReadAPIV2ResponsePayload(&user, response)
	}
}

func (s Client) ListUsers() (v2.ListUsersResponse, error) {
	var users v2.ListUsersResponse

	if response, err := s.Request(http.MethodGet, "api/v2/bloodhound-users", nil, nil); err != nil {
		return users, err
	} else {
		defer response.Body.Close()

		if api.IsErrorResponse(response) {
			return users, ReadAPIError(response)
		}

		return users, api.ReadAPIV2ResponsePayload(&users, response)
	}
}

func (s Client) UserAddRole(userID uuid.UUID, roleID int32) error {
	if response, err := s.Request(http.MethodPost, fmt.Sprintf("api/v2/bloodhound-users/%s/roles/%d", userID, roleID), nil, nil); err != nil {
		return err
	} else {
		defer response.Body.Close()

		if api.IsErrorResponse(response) {
			return ReadAPIError(response)
		}
	}

	return nil
}

func (s Client) UserRemoveRole(userID uuid.UUID, roleID int32) error {
	if response, err := s.Request(http.MethodDelete, fmt.Sprintf("api/v2/bloodhound-users/%s/roles/%d", userID, roleID), nil, nil); err != nil {
		return err
	} else {
		defer response.Body.Close()

		if api.IsErrorResponse(response) {
			return ReadAPIError(response)
		}
	}

	return nil
}

func (s Client) GetPermission(id int32) (model.Permission, error) {
	var permission model.Permission

	if response, err := s.Request(http.MethodGet, fmt.Sprintf("api/v2/permissions/%d", id), nil, nil); err != nil {
		return permission, err
	} else {
		defer response.Body.Close()

		if api.IsErrorResponse(response) {
			return permission, ReadAPIError(response)
		}

		return permission, api.ReadAPIV2ResponsePayload(&permission, response)
	}
}

func (s Client) ListPermissions() (v2.ListPermissionsResponse, error) {
	var permissions v2.ListPermissionsResponse

	if response, err := s.Request(http.MethodGet, "api/v2/permissions", nil, nil); err != nil {
		return permissions, err
	} else {
		defer response.Body.Close()

		if api.IsErrorResponse(response) {
			return permissions, ReadAPIError(response)
		}

		return permissions, api.ReadAPIV2ResponsePayload(&permissions, response)
	}
}

func (s Client) GetRole(id int32) (model.Role, error) {
	var role model.Role

	if response, err := s.Request(http.MethodGet, fmt.Sprintf("api/v2/auth/roles/%d", id), nil, nil); err != nil {
		return role, err
	} else {
		defer response.Body.Close()

		if api.IsErrorResponse(response) {
			return role, ReadAPIError(response)
		}

		return role, api.ReadAPIV2ResponsePayload(&role, response)
	}
}

func (s Client) ListRoles() (v2.ListRolesResponse, error) {
	var roles v2.ListRolesResponse

	if response, err := s.Request(http.MethodGet, "api/v2/roles", nil, nil); err != nil {
		return roles, err
	} else {
		defer response.Body.Close()

		if api.IsErrorResponse(response) {
			return roles, ReadAPIError(response)
		}

		return roles, api.ReadAPIV2ResponsePayload(&roles, response)
	}
}

func (s Client) GetSelf() (model.User, error) {
	var user model.User

	if response, err := s.Request(http.MethodGet, "api/v2/self", nil, nil); err != nil {
		return user, err
	} else {
		defer response.Body.Close()

		if api.IsErrorResponse(response) {
			return user, ReadAPIError(response)
		}

		return user, api.ReadAPIV2ResponsePayload(&user, response)
	}
}

func (s Client) SetUserSecret(userID uuid.UUID, secret string, needsPasswordReset bool) error {
	setUserSecret := v2.SetUserSecretRequest{
		Secret:             secret,
		NeedsPasswordReset: needsPasswordReset,
	}

	if response, err := s.Request(http.MethodPut, fmt.Sprintf("api/v2/bloodhound-users/%s/secret", userID), nil, setUserSecret); err != nil {
		return err
	} else {
		defer response.Body.Close()

		if api.IsErrorResponse(response) {
			return ReadAPIError(response)
		}

		return nil
	}
}

func (s Client) CreateUserToken(userID uuid.UUID, tokenName string) (model.AuthToken, error) {
	var token model.AuthToken

	body := map[string]any{"token_name": tokenName, "user_id": userID}

	if response, err := s.Request(http.MethodPost, "api/v2/tokens", nil, body); err != nil {
		return token, err
	} else {
		defer response.Body.Close()

		if api.IsErrorResponse(response) {
			return token, ReadAPIError(response)
		}

		return token, api.ReadAPIV2ResponsePayload(&token, response)
	}
}

func (s Client) DeleteUserToken(userID, tokenID uuid.UUID) error {
	if response, err := s.Request(http.MethodDelete, fmt.Sprintf("api/v2/tokens/%s", tokenID.String()), nil, nil); err != nil {
		return err
	} else {
		defer response.Body.Close()

		if api.IsErrorResponse(response) {
			return ReadAPIError(response)
		}
	}

	return nil
}

func (s Client) LoginSecret(username, secret string) (api.LoginResponse, error) {
	var (
		loginResponse api.LoginResponse
		loginRequest  = api.LoginRequest{
			LoginMethod: auth.ProviderTypeSecret,
			Username:    username,
			Secret:      secret,
		}
	)

	if response, err := s.Request(http.MethodPost, "api/v2/login", nil, loginRequest); err != nil {
		return loginResponse, err
	} else {
		defer response.Body.Close()

		if api.IsErrorResponse(response) {
			return loginResponse, ReadAPIError(response)
		}

		return loginResponse, api.ReadAPIV2ResponsePayload(&loginResponse, response)
	}
}

func (s Client) LoginSAML(organization, username string) error {
	panic("TODO")
}
