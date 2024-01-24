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
//
//go:generate go run go.uber.org/mock/mockgen -source=query.go -copyright_file=../../../../LICENSE.header -destination=./mocks/query.go -package=graph_mocks .

package graph

type ValueMapper interface {
	Next() (any, error)
	Map(target any) error
	MapOptions(target ...any) (any, error)
	Scan(targets ...any) error
}

type Scanner interface {
	Next() bool
	Values() (ValueMapper, error)
	Scan(targets ...any) error
}

type Result interface {
	Scanner

	Error() error
	Close()
}

type ErrorResult struct {
	err error
}

func (s ErrorResult) Next() bool {
	return false
}

func (s ErrorResult) Values() (ValueMapper, error) {
	return nil, s.err
}

func (s ErrorResult) Scan(targets ...any) error {
	return s.err
}

func (s ErrorResult) Error() error {
	return s.err
}

func (s ErrorResult) Close() {
}

func NewErrorResult(err error) Result {
	return ErrorResult{
		err: err,
	}
}

// Criteria is a top-level alias for communicating structured query filter criteria to a query generator.
type Criteria any

// CriteriaProvider is a function delegate that returns criteria.
type CriteriaProvider func() Criteria

// NodeQuery is an interface that covers all supported node query combinations. The contract supports a fluent
// interface to make query specifications more succinct.
type NodeQuery interface {
	// Filter applies the given criteria to this query.
	Filter(criteria Criteria) NodeQuery

	// Filterf applies the given criteria provider function to this query.
	Filterf(criteriaDelegate CriteriaProvider) NodeQuery

	// Query completes the query and hands the raw result to the given delegate for unmarshalling
	Query(delegate func(results Result) error, finalCriteria ...Criteria) error

	// Delete deletes any candidate nodes that match the query criteria
	Delete() error

	// Update updates all candidate nodes with the given properties
	Update(properties *Properties) error

	// OrderBy sets the OrderBy clause of the NodeQuery.
	OrderBy(criteria ...Criteria) NodeQuery

	// Offset sets an offset for the result set of the query. Using this function will enforce order on the result set.
	Offset(skip int) NodeQuery

	// Limit sets a maximum number of results to collect from the database.
	Limit(skip int) NodeQuery

	// Count completes the query and returns a tuple containing the count of results that were addressed by the
	// database and any error encountered during execution.
	Count() (int64, error)

	// First completes the query and returns the result and any error encountered during execution.
	First() (*Node, error)

	// Fetch completes the query and captures a cursor for iterating the result set. This cursor is passed to the given
	// delegate. Errors from the delegate are returned upwards as the error result of this call.
	Fetch(delegate func(cursor Cursor[*Node]) error) error

	// FetchIDs completes the query and captures a cursor for iterating the result set. This cursor is passed to the given
	// delegate. Errors from the delegate are returned upwards as the error result of this call.
	FetchIDs(delegate func(cursor Cursor[ID]) error) error

	// FetchKinds returns the ID and Kinds of matched nodes and omits property fetching
	FetchKinds(func(cursor Cursor[KindsResult]) error) error
}

// RelationshipQuery is an interface that covers all supported relationship query combinations. The contract supports a
// fluent interface to make query specifications more succinct.
type RelationshipQuery interface {
	// Filter applies the given criteria to this query.
	Filter(criteria Criteria) RelationshipQuery

	// Filterf applies the given criteria provider function to this query.
	Filterf(criteriaDelegate CriteriaProvider) RelationshipQuery

	// Update replaces the properties of all candidate relationships that matches the query criteria with the
	// given properties
	Update(properties *Properties) error

	// Delete deletes any candidate relationships that match the query criteria
	Delete() error

	// OrderBy sets the OrderBy clause of the RelationshipQuery.
	OrderBy(criteria ...Criteria) RelationshipQuery

	// Offset sets an offset for the result set of the query. Using this function will enforce order on the result set.
	Offset(skip int) RelationshipQuery

	// Limit sets a maximum number of results to collect from the database.
	Limit(skip int) RelationshipQuery

	// Count completes the query and returns a tuple containing the count of results that were addressed by the
	// database and any error encountered during execution.
	Count() (int64, error)

	// First completes the query and returns the result and any error encountered during execution.
	First() (*Relationship, error)

	// Query completes the query and hands the raw result to the given delegate for unmarshalling
	Query(delegate func(results Result) error, finalCriteria ...Criteria) error

	// Fetch completes the query and captures a cursor for iterating the result set. This cursor is passed to the given
	// delegate. Errors from the delegate are returned upwards as the error result of this call.
	Fetch(delegate func(cursor Cursor[*Relationship]) error) error

	// FetchDirection completes the query and captures a cursor for iterating through the relationship related nodes
	// for the given path direction
	FetchDirection(direction Direction, delegate func(cursor Cursor[DirectionalResult]) error) error

	// FetchIDs completes the query and captures a cursor for iterating the result set. This cursor is passed to the given
	// delegate. Errors from the delegate are returned upwards as the error result of this call.
	FetchIDs(delegate func(cursor Cursor[ID]) error) error

	//
	FetchTriples(delegate func(cursor Cursor[RelationshipTripleResult]) error) error

	//
	FetchAllShortestPaths(delegate func(cursor Cursor[Path]) error) error

	// FetchKinds returns the ID, Kind, Start ID and End ID of matched relationships and omits property fetching
	FetchKinds(delegate func(cursor Cursor[RelationshipKindsResult]) error) error
}
