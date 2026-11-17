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

package post_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/specterops/bloodhound/packages/go/analysis/post"
	"github.com/specterops/bloodhound/packages/go/graphschema/common"
	"github.com/specterops/dawgs/graph"
	"github.com/stretchr/testify/require"
)

type sinkTestDatabase struct {
	graph.Database
	graph.Batch
	relationships map[graph.ID]bool
	created       []post.EnsureRelationshipJob
	batchError    error
	createError   error
	flushError    error
	deleteError   error
	writerStarted chan struct{}
	writerRelease chan struct{}
	flushStarted  chan struct{}
	flushRelease  chan struct{}
}

func (s *sinkTestDatabase) BatchOperation(ctx context.Context, delegate graph.BatchDelegate, _ ...graph.BatchOption) error {
	if s.batchError != nil {
		close(s.writerStarted)
		<-s.writerRelease
		return s.batchError
	}
	if err := delegate(s); err != nil {
		return err
	}
	if s.flushStarted != nil {
		close(s.flushStarted)
		s.flushStarted = nil
		<-s.flushRelease
	}
	return s.flushError
}

func (s *sinkTestDatabase) CreateRelationshipByIDs(fromID, toID graph.ID, kind graph.Kind, properties *graph.Properties) error {
	if s.createError != nil {
		close(s.writerStarted)
		<-s.writerRelease
		return s.createError
	}
	s.created = append(s.created, post.EnsureRelationshipJob{
		FromID: fromID, ToID: toID, Kind: kind, RelProperties: properties.Map,
	})
	return nil
}

func (s *sinkTestDatabase) DeleteRelationship(relationshipID graph.ID) error {
	if s.deleteError != nil {
		return s.deleteError
	}
	delete(s.relationships, relationshipID)
	return nil
}

func newSinkTestDatabase() (*sinkTestDatabase, *post.Tracker) {
	var (
		database = &sinkTestDatabase{relationships: map[graph.ID]bool{77: true}}
		builder  = post.NewTrackerBuilder()
	)
	builder.TrackEdge(77, 1, 2, graph.StringKind("TrackedRelationship"), nil)
	return database, builder.Build()
}

func waitForSinkResult[T any](t *testing.T, resultChannel <-chan T) T {
	t.Helper()
	select {
	case result := <-resultChannel:
		return result
	case <-time.After(5 * time.Second):
		t.Fatal("sink operation did not finish")
		var zero T
		return zero
	}
}

func finishSink(t *testing.T, sink *post.FilteredRelationshipSink) error {
	t.Helper()
	var resultChannel = make(chan error, 1)
	go func() { resultChannel <- sink.Done() }()
	return waitForSinkResult(t, resultChannel)
}

func TestFilteredRelationshipSink_DonePrunesEmptySuccessfulGeneration(t *testing.T) {
	var (
		database, tracker = newSinkTestDatabase()
		sink              = post.NewFilteredRelationshipSink(t.Context(), "test", database, tracker)
	)
	require.NoError(t, finishSink(t, sink))
	require.Empty(t, database.relationships)
	require.NoError(t, sink.Done())
	require.NoError(t, sink.Abort())
	require.False(t, sink.Submit(t.Context(), post.EnsureRelationshipJob{}))
}

func TestFilteredRelationshipSink_DonePreservesVisitedEdgesAndInsertsNewEdges(t *testing.T) {
	var (
		database, tracker = newSinkTestDatabase()
		kind              = graph.StringKind("TrackedRelationship")
		sink              = post.NewFilteredRelationshipSink(t.Context(), "test", database, tracker)
	)
	require.True(t, sink.Submit(t.Context(), post.EnsureRelationshipJob{FromID: 1, ToID: 2, Kind: kind}))
	require.True(t, sink.Submit(t.Context(), post.EnsureRelationshipJob{
		FromID: 3, ToID: 4, Kind: kind, RelProperties: map[string]any{"isacl": true},
	}))
	require.NoError(t, finishSink(t, sink))
	require.True(t, database.relationships[77])
	require.Len(t, database.created, 1)
	require.Equal(t, graph.ID(3), database.created[0].FromID)
	require.Equal(t, graph.ID(4), database.created[0].ToID)
	require.Equal(t, true, database.created[0].RelProperties["isacl"])
	require.IsType(t, time.Time{}, database.created[0].RelProperties[common.FirstSeen.String()])
	require.Equal(t, int32(1), *sink.Stats().RelationshipsCreated[kind])
}

