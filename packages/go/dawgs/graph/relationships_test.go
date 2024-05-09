package graph_test

import (
	"github.com/specterops/bloodhound/dawgs/graph"
	"github.com/stretchr/testify/require"
	"testing"
	"unsafe"
)

func Test_RelationshipSizeOf(t *testing.T) {
	relationship := graph.Relationship{ID: graph.ID(1)}
	initialSize := int64(relationship.SizeOf())

	// ensure that the initial size accounts for all the struct fields
	sizeOfIDs := 3 * int64(unsafe.Sizeof(relationship.StartID))
	sizeOfKind := int64(unsafe.Sizeof(relationship.Kind))
	sizeOfProperties := int64(unsafe.Sizeof(relationship.Properties))
	require.Greater(t, initialSize, sizeOfIDs+sizeOfKind+sizeOfProperties)

	// ID, StartID, EndID and Kind have zero-value sizes that aren't impacted by
	// reassignment to a non-zero value. Therefore, skipping testing of those fields.

	// ensure that reassignment of the Properties field affects size
	relationship.Properties = graph.NewProperties()
	newSize := int64(relationship.SizeOf())
	require.Greater(t, newSize, initialSize)
}
