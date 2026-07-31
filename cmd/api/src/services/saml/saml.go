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
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/crewjam/saml"
)

//go:generate go run go.uber.org/mock/mockgen -copyright_file=../../../../../LICENSE.header -destination=./mocks/saml.go -package=mocks . Service

// Service serves as a lightweight wrapper around the SAML package.
type Service interface {
	MakeAuthenticationRequest(serviceProvider saml.ServiceProvider, idpURL string, binding string, resultBinding string) (*saml.AuthnRequest, error)
	ParseResponse(serviceProvider saml.ServiceProvider, req *http.Request, possibleRequestIDs []string) (*ValidatedResponse, error)
}

type Client struct{}

type ValidatedResponse struct {
	Assertion            *saml.Assertion
	ResponseID           string
	ResponseIssueInstant time.Time
}

// MakeAuthenticationRequest abstracts creating an SAML authentication request using
// the HTTP-Redirect binding. It returns a URL that we will redirect the user to in order to start the auth process.
func (c *Client) MakeAuthenticationRequest(serviceProvider saml.ServiceProvider, idpURL string, binding string, resultBinding string) (*saml.AuthnRequest, error) {
	return serviceProvider.MakeAuthenticationRequest(idpURL, binding, resultBinding)
}

// ParseResponse abstracts the handling/validation of the IDP response.
// The purpose is to extract the SAML IDP response received in req, resolves
// artifacts when necessary, validates it, and returns the verified assertion.
func (c *Client) ParseResponse(serviceProvider saml.ServiceProvider, req *http.Request, possibleRequestIDs []string) (*ValidatedResponse, error) {
	var parsedResponse saml.Response

	configuredRequestIDValidator := serviceProvider.ValidateRequestID
	serviceProvider.ValidateRequestID = func(response saml.Response, possibleRequestIDs []string) error {
		parsedResponse = response

		if configuredRequestIDValidator != nil {
			return configuredRequestIDValidator(response, possibleRequestIDs)
		}

		return validateRequestID(serviceProvider.AllowIDPInitiated, response, possibleRequestIDs)
	}

	assertion, err := serviceProvider.ParseResponse(req, possibleRequestIDs)
	if err != nil {
		return nil, err
	} else if parsedResponse.ID == "" {
		return nil, errors.New("SAML response is missing its ID")
	} else if assertion == nil {
		return nil, errors.New("SAML response is missing its assertion")
	} else if assertion.ID == "" {
		return nil, errors.New("SAML assertion is missing its ID")
	}

	return &ValidatedResponse{
		Assertion:            assertion,
		ResponseID:           parsedResponse.ID,
		ResponseIssueInstant: parsedResponse.IssueInstant,
	}, nil
}

func validateRequestID(allowIDPInitiated bool, response saml.Response, possibleRequestIDs []string) error {
	if allowIDPInitiated {
		return nil
	}

	for _, possibleRequestID := range possibleRequestIDs {
		if response.InResponseTo == possibleRequestID {
			return nil
		}
	}

	return fmt.Errorf("`InResponseTo` does not match any of the possible request IDs (expected %v)", possibleRequestIDs)
}
