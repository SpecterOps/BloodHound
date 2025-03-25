
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

CREATE SEQUENCE saved_queries_permissions_query_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;
ALTER SEQUENCE saved_queries_permissions_query_id_seq OWNED BY saved_queries_permissions.query_id;

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

ALTER TABLE ONLY saved_queries_permissions ALTER COLUMN query_id SET DEFAULT nextval('saved_queries_permissions_query_id_seq'::regclass);

ALTER TABLE ONLY user_sessions ALTER COLUMN id SET DEFAULT nextval('user_sessions_id_seq'::regclass);

INSERT INTO datapipe_status VALUES (true, 'idle', '2025-03-21 12:58:42.971311-05', NULL);

INSERT INTO feature_flags VALUES (1, '2025-03-21 12:56:17.075059-05', '2025-03-21 12:56:17.075059-05', 'butterfly_analysis', 'Enhanced Asset Inbound-Outbound Exposure Analysis', 'Enables more extensive analysis of attack path findings that allows BloodHound to help the user prioritize remediation of the most exposed assets.', false, false);
INSERT INTO feature_flags VALUES (2, '2025-03-21 12:56:17.075474-05', '2025-03-21 12:56:17.075474-05', 'enable_saml_sso', 'SAML Single Sign-On Support', 'Enables SSO authentication flows and administration panels to third party SAML identity providers.', true, false);
INSERT INTO feature_flags VALUES (3, '2025-03-21 12:56:17.07564-05', '2025-03-21 12:56:17.07564-05', 'scope_collection_by_ou', 'Enable SharpHound OU Scoped Collections', 'Enables scoping SharpHound collections to specific lists of OUs.', true, false);
INSERT INTO feature_flags VALUES (4, '2025-03-21 12:56:17.075794-05', '2025-03-21 12:56:17.075794-05', 'azure_support', 'Enable Azure Support', 'Enables Azure support.', true, false);
INSERT INTO feature_flags VALUES (5, '2025-03-21 12:56:17.075951-05', '2025-03-21 12:56:17.075951-05', 'reconciliation', 'Reconciliation', 'Enables Reconciliation', true, false);
INSERT INTO feature_flags VALUES (6, '2025-03-21 12:56:17.076129-05', '2025-03-21 12:56:17.076129-05', 'entity_panel_cache', 'Enable application level caching', 'Enables the use of application level caching for entity panel queries', true, false);
INSERT INTO feature_flags VALUES (7, '2025-03-21 12:56:17.076293-05', '2025-03-21 12:56:17.076293-05', 'adcs', 'Enable collection and processing of Active Directory Certificate Services Data', 'Enables the ability to collect, analyze, and explore Active Directory Certificate Services data and previews new attack paths.', false, false);
INSERT INTO feature_flags VALUES (8, '2025-03-21 12:58:42.987985-05', '2025-03-21 12:58:42.987985-05', 'dark_mode', 'Dark Mode', 'Allows users to enable or disable dark mode via a toggle in the settings menu', false, true);
INSERT INTO feature_flags VALUES (9, '2025-03-21 12:58:43.022829-05', '2025-03-21 12:58:43.022829-05', 'pg_migration_dual_ingest', 'PostgreSQL Migration Dual Ingest', 'Enables dual ingest pathing for both Neo4j and PostgreSQL.', false, false);
INSERT INTO feature_flags VALUES (10, '2025-03-21 12:58:43.038933-05', '2025-03-21 12:58:43.038933-05', 'clear_graph_data', 'Clear Graph Data', 'Enables the ability to delete all nodes and edges from the graph database.', true, false);
INSERT INTO feature_flags VALUES (11, '2025-03-21 12:58:43.039466-05', '2025-03-21 12:58:43.039466-05', 'risk_exposure_new_calculation', 'Use new tier zero risk exposure calculation', 'Enables the use of new tier zero risk exposure metatree metrics.', false, false);
INSERT INTO feature_flags VALUES (12, '2025-03-21 12:58:43.039655-05', '2025-03-21 12:58:43.039655-05', 'fedramp_eula', 'FedRAMP EULA', 'Enables showing the FedRAMP EULA on every login. (Enterprise only)', false, false);
INSERT INTO feature_flags VALUES (13, '2025-03-21 12:58:43.039915-05', '2025-03-21 12:58:43.039915-05', 'auto_tag_t0_parent_objects', 'Automatically add parent OUs and containers of Tier Zero AD objects to Tier Zero', 'Parent OUs and containers of Tier Zero AD objects are automatically added to Tier Zero during analysis. Containers are only added if they have a Tier Zero child object with ACL inheritance enabled.', true, true);

