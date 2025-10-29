-- Add Audit Log permission and Auditor role 
INSERT INTO permissions (authority, name, created_at, updated_at) VALUES ('audit_log', 'Read', current_timestamp, current_timestamp) ON CONFLICT DO NOTHING;

INSERT INTO roles (name, description, created_at, updated_at) VALUES 
 ('Auditor', 'Can read data and audit logs', current_timestamp, current_timestamp) ON CONFLICT DO NOTHING;

INSERT INTO roles_permissions (role_id, permission_id)
SELECT r.id, p.id
FROM roles r
JOIN permissions p
  ON (
    (r.name = 'Auditor' AND (p.authority, p.name) IN (
        ('app', 'ReadAppConfig'),
        ('risks', 'GenerateReport'),
        ('audit_log', 'Read'),
        ('auth', 'CreateToken'),
        ('auth', 'ManageSelf'),
        ('auth', 'ReadUsers'),
        ('graphdb', 'Read'),
        ('saved_queries', 'Read'),
        ('clients', 'Read')
    ))) 
ON CONFLICT DO NOTHING;

INSERT INTO roles_permissions (role_id, permission_id)
VALUES ((SELECT id FROM roles WHERE roles.name = 'Administrator'),
        (SELECT id FROM permissions WHERE permissions.authority = 'audit_log' and permissions.name = 'Read'))
ON CONFLICT DO NOTHING;