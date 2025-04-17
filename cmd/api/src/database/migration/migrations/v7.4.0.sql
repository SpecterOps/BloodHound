-- Add asset_group_tag_selector_nodes table
CREATE TABLE IF NOT EXISTS asset_group_tag_selector_nodes
(
  selector_id int NOT NULL,
  node_id bigint NOT NULL,
  certified int NOT NULL DEFAULT 0,
  certified_by text,
  source int,
  created_at timestamp with time zone,
  updated_at timestamp with time zone,
  CONSTRAINT fk_asset_group_tag_selectors_asset_group_tag_selector_nodes FOREIGN KEY (selector_id) REFERENCES asset_group_tag_selectors(id) ON DELETE CASCADE,
  PRIMARY KEY (selector_id, node_id)
  );
