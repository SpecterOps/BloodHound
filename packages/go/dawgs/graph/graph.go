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

package graph

//go:generate go run go.uber.org/mock/mockgen -source=graph.go -copyright_file=../../../../LICENSE.header -destination=./mocks/graph.go -package=graph_mocks .

import (
	"context"
	"errors"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/specterops/bloodhound/dawgs/util/size"
)

const (
	DirectionInbound  Direction = 0
	DirectionOutbound Direction = 1
	DirectionBoth     Direction = 2
	End                         = DirectionInbound
	Start                       = DirectionOutbound
)

var ErrInvalidDirection = errors.New("must be called with either an inbound or outbound direction")

// Direction describes the direction of a graph traversal. A Direction may be either Inbound or DirectionOutbound.
type Direction int

// Reverse returns the reverse of the current direction.
func (s Direction) Reverse() (Direction, error) {
	switch s {
	case DirectionInbound:
		return DirectionOutbound, nil

	case DirectionOutbound:
		return DirectionInbound, nil

	default:
		return 0, ErrInvalidDirection
	}
}

// PickID picks either the start or end Node ID from a Relationship depending on the direction of the receiver.
func (s Direction) PickID(start, end ID) (ID, error) {
	switch s {
	case DirectionInbound:
		return end, nil

	case DirectionOutbound:
		return start, nil

	default:
		return 0, ErrInvalidDirection
	}
}

// PickReverseID picks either the start or end Node ID from a Relationship depending on the direction of the receiver.
func (s Direction) PickReverseID(start, end ID) (ID, error) {
	switch s {
	case DirectionInbound:
		return start, nil

	case DirectionOutbound:
		return end, nil

	default:
		return 0, ErrInvalidDirection
	}
}

// Pick picks either the start or end Node ID from a Relationship depending on the direction of the receiver.
func (s Direction) Pick(relationship *Relationship) (ID, error) {
	return s.PickID(relationship.StartID, relationship.EndID)
}

// PickReverse picks either the start or end Node ID from a Relationship depending on the direction of the receiver.
func (s Direction) PickReverse(relationship *Relationship) (ID, error) {
	return s.PickReverseID(relationship.StartID, relationship.EndID)
}

func (s Direction) String() string {
	switch s {
	case DirectionInbound:
		return "inbound"
	case DirectionOutbound:
		return "outbound"
	case DirectionBoth:
		return "both"
	default:
		return "invalid"
	}
}

// ID is a 32-bit database Entity identifier type. Negative ID value associations in DAWGS drivers are not recommended
// and should not be considered during driver implementation.
type ID uint32

// Uint64 returns the ID typed as an uint64 and is shorthand for uint64(id).
func (s ID) Uint64() uint64 {
	return uint64(s)
}

// Uint32 returns the ID typed as an uint32 and is shorthand for uint32(id).
func (s ID) Uint32() uint32 {
	return uint32(s)
}

// Int64 returns the ID typed as an int64 and is shorthand for int64(id).
func (s ID) Int64() int64 {
	return int64(s)
}

// String formats the int64 value of the ID as a string.
func (s ID) String() string {
	return strconv.FormatInt(s.Int64(), 10)
}

// PropertyValue is an interface that offers type negotiation for property values to reduce the boilerplate required
// handle property values.
type PropertyValue interface {
	// IsNil returns true if the property value is nil.
	IsNil() bool

	// Bool returns the property value as a bool along with any type negotiation error information.
	Bool() (bool, error)

	// Int returns the property value as an int along with any type negotiation error information.
	Int() (int, error)

	// Int64 returns the property value as an int64 along with any type negotiation error information.
	Int64() (int64, error)

	// Int64Slice returns the property value as an int64 slice along with any type negotiation error information.
	Int64Slice() ([]int64, error)

	// IDSlice returns the property value as a ID slice along with any type negotiation error information.
	IDSlice() ([]ID, error)

	//StringSlice returns the property value as a string slice along with any type negotiation error information.
	StringSlice() ([]string, error)

	// Uint64 returns the property value as an uint64 along with any type negotiation error information.
	Uint64() (uint64, error)

	// Float64 returns the property value as a float64 along with any type negotiation error information.
	Float64() (float64, error)

	// String returns the property value as a string along with any type negotiation error information.
	String() (string, error)

	// Time returns the property value as time.Time along with any type negotiation error information.
	Time() (time.Time, error)

	// Any returns the property value typed as any. This function may return a null reference.
	Any() any
}

// Entity is a base type for all DAWGS database entity types. This type is often only useful when writing a query
// generator.
type Entity interface {
	// ID returns the ID of the Entity.
	ID() ID

	// Properties returns the Properties instance for this Entity.
	Properties() *Properties

	// SizeOf returns the approximate size of this Entity instance.
	SizeOf() size.Size
}

