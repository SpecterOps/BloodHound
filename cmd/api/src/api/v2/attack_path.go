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

package v2

import (
	"time"
)

const (
	ErrorParseParams        = "unable to parse request parameters"
	ErrorDecodeParams       = "unable to decode request parameters"
	ErrorNoDomainId         = "no domain id specified in url"
	ErrorNoFindingType      = "no finding type specified"
	ErrorInvalidFindingType = "invalid finding type specified: %v"
	ErrorInvalidRFC3339     = "invalid RFC-3339 datetime format: %v"
)

type RiskAcceptRequest struct {
	RiskType    string    `json:"risk_type"`
	AcceptUntil time.Time `json:"accept_until"`
	Accepted    bool      `json:"accepted"` // DEPRECATED remove this field for V3
}
