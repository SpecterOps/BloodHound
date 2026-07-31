// Copyright 2026 Specter Ops, Inc.
//
// Licensed under the Apache License, Version 2.0
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//	http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//
// SPDX-License-Identifier: Apache-2.0

package handlers

// ExtensionDetailsView is the JSON shape of the extension details embedded in a
// NodeKindView.
type ExtensionDetailsView struct {
	ExtensionID int32  `json:"extension_id"`
	Name        string `json:"name"`
	DisplayName string `json:"display_name"`
	Namespace   string `json:"namespace"`
	Version     string `json:"version"`
}
