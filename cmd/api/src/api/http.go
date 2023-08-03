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

package api

//go:generate go run go.uber.org/mock/mockgen -copyright_file=../../../../LICENSE.header -destination=./mocks/http.go -package=mocks . HTTPClient

import "net/http"

type HTTPClient interface {
	Do(request *http.Request) (*http.Response, error)
	CloseIdleConnections()
}

type httpClient struct {
	client *http.Client
}

func NewHTTPClient() HTTPClient {
	return httpClient{
		client: &http.Client{},
	}
}

func WrapHTTPClient(client *http.Client) HTTPClient {
	return httpClient{
		client: client,
	}
}

func (s httpClient) Do(request *http.Request) (*http.Response, error) {
	return s.client.Do(request)
}

func (s httpClient) CloseIdleConnections() {
	s.client.CloseIdleConnections()
}
