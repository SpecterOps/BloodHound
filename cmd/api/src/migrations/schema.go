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

package migrations

import (
	"github.com/specterops/bloodhound/dawgs/graph"
	"github.com/specterops/bloodhound/graphschema/ad"
	"github.com/specterops/bloodhound/graphschema/azure"
	"github.com/specterops/bloodhound/graphschema/common"
)

func CurrentSchema() *graph.Schema {
	bhSchema := graph.NewSchema()

	bhSchema.DefineKinds(ad.NodeKinds()...)
	bhSchema.DefineKinds(azure.NodeKinds()...)

	bhSchema.ConstrainProperty(common.ObjectID.String(), graph.FullTextSearchIndex)

	bhSchema.IndexProperty(common.Name.String(), graph.FullTextSearchIndex)
	bhSchema.IndexProperty(common.SystemTags.String(), graph.FullTextSearchIndex)
	bhSchema.IndexProperty(common.UserTags.String(), graph.FullTextSearchIndex)

	bhSchema.ForKinds(ad.Entity).Index(ad.DistinguishedName.String(), graph.BTreeIndex)

	bhSchema.ForKinds(ad.NodeKinds()...).
		Index(ad.DomainFQDN.String(), graph.BTreeIndex).
		Index(ad.DomainSID.String(), graph.BTreeIndex)

	bhSchema.ForKinds(azure.NodeKinds()...).
		Index(azure.TenantID.String(), graph.BTreeIndex)

	bhSchema.ForKinds(ad.RootCA, ad.EnterpriseCA, ad.AIACA).
		Index(ad.CertThumbprint.String(), graph.BTreeIndex)

	return bhSchema
}
