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

CREATE TABLE IF NOT EXISTS client_ingest_tasks (
    file_name text,
    request_guid text,
    task_id bigint,
    id bigint NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone
);

CREATE SEQUENCE IF NOT EXISTS client_ingest_tasks_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;
ALTER SEQUENCE client_ingest_tasks_id_seq OWNED BY client_ingest_tasks.id;

CREATE TABLE IF NOT EXISTS client_scheduled_jobs (
    client_id text,
    client_name text,
    client_schedule_id integer,
    execution_time timestamp with time zone,
    status bigint,
    status_message text,
    start_time timestamp with time zone,
    end_time timestamp with time zone,
    log_path text,
    session_collection boolean,
    local_group_collection boolean,
    ad_structure_collection boolean,
    all_trusted_domains boolean,
    domain_controller text,
    event_title text,
    last_ingest timestamp with time zone,
    ous text[],
    domains text[],
    id bigint NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone
);

CREATE SEQUENCE IF NOT EXISTS client_scheduled_jobs_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;
ALTER SEQUENCE client_scheduled_jobs_id_seq OWNED BY client_scheduled_jobs.id;

CREATE TABLE IF NOT EXISTS client_schedules (
    client_id text,
    rrule text,
    session_collection boolean,
    local_group_collection boolean,
    ad_structure_collection boolean,
    all_trusted_domains boolean,
    next_scheduled_at timestamp with time zone,
    deleted_at timestamp with time zone,
    ous text[],
    domains text[],
    id integer NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone
);

CREATE SEQUENCE IF NOT EXISTS client_schedules_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;
ALTER SEQUENCE client_schedules_id_seq OWNED BY client_schedules.id;

CREATE TABLE IF NOT EXISTS clients (
    name text,
    ip_address text,
    hostname text,
    configured_user text,
    last_checkin timestamp with time zone,
    current_job_id bigint,
    completed_job_count integer,
    domain_controller text,
    version text,
    user_sid text,
    deleted_at timestamp with time zone,
    type text,
    id text NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone
);

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

