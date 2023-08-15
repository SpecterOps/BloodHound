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

package neo4j_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/specterops/bloodhound/cypher/model"
	"github.com/specterops/bloodhound/dawgs/query"
	"github.com/specterops/bloodhound/dawgs/query/neo4j"

	"github.com/specterops/bloodhound/dawgs/graph"
	"github.com/stretchr/testify/require"
)

var (
	SystemTags = "system_tags"

	User         = graph.StringKind("User")
	Domain       = graph.StringKind("Domain")
	Computer     = graph.StringKind("Computer")
	Group        = graph.StringKind("Group")
	HasSession   = graph.StringKind("HasSession")
	GenericWrite = graph.StringKind("GenericWrite")
)

type QueryOutputAssertion struct {
	Query      string
	Parameters map[string]any
}

func expectAnalysisError(rawQuery *model.RegularQuery) func(t *testing.T) {
	return func(t *testing.T) {
		require.NotNil(t, neo4j.NewQueryBuilder(rawQuery).Prepare())
	}
}

func assertQueryShortestPathResult(rawQuery *model.RegularQuery, expectedOutput string, expectedParameters ...map[string]any) func(t *testing.T) {
	return func(t *testing.T) {
		builder := neo4j.NewQueryBuilder(rawQuery)

		// Validate that building the query didn't throw an error
		require.Nil(t, builder.PrepareAllShortestPaths())

		if len(expectedParameters) == 1 {
			require.Equal(t, expectedParameters[0], builder.Parameters)
		}

		output, err := builder.Render()

		require.Nil(t, err)
		require.Equal(t, expectedOutput, output)
	}
}

func assertQueryResult(rawQuery *model.RegularQuery, expectedOutput string, expectedParameters ...map[string]any) func(t *testing.T) {
	return func(t *testing.T) {
		var (
			builder    = neo4j.NewQueryBuilder(rawQuery)
			prepareErr = builder.Prepare()
		)

		// Validate that building the query didn't throw an error
		if prepareErr != nil {
			require.Nilf(t, prepareErr, prepareErr.Error())
		}

		if len(expectedParameters) == 1 {
			require.Equal(t, expectedParameters[0], builder.Parameters)
		}

		output, err := builder.Render()

		require.Nil(t, err)
		require.Equal(t, expectedOutput, output)
	}
}

func assertOneOfQueryResult(rawQuery *model.RegularQuery, expectations []QueryOutputAssertion) func(t *testing.T) {
	return func(t *testing.T) {
		builder := neo4j.NewQueryBuilder(rawQuery)

		// Validate that building the query didn't throw an error
		require.Nil(t, builder.Prepare())

		output, err := builder.Render()
		require.Nil(t, err)

		var matchingExpectation *QueryOutputAssertion

		for _, expectation := range expectations {
			if expectation.Query == output {
				matchingExpectation = &expectation
				break
			}
		}

		if matchingExpectation == nil {
			msg := fmt.Sprintf("Rendered query did not match any given options.\nActual:\n\t%s\nExpected one of: ", output)

			for _, expectation := range expectations {
				msg += "\n\t" + expectation.Query
			}

			t.Fatalf(msg)
		} else if matchingExpectation.Parameters != nil {
			require.Equal(t, matchingExpectation.Parameters, builder.Parameters)
		}
	}
}

func TestQueryBuilder_RenderShortestPaths(t *testing.T) {
	t.Run("Shortest Paths with Unbound Relationship", assertQueryShortestPathResult(query.SinglePartQuery(
		query.Where(
			query.And(
				query.Equals(query.StartProperty("objectid"), "12345"),
				query.KindIn(query.Start(), graph.StringKind("A"), graph.StringKind("B")),

				query.Equals(query.EndProperty("objectid"), "56789"),
				query.KindIn(query.End(), graph.StringKind("B")),
			),
		),

		query.Returning(
			query.Path(),
		),
	), "match p = allShortestPaths((s)-[*]->(e)) where s.objectid = $p0 and (s:A or s:B) and e.objectid = $p1 and e:B return p", map[string]any{
		"p0": "12345",
		"p1": "56789",
	}))

	t.Run("Shortest Paths with Bound Relationship", assertQueryShortestPathResult(query.SinglePartQuery(
		query.Where(
			query.And(
				query.Equals(query.StartProperty("objectid"), "12345"),
				query.KindIn(query.Start(), graph.StringKind("A"), graph.StringKind("B")),
				query.KindIn(query.Relationship(), graph.StringKind("R1"), graph.StringKind("R2")),
				query.Equals(query.EndProperty("objectid"), "56789"),
				query.KindIn(query.End(), graph.StringKind("B")),
			),
		),

		query.Returning(
			query.Path(),
		),
	), "match p = allShortestPaths((s)-[r:R1|R2*]->(e)) where s.objectid = $p0 and (s:A or s:B) and e.objectid = $p1 and e:B return p", map[string]any{
		"p0": "12345",
		"p1": "56789",
	}))
}