INSERT INTO parameters VALUES ('auth.password_expiration_window', 'Local Auth Password Expiry Window', 'This configuration parameter sets the local auth password expiry window for users that have valid auth secrets. Values for this configuration must follow the duration specification of ISO-8601.', '{"duration": "P90D"}', 1, '2025-03-21 12:56:17.07645-05', '2025-03-21 12:56:17.07645-05');
INSERT INTO parameters VALUES ('neo4j.configuration', 'Neo4j Configuration Parameters', 'This configuration parameter sets the BatchWriteSize and the BatchFlushSize for Neo4J.', '{"batch_write_size": 20000, "write_flush_size": 100000}', 2, '2025-03-21 12:56:17.077247-05', '2025-03-21 12:56:17.077247-05');

INSERT INTO permissions VALUES ('app', 'ReadAppConfig', 1, '2025-03-21 12:56:17.066242-05', '2025-03-21 12:56:17.066242-05');
INSERT INTO permissions VALUES ('app', 'WriteAppConfig', 2, '2025-03-21 12:56:17.066532-05', '2025-03-21 12:56:17.066532-05');
INSERT INTO permissions VALUES ('risks', 'GenerateReport', 3, '2025-03-21 12:56:17.066707-05', '2025-03-21 12:56:17.066707-05');
INSERT INTO permissions VALUES ('risks', 'ManageRisks', 4, '2025-03-21 12:56:17.066878-05', '2025-03-21 12:56:17.066878-05');
INSERT INTO permissions VALUES ('auth', 'CreateToken', 5, '2025-03-21 12:56:17.067068-05', '2025-03-21 12:56:17.067068-05');
INSERT INTO permissions VALUES ('auth', 'ManageAppConfig', 6, '2025-03-21 12:56:17.067239-05', '2025-03-21 12:56:17.067239-05');
INSERT INTO permissions VALUES ('auth', 'ManageProviders', 7, '2025-03-21 12:56:17.067393-05', '2025-03-21 12:56:17.067393-05');
INSERT INTO permissions VALUES ('auth', 'ManageSelf', 8, '2025-03-21 12:56:17.067537-05', '2025-03-21 12:56:17.067537-05');
INSERT INTO permissions VALUES ('auth', 'ManageUsers', 9, '2025-03-21 12:56:17.067704-05', '2025-03-21 12:56:17.067704-05');
INSERT INTO permissions VALUES ('clients', 'Manage', 10, '2025-03-21 12:56:17.067872-05', '2025-03-21 12:56:17.067872-05');
INSERT INTO permissions VALUES ('clients', 'Tasking', 11, '2025-03-21 12:56:17.068033-05', '2025-03-21 12:56:17.068033-05');
INSERT INTO permissions VALUES ('collection', 'ManageJobs', 12, '2025-03-21 12:56:17.068194-05', '2025-03-21 12:56:17.068194-05');
INSERT INTO permissions VALUES ('graphdb', 'Read', 13, '2025-03-21 12:56:17.068354-05', '2025-03-21 12:56:17.068354-05');
INSERT INTO permissions VALUES ('graphdb', 'Write', 14, '2025-03-21 12:56:17.068512-05', '2025-03-21 12:56:17.068512-05');
INSERT INTO permissions VALUES ('saved_queries', 'Read', 15, '2025-03-21 12:58:43.040108-05', '2025-03-21 12:58:43.040108-05');
INSERT INTO permissions VALUES ('saved_queries', 'Write', 16, '2025-03-21 12:58:43.040475-05', '2025-03-21 12:58:43.040475-05');
INSERT INTO permissions VALUES ('clients', 'Read', 17, '2025-03-21 12:58:43.040655-05', '2025-03-21 12:58:43.040655-05');
INSERT INTO permissions VALUES ('db', 'Wipe', 18, '2025-03-21 12:58:43.04082-05', '2025-03-21 12:58:43.04082-05');
INSERT INTO permissions VALUES ('graphdb', 'Mutate', 19, '2025-03-21 12:58:43.041001-05', '2025-03-21 12:58:43.041001-05');

INSERT INTO roles VALUES ('Administrator', 'Can manage users, clients, and application configuration', 1, '2025-03-21 12:56:17.0687-05', '2025-03-21 12:56:17.0687-05');
INSERT INTO roles VALUES ('User', 'Can read data, modify asset group memberships', 2, '2025-03-21 12:56:17.069018-05', '2025-03-21 12:56:17.069018-05');
INSERT INTO roles VALUES ('Read-Only', 'Used for integrations', 3, '2025-03-21 12:56:17.06919-05', '2025-03-21 12:56:17.06919-05');
INSERT INTO roles VALUES ('Upload-Only', 'Used for data collection clients, can post data but cannot read data', 4, '2025-03-21 12:56:17.069362-05', '2025-03-21 12:56:17.069362-05');
INSERT INTO roles VALUES ('Power User', 'Can upload data, manage clients, and perform any action a User can', 5, '2025-03-21 12:58:43.041242-05', '2025-03-21 12:58:43.041242-05');