CREATE TABLE IF NOT EXISTS domain_details (
    name text,
    object_id text NOT NULL,
    "exists" boolean,
    type text
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

CREATE TABLE IF NOT EXISTS list_findings (
    principal text,
    principal_kind text,
    finding text,
    domain_sid text,
    props jsonb,
    accepted_until timestamp with time zone,
    id integer NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone
);

CREATE SEQUENCE IF NOT EXISTS list_findings_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;
ALTER SEQUENCE list_findings_id_seq OWNED BY list_findings.id;

CREATE TABLE IF NOT EXISTS ou_details (
    name text,
    object_id text NOT NULL,
    "exists" boolean,
    distinguished_name text,
    type text
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

CREATE TABLE IF NOT EXISTS relationship_findings (
    from_principal text,
    to_principal text,
    from_principal_props jsonb,
    from_principal_kind text,
    to_principal_props jsonb,
    to_principal_kind text,
    rel_props jsonb,
    combo_graph_relation_id bigint,
    finding text,
    domain_sid text,
    principals_hash text,
    accepted_until timestamp with time zone,
    id integer NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone
);

CREATE SEQUENCE IF NOT EXISTS relationship_findings_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;
ALTER SEQUENCE relationship_findings_id_seq OWNED BY relationship_findings.id;

CREATE TABLE IF NOT EXISTS risk_counts (
    composite_risk bigint,
    finding_count bigint,
    impacted_asset_count bigint,
    domain_sid text,
    finding text,
    id bigint NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone
);

CREATE SEQUENCE IF NOT EXISTS risk_counts_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;
ALTER SEQUENCE risk_counts_id_seq OWNED BY risk_counts.id;

CREATE TABLE IF NOT EXISTS risk_posture_stats (
    domain_sid text,
    exposure_index numeric,
    tier_zero_count bigint,
    critical_risk_count bigint,
    id bigint NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone
);

CREATE SEQUENCE IF NOT EXISTS risk_posture_stats_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;
ALTER SEQUENCE risk_posture_stats_id_seq OWNED BY risk_posture_stats.id;

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
ALTER TABLE ONLY client_ingest_tasks ALTER COLUMN id SET DEFAULT nextval('client_ingest_tasks_id_seq'::regclass);
ALTER TABLE ONLY client_scheduled_jobs ALTER COLUMN id SET DEFAULT nextval('client_scheduled_jobs_id_seq'::regclass);
ALTER TABLE ONLY client_schedules ALTER COLUMN id SET DEFAULT nextval('client_schedules_id_seq'::regclass);
ALTER TABLE ONLY domain_collection_results ALTER COLUMN id SET DEFAULT nextval('domain_collection_results_id_seq'::regclass);
ALTER TABLE ONLY feature_flags ALTER COLUMN id SET DEFAULT nextval('feature_flags_id_seq'::regclass);
ALTER TABLE ONLY file_upload_jobs ALTER COLUMN id SET DEFAULT nextval('file_upload_jobs_id_seq'::regclass);
ALTER TABLE ONLY ingest_tasks ALTER COLUMN id SET DEFAULT nextval('ingest_tasks_id_seq'::regclass);
ALTER TABLE ONLY list_findings ALTER COLUMN id SET DEFAULT nextval('list_findings_id_seq'::regclass);
ALTER TABLE ONLY parameters ALTER COLUMN id SET DEFAULT nextval('parameters_id_seq'::regclass);
ALTER TABLE ONLY permissions ALTER COLUMN id SET DEFAULT nextval('permissions_id_seq'::regclass);
ALTER TABLE ONLY relationship_findings ALTER COLUMN id SET DEFAULT nextval('relationship_findings_id_seq'::regclass);
ALTER TABLE ONLY risk_counts ALTER COLUMN id SET DEFAULT nextval('risk_counts_id_seq'::regclass);
ALTER TABLE ONLY risk_posture_stats ALTER COLUMN id SET DEFAULT nextval('risk_posture_stats_id_seq'::regclass);
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
ALTER TABLE ONLY client_ingest_tasks
    ADD CONSTRAINT client_ingest_tasks_pkey PRIMARY KEY (id);
ALTER TABLE ONLY client_scheduled_jobs
    ADD CONSTRAINT client_scheduled_jobs_pkey PRIMARY KEY (id);
ALTER TABLE ONLY client_schedules
    ADD CONSTRAINT client_schedules_pkey PRIMARY KEY (id);
ALTER TABLE ONLY clients
    ADD CONSTRAINT clients_pkey PRIMARY KEY (id);
ALTER TABLE ONLY domain_collection_results
    ADD CONSTRAINT domain_collection_results_pkey PRIMARY KEY (id);
ALTER TABLE ONLY domain_details
    ADD CONSTRAINT domain_details_pkey PRIMARY KEY (object_id);
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
ALTER TABLE ONLY list_findings
    ADD CONSTRAINT list_findings_pkey PRIMARY KEY (id);
ALTER TABLE ONLY ou_details
    ADD CONSTRAINT ou_details_pkey PRIMARY KEY (object_id);
ALTER TABLE ONLY parameters
    ADD CONSTRAINT parameters_key_key UNIQUE (key);
ALTER TABLE ONLY parameters
    ADD CONSTRAINT parameters_pkey PRIMARY KEY (id);
ALTER TABLE ONLY permissions
    ADD CONSTRAINT permissions_pkey PRIMARY KEY (id);
ALTER TABLE ONLY relationship_findings
    ADD CONSTRAINT relationship_findings_pkey PRIMARY KEY (id);
ALTER TABLE ONLY risk_counts
    ADD CONSTRAINT risk_counts_pkey PRIMARY KEY (id);
ALTER TABLE ONLY risk_posture_stats
    ADD CONSTRAINT risk_posture_stats_pkey PRIMARY KEY (id);
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
CREATE INDEX IF NOT EXISTS idx_client_schedules_deleted_at ON client_schedules USING btree (deleted_at);
CREATE INDEX IF NOT EXISTS idx_clients_deleted_at ON clients USING btree (deleted_at);
CREATE INDEX IF NOT EXISTS idx_list_findings_domain_s_id ON list_findings USING btree (domain_sid);
CREATE INDEX IF NOT EXISTS idx_list_findings_finding ON list_findings USING btree (finding);
CREATE INDEX IF NOT EXISTS idx_list_findings_principal ON list_findings USING btree (principal);
CREATE INDEX IF NOT EXISTS idx_relationship_findings_domain_s_id ON relationship_findings USING btree (domain_sid);
CREATE INDEX IF NOT EXISTS idx_relationship_findings_finding ON relationship_findings USING btree (finding);
CREATE INDEX IF NOT EXISTS idx_relationship_findings_principals_hash ON relationship_findings USING btree (principals_hash);
CREATE INDEX IF NOT EXISTS idx_risk_counts_domain_s_id ON risk_counts USING btree (domain_sid);
CREATE INDEX IF NOT EXISTS idx_risk_counts_finding ON risk_counts USING btree (finding);
CREATE INDEX IF NOT EXISTS idx_saml_providers_name ON saml_providers USING btree (name);
CREATE UNIQUE INDEX IF NOT EXISTS idx_saved_queries_composite_index ON saved_queries USING btree (user_id, name);
CREATE INDEX IF NOT EXISTS idx_users_principal_name ON users USING btree (principal_name);

ALTER TABLE ONLY asset_group_collection_entries
    ADD CONSTRAINT fk_asset_group_collections_entries FOREIGN KEY (asset_group_collection_id) REFERENCES asset_group_collections(id) ON DELETE CASCADE;
ALTER TABLE ONLY asset_group_collections
    ADD CONSTRAINT fk_asset_groups_collections FOREIGN KEY (asset_group_id) REFERENCES asset_groups(id) ON DELETE CASCADE;
ALTER TABLE ONLY asset_group_selectors
    ADD CONSTRAINT fk_asset_groups_selectors FOREIGN KEY (asset_group_id) REFERENCES asset_groups(id) ON DELETE CASCADE;
ALTER TABLE ONLY client_schedules
    ADD CONSTRAINT fk_clients_schedules FOREIGN KEY (client_id) REFERENCES clients(id);
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
