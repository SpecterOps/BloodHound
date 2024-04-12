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
	"hash"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/specterops/bloodhound/headers"
)

const ErrorTemplateHMACSignature string = "unable to compute hmac signature: %w"

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

func NewSelfDestructingTempFile(dir, prefix string) (*SelfDestructingTempFile, error) {
	if file, err := os.CreateTemp(dir, prefix); err != nil {
		return nil, err
	} else {
		return &SelfDestructingTempFile{
			file:      file,
			destroyed: false,
		}, err
	}
}

type SelfDestructingTempFile struct {
	file      *os.File
	destroyed bool
}

func (s *SelfDestructingTempFile) SelfDestruct() error {
	if !s.destroyed {
		name := s.file.Name()

		if err := s.file.Close(); err != nil {
			return err
		} else if err := os.Remove(name); err != nil && !os.IsNotExist(err) {
			return err
		}
		s.destroyed = true
	}
	return nil
}

func (s *SelfDestructingTempFile) Read(b []byte) (int, error) {
	n, err := s.file.Read(b)
	if err == io.EOF {
		s.SelfDestruct()
	}
	return n, err
}

func (s *SelfDestructingTempFile) Write(b []byte) (int, error) {
	return s.file.Write(b)
}

func (s *SelfDestructingTempFile) Close() error {
	if !s.destroyed {
		return s.SelfDestruct()
	} else {
		return s.file.Close()
	}
}

func (s *SelfDestructingTempFile) Seek(offset int64, whence int) (int64, error) {
	return s.file.Seek(offset, whence)
}

func (s *SelfDestructingTempFile) Name() string {
	return s.file.Name()
}

// NewRequestSignature generates the BloodHound request signature using the provided hash function.
// NOTE: The given io.Reader will be read to EOF. Consider using io.TeeReader so that the body may be read again after the signature has been created.
func NewRequestSignature(hasher func() hash.Hash, key string, datetime string, requestMethod string, requestURI string, body io.Reader) ([]byte, error) {
	if hasher == nil {
		return nil, fmt.Errorf(ErrorTemplateHMACSignature, fmt.Errorf("hasher must not be nil"))
	}

	digester := hmac.New(hasher, []byte(key))

	// OperationKey is the first HMAC digest link in the signature chain. This prevents replay attacks that seek to
	// modify the request method or URI. It is composed of concatenating the request method and the request URI with
	// no delimiter and computing the HMAC digest using the token key as the digest secret.
	//
	// Example: GET /api/v2/test/resource HTTP/1.1
	// Signature Component: GET/api/v2/test/resource
	if _, err := digester.Write([]byte(requestMethod + requestURI)); err != nil {
		return nil, fmt.Errorf(ErrorTemplateHMACSignature, err)
	}

	// DateKey is the next HMAC digest link in the signature chain. This encodes the RFC3339 formatted datetime
	// value as part of the signature to the hour to prevent replay attacks that are older than max two hours. This
	// value is added to the signature chain by cutting off all values from the RFC3339 formatted datetime from the
	// hours value forward:
	//
	// Example: 2020-12-01T23:59:60Z
	// Signature Component: 2020-12-01T23
	digester = hmac.New(hasher, digester.Sum(nil))

	if _, err := digester.Write([]byte(datetime[:13])); err != nil {
		return nil, fmt.Errorf(ErrorTemplateHMACSignature, err)
	}

	// Body signing is the last HMAC digest link in the signature chain. This encodes the request body as part of
	// the signature to prevent replay attacks that seek to modify the payload of a signed request. In the case
	// where there is no body content the HMAC digest is computed anyway, simply with no values written to the
	// digester.
	digester = hmac.New(hasher, digester.Sum(nil))

	if body != nil {
		if _, err := io.Copy(digester, body); err != nil {
			return nil, fmt.Errorf(ErrorTemplateHMACSignature, err)
		}
	}

	return digester.Sum(nil), nil
}

// SignRequestAtTime signs a given HTTP request using the BHE request signature scheme. The passed-in time value is used
// for the DateKey portion of the signature digest.
func SignRequestAtTime(hasher func() hash.Hash, id string, token string, datetime time.Time, request *http.Request) error {
	datetimeFormatted := datetime.Format(time.RFC3339)
	var (
		buffer bytes.Buffer
		tee    io.Reader
	)

	if request.Body != nil {
		tee = io.TeeReader(request.Body, &buffer)
	}

	if signature, err := NewRequestSignature(hasher, token, datetimeFormatted, request.Method, request.URL.Path, tee); err != nil {
		return err
	} else {
		// Overwrite the request body reader if the request body wasn't nil
		if request.Body != nil {
			request.Body = io.NopCloser(&buffer)
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
	return SignRequestAtTime(sha256.New, tokenID, token, time.Now(), request)
}
