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

//go:generate go run go.uber.org/mock/mockgen -copyright_file=../../../../../../LICENSE.header -destination=./mocks/credential.go -package=mocks . CredentialsHandler

import (
	"fmt"
	"net/http"

	"github.com/golang-jwt/jwt/v4"
	"github.com/specterops/bloodhound/headers"
	"github.com/specterops/bloodhound/src/api"
	"github.com/specterops/bloodhound/src/auth"
)

type CredentialsHandler interface {
	Handle(request *http.Request) error
}

type TokenCredentialsHandler struct {
	TokenID  string
	TokenKey string
}

func (s *TokenCredentialsHandler) Handle(request *http.Request) error {
	return api.SignRequest(s.TokenID, s.TokenKey, request)
}

type SecretCredentialsHandler struct {
	Username string
	Secret   string
	Client   Client
	jwt      string
	session  auth.SessionData
}

func (s *SecretCredentialsHandler) SetSessionToken(sessionToken string) error {
	var jwtParser jwt.Parser

	if _, _, err := jwtParser.ParseUnverified(sessionToken, &s.session); err != nil {
		return fmt.Errorf("failed pasring JWT session token: %w", err)
	} else {
		s.jwt = sessionToken
	}

	return nil
}

func (s *SecretCredentialsHandler) login() error {
	if resp, err := s.Client.LoginSecret(s.Username, s.Secret); err != nil {
		return err
	} else {
		return s.SetSessionToken(resp.SessionToken)
	}
}

func (s *SecretCredentialsHandler) Handle(request *http.Request) error {
	if s.jwt == "" || s.session.Valid() != nil {
		if err := s.login(); err != nil {
			return err
		}
	}

	request.Header.Set(headers.Authorization.String(), fmt.Sprintf("%s %s", api.AuthorizationSchemeBearer, s.jwt))
	return nil
}