func TestQueryBuilder_Render(t *testing.T) {
	// Node Queries
	t.Run("Node Count", assertQueryResult(query.SinglePartQuery(
		query.Where(
			query.In(query.NodeID(), []graph.ID{1, 2, 3, 4}),
		),

		query.Returning(
			query.Count(query.Node()),
		),

		query.Limit(10),
		query.Offset(20),
	), "match (n) where id(n) in $p0 return count(n) skip 20 limit 10", map[string]any{
		"p0": []graph.ID{1, 2, 3, 4},
	}))

	t.Run("Node Item", assertQueryResult(query.SinglePartQuery(
		query.Where(
			query.In(query.NodeProperty("prop"), []int{1, 2, 3, 4}),
		),

		query.Returning(
			query.Count(query.Node()),
		),
	), "match (n) where n.prop in $p0 return count(n)"))

	// TODO: Revisit parameter reuse
	//
	//reusedLiteral := query3.Literal([]int{1, 2, 3, 4})
	//
	//t.Run("Node Item with Reused Literal", assertQueryResult(query3.Query(
	//	query3.Where(
	//		query3.And(
	//			query3.In(query3.NodeProperty("prop"), reusedLiteral),
	//			query3.In(query3.NodeProperty("other_prop"), reusedLiteral),
	//		),
	//	),
	//
	//	query3.Returning(
	//		query3.Count(query3.Node()),
	//	),
	//), "match (n) where n.prop in $p0 and n.other_prop in $p0 return count(n)"))

	t.Run("Distinct Item", assertQueryResult(query.SinglePartQuery(
		query.Where(
			query.In(query.NodeProperty("prop"), []int{1, 2, 3, 4}),
		),

		query.ReturningDistinct(
			query.NodeProperty("prop"),
		),
	), "match (n) where n.prop in $p0 return distinct n.prop"))

	t.Run("Count Distinct Item", assertQueryResult(query.SinglePartQuery(
		query.Where(
			query.In(query.NodeProperty("prop"), []int{1, 2, 3, 4}),
		),

		query.Returning(
			query.CountDistinct(query.NodeProperty("prop")),
		),
	), "match (n) where n.prop in $p0 return count(distinct n.prop)"))

	t.Run("Set Node Labels", assertQueryResult(query.SinglePartQuery(
		query.Where(
			query.In(query.NodeProperty("prop"), []int{1, 2, 3, 4}),
		),

		query.Update(
			query.AddKind(query.Node(), Domain),
			query.AddKind(query.Node(), User),
		),

		query.Returning(
			query.Count(query.Node()),
		),
	), "match (n) where n.prop in $p0 set n:Domain set n:User return count(n)"))

	t.Run("Remove Node Labels", assertQueryResult(query.SinglePartQuery(
		query.Where(
			query.In(query.NodeProperty("prop"), []int{1, 2, 3, 4}),
		),

		query.Update(
			query.DeleteKind(query.Node(), Domain),
			query.DeleteKind(query.Node(), User),
		),

		query.Returning(
			query.Count(query.Node()),
		),
	), "match (n) where n.prop in $p0 remove n:Domain remove n:User return count(n)"))

	t.Run("Multiple Node ID References", assertQueryResult(query.SinglePartQuery(
		query.Where(
			query.And(
				query.Equals(query.NodeProperty("name"), "name"),
				query.In(query.NodeID(), []int{1, 2, 3, 4}),
			),
		),

		query.Returning(
			query.Identity(query.Node()),
			query.Property(query.Node(), "value"),
		),

		query.Limit(10),
		query.Offset(20),
	), "match (n) where n.name = $p0 and id(n) in $p1 return id(n), n.value skip 20 limit 10"))

	// Create node
	t.Run("Create Node", assertQueryResult(query.SinglePartQuery(
		query.Create(
			query.NodePattern(
				graph.Kinds{Domain, Computer},
				query.Parameter(map[string]any{
					"prop1": 1234,
				}),
			),
		),

		query.Returning(
			query.Identity(query.Node()),
		),
	),
		"create (n:Domain:Computer $p0) return id(n)",
		map[string]any{
			"p0": map[string]any{
				"prop1": 1234,
			},
		},
	))

	// Set with node

	t.Run("DeleteProperty with Multiple Node ID References", assertQueryResult(query.SinglePartQuery(
		query.Where(
			query.And(
				query.Equals(query.NodeProperty("name"), "name"),
				query.In(query.NodeID(), []int{1, 2, 3, 4}),
			),
		),

		query.Update(
			query.DeleteProperty(query.NodeProperty("other")),
			query.DeleteProperty(query.NodeProperty("other2")),
		),

		query.Returning(
			query.Identity(query.Node()),
			query.Property(query.Node(), "value"),
		),

		query.Limit(10),
		query.Offset(20),
	), "match (n) where n.name = $p0 and id(n) in $p1 remove n.other remove n.other2 return id(n), n.value skip 20 limit 10"))

	properties := graph.NewProperties()
	properties.Set("test_1", "value_1")
	properties.Set("test_2", "value_2")

	t.Run("Set from Map", assertOneOfQueryResult(query.SinglePartQuery(
		query.Where(
			query.Equals(query.NodeProperty("objectid"), "12345"),
		),

		query.Update(
			query.SetProperties(query.Node(), properties.ModifiedProperties()),
		),
	), []QueryOutputAssertion{
		{
			Query: "match (n) where n.objectid = $p0 set n.test_1 = $p1, n.test_2 = $p2",
			Parameters: map[string]any{
				"p0": "12345",
				"p1": "value_1",
				"p2": "value_2",
			},
		},
		{
			Query: "match (n) where n.objectid = $p0 set n.test_2 = $p1, n.test_1 = $p2",
			Parameters: map[string]any{
				"p0": "12345",
				"p1": "value_2",
				"p2": "value_1",
			},
		},
	}))

	properties.Delete("test_1")
	properties.Delete("test_2")

	t.Run("DeleteProperty from Map", assertOneOfQueryResult(query.SinglePartQuery(
		query.Where(
			query.Equals(query.NodeProperty("objectid"), "12345"),
		),

		query.Update(
			query.DeleteProperties(query.Node(), properties.DeletedProperties()...),
		),
	), []QueryOutputAssertion{
		{
			Query: "match (n) where n.objectid = $p0 remove n.test_2, n.test_1",
			Parameters: map[string]any{
				"p0": "12345",
			},
		},
		{
			Query: "match (n) where n.objectid = $p0 remove n.test_1, n.test_2",
			Parameters: map[string]any{
				"p0": "12345",
			},
		},
	}))

	t.Run("Set with Multiple Node ID References", assertQueryResult(query.SinglePartQuery(
		query.Where(
			query.And(
				query.Equals(query.NodeProperty("name"), "name"),
				query.In(query.NodeID(), []int{1, 2, 3, 4}),
			),
		),

		query.Update(
			query.SetProperty(query.NodeProperty("other"), "value"),
		),

		query.Returning(
			query.Identity(query.Node()),
			query.Property(query.Node(), "value"),
		),

		query.Limit(10),
		query.Offset(20),
	), "match (n) where n.name = $p0 and id(n) in $p1 set n.other = $p2 return id(n), n.value skip 20 limit 10"))

	updatedNode := graph.NewNode(graph.ID(1), graph.NewProperties(), User, Domain, Computer)
	updatedNode.Properties.Set("test_1", "value_1")
	updatedNode.Properties.Delete("test_2")

	t.Run("Node Set and Remove Multiple Kinds and Properties", assertQueryResult(query.SinglePartQuery(
		query.Where(
			query.Equals(query.NodeID(), updatedNode.ID),
		),

		query.Updatef(func() graph.Criteria {
			var (
				properties       = updatedNode.Properties
				updateStatements = []graph.Criteria{
					query.AddKinds(query.Node(), updatedNode.Kinds),
				}
			)

			if modifiedProperties := properties.ModifiedProperties(); len(modifiedProperties) > 0 {
				updateStatements = append(updateStatements, query.SetProperties(query.Node(), modifiedProperties))
			}

			if deletedProperties := properties.DeletedProperties(); len(deletedProperties) > 0 {
				updateStatements = append(updateStatements, query.DeleteProperties(query.Node(), deletedProperties...))
			}

			return updateStatements
		}),
	), "match (n) where id(n) = $p0 set n:User:Domain:Computer set n.test_1 = $p1 remove n.test_2"))

	t.Run("Node has Relationships", assertQueryResult(query.SinglePartQuery(
		query.Where(
			query.HasRelationships(query.Node()),
		),

		query.Returning(
			query.Node(),
		),
	), "match (n) where (n)<-[]->() return n"))

	t.Run("Node has Relationships Order by Node Item", assertQueryResult(query.SinglePartQuery(
		query.Where(
			query.HasRelationships(query.Node()),
		),

		query.Returning(
			query.Node(),
		),

		query.OrderBy(
			query.Order(query.NodeProperty("value"), query.Ascending()),
		),
	), "match (n) where (n)<-[]->() return n order by n.value asc"))

	t.Run("Node has Relationships Order by Node Item", assertQueryResult(query.SinglePartQuery(
		query.Where(
			query.HasRelationships(query.Node()),
		),

		query.Returning(
			query.Node(),
		),

		query.OrderBy(
			query.Order(query.NodeProperty("value_1"), query.Ascending()),
			query.Order(query.NodeProperty("value_2"), query.Descending()),
		),
	), "match (n) where (n)<-[]->() return n order by n.value_1 asc, n.value_2 desc"))

	t.Run("Node has Relationships Order by Node Item with Limit and Offset", assertQueryResult(query.SinglePartQuery(
		query.Where(
			query.HasRelationships(query.Node()),
		),

		query.Returning(
			query.Node(),
		),

		query.OrderBy(
			query.Order(query.NodeProperty("value_1"), query.Ascending()),
			query.Order(query.NodeProperty("value_2"), query.Descending()),
		),

		query.Limit(10),
		query.Offset(20),
	), "match (n) where (n)<-[]->() return n order by n.value_1 asc, n.value_2 desc skip 20 limit 10"))

	t.Run("Node has no Relationships", assertQueryResult(query.SinglePartQuery(
		query.Where(
			query.Not(query.HasRelationships(query.Node())),
		),

		query.Returning(
			query.Node(),
		),
	), "match (n) where not ((n)<-[]->()) return n"))

	t.Run("Node Datetime Before", assertQueryResult(query.SinglePartQuery(
		query.Where(
			query.And(
				query.Before(query.NodeProperty("lastseen"), time.Now().UTC()),
				query.In(query.NodeID(), []int{1, 2, 3, 4}),
			),
		),

		query.Returning(
			query.Node(),
		),
	), "match (n) where n.lastseen < $p0 and id(n) in $p1 return n"))

	t.Run("Node Datetime Before or Equal to", assertQueryResult(query.SinglePartQuery(
		query.Where(
			query.And(
				query.LessThanOrEquals(query.NodeProperty("lastseen"), time.Now().UTC()),
				query.In(query.NodeID(), []int{1, 2, 3, 4}),
			),
		),

		query.Returning(
			query.Node(),
		),
	), "match (n) where n.lastseen <= $p0 and id(n) in $p1 return n"))

	t.Run("Node Datetime After", assertQueryResult(query.SinglePartQuery(
		query.Where(
			query.And(
				query.After(query.NodeProperty("lastseen"), time.Now().UTC()),
				query.In(query.NodeID(), []int{1, 2, 3, 4}),
			),
		),

		query.Returning(
			query.Node(),
		),
	), "match (n) where n.lastseen > $p0 and id(n) in $p1 return n"))

	t.Run("Node Datetime After or Equal to", assertQueryResult(query.SinglePartQuery(
		query.Where(
			query.And(
				query.GreaterThanOrEquals(query.NodeProperty("lastseen"), time.Now().UTC()),
				query.In(query.NodeID(), []int{1, 2, 3, 4}),
			),
		),

		query.Returning(
			query.Node(),
		),
	), "match (n) where n.lastseen >= $p0 and id(n) in $p1 return n"))

	t.Run("Node PropertyExists", assertQueryResult(query.SinglePartQuery(
		query.Where(
			query.And(
				query.Exists(query.NodeProperty("lastseen")),
				query.In(query.NodeID(), []int{1, 2, 3, 4}),
			),
		),

		query.Returning(
			query.Node(),
		),
	), "match (n) where n.lastseen is not null and id(n) in $p0 return n"))

	t.Run("Select Node Kinds", assertQueryResult(query.SinglePartQuery(
		query.Where(
			query.And(
				query.Kind(query.Node(), Domain),
			),
		),

		query.Returning(
			query.KindsOf(query.Node()),
		),
	), "match (n) where n:Domain return labels(n)"))

	t.Run("Select Node ID and Kinds", assertQueryResult(query.SinglePartQuery(
		query.Where(
			query.And(
				query.Kind(query.Node(), Domain),
			),
		),

		query.Returning(
			query.NodeID(),
			query.KindsOf(query.Node()),
		),
	), "match (n) where n:Domain return id(n), labels(n)"))

	t.Run("Node Kind Match", assertQueryResult(query.SinglePartQuery(
		query.Where(
			query.And(
				query.Kind(query.Node(), Domain),
			),
		),

		query.Returning(
			query.Node(),
		),
	), "match (n) where n:Domain return n"))

	t.Run("Node Kind In", assertQueryResult(query.SinglePartQuery(
		query.Where(
			query.And(
				query.KindIn(query.Node(), Domain, User, Group),
			),
		),

		query.Returning(
			query.Node(),
		),
	), "match (n) where (n:Domain or n:User or n:Group) return n"))

	t.Run("Node String Item Contains", assertQueryResult(query.SinglePartQuery(
		query.Where(
			query.StringContains(query.NodeProperty("tags"), "tag_1"),
		),

		query.Returning(
			query.Node(),
		),
	), "match (n) where n.tags contains $p0 return n"))

	t.Run("Node String Item Starts With", assertQueryResult(query.SinglePartQuery(
		query.Where(
			query.StringStartsWith(query.NodeProperty("tags"), "tag_1"),
		),

		query.Returning(
			query.Node(),
		),
	), "match (n) where n.tags starts with $p0 return n"))

	t.Run("Node String Item Ends With", assertQueryResult(query.SinglePartQuery(
		query.Where(
			query.StringEndsWith(query.NodeProperty("tags"), "tag_1"),
		),

		query.Returning(
			query.Node(),
		),
	), "match (n) where n.tags ends with $p0 return n"))

	t.Run("Node String Item Case Insensitive Contains", assertQueryResult(query.SinglePartQuery(
		query.Where(
			query.CaseInsensitiveStringContains(query.NodeProperty("tags"), "tag_1"),
		),

		query.Returning(
			query.Node(),
		),
	), "match (n) where toLower(n.tags) contains $p0 return n"))

	t.Run("Node String Item Case Insensitive Starts With", assertQueryResult(query.SinglePartQuery(
		query.Where(
			query.CaseInsensitiveStringStartsWith(query.NodeProperty("tags"), "tag_1"),
		),

		query.Returning(
			query.Node(),
		),
	), "match (n) where toLower(n.tags) starts with $p0 return n"))

	t.Run("Node String Item Case Insensitive Ends With", assertQueryResult(query.SinglePartQuery(
		query.Where(
			query.CaseInsensitiveStringEndsWith(query.NodeProperty("tags"), "tag_1"),
		),

		query.Returning(
			query.Node(),
		),
	), "match (n) where toLower(n.tags) ends with $p0 return n"))

	t.Run("Node Delete", assertQueryResult(query.SinglePartQuery(
		query.Where(
			query.In(query.Node(), []graph.ID{1, 2, 3}),
		),

		query.Delete(
			query.Node(),
		),
	), "match (n) where n in $p0 detach delete n"))

	// Relationship Queries
	t.Run("Empty Relationship Query", assertQueryResult(query.SinglePartQuery(
		query.Returning(
			query.RelationshipID(),
		),
	), "match ()-[r]->() return id(r)"))

	t.Run("Empty Start Node Query", assertQueryResult(query.SinglePartQuery(
		query.Returning(
			query.StartID(),
		),
	), "match (s)-[]->() return id(s)"))

	t.Run("Empty End Node Query", assertQueryResult(query.SinglePartQuery(
		query.Returning(
			query.EndID(),
		),
	), "match ()-[]->(e) return id(e)"))

	t.Run("Returning Relationship Kind Query", assertQueryResult(query.SinglePartQuery(
		query.Returning(
			query.RelationshipID(),
			query.KindsOf(query.Relationship()),
		),
	), "match ()-[r]->() return id(r), type(r)"))

	t.Run("Returning Start and Relationship Kind Query", assertQueryResult(query.SinglePartQuery(
		query.Returning(
			query.RelationshipID(),
			query.KindsOf(query.Relationship()),
			query.KindsOf(query.Start()),
		),
	), "match (s)-[r]->() return id(r), type(r), labels(s)"))

	t.Run("Returning End and Relationship Kind Query", assertQueryResult(query.SinglePartQuery(
		query.Returning(
			query.RelationshipID(),
			query.KindsOf(query.Relationship()),
			query.KindsOf(query.End()),
		),
	), "match ()-[r]->(e) return id(r), type(r), labels(e)"))

	t.Run("Returning Start, End and Relationship Kind Query", assertQueryResult(query.SinglePartQuery(
		query.Returning(
			query.RelationshipID(),
			query.KindsOf(query.Relationship()),
			query.KindsOf(query.Start()),
			query.KindsOf(query.End()),
		),
	), "match (s)-[r]->(e) return id(r), type(r), labels(s), labels(e)"))

	t.Run("Relationship Item and ID References", assertQueryResult(query.SinglePartQuery(
		query.Where(
			query.And(
				query.Equals(query.RelationshipProperty("name"), "name"),
				query.In(query.RelationshipID(), []int{1, 2, 3, 4}),
			),
		),

		query.Returning(
			query.RelationshipID(),
			query.Property(query.Relationship(), "value"),
		),

		query.Offset(20),
	), "match ()-[r]->() where r.name = $p0 and id(r) in $p1 return id(r), r.value skip 20"))

	t.Run("Relationship Select Start References", assertQueryResult(query.SinglePartQuery(
		query.Where(
			query.And(
				query.Equals(query.RelationshipProperty("name"), "name"),
				query.In(query.RelationshipID(), []int{1, 2, 3, 4}),
			),
		),

		query.Returning(
			query.StartID(),
			query.Property(query.Relationship(), "value"),
		),

		query.Offset(20),
	), "match (s)-[r]->() where r.name = $p0 and id(r) in $p1 return id(s), r.value skip 20"))

	t.Run("Relationship Start Node ID Reference", assertQueryResult(query.SinglePartQuery(
		query.Where(
			query.And(
				query.Equals(query.StartID(), 1),
				query.Equals(query.RelationshipProperty("name"), "name"),
				query.In(query.RelationshipID(), []int{1, 2, 3, 4}),
			),
		),

		query.Returning(
			query.RelationshipID(),
			query.Property(query.Relationship(), "value"),
		),

		query.Offset(20),
	), "match (s)-[r]->() where id(s) = $p0 and r.name = $p1 and id(r) in $p2 return id(r), r.value skip 20"))

	t.Run("Relationship End Node ID Reference", assertQueryResult(query.SinglePartQuery(
		query.Where(
			query.And(
				query.Equals(query.EndID(), 1),
				query.Equals(query.RelationshipProperty("name"), "name"),
				query.In(query.RelationshipID(), []int{1, 2, 3, 4}),
			),
		),

		query.Returning(
			query.RelationshipID(),
			query.Property(query.Relationship(), "value"),
		),

		query.Offset(20),
	), "match ()-[r]->(e) where id(e) = $p0 and r.name = $p1 and id(r) in $p2 return id(r), r.value skip 20"))

	t.Run("Relationship Start and End Node ID References", assertQueryResult(query.SinglePartQuery(
		query.Where(
			query.And(
				query.Equals(query.StartID(), 1),
				query.Equals(query.EndID(), 1),
				query.Equals(query.RelationshipProperty("name"), "name"),
				query.In(query.RelationshipID(), []int{1, 2, 3, 4}),
			),
		),

		query.Returning(
			query.RelationshipID(),
			query.Property(query.Relationship(), "value"),
		),
	), "match (s)-[r]->(e) where id(s) = $p0 and id(e) = $p1 and r.name = $p2 and id(r) in $p3 return id(r), r.value"))

	t.Run("Relationship Kind Match without Joining Expression", assertQueryResult(query.SinglePartQuery(
		query.Where(
			query.KindIn(query.Relationship(), Domain, User, GenericWrite, HasSession),
		),

		query.Returning(
			query.RelationshipID(),
			query.Property(query.Relationship(), "value"),
		),
	), "match ()-[r:Domain|User|GenericWrite|HasSession]->() return id(r), r.value"))

	t.Run("Relationship Kind Match", assertQueryResult(query.SinglePartQuery(
		query.Where(
			query.And(
				query.Equals(query.StartID(), 1),
				query.KindIn(query.Relationship(), HasSession),
				query.Equals(query.RelationshipProperty("name"), "name"),
				query.In(query.RelationshipID(), []int{1, 2, 3, 4}),
			),
		),

		query.Returning(
			query.RelationshipID(),
			query.Property(query.Relationship(), "value"),
		),
	), "match (s)-[r:HasSession]->() where id(s) = $p0 and r.name = $p1 and id(r) in $p2 return id(r), r.value"))

	updatedRelationship := graph.NewRelationship(graph.ID(1), graph.ID(2), graph.ID(3), graph.NewProperties(), HasSession)
	updatedRelationship.Properties.Set("test_1", "value_1")
	updatedRelationship.Properties.Delete("test_2")

	t.Run("Relationship Set and Remove Multiple Kinds and Properties", assertQueryResult(query.SinglePartQuery(
		query.Where(
			query.Equals(query.RelationshipID(), updatedRelationship.ID),
		),

		query.Updatef(func() graph.Criteria {
			var (
				properties       = updatedRelationship.Properties
				updateStatements []graph.Criteria
			)

			if modifiedProperties := properties.ModifiedProperties(); len(modifiedProperties) > 0 {
				updateStatements = append(updateStatements, query.SetProperties(query.Relationship(), modifiedProperties))
			}

			if deletedProperties := properties.DeletedProperties(); len(deletedProperties) > 0 {
				updateStatements = append(updateStatements, query.DeleteProperties(query.Relationship(), deletedProperties...))
			}

			return updateStatements
		}),
	), "match ()-[r]->() where id(r) = $p0 set r.test_1 = $p1 remove r.test_2"))

	t.Run("Relationship Kind Match in", assertQueryResult(query.SinglePartQuery(
		query.Where(
			query.And(
				query.Equals(query.StartID(), 1),
				query.KindIn(query.Relationship(), HasSession, GenericWrite),
				query.Equals(query.RelationshipProperty("name"), "name"),
				query.In(query.RelationshipID(), []int{1, 2, 3, 4}),
			),
		),

		query.Returning(
			query.RelationshipID(),
			query.Property(query.Relationship(), "value"),
		),
	), "match (s)-[r:HasSession|GenericWrite]->() where id(s) = $p0 and r.name = $p1 and id(r) in $p2 return id(r), r.value"))

	t.Run("Relationship Kind Match in and Start Node Kind Match in", assertQueryResult(query.SinglePartQuery(
		query.Where(
			query.And(
				query.KindIn(query.Start(), User, Computer),
				query.KindIn(query.Relationship(), HasSession, GenericWrite),
				query.Equals(query.RelationshipProperty("name"), "name"),
				query.In(query.RelationshipID(), []int{1, 2, 3, 4}),
			),
		),

		query.Returning(
			query.RelationshipID(),
			query.Property(query.Relationship(), "value"),
		),
	), "match (s)-[r:HasSession|GenericWrite]->() where (s:User or s:Computer) and r.name = $p0 and id(r) in $p1 return id(r), r.value"))

	t.Run("Relationship Kind Match in and Delete Start Node and Relationship", assertQueryResult(query.SinglePartQuery(
		query.Where(
			query.And(
				query.KindIn(query.Relationship(), HasSession, GenericWrite),
			),
		),

		query.Delete(
			query.Start(),
			query.Relationship(),
		),
	), "match (s)-[r:HasSession|GenericWrite]->() delete s, r"))

	t.Run("Relationship Kind Match in and Delete Start Node and Relationship Returning Count Relationships Deleted", assertQueryResult(query.SinglePartQuery(
		query.Where(
			query.And(
				query.KindIn(query.Relationship(), HasSession, GenericWrite),
			),
		),

		query.Delete(
			query.Start(),
			query.Relationship(),
		),

		query.Returning(
			query.Count(query.Relationship()),
		),
	), "match (s)-[r:HasSession|GenericWrite]->() delete s, r return count(r)"))

	t.Run("Create Relationship", assertQueryResult(query.SinglePartQuery(
		query.Create(
			query.StartNodePattern(
				graph.Kinds{Computer},
				query.Parameter(map[string]any{
					"prop1": 1234,
				}),
			),
			query.RelationshipPattern(
				HasSession,
				query.Parameter(map[string]any{
					"prop1": 1234,
				}),
				graph.DirectionOutbound,
			),
			query.EndNodePattern(
				graph.Kinds{User},
				query.Parameter(map[string]any{
					"prop1": 1234,
				}),
			),
		),

		query.Returning(
			query.Identity(query.Relationship()),
		),
	),
		"create (s:Computer $p0)-[r:HasSession $p1]->(e:User $p2) return id(r)",
		map[string]any{
			"p0": map[string]any{
				"prop1": 1234,
			},
			"p1": map[string]any{
				"prop1": 1234,
			},
			"p2": map[string]any{
				"prop1": 1234,
			},
		},
	))

	t.Run("Create Relationship with Match", assertQueryResult(query.SinglePartQuery(
		query.Where(
			query.And(
				query.Equals(query.StartID(), 1),
				query.Equals(query.EndID(), 2),
			),
		),

		query.Create(
			query.Start(),
			query.RelationshipPattern(
				HasSession,
				query.Parameter(map[string]any{
					"prop1": 1234,
				}),
				graph.DirectionOutbound,
			),
			query.End(),
		),

		query.Returning(
			query.Identity(query.Relationship()),
		),
	),
		"match (s), (e) where id(s) = $p0 and id(e) = $p1 create (s)-[r:HasSession $p2]->(e) return id(r)",
		map[string]any{
			"p0": 1,
			"p1": 2,
			"p2": map[string]any{
				"prop1": 1234,
			},
		},
	))

	t.Run("Not String Contains Operator Rewrite", assertQueryResult(query.SinglePartQuery(
		query.Where(
			query.Not(
				query.StringContains(query.Property(query.Node(), SystemTags), "admin_tier_0"),
			),
		),

		query.Returning(
			query.Count(query.Node()),
		),
	), "match (n) where (not (n.system_tags contains $p0) or n.system_tags is null) return count(n)"))

	t.Run("Is Not Null", assertQueryResult(query.SinglePartQuery(
		query.Where(
			query.IsNotNull(
				query.Property(query.Node(), SystemTags),
			),
		),
		query.Returning(
			query.Count(query.Node()),
		)),
		"match (n) where n.system_tags is not null return count(n)"))

	t.Run("Is Null", assertQueryResult(query.SinglePartQuery(
		query.Where(
			query.IsNull(
				query.Property(query.Node(), SystemTags),
			),
		),
		query.Returning(
			query.Count(query.Node()),
		)),
		"match (n) where n.system_tags is null return count(n)"))
}

