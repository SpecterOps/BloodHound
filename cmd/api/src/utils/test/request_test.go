// Copyright 2026 Specter Ops, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//
// SPDX-License-Identifier: Apache-2.0

package test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type responseBodyTestingT struct {
	failed bool
}

func (s *responseBodyTestingT) Errorf(string, ...any) {
	s.failed = true
}

func TestNormalizeResponseJSONBody(t *testing.T) {
	const (
		errorResponseBody = `{"http_status":400,"timestamp":"2026-10-07T12:34:56Z","errors":[{"message":"invalid request"}]}`
		expectedBody      = `{"http_status":400,"timestamp":"0001-01-01T00:00:00Z","errors":[{"message":"invalid request"}]}`
	)

	actualBody := normalizeResponseJSONBody(errorResponseBody)

	require.JSONEq(t, expectedBody, actualBody)
}

func TestJSONEq(t *testing.T) {
	t.Run("matches nested values regardless of object key order", func(t *testing.T) {
		assertionT := &responseBodyTestingT{}

		matches := assert.JSONEq(
			assertionT,
			`{"users":[{"id":1,"roles":["admin","user"]}]}`,
			`{"users":[{"roles":["admin","user"],"id":1}]}`,
		)

		require.True(t, matches)
		require.False(t, assertionT.failed)
	})

	t.Run("detects a changed nested value", func(t *testing.T) {
		assertionT := &responseBodyTestingT{}

		matches := assert.JSONEq(
			assertionT,
			`{"users":[{"id":1,"roles":["admin","user"]}]}`,
			`{"users":[{"id":2,"roles":["admin","user"]}]}`,
		)

		require.False(t, matches)
		require.True(t, assertionT.failed)
	})

	t.Run("detects a changed array order", func(t *testing.T) {
		assertionT := &responseBodyTestingT{}

		matches := assert.JSONEq(
			assertionT,
			`{"roles":["admin","user"]}`,
			`{"roles":["user","admin"]}`,
		)

		require.False(t, matches)
		require.True(t, assertionT.failed)
	})

	t.Run("detects invalid actual JSON", func(t *testing.T) {
		assertionT := &responseBodyTestingT{}

		matches := assert.JSONEq(assertionT, `{"id":1}`, `not json`)

		require.False(t, matches)
		require.True(t, assertionT.failed)
	})
}
