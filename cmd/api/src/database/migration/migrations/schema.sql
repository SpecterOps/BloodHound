-- Copyright 2023 Specter Ops, Inc.
--
-- Licensed under the Apache License, Version 2.0
-- you may not use this file except in compliance with the License.
-- You may obtain a copy of the License at
--
--     http://www.apache.org/licenses/LICENSE-2.0
--
-- Unless required by applicable law or agreed to in writing, software
-- distributed under the License is distributed on an "AS IS" BASIS,
-- WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
-- See the License for the specific language governing permissions and
-- limitations under the License.
--
-- SPDX-License-Identifier: Apache-2.0

CREATE TABLE IF NOT EXISTS ad_data_quality_aggregations (
    domains bigint,
    users bigint,
    groups bigint,
    computers bigint,
    ous bigint,
    containers bigint,
    gpos bigint,
    acls bigint,
    sessions bigint,
    relationships bigint,
    session_completeness numeric,
    local_group_completeness numeric,
    run_id text,
    id integer NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone
);

CREATE SEQUENCE IF NOT EXISTS ad_data_quality_aggregations_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;
ALTER SEQUENCE ad_data_quality_aggregations_id_seq OWNED BY ad_data_quality_aggregations.id;

CREATE TABLE IF NOT EXISTS ad_data_quality_stats (
    domain_sid text,
    users bigint,
    groups bigint,
    computers bigint,
    ous bigint,
    containers bigint,
    gpos bigint,
    acls bigint,
    sessions bigint,
    relationships bigint,
    session_completeness numeric,
    local_group_completeness numeric,
    run_id text,
    id integer NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone
);

CREATE SEQUENCE IF NOT EXISTS ad_data_quality_stats_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;
ALTER SEQUENCE ad_data_quality_stats_id_seq OWNED BY ad_data_quality_stats.id;

CREATE TABLE IF NOT EXISTS asset_group_collection_entries (
    asset_group_collection_id bigint,
    object_id text,
    node_label text,
    properties jsonb,
    id bigint NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone
);

CREATE SEQUENCE IF NOT EXISTS asset_group_collection_entries_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;
ALTER SEQUENCE asset_group_collection_entries_id_seq OWNED BY asset_group_collection_entries.id;

CREATE TABLE IF NOT EXISTS asset_group_collections (
    asset_group_id integer,
    id bigint NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone
);

CREATE SEQUENCE IF NOT EXISTS asset_group_collections_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;
ALTER SEQUENCE asset_group_collections_id_seq OWNED BY asset_group_collections.id;

CREATE TABLE IF NOT EXISTS asset_group_selectors (
    asset_group_id integer,
    name text,
    selector text,
    system_selector boolean,
    id integer NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone
);

CREATE SEQUENCE IF NOT EXISTS asset_group_selectors_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;
ALTER SEQUENCE asset_group_selectors_id_seq OWNED BY asset_group_selectors.id;

CREATE TABLE IF NOT EXISTS asset_groups (
    name text,
    tag text,
    system_group boolean,
    id integer NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone
);

CREATE SEQUENCE IF NOT EXISTS asset_groups_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;
ALTER SEQUENCE asset_groups_id_seq OWNED BY asset_groups.id;

CREATE TABLE IF NOT EXISTS audit_logs (
    id bigint NOT NULL,
    created_at timestamp with time zone,
    actor_id text,
    actor_name text,
    action text,
    fields jsonb,
    request_id text
);

CREATE SEQUENCE IF NOT EXISTS audit_logs_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;
ALTER SEQUENCE audit_logs_id_seq OWNED BY audit_logs.id;

CREATE TABLE IF NOT EXISTS auth_secrets (
    user_id text,
    digest text,
    digest_method text,
    expires_at timestamp with time zone,
    totp_secret text,
    totp_activated boolean,
    id integer NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone
);

CREATE SEQUENCE IF NOT EXISTS auth_secrets_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;
ALTER SEQUENCE auth_secrets_id_seq OWNED BY auth_secrets.id;

