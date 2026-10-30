// Copyright 2026 Specter Ops, Inc.
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

package generator

import "testing"

func TestDMSADisplayMetadata(t *testing.T) {
	metadata, found := nodeIcons["DMSA"]
	if !found {
		t.Fatal("DMSA display metadata is not registered")
	}

	if metadata.Icon == "" || metadata.Color == "" {
		t.Fatalf("DMSA display metadata is incomplete: %#v", metadata)
	}
}
