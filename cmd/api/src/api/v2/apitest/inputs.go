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

package apitest

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"net/url"

	"github.com/specterops/bloodhound/src/test/must"
)

type Input struct {
	query   url.Values
	urlVars map[string]string
	request *http.Request
}

type InputFunc func(input *Input)

func SetHeader(input *Input, key string, value string) {
	input.request.Header.Set(key, value)
}

func SetContext(input *Input, ctx context.Context) {
	input.request = input.request.Clone(ctx)
}

func BodyBytes(input *Input, body []byte) {
	input.request.Body = io.NopCloser(bytes.NewReader(body))
}

func BodyString(input *Input, body string) {
	BodyBytes(input, []byte(body))
}

func BodyStruct(input *Input, body any) {
	data := must.MarshalJSON(body)
	BodyBytes(input, data)
}

func AddQueryParam(input *Input, key string, value string) {
	input.query.Add(key, value)
}

func DeleteQueryParam(input *Input, key string) {
	input.query.Del(key)
}

func SetURLVar(input *Input, key string, value string) {
	input.urlVars[key] = value
}

func DeleteURLVar(input *Input, key string) {
	delete(input.urlVars, key)
}
