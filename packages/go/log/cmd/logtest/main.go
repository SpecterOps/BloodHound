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

package main

import (
	"github.com/specterops/bloodhound/log"
)

func main() {
	log.Infof("This is an info log message: %s", "test")
	log.Warnf("This is a warning log message: %s", "test")
	log.Errorf("This is a error log message: %s", "test")
	log.Fatalf("This is a fatal log message and will kill the application with exit 1: %s", "test")
	log.Errorf("This should never be seen, the Fatalf call is broken!")
}