CREATE TABLE IF NOT EXISTS auth_tokens (
    user_id text,
    client_id text,
    name text,
    key text,
    hmac_method text,
    last_access timestamp with time zone,
    id text NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone
);

CREATE TABLE IF NOT EXISTS azure_data_quality_aggregations (
    tenants bigint,
    users bigint,
    groups bigint,
    apps bigint,
    service_principals bigint,
    devices bigint,
    management_groups bigint,
    subscriptions bigint,
    resource_groups bigint,
    vms bigint,
    key_vaults bigint,
    relationships bigint,
    run_id text,
    id integer NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone
);

CREATE SEQUENCE IF NOT EXISTS azure_data_quality_aggregations_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;
ALTER SEQUENCE azure_data_quality_aggregations_id_seq OWNED BY azure_data_quality_aggregations.id;

CREATE TABLE IF NOT EXISTS azure_data_quality_stats (
    tenant_id text,
    users bigint,
    groups bigint,
    apps bigint,
    service_principals bigint,
    devices bigint,
    management_groups bigint,
    subscriptions bigint,
    resource_groups bigint,
    vms bigint,
    key_vaults bigint,
    relationships bigint,
    run_id text,
    id integer NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone
);

CREATE SEQUENCE IF NOT EXISTS azure_data_quality_stats_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;
ALTER SEQUENCE azure_data_quality_stats_id_seq OWNED BY azure_data_quality_stats.id;

CREATE TABLE IF NOT EXISTS domain_collection_results (
    job_id bigint,
    domain_name text,
    success boolean,
    message text,
    user_count bigint,
    group_count bigint,
    computer_count bigint,
    gpo_count bigint,
    ou_count bigint,
    container_count bigint,
    deleted_count bigint,
    id bigint NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone
);

CREATE SEQUENCE IF NOT EXISTS domain_collection_results_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;
ALTER SEQUENCE domain_collection_results_id_seq OWNED BY domain_collection_results.id;

CREATE TABLE IF NOT EXISTS feature_flags (
    id integer NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    key text,
    name text,
    description text,
    enabled boolean,
    user_updatable boolean
);

CREATE SEQUENCE IF NOT EXISTS feature_flags_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;
ALTER SEQUENCE feature_flags_id_seq OWNED BY feature_flags.id;

CREATE TABLE IF NOT EXISTS file_upload_jobs (
    user_id text,
    user_email_address text,
    status bigint,
    status_message text,
    start_time timestamp with time zone,
    end_time timestamp with time zone,
    last_ingest timestamp with time zone,
    id bigint NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone
);

CREATE SEQUENCE IF NOT EXISTS file_upload_jobs_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;
ALTER SEQUENCE file_upload_jobs_id_seq OWNED BY file_upload_jobs.id;

CREATE TABLE IF NOT EXISTS ingest_tasks (
    file_name text,
    request_guid text,
    task_id bigint,
    id bigint NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone
);

CREATE SEQUENCE IF NOT EXISTS ingest_tasks_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;
ALTER SEQUENCE ingest_tasks_id_seq OWNED BY ingest_tasks.id;

CREATE TABLE IF NOT EXISTS installations (
    id text NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone
);

CREATE TABLE IF NOT EXISTS parameters (
    key text,
    name text,
    description text,
    value jsonb,
    id integer NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone
);

CREATE SEQUENCE IF NOT EXISTS parameters_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;
ALTER SEQUENCE parameters_id_seq OWNED BY parameters.id;

CREATE TABLE IF NOT EXISTS permissions (
    authority text,
    name text,
    id integer NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone
);

CREATE SEQUENCE IF NOT EXISTS permissions_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;
ALTER SEQUENCE permissions_id_seq OWNED BY permissions.id;

CREATE TABLE IF NOT EXISTS roles (
    name text,
    description text,
    id integer NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone
);

CREATE SEQUENCE IF NOT EXISTS roles_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;
ALTER SEQUENCE roles_id_seq OWNED BY roles.id;

CREATE TABLE IF NOT EXISTS roles_permissions (
    role_id integer NOT NULL,
    permission_id integer NOT NULL
);

