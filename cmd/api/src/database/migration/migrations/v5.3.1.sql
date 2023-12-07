CREATE INDEX IF NOT EXISTS idx_asset_group_collections_asset_group_id ON asset_group_collections USING btree (asset_group_id);
CREATE INDEX IF NOT EXISTS idx_asset_group_collections_created_at ON asset_group_collections USING btree (created_at);
CREATE INDEX IF NOT EXISTS idx_asset_group_collections_updated_at ON asset_group_collections USING btree (updated_at);

CREATE INDEX IF NOT EXISTS idx_asset_group_collection_entries_asset_group_collection_id ON asset_group_collection_entries USING btree (asset_group_collection_id);
CREATE INDEX IF NOT EXISTS idx_asset_group_collection_entries_created_at ON asset_group_collection_entries USING btree (created_at);
CREATE INDEX IF NOT EXISTS idx_asset_group_collection_entries_updated_at ON asset_group_collection_entries USING btree (updated_at);

TRUNCATE asset_group_collection_entries;
