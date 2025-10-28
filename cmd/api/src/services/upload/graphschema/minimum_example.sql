-- Insert collector
INSERT INTO collectors (name, display_name, version) 
VALUES ('sharp-hound', 'SharpHound', 'v1.0.0');

-- Insert schema
INSERT INTO graph_schemas (collector_id, name) 
VALUES (1, 'Active Directory Schema v1.0');

-- Insert properties
INSERT INTO schema_properties (schema_id, symbol, display_name, representation, data_type, description) VALUES
(1, 'ObjectID', 'Object ID', 'objectid', 'string', 'Unique identifier for the object'),
(1, 'Name', 'Name', 'name', 'string', 'Display name of the object'),
(1, 'Enabled', 'Enabled', 'enabled', 'bool', 'Whether the account is enabled'),
(1, 'LastLogon', 'Last Logon', 'lastlogon', 'timestamp', 'Last successful logon time'),
(1, 'OperatingSystem', 'Operating System', 'operatingsystem', 'string', 'Computer operating system');

-- Insert node kinds
INSERT INTO schema_node_kinds (schema_id, symbol, representation, description) VALUES
(1, 'User', 'User', 'Active Directory user account'),
(1, 'Computer', 'Computer', 'Active Directory computer object'),
(1, 'Group', 'Group', 'Active Directory security group');

-- Insert node-property associations
INSERT INTO schema_node_properties (node_id, property_id, is_required) VALUES
-- User properties
(1, 1, true),  -- ObjectID required
(1, 2, true),  -- Name required  
(1, 3, false), -- Enabled optional
(1, 4, false), -- LastLogon optional
-- Computer properties
(2, 1, true),  -- ObjectID required
(2, 2, true),  -- Name required
(2, 3, false), -- Enabled optional
(2, 5, false), -- OperatingSystem optional
-- Group properties  
(3, 1, true),  -- ObjectID required
(3, 2, true);  -- Name required

-- Insert relationship kinds  
INSERT INTO schema_relationship_kinds (schema_id, symbol, display_name, description) VALUES
(1, 'HasSession', 'Has Session', 'Computer has an active user session'),
(1, 'MemberOf', 'Member Of', 'Principal is a member of a group'),
(1, 'GenericWrite', 'Generic Write', 'Principal has write access to target');

-- Insert relationship constraints
INSERT INTO schema_relationship_constraints (relationship_id, source_id, target_id) VALUES
(1, 1, 2), -- HasSession: User -> Computer
(2, 1, 3), -- MemberOf: User -> Group
(2, 2, 3), -- MemberOf: Computer -> Group  
(3, 1, 1), -- GenericWrite: User -> User
(3, 1, 2), -- GenericWrite: User -> Computer
(3, 1, 3); -- GenericWrite: User -> Group