CREATE TABLE IF NOT EXISTS saml_providers (
    name text,
    display_name text,
    issuer_uri text,
    single_sign_on_uri text,
    metadata_xml bytea,
    ous text[],
    id integer NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone
);

CREATE SEQUENCE IF NOT EXISTS saml_providers_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;
ALTER SEQUENCE saml_providers_id_seq OWNED BY saml_providers.id;

CREATE TABLE IF NOT EXISTS saved_queries (
    user_id text,
    name text,
    query text,
    id bigint NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone
);

CREATE SEQUENCE IF NOT EXISTS saved_queries_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;
ALTER SEQUENCE saved_queries_id_seq OWNED BY saved_queries.id;

CREATE TABLE IF NOT EXISTS user_sessions (
    user_id text,
    auth_provider_type bigint,
    auth_provider_id integer,
    expires_at timestamp with time zone,
    id bigint NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone
);

CREATE SEQUENCE IF NOT EXISTS user_sessions_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;
ALTER SEQUENCE user_sessions_id_seq OWNED BY user_sessions.id;

CREATE TABLE IF NOT EXISTS users (
    saml_provider_id integer,
    first_name text,
    last_name text,
    email_address text,
    principal_name text,
    last_login timestamp with time zone,
    is_disabled boolean,
    eula_accepted boolean,
    id text NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone
);

CREATE TABLE IF NOT EXISTS users_roles (
    user_id text NOT NULL,
    role_id integer NOT NULL
);

ALTER TABLE ONLY ad_data_quality_aggregations ALTER COLUMN id SET DEFAULT nextval('ad_data_quality_aggregations_id_seq'::regclass);
ALTER TABLE ONLY ad_data_quality_stats ALTER COLUMN id SET DEFAULT nextval('ad_data_quality_stats_id_seq'::regclass);
ALTER TABLE ONLY asset_group_collection_entries ALTER COLUMN id SET DEFAULT nextval('asset_group_collection_entries_id_seq'::regclass);
ALTER TABLE ONLY asset_group_collections ALTER COLUMN id SET DEFAULT nextval('asset_group_collections_id_seq'::regclass);
ALTER TABLE ONLY asset_group_selectors ALTER COLUMN id SET DEFAULT nextval('asset_group_selectors_id_seq'::regclass);
ALTER TABLE ONLY asset_groups ALTER COLUMN id SET DEFAULT nextval('asset_groups_id_seq'::regclass);
ALTER TABLE ONLY audit_logs ALTER COLUMN id SET DEFAULT nextval('audit_logs_id_seq'::regclass);
ALTER TABLE ONLY auth_secrets ALTER COLUMN id SET DEFAULT nextval('auth_secrets_id_seq'::regclass);
ALTER TABLE ONLY azure_data_quality_aggregations ALTER COLUMN id SET DEFAULT nextval('azure_data_quality_aggregations_id_seq'::regclass);
ALTER TABLE ONLY azure_data_quality_stats ALTER COLUMN id SET DEFAULT nextval('azure_data_quality_stats_id_seq'::regclass);
ALTER TABLE ONLY domain_collection_results ALTER COLUMN id SET DEFAULT nextval('domain_collection_results_id_seq'::regclass);
ALTER TABLE ONLY feature_flags ALTER COLUMN id SET DEFAULT nextval('feature_flags_id_seq'::regclass);
ALTER TABLE ONLY file_upload_jobs ALTER COLUMN id SET DEFAULT nextval('file_upload_jobs_id_seq'::regclass);
ALTER TABLE ONLY ingest_tasks ALTER COLUMN id SET DEFAULT nextval('ingest_tasks_id_seq'::regclass);
ALTER TABLE ONLY parameters ALTER COLUMN id SET DEFAULT nextval('parameters_id_seq'::regclass);
ALTER TABLE ONLY permissions ALTER COLUMN id SET DEFAULT nextval('permissions_id_seq'::regclass);
ALTER TABLE ONLY roles ALTER COLUMN id SET DEFAULT nextval('roles_id_seq'::regclass);
ALTER TABLE ONLY saml_providers ALTER COLUMN id SET DEFAULT nextval('saml_providers_id_seq'::regclass);
ALTER TABLE ONLY saved_queries ALTER COLUMN id SET DEFAULT nextval('saved_queries_id_seq'::regclass);
ALTER TABLE ONLY user_sessions ALTER COLUMN id SET DEFAULT nextval('user_sessions_id_seq'::regclass);
ALTER TABLE ONLY ad_data_quality_aggregations
    ADD CONSTRAINT ad_data_quality_aggregations_pkey PRIMARY KEY (id);