INSERT INTO roles_permissions VALUES (1, 1);
INSERT INTO roles_permissions VALUES (1, 2);
INSERT INTO roles_permissions VALUES (1, 3);
INSERT INTO roles_permissions VALUES (1, 4);
INSERT INTO roles_permissions VALUES (1, 5);
INSERT INTO roles_permissions VALUES (1, 6);
INSERT INTO roles_permissions VALUES (1, 7);
INSERT INTO roles_permissions VALUES (1, 8);
INSERT INTO roles_permissions VALUES (1, 9);
INSERT INTO roles_permissions VALUES (1, 10);
INSERT INTO roles_permissions VALUES (1, 11);
INSERT INTO roles_permissions VALUES (1, 12);
INSERT INTO roles_permissions VALUES (1, 13);
INSERT INTO roles_permissions VALUES (1, 14);
INSERT INTO roles_permissions VALUES (2, 1);
INSERT INTO roles_permissions VALUES (2, 3);
INSERT INTO roles_permissions VALUES (2, 5);
INSERT INTO roles_permissions VALUES (2, 8);
INSERT INTO roles_permissions VALUES (2, 13);
INSERT INTO roles_permissions VALUES (3, 1);
INSERT INTO roles_permissions VALUES (3, 3);
INSERT INTO roles_permissions VALUES (3, 8);
INSERT INTO roles_permissions VALUES (3, 13);
INSERT INTO roles_permissions VALUES (4, 11);
INSERT INTO roles_permissions VALUES (4, 14);
INSERT INTO roles_permissions VALUES (1, 15);
INSERT INTO roles_permissions VALUES (1, 16);
INSERT INTO roles_permissions VALUES (1, 17);
INSERT INTO roles_permissions VALUES (1, 18);
INSERT INTO roles_permissions VALUES (1, 19);
INSERT INTO roles_permissions VALUES (2, 15);
INSERT INTO roles_permissions VALUES (2, 16);
INSERT INTO roles_permissions VALUES (2, 17);
INSERT INTO roles_permissions VALUES (3, 5);
INSERT INTO roles_permissions VALUES (5, 1);
INSERT INTO roles_permissions VALUES (5, 2);
INSERT INTO roles_permissions VALUES (5, 3);
INSERT INTO roles_permissions VALUES (5, 4);
INSERT INTO roles_permissions VALUES (5, 5);
INSERT INTO roles_permissions VALUES (5, 8);
INSERT INTO roles_permissions VALUES (5, 10);
INSERT INTO roles_permissions VALUES (5, 17);
INSERT INTO roles_permissions VALUES (5, 11);
INSERT INTO roles_permissions VALUES (5, 12);
INSERT INTO roles_permissions VALUES (5, 14);
INSERT INTO roles_permissions VALUES (5, 13);
INSERT INTO roles_permissions VALUES (5, 15);
INSERT INTO roles_permissions VALUES (5, 16);
INSERT INTO roles_permissions VALUES (5, 19);

SELECT pg_catalog.setval('ad_data_quality_aggregations_id_seq', 1, false);

SELECT pg_catalog.setval('ad_data_quality_stats_id_seq', 1, false);

SELECT pg_catalog.setval('asset_group_collection_entries_id_seq', 1, false);

SELECT pg_catalog.setval('asset_group_collections_id_seq', 1, false);

SELECT pg_catalog.setval('asset_group_selectors_id_seq', 1, false);

SELECT pg_catalog.setval('asset_groups_id_seq', 2, true);

SELECT pg_catalog.setval('audit_logs_id_seq', 1, false);

SELECT pg_catalog.setval('auth_secrets_id_seq', 1, false);

SELECT pg_catalog.setval('azure_data_quality_aggregations_id_seq', 1, false);

SELECT pg_catalog.setval('azure_data_quality_stats_id_seq', 1, false);

SELECT pg_catalog.setval('domain_collection_results_id_seq', 1, false);

SELECT pg_catalog.setval('feature_flags_id_seq', 13, true);

SELECT pg_catalog.setval('file_upload_jobs_id_seq', 1, false);

SELECT pg_catalog.setval('ingest_tasks_id_seq', 1, false);

SELECT pg_catalog.setval('parameters_id_seq', 2, true);

SELECT pg_catalog.setval('permissions_id_seq', 19, true);

SELECT pg_catalog.setval('roles_id_seq', 5, true);

SELECT pg_catalog.setval('saml_providers_id_seq', 1, false);

