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

package params

import (
	"fmt"
	"net/http"
	"regexp"

	"github.com/gorilla/mux"
)

// Query parameters
var (
	StartNode         = newParam("start_node", nil)
	EndNode           = newParam("end_node", nil)
	RelationshipKinds = newParam("relationship_kinds", containsPredicate)
)

// param is an immutable path or query parameter
type param struct {
	name  string
	regex *regexp.Regexp
}

func (s param) String() string {
	return s.name
}

func (s param) Regexp() *regexp.Regexp {
	return s.regex
}

func (s param) RouteMatcher() string {
	if s.regex == nil {
		return fmt.Sprintf("{%s}", s.name)
	} else {
		return fmt.Sprintf("{%s:%s}", s.name, s.regex.String())
	}
}

func newParam(name string, regexp *regexp.Regexp) param {
	return param{
		name:  name,
		regex: regexp,
	}
}

func GetPathVariables(request *http.Request) map[string]string {
	return mux.Vars(request)
}
