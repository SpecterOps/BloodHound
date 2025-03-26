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

CREATE TABLE ad_data_quality_aggregations (
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
CREATE SEQUENCE ad_data_quality_aggregations_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;
ALTER SEQUENCE ad_data_quality_aggregations_id_seq OWNED BY ad_data_quality_aggregations.id;

CREATE TABLE ad_data_quality_stats (
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
CREATE SEQUENCE ad_data_quality_stats_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;
ALTER SEQUENCE ad_data_quality_stats_id_seq OWNED BY ad_data_quality_stats.id;

CREATE TABLE analysis_request_switch (
    singleton boolean DEFAULT true NOT NULL,
    request_type text NOT NULL,
    requested_by text NOT NULL,
    requested_at timestamp with time zone NOT NULL,
    CONSTRAINT singleton_uni CHECK (singleton)
);
CREATE TABLE asset_group_collection_entries (
    asset_group_collection_id bigint,
    object_id text,
    node_label text,
    properties jsonb,
    id bigint NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone
);
CREATE SEQUENCE asset_group_collection_entries_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;
ALTER SEQUENCE asset_group_collection_entries_id_seq OWNED BY asset_group_collection_entries.id;

CREATE TABLE asset_group_collections (
    asset_group_id integer,
    id bigint NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone
);
CREATE SEQUENCE asset_group_collections_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;
ALTER SEQUENCE asset_group_collections_id_seq OWNED BY asset_group_collections.id;

CREATE TABLE asset_group_selectors (
    asset_group_id integer,
    name text,
    selector text,
    system_selector boolean,
    id integer NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone
);
CREATE SEQUENCE asset_group_selectors_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;
ALTER SEQUENCE asset_group_selectors_id_seq OWNED BY asset_group_selectors.id;

CREATE TABLE asset_groups (
    name text NOT NULL,
    tag text NOT NULL,
    system_group boolean,
    id integer NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone
);
CREATE SEQUENCE asset_groups_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;
ALTER SEQUENCE asset_groups_id_seq OWNED BY asset_groups.id;

CREATE TABLE audit_logs (
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
    CONSTRAINT status_check CHECK (((status)::text = ANY ((ARRAY['intent'::character varying, 'success'::character varying, 'failure'::character varying])::text[])))
);
CREATE SEQUENCE audit_logs_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;
ALTER SEQUENCE audit_logs_id_seq OWNED BY audit_logs.id;

CREATE TABLE auth_secrets (
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
CREATE SEQUENCE auth_secrets_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;
ALTER SEQUENCE auth_secrets_id_seq OWNED BY auth_secrets.id;

CREATE TABLE auth_tokens (
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
CREATE TABLE azure_data_quality_aggregations (
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
CREATE SEQUENCE azure_data_quality_aggregations_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;
ALTER SEQUENCE azure_data_quality_aggregations_id_seq OWNED BY azure_data_quality_aggregations.id;

CREATE TABLE azure_data_quality_stats (
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
CREATE SEQUENCE azure_data_quality_stats_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;
ALTER SEQUENCE azure_data_quality_stats_id_seq OWNED BY azure_data_quality_stats.id;

CREATE TABLE datapipe_status (
    singleton boolean DEFAULT true NOT NULL,
    status text NOT NULL,
    updated_at timestamp with time zone NOT NULL,
    last_complete_analysis_at timestamp with time zone,
    CONSTRAINT singleton_uni CHECK (singleton)
);
CREATE TABLE domain_collection_results (
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
CREATE SEQUENCE domain_collection_results_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;
ALTER SEQUENCE domain_collection_results_id_seq OWNED BY domain_collection_results.id;

CREATE TABLE feature_flags (
    id integer NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    key text,
    name text,
    description text,
    enabled boolean,
    user_updatable boolean
);
CREATE SEQUENCE feature_flags_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;
ALTER SEQUENCE feature_flags_id_seq OWNED BY feature_flags.id;

CREATE TABLE file_upload_jobs (
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
CREATE SEQUENCE file_upload_jobs_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;
ALTER SEQUENCE file_upload_jobs_id_seq OWNED BY file_upload_jobs.id;

CREATE TABLE ingest_tasks (
    file_name text,
    request_guid text,
    task_id bigint,
    id bigint NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    file_type integer DEFAULT 0
);
CREATE SEQUENCE ingest_tasks_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;
ALTER SEQUENCE ingest_tasks_id_seq OWNED BY ingest_tasks.id;

CREATE TABLE installations (
    id text NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone
);
CREATE TABLE parameters (
    key text,
    name text,
    description text,
    value jsonb,
    id integer NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone
);
CREATE SEQUENCE parameters_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;
ALTER SEQUENCE parameters_id_seq OWNED BY parameters.id;

CREATE TABLE permissions (
    authority text,
    name text,
    id integer NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone
);
CREATE SEQUENCE permissions_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;
ALTER SEQUENCE permissions_id_seq OWNED BY permissions.id;

CREATE TABLE roles (
    name text,
    description text,
    id integer NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone
);
CREATE SEQUENCE roles_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;
ALTER SEQUENCE roles_id_seq OWNED BY roles.id;

CREATE TABLE roles_permissions (
    role_id integer NOT NULL,
    permission_id integer NOT NULL
);
CREATE TABLE saml_providers (
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
CREATE SEQUENCE saml_providers_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;
ALTER SEQUENCE saml_providers_id_seq OWNED BY saml_providers.id;

CREATE TABLE saved_queries (
    user_id text,
    name text,
    query text,
    id bigint NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    description text DEFAULT ''::text
);
CREATE SEQUENCE saved_queries_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;
ALTER SEQUENCE saved_queries_id_seq OWNED BY saved_queries.id;

CREATE TABLE saved_queries_permissions (
    id bigint NOT NULL,
    shared_to_user_id text,
    query_id bigint NOT NULL,
    public boolean DEFAULT false NOT NULL,
    created_at timestamp with time zone DEFAULT now(),
    updated_at timestamp with time zone DEFAULT now()
);
CREATE SEQUENCE saved_queries_permissions_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;
ALTER SEQUENCE saved_queries_permissions_id_seq OWNED BY saved_queries_permissions.id;

CREATE TABLE user_sessions (
    user_id text,
    auth_provider_type bigint,
    auth_provider_id integer,
    expires_at timestamp with time zone,
    id bigint NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    flags jsonb
);
CREATE SEQUENCE user_sessions_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;
ALTER SEQUENCE user_sessions_id_seq OWNED BY user_sessions.id;

CREATE TABLE users (
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
CREATE TABLE users_roles (
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
ALTER TABLE ONLY saved_queries_permissions ALTER COLUMN id SET DEFAULT nextval('saved_queries_permissions_id_seq'::regclass);
ALTER TABLE ONLY user_sessions ALTER COLUMN id SET DEFAULT nextval('user_sessions_id_seq'::regclass);

ALTER TABLE ONLY ad_data_quality_aggregations ADD CONSTRAINT ad_data_quality_aggregations_pkey PRIMARY KEY (id);
ALTER TABLE ONLY ad_data_quality_stats ADD CONSTRAINT ad_data_quality_stats_pkey PRIMARY KEY (id);
ALTER TABLE ONLY analysis_request_switch ADD CONSTRAINT analysis_request_switch_pkey PRIMARY KEY (singleton);
ALTER TABLE ONLY asset_group_collection_entries ADD CONSTRAINT asset_group_collection_entries_pkey PRIMARY KEY (id);
ALTER TABLE ONLY asset_group_collections ADD CONSTRAINT asset_group_collections_pkey PRIMARY KEY (id);
ALTER TABLE ONLY asset_group_selectors Add CONSTRAINT asset_group_selectors_name_assetgroupid_key UNIQUE (name, asset_group_id);
ALTER TABLE ONLY asset_group_selectors Add CONSTRAINT asset_group_selectors_pkey PRIMARY KEY (id);
ALTER TABLE ONLY asset_groups Add CONSTRAINT asset_groups_name_key UNIQUE (name);
ALTER TABLE ONLY asset_groups Add CONSTRAINT asset_groups_pkey PRIMARY KEY (id);
ALTER TABLE ONLY asset_groups Add CONSTRAINT asset_groups_tag_key UNIQUE (tag);
ALTER TABLE ONLY audit_logs Add CONSTRAINT audit_logs_pkey PRIMARY KEY (id);
ALTER TABLE ONLY auth_secrets Add CONSTRAINT auth_secrets_pkey PRIMARY KEY (id);
ALTER TABLE ONLY auth_tokens Add CONSTRAINT auth_tokens_pkey PRIMARY KEY (id);
ALTER TABLE ONLY azure_data_quality_aggregations ADD CONSTRAINT azure_data_quality_aggregations_pkey PRIMARY KEY (id);
ALTER TABLE ONLY azure_data_quality_stats Add CONSTRAINT azure_data_quality_stats_pkey PRIMARY KEY (id);
ALTER TABLE ONLY datapipe_status Add CONSTRAINT datapipe_status_pkey PRIMARY KEY (singleton);
ALTER TABLE ONLY domain_collection_results Add CONSTRAINT domain_collection_results_pkey PRIMARY KEY (id);
ALTER TABLE ONLY feature_flags Add CONSTRAINT feature_flags_key_key UNIQUE (key);
ALTER TABLE ONLY feature_flags Add CONSTRAINT feature_flags_pkey PRIMARY KEY (id);
ALTER TABLE ONLY file_upload_jobs Add CONSTRAINT file_upload_jobs_pkey PRIMARY KEY (id);
ALTER TABLE ONLY ingest_tasks Add CONSTRAINT ingest_tasks_pkey PRIMARY KEY (id);
ALTER TABLE ONLY installations Add CONSTRAINT installations_pkey PRIMARY KEY (id);
ALTER TABLE ONLY parameters Add CONSTRAINT parameters_key_key UNIQUE (key);
ALTER TABLE ONLY parameters Add CONSTRAINT parameters_pkey PRIMARY KEY (id);
ALTER TABLE ONLY permissions Add CONSTRAINT permissions_authority_name_key UNIQUE (authority, name);
ALTER TABLE ONLY permissions Add CONSTRAINT permissions_pkey PRIMARY KEY (id);
ALTER TABLE ONLY roles Add CONSTRAINT roles_name_key UNIQUE (name);
ALTER TABLE ONLY roles_permissions Add CONSTRAINT roles_permissions_pkey PRIMARY KEY (role_id, permission_id);
ALTER TABLE ONLY roles Add CONSTRAINT roles_pkey PRIMARY KEY (id);
ALTER TABLE ONLY saml_providers Add CONSTRAINT saml_providers_name_key UNIQUE (name);
ALTER TABLE ONLY saml_providers Add CONSTRAINT saml_providers_pkey PRIMARY KEY (id);
ALTER TABLE ONLY saved_queries_permissions Add CONSTRAINT saved_queries_permissions_pkey PRIMARY KEY (id);
ALTER TABLE ONLY saved_queries_permissions Add CONSTRAINT saved_queries_permissions_shared_to_user_id_query_id_key UNIQUE (shared_to_user_id, query_id);
ALTER TABLE ONLY saved_queries Add CONSTRAINT saved_queries_pkey PRIMARY KEY (id);
ALTER TABLE ONLY user_sessions Add CONSTRAINT user_sessions_pkey PRIMARY KEY (id);
ALTER TABLE ONLY users Add CONSTRAINT users_pkey PRIMARY KEY (id);
ALTER TABLE ONLY users Add CONSTRAINT users_principal_name_key UNIQUE (principal_name);
ALTER TABLE ONLY users_roles Add CONSTRAINT users_roles_pkey PRIMARY KEY (user_id, role_id);

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

ALTER TABLE ONLY asset_group_collection_entries Add CONSTRAINT fk_asset_group_collections_entries FOREIGN KEY (asset_group_collection_id) REFERENCES asset_group_collections(id) ON DELETE CASCADE;
ALTER TABLE ONLY asset_group_collections Add CONSTRAINT fk_asset_groups_collections FOREIGN KEY (asset_group_id) REFERENCES asset_groups(id) ON DELETE CASCADE;
ALTER TABLE ONLY asset_group_selectors Add CONSTRAINT fk_asset_groups_selectors FOREIGN KEY (asset_group_id) REFERENCES asset_groups(id) ON DELETE CASCADE;
ALTER TABLE ONLY file_upload_jobs Add CONSTRAINT fk_file_upload_jobs_user FOREIGN KEY (user_id) REFERENCES users(id);
ALTER TABLE ONLY roles_permissions Add CONSTRAINT fk_roles_permissions_permission FOREIGN KEY (permission_id) REFERENCES permissions(id);
ALTER TABLE ONLY roles_permissions Add CONSTRAINT fk_roles_permissions_role FOREIGN KEY (role_id) REFERENCES roles(id);
ALTER TABLE ONLY user_sessions Add CONSTRAINT fk_user_sessions_user FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE;
ALTER TABLE ONLY auth_secrets Add CONSTRAINT fk_users_auth_secret FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE;
ALTER TABLE ONLY auth_tokens Add CONSTRAINT fk_users_auth_tokens FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE;
ALTER TABLE ONLY users_roles Add CONSTRAINT fk_users_roles_role FOREIGN KEY (role_id) REFERENCES roles(id);
ALTER TABLE ONLY users_roles Add CONSTRAINT fk_users_roles_user FOREIGN KEY (user_id) REFERENCES users(id);
ALTER TABLE ONLY users Add CONSTRAINT fk_users_saml_provider FOREIGN KEY (saml_provider_id) REFERENCES saml_providers(id);
ALTER TABLE ONLY saved_queries_permissions Add CONSTRAINT saved_queries_permissions_query_id_fkey FOREIGN KEY (query_id) REFERENCES saved_queries(id) ON DELETE CASCADE;
ALTER TABLE ONLY saved_queries_permissions Add CONSTRAINT saved_queries_permissions_shared_to_user_id_fkey FOREIGN KEY (shared_to_user_id) REFERENCES users(id) ON DELETE CASCADE;

INSERT INTO asset_groups (name, tag, system_group, created_at, updated_at) VALUES ('Admin Tier Zero', 'admin_tier_0', true, current_timestamp, current_timestamp);

INSERT INTO datapipe_status VALUES (true, 'idle', current_timestamp, NULL);

INSERT INTO feature_flags VALUES (1, current_timestamp, current_timestamp, 'butterfly_analysis', 'Enhanced Asset Inbound-Outbound Exposure Analysis', 'Enables more extensive analysis of attack path findings that allows BloodHound to help the user prioritize remediation of the most exposed assets.', false, false);
INSERT INTO feature_flags VALUES (2, current_timestamp, current_timestamp, 'enable_saml_sso', 'SAML Single Sign-On Support', 'Enables SSO authentication flows and administration panels to third party SAML identity providers.', true, false);
INSERT INTO feature_flags VALUES (3, current_timestamp, current_timestamp, 'scope_collection_by_ou', 'Enable SharpHound OU Scoped Collections', 'Enables scoping SharpHound collections to specific lists of OUs.', true, false);
INSERT INTO feature_flags VALUES (4, current_timestamp, current_timestamp, 'azure_support', 'Enable Azure Support', 'Enables Azure support.', true, false);
INSERT INTO feature_flags VALUES (6, current_timestamp, current_timestamp, 'entity_panel_cache', 'Enable application level caching', 'Enables the use of application level caching for entity panel queries', true, false);
INSERT INTO feature_flags VALUES (7, current_timestamp, current_timestamp, 'adcs', 'Enable collection and processing of Active Directory Certificate Services Data', 'Enables the ability to collect, analyze, and explore Active Directory Certificate Services data and previews new attack paths.', false, false);
INSERT INTO feature_flags VALUES (8, current_timestamp, current_timestamp, 'dark_mode', 'Dark Mode', 'Allows users to enable or disable dark mode via a toggle in the settings menu', false, true);
INSERT INTO feature_flags VALUES (9, current_timestamp, current_timestamp, 'pg_migration_dual_ingest', 'PostgreSQL Migration Dual Ingest', 'Enables dual ingest pathing for both Neo4j and PostgreSQL.', false, false);
INSERT INTO feature_flags VALUES (10, current_timestamp, current_timestamp, 'clear_graph_data', 'Clear Graph Data', 'Enables the ability to delete all nodes and edges from the graph database.', true, false);
INSERT INTO feature_flags VALUES (11, current_timestamp, current_timestamp, 'risk_exposure_new_calculation', 'Use new tier zero risk exposure calculation', 'Enables the use of new tier zero risk exposure metatree metrics.', false, false);
INSERT INTO feature_flags VALUES (12, current_timestamp, current_timestamp, 'fedramp_eula', 'FedRAMP EULA', 'Enables showing the FedRAMP EULA on every login. (Enterprise only)', false, false);
INSERT INTO feature_flags VALUES (13, current_timestamp, current_timestamp, 'auto_tag_t0_parent_objects', 'Automatically add parent OUs and containers of Tier Zero AD objects to Tier Zero', 'Parent OUs and containers of Tier Zero AD objects are automatically added to Tier Zero during analysis. Containers are only added if they have a Tier Zero child object with ACL inheritance enabled.', true, true);
INSERT INTO feature_flags VALUES (14, current_timestamp, current_timestamp, 'oidc_support', 'OIDC Support', 'Enables OpenID Connect authentication support for SSO Authentication.', false, false);

INSERT INTO parameters VALUES ('auth.password_expiration_window', 'Local Auth Password Expiry Window', 'This configuration parameter sets the local auth password expiry window for users that have valid auth secrets. Values for this configuration must follow the duration specification of ISO-8601.', '{"duration": "P90D"}', 1, current_timestamp, current_timestamp);
INSERT INTO parameters VALUES ('neo4j.configuration', 'Neo4j Configuration Parameters', 'This configuration parameter sets the BatchWriteSize and the BatchFlushSize for Neo4J.', '{"batch_write_size": 20000, "write_flush_size": 100000}', 2, current_timestamp, current_timestamp);
INSERT INTO parameters VALUES ('analysis.citrix_rdp_support', 'Citrix RDP Support', 'This configuration parameter toggles Citrix support during post-processing. When enabled, computers identified with a ''Direct Access Users'' local group will assume that Citrix is installed and CanRDP edges will require membership of both ''Direct Access Users'' and ''Remote Desktop Users'' local groups on the computer.', '{ "enabled": false }', 3, current_timestamp, current_timestamp);
INSERT INTO parameters VALUES ('prune.ttl', 'Prune Retention TTL Configuration Parameters','This configuration parameter sets the retention TTLs during analysis pruning.', '{"base_ttl": "P7D", "has_session_edge_ttl": "P3D" }', 4, current_timestamp, current_timestamp);
INSERT INTO parameters VALUES ('analysis.reconciliation', 'Reconciliation', 'This configuration parameter enables / disables reconciliation during analysis.', '{"enabled": true}', 5, current_timestamp, current_timestamp);

INSERT INTO permissions VALUES ('app', 'ReadAppConfig', 1, current_timestamp, current_timestamp);
INSERT INTO permissions VALUES ('app', 'WriteAppConfig', 2, current_timestamp, current_timestamp);
INSERT INTO permissions VALUES ('risks', 'GenerateReport', 3, current_timestamp, current_timestamp);
INSERT INTO permissions VALUES ('risks', 'ManageRisks', 4, current_timestamp, current_timestamp);
INSERT INTO permissions VALUES ('auth', 'CreateToken', 5, current_timestamp, current_timestamp);
INSERT INTO permissions VALUES ('auth', 'ManageAppConfig', 6, current_timestamp, current_timestamp);
INSERT INTO permissions VALUES ('auth', 'ManageProviders', 7, current_timestamp, current_timestamp);
INSERT INTO permissions VALUES ('auth', 'ManageSelf', 8, current_timestamp, current_timestamp);
INSERT INTO permissions VALUES ('auth', 'ManageUsers', 9, current_timestamp, current_timestamp);
INSERT INTO permissions VALUES ('clients', 'Manage', 10, current_timestamp, current_timestamp);
INSERT INTO permissions VALUES ('clients', 'Tasking', 11, current_timestamp, current_timestamp);
INSERT INTO permissions VALUES ('collection', 'ManageJobs', 12, current_timestamp, current_timestamp);
INSERT INTO permissions VALUES ('graphdb', 'Read', 13, current_timestamp, current_timestamp);
INSERT INTO permissions VALUES ('graphdb', 'Write', 14, current_timestamp, current_timestamp);
INSERT INTO permissions VALUES ('saved_queries', 'Read', 15, current_timestamp, current_timestamp);
INSERT INTO permissions VALUES ('saved_queries', 'Write', 16, current_timestamp, current_timestamp);
INSERT INTO permissions VALUES ('clients', 'Read', 17, current_timestamp, current_timestamp);
INSERT INTO permissions VALUES ('db', 'Wipe', 18, current_timestamp, current_timestamp);
INSERT INTO permissions VALUES ('graphdb', 'Mutate', 19, current_timestamp, current_timestamp);

INSERT INTO roles VALUES ('Administrator', 'Can manage users, clients, and application configuration', 1, current_timestamp, current_timestamp);
INSERT INTO roles VALUES ('User', 'Can read data, modify asset group memberships', 2, current_timestamp, current_timestamp);
INSERT INTO roles VALUES ('Read-Only', 'Used for integrations', 3, current_timestamp, current_timestamp);
INSERT INTO roles VALUES ('Upload-Only', 'Used for data collection clients, can post data but cannot read data', 4, current_timestamp, current_timestamp);
INSERT INTO roles VALUES ('Power User', 'Can upload data, manage clients, and perform any action a User can', 5, current_timestamp, current_timestamp);

INSERT INTO roles_permissions (role_id, permission_id)
SELECT r.id, p.id
FROM roles r, permissions p
WHERE r.name = 'Administrator' AND p.name IN (
    'ReadAppConfig', 'WriteAppConfig', 'GenerateReport', 'ManageRisks',
    'CreateToken', 'ManageAppConfig', 'ManageProviders', 'ManageSelf',
    'ManageUsers', 'Manage', 'Tasking', 'ManageJobs', 'Read', 'Write'
);

-- User
INSERT INTO roles_permissions (role_id, permission_id)
SELECT r.id, p.id
FROM roles r, permissions p
WHERE r.name = 'User' AND p.name IN (
    'ReadAppConfig', 'GenerateReport', 'CreateToken', 'ManageSelf', 'Manage', 'Read'
);

-- Read-Only
INSERT INTO roles_permissions (role_id, permission_id)
SELECT r.id, p.id
FROM roles r, permissions p
WHERE r.name = 'Read-Only' AND p.name IN (
    'ReadAppConfig', 'GenerateReport', 'ManageSelf', 'Read', 'CreateToken'
);

-- Upload-Only
INSERT INTO roles_permissions (role_id, permission_id)
SELECT r.id, p.id
FROM roles r, permissions p
WHERE r.name = 'Upload-Only' AND p.name IN (
    'Tasking', 'Write'
);

-- Power User
INSERT INTO roles_permissions (role_id, permission_id)
SELECT r.id, p.id
FROM roles r, permissions p
WHERE r.name = 'Power User' AND p.name IN (
    'ReadAppConfig', 'WriteAppConfig', 'GenerateReport', 'ManageRisks',
    'CreateToken', 'ManageSelf', 'Manage', 'Read', 'Tasking', 'ManageJobs',
    'Write', 'Mutate', 'Read', 'Mutate'
);

-- Make sure the sequence values of things all line up at the end
SELECT pg_catalog.setval('ad_data_quality_aggregations_id_seq', COALESCE(MAX(id), 1), true) FROM ad_data_quality_aggregations;
SELECT pg_catalog.setval('ad_data_quality_stats_id_seq', COALESCE(MAX(id), 1), true) FROM ad_data_quality_stats;
SELECT pg_catalog.setval('asset_group_collection_entries_id_seq', COALESCE(MAX(id), 1), true) FROM asset_group_collection_entries;
SELECT pg_catalog.setval('asset_group_collections_id_seq', COALESCE(MAX(id), 1), true) FROM asset_group_collections;
SELECT pg_catalog.setval('asset_group_selectors_id_seq', COALESCE(MAX(id), 1), true) FROM asset_group_selectors;
SELECT pg_catalog.setval('asset_groups_id_seq', COALESCE(MAX(id), 1), true) FROM asset_groups;
SELECT pg_catalog.setval('audit_logs_id_seq', COALESCE(MAX(id), 1), true) FROM audit_logs;
SELECT pg_catalog.setval('auth_secrets_id_seq', COALESCE(MAX(id), 1), true) FROM auth_secrets;
SELECT pg_catalog.setval('azure_data_quality_aggregations_id_seq', COALESCE(MAX(id), 1), true) FROM azure_data_quality_aggregations;
SELECT pg_catalog.setval('azure_data_quality_stats_id_seq', COALESCE(MAX(id), 1), true) FROM azure_data_quality_stats;
SELECT pg_catalog.setval('domain_collection_results_id_seq', COALESCE(MAX(id), 1), true) FROM domain_collection_results;
SELECT pg_catalog.setval('feature_flags_id_seq', COALESCE(MAX(id), 1), true) FROM feature_flags;
SELECT pg_catalog.setval('file_upload_jobs_id_seq', COALESCE(MAX(id), 1), true) FROM file_upload_jobs;
SELECT pg_catalog.setval('ingest_tasks_id_seq', COALESCE(MAX(id), 1), true) FROM ingest_tasks;
SELECT pg_catalog.setval('parameters_id_seq', COALESCE(MAX(id), 1), true) FROM parameters;
SELECT pg_catalog.setval('permissions_id_seq', COALESCE(MAX(id), 1), true) FROM permissions;
SELECT pg_catalog.setval('roles_id_seq', COALESCE(MAX(id), 1), true) FROM roles;
SELECT pg_catalog.setval('saml_providers_id_seq', COALESCE(MAX(id), 1), true) FROM saml_providers;
SELECT pg_catalog.setval('saved_queries_id_seq', COALESCE(MAX(id), 1), true) FROM saved_queries;
SELECT pg_catalog.setval('saved_queries_permissions_id_seq', COALESCE(MAX(id), 1), true) FROM saved_queries_permissions;
SELECT pg_catalog.setval('user_sessions_id_seq', COALESCE(MAX(id), 1), true) FROM user_sessions;