SELECT pg_catalog.setval('saved_queries_id_seq', 1, false);

SELECT pg_catalog.setval('saved_queries_permissions_id_seq', 1, false);

SELECT pg_catalog.setval('saved_queries_permissions_query_id_seq', 1, false);

SELECT pg_catalog.setval('user_sessions_id_seq', 1, false);

ALTER TABLE ONLY ad_data_quality_aggregations
    ADD CONSTRAINT ad_data_quality_aggregations_pkey PRIMARY KEY (id);

ALTER TABLE ONLY ad_data_quality_stats
    ADD CONSTRAINT ad_data_quality_stats_pkey PRIMARY KEY (id);

ALTER TABLE ONLY analysis_request_switch
    ADD CONSTRAINT analysis_request_switch_pkey PRIMARY KEY (singleton);

ALTER TABLE ONLY asset_group_collection_entries
    ADD CONSTRAINT asset_group_collection_entries_pkey PRIMARY KEY (id);

ALTER TABLE ONLY asset_group_collections
    ADD CONSTRAINT asset_group_collections_pkey PRIMARY KEY (id);

ALTER TABLE ONLY asset_group_selectors
    ADD CONSTRAINT asset_group_selectors_name_assetgroupid_key UNIQUE (name, asset_group_id);

ALTER TABLE ONLY asset_group_selectors
    ADD CONSTRAINT asset_group_selectors_pkey PRIMARY KEY (id);

ALTER TABLE ONLY asset_groups
    ADD CONSTRAINT asset_groups_name_key UNIQUE (name);

ALTER TABLE ONLY asset_groups
    ADD CONSTRAINT asset_groups_pkey PRIMARY KEY (id);

ALTER TABLE ONLY asset_groups
    ADD CONSTRAINT asset_groups_tag_key UNIQUE (tag);

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

ALTER TABLE ONLY datapipe_status
    ADD CONSTRAINT datapipe_status_pkey PRIMARY KEY (singleton);

ALTER TABLE ONLY domain_collection_results
    ADD CONSTRAINT domain_collection_results_pkey PRIMARY KEY (id);

ALTER TABLE ONLY feature_flags
    ADD CONSTRAINT feature_flags_key_key UNIQUE (key);

ALTER TABLE ONLY feature_flags
    ADD CONSTRAINT feature_flags_pkey PRIMARY KEY (id);

ALTER TABLE ONLY file_upload_jobs
    ADD CONSTRAINT file_upload_jobs_pkey PRIMARY KEY (id);

ALTER TABLE ONLY ingest_tasks
    ADD CONSTRAINT ingest_tasks_pkey PRIMARY KEY (id);

ALTER TABLE ONLY installations
    ADD CONSTRAINT installations_pkey PRIMARY KEY (id);

ALTER TABLE ONLY parameters
    ADD CONSTRAINT parameters_key_key UNIQUE (key);

ALTER TABLE ONLY parameters
    ADD CONSTRAINT parameters_pkey PRIMARY KEY (id);

ALTER TABLE ONLY permissions
    ADD CONSTRAINT permissions_authority_name_key UNIQUE (authority, name);

ALTER TABLE ONLY permissions
    ADD CONSTRAINT permissions_pkey PRIMARY KEY (id);

ALTER TABLE ONLY roles
    ADD CONSTRAINT roles_name_key UNIQUE (name);

ALTER TABLE ONLY roles_permissions
    ADD CONSTRAINT roles_permissions_pkey PRIMARY KEY (role_id, permission_id);

ALTER TABLE ONLY roles
    ADD CONSTRAINT roles_pkey PRIMARY KEY (id);

ALTER TABLE ONLY saml_providers
    ADD CONSTRAINT saml_providers_name_key UNIQUE (name);

ALTER TABLE ONLY saml_providers
    ADD CONSTRAINT saml_providers_pkey PRIMARY KEY (id);

ALTER TABLE ONLY saved_queries_permissions
    ADD CONSTRAINT saved_queries_permissions_pkey PRIMARY KEY (id);

ALTER TABLE ONLY saved_queries_permissions
    ADD CONSTRAINT saved_queries_permissions_shared_to_user_id_query_id_key UNIQUE (shared_to_user_id, query_id);

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

ALTER TABLE ONLY saved_queries_permissions
    ADD CONSTRAINT saved_queries_permissions_query_id_fkey FOREIGN KEY (query_id) REFERENCES saved_queries(id) ON DELETE CASCADE;

ALTER TABLE ONLY saved_queries_permissions
    ADD CONSTRAINT saved_queries_permissions_shared_to_user_id_fkey FOREIGN KEY (shared_to_user_id) REFERENCES users(id) ON DELETE CASCADE;
