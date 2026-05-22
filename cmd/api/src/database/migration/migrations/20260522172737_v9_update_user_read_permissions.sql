-- +goose Up
INSERT INTO permissions (authority, name, created_at, updated_at)
VALUES ('auth',
        'ReadUsersMinimal',
        current_timestamp,
        current_timestamp)
ON CONFLICT DO NOTHING;

DELETE FROM roles_permissions
WHERE permission_id = (SELECT id 
                        FROM permissions 
                        WHERE authority = 'auth' AND name = 'ReadUsers')
    AND role_id IN (SELECT id 
                    FROM roles 
                    WHERE name IN ('User', 'Read-Only', 'Power User'));

INSERT INTO roles_permissions (role_id, permission_id)
SELECT r.id, p.id
FROM roles r
JOIN permissions p
    ON (
      (r.name IN ('User', 'Read-Only', 'Power User', 'Auditor', 'Administrator') AND (p.authority, p.name) IN (
            ('auth', 'ReadUsersMinimal')
      ))
)
ON CONFLICT DO NOTHING;

-- +goose Down
