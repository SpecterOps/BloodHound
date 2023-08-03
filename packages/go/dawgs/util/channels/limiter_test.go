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

package channels

import (
	"context"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

func TestConcurrencyLimiter(t *testing.T) {
	var (
		limiter = NewConcurrencyLimiter(2)
	)

	require.True(t, limiter.Acquire(context.Background()))
	require.True(t, limiter.Acquire(context.Background()))

	timeoutCtx, done := context.WithTimeout(context.Background(), time.Millisecond)
	defer done()

	require.False(t, limiter.Acquire(timeoutCtx))
}
