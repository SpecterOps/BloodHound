-- +goose Up

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
CREATE EXTENSION IF NOT EXISTS pg_trgm;

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
    updated_at timestamp with time zone,
    aiacas bigint DEFAULT 0,
    rootcas bigint DEFAULT 0,
    enterprisecas bigint DEFAULT 0,
    ntauthstores bigint DEFAULT 0,
    certtemplates bigint DEFAULT 0,
    issuancepolicies bigint DEFAULT 0
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
    updated_at timestamp with time zone,
    aiacas bigint DEFAULT 0,
    rootcas bigint DEFAULT 0,
    enterprisecas bigint DEFAULT 0,
    ntauthstores bigint DEFAULT 0,
    certtemplates bigint DEFAULT 0,
    issuancepolicies bigint DEFAULT 0
);
CREATE SEQUENCE IF NOT EXISTS ad_data_quality_stats_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;
ALTER SEQUENCE ad_data_quality_stats_id_seq OWNED BY ad_data_quality_stats.id;

CREATE TABLE IF NOT EXISTS analysis_request_switch (
    singleton boolean DEFAULT true NOT NULL,
    request_type text NOT NULL,
    requested_by text NOT NULL,
    requested_at timestamp with time zone NOT NULL,
    CONSTRAINT singleton_uni CHECK (singleton)
);
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
    name text NOT NULL,
    tag text NOT NULL,
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
    request_id text,
    actor_email character varying(330) DEFAULT NULL::character varying,
    source_ip_address text DEFAULT NULL::character varying,
    status character varying(15) DEFAULT 'intent'::character varying,
    commit_id text,
    CONSTRAINT status_check CHECK (status IN ('intent', 'success', 'failure'))
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
    updated_at timestamp with time zone,
    automation_accounts bigint DEFAULT 0,
    container_registries bigint DEFAULT 0,
    function_apps bigint DEFAULT 0,
    logic_apps bigint DEFAULT 0,
    managed_clusters bigint DEFAULT 0,
    vm_scale_sets bigint DEFAULT 0,
    web_apps bigint DEFAULT 0
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
    updated_at timestamp with time zone,
    automation_accounts bigint DEFAULT 0,
    container_registries bigint DEFAULT 0,
    function_apps bigint DEFAULT 0,
    logic_apps bigint DEFAULT 0,
    managed_clusters bigint DEFAULT 0,
    vm_scale_sets bigint DEFAULT 0,
    web_apps bigint DEFAULT 0
);
CREATE SEQUENCE IF NOT EXISTS azure_data_quality_stats_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;
ALTER SEQUENCE azure_data_quality_stats_id_seq OWNED BY azure_data_quality_stats.id;

CREATE TABLE IF NOT EXISTS datapipe_status (
    singleton boolean DEFAULT true NOT NULL,
    status text NOT NULL,
    updated_at timestamp with time zone NOT NULL,
    last_complete_analysis_at timestamp with time zone,
    CONSTRAINT singleton_uni CHECK (singleton)
);

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
    updated_at timestamp with time zone,
    total_files integer DEFAULT 0,
    failed_files integer DEFAULT 0
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
    updated_at timestamp with time zone,
    file_type integer DEFAULT 0
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
    updated_at timestamp with time zone,
    description text DEFAULT ''::text
);
CREATE SEQUENCE IF NOT EXISTS saved_queries_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;
ALTER SEQUENCE saved_queries_id_seq OWNED BY saved_queries.id;

CREATE TABLE IF NOT EXISTS saved_queries_permissions (
    id bigint NOT NULL,
    shared_to_user_id text,
    query_id bigint NOT NULL,
    public boolean DEFAULT false NOT NULL,
    created_at timestamp with time zone DEFAULT now(),
    updated_at timestamp with time zone DEFAULT now()
);
CREATE SEQUENCE IF NOT EXISTS saved_queries_permissions_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;
ALTER SEQUENCE saved_queries_permissions_id_seq OWNED BY saved_queries_permissions.id;

CREATE TABLE IF NOT EXISTS user_sessions (
    user_id text,
    auth_provider_type bigint,
    auth_provider_id integer,
    expires_at timestamp with time zone,
    id bigint NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    flags jsonb
);

---
--- NOTE: This is bad, and will be removed in a future migration. We should
---       consider using a foriegn key for query_id as well.
---
CREATE SEQUENCE saved_queries_permissions_query_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;
ALTER SEQUENCE saved_queries_permissions_query_id_seq OWNED BY saved_queries_permissions.query_id;

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
ALTER TABLE ONLY feature_flags ALTER COLUMN id SET DEFAULT nextval('feature_flags_id_seq'::regclass);
ALTER TABLE ONLY file_upload_jobs ALTER COLUMN id SET DEFAULT nextval('file_upload_jobs_id_seq'::regclass);
ALTER TABLE ONLY ingest_tasks ALTER COLUMN id SET DEFAULT nextval('ingest_tasks_id_seq'::regclass);
ALTER TABLE ONLY parameters ALTER COLUMN id SET DEFAULT nextval('parameters_id_seq'::regclass);
ALTER TABLE ONLY permissions ALTER COLUMN id SET DEFAULT nextval('permissions_id_seq'::regclass);
ALTER TABLE ONLY roles ALTER COLUMN id SET DEFAULT nextval('roles_id_seq'::regclass);
ALTER TABLE ONLY saml_providers ALTER COLUMN id SET DEFAULT nextval('saml_providers_id_seq'::regclass);
ALTER TABLE ONLY saved_queries ALTER COLUMN id SET DEFAULT nextval('saved_queries_id_seq'::regclass);
ALTER TABLE ONLY saved_queries_permissions ALTER COLUMN id SET DEFAULT nextval('saved_queries_permissions_id_seq'::regclass);
ALTER TABLE ONLY saved_queries_permissions ALTER COLUMN query_id SET DEFAULT nextval('saved_queries_permissions_query_id_seq'::regclass);
ALTER TABLE ONLY user_sessions ALTER COLUMN id SET DEFAULT nextval('user_sessions_id_seq'::regclass);

ALTER TABLE ONLY ad_data_quality_aggregations ADD CONSTRAINT ad_data_quality_aggregations_pkey PRIMARY KEY (id);
ALTER TABLE ONLY ad_data_quality_stats ADD CONSTRAINT ad_data_quality_stats_pkey PRIMARY KEY (id);
ALTER TABLE ONLY analysis_request_switch ADD CONSTRAINT analysis_request_switch_pkey PRIMARY KEY (singleton);
ALTER TABLE ONLY asset_group_collection_entries ADD CONSTRAINT asset_group_collection_entries_pkey PRIMARY KEY (id);
ALTER TABLE ONLY asset_group_collections ADD CONSTRAINT asset_group_collections_pkey PRIMARY KEY (id);
ALTER TABLE ONLY asset_group_selectors ADD CONSTRAINT asset_group_selectors_name_assetgroupid_key UNIQUE (name, asset_group_id);
ALTER TABLE ONLY asset_group_selectors ADD CONSTRAINT asset_group_selectors_pkey PRIMARY KEY (id);
ALTER TABLE ONLY asset_groups ADD CONSTRAINT asset_groups_name_key UNIQUE (name);
ALTER TABLE ONLY asset_groups ADD CONSTRAINT asset_groups_pkey PRIMARY KEY (id);
ALTER TABLE ONLY asset_groups ADD CONSTRAINT asset_groups_tag_key UNIQUE (tag);
ALTER TABLE ONLY audit_logs ADD CONSTRAINT audit_logs_pkey PRIMARY KEY (id);
ALTER TABLE ONLY auth_secrets ADD CONSTRAINT auth_secrets_pkey PRIMARY KEY (id);
ALTER TABLE ONLY auth_tokens ADD CONSTRAINT auth_tokens_pkey PRIMARY KEY (id);
ALTER TABLE ONLY azure_data_quality_aggregations ADD CONSTRAINT azure_data_quality_aggregations_pkey PRIMARY KEY (id);
ALTER TABLE ONLY azure_data_quality_stats ADD CONSTRAINT azure_data_quality_stats_pkey PRIMARY KEY (id);
ALTER TABLE ONLY datapipe_status ADD CONSTRAINT datapipe_status_pkey PRIMARY KEY (singleton);
ALTER TABLE ONLY feature_flags ADD CONSTRAINT feature_flags_key_key UNIQUE (key);
ALTER TABLE ONLY feature_flags ADD CONSTRAINT feature_flags_pkey PRIMARY KEY (id);
ALTER TABLE ONLY file_upload_jobs ADD CONSTRAINT file_upload_jobs_pkey PRIMARY KEY (id);
ALTER TABLE ONLY ingest_tasks ADD CONSTRAINT ingest_tasks_pkey PRIMARY KEY (id);
ALTER TABLE ONLY installations ADD CONSTRAINT installations_pkey PRIMARY KEY (id);
ALTER TABLE ONLY parameters ADD CONSTRAINT parameters_key_key UNIQUE (key);
ALTER TABLE ONLY parameters ADD CONSTRAINT parameters_pkey PRIMARY KEY (id);
ALTER TABLE ONLY permissions ADD CONSTRAINT permissions_authority_name_key UNIQUE (authority, name);
ALTER TABLE ONLY permissions ADD CONSTRAINT permissions_pkey PRIMARY KEY (id);
ALTER TABLE ONLY roles ADD CONSTRAINT roles_name_key UNIQUE (name);
ALTER TABLE ONLY roles_permissions ADD CONSTRAINT roles_permissions_pkey PRIMARY KEY (role_id, permission_id);
ALTER TABLE ONLY roles ADD CONSTRAINT roles_pkey PRIMARY KEY (id);
ALTER TABLE ONLY saml_providers ADD CONSTRAINT saml_providers_name_key UNIQUE (name);
ALTER TABLE ONLY saml_providers ADD CONSTRAINT saml_providers_pkey PRIMARY KEY (id);
ALTER TABLE ONLY saved_queries_permissions ADD CONSTRAINT saved_queries_permissions_pkey PRIMARY KEY (id);
ALTER TABLE ONLY saved_queries_permissions ADD CONSTRAINT saved_queries_permissions_shared_to_user_id_query_id_key UNIQUE (shared_to_user_id, query_id);
ALTER TABLE ONLY saved_queries ADD CONSTRAINT saved_queries_pkey PRIMARY KEY (id);
ALTER TABLE ONLY user_sessions ADD CONSTRAINT user_sessions_pkey PRIMARY KEY (id);
ALTER TABLE ONLY users ADD CONSTRAINT users_pkey PRIMARY KEY (id);
ALTER TABLE ONLY users ADD CONSTRAINT users_principal_name_key UNIQUE (principal_name);
ALTER TABLE ONLY users_roles ADD CONSTRAINT users_roles_pkey PRIMARY KEY (user_id, role_id);

CREATE INDEX idx_ad_asset_groups_created_at ON asset_groups USING btree (created_at);
CREATE INDEX idx_ad_asset_groups_updated_at ON asset_groups USING btree (updated_at);
CREATE INDEX idx_ad_data_quality_aggregations_created_at ON ad_data_quality_aggregations USING btree (created_at);
CREATE INDEX idx_ad_data_quality_aggregations_run_id ON ad_data_quality_aggregations USING btree (run_id);
CREATE INDEX idx_ad_data_quality_aggregations_updated_at ON ad_data_quality_aggregations USING btree (updated_at);
CREATE INDEX idx_ad_data_quality_stats_run_id ON ad_data_quality_stats USING btree (run_id);
CREATE INDEX idx_asset_group_collection_entries_asset_group_collection_id ON asset_group_collection_entries USING btree (asset_group_collection_id);
CREATE INDEX idx_asset_group_collection_entries_created_at ON asset_group_collection_entries USING btree (created_at);
CREATE INDEX idx_asset_group_collection_entries_updated_at ON asset_group_collection_entries USING btree (updated_at);
CREATE INDEX idx_asset_group_collections_asset_group_id ON asset_group_collections USING btree (asset_group_id);
CREATE INDEX idx_asset_group_collections_created_at ON asset_group_collections USING btree (created_at);
CREATE INDEX idx_asset_group_collections_updated_at ON asset_group_collections USING btree (updated_at);
CREATE INDEX idx_audit_logs_action ON audit_logs USING btree (action);
CREATE INDEX idx_audit_logs_actor_email ON audit_logs USING btree (actor_email);
CREATE INDEX idx_audit_logs_actor_id ON audit_logs USING btree (actor_id);
CREATE INDEX idx_audit_logs_created_at ON audit_logs USING btree (created_at);
CREATE INDEX idx_audit_logs_source_ip_address ON audit_logs USING btree (source_ip_address);
CREATE INDEX idx_audit_logs_status ON audit_logs USING btree (status);
CREATE INDEX idx_azure_data_quality_aggregations_created_at ON azure_data_quality_aggregations USING btree (created_at);
CREATE INDEX idx_azure_data_quality_aggregations_run_id ON azure_data_quality_aggregations USING btree (run_id);
CREATE INDEX idx_azure_data_quality_stats_created_at ON azure_data_quality_stats USING btree (created_at);
CREATE INDEX idx_azure_data_quality_stats_run_id ON azure_data_quality_stats USING btree (run_id);
CREATE INDEX idx_azure_data_quality_stats_updated_at ON azure_data_quality_stats USING btree (updated_at);
CREATE INDEX idx_file_upload_jobs_created_at ON file_upload_jobs USING btree (created_at);
CREATE INDEX idx_file_upload_jobs_end_time ON file_upload_jobs USING btree (end_time);
CREATE INDEX idx_file_upload_jobs_start_time ON file_upload_jobs USING btree (start_time);
CREATE INDEX idx_file_upload_jobs_status ON file_upload_jobs USING btree (status);
CREATE INDEX idx_file_upload_jobs_updated_at ON file_upload_jobs USING btree (updated_at);
CREATE INDEX idx_ingest_tasks_task_id ON ingest_tasks USING btree (task_id);
CREATE INDEX idx_saml_providers_name ON saml_providers USING btree (name);
CREATE UNIQUE INDEX idx_saved_queries_composite_index ON saved_queries USING btree (user_id, name);
CREATE INDEX idx_saved_queries_description ON saved_queries USING gin (description gin_trgm_ops);
CREATE INDEX idx_saved_queries_name ON saved_queries USING gin (name gin_trgm_ops);
CREATE INDEX idx_users_eula_accepted ON users USING btree (eula_accepted);
CREATE INDEX idx_users_principal_name ON users USING btree (principal_name);

ALTER TABLE ONLY asset_group_collection_entries ADD CONSTRAINT fk_asset_group_collections_entries FOREIGN KEY (asset_group_collection_id) REFERENCES asset_group_collections(id) ON DELETE CASCADE;
ALTER TABLE ONLY asset_group_collections ADD CONSTRAINT fk_asset_groups_collections FOREIGN KEY (asset_group_id) REFERENCES asset_groups(id) ON DELETE CASCADE;
ALTER TABLE ONLY asset_group_selectors ADD CONSTRAINT fk_asset_groups_selectors FOREIGN KEY (asset_group_id) REFERENCES asset_groups(id) ON DELETE CASCADE;
ALTER TABLE ONLY file_upload_jobs ADD CONSTRAINT fk_file_upload_jobs_user FOREIGN KEY (user_id) REFERENCES users(id);
ALTER TABLE ONLY roles_permissions ADD CONSTRAINT fk_roles_permissions_permission FOREIGN KEY (permission_id) REFERENCES permissions(id);
ALTER TABLE ONLY roles_permissions ADD CONSTRAINT fk_roles_permissions_role FOREIGN KEY (role_id) REFERENCES roles(id);
ALTER TABLE ONLY user_sessions ADD CONSTRAINT fk_user_sessions_user FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE;
ALTER TABLE ONLY auth_secrets ADD CONSTRAINT fk_users_auth_secret FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE;
ALTER TABLE ONLY auth_tokens ADD CONSTRAINT fk_users_auth_tokens FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE;
ALTER TABLE ONLY users_roles ADD CONSTRAINT fk_users_roles_role FOREIGN KEY (role_id) REFERENCES roles(id);
ALTER TABLE ONLY users_roles ADD CONSTRAINT fk_users_roles_user FOREIGN KEY (user_id) REFERENCES users(id);
ALTER TABLE ONLY users ADD CONSTRAINT fk_users_saml_provider FOREIGN KEY (saml_provider_id) REFERENCES saml_providers(id);
ALTER TABLE ONLY saved_queries_permissions ADD CONSTRAINT saved_queries_permissions_query_id_fkey FOREIGN KEY (query_id) REFERENCES saved_queries(id) ON DELETE CASCADE;
ALTER TABLE ONLY saved_queries_permissions ADD CONSTRAINT saved_queries_permissions_shared_to_user_id_fkey FOREIGN KEY (shared_to_user_id) REFERENCES users(id) ON DELETE CASCADE;

