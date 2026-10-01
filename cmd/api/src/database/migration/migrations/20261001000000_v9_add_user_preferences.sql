-- +goose Up
CREATE TABLE user_preferences (
    -- users.id stores UUIDs as text; the foreign key must use the same type.
    user_id text PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE,
    ui_storage bytea NOT NULL DEFAULT ''::bytea,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    CONSTRAINT user_preferences_ui_storage_size CHECK (octet_length(ui_storage) <= 65536)
);

-- +goose Down
DROP TABLE user_preferences;