ALTER TABLE ONLY ad_data_quality_stats
    ADD CONSTRAINT ad_data_quality_stats_pkey PRIMARY KEY (id);
ALTER TABLE ONLY asset_group_collection_entries
    ADD CONSTRAINT asset_group_collection_entries_pkey PRIMARY KEY (id);
ALTER TABLE ONLY asset_group_collections
    ADD CONSTRAINT asset_group_collections_pkey PRIMARY KEY (id);
ALTER TABLE ONLY asset_group_selectors
    ADD CONSTRAINT asset_group_selectors_name_key UNIQUE (name);
ALTER TABLE ONLY asset_group_selectors
    ADD CONSTRAINT asset_group_selectors_pkey PRIMARY KEY (id);
ALTER TABLE ONLY asset_groups
    ADD CONSTRAINT asset_groups_pkey PRIMARY KEY (id);
ALTER TABLE ONLY audit_logs
    ADD CONSTRAINT audit_logs_pkey PRIMARY KEY (id);
ALTER TABLE ONLY auth_secrets
    ADD CONSTRAINT auth_secrets_pkey PRIMARY KEY (id);
ALTER TABLE ONLY auth_tokens
    ADD CONSTRAINT auth_tokens_pkey PRIMARY KEY (id);
ALTER TABLE ONLY azure_data_quality_aggregations
    ADD CONSTRAINT azure_data_quality_aggregations_pkey PRIMARY KEY (id);
ALTER TABLE ONLY azure_data_quality_stats
    ADD CONSTRAINT azure_data_quality_stats_pkey PRIMARY KEY (id);
ALTER TABLE ONLY domain_collection_results
    ADD CONSTRAINT domain_collection_results_pkey PRIMARY KEY (id);
ALTER TABLE ONLY feature_flags
    ADD CONSTRAINT feature_flags_key_key UNIQUE (key);
ALTER TABLE ONLY feature_flags
    ADD CONSTRAINT feature_flags_pkey PRIMARY KEY (id);
ALTER TABLE ONLY file_upload_jobs
    ADD CONSTRAINT file_upload_jobs_pkey PRIMARY KEY (id);
ALTER TABLE ONLY asset_group_selectors
    ADD CONSTRAINT idx_asset_group_selectors_name UNIQUE (name);
ALTER TABLE ONLY ingest_tasks
    ADD CONSTRAINT ingest_tasks_pkey PRIMARY KEY (id);
ALTER TABLE ONLY installations
    ADD CONSTRAINT installations_pkey PRIMARY KEY (id);
ALTER TABLE ONLY parameters
    ADD CONSTRAINT parameters_key_key UNIQUE (key);
ALTER TABLE ONLY parameters
    ADD CONSTRAINT parameters_pkey PRIMARY KEY (id);
ALTER TABLE ONLY permissions
    ADD CONSTRAINT permissions_pkey PRIMARY KEY (id);
ALTER TABLE ONLY roles_permissions
    ADD CONSTRAINT roles_permissions_pkey PRIMARY KEY (role_id, permission_id);
ALTER TABLE ONLY roles
    ADD CONSTRAINT roles_pkey PRIMARY KEY (id);
ALTER TABLE ONLY saml_providers
    ADD CONSTRAINT saml_providers_name_key UNIQUE (name);
ALTER TABLE ONLY saml_providers
    ADD CONSTRAINT saml_providers_pkey PRIMARY KEY (id);
ALTER TABLE ONLY saved_queries
    ADD CONSTRAINT saved_queries_pkey PRIMARY KEY (id);
ALTER TABLE ONLY user_sessions
    ADD CONSTRAINT user_sessions_pkey PRIMARY KEY (id);