--
-- Application startup requires some basic data to start up. One day, we might want to explore moving this to a config file
--

INSERT INTO asset_groups (name, tag, system_group, created_at, updated_at) VALUES
    ('Admin Tier Zero', 'admin_tier_0', true, current_timestamp, current_timestamp),
    ('Owned', 'owned', true, current_timestamp, current_timestamp);

INSERT INTO datapipe_status VALUES (true, 'idle', current_timestamp, NULL);

INSERT INTO feature_flags (key, name, description, enabled, user_updatable, created_at, updated_at) VALUES
(
    'butterfly_analysis', 'Enhanced Asset Inbound-Outbound Exposure Analysis',
    'Enables more extensive analysis of attack path findings that allows BloodHound to help the user prioritize remediation of the most exposed assets.',
    false, false, current_timestamp, current_timestamp
), (
    'enable_saml_sso', 'SAML Single Sign-On Support',
    'Enables SSO authentication flows and administration panels to third party SAML identity providers.',
    true, false, current_timestamp, current_timestamp
), (
    'scope_collection_by_ou',
    'Enable SharpHound OU Scoped Collections',
    'Enables scoping SharpHound collections to specific lists of OUs.',
    true, false, current_timestamp, current_timestamp
), (
    'azure_support',
    'Enable Azure Support',
    'Enables Azure support.',
    true, false, current_timestamp, current_timestamp
), (
    'entity_panel_cache',
    'Enable application level caching',
    'Enables the use of application level caching for entity panel queries',
    true, false, current_timestamp, current_timestamp
), (
    'adcs',
    'Enable collection and processing of Active Directory Certificate Services Data',
    'Enables the ability to collect, analyze, and explore Active Directory Certificate Services data and previews new attack paths.',
    false, false, current_timestamp, current_timestamp
), (
    'dark_mode',
    'Dark Mode',
    'Allows users to enable or disable dark mode via a toggle in the settings menu',
    true, false, current_timestamp, current_timestamp
), (
    'pg_migration_dual_ingest',
    'PostgreSQL Migration Dual Ingest',
    'Enables dual ingest pathing for both Neo4j and PostgreSQL.',
    false, false, current_timestamp, current_timestamp
), (
    'clear_graph_data',
    'Clear Graph Data',
    'Enables the ability to delete all nodes and edges from the graph database.',
    true, false, current_timestamp, current_timestamp
), (
    'risk_exposure_new_calculation',
    'Use new tier zero risk exposure calculation',
    'Enables the use of new tier zero risk exposure metatree metrics.',
    false, false, current_timestamp, current_timestamp
), (
    'fedramp_eula',
    'FedRAMP EULA',
    'Enables showing the FedRAMP EULA on every login. (Enterprise only)',
    false, false, current_timestamp, current_timestamp
), (
    'auto_tag_t0_parent_objects',
    'Automatically add parent OUs and containers of Tier Zero AD objects to Tier Zero',
    'Parent OUs and containers of Tier Zero AD objects are automatically added to Tier Zero during analysis. Containers are only added if they have a Tier Zero child object with ACL inheritance enabled.',
    true, true, current_timestamp, current_timestamp
), (
    'oidc_support',
    'OIDC Support',
    'Enables OpenID Connect authentication support for SSO Authentication.',
    false, false, current_timestamp, current_timestamp
);


INSERT INTO parameters (key, name, description, value, created_at, updated_at) VALUES (
    'auth.password_expiration_window',
    'Local Auth Password Expiry Window',
    'This configuration parameter sets the local auth password expiry window for users that have valid auth secrets. Values for this configuration must follow the duration specification of ISO-8601.',
    '{"duration": "P90D"}',
    current_timestamp, current_timestamp
), (
    'neo4j.configuration',
    'Neo4j Configuration Parameters',
    'This configuration parameter sets the BatchWriteSize and the BatchFlushSize for Neo4J.',
    '{"batch_write_size": 20000, "write_flush_size": 100000}',
    current_timestamp, current_timestamp
), (
    'analysis.citrix_rdp_support',
    'Citrix RDP Support',
    'This configuration parameter toggles Citrix support during post-processing. When enabled, computers identified with a ''Direct Access Users'' local group will assume that Citrix is installed and CanRDP edges will require membership of both ''Direct Access Users'' and ''Remote Desktop Users'' local groups on the computer.', '{ "enabled": false }',
    current_timestamp, current_timestamp
), (
    'prune.ttl',
    'Prune Retention TTL Configuration Parameters',
    'This configuration parameter sets the retention TTLs during analysis pruning.', '{"base_ttl": "P7D", "has_session_edge_ttl": "P3D" }',
    current_timestamp, current_timestamp
), (
    'analysis.reconciliation',
    'Reconciliation',
    'This configuration parameter enables / disables reconciliation during analysis.', '{"enabled": true}',
    current_timestamp, current_timestamp
);

INSERT INTO permissions (authority, name, created_at, updated_at) VALUES
('app', 'ReadAppConfig', current_timestamp, current_timestamp),
('app', 'WriteAppConfig', current_timestamp, current_timestamp),
('risks', 'GenerateReport', current_timestamp, current_timestamp),
 ('risks', 'ManageRisks', current_timestamp, current_timestamp),
 ('auth', 'CreateToken', current_timestamp, current_timestamp),
 ('auth', 'ManageProviders', current_timestamp, current_timestamp),
 ('auth', 'ManageSelf', current_timestamp, current_timestamp),
 ('auth', 'ManageUsers', current_timestamp, current_timestamp),
 ('clients', 'Manage', current_timestamp, current_timestamp),
 ('clients', 'Tasking', current_timestamp, current_timestamp),
 ('collection', 'ManageJobs', current_timestamp, current_timestamp),
 ('graphdb', 'Read', current_timestamp, current_timestamp),
 ('graphdb', 'Write', current_timestamp, current_timestamp),
 ('saved_queries', 'Read', current_timestamp, current_timestamp),
 ('saved_queries', 'Write', current_timestamp, current_timestamp),
 ('clients', 'Read', current_timestamp, current_timestamp),
 ('db', 'Wipe', current_timestamp, current_timestamp),
 ('graphdb', 'Mutate', current_timestamp, current_timestamp);

INSERT INTO roles (name, description, created_at, updated_at) VALUES
 ('Administrator', 'Can manage users, clients, and application configuration', current_timestamp, current_timestamp),
 ('User', 'Can read data, modify asset group memberships', current_timestamp, current_timestamp),
 ('Read-Only', 'Used for integrations', current_timestamp, current_timestamp),
 ('Upload-Only', 'Used for data collection clients, can post data but cannot read data', current_timestamp, current_timestamp),
 ('Power User', 'Can upload data, manage clients, and perform any action a User can', current_timestamp, current_timestamp);

INSERT INTO roles_permissions (role_id, permission_id)
SELECT r.id, p.id
FROM roles r
JOIN permissions p
  ON (
    (r.name = 'Administrator' AND (p.authority, p.name) IN (
        ('app', 'ReadAppConfig'),
        ('app', 'WriteAppConfig'),
        ('risks', 'GenerateReport'),
        ('risks', 'ManageRisks'),
        ('auth', 'CreateToken'),
        ('auth', 'ManageProviders'),
        ('auth', 'ManageSelf'),
        ('auth', 'ManageUsers'),
        ('clients', 'Manage'),
        ('clients', 'Tasking'),
        ('collection', 'ManageJobs'),
        ('graphdb', 'Read'),
        ('graphdb', 'Write'),
        ('saved_queries', 'Read'),
        ('saved_queries', 'Write'),
        ('clients', 'Read'),
        ('graphdb', 'Mutate'),
        ('db', 'Wipe')
    ))
    OR
    (r.name = 'User' AND (p.authority, p.name) IN (
        ('app', 'ReadAppConfig'),
        ('risks', 'GenerateReport'),
        ('auth', 'CreateToken'),
        ('auth', 'ManageSelf'),
        ('clients', 'Read'),
        ('saved_queries', 'Write'),
        ('saved_queries', 'Read'),
        ('graphdb', 'Read')
    ))
    OR
    (r.name = 'Read-Only' AND (p.authority, p.name) IN (
        ('app', 'ReadAppConfig'),
        ('risks', 'GenerateReport'),
        ('auth', 'ManageSelf'),
        ('auth', 'CreateToken'),
        ('saved_queries', 'Read'),
        ('graphdb', 'Read')
    ))
    OR
    (r.name = 'Upload-Only' AND (p.authority, p.name) IN (
        ('clients', 'Tasking'),
        ('graphdb', 'Write')
    ))
    OR
    (r.name = 'Power User' AND (p.authority, p.name) IN (
        ('app', 'ReadAppConfig'),
        ('app', 'WriteAppConfig'),
        ('risks', 'GenerateReport'),
        ('risks', 'ManageRisks'),
        ('auth', 'CreateToken'),
        ('auth', 'ManageSelf'),
        ('clients', 'Manage'),
        ('clients', 'Read'),
        ('clients', 'Tasking'),
        ('collection', 'ManageJobs'),
        ('graphdb', 'Write'),
        ('graphdb', 'Read'),
        ('saved_queries', 'Read'),
        ('saved_queries', 'Write'),
        ('graphdb', 'Mutate')
    ))
  );

-- Make sure the sequence values of things all line up at the end for the table we inserted
SELECT pg_catalog.setval('asset_groups_id_seq', MAX(id), true) FROM asset_groups;
SELECT pg_catalog.setval('feature_flags_id_seq', MAX(id), true) FROM feature_flags;
SELECT pg_catalog.setval('parameters_id_seq', MAX(id), true) FROM parameters;
SELECT pg_catalog.setval('permissions_id_seq', MAX(id), true) FROM permissions;
SELECT pg_catalog.setval('roles_id_seq', MAX(id), true) FROM roles;
-- Copyright 2024 Specter Ops, Inc.
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

-- Add Scheduled Analysis Configs
INSERT INTO parameters (key, name, description, value, created_at, updated_at)
VALUES ('analysis.scheduled',
        'Scheduled Analysis',
        'This configuration parameter allows setting a schedule for analysis. When enabled, analysis will only run when the scheduled time arrives',
        '{
          "enabled": false,
          "rrule": ""
        }',
        current_timestamp, current_timestamp)
ON CONFLICT DO NOTHING;

-- Add last analysis time to datapipe status so we can track scheduled analysis time properly
ALTER TABLE datapipe_status
    ADD COLUMN IF NOT EXISTS "last_analysis_run_at" TIMESTAMP with time zone;

-- SSO Provider
CREATE TABLE IF NOT EXISTS sso_providers
(
    id         SERIAL PRIMARY KEY,
    name       TEXT    NOT NULL,
    slug       TEXT    NOT NULL,
    type       INTEGER NOT NULL,

    updated_at TIMESTAMP WITH TIME ZONE DEFAULT now(),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT now(),

    UNIQUE (name),
    UNIQUE (slug)
);

-- OIDC Provider
CREATE TABLE IF NOT EXISTS oidc_providers
(
    id              SERIAL PRIMARY KEY,
    client_id       TEXT                                                    NOT NULL,
    issuer          TEXT                                                    NOT NULL,
    sso_provider_id INTEGER REFERENCES sso_providers (id) ON DELETE CASCADE NULL,

    updated_at      TIMESTAMP WITH TIME ZONE DEFAULT now(),
    created_at      TIMESTAMP WITH TIME ZONE DEFAULT now()
);

-- Create the reference from saml_providers to sso_providers
ALTER TABLE ONLY saml_providers
    ADD COLUMN IF NOT EXISTS sso_provider_id INTEGER NULL;
ALTER TABLE ONLY saml_providers
    DROP CONSTRAINT IF EXISTS fk_saml_provider_sso_provider;
ALTER TABLE ONLY saml_providers
    ADD CONSTRAINT fk_saml_provider_sso_provider FOREIGN KEY (sso_provider_id) REFERENCES sso_providers (id) ON DELETE CASCADE;

-- Backfill our sso_providers table with the existing data from saml_providers
-- The hardcoded type is determined by the AuthProvider
-- See:
-- https://github.com/SpecterOps/BloodHound/blob/main/cmd/api/src/model/auth.go#L565
INSERT INTO sso_providers(name, slug, type) (SELECT name, lower(replace(name, ' ', '-')), 1
                                             FROM saml_providers
                                             WHERE sso_provider_id IS NULL)
ON CONFLICT DO NOTHING;

-- Backfill the references from the newly created sso_provider entries
UPDATE saml_providers
SET sso_provider_id = (SELECT id FROM sso_providers WHERE name = saml_providers.name)
WHERE saml_providers.sso_provider_id IS NULL;

-- Add the sso_provider to the users table
ALTER TABLE ONLY users
    ADD COLUMN IF NOT EXISTS sso_provider_id INTEGER NULL;
ALTER TABLE ONLY users
    DROP CONSTRAINT IF EXISTS fk_users_sso_provider;
ALTER TABLE ONLY users
    ADD CONSTRAINT fk_users_sso_provider FOREIGN KEY (sso_provider_id) REFERENCES sso_providers (id) ON DELETE SET NULL;

-- Backfill users with their proper sso_provider when they have a saml_provider_id
UPDATE users u
SET sso_provider_id = (SELECT sso.id
                       FROM saml_providers saml
                                JOIN sso_providers sso ON sso.id = saml.sso_provider_id
                       WHERE u.saml_provider_id = saml.id)
WHERE sso_provider_id IS NULL
  AND saml_provider_id IS NOT NULL;
-- Copyright 2024 Specter Ops, Inc.
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

-- Add updated posture page feature flag
INSERT INTO feature_flags (created_at, updated_at, key, name, description, enabled, user_updatable)
VALUES (current_timestamp,
        current_timestamp,
        'updated_posture_page',
        'Updated Posture Page',
        'Enables the updated version of the posture page in the UI application',
        false,
        false)
ON CONFLICT DO NOTHING;

INSERT INTO permissions (authority, name, created_at, updated_at) VALUES ('graphdb', 'Ingest', current_timestamp, current_timestamp) ON CONFLICT DO NOTHING;

-- Grant the Upload-Only user GraphDBIngest permissions
INSERT INTO roles_permissions (role_id, permission_id)
VALUES ((SELECT id FROM roles WHERE roles.name = 'Upload-Only'),
        (SELECT id FROM permissions WHERE permissions.authority = 'graphdb' and permissions.name = 'Ingest'))
ON CONFLICT DO NOTHING;

-- Grant the Power User user GraphDBIngest permissions
INSERT INTO roles_permissions (role_id, permission_id)
VALUES ((SELECT id FROM roles WHERE roles.name = 'Power User'),
        (SELECT id FROM permissions WHERE permissions.authority = 'graphdb' and permissions.name = 'Ingest'))
ON CONFLICT DO NOTHING;

-- Grant the Admininstrator user GraphDBIngest permissions
INSERT INTO roles_permissions (role_id, permission_id)
VALUES ((SELECT id FROM roles WHERE roles.name = 'Administrator'),
        (SELECT id FROM permissions WHERE permissions.authority = 'graphdb' and permissions.name = 'Ingest'))
ON CONFLICT DO NOTHING;

-- Remove the GraphDBWrite permission from the Upload-Only role
DELETE FROM roles_permissions
WHERE role_id = (SELECT id FROM roles WHERE roles.name = 'Upload-Only')
AND permission_id = (SELECT id FROM permissions WHERE permissions.authority = 'graphdb' AND permissions.name = 'Write');

-- Set the user's saml_provider_id to null when an sso_provider or saml_provider is deleted
ALTER TABLE ONLY users
    DROP CONSTRAINT IF EXISTS fk_users_saml_provider;
ALTER TABLE ONLY users
    ADD CONSTRAINT fk_users_saml_provider FOREIGN KEY (saml_provider_id) REFERENCES saml_providers (id) ON DELETE SET NULL;

-- Backfill users with their proper sso_provider when they have a saml_provider_id
UPDATE users u
SET sso_provider_id = (SELECT sso.id
                       FROM saml_providers saml
                                JOIN sso_providers sso ON sso.id = saml.sso_provider_id
                       WHERE u.saml_provider_id = saml.id)
WHERE sso_provider_id IS NULL
  AND saml_provider_id IS NOT NULL;
-- Copyright 2024 Specter Ops, Inc.
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

