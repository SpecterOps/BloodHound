-- +goose Up
INSERT INTO permissions (authority, name, created_at, updated_at)
VALUES ('graphdb',
        'IngestRead',
        current_timestamp,
        current_timestamp)
ON CONFLICT DO NOTHING;

INSERT INTO roles_permissions (role_id, permission_id)
SELECT r.id, p.id
FROM roles r
JOIN permissions p
    ON (
      (r.name IN ('Auditor') AND (p.authority, p.name) IN (
            ('graphdb', 'IngestRead')
      ))
)
ON CONFLICT DO NOTHING;

INSERT INTO roles_permissions (role_id, permission_id)
SELECT r.role_id, p.id
FROM roles_permissions r
JOIN permissions f
    ON f.id = r.permission_id
JOIN permissions p
    ON (p.authority, p.name) = ('graphdb', 'IngestRead')
WHERE (f.authority, f.name) = ('graphdb', 'Ingest')
ON CONFLICT DO NOTHING;

-- +goose Down
DELETE FROM roles_permissions
WHERE permission_id = (SELECT id FROM permissions WHERE authority = 'graphdb' AND name = 'IngestRead');

DELETE FROM permissions WHERE authority = 'graphdb' AND name = 'IngestRead';
