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

package headers

//go:generate go run cmd/generate.go

type Header string

func (s Header) String() string {
	return string(s)
}

// Non-standard headers
const (
	RequestDate Header = "RequestDate"
	RequestID   Header = "RequestID"
	Signature   Header = "Signature" // https://www.ietf.org/archive/id/draft-ietf-httpbis-message-signatures-04.html#name-the-signature-http-header
)
