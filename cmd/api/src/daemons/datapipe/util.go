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

package datapipe

import (
	"math/rand"
	"time"
)

func RandomDurationBetween(min, max time.Duration) time.Duration {
	r := rand.New(rand.NewSource(time.Now().Unix()))

	durationRange := max - min
	return min + time.Duration(r.Int63())%durationRange
}