-- Drop column saml_provider_id from users table
ALTER TABLE ONLY users
DROP CONSTRAINT IF EXISTS fk_users_saml_provider;
ALTER TABLE ONLY users
DROP COLUMN IF EXISTS saml_provider_id;

-- Add root_uri_version and backfill existing saml providers to 1 or "/v1/login/saml/"
ALTER TABLE ONLY saml_providers
  ADD COLUMN IF NOT EXISTS root_uri_version INTEGER NOT NULL DEFAULT 1;

-- Update root_uri_version to default to 2 or "/v2/sso/" for newly created saml providers
ALTER TABLE ONLY saml_providers
  ALTER COLUMN root_uri_version SET DEFAULT 2;

-- Set the `updated_posture_page` feature flag to true
UPDATE feature_flags SET enabled = true WHERE key = 'updated_posture_page';

-- Fix users in bad state due to sso bug
DELETE FROM auth_secrets WHERE id IN (SELECT auth_secrets.id FROM auth_secrets JOIN users ON users.id = auth_secrets.user_id WHERE users.sso_provider_id IS NOT NULL);

-- Set the `oidc_support` feature flag to true
UPDATE feature_flags SET enabled = true WHERE key = 'oidc_support';
-- Copyright 2025 Specter Ops, Inc.
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

-- Delete the `updated_posture_page` feature flag
DELETE FROM feature_flags WHERE key = 'updated_posture_page';

-- Add new config column in sso_providers table
ALTER TABLE IF EXISTS sso_providers ADD COLUMN IF NOT EXISTS config jsonb;

-- Update sso_providers table by backfilling existing sso providers' new config column with default values
UPDATE sso_providers set config = '{"auto_provision": {"enabled": false, "default_role_id": 0, "role_provision": false}}';
-- Copyright 2025 Specter Ops, Inc.
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

-- Prepend all found duplicate emails on user table with user id in preparation for unique constraint
UPDATE users SET email_address = id || '-' || lower(email_address) where lower(email_address) in (SELECT distinct(lower(email_address)) FROM users GROUP BY lower(email_address) HAVING count(lower(email_address)) > 1);

-- Add unique constraint on user emails
ALTER TABLE IF EXISTS users
  DROP CONSTRAINT IF EXISTS users_email_address_key;
ALTER TABLE IF EXISTS users
  ADD CONSTRAINT users_email_address_key UNIQUE (email_address);

-- Add `back_button_support` feature flag
INSERT INTO feature_flags (created_at, updated_at, key, name, description, enabled, user_updatable)
VALUES (current_timestamp,
        current_timestamp,
        'back_button_support',
        'Back Button Support',
        'Enable users to quickly navigate between views in a wider range of scenarios by utilizing the browser navigation buttons.',
        false,
        false)
ON CONFLICT DO NOTHING;

-- Add `tier_management_engine` feature flag
INSERT INTO feature_flags (created_at, updated_at, key, name, description, enabled, user_updatable)
VALUES (current_timestamp,
        current_timestamp,
        'tier_management_engine',
        'Tier Management Engine',
        'Updates the managed assets selector engine and the asset management page.',
        false,
        false)
ON CONFLICT DO NOTHING;

-- Add `NTLM Post Processing` feature flag
INSERT INTO feature_flags (created_at, updated_at, key, name, description, enabled, user_updatable)
VALUES (current_timestamp,
        current_timestamp,
        'ntlm_post_processing',
        'NTLM Post Processing Support',
        'Enable the post processing of NTLM relay attack paths, this will enable the creation of CoerceAndRelayNTLMTo[LDAP, LDAPS, ADCS, SMB] edges.',
        false,
        true)
ON CONFLICT DO NOTHING;
-- Copyright 2025 Specter Ops, Inc.
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

-- Set `back_button_support` feature flag as user updatable
UPDATE feature_flags SET user_updatable = true WHERE key = 'back_button_support';

-- Specify the `back_button_support` feature flag is currently only for BHCE users
UPDATE feature_flags SET description = 'Enable users to quickly navigate between views in a wider range of scenarios by utilizing the browser navigation buttons. Currently for BloodHound Community Edition users only.' WHERE key = 'back_button_support';

-- Add trusted_proxies
INSERT INTO parameters (key, name, description, value, created_at, updated_at)
VALUES ('http.trusted_proxies', 'Trusted Proxies',
        'This configuration parameter defines the number of trusted reverse proxies for enforcing our current rate limiting middleware',
        '{"trusted_proxies": 0}',
        current_timestamp, current_timestamp)
ON CONFLICT DO NOTHING;
-- Copyright 2025 Specter Ops, Inc.
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

-- This table is normally created by dawgs, as defined in schema_up.sql
-- We add it here to maintain a new FK to asset_group_tags below regardless
-- of graph driver selected. Any future changes to the schema should be reflected
-- in `schema_up.sql` as well
CREATE TABLE IF NOT EXISTS kind
(
  id   SMALLSERIAL,
  name varchar(256) not null,
  primary key (id),
  unique (name)
);

-- Add asset_group_tags table
CREATE TABLE IF NOT EXISTS asset_group_tags
(
    id SERIAL NOT NULL,
    type int NOT NULL,
    kind_id smallint,
    name text NOT NULL,
    description text NOT NULL DEFAULT '',
    created_at timestamp with time zone,
    created_by text,
    updated_at timestamp with time zone,
    updated_by text,
    deleted_at timestamp with time zone,
    deleted_by text,
    position integer,
    require_certify boolean,
    PRIMARY KEY (id),
    CONSTRAINT fk_kind_asset_group_tags FOREIGN KEY (kind_id) REFERENCES kind(id)
);

-- Add partial unique index for name for asset_group_tags
CREATE UNIQUE INDEX IF NOT EXISTS agl_name_unique_index ON asset_group_tags (name)
    WHERE deleted_at IS NULL;

-- Create tier xero record
WITH inserted_kind AS (
INSERT INTO kind (name) VALUES ('Tag_Tier_Zero') ON CONFLICT DO NOTHING
  RETURNING id)
INSERT INTO asset_group_tags (name, type, kind_id, description, created_by, created_at, updated_by, updated_at, position, require_certify)
  VALUES ('Tier Zero', 1, (SELECT id FROM inserted_kind), 'Tier Zero', 'SYSTEM', current_timestamp, 'SYSTEM', current_timestamp, 1, FALSE)
  ON CONFLICT DO NOTHING;

-- Add asset_group_history tables
CREATE TABLE IF NOT EXISTS asset_group_history
(
    id BIGSERIAL NOT NULL,
    actor text NOT NULL,
    action text NOT NULL,
    target text,
    asset_group_tag_id int NOT NULL,
    environment_id text,
    note text,
    created_at timestamp with time zone,
    PRIMARY KEY (id),
    CONSTRAINT fk_asset_group_history_asset_group_tags FOREIGN KEY (asset_group_tag_id) REFERENCES asset_group_tags(id)
);

-- As of v8.2.0, the auto_certify column type has been changed from a boolean to an integer type
-- Add asset_group_tag_selectors table
CREATE TABLE IF NOT EXISTS asset_group_tag_selectors
(
    id SERIAL NOT NULL,
    asset_group_tag_id int,
    created_at timestamp with time zone,
    created_by text,
    updated_at timestamp with time zone,
    updated_by text,
    disabled_at timestamp with time zone,
    disabled_by text,
    name text NOT NULL,
    description text NOT NULL DEFAULT '',
    is_default boolean NOT NULL DEFAULT FALSE,
    allow_disable boolean NOT NULL DEFAULT TRUE,
    auto_certify boolean NOT NULL DEFAULT FALSE,
    PRIMARY KEY (id),
    CONSTRAINT fk_asset_group_tags_asset_group_selectors FOREIGN KEY (asset_group_tag_id) REFERENCES asset_group_tags(id) ON DELETE CASCADE
);

-- Add asset_group_tag_selector_seeds table
CREATE TABLE IF NOT EXISTS asset_group_tag_selector_seeds
(
    selector_id int NOT NULL,
    type int NOT NULL,
    value text NOT NULL,
    CONSTRAINT fk_asset_group_tag_selectors_asset_group_tag_selector_seeds FOREIGN KEY (selector_id) REFERENCES asset_group_tag_selectors(id) ON DELETE CASCADE
);

-- generic ingest
ALTER TABLE IF EXISTS file_upload_jobs RENAME TO ingest_jobs;
ALTER TABLE ingest_tasks ADD COLUMN IF NOT EXISTS is_generic BOOLEAN NOT NULL DEFAULT FALSE;

-- GA for ntlm post processing
UPDATE feature_flags SET user_updatable = false WHERE key = 'ntlm_post_processing';
UPDATE feature_flags SET enabled = true WHERE key = 'ntlm_post_processing';
-- Copyright 2025 Specter Ops, Inc.
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

ALTER TABLE asset_group_history
	ADD COLUMN IF NOT EXISTS email VARCHAR(330) DEFAULT NULL;

-- Populate email for existing records by looking up the email address from the users table
UPDATE asset_group_history
	SET email = (SELECT email_address FROM users WHERE asset_group_history.actor = users.id)
	WHERE email IS NULL AND actor != 'SYSTEM';

-- Add asset_group_tag_selector_nodes table
CREATE TABLE IF NOT EXISTS asset_group_tag_selector_nodes
(
	selector_id int NOT NULL,
	node_id bigint NOT NULL,
	certified int NOT NULL DEFAULT 0,
	certified_by text,
	source int,
	created_at timestamp with time zone,
	updated_at timestamp with time zone,
	CONSTRAINT fk_asset_group_tag_selectors_asset_group_tag_selector_nodes FOREIGN KEY (selector_id) REFERENCES asset_group_tag_selectors(id) ON DELETE CASCADE,
	PRIMARY KEY (selector_id, node_id)
);

-- Add custom_node_kinds table
CREATE TABLE IF NOT EXISTS custom_node_kinds (
  id            SERIAL        PRIMARY KEY,
  kind_name     VARCHAR(256)  NOT NULL,
  config        JSONB         NOT NULL,

  created_at    TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
  updated_at    TIMESTAMP WITH TIME ZONE DEFAULT NOW(),

  unique(kind_name)
);

-- Migrate existing Tier Zero selectors
WITH inserted_selector AS (
  INSERT INTO asset_group_tag_selectors (asset_group_tag_id, created_at, created_by, updated_at, updated_by, name, description, is_default, allow_disable, auto_certify)
  SELECT (SELECT id FROM asset_group_tags WHERE name = 'Tier Zero'), current_timestamp, 'SYSTEM', current_timestamp, 'SYSTEM', s.name, s.selector, false, true, false
  FROM asset_group_selectors s JOIN asset_groups ag ON ag.id = s.asset_group_id
  WHERE ag.tag = 'admin_tier_0' and NOT EXISTS(SELECT 1 FROM asset_group_tag_selectors WHERE name = s.name)
  RETURNING id, description
)
INSERT INTO asset_group_tag_selector_seeds (selector_id, type, value) SELECT id, 1, description FROM inserted_selector;

-- Migrate existing Owned selectors
WITH inserted_kind AS (
  INSERT INTO kind (name)
  SELECT 'Tag_' || replace(name, ' ', '_') as name
  FROM asset_groups
  WHERE tag = 'owned'
  ON CONFLICT DO NOTHING
  RETURNING id, name
),
inserted_tag AS (
  INSERT INTO asset_group_tags (kind_id, type, name, description, created_at, created_by, updated_at, updated_by)
  SELECT ik.id, 3, ag.name, ag.name, current_timestamp, 'SYSTEM', current_timestamp, 'SYSTEM'
  FROM inserted_kind ik JOIN asset_groups ag ON ik.name = 'Tag_' || replace(ag.name, ' ', '_')
  RETURNING id, name
),
inserted_selector AS (
  INSERT INTO asset_group_tag_selectors (asset_group_tag_id, created_at, created_by, updated_at, updated_by, name, description, is_default, allow_disable, auto_certify)
  SELECT (SELECT id from inserted_tag), current_timestamp, 'SYSTEM', current_timestamp, 'SYSTEM', s.name, s.selector, false, true, false
  FROM asset_group_selectors s JOIN asset_groups ag ON ag.id = s.asset_group_id
  WHERE ag.tag = 'owned' and NOT EXISTS(SELECT 1 FROM asset_group_tag_selectors WHERE name = s.name)
  RETURNING id, description
)
INSERT INTO asset_group_tag_selector_seeds (selector_id, type, value) SELECT id, 1, description FROM inserted_selector;

-- Populate default cypher selectors
-- Add the following to the GA release migration to enable these for bootstrapped instances
-- UPDATE asset_group_tag_selectors SET disabled_at = NULL, disabled_by = NULL WHERE is_default = true AND created_at > current_timestamp - '1 min'::interval