func TestFilteredRelationshipSink_AbortPreservesUnvisitedEdges(t *testing.T) {
	var (
		database, tracker = newSinkTestDatabase()
		sink              = post.NewFilteredRelationshipSink(t.Context(), "test", database, tracker)
		resultChannel     = make(chan error, 1)
	)
	go func() {
		resultChannel <- sink.Abort()
	}()
	require.NoError(t, waitForSinkResult(t, resultChannel))
	require.True(t, database.relationships[77])
	require.NoError(t, sink.Abort())
	require.ErrorIs(t, sink.Done(), context.Canceled)
	require.False(t, sink.Submit(t.Context(), post.EnsureRelationshipJob{}))
}

func TestFilteredRelationshipSink_ParentCancellationPreservesUnvisitedEdges(t *testing.T) {
	var (
		database, tracker = newSinkTestDatabase()
		ctx, cancel       = context.WithCancel(t.Context())
		sink              = post.NewFilteredRelationshipSink(ctx, "test", database, tracker)
	)
	cancel()
	require.ErrorIs(t, finishSink(t, sink), context.Canceled)
	require.True(t, database.relationships[77])
	require.False(t, sink.Submit(t.Context(), post.EnsureRelationshipJob{}))
}

func TestFilteredRelationshipSink_WriterFailureUnblocksSubmissionsAndPreservesEdges(t *testing.T) {
	for _, failureMode := range []string{"batch", "create"} {
		t.Run(failureMode, func(t *testing.T) {
			var (
				database, tracker = newSinkTestDatabase()
				writerError       = errors.New("writer failed")
				resultChannel     = make(chan bool, 1)
			)
			database.writerStarted = make(chan struct{})
			database.writerRelease = make(chan struct{})
			if failureMode == "batch" {
				database.batchError = writerError
			} else {
				database.createError = writerError
			}
			var sink = post.NewFilteredRelationshipSink(t.Context(), "test", database, tracker)
			go func() {
				for submitted := 0; submitted < 10000; submitted++ {
					if !sink.Submit(context.Background(), post.EnsureRelationshipJob{
						FromID: 3, ToID: 4, Kind: graph.StringKind("NewRelationship"),
					}) {
						resultChannel <- false
						return
					}
				}
				resultChannel <- true
			}()
			waitForSinkResult(t, database.writerStarted)
			close(database.writerRelease)
			require.False(t, waitForSinkResult(t, resultChannel))
			require.ErrorIs(t, finishSink(t, sink), writerError)
			require.True(t, database.relationships[77])
			require.ErrorIs(t, sink.Done(), writerError)
			require.ErrorIs(t, sink.Abort(), writerError)
		})
	}
}

func TestFilteredRelationshipSink_FlushFailurePreservesUnvisitedEdges(t *testing.T) {
	var (
		database, tracker = newSinkTestDatabase()
		writerError       = errors.New("batch flush failed")
	)
	database.flushError = writerError
	var sink = post.NewFilteredRelationshipSink(t.Context(), "test", database, tracker)
	require.True(t, sink.Submit(t.Context(), post.EnsureRelationshipJob{
		FromID: 3, ToID: 4, Kind: graph.StringKind("NewRelationship"),
	}))
	require.ErrorIs(t, finishSink(t, sink), writerError)
	require.True(t, database.relationships[77])
}

func TestFilteredRelationshipSink_DoneReturnsPruningError(t *testing.T) {
	var (
		database, tracker = newSinkTestDatabase()
		deleteError       = errors.New("deleting relationship failed")
	)
	database.deleteError = deleteError
	var sink = post.NewFilteredRelationshipSink(t.Context(), "test", database, tracker)
	require.ErrorIs(t, finishSink(t, sink), deleteError)
	require.True(t, database.relationships[77])
}

func TestFilteredRelationshipSink_SubmitDuringShutdownIsRejected(t *testing.T) {
	var (
		database, tracker = newSinkTestDatabase()
		flushStarted      = make(chan struct{})
		resultChannel     = make(chan error, 1)
	)
	database.flushStarted = flushStarted
	database.flushRelease = make(chan struct{})
	var sink = post.NewFilteredRelationshipSink(t.Context(), "test", database, tracker)
	go func() { resultChannel <- sink.Done() }()
	waitForSinkResult(t, flushStarted)
	require.False(t, sink.Submit(context.Background(), post.EnsureRelationshipJob{}))
	close(database.flushRelease)
	require.NoError(t, waitForSinkResult(t, resultChannel))
	require.Empty(t, database.relationships)
}
