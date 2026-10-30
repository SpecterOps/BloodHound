// Copyright 2026 Specter Ops, Inc.
//
// Licensed under the Apache License, Version 2.0
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//	http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//
// SPDX-License-Identifier: Apache-2.0

package duration_test

import (
	"testing"
	"time"

	"github.com/specterops/bloodhound/packages/go/duration"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUnmarshalJson(t *testing.T) {
	cases := []struct {
		input  string
		output time.Duration
	}{
		{
			input:  `"PT5H"`,
			output: 5 * time.Hour,
		},
		{
			input:  `"PT5H"`,
			output: 5 * time.Hour,
		},
		{
			input:  `"P2D"`,
			output: 48 * time.Hour,
		},
		{
			input:  `"PT1H30M"`,
			output: time.Hour + 30*time.Minute,
		},
		{
			input:  `"P1DT2H30M"`,
			output: 24*time.Hour + 2*time.Hour + 30*time.Minute,
		},
		{
			input:  `"P3DT4H15M"`,
			output: 3*24*time.Hour + 4*time.Hour + 15*time.Minute,
		},
		{
			input:  `"P1DT0H45M"`,
			output: 24*time.Hour + 45*time.Minute,
		},
	}

	for _, cc := range cases {
		var d duration.ISODuration
		err := d.UnmarshalJSON([]byte(cc.input))
		require.NoError(t, err)
		assert.Equal(t, duration.ISODuration(cc.output), d)
	}
}

func TestMarshalJson(t *testing.T) {
	cases := []struct {
		input  time.Duration
		output string
	}{
		{
			input:  5 * time.Hour,
			output: `"PT5H"`,
		},
		{
			input:  2 * 24 * time.Hour,
			output: `"P2D"`,
		},
		{
			input:  time.Hour + 30*time.Minute,
			output: `"PT1H30M"`,
		},
		{
			input:  24*time.Hour + 2*time.Hour + 30*time.Minute,
			output: `"P1DT2H30M"`,
		},
		{
			input:  3*24*time.Hour + 4*time.Hour + 15*time.Minute,
			output: `"P3DT4H15M"`,
		},
		{
			input:  24*time.Hour + 45*time.Minute,
			output: `"P1DT45M"`,
		},
	}

	for _, cc := range cases {
		d := duration.ISODuration(cc.input)
		output, err := d.MarshalJSON()
		require.NoError(t, err)
		assert.Equal(t, cc.output, string(output))
	}
}
