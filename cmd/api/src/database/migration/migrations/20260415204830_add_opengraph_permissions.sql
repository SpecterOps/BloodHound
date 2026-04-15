-- Add OpenGraph permissions to permissions table
-- +goose Up
INSERT INTO permissions(created_at, updated_at, authority, name)
VALUES (
        current_timestamp,
        current_timestamp,
        'opengraph',
        'Read'
       ), 
       (
        current_timestamp,
        current_timestamp,
        'opengraph',
        'Write'
       )
ON CONFLICT DO NOTHING;

-- +goose Down
DELETE FROM permissions 
WHERE authority = 'opengraph' 
  AND name IN ('Read', 'Write');
