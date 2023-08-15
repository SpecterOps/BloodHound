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

package lab

import (
	"fmt"
	"os"
	"strings"
)

const logPrefix = "[LAB_DEBUG] "

var debugMode = false

func init() {
	if strings.ToLower(os.Getenv("LAB_DEBUG")) == "true" {
		debugMode = true
	}
}

func log(message string) {
	fmt.Printf("%s%s\n", logPrefix, message)
}

func logf(format string, a ...any) {
	log(fmt.Sprintf(format, a...))
}

func logTopology(topology []any) {
	if debugMode {
		log("Fixture Topology")
		numFixtures := len(topology)
		for i := numFixtures - 1; i >= 0; i-- {
			fixture := topology[i]
			logf("%d: %T", numFixtures-i, fixture)
		}
	}
}
