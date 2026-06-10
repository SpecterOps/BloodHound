-- +goose Up
INSERT INTO permissions (authority, name, created_at, updated_at)
VALUES ('auth',
        'ReadProviders',
        current_timestamp,
        current_timestamp)
ON CONFLICT DO NOTHING;

INSERT INTO roles_permissions (role_id, permission_id)
SELECT r.id, p.id
FROM roles r
JOIN permissions p
    ON (
      (r.name IN ('Auditor') AND (p.authority, p.name) IN (
            ('auth', 'ReadProviders')
      ))
)
ON CONFLICT DO NOTHING;

INSERT INTO roles_permissions (role_id, permission_id)
SELECT r.role_id, p.id
FROM roles_permissions r
JOIN permissions f
    ON f.id = r.permission_id
JOIN permissions p
    ON (p.authority, p.name) = ('auth', 'ReadProviders')
WHERE (f.authority, f.name) = ('auth', 'ManageProviders')
ON CONFLICT DO NOTHING;

-- +goose Down
DELETE FROM roles_permissions
WHERE permission_id = (SELECT id FROM permissions WHERE authority = 'auth' AND name = 'ReadProviders');

DELETE FROM permissions WHERE authority = 'auth' AND name = 'ReadProviders';