ALTER TABLE ONLY users
    ADD CONSTRAINT users_pkey PRIMARY KEY (id);
ALTER TABLE ONLY users
    ADD CONSTRAINT users_principal_name_key UNIQUE (principal_name);
ALTER TABLE ONLY users_roles
    ADD CONSTRAINT users_roles_pkey PRIMARY KEY (user_id, role_id);

CREATE INDEX IF NOT EXISTS idx_ad_data_quality_aggregations_run_id ON ad_data_quality_aggregations USING btree (run_id);
CREATE INDEX IF NOT EXISTS idx_ad_data_quality_stats_run_id ON ad_data_quality_stats USING btree (run_id);
CREATE INDEX IF NOT EXISTS idx_audit_logs_action ON audit_logs USING btree (action);
CREATE INDEX IF NOT EXISTS idx_audit_logs_actor_id ON audit_logs USING btree (actor_id);
CREATE INDEX IF NOT EXISTS idx_audit_logs_created_at ON audit_logs USING btree (created_at);
CREATE INDEX IF NOT EXISTS idx_azure_data_quality_aggregations_run_id ON azure_data_quality_aggregations USING btree (run_id);
CREATE INDEX IF NOT EXISTS idx_azure_data_quality_stats_run_id ON azure_data_quality_stats USING btree (run_id);
CREATE INDEX IF NOT EXISTS idx_saml_providers_name ON saml_providers USING btree (name);
CREATE UNIQUE INDEX IF NOT EXISTS idx_saved_queries_composite_index ON saved_queries USING btree (user_id, name);
CREATE INDEX IF NOT EXISTS idx_users_principal_name ON users USING btree (principal_name);

ALTER TABLE ONLY asset_group_collection_entries
    ADD CONSTRAINT fk_asset_group_collections_entries FOREIGN KEY (asset_group_collection_id) REFERENCES asset_group_collections(id) ON DELETE CASCADE;
ALTER TABLE ONLY asset_group_collections
    ADD CONSTRAINT fk_asset_groups_collections FOREIGN KEY (asset_group_id) REFERENCES asset_groups(id) ON DELETE CASCADE;
ALTER TABLE ONLY asset_group_selectors
    ADD CONSTRAINT fk_asset_groups_selectors FOREIGN KEY (asset_group_id) REFERENCES asset_groups(id) ON DELETE CASCADE;
ALTER TABLE ONLY file_upload_jobs
    ADD CONSTRAINT fk_file_upload_jobs_user FOREIGN KEY (user_id) REFERENCES users(id);
ALTER TABLE ONLY roles_permissions
    ADD CONSTRAINT fk_roles_permissions_permission FOREIGN KEY (permission_id) REFERENCES permissions(id);
ALTER TABLE ONLY roles_permissions
    ADD CONSTRAINT fk_roles_permissions_role FOREIGN KEY (role_id) REFERENCES roles(id);
ALTER TABLE ONLY user_sessions
    ADD CONSTRAINT fk_user_sessions_user FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE;
ALTER TABLE ONLY auth_secrets
    ADD CONSTRAINT fk_users_auth_secret FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE;
ALTER TABLE ONLY auth_tokens
    ADD CONSTRAINT fk_users_auth_tokens FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE;
ALTER TABLE ONLY users_roles
    ADD CONSTRAINT fk_users_roles_role FOREIGN KEY (role_id) REFERENCES roles(id);
ALTER TABLE ONLY users_roles
    ADD CONSTRAINT fk_users_roles_user FOREIGN KEY (user_id) REFERENCES users(id);
ALTER TABLE ONLY users
    ADD CONSTRAINT fk_users_saml_provider FOREIGN KEY (saml_provider_id) REFERENCES saml_providers(id);

-- Populate asset group table
INSERT INTO asset_groups (name, tag, system_group, created_at, updated_at) VALUES ('Admin Tier Zero', 'admin_tier_0', true, current_timestamp, current_timestamp);

