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

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/specterops/bloodhound/src/auth"
	"github.com/specterops/bloodhound/headers"
)

// tee takes a source reader and two writers. The function reads from the source until exhaustion. Each read is written
// serially to both writers.
func tee(reader io.Reader, outA, outB io.Writer) error {
	// Ignore readers that are nil to begin with. This covers the case where a request is being signed but contains
	// no body.
	if reader == nil {
		return nil
	}

	// Internal read buffer for splitting out to the other writers
	buffer := make([]byte, 4096)

	for {
		read, err := reader.Read(buffer)

		if read > 0 {
			if _, err := outA.Write(buffer[:read]); err != nil {
				return err
			}

			if _, err := outB.Write(buffer[:read]); err != nil {
				return err
			}
		}

		if err != nil {
			if err != io.EOF {
				return err
			}

			return nil
		}
	}
}

func GenerateRequestSignature(token, datetime, hmacMethod, requestMethod, requestURI string, body io.Reader) (io.Reader, []byte, error) {
	if hmacMethod != auth.HMAC_SHA2_256 {
		return nil, nil, fmt.Errorf("unsupported HMAC method: %s", hmacMethod)
	}

	// OperationKey is the first HMAC digest link in the signature chain. This prevents replay attacks that seek to
	// modify the request method or URI. It is composed of concatenating the request method and the request URI with
	// no delimiter and computing the HMAC digest using the token key as the digest secret.
	//
	// Example: GET /api/v2/test/resource HTTP/1.1
	// Signature Component: GET/api/v2/test/resource
	digester := hmac.New(sha256.New, []byte(token))

	if _, err := digester.Write([]byte(requestMethod + requestURI)); err != nil {
		return nil, nil, err
	}

	// DateKey is the next HMAC digest link in the signature chain. This encodes the RFC3339 formatted datetime
	// value as part of the signature to the hour to prevent replay attacks that are older than max two hours. This
	// value is added to the signature chain by cutting off all values from the RFC3339 formatted datetime from the
	// hours value forward:
	//
	// Example: 2020-12-01T23:59:60Z
	// Signature Component: 2020-12-01T23
	digester = hmac.New(sha256.New, digester.Sum(nil))

	if _, err := digester.Write([]byte(datetime[:13])); err != nil {
		return nil, nil, err
	}

	// Body signing is the last HMAC digest link in the signature chain. This encodes the request body as part of
	// the signature to prevent replay attacks that seek to modify the payload of a signed request. In the case
	// where there is no body content the HMAC digest is computed anyway, simply with no values written to the
	// digester.
	digester = hmac.New(sha256.New, digester.Sum(nil))

	readBody := &bytes.Buffer{}
	if err := tee(body, readBody, digester); err != nil {
		return nil, nil, err
	}

	return readBody, digester.Sum(nil), nil
}

// SignRequestAtTime signs a given HTTP request using the BHE request signature scheme. The passed-in time value is used
// for the DateKey portion of the signature digest.
func SignRequestAtTime(id, token, hmacMethod string, datetime time.Time, request *http.Request) error {
	datetimeFormatted := datetime.Format(time.RFC3339)

	if readBody, signature, err := GenerateRequestSignature(token, datetimeFormatted, hmacMethod, request.Method, request.URL.Path, request.Body); err != nil {
		return err
	} else {
		// Overwrite the request body reader if the request body wasn't nil
		if request.Body != nil {
			request.Body = io.NopCloser(readBody)
		}

		// Set the request headers
		request.Header.Set(headers.Authorization.String(), fmt.Sprintf("%s %s", AuthorizationSchemeBHESignature, id))
		request.Header.Set(headers.RequestDate.String(), datetimeFormatted)
		request.Header.Set(headers.Signature.String(), base64.StdEncoding.EncodeToString(signature))
	}

	return nil
}

// SignRequest signs a given HTTP request using the BHE request signature scheme. Note: signatures are time-sensitive and
// may only be valid for a maximum period of 2 hours.
func SignRequest(tokenID, token string, request *http.Request) error {
	return SignRequestAtTime(tokenID, token, auth.HMAC_SHA2_256, time.Now(), request)
}

type readerDelegatedCloser struct {
	source io.Reader
	closer io.Closer
}

func (s readerDelegatedCloser) Read(p []byte) (n int, err error) {
	return s.source.Read(p)
}

func (s readerDelegatedCloser) Close() error {
	return s.closer.Close()
}
