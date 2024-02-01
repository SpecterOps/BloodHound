// Copyright 2024 Specter Ops, Inc.
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

package ad

import (
	"context"
	"errors"

	"github.com/specterops/bloodhound/analysis"
	"github.com/specterops/bloodhound/analysis/impact"
	"github.com/specterops/bloodhound/dawgs/cardinality"
	"github.com/specterops/bloodhound/dawgs/graph"
	"github.com/specterops/bloodhound/dawgs/ops"
	"github.com/specterops/bloodhound/dawgs/query"
	"github.com/specterops/bloodhound/dawgs/util/channels"
	"github.com/specterops/bloodhound/graphschema/ad"
	"github.com/specterops/bloodhound/log"
)

func PostADCSESC9a(ctx context.Context, tx graph.Transaction, outC chan<- analysis.CreatePostRelationshipJob, groupExpansions impact.PathAggregator, eca, domain *graph.Node, cache ADCSCache) error {
	results := cardinality.NewBitmap32()

	if canAbuseWeakCertBindingRels, err := FetchCanAbuseWeakCertBindingRels(tx, eca); err != nil {
		if graph.IsErrNotFound(err) {
			return nil
		}

		return err
	} else if len(canAbuseWeakCertBindingRels) == 0 {
		return nil
	} else if publishedCertTemplates, ok := cache.PublishedTemplateCache[eca.ID]; !ok {
		return nil
	} else {
		for _, template := range publishedCertTemplates {
			if valid, err := isCertTemplateValidForESC9a(template); err != nil {
				if !errors.Is(err, graph.ErrPropertyNotFound) {
					log.Errorf("Error checking cert template validity for template %d: %v", template.ID, err)
				} else {
					log.Debugf("Error checking cert template validity for template %d: %v", template.ID, err)
				}
			} else if !valid {
				continue
			} else if certTemplateControllers, ok := cache.CertTemplateControllers[template.ID]; !ok {
				log.Debugf("Failed to retrieve controllers for cert template %d from cache", template.ID)
				continue
			} else if ecaControllers, ok := cache.EnterpriseCAEnrollers[eca.ID]; !ok {
				log.Debugf("Failed to retrieve controllers for enterprise ca %d from cache", eca.ID)
				continue
			} else {
				//Expand controllers for the eca + template completely because we don't do group shortcutting here
				var (
					victimBitmap = expandNodeSliceToBitmapWithoutGroups(certTemplateControllers, groupExpansions)
					ecaBitmap    = expandNodeSliceToBitmapWithoutGroups(ecaControllers, groupExpansions)
				)

				victimBitmap.And(ecaBitmap)
				//Use our id list to filter down to users
				if userNodes, err := ops.FetchNodeSet(tx.Nodes().Filterf(func() graph.Criteria {
					return query.And(
						query.KindIn(query.Node(), ad.User),
						query.InIDs(query.NodeID(), cardinality.DuplexToGraphIDs(victimBitmap)...),
					)
				})); err != nil {
					if !graph.IsErrNotFound(err) {
						return err
					}
				} else if len(userNodes) > 0 {
					if subjRequireDns, err := template.Properties.Get(ad.SubjectAltRequireDNS.String()).Bool(); err != nil {
						log.Debugf("Failed to retrieve subjectAltRequireDNS for template %d: %v", template.ID, err)
						victimBitmap.Xor(cardinality.NodeSetToDuplex(userNodes))
					} else if subjRequireDomainDns, err := template.Properties.Get(ad.SubjectAltRequireDomainDNS.String()).Bool(); err != nil {
						log.Debugf("Failed to retrieve subjectAltRequireDomainDNS for template %d: %v", template.ID, err)
						victimBitmap.Xor(cardinality.NodeSetToDuplex(userNodes))
					} else if subjRequireDns || subjRequireDomainDns {
						//If either of these properties is true, we need to remove all these users from our victims list
						victimBitmap.Xor(cardinality.NodeSetToDuplex(userNodes))
					}
				}

				if attackers, err := ops.FetchStartNodes(tx.Relationships().Filterf(func() graph.Criteria {
					return query.And(
						query.KindIn(query.Start(), ad.Group, ad.User, ad.Computer),
						query.KindIn(query.Relationship(), ad.GenericAll, ad.GenericWrite, ad.Owns, ad.WriteOwner, ad.WriteDACL),
						query.InIDs(query.EndID(), cardinality.DuplexToGraphIDs(victimBitmap)...),
					)
				})); err != nil {
					return err
				} else {
					results.Or(cardinality.NodeSetToDuplex(attackers))
				}
			}
		}

		results.Each(func(value uint32) bool {
			if !channels.Submit(ctx, outC, analysis.CreatePostRelationshipJob{
				FromID: graph.ID(value),
				ToID:   domain.ID,
				Kind:   ad.ADCSESC9a,
			}) {
				return false
			} else {
				return true
			}
		})

		return nil
	}
}

func expandNodeSliceToBitmapWithoutGroups(nodes []*graph.Node, groupExpansions impact.PathAggregator) cardinality.Duplex[uint32] {
	var bitmap = cardinality.NewBitmap32()
	for _, controller := range nodes {
		if controller.Kinds.ContainsOneOf(ad.Group) {
			groupExpansions.Cardinality(controller.ID.Uint32()).(cardinality.Duplex[uint32]).Each(func(id uint32) bool {
				//Check group expansions against each id, if cardinality is 0 than its not a group
				if groupExpansions.Cardinality(id).Cardinality() == 0 {
					bitmap.Add(id)
				}

				return true
			})
		} else {
			bitmap.Add(controller.ID.Uint32())
		}
	}

	return bitmap
}

func isCertTemplateValidForESC9a(ct *graph.Node) (bool, error) {
	if reqManagerApproval, err := ct.Properties.Get(ad.RequiresManagerApproval.String()).Bool(); err != nil {
		return false, err
	} else if reqManagerApproval {
		return false, nil
	} else if authenticationEnabled, err := ct.Properties.Get(ad.AuthenticationEnabled.String()).Bool(); err != nil {
		return false, err
	} else if !authenticationEnabled {
		return false, nil
	} else if noSecurityExtension, err := ct.Properties.Get(ad.NoSecurityExtension.String()).Bool(); err != nil {
		return false, err
	} else if !noSecurityExtension {
		return false, nil
	} else if enrolleeSuppliesSubject, err := ct.Properties.Get(ad.EnrolleeSuppliesSubject.String()).Bool(); err != nil {
		return false, err
	} else if enrolleeSuppliesSubject {
		return false, nil
	} else if schemaVersion, err := ct.Properties.Get(ad.SchemaVersion.String()).Float64(); err != nil {
		return false, err
	} else if authorizedSignatures, err := ct.Properties.Get(ad.AuthorizedSignatures.String()).Float64(); err != nil {
		return false, err
	} else if schemaVersion > 1 && authorizedSignatures > 0 {
		return false, nil
	} else if subjectAltRequireUPN, err := ct.Properties.Get(ad.SubjectAltRequireUPN.String()).Bool(); err != nil {
		return false, err
	} else if subjectAltRequireSPN, err := ct.Properties.Get(ad.SubjectAltRequireSPN.String()).Bool(); err != nil {
		return false, err
	} else if subjectAltRequireSPN || subjectAltRequireUPN {
		return true, nil
	} else {
		return false, nil
	}
}
