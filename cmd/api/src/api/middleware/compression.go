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

package middleware

import (
	"compress/gzip"
	"compress/zlib"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/specterops/bloodhound/headers"
	"github.com/specterops/bloodhound/log"
	"github.com/specterops/bloodhound/src/api"
)

var errUnsupportedEncoding = errors.New("content encoding is not supported")

type GzipResponseWriter struct {
	http.ResponseWriter
	gw *gzip.Writer
}

func NewGzipResponseWriter(w http.ResponseWriter) *GzipResponseWriter {
	return &GzipResponseWriter{
		ResponseWriter: w,
		gw:             gzip.NewWriter(w),
	}
}

func (s *GzipResponseWriter) Write(p []byte) (int, error) {
	return s.gw.Write(p)
}

func (s *GzipResponseWriter) Close() error {
	return s.gw.Close()
}

func CompressionMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(responseWriter http.ResponseWriter, request *http.Request) {
		var (
			gw  *GzipResponseWriter
			err error
		)

		if contentEncodingString := strings.Join(request.Header.Values(headers.ContentEncoding.String()), ","); contentEncodingString != "" { // "Content-Encoding: gzip, deflate; Content-Encoding: br;" = "gzip, deflate, br"
			for _, encoding := range strings.Split(contentEncodingString, ",") {
				encoding = strings.TrimSpace(encoding)
				request.Body, err = wrapBody(encoding, request.Body)
				if err != nil {
					errMsg := fmt.Sprintf("failed to create reader for %s encoding: %v", encoding, err)
					log.Warnf(errMsg)
					if errors.Is(err, errUnsupportedEncoding) {
						api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusUnsupportedMediaType, fmt.Sprintf("Error trying to read request: %s", errMsg), request), responseWriter)
					} else {
						api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusBadRequest, fmt.Sprintf("Error trying to read request: %s", errMsg), request), responseWriter)
					}
					return
				}
			}
		}

		if acceptEncodingString := strings.Join(request.Header.Values(headers.AcceptEncoding.String()), ","); acceptEncodingString != "" {
			// For simplicity, we will only honor a "gzip-or-not" compression strategy, without regard to quality values.
			// In the future we *may* choose to support Accept-Encoding quality values:
			// https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/Accept-Encoding
			// https://developer.mozilla.org/en-US/docs/Glossary/Quality_values
			if strings.Contains(acceptEncodingString, "gzip") {
				gw = NewGzipResponseWriter(responseWriter)
			}
		}

		if gw != nil {
			responseWriter = gw
			responseWriter.Header().Set(headers.ContentEncoding.String(), "gzip")
			defer gw.Close()
		}
		next.ServeHTTP(responseWriter, request)
	})
}

func wrapBody(encoding string, body io.ReadCloser) (io.ReadCloser, error) {
	var (
		newBody = body
		err     error
	)
	switch encoding {
	case "gzip", "x-gzip":
		newBody, err = gzip.NewReader(body)
	case "deflate":
		newBody, err = zlib.NewReader(body)
	default:
		log.Infof("unsupported encoding detected: %s", encoding)
		err = errUnsupportedEncoding
	}
	return newBody, err
}
