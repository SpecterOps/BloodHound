package changestream

import (
	"context"
	"fmt"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"time"
)

const (
	assertTableSQL = `
create table if not exists node_change_stream (
	id bigint generated always as identity not null,
	target_node text not null,
	kind_ids smallint[] not null,
	properties_hash bytea not null,
	property_fields text[] not null,
	change_type integer not null,
	created_at timestamp with time zone not null,

	primary key (id, created_at)
) partition by range (created_at);

create index if not exists node_change_stream_target_node_index on node_change_stream using hash (target_node);
create index if not exists node_change_stream_created_at_index on node_change_stream using btree (created_at);

create table if not exists edge_change_stream (
	id bigint generated always as identity not null,
	target_node text not null,
	related_node text not null,
	kind_id smallint not null,
	properties_hash bytea not null,
	property_fields text[] not null,
	identity_hash bytea not null,
	change_type integer not null,
	created_at timestamp with time zone not null,

	primary key (id, created_at)
) partition by range (created_at);

create index if not exists edge_change_stream_target_node_index on edge_change_stream using hash (target_node);
create index if not exists edge_change_stream_edge_index on edge_change_stream using hash (identity_hash);
create index if not exists edge_change_stream_created_at_index on edge_change_stream using btree (created_at);

-- Below are some optional indexes that were not prioritized but may be useful later
--
-- create index if not exists change_stream_related_node_index on change_stream using hash (related_node);
-- create index if not exists change_stream_kind_ids_index on node using gin (kind_ids);
-- create index if not exists change_stream_properties_hash_index on change_stream using hash (properties_hash);
--
`

	insertNodeChangeSQL = `insert into node_change_stream (target_node, kind_ids, property_fields, properties_hash, change_type, created_at) select unnest($1::text[]), unnest($2::text[])::int2[], unnest($3::text[])::text[], unnest($4::bytea[]), 0, now();`

	insertEdgeChangeSQL = `insert into edge_change_stream (target_node, related_node, kind_id, property_fields, properties_hash, identity_hash, change_type, created_at) select unnest($1::text[]), unnest($2::text[]), unnest($3::int2[]), unnest($4::text[])::text[], unnest($5::bytea[]), unnest($6::bytea[]), 0, now();`

	lastNodeChangeSQL = `select cs.properties_hash != $2, cs.change_type from node_change_stream cs where cs.target_node = $1 order by created_at desc limit 1;`

	lastEdgeChangeSQL = `select cs.properties_hash != $2, cs.change_type from edge_change_stream cs where cs.identity_hash = $1 order by created_at desc limit 1;`

	createChangeStreamPartitionsSQLFmt = `
create table if not exists node_change_stream_%s partition of node_change_stream for values from ('%s') to ('%s');
create table if not exists edge_change_stream_%s partition of edge_change_stream for values from ('%s') to ('%s');
`

	disableSynchronousCommitSQL = `set local synchronous_commit = 'off';`
)

func disableSynchronousCommitForConn(ctx context.Context, conn *pgx.Conn) error {
	_, err := conn.Exec(ctx, disableSynchronousCommitSQL)
	return err
}

const (
	tablePartitionRangeDuration   = time.Hour
	tablePartitionRangeFormatStr  = "2006-01-02 15:00:00"
	tablePartitionSuffixFormatStr = "2006_01_02_15"
)

func nodeChangePartitionName(now time.Time) string {
	return "node_change_stream_" + now.Format(tablePartitionSuffixFormatStr)
}

func edgeChangePartitionName(now time.Time) string {
	return "edge_change_stream_" + now.Format(tablePartitionSuffixFormatStr)
}

func shouldAssertNextPartition(lastPartitionAssert time.Time) bool {
	var (
		now                     = time.Now()
		lastPartitionRangeStr = lastPartitionAssert.Format(tablePartitionRangeFormatStr)
		nowPartitionRangeStr  = now.Format(tablePartitionRangeFormatStr)
	)

	return lastPartitionRangeStr != nowPartitionRangeStr
}

func assertChangelogPartition(ctx context.Context, pgxPool *pgxpool.Pool) error {
	var (
		now                  = time.Now()
		partitionTableSuffix = now.Format(tablePartitionSuffixFormatStr)
		partitionRangeStart  = now.Format(tablePartitionRangeFormatStr)
		partitionRangeEnd    = now.Add(tablePartitionRangeDuration).Format(tablePartitionRangeFormatStr)
		assertSQL            = fmt.Sprintf(createChangeStreamPartitionsSQLFmt, partitionTableSuffix, partitionRangeStart, partitionRangeEnd, partitionTableSuffix, partitionRangeStart, partitionRangeEnd)
		_, err               = pgxPool.Exec(ctx, assertSQL)
	)

	return err
}

func newPGXPool(ctx context.Context, connectionStr string) (*pgxpool.Pool, error) {
	poolCtx, done := context.WithTimeout(ctx, time.Second*10)
	defer done()

	poolCfg, err := pgxpool.ParseConfig(connectionStr)
	if err != nil {
		return nil, err
	}

	poolCfg.MinConns = 5
	poolCfg.MaxConns = 5

	// Disable sync commits for the changelog
	poolCfg.AfterConnect = disableSynchronousCommitForConn

	return pgxpool.NewWithConfig(poolCtx, poolCfg)
}