type NodeUpdate struct {
	Node               *Node
	IdentityKind       Kind
	IdentityProperties []string
}

func (s NodeUpdate) Key() (string, error) {
	key := strings.Builder{}

	slices.Sort(s.IdentityProperties)

	for _, identityProperty := range s.IdentityProperties {
		if propertyValue, err := s.Node.Properties.Get(identityProperty).String(); err != nil {
			return "", err
		} else {
			key.WriteString(propertyValue)
		}
	}

	return key.String(), nil
}

type RelationshipUpdate struct {
	Relationship            *Relationship
	IdentityProperties      []string
	Start                   *Node
	StartIdentityKind       Kind
	StartIdentityProperties []string
	End                     *Node
	EndIdentityKind         Kind
	EndIdentityProperties   []string
}

func (s RelationshipUpdate) Key() (string, error) {
	var (
		key             = strings.Builder{}
		startNodeUpdate = NodeUpdate{
			Node:               s.Start,
			IdentityKind:       s.StartIdentityKind,
			IdentityProperties: s.StartIdentityProperties,
		}

		endNodeUpdate = NodeUpdate{
			Node:               s.End,
			IdentityKind:       s.EndIdentityKind,
			IdentityProperties: s.EndIdentityProperties,
		}
	)

	key.WriteString(s.Relationship.Kind.String())

	if startKey, err := startNodeUpdate.Key(); err != nil {
		return "", err
	} else if endKey, err := endNodeUpdate.Key(); err != nil {
		return "", err
	} else {
		key.WriteString(startKey)

		slices.Sort(s.IdentityProperties)

		for _, identityProperty := range s.IdentityProperties {
			if propertyValue, err := s.Relationship.Properties.Get(identityProperty).String(); err != nil {
				return "", err
			} else {
				key.WriteString(propertyValue)
			}
		}

		key.WriteString(endKey)
		return key.String(), nil
	}
}

func (s RelationshipUpdate) IdentityPropertiesMap() map[string]any {
	identityPropertiesMap := make(map[string]any, len(s.IdentityProperties))

	for _, identityPropertyKey := range s.IdentityProperties {
		identityPropertiesMap[identityPropertyKey] = s.Relationship.Properties.Get(identityPropertyKey).Any()
	}

	return identityPropertiesMap
}

func (s RelationshipUpdate) StartIdentityPropertiesMap() map[string]any {
	identityPropertiesMap := make(map[string]any, len(s.StartIdentityProperties))

	for _, identityPropertyKey := range s.StartIdentityProperties {
		identityPropertiesMap[identityPropertyKey] = s.Start.Properties.Get(identityPropertyKey).Any()
	}

	return identityPropertiesMap
}

func (s RelationshipUpdate) EndIdentityPropertiesMap() map[string]any {
	identityPropertiesMap := make(map[string]any, len(s.EndIdentityProperties))

	for _, identityPropertyKey := range s.EndIdentityProperties {
		identityPropertiesMap[identityPropertyKey] = s.End.Properties.Get(identityPropertyKey).Any()
	}

	return identityPropertiesMap
}

type Batch interface {
	// WithGraph scopes the transaction to a specific graph. If the driver for the transaction does not support
	// multiple  graphs the resulting transaction will target the default graph instead and this call becomes a no-op.
	WithGraph(graphSchema Graph) Batch

	// CreateNode creates a new Node in the database and returns the creation as a NodeResult.
	CreateNode(node *Node) error

	// DeleteNode deletes a node by the given ID.
	DeleteNode(id ID) error

	// Nodes begins a batch query that can be used to update or delete nodes.
	Nodes() NodeQuery

	// Relationships begins a batch query that can be used to update or delete relationships.
	Relationships() RelationshipQuery

	// UpdateNodeBy is a stop-gap until the query interface can better support targeted batch create-update operations.
	// Nodes identified by the NodeUpdate criteria will either be updated or in the case where the node does not yet
	// exist, created.
	UpdateNodeBy(update NodeUpdate) error

	// TODO: Existing batch logic expects this to perform an upsert on conficts with (start_id, end_id, kind). This is incorrect and should be refactored
	CreateRelationship(relationship *Relationship) error

	// CreateRelationshipByIDs creates a new Relationship from the start Node to the end Node with the given Kind and
	// Properties and returns the creation as a RelationshipResult.
	//
	// Deprecated: Use CreateRelationship
	CreateRelationshipByIDs(startNodeID, endNodeID ID, kind Kind, properties *Properties) error

	// DeleteRelationship deletes a relationship by the given ID.
	DeleteRelationship(id ID) error

	// UpdateRelationshipBy is a stop-gap until the query interface can better support targeted batch create-update
	// operations. Relationships identified by the RelationshipUpdate criteria will either be updated or in the case
	// where the relationship does not yet exist, created.
	UpdateRelationshipBy(update RelationshipUpdate) error

	// Commit calls to commit this batch transaction right away.
	Commit() error
}

