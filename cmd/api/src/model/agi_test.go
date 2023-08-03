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

package model

import (
	"testing"
)

func TestAssetGroupSelectorSpec_Validate(t *testing.T) {
	//happyPath := AssetGroupSelectorSpec{
	//	SelectorName:   "test",
	//	EntityObjectID: "S-1-5-21-570004220-2248230615-4072641716-544",
	//	Action:         SelectorSpecActionAdd,
	//}
	//
	//assert.Nil(t, happyPath.Validate(), "Expected valid AG selector spec to validate")
	//
	//bustedEntityObjectID := AssetGroupSelectorSpec{
	//	SelectorName:   "troll",
	//	EntityObjectID: "I CONTROL YOU",
	//	Action:         SelectorSpecActionRemove,
	//}
	//
	//assert.Equalf(t, bustedEntityObjectID.Validate(), ErrEntityObjectIDInvalid, "Expected bad object ID to fail validation")
	//
	//bustedAction := AssetGroupSelectorSpec{
	//	SelectorName:   "dayman",
	//	EntityObjectID: "S-1-5-21-570004220-2248230615-4072641716-544",
	//	Action:         "foobar",
	//}
	//
	//assert.Equalf(t, bustedAction.Validate(), ErrActionInvalid, "Expected bad node label to fail validation")
}
