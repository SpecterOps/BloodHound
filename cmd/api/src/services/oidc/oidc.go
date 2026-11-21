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

package oidc

import (
	"context"
	"errors"
	"net/http"

	"github.com/coreos/go-oidc/v3/oidc"
)

//go:generate go run go.uber.org/mock/mockgen -copyright_file=../../../../../LICENSE.header -destination=./mocks/oidc.go -package=mocks . Service

// Serves as a lightweight wrapper around the oidc package which allows for
// OpenID Connect client logic to be abstracted.

var oidcHTTPClient = &http.Client{
	CheckRedirect: func(request *http.Request, via []*http.Request) error {
		if request.URL.Scheme != "https" {
			return errors.New("OIDC redirect must use HTTPS")
		}

		// Preserve Go's default redirect limit when using a custom CheckRedirect
		if len(via) >= 10 {
			return errors.New("OIDC request reached the redirect limit of 10 requests")
		}

		return nil
	},
}

// DoRequest sends an OIDC request and rejects redirects to non-HTTPS URLs
func DoRequest(request *http.Request) (*http.Response, error) {
	return oidcHTTPClient.Do(request)
}

type Service interface {
	NewProvider(ctx context.Context, issuer string) (*oidc.Provider, error)
}

type Client struct{}

func (c *Client) NewProvider(ctx context.Context, issuer string) (*oidc.Provider, error) {
	discoveryContext := oidc.ClientContext(ctx, oidcHTTPClient)

	return oidc.NewProvider(discoveryContext, issuer)
}
