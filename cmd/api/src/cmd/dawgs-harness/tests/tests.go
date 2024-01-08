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

package tests

import (
	"fmt"
	"strconv"

	"github.com/specterops/bloodhound/dawgs/graph"
	"github.com/specterops/bloodhound/dawgs/ops"
	"github.com/specterops/bloodhound/dawgs/query"
	"github.com/specterops/bloodhound/graphschema/ad"
	"github.com/specterops/bloodhound/graphschema/common"
)

var (
	SimpleRelationshipsToCreate int
	StartNodeIDs                []graph.ID
	EndNodeIDs                  []graph.ID
	RelationshipIDs             []graph.ID
)

func validateFetches(numExpectedFetches, resultsFetched int) error {
	if resultsFetched != numExpectedFetches {
		return fmt.Errorf("expected to fetch %d results but only fetched %d", numExpectedFetches, resultsFetched)
	}

	return nil
}

func FetchNodesByID(testCase *TestCase) any {
	return func(tx graph.Transaction) error {
		for _, nodeID := range append(StartNodeIDs, EndNodeIDs...) {
			if err := testCase.Sample(func() error {
				if node, err := ops.FetchNode(tx, nodeID); err != nil {
					return err
				} else {
					_, err := node.Properties.Get(common.ObjectID.String()).String()
					return err
				}
			}); err != nil {
				return err
			}
		}

		return nil
	}
}

func FetchNodesByProperty(propertyName string, maxFetches int) func(testCase *TestCase) any {
	return func(testCase *TestCase) any {
		return func(tx graph.Transaction) error {
			resultsFetched := 0

			for iteration := 0; iteration < maxFetches; iteration++ {
				propertyValue := "batch start node " + strconv.Itoa(iteration)

				if iteration == 0 {
					testCase.EnableAnalysis()
				} else {
					testCase.DisableAnalysis()
				}

				if err := testCase.Sample(func() error {
					return tx.Nodes().Filterf(func() graph.Criteria {
						return query.And(
							query.Kind(query.Node(), ad.Entity),
							query.Equals(query.NodeProperty(propertyName), propertyValue),
						)
					}).Fetch(func(cursor graph.Cursor[*graph.Node]) error {
						for node := range cursor.Chan() {
							if actualPropertyValue, err := node.Properties.Get(common.Name.String()).String(); err != nil {
								return err
							} else if propertyValue != actualPropertyValue {
								return fmt.Errorf("expected node name to be %s but got %s", propertyValue, actualPropertyValue)
							}

							resultsFetched++
						}

						return cursor.Error()
					})
				}); err != nil {
					return err
				}
			}

			return validateFetches(maxFetches, resultsFetched)
		}
	}
}

func FetchNodesByPropertySlice(propertyName string) func(testCase *TestCase) any {
	return func(testCase *TestCase) any {
		return func(tx graph.Transaction) error {
			var (
				numExpectedFetches = len(StartNodeIDs)
				resultsFetched     = 0
				propertyValues     = make([]string, numExpectedFetches)
			)

			for iteration := 0; iteration < len(StartNodeIDs); iteration++ {
				propertyValues[iteration] = "start node " + strconv.Itoa(iteration)
			}

			if err := testCase.Sample(func() error {
				return tx.Nodes().Filterf(func() graph.Criteria {
					return query.And(
						query.Kind(query.Node(), ad.Entity),
						query.In(query.NodeProperty(propertyName), propertyValues),
					)
				}).Fetch(func(cursor graph.Cursor[*graph.Node]) error {
					for node := range cursor.Chan() {
						if _, err := node.Properties.Get(common.Name.String()).String(); err != nil {
							return err
						}

						resultsFetched++
					}

					return cursor.Error()
				})
			}); err != nil {
				return err
			}

			return validateFetches(numExpectedFetches, resultsFetched)
		}
	}
}

func FetchRelationshipByStartNodeProperty(testCase *TestCase) any {
	return func(tx graph.Transaction) error {
		var (
			numExpectedFetches = len(RelationshipIDs)
			resultsFetched     = 0
		)

		for iteration := 0; iteration < numExpectedFetches; iteration++ {
			var (
				iterationStr     = strconv.Itoa(iteration)
				nodeName         = "batch start node " + iterationStr
				relationshipName = "batch relationship " + iterationStr
			)

			if err := testCase.Sample(func() error {
				return tx.Relationships().Filterf(func() graph.Criteria {
					return query.And(
						query.Kind(query.Start(), ad.Entity),
						query.Equals(query.StartProperty(common.Name.String()), nodeName),
					)
				}).Fetch(func(cursor graph.Cursor[*graph.Relationship]) error {
					for relationship := range cursor.Chan() {
						if actualRelationshipName, err := relationship.Properties.Get(common.Name.String()).String(); err != nil {
							return err
						} else if relationshipName != actualRelationshipName {
							return fmt.Errorf("expected relationship name to be %s but got %s", relationshipName, actualRelationshipName)
						}

						resultsFetched++
					}

					return cursor.Error()
				})
			}); err != nil {
				return err
			}
		}

		return validateFetches(numExpectedFetches, resultsFetched)
	}
}