WITH src_data AS (
	SELECT * FROM (VALUES
-- START
('Application Administrator', false, true, E'MATCH (n:AZRole) \nWHERE n.objectid STARTS WITH ''9B895D92-2CD3-44C7-9D02-A6AC2D5EA5C3@''\nRETURN n;', E'The Application Administrator role can control tenant-resident apps. This includes creating new credentials for apps, which can be used to authenticate the tenant as the app''s service principal and abuse the service principal privileges. The role is therefore considered Tier Zero if the tenant contains any Tier Zero service principals.'),
('Knowledge Administrator', false, true, E'MATCH (n:AZRole) \nWHERE n.objectid STARTS WITH ''B5A8DCF3-09D5-43A9-A639-8E29EF291470@''\nRETURN n;', E'The Knowledge Administrator role can control non-role-assignable groups. If any non-role-assignable group has compromising permissions over a Tier Zero asset (e.g. Contributor on a domain controller Azure VM), then the Knowledge Administrator role can add arbitrary principals to the given group and compromise Tier Zero. If no non-role-assignable group has compromising permissions over a Tier Zero asset, then there is no attack path to Tier Zero from the Knowledge Administrator role. It therefore depends on the usage of non-role-assignable groups whether the role should be considered Tier Zero.'),
('Account Operators', false, true, E'MATCH (n:Group)\nWHERE n.objectid ENDS WITH ''S-1-5-32-548''\nRETURN n;', E'The Account Operators group has GenericAll in the default security descriptor on the AD object classes: User, Group, and Computer. That means all objects of these types will be under full control of Account Operators unless they are protected with AdminSDHolder. Not all Tier Zero objects will be protected with AdminSDHolder typically, as not all Tier Zero objects will be included in Protected Accounts and Groups. This means Account Operators members have a path to compromise Tier Zero most often.\n\nIt is possible to delete all GenericAll ACEs for Account Operators on Tier Zero objects. To protect future Tier Zero objects, one would have to either remove the Account Operators ACE from the default security descriptors or implement a process of removing the ACEs as Tier Zero objects are being created. However, we recommend not using the group and classifying it as Tier Zero instead.'),
('Administrators', true, false, E'MATCH (n:Group)\nWHERE n.objectid ENDS WITH ''S-1-5-32-544''\nRETURN n;', E'The Administrators group has full control over most of AD''s essential objects and are inarguably part of Tier Zero.'),
('Backup Operators', true, false, E'MATCH (n:Group)\nWHERE n.objectid ENDS WITH ''S-1-5-32-551''\nRETURN n;', E'The Backup Operators group has the SeBackupPrivilege and SeRestorePrivilege rights on the domain controllers by default. These privileges allow members to access all files on the domain controllers, regardless of their permission, through backup and restore operations. Additionally, Backup Operators have full remote access to the registry of domain controllers. To compromise the domain, members of Backup Operators can dump the registry hives of a domain controller remotely, extract the domain controller account credentials, and perform a DCSync attack. Alternative ways to compromise the domain exist as well. The group is considered Tier Zero because of these known abuse techniques.'),
('Domain Admins', true, false, E'MATCH (n:Group)\nWHERE n.objectid ENDS WITH ''-512''\nRETURN n;', E'The Domain Admins group has full control over most of AD''s essential objects and are inarguably part of Tier Zero.'),
('Enterprise Admins', true, false, E'MATCH (n:Group)\nWHERE n.objectid ENDS WITH ''-519''\nRETURN n;', E'The Enterprise Admins group has full control over most of AD''s essential objects and are inarguably part of Tier Zero.'),
('Print Operators', false, true, E'MATCH (n:Group)\nWHERE n.objectid ENDS WITH ''S-1-5-32-550''\nRETURN n;', E'The Print Operators group has the local privilege on the domain controllers to load device drivers and can log on locally on domain controllers by default.\n\nIt is feasible to remove the logon privilege from the group on the domain controllers, such that the group has no known abusable path to Tier Zero. However, the local privilege to load device drivers is considered a security dependency for the domain controllers, and the group is therefore considered Tier Zero.'),
('Schema Admins', true, false, E'MATCH (n:Group)\nWHERE n.objectid ENDS WITH ''-518''\nRETURN n;', E'The Schema Admins group has full control over the AD schema. This allows the group members to create or modify ACEs for future AD objects. An attacker could grant full control to a compromised principal on any object type and wait for the next Tier Zero asset to be created, to then have a path to Tier Zero. This attack could be remediated by removing any unwanted ACEs on objects before they are promoted to Tier Zero, but we recommend considering the group as Tier Zero instead.'),
('Server Operators', false, true, E'MATCH (n:Group)\nWHERE n.objectid ENDS WITH ''S-1-5-32-549''\nRETURN n;', E'The Server Operators group has local privileges on the domain controllers and perform administrative operations as creating backups of all files. The group can log on locally on domain controllers by default.\n\nIt is feasible to remove the logon privilege from the group on the domain controllers, such that the group has no known abusable path to Tier Zero. However, the local privileges are considered security dependencies for the domain controllers, and the groups are therefore considered Tier Zero.'),
('Administrator', true, false, E'MATCH (n:User)\nWHERE n.objectid ENDS WITH ''-500''\nRETURN n;', E'The built-in Administrator account has admin access to DCs by default and is therefore Tier Zero.'),
('AdminSDHolder', true, false, E'MATCH (n:Domain)\nMATCH (m:Container)\nWHERE m.distinguishedname = ''CN=ADMINSDHOLDER,CN=SYSTEM,'' + n.distinguishedname\nRETURN m;', E'The permissions configured on AdminSDHolder is a template that will be applied on Protected Groups and Users with SDProp, by default every hour. Control over AdminSDHolder means you have control over the Protected Groups (and their members) and Users, which include Tier Zero groups such as Domain Admins. The AdminSDHolder container is therefore a Tier Zero object.'),
('Domain root object', true, false, E'MATCH (n:Domain)\nRETURN n;', E'An attacker with control over the domain root object can compromise the domain in multiple ways, for example by a DCSync attack (see reference). The domain root object is therefore Tier Zero.'),
('KRBTGT objects', false, true, E'MATCH (n:User)\nWHERE n.objectid ENDS WITH ''-502''\nRETURN n;', E'The krbtgt''s credentials allow one to create golden ticket and compromise the domain. Therefore, if you obtain the credentials of this account, then you can authenticate as any Tier Zero user. However, there is currently no known privilege on the object to obtain the Kerberos keys or to compromise the account in any other way. When you reset the password of krbtgt, AD will ignore your password input and use a random string instead. So, the reset password privilege does not work for a compromise. An attacker could use the reset password privilege to harm Tier Zero, as a double password reset causes all Kerberos TGTs in the domain to become invalid. So, since control over the account can harm Tier Zero, and there is no reason for delegating control to non-Tier Zero, the krbtgt is Tier Zero.'),
('Read-Only Domain Controllers', false, true, E'MATCH (n:Computer)-[:MemberOf]->(m:Group) \nWHERE m.objectid ENDS WITH ''-521''\nRETURN n;', E'An attacker with control over a RODC computer object can compromise Tier Zero principals. The attacker can modify the msDS-RevealOnDemandGroup and msDS-NeverRevealGroup attributes of the RODC computer object such that the RODC can retrieve the credentials of a targeted Tier Zero principal. The attacker can obtain admin access to the OS of the RODC through the managedBy attribute, from where they can obtain the credentials of the RODC krbtgt account. With that, the attacker can create a RODC golden ticket for the target principal. This ticket can be converted to a real golden ticket as the target has been added to the msDS-RevealOnDemandGroup attribute and is not protected by the msDS-NeverRevealGroup attribute. Therefore, the RODC computer object is Tier Zero.'),
('Global Administrator', true, false, E'MATCH (n:AZRole) \nWHERE n.objectid STARTS WITH ''62E90394-69F5-4237-9190-012177145E10@''\nRETURN n;', E'The Global Administrator role is the highest privilege role in Entra ID and inarguably part of Tier Zero. It can do almost anything, and grant permission to do the things it cannot do.'),
('Partner Tier2 Support', true, false, E'MATCH (n:AZRole) \nWHERE n.objectid STARTS WITH ''E00E864A-17C5-4A4B-9C06-F5B95A8D5BD8@''\nRETURN n;', E'The Partner Tier2 Support role can reset the password for any principal, including principals with the Global Administrator role. The role is therefore considered Tier Zero.'),
('Privileged Authentication Administrator', true, false, E'MATCH (n:AZRole) \nWHERE n.objectid STARTS WITH ''7BE44C8A-ADAF-4E2A-84D6-AB2649E08A13@''\nRETURN n;', E'The Privileged Authentication Administrator role can set or reset any authentication method (including passwords) for any principal, including principals with the Global Administrator role. The role is therefore considered Tier Zero.'),
('Privileged Role Administrator', true, false, E'MATCH (n:AZRole) \nWHERE n.objectid STARTS WITH ''E8611AB8-C189-46E8-94E1-60213AB1F814@''\nRETURN n;', E'The Privileged Role Administrator role can grant any other admin role to any principal at the tenant level. The role is therefore considered Tier Zero.'),
('Enterprise CA Computers', false, true, E'MATCH (n:Computer)-[:HostsCAService]->(:EnterpriseCA)\nRETURN n;', E'Enterprise CAs can by default issue certificates that enable authentication as anyone, thereby allowing takeover of Tier Zero. An attacker with admin rights on an enterprise CA can obtain a certificate as any user in different ways. One option is to dump the private key of the CA and craft a ''golden certificate'' as a target user. This attack can be prevented by protecting the private key with hardware. Alternatively, the attacker can publish any template, modify pending certificate requests, and issue denied requests, which typically also enable a takeover of Tier Zero. Enterprise CA computer objects are therefore Tier Zero.\n\nIf the enterprise CA certificate is removed from the NTAuth store, then certificates from this CA cannot be used for domain authentication, thus preventing a Tier Zero takeover.'),
('Exchange Trusted Subsystem', false, true, E'MATCH (n:Group)\nWHERE n.name STARTS WITH ''EXCHANGE TRUSTED SUBSYSTEM@''\nRETURN n;', E'The Exchange Trusted Subsystem group has takeover permissions on all users with the default ACL inheritance enabled from the domain, regardless of the permission model Exchange is configured to. The compromising permission is write access to the AltSecurityIdentities attribute, which allows an attacker to add an explicit mapping for the user for domain authentication. Typically, some Tier Zero users inherit permissions from the domain. The group is therefore Tier Zero.\n\nThe group can only be treated as non-Tier Zero if all Tier Zero users are protected from this compromising permission.'),
('Exchange Windows Permissions', false, true, E'MATCH (n:Group)\nWHERE n.name STARTS WITH ''EXCHANGE WINDOWS PERMISSIONS@''\nRETURN n;', E'The Exchange Windows Permissions group has takeover permissions on all users (WriteDACL and reset password) and all groups (edit membership) with the default ACL inheritance enabled from the domain, if Exchange is configured with the default shared permission model or the RBAC split model. Typically, some Tier Zero users and groups inherit permissions from the domain. The group is therefore Tier Zero.\n\nIf Exchange is configured in the AD split model, then this group has no compromising permissions and can be treated as non-Tier Zero.'),
('DNS Admins', false, true, E'MATCH (n:Group)\nWHERE n.name STARTS WITH ''DNSADMINS@''\nRETURN n;', E'DnsAdmins controls DNS which enables an attacker to trick a privileged victim to authenticate against an attacker-controlled host as it was another host. This enables a Kerberos relay attack. Also, control over DNS enables disruption of Tier Zero since Kerberos depends on DNS by default.\n\nThe group could previously use a feature in the Microsoft DNS management protocol to make the DNS service load any DLL and thereby obtain a session as SYSTEM on the DNS server. This vulnerability was patched in Dec 2021.'),
('Domain Controllers', true, false, E'MATCH (n:Group)\nWHERE n.objectid ENDS WITH ''-516''\nRETURN n;', E'The Domain Controllers group has the GetChangesAll privilege on the domain. This is not enough to perform DCSync, where the GetChanges privilege is also required.\n\nThere are no known ways to abuse membership in this group to compromise Tier Zero. However, the GetChangesAll privilege is considered a security dependency that should only be held by Tier Zero principals. Additionally, control over the group allows one to impact the operability of Tier Zero by removing domain controllers from the group, which breaks AD replication. The group is therefore considered Tier Zero.'),
('Intune Administrator', false, true, E'MATCH (n:AZRole) \nWHERE n.objectid STARTS WITH ''3A2C62DB-5318-420D-8D74-23AFFEE5D9D5@''\nRETURN n;', E'The Intune Administrator role has permission to execute scripts locally on Entra-managed devices. The role has therefore a potential attack path to Tier Zero through Entra-managed devices used by Tier Zero principals. Furthermore, the Intune Administrator role can manage Conditional Access, which can be abused to lower the security of Tier Zero or prevent the operability of Tier Zero. The role is therefore considered Tier Zero.'),
('Security Administrator', false, true, E'MATCH (n:AZRole) \nWHERE n.objectid STARTS WITH ''194AE4CB-B126-40B2-BD5B-6091B380977D@''\nRETURN n;', E'The Security Administrator role has access to Live Response API (if not disabled) with permission to execute scripts locally on Entra-managed devices. The role has therefore a potential attack path to Tier Zero through Entra-managed devices used by Tier Zero principals. Furthermore, the Security Administrator role can manage Conditional Access, which can be abused to lower the security of Tier Zero or prevent the operability of Tier Zero. The role is therefore considered Tier Zero.'),
('Cert Publishers', false, true, E'MATCH (n:Group)\nWHERE n.objectid ENDS WITH ''-517''\nRETURN n;', E'The Cert Publishers group has full control permissions on root CA and AIA CA objects. This enables an attacker to add or remove certificates for these objects, which are trusted throughout the AD forest. As certificate authentication requires the certificate to chain up to a trusted root CA, an attacker could prevent successful authentication for AD accounts and disrupt Tier Zero operations. The group is therefore Tier Zero.\n\nIn some environments, the group also has full control over the NTAuth store. In that scenario, the group can take over the forest by adding a forged root certificate, making it trusted for NTAuth.'),
('NTAuth store', false, true, E'MATCH (n:NTAuthStore) \nRETURN n;', E'The NTAuth store is a security dependency for Tier Zero. A certificate that impersonates any user in AD must chain up to a trusted root CA and be issued by a CA trusted by the NTAuth store. With control over a root CA and the NTAuth store, an attacker can make an attacker-controlled root CA certificate meet these requirements and issue certificates as anyone, taking over Tier Zero. Control over the NTAuth store alone may be sufficient to disrupt Tier Zero operations, as the attacker can delete CA certificates that Tier Zero principals or systems rely on for authentication. The NTAuth store is therefore Tier Zero.'),
('Root CA object', false, true, E'MATCH (n:RootCA) \nRETURN n;', E'A root CA is a security dependency for Tier Zero. A certificate that impersonates any user in AD must chain up to a trusted root CA and be issued by a CA trusted by the NTAuth store. With control over a root CA and the NTAuth store, an attacker can make an attacker-controlled root CA certificate meet these requirements and issue certificates as anyone, taking over Tier Zero. Control over a root CA alone may be sufficient to disrupt Tier Zero operations, as the attacker can delete root CA certificates that Tier Zero principals or systems rely on for authentication. Root CA objects are therefore Tier Zero.'),
('Enterprise Key Admins', true, false, E'MATCH (n:Group)\nWHERE n.objectid ENDS WITH ''-527''\nRETURN n;', E'The Enterprise Key Admins group has write access to the msds-KeyCredentialLink attribute on all users (not protected by AdminSDHolder) and on all computers in the AD forest. This enables the group to compromise all these principals through Shadow Credentials attacks. The group is therefore considered Tier Zero.'),
('Key Admins', true, false, E'MATCH (n:Group)\nWHERE n.objectid ENDS WITH ''-526''\nRETURN n;', E'The Key Admins group has write access to the msds-KeyCredentialLink attribute on all users (not protected by AdminSDHolder) and on all computers in the AD domain. This enables the group to compromise all these principals through Shadow Credentials attacks. The group is therefore considered Tier Zero.'),
('Cryptographic Operators', false, true, E'MATCH (n:Group)\nWHERE n.objectid ENDS WITH ''S-1-5-32-569''\nRETURN n;', E'The Cryptographic Operators group has the local privilege on domain controllers to perform cryptographic operations but no privilege to log in.\n\nThere are no known ways to abuse the membership of the group to compromise Tier Zero. The local privilege the group has on the domain controllers is considered security dependencies, and the group is therefore considered Tier Zero.'),
('Distributed COM Users', false, true, E'MATCH (n:Group)\nWHERE n.objectid ENDS WITH ''S-1-5-32-562''\nRETURN n;', E'The Distributed COM Users group has local privileges on domain controllers to launch, activate, and use Distributed COM objects but no privilege to log in.\n\nThere are no known ways to abuse the membership of the group to compromise Tier Zero. The local privileges the group has on the DCs are considered security dependency, and the group is therefore considered Tier Zero.'),
('AIA CA (AD object)', false, true, E'MATCH (n:AIACA) \nRETURN n;', E'The AIA CA objects may represent offline enterprise CAs or cross CAs. In such cases, deleting the AIA CA object would cause certificates, potentially of Tier Zero principals, to lose trust. We therefore recommend to treat AIACAs as Tier Zero.'),
('Enterprise CA (AD object)', false, true, E'MATCH (n:EnterpriseCA) \nRETURN n;', E'Control over an enterprise CA object enables an attacker to publish certificate templates. If any templates that allow ADCS domain escalation exist but are unpublished, then control over the enterprise CA object could enable a takeover of Tier Zero. An attacker could potentially also disrupt or takeover Tier Zero by deleting the certificate of the enterprise CA or changing the DNShostName of the enterprise CA to an attcker-controlled host. Enterprise CA objects are therfore Tier Zero.\n\nIf the enterprise CA certificate is removed from the NTAuth store, certificates from this CA cannot be used for domain authentication, thus preventing a Tier Zero takeover.'),
('Performance Log Users', false, true, E'MATCH (n:Group)\nWHERE n.objectid ENDS WITH ''S-1-5-32-559''\nRETURN n;', E'The Performance Log Users group has local privileges on domain controllers to launch, activate, and use Distributed COM objects but no privilege to log in.\n\nThere are no known ways to abuse the membership of the group to compromise Tier Zero. The local privileges the group has on the DCs are considered security dependency, and the group is therefore considered Tier Zero.'),
('Certificate template', false, true, E'MATCH (n:CertTemplate) \nRETURN n;', E'Control over a certificate template enables the ADCS ESC4 attack and Tier Zero takeover if the template is published to a CA trusted in the NTAuth store and that chains up to a trusted root CA. There are default templates that meet this requirement; others remain unpublished. A template cannot be used if it is not published, making control over an unpublished object less concerning. However, if it is ever published, it becomes a risk. We, therefore, recommend treating all certificate templates as Tier Zero objects, whether published or not.'),
('Azure tenant object', true, false, E'MATCH (n:AZTenant) \nRETURN n;', E'An attacker with control of the Tenant Root Object has control of all identities, applications, roles, and devices that reside in that tenant. Further, control of the Tenant Root Object enables an attacker to gain control of all Azure Resource Manager subscriptions that trust the tenant. This object is therefore considered Tier Zero.'),
('Enterprise Domain Controllers', true, false, E'MATCH (n:Group)\nWHERE n.objectid ENDS WITH ''-1-5-9''\nRETURN n;', E'There are no known ways to abuse membership in this group to compromise Tier Zero. However, the GetChangesAll privilege is considered a security dependency that should only be held by Tier Zero principals. Additionally, control over the group allows one to impact the operability of Tier Zero by removing domain controllers from the group, which breaks AD replication. The group is therefore considered Tier Zero."'),
('AZUREADSSOACC object', false, true, E'MATCH (n:Computer)\nWHERE n.samaccountname = ''AZUREADSSOACC$''\nRETURN n;', E'Microsoft automatically creates the AZUREADSSOACC account when enabling Seamless SSO. When configured for Seamless SSO, this object has the ability to modify any synced object within an Azure environment, granting significant control over the organization.')
-- END
	) AS s (name, enabled, allow_disable, cypher, description)
), inserted_selectors AS (
	INSERT INTO asset_group_tag_selectors (
		asset_group_tag_id,
		created_at,
		created_by,
		updated_at,
		updated_by,
		disabled_at,
		disabled_by,
		name,
		description,
		is_default,
		allow_disable,
		auto_certify
	)
	SELECT
		(SELECT id FROM asset_group_tags WHERE name = 'Tier Zero'),
		current_timestamp,
		'SYSTEM',
		current_timestamp,
		'SYSTEM',
		CASE WHEN NOT d.enabled THEN current_timestamp ELSE NULL END,
		CASE WHEN NOT d.enabled THEN 'SYSTEM' ELSE NULL END,
		d.name,
		d.description,
		true,
		d.allow_disable,
		false
	FROM src_data d WHERE NOT EXISTS(SELECT 1 FROM asset_group_tag_selectors WHERE name = d.name)
	RETURNING id, name
)
INSERT INTO asset_group_tag_selector_seeds (selector_id, type, value)
SELECT
	s.id,
	2,
	d.cypher
