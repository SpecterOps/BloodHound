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

package errors

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestError_Error(t *testing.T) {
	input := Error("test error")
	output := input.Error()

	// output should have the same value and be a string type
	require.Equal(t, "test error", output)
	require.Equal(t, "string", reflect.TypeOf(output).Name())
}

func TestError_New(t *testing.T) {
	output := New("test error")

	// output should have the same value and be an Error type
	require.Equal(t, "test error", output.Error())
	require.Equal(t, "Error", reflect.TypeOf(output).Name())
}

func TestError_Is(t *testing.T) {
	errTest := fmt.Errorf("test error")
	err := errTest
	require.True(t, Is(err, errTest))
}

func TestError_As(t *testing.T) {
	err := fmt.Errorf("test error")
	errTest2 := fmt.Errorf("this should go away")

	require.True(t, As(err, &errTest2))
	require.Equal(t, err, errTest2)
}
