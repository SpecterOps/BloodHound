-- +goose Up

-- Add ON DELETE CASCADE constraint to the roles_permissions table 
ALTER TABLE roles_permissions
DROP CONSTRAINT fk_roles_permissions_permission;

ALTER TABLE roles_permissions
ADD CONSTRAINT fk_roles_permissions_permission 
FOREIGN KEY (permission_id) REFERENCES permissions(id) ON DELETE CASCADE;

INSERT INTO permissions(authority, name, created_at, updated_at)
VALUES ('analysis', 'Request', current_timestamp, current_timestamp)
ON CONFLICT DO NOTHING;

INSERT INTO roles_permissions(role_id, permission_id)
SELECT r.id, p.id 
FROM roles r 
    JOIN permissions p
        ON (p.authority, p.name) = ('analysis', 'Request')
WHERE r.name IN ('Administrator')
ON CONFLICT DO NOTHING;


-- +goose Down
DELETE FROM permissions
WHERE authority = 'analysis' AND name = 'Request';

-- ON DELETE CASCADE should delete the role_permission association automatically 

-- Remove ON DELETE CASCADE constraint
ALTER TABLE roles_permissions
DROP CONSTRAINT fk_roles_permissions_permission;

ALTER TABLE roles_permissions
ADD CONSTRAINT fk_roles_permissions_permission 
FOREIGN KEY (permission_id) REFERENCES permissions(id);


