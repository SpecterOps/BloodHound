-- Add OpenGraph Write permissions to Admin role
-- +goose Up
INSERT INTO roles_permissions (role_id, permission_id)
SELECT r.id, p.id
FROM roles r
JOIN permissions p
ON (p.authority, p.name) = ('opengraph', 'Write')
WHERE r.name IN ('Administrator')
ON CONFLICT DO NOTHING;

-- +goose Down
DELETE FROM roles_permissions
WHERE role_id IN (
    SELECT id FROM roles 
    WHERE name IN ('Administrator')
)
AND permission_id = (
    SELECT id FROM permissions 
    WHERE authority = 'opengraph' AND name = 'Write'
);