FROM inserted_selectors s JOIN src_data d ON d.name = s.name;

-- Enable `back_button_support` feature flag and block users from updating it.
UPDATE feature_flags SET user_updatable = false, enabled = true WHERE key = 'back_button_support';

UPDATE feature_flags SET "user_updatable" = true WHERE "key" = 'tier_management_engine';
-- Copyright 2025 Specter Ops, Inc.
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

-- is_generic column not actually needed.
ALTER TABLE ingest_tasks
DROP COLUMN IF EXISTS is_generic;

-- create explore_table_view feature flag, disable it, and make it non user-updatable.
INSERT INTO feature_flags (created_at, updated_at, key, name, description, enabled, user_updatable)
VALUES (current_timestamp,
        current_timestamp,
        'explore_table_view',
        'Explore Table View',
        'Adds a layout option to the Explore page that will display all nodes in a table view. It also will automatically display the table when a cypher query returned only nodes.',
        false,
        false)
ON CONFLICT DO NOTHING;

 -- Add Tier Management Parameter
INSERT INTO parameters (key, name, description, value, created_at, updated_at) VALUES ('analysis.tiering', 'Multi-Tier Analysis Configuration', 'This configuration parameter determines the limits of tiering with respect to analysis', '{"tier_limit": 1, "label_limit": 0, "multi_tier_analysis_enabled": false}', current_timestamp, current_timestamp) ON CONFLICT DO NOTHING;
-- Copyright 2025 Specter Ops, Inc.
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

-- Add analysis_enabled flag to asset_group_tags
ALTER TABLE asset_group_tags ADD COLUMN IF NOT EXISTS analysis_enabled BOOL;

-- Set analysis_enabled to true for tier zero and false for other tiers
UPDATE asset_group_tags SET analysis_enabled = position = 1 WHERE type = 1 AND analysis_enabled IS NULL;

-- Add EULA custom text
INSERT INTO parameters (key, name, description, value, created_at, updated_at) VALUES ('eula.custom_text', 'EULA Custom Text', 'This configuration parameter overrides the EULA agreement text with provided text.', '{"custom_text": ""}', current_timestamp, current_timestamp) ON CONFLICT DO NOTHING;

-- Add Auth Session TTL Hours
INSERT INTO parameters (key, name, description, value, created_at, updated_at) VALUES ('auth.session_ttl_hours', 'Auth Session TTL Hours', 'This configuration parameter determines the length of time in hours a logged in session stays active before expiration.', '{"hours": 8}', current_timestamp, current_timestamp) ON CONFLICT DO NOTHING;

-- Retire the `auto_tag_t0_parent_objects` feature flag
UPDATE feature_flags SET enabled = true, user_updatable = false WHERE key = 'auto_tag_t0_parent_objects';

-- Add RO-DC default selector to Tier Zero
WITH src_data AS (
  SELECT * FROM (VALUES
-- START
('Read-Only DCs', false, true, E'MATCH (n:Computer)\nWHERE n.isReadOnlyDC = true\nRETURN n;', E'An attacker with control over a RODC computer object can compromise Tier Zero principals. The attacker can modify the msDS-RevealOnDemandGroup and msDS-NeverRevealGroup attributes of the RODC computer object such that the RODC can retrieve the credentials of a targeted Tier Zero principal. The attacker can obtain admin access to the OS of the RODC through the managedBy attribute, from where they can obtain the credentials of the RODC krbtgt account. With that, the attacker can create a RODC golden ticket for the target principal. This ticket can be converted to a real golden ticket as the target has been added to the msDS-RevealOnDemandGroup attribute and is not protected by the msDS-NeverRevealGroup attribute. Therefore, the RODC computer object is Tier Zero.')
-- END
  ) AS s (name, enabled, allow_disable, cypher, description)
), inserted_selectors AS (
INSERT INTO asset_group_tag_selectors (
  asset_group_tag_id,
  created_at,
  created_by,
  updated_at,
  updated_by,
  disabled_at,
  disabled_by,
  name,
  description,
  is_default,
  allow_disable,
  auto_certify
)
SELECT
  (SELECT id FROM asset_group_tags WHERE type = 1 and position = 1 LIMIT 1),
  current_timestamp,
  'SYSTEM',
  current_timestamp,
  'SYSTEM',
  CASE WHEN NOT d.enabled THEN current_timestamp ELSE NULL END,
  CASE WHEN NOT d.enabled THEN 'SYSTEM' ELSE NULL END,
  d.name,
  d.description,
  true,
  d.allow_disable,
  false
FROM src_data d WHERE NOT EXISTS(SELECT 1 FROM asset_group_tag_selectors WHERE name = d.name)
  RETURNING id, name
)
INSERT INTO asset_group_tag_selector_seeds (selector_id, type, value)
SELECT
  s.id,
  2,
  d.cypher
FROM inserted_selectors s JOIN src_data d ON d.name = s.name;
-- Copyright 2025 Specter Ops, Inc.
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
create table if not exists source_kinds (
  id smallserial,
  name varchar(256) not null,
  primary key (id),
  unique (name)
);
INSERT INTO source_kinds (name)
VALUES ('Base'),
  ('AZBase') ON CONFLICT (name) DO NOTHING;
ALTER TABLE analysis_request_switch
ADD COLUMN IF NOT EXISTS delete_all_graph boolean DEFAULT false,
  ADD COLUMN IF NOT EXISTS delete_sourceless_graph boolean DEFAULT false,
  ADD COLUMN IF NOT EXISTS delete_source_kinds text [] DEFAULT ARRAY []::text [];
-- Remove the ReadAppConfig / WriteAppConfig from power users role
DELETE FROM roles_permissions
WHERE role_id = (
    SELECT id
    FROM roles
    WHERE roles.name = 'Power User'
  )
  AND permission_id IN (
    SELECT id
    FROM permissions
    WHERE permissions.authority = 'app'
      AND permissions.name IN ('WriteAppConfig')
  );
-- Add name index to asset_group_tag_selectors table for search
CREATE INDEX IF NOT EXISTS idx_asset_group_tag_selectors_name ON asset_group_tag_selectors USING btree (name);
-- if the explore_table_view flag doesnt exist, create explore_table_view feature flag, disable it, and make it non user-updatable.
-- this same query exists in 7.5.0, however it was merged after the release was cut so any tenants created before 7.5.0 will be missing this flag.
INSERT INTO feature_flags (created_at, updated_at, key, name, description, enabled, user_updatable)
VALUES (current_timestamp,
        current_timestamp,
        'explore_table_view',
        'Explore Table View',
        'Adds a layout option to the Explore page that will display all nodes in a table view. It also will automatically display the table when a cypher query returned only nodes.',
        false,
        false)
ON CONFLICT DO NOTHING;

-- enable explore_table_view feature flag
UPDATE feature_flags
SET enabled = true
WHERE key = 'explore_table_view';


-- Add Incoming Forest Trust Builders selector
WITH s AS (
  INSERT INTO asset_group_tag_selectors (
      asset_group_tag_id,
      created_at,
      created_by,
      updated_at,
      updated_by,
      disabled_at,
      disabled_by,
      name,
      description,
      is_default,
      allow_disable,
      auto_certify
    )
  SELECT
      (
        SELECT id
        FROM asset_group_tags
        WHERE name = 'Tier Zero'
      ),
      current_timestamp,
      'SYSTEM',
      current_timestamp,
      'SYSTEM',
      current_timestamp,
      'SYSTEM',
      'Incoming Forest Trust Builders',
      E'Members of this group can create incoming trusts that allow TGT delegation which can lead to compromise of the forest.',
      true,
      true,
      false
  WHERE NOT EXISTS(SELECT 1 FROM asset_group_tag_selectors WHERE name = 'Incoming Forest Trust Builders')
  RETURNING id
)
INSERT INTO asset_group_tag_selector_seeds (selector_id, type, value)
SELECT
	s.id,
	2,
	E'MATCH (n:Group) \nWHERE n.objectid ENDS WITH ''-557''\nRETURN n;'
FROM s;-- Copyright 2025 Specter Ops, Inc.
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

-- Environment Targeted Access Control Feature Flag
INSERT INTO feature_flags (created_at, updated_at, key, name, description, enabled, user_updatable)
VALUES (current_timestamp,
        current_timestamp,
        'targeted_access_control',
        'Targeted Access Control',
        'Enable power users and admins to set targeted access controls on users',
        false,
        false)
ON CONFLICT DO NOTHING;

-- Environment Targeted Access Control
CREATE TABLE IF NOT EXISTS environment_access_control (
    id BIGSERIAL PRIMARY KEY,
    user_id TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    environment TEXT NOT NULL,
    created_at timestamp with time zone DEFAULT current_timestamp,
    updated_at timestamp with time zone,
    CONSTRAINT environment_access_control_user_env_key UNIQUE (user_id, environment),
    CONSTRAINT environment_not_blank CHECK (btrim(environment) <> '')
);

ALTER TABLE users ADD COLUMN IF NOT EXISTS all_environments BOOL DEFAULT TRUE;

-- Add denormalized property columns to asset_group_tag_selector_nodes table
ALTER TABLE asset_group_tag_selector_nodes
  ADD COLUMN IF NOT EXISTS node_primary_kind TEXT,
  ADD COLUMN IF NOT EXISTS node_environment_id TEXT,
  ADD COLUMN IF NOT EXISTS node_object_id TEXT,
  ADD COLUMN IF NOT EXISTS node_name TEXT;

-- Add indexes for the above new columns added to asset_group_tag_selector_nodes table
CREATE INDEX IF NOT EXISTS idx_agt_selector_nodes_primary_kind ON asset_group_tag_selector_nodes USING btree (node_primary_kind);
CREATE INDEX IF NOT EXISTS idx_agt_selector_nodes_environment_id ON asset_group_tag_selector_nodes USING btree (node_environment_id);
CREATE INDEX IF NOT EXISTS idx_agt_selector_nodes_object_id ON asset_group_tag_selector_nodes USING btree (node_object_id);
CREATE INDEX IF NOT EXISTS idx_agt_selector_nodes_name ON asset_group_tag_selector_nodes USING btree (node_name);

ALTER TABLE asset_group_tags
        ADD COLUMN IF NOT EXISTS glyph TEXT UNIQUE;

-- File Ingest Details
ALTER TABLE ingest_tasks ADD COLUMN IF NOT EXISTS original_file_name text NOT NULL DEFAULT '';
ALTER TABLE ingest_tasks RENAME COLUMN file_name TO stored_file_name;
CREATE TABLE IF NOT EXISTS completed_tasks (
    id BIGSERIAL PRIMARY KEY,
    ingest_job_id BIGINT NOT NULL REFERENCES ingest_jobs(id) ON DELETE CASCADE,
    file_name TEXT NOT NULL,
    parent_file_name TEXT NOT NULL,
    errors TEXT[] NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone
);
CREATE INDEX IF NOT EXISTS idx_completed_tasks_ingest_job_id ON completed_tasks USING btree (ingest_job_id);

-- Add FinishedJobsLog rework feature flag
INSERT INTO feature_flags (created_at, updated_at, key, name, description, enabled, user_updatable)
VALUES (
  current_timestamp,
  current_timestamp,
  'finished_jobs_log_v2',
  'Finished Jobs Log Update',
  'An updated Finished Jobs Log with filtering and more info.',
  false,
  false
)
ON CONFLICT DO NOTHING;
-- Copyright 2025 Specter Ops, Inc.
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
-- Add OpenGraph Phase 2 feature flag
INSERT INTO feature_flags (created_at, updated_at, key, name, description, enabled, user_updatable)
VALUES (
           current_timestamp,
           current_timestamp,
           'open_graph_phase_2',
           'Open Graph Phase 2',
           'Open Graph Phase 2 features',
           false,
           false
       )
ON CONFLICT DO NOTHING;

INSERT INTO permissions(created_at, updated_at, authority, name)
VALUES (
        current_timestamp,
        current_timestamp,
        'auth',
        'ReadUsers'
       )
ON CONFLICT DO NOTHING;

INSERT INTO roles_permissions (role_id, permission_id)
SELECT r.id, p.id
FROM roles r
JOIN permissions p
ON (p.authority, p.name) = ('auth', 'ReadUsers')
WHERE r.name IN ('Administrator', 'User', 'Read-Only', 'Power User')
ON CONFLICT DO NOTHING;

INSERT INTO feature_flags (created_at, updated_at, key, name, description, enabled, user_updatable)
VALUES (
           current_timestamp,
           current_timestamp,
           'changelog',
           'Changelog',
           'This flag allows the application to query the changelog daemon for deduplication of ingest payloads.',
           false,
           false
       )
ON CONFLICT DO NOTHING;

-- Add Stale Client Updated Logic rework parameter
INSERT INTO parameters (key, name, description, value, created_at, updated_at)
VALUES (
         'pipeline.updated_stale_client',
        'Stale Client Updated Logic',
        'Is used to updated the logic used for if a job has become stale. With this enabled, rather than checking the last ingest time, the last checkin time of the client is checked to timeout the job.',
        '{"enabled": true}',
           current_timestamp,
           current_timestamp
       )
  ON CONFLICT DO NOTHING;
-- Copyright 2025 Specter Ops, Inc.
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


-- Set all_environments to true for existing users
UPDATE users SET all_environments = true;

-- Rename environment to environment_id to prepare for data partitioning, if the column does not exist then we throw away the error for idempotence
-- +goose StatementBegin
DO
$$
    BEGIN
        ALTER TABLE environment_access_control
            RENAME COLUMN environment TO environment_id;
    EXCEPTION
        WHEN undefined_column THEN
            NULL;
        WHEN undefined_table THEN
            NULL;
    END;
$$;
-- +goose StatementEnd

-- This migration changes the auto_certify column type from a boolean to an integer type
-- Then it converts the previous boolean values into enum-like integer values of 0, 1, or 2
-- +goose StatementBegin
DO $$
	BEGIN
		IF (
            SELECT data_type
            FROM information_schema.columns
            WHERE table_name = 'asset_group_tag_selectors' AND column_name = 'auto_certify'
        ) = 'boolean' THEN
		    ALTER TABLE asset_group_tag_selectors ADD COLUMN auto_certify_int INTEGER NOT NULL DEFAULT 0;

		    -- 0 means disabled
		    -- 1 is enabled for all objects (seeds, children, parents)
		    -- 2 is enabled for seeds only
		    UPDATE asset_group_tag_selectors selectors
		    SET auto_certify_int = CASE
		        WHEN selectors.is_default THEN 2
		        WHEN selectors.auto_certify = TRUE THEN 1
		        WHEN EXISTS (
		            SELECT *
		            FROM asset_group_tag_selector_seeds seeds
		            WHERE seeds.type = 1 AND seeds.selector_id = selectors.id
		        ) THEN 2
		        ELSE 0
		    END
		    FROM asset_group_tags tags
		    WHERE selectors.asset_group_tag_id = tags.id;

		    ALTER TABLE asset_group_tag_selectors DROP COLUMN auto_certify;
		    ALTER TABLE asset_group_tag_selectors RENAME COLUMN auto_certify_int TO auto_certify;
		END IF;
	END;
$$;
-- +goose StatementEnd

CREATE INDEX IF NOT EXISTS idx_agt_history_actor ON asset_group_history USING btree (actor);
CREATE INDEX IF NOT EXISTS idx_agt_history_action ON asset_group_history USING btree (action);
CREATE INDEX IF NOT EXISTS idx_agt_history_target ON asset_group_history USING btree (target);
CREATE INDEX IF NOT EXISTS idx_agt_history_email ON asset_group_history USING btree (email);
CREATE INDEX IF NOT EXISTS idx_agt_history_env_id ON asset_group_history USING btree (environment_id);
CREATE INDEX IF NOT EXISTS idx_agt_history_created_at ON asset_group_history USING btree (created_at);

