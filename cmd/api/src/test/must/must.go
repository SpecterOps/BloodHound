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

package must

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/specterops/bloodhound/src/database/types"
	"github.com/gofrs/uuid"
	"github.com/specterops/bloodhound/crypto"
)

func MarshalJSON(value any) []byte {
	if content, err := json.MarshalIndent(value, "", "  "); err != nil {
		panic(fmt.Sprintf("failed JSON marshaling: %v", err))
	} else {
		return content
	}
}

func MarshalJSONReader(value any) io.Reader {
	return bytes.NewBuffer(MarshalJSON(value))
}

func ParseURL(rawURL string) url.URL {
	if urlPtr, err := url.Parse(rawURL); err != nil {
		panic(fmt.Sprintf("bad url: %s: %v", rawURL, err))
	} else {
		return *urlPtr
	}
}

func ParseQuery(rawQuery string) url.Values {
	if queryPtr, err := url.ParseQuery(rawQuery); err != nil {
		panic(fmt.Sprintf("bad query string: %s: %v", rawQuery, err))
	} else {
		return queryPtr
	}
}

func ParseTime(layout, value string) time.Time {
	if parsedTime, err := time.Parse(layout, value); err != nil {
		panic(fmt.Sprintf("failed parsing time %s from date time format %s: %v", value, layout, err))
	} else {
		return parsedTime
	}
}

func NewHTTPRequest(method, url string, body io.Reader) *http.Request {
	if request, err := http.NewRequest(method, url, body); err != nil {
		panic(fmt.Sprintf("failed creating a new HTTP request: %v", err))
	} else {
		return request
	}
}

func NewJSONBObject(object any) types.JSONBObject {
	if object, err := types.NewJSONBObject(object); err != nil {
		panic(fmt.Sprintf("failed creating a new JSONBObject: %v", err))
	} else {
		return object
	}
}

func NewUUIDv4() uuid.UUID {
	if newUUID, err := uuid.NewV4(); err != nil {
		panic(fmt.Sprintf("failed generating a v4 UUID: %v", err))
	} else {
		return newUUID
	}
}

func NewNullUUIDv4() uuid.NullUUID {
	if newUUID, err := uuid.NewV4(); err != nil {
		panic(fmt.Sprintf("failed generating a v4 UUID: %v", err))
	} else {
		return uuid.NullUUID{
			UUID:  newUUID,
			Valid: true,
		}
	}
}

func GenerateX509StringPair(organization string) (string, string) {
	if cert, key, err := crypto.TLSGenerateStringPair(organization); err != nil {
		panic(fmt.Sprintf("failed generating a fresh x509 cert/key pair: %v", err))
	} else {
		return cert, key
	}
}
