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

package test

import (
	"testing"
	"time"

	"github.com/specterops/bloodhound/dawgs/graph"
	"github.com/stretchr/testify/require"
)

func RequireProperty[T any](t *testing.T, expected T, actual graph.PropertyValue, msg ...any) {
	var (
		value = actual.Any()
		err   error
	)

	switch any(expected).(type) {
	case time.Time:
		value, err = actual.Time()
	}

	require.Nil(t, err, msg...)
	require.Equal(t, expected, value, msg...)
}
