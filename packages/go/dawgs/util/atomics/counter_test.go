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

package atomics_test

import (
	"sync"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/specterops/bloodhound/dawgs/util/atomics"
)

func TestNewCounterToMaximum(t *testing.T) {
	t.Run("Count to 0", func(t *testing.T) {
		var (
			counter    = atomics.NewCounter[uint32](0)
			iterations = 0
		)

		for !counter() {
			iterations++
		}

		require.Equal(t, 0, iterations)
	})

	t.Run("Count to 10", func(t *testing.T) {
		var (
			counter    = atomics.NewCounter[uint32](10)
			iterations = 0
		)

		for !counter() {
			iterations++
		}

		require.Equal(t, 10, iterations)
	})

	t.Run("10 goroutines counting to 10240", func(t *testing.T) {
		var (
			counter1  = atomics.NewCounter[uint32](10240)
			counter2  = atomics.NewCounter[uint32](10240)
			waitGroup = &sync.WaitGroup{}
		)

		for workerID := 0; workerID < 10; workerID++ {
			waitGroup.Add(1)

			go func() {
				defer waitGroup.Done()

				for !counter1() {
					if counter2() {
						t.Errorf("Check counter reached limit")
						return
					}
				}
			}()
		}

		waitGroup.Wait()
	})
}