func FetchDirectionalResultByStartNodeProperty(testCase *TestCase) any {
	return func(tx graph.Transaction) error {
		var (
			numExpectedFetches = len(RelationshipIDs)
			resultsFetched     = 0
		)

		for iteration := 0; iteration < numExpectedFetches; iteration++ {
			if iteration == 0 {
				testCase.EnableAnalysis()
			} else {
				testCase.DisableAnalysis()
			}

			var (
				iterationStr     = strconv.Itoa(iteration)
				nodeName         = "batch start node " + iterationStr
				relationshipName = "batch relationship " + iterationStr
			)

			if err := testCase.Sample(func() error {
				return tx.Relationships().Filterf(func() graph.Criteria {
					return query.And(
						query.Kind(query.Start(), ad.Entity),
						query.Equals(query.StartProperty(common.Name.String()), nodeName),
					)
				}).FetchDirection(graph.DirectionOutbound, func(cursor graph.Cursor[graph.DirectionalResult]) error {
					for directionalResult := range cursor.Chan() {
						if actualRelationshipName, err := directionalResult.Relationship.Properties.Get(common.Name.String()).String(); err != nil {
							return err
						} else if relationshipName != actualRelationshipName {
							return fmt.Errorf("expected relationship name to be %s but got %s", relationshipName, actualRelationshipName)
						}

						if actualNodeName, err := directionalResult.Node.Properties.Get(common.Name.String()).String(); err != nil {
							return err
						} else if nodeName != actualNodeName {
							return fmt.Errorf("expected node name to be %s but got %s", nodeName, actualNodeName)
						}

						resultsFetched++
					}

					return cursor.Error()
				})
			}); err != nil {
				return err
			}
		}

		return validateFetches(numExpectedFetches, resultsFetched)
	}
}

func FetchRelationshipsByPropertySlice(testCase *TestCase) any {
	return func(tx graph.Transaction) error {
		var (
			numExpectedFetches = len(RelationshipIDs)
			resultsFetched     = 0
			relationshipNames  = make([]string, numExpectedFetches)
		)

		for iteration := 0; iteration < numExpectedFetches; iteration++ {
			relationshipNames[iteration] = "relationship " + strconv.Itoa(iteration)
		}

		if err := testCase.Sample(func() error {
			return tx.Relationships().Filterf(func() graph.Criteria {
				return query.In(query.RelationshipProperty(common.Name.String()), relationshipNames)
			}).Fetch(func(cursor graph.Cursor[*graph.Relationship]) error {
				for relationship := range cursor.Chan() {
					if _, err := relationship.Properties.Get(common.Name.String()).String(); err != nil {
						return err
					}

					resultsFetched++
				}

				return cursor.Error()
			})
		}); err != nil {
			return err
		}

		return validateFetches(numExpectedFetches, resultsFetched)
	}
}

func FetchRelationshipsByProperty(propertyName string) func(testCase *TestCase) any {
	return func(testCase *TestCase) any {
		return func(tx graph.Transaction) error {
			var (
				numExpectedFetches = len(RelationshipIDs)
				resultsFetched     = 0
			)

			for iteration := 0; iteration < numExpectedFetches; iteration++ {
				relationshipName := "relationship " + strconv.Itoa(iteration)

				if err := testCase.Sample(func() error {
					return tx.Relationships().Filterf(func() graph.Criteria {
						return query.Equals(query.RelationshipProperty(propertyName), relationshipName)
					}).Fetch(func(cursor graph.Cursor[*graph.Relationship]) error {
						for relationship := range cursor.Chan() {
							if actualRelationshipName, err := relationship.Properties.Get(common.Name.String()).String(); err != nil {
								return err
							} else if relationshipName != actualRelationshipName {
								return fmt.Errorf("expected relationship name to be %s but got %s", relationshipName, actualRelationshipName)
							}

							resultsFetched++
						}

						return cursor.Error()
					})
				}); err != nil {
					return err
				}
			}

			return validateFetches(numExpectedFetches, resultsFetched)
		}
	}
}