-- Remigrate old custom AGI selectors to PZ selectors for any instances without PZ feature flag enabled
-- +goose StatementBegin
DO $$
  BEGIN
		IF
      (SELECT enabled FROM feature_flags WHERE key  = 'tier_management_engine') = false
    THEN
       -- Delete custom selectors
       DELETE FROM asset_group_tag_selectors WHERE is_default = false AND asset_group_tag_id IN ((SELECT id FROM asset_group_tags WHERE position = 1), (SELECT id FROM asset_group_tags WHERE type = 3));

       -- Re-Migrate existing Tier Zero selectors
       WITH inserted_selector AS (
         INSERT INTO asset_group_tag_selectors (asset_group_tag_id, created_at, created_by, updated_at, updated_by, name, description, is_default, allow_disable, auto_certify)
         SELECT (SELECT id FROM asset_group_tags WHERE position = 1), current_timestamp, 'SYSTEM', current_timestamp, 'SYSTEM', s.name, s.selector, false, true, 2
         FROM asset_group_selectors s JOIN asset_groups ag ON ag.id = s.asset_group_id
         WHERE ag.tag = 'admin_tier_0' and NOT EXISTS(SELECT 1 FROM asset_group_tag_selectors WHERE name = s.name)
         RETURNING id, description
         )
       INSERT INTO asset_group_tag_selector_seeds (selector_id, type, value) SELECT id, 1, description FROM inserted_selector;

      -- Re-Migrate existing Owned selectors
      WITH inserted_selector AS (
        INSERT INTO asset_group_tag_selectors (asset_group_tag_id, created_at, created_by, updated_at, updated_by, name, description, is_default, allow_disable, auto_certify)
        SELECT (SELECT id FROM asset_group_tags WHERE type = 3), current_timestamp, 'SYSTEM', current_timestamp, 'SYSTEM', s.name, s.selector, false, true, 0
        FROM asset_group_selectors s JOIN asset_groups ag ON ag.id = s.asset_group_id
        WHERE ag.tag = 'owned' and NOT EXISTS(SELECT 1 FROM asset_group_tag_selectors WHERE name = s.name)
          RETURNING id, description
          )
      INSERT INTO asset_group_tag_selector_seeds (selector_id, type, value) SELECT id, 1, description FROM inserted_selector;
    END IF;
  END;
$$;
-- +goose StatementEnd

-- Set all default selectors to enabled for bootstrapped instances
UPDATE asset_group_tag_selectors SET disabled_at = NULL, disabled_by = NULL WHERE is_default = true AND created_at > current_timestamp - '7 min'::interval;

-- Set PZ feature flag enabled for bootstrapped instances
UPDATE feature_flags SET enabled = true WHERE key = 'tier_management_engine' AND created_at > current_timestamp - '7 min'::interval;

-- Add unique constraint for asset group tag selectors name per asset group tag
-- Before we add unique constraint, rename any duplicates with `_X` to prevent constraint failing
WITH duplicate_selectors AS (
  SELECT id, name, asset_group_tag_id, ROW_NUMBER() OVER (PARTITION BY name, asset_group_tag_id ORDER BY id) AS rowNumber
  FROM asset_group_tag_selectors
)
UPDATE asset_group_tag_selectors agts
SET name = agts.name || '_' || rowNumber FROM duplicate_selectors
WHERE agts.id = duplicate_selectors.id AND duplicate_selectors.rowNumber > 1;

ALTER TABLE IF EXISTS asset_group_tag_selectors DROP CONSTRAINT IF EXISTS asset_group_tag_selectors_unique_name_asset_group_tag;
ALTER TABLE IF EXISTS asset_group_tag_selectors ADD CONSTRAINT asset_group_tag_selectors_unique_name_asset_group_tag UNIQUE ("name",asset_group_tag_id,is_default);

-- Fix naming inconsistencies for ETAC
-- +goose StatementBegin
DO
$$
    BEGIN
        IF EXISTS (SELECT
                   FROM pg_tables
                   WHERE schemaname = 'public'
                     AND tablename = 'environment_access_control')
            AND NOT EXISTS (SELECT
                            FROM pg_tables
                            WHERE schemaname = 'public'
                              AND tablename = 'environment_targeted_access_control')
        THEN
            ALTER TABLE public.environment_access_control
                RENAME TO environment_targeted_access_control;
        END IF;
    END
$$;
-- +goose StatementEnd
UPDATE feature_flags
SET key         = 'environment_targeted_access_control',
    name        = 'Environment Targeted Access Control',
    description = 'Enable power users and admins to set environment targeted access controls on users'
WHERE key = 'targeted_access_control';

-- Update RO-DC default selector within Tier Zero to use the correct attribute name
UPDATE asset_group_tag_selector_seeds
SET value = E'MATCH (n:Computer)\nWHERE n.isreadonlydc = true\nRETURN n;'
WHERE selector_id in (SELECT id FROM asset_group_tag_selectors WHERE name = 'Read-Only DCs' AND is_default = true);

-- Set Open Graph Phase 2 feature flag to enable UI behind it
UPDATE feature_flags SET enabled = true WHERE key = 'open_graph_phase_2';

-- Add Analysis file retention defaults
INSERT INTO parameters (key, name, description, value, created_at, updated_at) VALUES ('analysis.retain_ingest_files', 'Analysis Retain Ingest Files', 'This config param sets the default beehavior of ingest file retention', '{"enabled": false}', current_timestamp, current_timestamp) ON CONFLICT DO NOTHING;
-- Copyright 2025 Specter Ops, Inc.
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
    ))
    OR
    (r.name = 'Administrator' AND (p.authority, p.name) IN (
               ('audit_log', 'Read')
    ))
)
ON CONFLICT DO NOTHING;

-- Configuring all existing records of "SYSTEM" to "BloodHound" within the asset_group_tags, asset_group_tag_selectors, asset_group_tag_selector_nodes, and asset_group_history tables
UPDATE asset_group_tags
SET created_by = CASE WHEN created_by = 'SYSTEM' THEN 'BloodHound' ELSE created_by END,
    updated_by = CASE WHEN updated_by = 'SYSTEM' THEN 'BloodHound' ELSE updated_by END,
    deleted_by = CASE WHEN deleted_by = 'SYSTEM' THEN 'BloodHound' ELSE deleted_by END
WHERE created_by = 'SYSTEM' OR updated_by = 'SYSTEM' OR deleted_by = 'SYSTEM';

UPDATE asset_group_tag_selectors
SET created_by = CASE WHEN created_by = 'SYSTEM' THEN 'BloodHound' ELSE created_by END,
    updated_by = CASE WHEN updated_by = 'SYSTEM' THEN 'BloodHound' ELSE updated_by END,
    disabled_by = CASE WHEN disabled_by = 'SYSTEM' THEN 'BloodHound' ELSE disabled_by END
WHERE created_by = 'SYSTEM' OR updated_by = 'SYSTEM' OR disabled_by = 'SYSTEM';

UPDATE asset_group_tag_selector_nodes
SET certified_by = 'BloodHound'
WHERE certified_by = 'SYSTEM';

UPDATE asset_group_history
SET actor = 'BloodHound'
WHERE actor = 'SYSTEM';


-- Explicitly set glyph values for the default asset_group_tags
-- Find Tier Zero by position
UPDATE asset_group_tags SET glyph = 'gem' WHERE position = 1;
-- Find Owned by type
UPDATE asset_group_tags SET glyph = 'skull' WHERE type = 3;
-- Copyright 2026 Specter Ops, Inc.
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

-- OpenGraph Search feature flag
INSERT INTO feature_flags (created_at, updated_at, key, name, description, enabled, user_updatable)
VALUES (current_timestamp,
    current_timestamp,
    'opengraph_search',
    'OpenGraph Search',
    'Enable OpenGraph Search',
    false,
    false)
ON CONFLICT DO NOTHING;


-- OpenGraph graph schema - extensions (collectors)
CREATE TABLE IF NOT EXISTS schema_extensions (
    id SERIAL NOT NULL,
    name TEXT UNIQUE NOT NULL,
    display_name TEXT NOT NULL,
    version TEXT NOT NULL,
    is_builtin BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT current_timestamp,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT current_timestamp,
    deleted_at TIMESTAMP WITH TIME ZONE DEFAULT NULL,
    PRIMARY KEY (id)
);

-- OpenGraph schema_node_kinds -  stores node kinds for open graph extensions. This FK's to the DAWGS kind table directly.
CREATE TABLE IF NOT EXISTS schema_node_kinds (
    id SERIAL PRIMARY KEY,
    schema_extension_id INT NOT NULL REFERENCES schema_extensions (id) ON DELETE CASCADE, -- indicates which extension this node kind belongs to
    kind_id SMALLINT NOT NULL UNIQUE REFERENCES kind (id) ON DELETE CASCADE,
    display_name TEXT NOT NULL, -- can be different from name but usually isn't other than Base/Entity
    description TEXT NOT NULL, -- human-readable description of the kind
    is_display_kind BOOL NOT NULL DEFAULT FALSE,
    icon TEXT NOT NULL, -- font-awesome icon
    icon_color TEXT NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT current_timestamp,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT current_timestamp,
    deleted_at TIMESTAMP WITH TIME ZONE DEFAULT NULL
);

CREATE INDEX IF NOT EXISTS idx_graph_schema_node_kinds_extensions_id ON schema_node_kinds (schema_extension_id);

-- OpenGraph schema properties
CREATE TABLE IF NOT EXISTS schema_properties (
    id SERIAL NOT NULL,
    schema_extension_id INT NOT NULL,
    name TEXT NOT NULL,
    display_name TEXT NOT NULL,
    data_type TEXT NOT NULL,
    description TEXT NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT current_timestamp,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT current_timestamp,
    deleted_at TIMESTAMP WITH TIME ZONE DEFAULT NULL,
    CONSTRAINT fk_schema_extensions_schema_properties FOREIGN KEY (schema_extension_id) REFERENCES schema_extensions(id) ON DELETE CASCADE,
    UNIQUE (schema_extension_id, name)
);

CREATE INDEX IF NOT EXISTS idx_schema_properties_schema_extensions_id on schema_properties (schema_extension_id);

-- OpenGraph schema_edge_kinds - store edge kinds for open graph extensions. This FK's to the DAWGS kind table directly.
-- Renamed to schema_relationship_kinds
CREATE TABLE IF NOT EXISTS schema_edge_kinds (
    id SERIAL PRIMARY KEY,
    schema_extension_id INT NOT NULL REFERENCES schema_extensions (id) ON DELETE CASCADE, -- indicates which extension this edge kind belongs to
    kind_id SMALLINT NOT NULL UNIQUE REFERENCES kind (id) ON DELETE CASCADE,
    description TEXT NOT NULL, -- human-readable description of the edge-kind
    is_traversable BOOL NOT NULL DEFAULT FALSE, -- indicates whether the given edge-kind is traversable
    created_at TIMESTAMP WITH TIME ZONE DEFAULT current_timestamp,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT current_timestamp,
    deleted_at TIMESTAMP WITH TIME ZONE DEFAULT NULL
);

CREATE INDEX IF NOT EXISTS idx_schema_edge_kinds_extensions_id ON schema_edge_kinds (schema_extension_id);

-- OpenGraph schema_environments - stores environment mappings.
CREATE TABLE IF NOT EXISTS schema_environments (
    id SERIAL,
    schema_extension_id INTEGER NOT NULL REFERENCES schema_extensions(id) ON DELETE CASCADE,
    environment_kind_id INTEGER NOT NULL REFERENCES kind(id),
    source_kind_id INTEGER NOT NULL REFERENCES kind(id),
    PRIMARY KEY (id),
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT current_timestamp,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT current_timestamp,
    deleted_at TIMESTAMP WITH TIME ZONE DEFAULT NULL,
    UNIQUE(environment_kind_id,source_kind_id)
);

CREATE INDEX IF NOT EXISTS idx_schema_environments_extension_id ON schema_environments (schema_extension_id);

-- OpenGraph schema_relationship_findings - Individual findings. ie T0WriteOwner, T0ADCSESC1, T0DCSync
CREATE TABLE IF NOT EXISTS schema_relationship_findings (
    id SERIAL,
    schema_extension_id INTEGER NOT NULL REFERENCES schema_extensions(id) ON DELETE CASCADE,
    relationship_kind_id INTEGER NOT NULL REFERENCES kind(id),
    environment_id INTEGER NOT NULL REFERENCES schema_environments(id) ON DELETE CASCADE,
    name TEXT NOT NULL,
    display_name TEXT NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT current_timestamp,
    PRIMARY KEY(id),
    UNIQUE(name)
);

CREATE INDEX IF NOT EXISTS idx_schema_relationship_findings_extension_id ON schema_relationship_findings (schema_extension_id);
CREATE INDEX IF NOT EXISTS idx_schema_relationship_findings_environment_id ON schema_relationship_findings(environment_id);

-- OpenGraph remediation_content_type - ENUM type for remediation content categories, IF NOT EXISTS is not natively supported for type creation
-- +goose StatementBegin
DO $$
    BEGIN
        IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'remediation_content_type') THEN
            CREATE TYPE  remediation_content_type AS ENUM (
                'short_description',
                'long_description',
                'short_remediation',
                'long_remediation'
                );
        END IF;
    END
$$;
-- +goose StatementEnd

-- OpenGraph schema_remediations - Normalized remediation content table with FK to findings
CREATE TABLE IF NOT EXISTS schema_remediations (
    finding_id INTEGER NOT NULL REFERENCES schema_relationship_findings(id) ON DELETE CASCADE,
    content_type remediation_content_type NOT NULL,
    content TEXT STORAGE MAIN,
    PRIMARY KEY(finding_id, content_type)
);

-- Index for filtering by content_type (single content type queries)
CREATE INDEX IF NOT EXISTS idx_schema_remediations_content_type ON schema_remediations(content_type);

-- OpenGraph schema_environments_principal_kinds - Environment to principal mappings
CREATE TABLE IF NOT EXISTS schema_environments_principal_kinds (
    environment_id INTEGER NOT NULL REFERENCES schema_environments(id) ON DELETE CASCADE,
    principal_kind INTEGER NOT NULL REFERENCES kind(id),
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT current_timestamp,
    PRIMARY KEY(environment_id, principal_kind)
);

CREATE INDEX IF NOT EXISTS idx_schema_environments_principal_kinds_principal_kind ON schema_environments_principal_kinds (principal_kind);

-- Added to report warnings for opengraph files that attempt to create invalid relationships.
ALTER TABLE ingest_jobs
    ADD COLUMN IF NOT EXISTS partial_failed_files integer DEFAULT 0;

ALTER TABLE completed_tasks
    ADD COLUMN IF NOT EXISTS warnings TEXT[] NOT NULL DEFAULT '{}';

ALTER TABLE source_kinds
    ADD COLUMN IF NOT EXISTS active BOOLEAN DEFAULT true NOT NULL;

UPDATE source_kinds SET active = true WHERE active is NULL;

-- Enables Citrix RDP support by default
UPDATE parameters
SET
    value = '{ "enabled": true }'
WHERE key = 'analysis.citrix_rdp_support';

-- Drop old ETAC table if the old and new table exist due to a failed v8.3.0 migration
-- +goose StatementBegin
DO
$$
    BEGIN
        IF EXISTS (SELECT
                   FROM pg_tables
                   WHERE schemaname = 'public'
                     AND tablename = 'environment_targeted_access_control')
        THEN
            DROP TABLE IF EXISTS environment_access_control;
        END IF;
    END
$$;
-- +goose StatementEnd

-- Feature flag for Client Bearer Token Auth
INSERT INTO feature_flags (created_at, updated_at, key, name, description, enabled, user_updatable)
VALUES (current_timestamp,
        current_timestamp,
        'client_bearer_auth',
        'Client Bearer Auth',
        'Enable clients to be authenticated using bearer tokens.',
        false,
        false)
  ON CONFLICT DO NOTHING;

 -- Add AGT tuning parameter
INSERT INTO parameters (key, name, description, value, created_at, updated_at)
VALUES ('analysis.tagging', 'Analysis Tagging Configuration', 'This configuration parameter determines the limits used during the asset group tagging phase of analysis', '{"dawgs_worker_limit": 2, "expansion_worker_limit": 3, "selector_worker_limit": 7}', current_timestamp, current_timestamp)
ON CONFLICT DO NOTHING;

