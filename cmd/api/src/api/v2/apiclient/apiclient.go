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
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"time"

	"github.com/specterops/bloodhound/errors"
	"github.com/specterops/bloodhound/headers"
	"github.com/specterops/bloodhound/log"
	"github.com/specterops/bloodhound/mediatypes"
	"github.com/specterops/bloodhound/src/api"
)

type Client struct {
	Credentials CredentialsHandler
	Http        *http.Client
	ServiceURL  url.URL
}

func NewClient(rawServiceURL string) (Client, error) {
	if serviceURL, err := url.Parse(rawServiceURL); err != nil {
		return Client{}, err
	} else {
		return Client{
			Http: &http.Client{
				Timeout: time.Second * 5,
				Transport: &http.Transport{
					Proxy: http.ProxyFromEnvironment,
					DialContext: (&net.Dialer{
						Timeout:   time.Second,
						KeepAlive: 30 * time.Second,
					}).DialContext,
					ForceAttemptHTTP2:     true,
					MaxIdleConns:          100,
					IdleConnTimeout:       90 * time.Second,
					TLSHandshakeTimeout:   10 * time.Second,
					ExpectContinueTimeout: 1 * time.Second,
				},
			},
			ServiceURL: *serviceURL,
		}, nil
	}
}

func ReadAPIError(response *http.Response) error {
	if response.StatusCode == http.StatusNotFound {
		return errors.New("API returned a 404 error")
	}

	var apiError api.ErrorWrapper

	if err := api.ReadAPIV2ErrorResponsePayload(&apiError, response); err != nil {
		return err
	}

	return apiError
}

func (s Client) Request(method, path string, params url.Values, body any, header ...http.Header) (*http.Response, error) {
	endpoint := api.URLJoinPath(s.ServiceURL, path)
	endpoint.RawQuery = params.Encode()

	if request, err := http.NewRequest(method, endpoint.String(), nil); err != nil {
		return nil, err
	} else {
		// Set the header
		if len(header) > 0 {
			request.Header = header[0]
		}

		// query the Request and hand the response back to the user
		const (
			sleepInterval = time.Second * 5
			maxSleep      = sleepInterval * 5
		)

		started := time.Now()

		for {
			// Serialize the Request body - we expect a JSON serializable object here
			// This must be done on every retry, otherwise the buffer will be empty because it had been read
			if body != nil {
				buffer := &bytes.Buffer{}

				if err := json.NewEncoder(buffer).Encode(body); err != nil {
					return nil, err
				}

				request.Header.Set(headers.ContentType.String(), mediatypes.ApplicationJson.String())
				request.Body = io.NopCloser(buffer)
			}

			// Set our credentials either via signage or bearer token session
			// Credentials also have to be set with every attempt due to request signing
			if s.Credentials != nil {
				if err := s.Credentials.Handle(request); err != nil {
					return nil, err
				}
			}

			if response, err := s.Http.Do(request); err != nil {
				if time.Since(started) >= maxSleep {
					return nil, fmt.Errorf("waited %f seconds while retrying - Request failure cause: %w", maxSleep.Seconds(), err)
				}

				log.Infof("Request to %s failed with error: %v. Attempting a retry.", endpoint.String(), err)
				time.Sleep(sleepInterval)
			} else {
				return response, nil
			}
		}
	}
}

func (s Client) NewRequest(method string, path string, params url.Values, body io.ReadCloser) (*http.Request, error) {
	endpoint := api.URLJoinPath(s.ServiceURL, path)
	endpoint.RawQuery = params.Encode()

	return http.NewRequest(method, endpoint.String(), body)
}

func (s Client) Raw(request *http.Request) (*http.Response, error) {
	// Set our credentials either via signage or bearer token session
	if s.Credentials != nil {
		if err := s.Credentials.Handle(request); err != nil {
			return nil, err
		}
	}

	// query the Request and hand the response back to the user
	return s.Http.Do(request)
}
