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
package model_test

import (
	"testing"

	"github.com/specterops/bloodhound/cmd/api/src/model"
	"github.com/specterops/dawgs/graph"
)

func TestAssetGroupTagKindNameUsesPersistedKind(t *testing.T) {
	tag := model.AssetGroupTag{
		Name:             "T0 - CONTOSO PROD",
		KindNameOverride: "Tag_T0_-_CONTOSO_PROD_[tenant-id]",
	}

	if got, want := tag.ToKind(), graph.StringKind(tag.KindNameOverride); got != want {
		t.Fatalf("ToKind() = %s, want %s", got, want)
	}

	tag.KindNameOverride = ""
	if got, want := tag.KindName(), "Tag_T0_-_CONTOSO_PROD"; got != want {
		t.Fatalf("KindName() = %s, want %s", got, want)
	}
}
