-- Add OpenGraph Read permissions to specific roles 
-- +goose Up
INSERT INTO roles_permissions (role_id, permission_id)
SELECT r.id, p.id
FROM roles r
JOIN permissions p
ON (p.authority, p.name) = ('opengraph', 'Read')
WHERE r.name IN ('Administrator', 'User', 'Read-Only', 'Power User', 'Auditor')
ON CONFLICT DO NOTHING;

-- +goose Down
DELETE FROM roles_permissions
WHERE role_id IN (
    SELECT id FROM roles 
    WHERE name IN ('Administrator', 'User', 'Read-Only', 'Power User', 'Auditor')
)
AND permission_id = (
    SELECT id FROM permissions 
    WHERE authority = 'opengraph' AND name = 'Read'
);