-- upsert_kind checks to see if a kind exists in the kind table and inserts it if not.
-- A SELECT is used instead of an insert CTE with ON CONDITION DO as the latter will increment the kind's SERIAL id even
-- if the kind already exists.
-- +goose StatementBegin
CREATE OR REPLACE FUNCTION upsert_kind(node_kind_name TEXT) RETURNS kind AS $$
DECLARE
    kind_row kind%rowtype;
BEGIN
    LOCK kind;

    SELECT * INTO kind_row FROM kind WHERE kind.name = node_kind_name;

    IF kind_row.id IS NULL THEN
        INSERT INTO kind (name) VALUES (node_kind_name) RETURNING * INTO kind_row;
    END IF;

    RETURN kind_row;
END $$ LANGUAGE plpgsql;
-- +goose StatementEnd
-- Copyright 2026 Specter Ops, Inc.
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
-- OpenGraph Pathfinding feature flag
INSERT INTO feature_flags (created_at, updated_at, key, name, description, enabled, user_updatable)
VALUES (current_timestamp,
    current_timestamp,
    'opengraph_pathfinding',
    'OpenGraph Pathfinding',
    'Enable OpenGraph Pathfinding',
    false,
    false)
ON CONFLICT DO NOTHING;

-- Set `opengraph_search` feature flag to enabled by default
UPDATE feature_flags SET enabled = true WHERE key = 'opengraph_search';
-- Copyright 2026 Specter Ops, Inc.
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

ALTER TABLE IF EXISTS schema_environments
    DROP CONSTRAINT IF EXISTS schema_environments_source_kind_id_fkey;

ALTER TABLE IF EXISTS schema_environments
    ADD CONSTRAINT schema_environments_source_kind_id_fkey
    FOREIGN KEY (source_kind_id) REFERENCES source_kinds(id);


-- OpenGraph Findings feature flag
INSERT INTO feature_flags (created_at, updated_at, key, name, description, enabled, user_updatable)
VALUES (current_timestamp,
    current_timestamp,
    'opengraph_findings',
    'OpenGraph Findings',
    'Enable OpenGraph Findings',
    false,
    false)
ON CONFLICT DO NOTHING;

-- Add API Tokens parameter
INSERT INTO parameters (key, name, description, value, created_at, updated_at)
VALUES ('auth.api_tokens',
        'API Tokens',
        'This configuration parameter enables/disables authorization through API Tokens',
        '{"enabled":true}',
        current_timestamp,
        current_timestamp)
  ON CONFLICT DO NOTHING;

-- Add Timeouts parameter
INSERT INTO parameters (key, name, description, value, created_at, updated_at)
VALUES ('api.timeout_limit',
        'Query Timeout Limit',
        'This configuration parameter enables/disables a timeout limit for API Requests',
        '{"enabled":true}',
        current_timestamp,
        current_timestamp)
  ON CONFLICT DO NOTHING;

-- Update Scheduled Analysis description
UPDATE parameters SET description = 'This configuration parameter allows setting a schedule for analysis. When enabled, analysis will run when the scheduled time arrives or when manually requested' WHERE key = 'analysis.scheduled';

-- Add Namespace column to schema_extensions
ALTER TABLE schema_extensions
    ADD COLUMN IF NOT EXISTS namespace TEXT;

UPDATE schema_extensions SET namespace = LEFT(name, 3)
WHERE namespace IS NULL OR namespace = '';

ALTER TABLE schema_extensions
    ALTER COLUMN namespace SET NOT NULL;

-- +goose StatementBegin
DO $$
    BEGIN
        IF NOT EXISTS (
                      SELECT 1
                      FROM pg_constraint
                      WHERE conname = 'schema_extensions_namespace_unique'
        ) THEN
            ALTER TABLE schema_extensions
                ADD CONSTRAINT schema_extensions_namespace_unique UNIQUE (namespace);
        END IF;
    END$$;
-- +goose StatementEnd

ALTER TABLE IF EXISTS schema_edge_kinds RENAME TO schema_relationship_kinds;

-- Remove ETAC from feature flags since it has moved to DogTags
DELETE FROM feature_flags WHERE key = 'environment_targeted_access_control';

-- Drop unique name constraint before migrating to PZ in case AGT names are not unique
ALTER TABLE IF EXISTS asset_group_tag_selectors DROP CONSTRAINT IF EXISTS asset_group_tag_selectors_unique_name_asset_group_tag;

-- Remigrate old custom AGI selectors to PZ selectors for any instances without PZ feature flag enabled
-- +goose StatementBegin
DO $$
  BEGIN
		IF
      (SELECT enabled FROM feature_flags WHERE key  = 'tier_management_engine') = false
    THEN
       -- Delete custom selectors
       DELETE FROM asset_group_tag_selectors WHERE is_default = false AND asset_group_tag_id IN ((SELECT id FROM asset_group_tags WHERE position = 1), (SELECT id FROM asset_group_tags WHERE type = 3));

       -- Re-Migrate existing Tier Zero selectors
       WITH inserted_selector AS (
         INSERT INTO asset_group_tag_selectors (asset_group_tag_id, created_at, created_by, updated_at, updated_by, name, description, is_default, allow_disable, auto_certify)
         SELECT (SELECT id FROM asset_group_tags WHERE position = 1), current_timestamp, 'BloodHound', current_timestamp, 'BloodHound', s.name, s.selector, false, true, 2
         FROM asset_group_selectors s JOIN asset_groups ag ON ag.id = s.asset_group_id
         WHERE ag.tag = 'admin_tier_0' and NOT EXISTS(SELECT 1 FROM asset_group_tag_selectors WHERE name = s.name)
         RETURNING id, description
         )
       INSERT INTO asset_group_tag_selector_seeds (selector_id, type, value) SELECT id, 1, description FROM inserted_selector;

      -- Re-Migrate existing Owned selectors
      WITH inserted_selector AS (
        INSERT INTO asset_group_tag_selectors (asset_group_tag_id, created_at, created_by, updated_at, updated_by, name, description, is_default, allow_disable, auto_certify)
        SELECT (SELECT id FROM asset_group_tags WHERE type = 3), current_timestamp, 'BloodHound', current_timestamp, 'BloodHound', s.name, s.selector, false, true, 0
        FROM asset_group_selectors s JOIN asset_groups ag ON ag.id = s.asset_group_id
        WHERE ag.tag = 'owned' and NOT EXISTS(SELECT 1 FROM asset_group_tag_selectors WHERE name = s.name)
          RETURNING id, description
          )
      INSERT INTO asset_group_tag_selector_seeds (selector_id, type, value) SELECT id, 1, description FROM inserted_selector;
    END IF;
  END;
$$;
-- +goose StatementEnd

-- Before we add unique constraint, rename any duplicates with `_X` to prevent constraint failing
WITH duplicate_selectors AS (
  SELECT id, name, asset_group_tag_id, ROW_NUMBER() OVER (PARTITION BY name, asset_group_tag_id ORDER BY id) AS rowNumber
  FROM asset_group_tag_selectors
)
UPDATE asset_group_tag_selectors agts
SET name = agts.name || '_' || rowNumber FROM duplicate_selectors
WHERE agts.id = duplicate_selectors.id AND duplicate_selectors.rowNumber > 1;

-- Reinstate unique constraint for asset group tag selectors name per asset group tag
ALTER TABLE IF EXISTS asset_group_tag_selectors ADD CONSTRAINT asset_group_tag_selectors_unique_name_asset_group_tag UNIQUE ("name",asset_group_tag_id,is_default);-- Copyright 2026 Specter Ops, Inc.
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

-- OpenGraph Extension Management feature flag
INSERT INTO feature_flags (created_at, updated_at, key, name, description, enabled, user_updatable)
VALUES (current_timestamp,
        current_timestamp,
        'opengraph_extension_management',
        'OpenGraph Extension Management',
        'Enable OpenGraph Extension Management',
        false,
        false)
ON CONFLICT DO NOTHING;

-- Scheduled Analysis Configuration feature flag
INSERT INTO feature_flags (created_at, updated_at, key, name, description, enabled, user_updatable)
VALUES (current_timestamp,
        current_timestamp,
        'scheduled_analysis_configuration',
        'Scheduled Analysis Configuration',
        'Enable Scheduled Analysis Configuration form in the UI',
        false,
        false)
ON CONFLICT DO NOTHING;

-- OpenGraph Collector Platform Support feature flag
INSERT INTO feature_flags (created_at, updated_at, key, name, description, enabled, user_updatable)
VALUES (current_timestamp,
        current_timestamp,
        'opengraph_collector_platform_support',
        'OpenGraph Collector Platform Support',
        'Enable creation and communication with the OpenGraph Collector platform.',
        false,
        false)
ON CONFLICT DO NOTHING;

-- Remove pathfinding feature flag. We are keying off of opengraph_extension_management instead
DELETE FROM feature_flags WHERE key = 'opengraph_pathfinding';

-- upsert_kind is a stop-gap function used in open graph schema node,
-- relationship and source kind creation.
-- This function is needed to stave off kind table id exhaustion while maintaining
-- an acceptable level of performance. An exception will be raised if
-- the function fails to insert the kind after 5 attempts.
--
-- Underlying Issues: The kind table's id column is a SMALLINT, Postgres will
-- increase a SERIAL PK on conflict even if DO NOTHING or DO UPDATE is used and
-- a table lock on the Kind table greatly decreases performance.
-- +goose StatementBegin
CREATE OR REPLACE FUNCTION upsert_kind(node_kind_name TEXT) RETURNS kind AS $$
DECLARE
    kind_row kind%rowtype;
BEGIN
    -- Try to find existing kind based on name
    SELECT * INTO kind_row FROM kind WHERE name = node_kind_name;
    IF kind_row IS NOT NULL THEN
        RETURN kind_row;
    END IF;
    -- Insert with retry, handles the race condition where two transactions try to add the same kind at the same time
    FOR i IN 1..5 LOOP
        BEGIN
            INSERT INTO kind (name)
            VALUES (node_kind_name)
            RETURNING * INTO kind_row;
            RETURN kind_row;
        EXCEPTION
            WHEN unique_violation THEN
                -- Check if the insert conflict was for same kind
                SELECT * INTO kind_row FROM kind WHERE name = node_kind_name;
                IF kind_row IS NOT NULL THEN
                    RETURN kind_row;
                END IF;
        END;
    END LOOP;
    -- failed to insert kind after 5 retries
    RAISE EXCEPTION 'failed to insert kind % after 5 retries', node_kind_name;
END;
$$ LANGUAGE plpgsql;
-- +goose StatementEnd

-- Add FK from source_kind to kind:
-- add kind_id column, add any missing source_kinds to the kind table,
-- fill new kind_id column, add not null and unique constraint,
-- add fk to kind table, drop name column
ALTER TABLE source_kinds ADD COLUMN IF NOT EXISTS kind_id SMALLINT;
-- +goose StatementBegin
DO $$
    BEGIN
        IF EXISTS (
            SELECT 1
            FROM information_schema.columns
            WHERE table_schema = 'public' AND column_name = 'name' AND table_name = 'source_kinds'
        ) THEN
            INSERT INTO kind (name)
            SELECT name
            FROM source_kinds sk
            WHERE NOT EXISTS (
                             SELECT 1
                             FROM kind k
                             WHERE k.name = sk.name) ON CONFLICT DO NOTHING;
            UPDATE source_kinds sk SET kind_id = k.id FROM kind k WHERE sk.name = k.name;
        END IF;
        ALTER TABLE source_kinds ALTER COLUMN kind_id SET NOT NULL;
        IF NOT EXISTS (
                      SELECT 1
                      FROM pg_constraint
                      WHERE conname = 'source_kinds_kind_id_unique'
        ) THEN
            ALTER TABLE source_kinds
                ADD CONSTRAINT source_kinds_kind_id_unique UNIQUE (kind_id);
        END IF;
        IF NOT EXISTS (
                      SELECT 1
                      FROM pg_constraint
                      WHERE conname = 'fk_source_kinds_kind_id_kind'
        ) THEN
            ALTER TABLE source_kinds
                ADD CONSTRAINT fk_source_kinds_kind_id_kind FOREIGN KEY (kind_id) REFERENCES kind (id) ON DELETE CASCADE;
        END IF;
    END
$$;
-- +goose StatementEnd
ALTER TABLE source_kinds DROP COLUMN IF EXISTS name;

-- OpenGraph schema_list_findings
CREATE TABLE IF NOT EXISTS schema_list_findings (
    id SERIAL,
    schema_extension_id INTEGER NOT NULL REFERENCES schema_extensions(id) ON DELETE CASCADE,
    node_kind_id SMALLINT NOT NULL REFERENCES kind(id),
    environment_id INTEGER NOT NULL REFERENCES schema_environments(id) ON DELETE CASCADE,
    name TEXT NOT NULL,
    display_name TEXT NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT current_timestamp,
    PRIMARY KEY(id),
    UNIQUE(name)
);

CREATE INDEX IF NOT EXISTS idx_schema_list_findings_extension_id ON schema_list_findings (schema_extension_id);
CREATE INDEX IF NOT EXISTS idx_schema_list_findings_environment_id ON schema_list_findings(environment_id);

-- Drop unique name constraint before migrating to PZ in case AGT names are not unique
ALTER TABLE IF EXISTS asset_group_tag_selectors DROP CONSTRAINT IF EXISTS asset_group_tag_selectors_unique_name_asset_group_tag;

-- Remigrate old custom AGI selectors to PZ selectors for any instances without PZ feature flag enabled
-- +goose StatementBegin
DO $$
  BEGIN
		IF
      (SELECT enabled FROM feature_flags WHERE key  = 'tier_management_engine') = false
    THEN
       -- Delete custom selectors
       DELETE FROM asset_group_tag_selectors WHERE is_default = false AND asset_group_tag_id IN ((SELECT id FROM asset_group_tags WHERE position = 1), (SELECT id FROM asset_group_tags WHERE type = 3));

       -- Re-Migrate existing Tier Zero selectors
       WITH inserted_selector AS (
         INSERT INTO asset_group_tag_selectors (asset_group_tag_id, created_at, created_by, updated_at, updated_by, name, description, is_default, allow_disable, auto_certify)
         SELECT (SELECT id FROM asset_group_tags WHERE position = 1), current_timestamp, 'BloodHound', current_timestamp, 'BloodHound', s.name, s.selector, false, true, 2
         FROM asset_group_selectors s JOIN asset_groups ag ON ag.id = s.asset_group_id
         WHERE ag.tag = 'admin_tier_0' and NOT EXISTS(SELECT 1 FROM asset_group_tag_selectors WHERE name = s.name)
         RETURNING id, description
         )
       INSERT INTO asset_group_tag_selector_seeds (selector_id, type, value) SELECT id, 1, description FROM inserted_selector;

      -- Re-Migrate existing Owned selectors
      WITH inserted_selector AS (
        INSERT INTO asset_group_tag_selectors (asset_group_tag_id, created_at, created_by, updated_at, updated_by, name, description, is_default, allow_disable, auto_certify)
        SELECT (SELECT id FROM asset_group_tags WHERE type = 3), current_timestamp, 'BloodHound', current_timestamp, 'BloodHound', s.name, s.selector, false, true, 0
        FROM asset_group_selectors s JOIN asset_groups ag ON ag.id = s.asset_group_id
        WHERE ag.tag = 'owned' and NOT EXISTS(SELECT 1 FROM asset_group_tag_selectors WHERE name = s.name)
          RETURNING id, description
          )
      INSERT INTO asset_group_tag_selector_seeds (selector_id, type, value) SELECT id, 1, description FROM inserted_selector;
    END IF;
  END;
$$;
-- +goose StatementEnd

-- Before we add unique constraint, rename any duplicates with `_X` to prevent constraint failing
WITH duplicate_selectors AS (
  SELECT id, name, asset_group_tag_id, ROW_NUMBER() OVER (PARTITION BY name, asset_group_tag_id ORDER BY id) AS rowNumber
  FROM asset_group_tag_selectors
)
UPDATE asset_group_tag_selectors agts
SET name = agts.name || '_' || rowNumber FROM duplicate_selectors
WHERE agts.id = duplicate_selectors.id AND duplicate_selectors.rowNumber > 1;

-- Reinstate unique constraint for asset group tag selectors name per asset group tag
ALTER TABLE IF EXISTS asset_group_tag_selectors ADD CONSTRAINT asset_group_tag_selectors_unique_name_asset_group_tag UNIQUE ("name",asset_group_tag_id,is_default);
-- Copyright 2026 Specter Ops, Inc.
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
-- Drop the compound unique constraint on schema_environments (environment_kind_id, source_kind_id)
-- and add a unique constraint on just environment_kind_id
ALTER TABLE IF EXISTS schema_environments
    DROP CONSTRAINT IF EXISTS schema_environments_environment_kind_id_source_kind_id_key;