// Transaction is an interface that contains all operations that may be executed against a DAWGS driver. DAWGS drivers are
// expected to support all Transaction operations in-transaction.
type Transaction interface {
	// WithGraph scopes the transaction to a specific graph. If the driver for the transaction does not support
	// multiple  graphs the resulting transaction will target the default graph instead and this call becomes a no-op.
	WithGraph(graphSchema Graph) Transaction

	// CreateNode creates a new Node in the database and returns the creation as a NodeResult.
	CreateNode(properties *Properties, kinds ...Kind) (*Node, error)

	// UpdateNode updates a Node in the database with the given Node by ID. UpdateNode will not create missing Node
	// entries in the database. Use CreateNode first to create a new Node.
	UpdateNode(node *Node) error

	// Nodes creates a new NodeQuery and returns it.
	Nodes() NodeQuery

	// CreateRelationshipByIDs creates a new Relationship from the start Node to the end Node with the given Kind and
	// Properties and returns the creation as a RelationshipResult.
	CreateRelationshipByIDs(startNodeID, endNodeID ID, kind Kind, properties *Properties) (*Relationship, error)

	// UpdateRelationship updates a Relationship in the database with the given Relationship by ID. UpdateRelationship
	// will not create missing Relationship entries in the database. Use CreateRelationship first to create a new
	// Relationship.
	UpdateRelationship(relationship *Relationship) error

	// Relationships creates a new RelationshipQuery and returns it.
	Relationships() RelationshipQuery

	// Raw allows a user to pass raw queries directly to the database without translation.
	Raw(query string, parameters map[string]any) Result

	// Query allows a user to execute a given cypher query that will be translated to the target database.
	Query(query string, parameters map[string]any) Result

	// Commit calls to commit this transaction right away.
	Commit() error

	// TraversalMemoryLimit returns the traversal memory limit of
	TraversalMemoryLimit() size.Size
}

// TransactionDelegate represents a transactional database context actor. Errors returned from a TransactionDelegate
// result in the rollback of write enabled transactions. Successful execution of a TransactionDelegate (nil error
// return value) results in a transactional commit of work done within the TransactionDelegate.
type TransactionDelegate func(tx Transaction) error

// BatchDelegate represents a transactional database context actor.
type BatchDelegate func(batch Batch) error

// TransactionConfig is a generic configuration that may apply to all supported databases.
type TransactionConfig struct {
	Timeout      time.Duration
	DriverConfig any
}

// TransactionOption is a function that represents a configuration setting for the underlying database transaction.
type TransactionOption func(config *TransactionConfig)

// Database is a high-level interface representing transactional entry-points into DAWGS driver implementations.
type Database interface {
	// SetWriteFlushSize sets a new write flush interval on the current driver
	SetWriteFlushSize(interval int)

	// SetBatchWriteSize sets a new batch write interval on the current driver
	SetBatchWriteSize(interval int)

	// ReadTransaction opens up a new read transactional context in the database and then defers the context to the
	// given logic function.
	ReadTransaction(ctx context.Context, txDelegate TransactionDelegate, options ...TransactionOption) error

	// WriteTransaction opens up a new write transactional context in the database and then defers the context to the
	// given logic function.
	WriteTransaction(ctx context.Context, txDelegate TransactionDelegate, options ...TransactionOption) error

	// BatchOperation opens up a new write transactional context in the database and then defers the context to the
	// given logic function. Batch operations are fundamentally different between databases supported by DAWGS,
	// necessitating a different interface that lacks many of the convenience features of a regular read or write
	// transaction.
	BatchOperation(ctx context.Context, batchDelegate BatchDelegate) error

	// AssertSchema will apply the given schema to the underlying database.
	AssertSchema(ctx context.Context, dbSchema Schema) error

	// SetDefaultGraph sets the default graph namespace for the connection.
	SetDefaultGraph(ctx context.Context, graphSchema Graph) error

	// Run allows a user to pass statements directly to the database. Since results may rely on a transactional context
	// only an error is returned from this function
	Run(ctx context.Context, query string, parameters map[string]any) error

	// Close closes the database context and releases any pooled resources held by the instance.
	Close(ctx context.Context) error
}