func TestQueryBuilder_Analyze(t *testing.T) {
	// Don't allow node query references to intermingle with relationship query references
	t.Run("Should Reject Mixing Query Type References", expectAnalysisError(query.SinglePartQuery(
		query.Where(
			query.And(
				query.Equals(query.NodeID(), 1),
				query.Equals(query.Property(query.Relationship(), "name"), "name"),
				query.In(query.RelationshipID(), []int{1, 2, 3, 4}),
			),
		),

		query.Returning(
			query.RelationshipID(),
			query.Property(query.Relationship(), "value"),
		),

		query.Offset(20),
	)))

	t.Run("Should Reject Mixing Query Type References", expectAnalysisError(query.SinglePartQuery(
		query.Where(
			query.And(
				query.Equals(query.NodeID(), 1),
				query.Equals(query.Property(query.Relationship(), "name"), "name"),
				query.In(query.RelationshipID(), []int{1, 2, 3, 4}),
			),
		),

		query.Returning(
			query.RelationshipID(),
			query.Property(query.Relationship(), "value"),
		),

		query.Offset(20),
	)))

	t.Run("Should fail on bad query criteria", expectAnalysisError(query.SinglePartQuery(
		query.Node(),
	)))

	t.Run("Should fail on bad create criteria", expectAnalysisError(query.SinglePartQuery(
		query.Create(
			query.Where(
				query.And(),
			),
		),
	)))

	t.Run("Should fail on bad variable reference types for KindOf", expectAnalysisError(query.SinglePartQuery(
		query.Where(
			query.KindsOf(
				query.Create(),
			),
		),
	)))
}
