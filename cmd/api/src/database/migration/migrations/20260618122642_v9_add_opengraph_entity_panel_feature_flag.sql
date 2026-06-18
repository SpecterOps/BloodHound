-- +goose Up
INSERT INTO feature_flags (created_at, updated_at, key, name, description, enabled, user_updatable)
VALUES (current_timestamp, current_timestamp, 'opengraph_entity_panel', 'OpenGraph Entity Panel', 'Enable/Disable Customized Entity Panel Support for OpenGraph Extensions', false, false);

-- +goose Down
DELETE FROM feature_flags WHERE key = 'opengraph_entity_panel';