-- Populate permissions table
INSERT INTO permissions (authority, name, created_at, updated_at) VALUES ('app', 'ReadAppConfig', current_timestamp, current_timestamp);
INSERT INTO permissions (authority, name, created_at, updated_at) VALUES ('app', 'WriteAppConfig', current_timestamp, current_timestamp);
INSERT INTO permissions (authority, name, created_at, updated_at) VALUES ('risks', 'GenerateReport', current_timestamp, current_timestamp);
INSERT INTO permissions (authority, name, created_at, updated_at) VALUES ('risks', 'ManageRisks', current_timestamp, current_timestamp);
INSERT INTO permissions (authority, name, created_at, updated_at) VALUES ('auth', 'CreateToken', current_timestamp, current_timestamp);
INSERT INTO permissions (authority, name, created_at, updated_at) VALUES ('auth', 'ManageAppConfig', current_timestamp, current_timestamp);
INSERT INTO permissions (authority, name, created_at, updated_at) VALUES ('auth', 'ManageProviders', current_timestamp, current_timestamp);
INSERT INTO permissions (authority, name, created_at, updated_at) VALUES ('auth', 'ManageSelf', current_timestamp, current_timestamp);
INSERT INTO permissions (authority, name, created_at, updated_at) VALUES ('auth', 'ManageUsers', current_timestamp, current_timestamp);
INSERT INTO permissions (authority, name, created_at, updated_at) VALUES ('clients', 'Manage', current_timestamp, current_timestamp);
INSERT INTO permissions (authority, name, created_at, updated_at) VALUES ('clients', 'Tasking', current_timestamp, current_timestamp);
INSERT INTO permissions (authority, name, created_at, updated_at) VALUES ('collection', 'ManageJobs', current_timestamp, current_timestamp);
INSERT INTO permissions (authority, name, created_at, updated_at) VALUES ('graphdb', 'Read', current_timestamp, current_timestamp);
INSERT INTO permissions (authority, name, created_at, updated_at) VALUES ('graphdb', 'Write', current_timestamp, current_timestamp);

-- Populate roles table
INSERT INTO roles (name, description, created_at, updated_at) VALUES ('Administrator', 'Can manage users, clients, and application configuration', current_timestamp, current_timestamp);
INSERT INTO roles (name, description, created_at, updated_at) VALUES ('User', 'Can read data, modify asset group memberships', current_timestamp, current_timestamp);
INSERT INTO roles (name, description, created_at, updated_at) VALUES ('Read-Only', 'Used for integrations', current_timestamp, current_timestamp);
INSERT INTO roles (name, description, created_at, updated_at) VALUES ('Upload-Only', 'Used for data collection clients, can post data but cannot read data', current_timestamp, current_timestamp);


