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

package neo4j_test

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/specterops/bloodhound/dawgs/drivers/neo4j"
)

func TestValueMapper_MapOptions(t *testing.T) {
	mapper := neo4j.NewValueMapper([]any{1, 2, 3})

	var (
		intOption    int
		floatOption  float64
		stringOption string

		mappedPointer, err = mapper.MapOptions(&floatOption, &intOption)
		_, isIntPointer    = mappedPointer.(*int)
	)

	require.Nil(t, err)
	require.True(t, isIntPointer)
	require.Equal(t, 1, intOption)

	mappedPointer, err = mapper.MapOptions(&stringOption)

	require.Nil(t, mappedPointer)
	require.ErrorContains(t, err, "no matching target given")

}