func BatchDeleteEndNodesByID(testCase *TestCase) any {
	return func(batch graph.Batch) error {
		for _, nodeID := range EndNodeIDs {
			if err := testCase.Sample(func() error {
				return batch.DeleteNode(nodeID)
			}); err != nil {
				return err
			}
		}

		return nil
	}
}

func DeleteStartNodesByIDSlice(testCase *TestCase) any {
	return func(tx graph.Transaction) error {
		return testCase.Sample(func() error {
			return ops.DeleteNodes(tx, StartNodeIDs...)
		})
	}
}

func FetchRelationshipsByID(testCase *TestCase) any {
	return func(tx graph.Transaction) error {
		for _, relationshipID := range RelationshipIDs {
			if err := testCase.Sample(func() error {
				_, err := tx.Relationships().Filterf(func() graph.Criteria {
					return query.Equals(query.RelationshipID(), relationshipID)
				}).First()

				return err
			}); err != nil {
				return err
			}
		}

		return nil
	}
}

func NodeUpdateTests(testCase *TestCase) any {
	return func(tx graph.Transaction) error {
		for _, nodeID := range append(StartNodeIDs, EndNodeIDs...) {
			if err := testCase.Sample(func() error {
				if node, err := ops.FetchNode(tx, nodeID); err != nil {
					return err
				} else {
					node.Properties.Set(common.SystemTags.String(), "tag")
					node.Properties.Delete(common.Name.String())

					return tx.UpdateNode(node)
				}
			}); err != nil {
				return err
			}
		}

		return nil
	}
}

func BatchNodeAndRelationshipCreationTest(testCase *TestCase) any {
	return func(batch graph.Batch) error {
		for iteration := 0; iteration < SimpleRelationshipsToCreate; iteration++ {
			var (
				iterationStr              = strconv.Itoa(iteration)
				startNodePropertyValue    = "batch start node " + iterationStr
				endNodePropertyValue      = "batch end node " + iterationStr
				relationshipPropertyValue = "batch relationship " + iterationStr
			)

			var (
				startNode = graph.PrepareNode(graph.AsProperties(graph.PropertyMap{
					common.Name:     startNodePropertyValue,
					common.ObjectID: startNodePropertyValue,
				}), ad.Entity, ad.User)

				endNode = graph.PrepareNode(graph.AsProperties(graph.PropertyMap{
					common.Name:     endNodePropertyValue,
					common.ObjectID: endNodePropertyValue,
				}), ad.Entity, ad.Group)
			)

			if err := testCase.Sample(func() error {
				return batch.UpdateRelationshipBy(graph.RelationshipUpdate{
					Relationship: graph.PrepareRelationship(graph.AsProperties(graph.PropertyMap{
						common.Name:     relationshipPropertyValue,
						common.ObjectID: relationshipPropertyValue,
					}), ad.MemberOf),
					Start:                   startNode,
					StartIdentityKind:       ad.Entity,
					StartIdentityProperties: []string{common.ObjectID.String()},
					End:                     endNode,
					EndIdentityKind:         ad.Entity,
					EndIdentityProperties:   []string{common.ObjectID.String()},
				})
			}); err != nil {
				return err
			}
		}

		return nil
	}
}

func NodeAndRelationshipCreationTest(testCase *TestCase) any {
	return func(tx graph.Transaction) error {
		for iteration := 0; iteration < SimpleRelationshipsToCreate; iteration++ {
			var (
				iterationStr              = strconv.Itoa(iteration)
				startNodePropertyValue    = "start node " + iterationStr
				endNodePropertyValue      = "end node " + iterationStr
				relationshipPropertyValue = "relationship " + iterationStr
			)

			if err := testCase.Sample(func() error {
				if startNode, err := tx.CreateNode(graph.AsProperties(graph.PropertyMap{
					common.Name:     startNodePropertyValue,
					common.ObjectID: startNodePropertyValue,
				}), ad.Entity, ad.User); err != nil {
					return err
				} else if endNode, err := tx.CreateNode(graph.AsProperties(graph.PropertyMap{
					common.Name:     endNodePropertyValue,
					common.ObjectID: endNodePropertyValue,
				}), ad.Entity, ad.Group); err != nil {
					return err
				} else if relationship, err := tx.CreateRelationshipByIDs(startNode.ID, endNode.ID, ad.MemberOf, graph.AsProperties(graph.PropertyMap{
					common.Name:     relationshipPropertyValue,
					common.ObjectID: relationshipPropertyValue,
				})); err != nil {
					return err
				} else {
					StartNodeIDs[iteration] = startNode.ID
					EndNodeIDs[iteration] = endNode.ID
					RelationshipIDs[iteration] = relationship.ID
				}

				return nil
			}); err != nil {
				return err
			}
		}

		return nil
	}
}