-- Populate roles_permissions table
-- Administrator
INSERT INTO roles_permissions (role_id, permission_id) VALUES ((SELECT id FROM roles WHERE roles.name  = 'Administrator'), (SELECT id FROM permissions WHERE permissions.authority  = 'app'  and permissions.name = 'ReadAppConfig'));
INSERT INTO roles_permissions (role_id, permission_id) VALUES ((SELECT id FROM roles WHERE roles.name  = 'Administrator'), (SELECT id FROM permissions WHERE permissions.authority  = 'app'  and permissions.name = 'WriteAppConfig'));
INSERT INTO roles_permissions (role_id, permission_id) VALUES ((SELECT id FROM roles WHERE roles.name  = 'Administrator'), (SELECT id FROM permissions WHERE permissions.authority  = 'risks'  and permissions.name = 'GenerateReport'));
INSERT INTO roles_permissions (role_id, permission_id) VALUES ((SELECT id FROM roles WHERE roles.name  = 'Administrator'), (SELECT id FROM permissions WHERE permissions.authority  = 'risks'  and permissions.name = 'ManageRisks'));
INSERT INTO roles_permissions (role_id, permission_id) VALUES ((SELECT id FROM roles WHERE roles.name  = 'Administrator'), (SELECT id FROM permissions WHERE permissions.authority  = 'auth'  and permissions.name = 'CreateToken'));
INSERT INTO roles_permissions (role_id, permission_id) VALUES ((SELECT id FROM roles WHERE roles.name  = 'Administrator'), (SELECT id FROM permissions WHERE permissions.authority  = 'auth'  and permissions.name = 'ManageAppConfig'));
INSERT INTO roles_permissions (role_id, permission_id) VALUES ((SELECT id FROM roles WHERE roles.name  = 'Administrator'), (SELECT id FROM permissions WHERE permissions.authority  = 'auth'  and permissions.name = 'ManageProviders'));
INSERT INTO roles_permissions (role_id, permission_id) VALUES ((SELECT id FROM roles WHERE roles.name  = 'Administrator'), (SELECT id FROM permissions WHERE permissions.authority  = 'auth'  and permissions.name = 'ManageSelf'));
INSERT INTO roles_permissions (role_id, permission_id) VALUES ((SELECT id FROM roles WHERE roles.name  = 'Administrator'), (SELECT id FROM permissions WHERE permissions.authority  = 'auth'  and permissions.name = 'ManageUsers'));
INSERT INTO roles_permissions (role_id, permission_id) VALUES ((SELECT id FROM roles WHERE roles.name  = 'Administrator'), (SELECT id FROM permissions WHERE permissions.authority  = 'clients'  and permissions.name = 'Manage'));
INSERT INTO roles_permissions (role_id, permission_id) VALUES ((SELECT id FROM roles WHERE roles.name  = 'Administrator'), (SELECT id FROM permissions WHERE permissions.authority  = 'clients'  and permissions.name = 'Tasking'));
INSERT INTO roles_permissions (role_id, permission_id) VALUES ((SELECT id FROM roles WHERE roles.name  = 'Administrator'), (SELECT id FROM permissions WHERE permissions.authority  = 'collection'  and permissions.name = 'ManageJobs'));
INSERT INTO roles_permissions (role_id, permission_id) VALUES ((SELECT id FROM roles WHERE roles.name  = 'Administrator'), (SELECT id FROM permissions WHERE permissions.authority  = 'graphdb'  and permissions.name = 'Read'));
INSERT INTO roles_permissions (role_id, permission_id) VALUES ((SELECT id FROM roles WHERE roles.name  = 'Administrator'), (SELECT id FROM permissions WHERE permissions.authority  = 'graphdb'  and permissions.name = 'Write'));

-- User
INSERT INTO roles_permissions (role_id, permission_id) VALUES ((SELECT id FROM roles WHERE roles.name  = 'User'), (SELECT id FROM permissions WHERE permissions.authority  = 'app'  and permissions.name = 'ReadAppConfig'));
INSERT INTO roles_permissions (role_id, permission_id) VALUES ((SELECT id FROM roles WHERE roles.name  = 'User'), (SELECT id FROM permissions WHERE permissions.authority  = 'risks'  and permissions.name = 'GenerateReport'));
INSERT INTO roles_permissions (role_id, permission_id) VALUES ((SELECT id FROM roles WHERE roles.name  = 'User'), (SELECT id FROM permissions WHERE permissions.authority  = 'auth'  and permissions.name = 'CreateToken'));
INSERT INTO roles_permissions (role_id, permission_id) VALUES ((SELECT id FROM roles WHERE roles.name  = 'User'), (SELECT id FROM permissions WHERE permissions.authority  = 'auth'  and permissions.name = 'ManageSelf'));
INSERT INTO roles_permissions (role_id, permission_id) VALUES ((SELECT id FROM roles WHERE roles.name  = 'User'), (SELECT id FROM permissions WHERE permissions.authority  = 'clients'  and permissions.name = 'Manage'));
INSERT INTO roles_permissions (role_id, permission_id) VALUES ((SELECT id FROM roles WHERE roles.name  = 'User'), (SELECT id FROM permissions WHERE permissions.authority  = 'graphdb'  and permissions.name = 'Read'));

