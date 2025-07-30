// Copyright 2025 Specter Ops, Inc.
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

package saml

import (
	"net/http"

	"github.com/crewjam/saml"
)

//go:generate go run go.uber.org/mock/mockgen -copyright_file=../../../../../LICENSE.header -destination=./mocks/saml.go -package=mocks . Service

// Service serves as a lightweight wrapper around the SAML package.
type Service interface {
	MakeAuthenticationRequest(serviceProvider saml.ServiceProvider, idpURL string, binding string, resultBinding string) (*saml.AuthnRequest, error)
	ParseResponse(serviceProvider saml.ServiceProvider, req *http.Request, possibleRequestIDs []string) (*saml.Assertion, error)
}

type Client struct{}

// MakeAuthenticationRequest abstracts creating an SAML authentication request using
// the HTTP-Redirect binding. It returns a URL that we will redirect the user to in order to start the auth process.
func (c *Client) MakeAuthenticationRequest(serviceProvider saml.ServiceProvider, idpURL string, binding string, resultBinding string) (*saml.AuthnRequest, error) {
	return serviceProvider.MakeAuthenticationRequest(idpURL, binding, resultBinding)
}

// ParseResponse abstracts the handling/validation of the IDP response.
// The purpose is to extract the SAML IDP response received in req, resolves
// artifacts when necessary, validates it, and returns the verified assertion.
func (c *Client) ParseResponse(serviceProvider saml.ServiceProvider, req *http.Request, possibleRequestIDs []string) (*saml.Assertion, error) {
	return serviceProvider.ParseResponse(req, possibleRequestIDs)
}