-- +goose StatementBegin
DO $$
    BEGIN
        IF NOT EXISTS (
                      SELECT 1
                      FROM pg_constraint
                      WHERE conname = 'schema_environments_environment_kind_id_key'
        ) THEN
            ALTER TABLE schema_environments
                ADD CONSTRAINT schema_environments_environment_kind_id_key UNIQUE (environment_kind_id);
        END IF;
    END$$;
-- +goose StatementEnd

-- Drop list findings table, it was unused;
DROP TABLE IF EXISTS schema_list_findings;

-- Create schema_findings table
CREATE TABLE IF NOT EXISTS schema_findings (
  id SERIAL,
  type INTEGER NOT NULL,
  schema_extension_id INTEGER NOT NULL REFERENCES schema_extensions(id) ON DELETE CASCADE,
  environment_id INTEGER NOT NULL REFERENCES schema_environments(id) ON DELETE CASCADE,
  kind_id INTEGER NOT NULL REFERENCES kind(id),
  name TEXT NOT NULL,
  display_name TEXT NOT NULL,
  created_at TIMESTAMP WITH TIME ZONE DEFAULT current_timestamp,
  PRIMARY KEY(id),
  UNIQUE(name)
);
CREATE INDEX IF NOT EXISTS idx_schema_findings_extension_id ON schema_findings (schema_extension_id);
CREATE INDEX IF NOT EXISTS idx_schema_findings_environment_id ON schema_findings(environment_id);

-- Populate schema_findings from old schema_relationship_findings
-- +goose StatementBegin
DO $$
BEGIN
  IF EXISTS (SELECT FROM information_schema.tables WHERE table_name = 'schema_relationship_findings') THEN
    INSERT INTO schema_findings (id, type, schema_extension_id, kind_id, environment_id, name, display_name, created_at)
    SELECT id, 1, schema_extension_id, relationship_kind_id, environment_id, name, display_name, created_at
    FROM schema_relationship_findings ON CONFLICT (name) DO NOTHING;
  END IF;
END
$$;
-- +goose StatementEnd

-- Update schema_remediations to reference schema_findings instead of schema_relationship_findings
ALTER TABLE schema_remediations DROP CONSTRAINT IF EXISTS schema_remediations_finding_id_fkey;
ALTER TABLE schema_remediations ADD CONSTRAINT schema_remediations_finding_id_fkey FOREIGN KEY (finding_id) REFERENCES schema_findings(id) ON DELETE CASCADE;

-- Drop schema_relationship_findings
DROP TABLE IF EXISTS schema_relationship_findings;

-- Add subtypes table to schema_findings
CREATE TABLE IF NOT EXISTS schema_findings_subtypes (
  schema_finding_id INTEGER NOT NULL REFERENCES schema_findings(id) ON DELETE CASCADE,
  subtype TEXT NOT NULL,
  PRIMARY KEY(schema_finding_id, subtype)
);

-- Update the 'auth_tokens' table adding expiration column
ALTER TABLE auth_tokens
ADD COLUMN IF NOT EXISTS expires_at timestamp with time zone;

-- Add a column to `custom_node_kinds` to more easily correlate OpenGraph icons
ALTER TABLE IF EXISTS custom_node_kinds
    ADD COLUMN IF NOT EXISTS schema_node_kind_id INTEGER REFERENCES schema_node_kinds (id) ON DELETE SET NULL;

-- Add Posture PDF Export feature flag
INSERT INTO feature_flags (created_at, updated_at, key, name, description, enabled, user_updatable)
VALUES (current_timestamp,
        current_timestamp,
        'posture_pdf_export',
        'Posture PDF Export',
        'Enables PDF export from Posture page.',
        false,
        false)
ON CONFLICT DO NOTHING;

-- Backfill the custom_node_kinds table with any missing icon definitions from the schema_node_kinds table
-- +goose StatementBegin
DO $$
DECLARE schema_node_kind_record RECORD;
BEGIN
  FOR schema_node_kind_record IN
    SELECT
      schema_node_kinds.id,
      kind.name AS kind_name,
      schema_node_kinds.icon,
      schema_node_kinds.icon_color
    FROM schema_node_kinds
    JOIN kind ON schema_node_kinds.kind_id = kind.id
    JOIN schema_extensions ON schema_node_kinds.schema_extension_id = schema_extensions.id
    WHERE schema_node_kinds.icon IS NOT NULL
      AND schema_node_kinds.icon != ''
      AND schema_node_kinds.is_display_kind = true
      AND schema_node_kinds.deleted_at IS NULL
      AND schema_extensions.is_builtin = false
  LOOP
    IF NOT EXISTS (SELECT 1
      FROM custom_node_kinds
      WHERE schema_node_kind_id = schema_node_kind_record.id) THEN
        IF NOT EXISTS (SELECT 1
          FROM custom_node_kinds
          WHERE kind_name = schema_node_kind_record.kind_name) THEN
            INSERT INTO custom_node_kinds (kind_name, schema_node_kind_id, config)
            VALUES (schema_node_kind_record.kind_name, schema_node_kind_record.id, jsonb_build_object('icon', jsonb_build_object('type', 'font-awesome', 'name', schema_node_kind_record.icon, 'color', schema_node_kind_record.icon_color)));
        ELSE
          UPDATE custom_node_kinds SET schema_node_kind_id = schema_node_kind_record.id, config = jsonb_build_object('icon', jsonb_build_object('type', 'font-awesome', 'name', schema_node_kind_record.icon, 'color', schema_node_kind_record.icon_color)), updated_at = NOW() WHERE kind_name = schema_node_kind_record.kind_name;
        END IF;
    END IF;
  END LOOP;
END
$$;
-- +goose StatementEnd

-- The migrations below must occur before toggling PZ Feature Flag GA

-- Drop unique name constraint before migrating to PZ in case AGT names are not unique
ALTER TABLE IF EXISTS asset_group_tag_selectors DROP CONSTRAINT IF EXISTS asset_group_tag_selectors_unique_name_asset_group_tag;

-- Remigrate old custom AGI selectors to PZ selectors for any instances without PZ feature flag enabled
-- +goose StatementBegin
DO $$
BEGIN
		IF
(SELECT enabled FROM feature_flags WHERE key  = 'tier_management_engine') = false
  THEN
-- Delete custom selectors
DELETE FROM asset_group_tag_selectors WHERE is_default = false AND asset_group_tag_id IN ((SELECT id FROM asset_group_tags WHERE position = 1), (SELECT id FROM asset_group_tags WHERE type = 3));

-- Re-Migrate existing Tier Zero selectors
WITH inserted_selector AS (
INSERT INTO asset_group_tag_selectors (asset_group_tag_id, created_at, created_by, updated_at, updated_by, name, description, is_default, allow_disable, auto_certify)
SELECT (SELECT id FROM asset_group_tags WHERE position = 1), current_timestamp, 'BloodHound', current_timestamp, 'BloodHound', s.name, s.selector, false, true, 2
FROM asset_group_selectors s JOIN asset_groups ag ON ag.id = s.asset_group_id
WHERE ag.tag = 'admin_tier_0' and NOT EXISTS(SELECT 1 FROM asset_group_tag_selectors WHERE name = s.name)
  RETURNING id, description
         )
INSERT INTO asset_group_tag_selector_seeds (selector_id, type, value) SELECT id, 1, description FROM inserted_selector;

-- Re-Migrate existing Owned selectors
WITH inserted_selector AS (
INSERT INTO asset_group_tag_selectors (asset_group_tag_id, created_at, created_by, updated_at, updated_by, name, description, is_default, allow_disable, auto_certify)
SELECT (SELECT id FROM asset_group_tags WHERE type = 3), current_timestamp, 'BloodHound', current_timestamp, 'BloodHound', s.name, s.selector, false, true, 0
FROM asset_group_selectors s JOIN asset_groups ag ON ag.id = s.asset_group_id
WHERE ag.tag = 'owned' and NOT EXISTS(SELECT 1 FROM asset_group_tag_selectors WHERE name = s.name)
  RETURNING id, description
          )
INSERT INTO asset_group_tag_selector_seeds (selector_id, type, value) SELECT id, 1, description FROM inserted_selector;
END IF;
END;
$$;
-- +goose StatementEnd

-- Before we add unique constraint, rename any duplicates with `_X` to prevent constraint failing
WITH duplicate_selectors AS (
  SELECT id, name, asset_group_tag_id, ROW_NUMBER() OVER (PARTITION BY name, asset_group_tag_id ORDER BY id) AS rowNumber
  FROM asset_group_tag_selectors
)
UPDATE asset_group_tag_selectors agts
SET name = agts.name || '_' || rowNumber FROM duplicate_selectors
WHERE agts.id = duplicate_selectors.id AND duplicate_selectors.rowNumber > 1;

-- Reinstate unique constraint for asset group tag selectors name per asset group tag
ALTER TABLE IF EXISTS asset_group_tag_selectors ADD CONSTRAINT asset_group_tag_selectors_unique_name_asset_group_tag UNIQUE ("name",asset_group_tag_id,is_default);

-- GA Tier Management Engine (PZ)
UPDATE feature_flags SET enabled = true, user_updatable = false, updated_at = current_timestamp WHERE key = 'tier_management_engine';


-- Swap source_kind to INTEGER for id
ALTER TABLE source_kinds ALTER COLUMN id SET DATA TYPE INTEGER;
ALTER SEQUENCE source_kinds_id_seq AS INTEGER;
SELECT setval('source_kinds_id_seq', (SELECT MAX(id) FROM source_kinds), true);

-- Update upsert_kind function to use advisory lock
-- +goose StatementBegin
CREATE OR REPLACE FUNCTION upsert_kind(node_kind_name text) RETURNS kind AS $$
DECLARE
  kind_row kind%rowtype;
BEGIN
    -- This avoids full-table locking, serializing calls using the same kind name, thus preventing duplicate inserts for the same kind
    PERFORM pg_advisory_xact_lock(hashtext(node_kind_name));

    SELECT * INTO kind_row FROM kind WHERE name = node_kind_name;

    IF kind_row.id IS NULL THEN
        INSERT INTO kind (name) VALUES (node_kind_name) RETURNING * INTO kind_row;
    END IF;

    RETURN kind_row;
END $$
LANGUAGE plpgsql;
-- +goose StatementEnd

-- +goose StatementBegin
CREATE OR REPLACE FUNCTION upsert_source_kind(source_kind_name TEXT) RETURNS source_kinds AS $$
DECLARE
  kind_row kind%rowtype;
  source_kind_row source_kinds%rowtype;
BEGIN
    -- Use advisory lock to serialize calls with the same source kind name
    PERFORM pg_advisory_xact_lock(hashtext(source_kind_name));

    SELECT * INTO kind_row FROM upsert_kind(source_kind_name);

    -- Then, try to find existing source_kind by kind_id
    SELECT * INTO source_kind_row FROM source_kinds WHERE kind_id = kind_row.id;

    IF source_kind_row.id IS NULL THEN
        INSERT INTO source_kinds (kind_id, active) VALUES (kind_row.id, true) RETURNING * INTO source_kind_row;
    ELSE
      UPDATE source_kinds SET active = true WHERE id = source_kind_row.id RETURNING * INTO source_kind_row;
    END IF;

RETURN source_kind_row;
END $$
LANGUAGE plpgsql;
-- +goose StatementEnd
-- Copyright 2026 Specter Ops, Inc.
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

-- Fix schema_findings_pkey duplicate key value violation
SELECT setval('schema_findings_id_seq', COALESCE(MAX(id), 1), true) FROM schema_findings;-- Copyright 2026 Specter Ops, Inc.
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


-- Add support_account flag to users
ALTER TABLE users ADD COLUMN IF NOT EXISTS support_account BOOL DEFAULT FALSE;

-- Rename opengraph_collector_platform_support feature flag to openhound_support
UPDATE feature_flags
SET key         = 'openhound_support',
    name        = 'OpenHound Support',
    description = 'Enable creation and communication with OpenHound platform'
WHERE key = 'opengraph_collector_platform_support';

-- Create Read Jobs permissions
INSERT INTO permissions (authority, name, created_at, updated_at)
VALUES ('collection', 'ReadJobs', current_timestamp, current_timestamp)
  ON CONFLICT DO NOTHING;

-- Add CollectionReadJobs permission to Administrator and Auditor
INSERT INTO roles_permissions (role_id, permission_id)
SELECT r.id, p.id
FROM roles r
JOIN permissions p ON (p.authority, p.name) = ('collection', 'ReadJobs')
WHERE r.name IN ('Auditor', 'Administrator')
  ON CONFLICT DO NOTHING;

-- Remove unused permission
DELETE FROM roles_permissions
WHERE permission_id = (SELECT id FROM permissions WHERE authority='auth' and name = 'ManageAppConfig');

DELETE FROM permissions
WHERE authority='auth' and name = 'ManageAppConfig';


-- +goose StatementBegin
DO $$
  BEGIN
    IF EXISTS (
      SELECT 1
      FROM information_schema.table_constraints tc
          JOIN information_schema.constraint_column_usage ccu ON tc.constraint_name = ccu.constraint_name
      WHERE tc.table_name = 'schema_environments'
        AND tc.constraint_name = 'schema_environments_source_kind_id_fkey'
        AND ccu.table_name = 'source_kinds'
    )
   THEN
      ALTER TABLE schema_environments DROP CONSTRAINT schema_environments_source_kind_id_fkey;

      UPDATE schema_environments SET source_kind_id = (
        SELECT sk.kind_id FROM source_kinds sk WHERE sk.id = schema_environments.source_kind_id
      );

      ALTER TABLE schema_environments ADD CONSTRAINT schema_environments_source_kind_id_fkey FOREIGN KEY (source_kind_id) REFERENCES kind(id);

      DELETE FROM source_kinds where active = false;
      ALTER TABLE source_kinds DROP COLUMN IF EXISTS active;
  END IF;
END $$;
-- +goose StatementEnd

-- +goose StatementBegin
CREATE OR REPLACE FUNCTION upsert_source_kind(source_kind_name TEXT) RETURNS source_kinds AS $$
DECLARE
  kind_row kind%rowtype;
  source_kind_row source_kinds%rowtype;
BEGIN
    -- Use advisory lock to serialize calls with the same source kind name
    PERFORM pg_advisory_xact_lock(hashtext(source_kind_name));

    SELECT * INTO kind_row FROM upsert_kind(source_kind_name);

    -- Then, try to find existing source_kind by kind_id
    SELECT * INTO source_kind_row FROM source_kinds WHERE kind_id = kind_row.id;

    IF source_kind_row.id IS NULL THEN
            INSERT INTO source_kinds (kind_id) VALUES (kind_row.id) RETURNING * INTO source_kind_row;
    END IF;

RETURN source_kind_row;
END $$
LANGUAGE plpgsql;
-- +goose StatementEnd

-- Add API Key Expiration Support feature flag
INSERT INTO feature_flags (created_at, updated_at, key, name, description, enabled, user_updatable)
VALUES (current_timestamp,
        current_timestamp,
        'api_key_expiration_support',
        'API Key Expiration Support',
        'Enables API Key Expiration configuration options',
        false,
        false)
ON CONFLICT DO NOTHING;

-- Add API Tokens Expiration Parameter
INSERT INTO parameters (key, name, description, value, created_at, updated_at)
VALUES ('auth.api_token_expiration',
        'API Token Expiration',
        'This configuration parameter enables/disables created API tokens to expire after the set number of days.',
        '{"enabled":false, "expiration_period":90}',
        current_timestamp,
        current_timestamp)
ON CONFLICT DO NOTHING;

-- Add column to allow deletion of relationships by kind
ALTER TABLE analysis_request_switch ADD COLUMN IF NOT EXISTS delete_relationships text [] DEFAULT ARRAY []::text [];

-- Auditors to have all environments access
UPDATE users
SET all_environments = true
WHERE id IN (
  SELECT u.id
  FROM users u
         JOIN users_roles ur ON ur.user_id = u.id
         JOIN roles r ON ur.role_id = r.id
  WHERE r.name = 'Auditor'
);

-- Make opengraph_extension_management user updatable
UPDATE feature_flags
SET user_updatable = true,
    updated_at = current_timestamp
WHERE key = 'opengraph_extension_management';

-- Set client_bearer_auth feature flag to default to enabled
UPDATE feature_flags
SET enabled = true,
    updated_at = current_timestamp
WHERE key = 'client_bearer_auth';
-- +goose Down
-- Intentionally empty - baseline cannot be rolled back