-- Read-Only
INSERT INTO roles_permissions (role_id, permission_id) VALUES ((SELECT id FROM roles WHERE roles.name  = 'Read-Only'), (SELECT id FROM permissions WHERE permissions.authority  = 'app'  and permissions.name = 'ReadAppConfig'));
INSERT INTO roles_permissions (role_id, permission_id) VALUES ((SELECT id FROM roles WHERE roles.name  = 'Read-Only'), (SELECT id FROM permissions WHERE permissions.authority  = 'risks'  and permissions.name = 'GenerateReport'));
INSERT INTO roles_permissions (role_id, permission_id) VALUES ((SELECT id FROM roles WHERE roles.name  = 'Read-Only'), (SELECT id FROM permissions WHERE permissions.authority  = 'auth'  and permissions.name = 'ManageSelf'));
INSERT INTO roles_permissions (role_id, permission_id) VALUES ((SELECT id FROM roles WHERE roles.name  = 'Read-Only'), (SELECT id FROM permissions WHERE permissions.authority  = 'graphdb'  and permissions.name = 'Read'));

-- Upload-Only
INSERT INTO roles_permissions (role_id, permission_id) VALUES ((SELECT id FROM roles WHERE roles.name  = 'Upload-Only'), (SELECT id FROM permissions WHERE permissions.authority  = 'clients'  and permissions.name = 'Tasking'));
INSERT INTO roles_permissions (role_id, permission_id) VALUES ((SELECT id FROM roles WHERE roles.name  = 'Upload-Only'), (SELECT id FROM permissions WHERE permissions.authority  = 'graphdb'  and permissions.name = 'Write'));

-- Populate feature_flags table
INSERT INTO feature_flags (created_at, updated_at, key, name, description, enabled, user_updatable) VALUES (current_timestamp, current_timestamp, 'butterfly_analysis', 'Enhanced Asset Inbound-Outbound Exposure Analysis', 'Enables more extensive analysis of attack path findings that allows BloodHound to help the user prioritize remediation of the most exposed assets.', false, false);
INSERT INTO feature_flags (created_at, updated_at, key, name, description, enabled, user_updatable) VALUES (current_timestamp, current_timestamp, 'enable_saml_sso', 'SAML Single Sign-On Support', 'Enables SSO authentication flows and administration panels to third party SAML identity providers.', true, false);
INSERT INTO feature_flags (created_at, updated_at, key, name, description, enabled, user_updatable) VALUES (current_timestamp, current_timestamp, 'scope_collection_by_ou', 'Enable SharpHound OU Scoped Collections', 'Enables scoping SharpHound collections to specific lists of OUs.', true, false);
INSERT INTO feature_flags (created_at, updated_at, key, name, description, enabled, user_updatable) VALUES (current_timestamp, current_timestamp, 'azure_support', 'Enable Azure Support', 'Enables Azure support.', true, false);
INSERT INTO feature_flags (created_at, updated_at, key, name, description, enabled, user_updatable) VALUES (current_timestamp, current_timestamp, 'reconciliation', 'Reconciliation', 'Enables Reconciliation', true, false);
INSERT INTO feature_flags (created_at, updated_at, key, name, description, enabled, user_updatable) VALUES (current_timestamp, current_timestamp, 'entity_panel_cache', 'Enable application level caching', 'Enables the use of application level caching for entity panel queries', true, false);
INSERT INTO feature_flags (created_at, updated_at, key, name, description, enabled, user_updatable) VALUES (current_timestamp, current_timestamp, 'adcs', 'Enable collection and processing of Active Directory Certificate Services Data', 'Enables the ability to collect, analyze, and explore Active Directory Certificate Services data and previews new attack paths.', false, false );


-- Populate parameters table
INSERT INTO parameters (key, name, description, value, id, created_at, updated_at) VALUES ('auth.password_expiration_window', 'Local Auth Password Expiry Window', 'This configuration parameter sets the local auth password expiry window for users that have valid auth secrets. Values for this configuration must follow the duration specification of ISO-8601.', '{"duration": "P90D"}', 1, current_timestamp, current_timestamp);
INSERT INTO parameters (key, name, description, value, id, created_at, updated_at) VALUES ('neo4j.configuration', 'Neo4j Configuration Parameters', 'This configuration parameter sets the BatchWriteSize and the BatchFlushSize for Neo4J.', '{"batch_write_size": 20000, "write_flush_size": 100000}', 2, current_timestamp, current_timestamp);
