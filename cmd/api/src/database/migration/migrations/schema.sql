--
-- PostgreSQL database dump
--

-- Dumped from database version 13.2 (Debian 13.2-1.pgdg100+1)
-- Dumped by pg_dump version 13.2 (Debian 13.2-1.pgdg100+1)

SET statement_timeout = 0;
SET lock_timeout = 0;
SET idle_in_transaction_session_timeout = 0;
SET client_encoding = 'UTF8';
SET standard_conforming_strings = on;
SET check_function_bodies = false;
SET xmloption = content;
SET client_min_messages = warning;
SET row_security = off;
SET default_tablespace = '';
SET default_table_access_method = heap;

--
-- Name: ad_data_quality_aggregations; Type: TABLE; Schema: public; Owner: bhe
--

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

--
-- Name: ad_data_quality_aggregations_id_seq; Type: SEQUENCE; Schema: public; Owner: bhe
--

CREATE SEQUENCE IF NOT EXISTS ad_data_quality_aggregations_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;

--
-- Name: ad_data_quality_aggregations_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: bhe
--
-- -- Unsure if necessaryUnsure if necessary
-- -- ALTER SEQUENCE ad_data_quality_aggregations_id_seq OWNED BY ad_data_quality_aggregations.id;


--
-- Name: ad_data_quality_stats; Type: TABLE; Schema: public; Owner: bhe
--

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

--
-- Name: ad_data_quality_stats_id_seq; Type: SEQUENCE; Schema: public; Owner: bhe
--

CREATE SEQUENCE IF NOT EXISTS ad_data_quality_stats_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;

--
-- Name: ad_data_quality_stats_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: bhe
--
-- -- Unsure if necessaryUnsure if necessary
-- -- ALTER SEQUENCE ad_data_quality_stats_id_seq OWNED BY ad_data_quality_stats.id;


--
-- Name: asset_group_collection_entries; Type: TABLE; Schema: public; Owner: bhe
--

CREATE TABLE IF NOT EXISTS asset_group_collection_entries (
    asset_group_collection_id bigint,
    object_id text,
    node_label text,
    properties jsonb,
    id bigint NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone
);

--
-- Name: asset_group_collection_entries_id_seq; Type: SEQUENCE; Schema: public; Owner: bhe
--

CREATE SEQUENCE IF NOT EXISTS asset_group_collection_entries_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;

--
-- Name: asset_group_collection_entries_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: bhe
--
-- Unsure if necessary
-- ALTER SEQUENCE asset_group_collection_entries_id_seq OWNED BY asset_group_collection_entries.id;


--
-- Name: asset_group_collections; Type: TABLE; Schema: public; Owner: bhe
--

CREATE TABLE IF NOT EXISTS asset_group_collections (
    asset_group_id integer,
    id bigint NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone
);

--
-- Name: asset_group_collections_id_seq; Type: SEQUENCE; Schema: public; Owner: bhe
--

CREATE SEQUENCE IF NOT EXISTS asset_group_collections_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;

--
-- Name: asset_group_collections_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: bhe
--
-- Unsure if necessary
-- ALTER SEQUENCE asset_group_collections_id_seq OWNED BY asset_group_collections.id;


--
-- Name: asset_group_selectors; Type: TABLE; Schema: public; Owner: bhe
--

CREATE TABLE IF NOT EXISTS asset_group_selectors (
    asset_group_id integer,
    name text,
    selector text,
    system_selector boolean,
    id integer NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone
);

--
-- Name: asset_group_selectors_id_seq; Type: SEQUENCE; Schema: public; Owner: bhe
--

CREATE SEQUENCE IF NOT EXISTS asset_group_selectors_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;

--
-- Name: asset_group_selectors_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: bhe
--
-- Unsure if necessary
-- ALTER SEQUENCE asset_group_selectors_id_seq OWNED BY asset_group_selectors.id;


--
-- Name: asset_groups; Type: TABLE; Schema: public; Owner: bhe
--

CREATE TABLE IF NOT EXISTS asset_groups (
    name text,
    tag text,
    system_group boolean,
    id integer NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone
);

--
-- Name: asset_groups_id_seq; Type: SEQUENCE; Schema: public; Owner: bhe
--

CREATE SEQUENCE IF NOT EXISTS asset_groups_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;

--
-- Name: asset_groups_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: bhe
--
-- Unsure if necessary
-- ALTER SEQUENCE asset_groups_id_seq OWNED BY asset_groups.id;


--
-- Name: audit_logs; Type: TABLE; Schema: public; Owner: bhe
--

CREATE TABLE IF NOT EXISTS audit_logs (
    id bigint NOT NULL,
    created_at timestamp with time zone,
    actor_id text,
    actor_name text,
    action text,
    fields jsonb,
    request_id text
);

--
-- Name: audit_logs_id_seq; Type: SEQUENCE; Schema: public; Owner: bhe
--

CREATE SEQUENCE IF NOT EXISTS audit_logs_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;

--
-- Name: audit_logs_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: bhe
--
-- Unsure if necessary
-- ALTER SEQUENCE audit_logs_id_seq OWNED BY audit_logs.id;


--
-- Name: auth_secrets; Type: TABLE; Schema: public; Owner: bhe
--

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

--
-- Name: auth_secrets_id_seq; Type: SEQUENCE; Schema: public; Owner: bhe
--

CREATE SEQUENCE IF NOT EXISTS auth_secrets_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;

--
-- Name: auth_secrets_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: bhe
--
-- Unsure if necessary
-- ALTER SEQUENCE auth_secrets_id_seq OWNED BY auth_secrets.id;


--
-- Name: auth_tokens; Type: TABLE; Schema: public; Owner: bhe
--

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

--
-- Name: azure_data_quality_aggregations; Type: TABLE; Schema: public; Owner: bhe
--

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

--
-- Name: azure_data_quality_aggregations_id_seq; Type: SEQUENCE; Schema: public; Owner: bhe
--

CREATE SEQUENCE IF NOT EXISTS azure_data_quality_aggregations_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;

--
-- Name: azure_data_quality_aggregations_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: bhe
--
-- Unsure if necessary
-- ALTER SEQUENCE azure_data_quality_aggregations_id_seq OWNED BY azure_data_quality_aggregations.id;


--
-- Name: azure_data_quality_stats; Type: TABLE; Schema: public; Owner: bhe
--

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

--
-- Name: azure_data_quality_stats_id_seq; Type: SEQUENCE; Schema: public; Owner: bhe
--

CREATE SEQUENCE IF NOT EXISTS azure_data_quality_stats_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;

--
-- Name: azure_data_quality_stats_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: bhe
--
-- Unsure if necessary
-- ALTER SEQUENCE azure_data_quality_stats_id_seq OWNED BY azure_data_quality_stats.id;


--
-- Name: client_ingest_tasks; Type: TABLE; Schema: public; Owner: bhe
--

CREATE TABLE IF NOT EXISTS client_ingest_tasks (
    file_name text,
    request_guid text,
    task_id bigint,
    id bigint NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone
);

--
-- Name: client_ingest_tasks_id_seq; Type: SEQUENCE; Schema: public; Owner: bhe
--

CREATE SEQUENCE IF NOT EXISTS client_ingest_tasks_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;

--
-- Name: client_ingest_tasks_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: bhe
--
-- Unsure if necessary
-- ALTER SEQUENCE client_ingest_tasks_id_seq OWNED BY client_ingest_tasks.id;


--
-- Name: client_scheduled_jobs; Type: TABLE; Schema: public; Owner: bhe
--

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

--
-- Name: client_scheduled_jobs_id_seq; Type: SEQUENCE; Schema: public; Owner: bhe
--

CREATE SEQUENCE IF NOT EXISTS client_scheduled_jobs_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;

--
-- Name: client_scheduled_jobs_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: bhe
--
-- Unsure if necessary
-- ALTER SEQUENCE client_scheduled_jobs_id_seq OWNED BY client_scheduled_jobs.id;


--
-- Name: client_schedules; Type: TABLE; Schema: public; Owner: bhe
--

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

--
-- Name: client_schedules_id_seq; Type: SEQUENCE; Schema: public; Owner: bhe
--

CREATE SEQUENCE IF NOT EXISTS client_schedules_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;

--
-- Name: client_schedules_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: bhe
--
-- Unsure if necessary
-- ALTER SEQUENCE client_schedules_id_seq OWNED BY client_schedules.id;


--
-- Name: clients; Type: TABLE; Schema: public; Owner: bhe
--

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

--
-- Name: domain_collection_results; Type: TABLE; Schema: public; Owner: bhe
--

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

--
-- Name: domain_collection_results_id_seq; Type: SEQUENCE; Schema: public; Owner: bhe
--

CREATE SEQUENCE IF NOT EXISTS domain_collection_results_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;

--
-- Name: domain_collection_results_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: bhe
--
-- Unsure if necessary
-- ALTER SEQUENCE domain_collection_results_id_seq OWNED BY domain_collection_results.id;


--
-- Name: domain_details; Type: TABLE; Schema: public; Owner: bhe
--

CREATE TABLE IF NOT EXISTS domain_details (
    name text,
    object_id text NOT NULL,
    "exists" boolean,
    type text
);

--
-- Name: feature_flags; Type: TABLE; Schema: public; Owner: bhe
--

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

--
-- Name: feature_flags_id_seq; Type: SEQUENCE; Schema: public; Owner: bhe
--

CREATE SEQUENCE IF NOT EXISTS feature_flags_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;

--
-- Name: feature_flags_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: bhe
--
-- Unsure if necessary
-- ALTER SEQUENCE feature_flags_id_seq OWNED BY feature_flags.id;


--
-- Name: file_upload_jobs; Type: TABLE; Schema: public; Owner: bhe
--

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

--
-- Name: file_upload_jobs_id_seq; Type: SEQUENCE; Schema: public; Owner: bhe
--

CREATE SEQUENCE IF NOT EXISTS file_upload_jobs_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;

--
-- Name: file_upload_jobs_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: bhe
--
-- Unsure if necessary
-- ALTER SEQUENCE file_upload_jobs_id_seq OWNED BY file_upload_jobs.id;


--
-- Name: ingest_tasks; Type: TABLE; Schema: public; Owner: bhe
--

CREATE TABLE IF NOT EXISTS ingest_tasks (
    file_name text,
    request_guid text,
    task_id bigint,
    id bigint NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone
);

--
-- Name: ingest_tasks_id_seq; Type: SEQUENCE; Schema: public; Owner: bhe
--

CREATE SEQUENCE IF NOT EXISTS ingest_tasks_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;

--
-- Name: ingest_tasks_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: bhe
--
-- Unsure if necessary
-- ALTER SEQUENCE ingest_tasks_id_seq OWNED BY ingest_tasks.id;


--
-- Name: installations; Type: TABLE; Schema: public; Owner: bhe
--

CREATE TABLE IF NOT EXISTS installations (
    id text NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone
);

--
-- Name: list_findings; Type: TABLE; Schema: public; Owner: bhe
--

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

--
-- Name: list_findings_id_seq; Type: SEQUENCE; Schema: public; Owner: bhe
--

CREATE SEQUENCE IF NOT EXISTS list_findings_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;

--
-- Name: list_findings_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: bhe
--
-- Unsure if necessary
-- ALTER SEQUENCE list_findings_id_seq OWNED BY list_findings.id;


--
-- Name: migrations; Type: TABLE; Schema: public; Owner: bhe
--

CREATE TABLE IF NOT EXISTS migrations (
    id integer NOT NULL,
    updated_at timestamp with time zone,
    major integer,
    minor integer,
    patch integer
);

--
-- Name: migrations_id_seq; Type: SEQUENCE; Schema: public; Owner: bhe
--

CREATE SEQUENCE IF NOT EXISTS migrations_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;

--
-- Name: migrations_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: bhe
--
-- Unsure if necessary
ALTER SEQUENCE migrations_id_seq OWNED BY migrations.id;


--
-- Name: ou_details; Type: TABLE; Schema: public; Owner: bhe
--

CREATE TABLE IF NOT EXISTS ou_details (
    name text,
    object_id text NOT NULL,
    "exists" boolean,
    distinguished_name text,
    type text
);

--
-- Name: parameters; Type: TABLE; Schema: public; Owner: bhe
--

CREATE TABLE IF NOT EXISTS parameters (
    key text,
    name text,
    description text,
    value jsonb,
    id integer NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone
);

--
-- Name: parameters_id_seq; Type: SEQUENCE; Schema: public; Owner: bhe
--

CREATE SEQUENCE IF NOT EXISTS parameters_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;

--
-- Name: parameters_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: bhe
--
-- Unsure if necessary
-- ALTER SEQUENCE parameters_id_seq OWNED BY parameters.id;


--
-- Name: permissions; Type: TABLE; Schema: public; Owner: bhe
--

CREATE TABLE IF NOT EXISTS permissions (
    authority text,
    name text,
    id integer NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone
);

--
-- Name: permissions_id_seq; Type: SEQUENCE; Schema: public; Owner: bhe
--

CREATE SEQUENCE IF NOT EXISTS permissions_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;

--
-- Name: permissions_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: bhe
--
-- Unsure if necessary
-- ALTER SEQUENCE permissions_id_seq OWNED BY permissions.id;


--
-- Name: relationship_findings; Type: TABLE; Schema: public; Owner: bhe
--

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

--
-- Name: relationship_findings_id_seq; Type: SEQUENCE; Schema: public; Owner: bhe
--

CREATE SEQUENCE IF NOT EXISTS relationship_findings_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;

--
-- Name: relationship_findings_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: bhe
--
-- Unsure if necessary
-- ALTER SEQUENCE relationship_findings_id_seq OWNED BY relationship_findings.id;


--
-- Name: risk_counts; Type: TABLE; Schema: public; Owner: bhe
--

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

--
-- Name: risk_counts_id_seq; Type: SEQUENCE; Schema: public; Owner: bhe
--

CREATE SEQUENCE IF NOT EXISTS risk_counts_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;

--
-- Name: risk_counts_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: bhe
--
-- Unsure if necessary
-- ALTER SEQUENCE risk_counts_id_seq OWNED BY risk_counts.id;


--
-- Name: risk_posture_stats; Type: TABLE; Schema: public; Owner: bhe
--

CREATE TABLE IF NOT EXISTS risk_posture_stats (
    domain_sid text,
    exposure_index numeric,
    tier_zero_count bigint,
    critical_risk_count bigint,
    id bigint NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone
);

--
-- Name: risk_posture_stats_id_seq; Type: SEQUENCE; Schema: public; Owner: bhe
--

CREATE SEQUENCE IF NOT EXISTS risk_posture_stats_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;

--
-- Name: risk_posture_stats_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: bhe
--
-- Unsure if necessary
-- ALTER SEQUENCE risk_posture_stats_id_seq OWNED BY risk_posture_stats.id;


--
-- Name: roles; Type: TABLE; Schema: public; Owner: bhe
--

CREATE TABLE IF NOT EXISTS roles (
    name text,
    description text,
    id integer NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone
);

--
-- Name: roles_id_seq; Type: SEQUENCE; Schema: public; Owner: bhe
--

CREATE SEQUENCE IF NOT EXISTS roles_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;

--
-- Name: roles_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: bhe
--
-- Unsure if necessary
-- ALTER SEQUENCE roles_id_seq OWNED BY roles.id;


--
-- Name: roles_permissions; Type: TABLE; Schema: public; Owner: bhe
--

CREATE TABLE IF NOT EXISTS roles_permissions (
    role_id integer NOT NULL,
    permission_id integer NOT NULL
);

--
-- Name: saml_providers; Type: TABLE; Schema: public; Owner: bhe
--

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

--
-- Name: saml_providers_id_seq; Type: SEQUENCE; Schema: public; Owner: bhe
--

CREATE SEQUENCE IF NOT EXISTS saml_providers_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;

--
-- Name: saml_providers_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: bhe
--
-- Unsure if necessary
-- ALTER SEQUENCE saml_providers_id_seq OWNED BY saml_providers.id;


--
-- Name: saved_queries; Type: TABLE; Schema: public; Owner: bhe
--

CREATE TABLE IF NOT EXISTS saved_queries (
    user_id text,
    name text,
    query text,
    id bigint NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone
);

--
-- Name: saved_queries_id_seq; Type: SEQUENCE; Schema: public; Owner: bhe
--

CREATE SEQUENCE IF NOT EXISTS saved_queries_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;

--
-- Name: saved_queries_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: bhe
--
-- Unsure if necessary
-- ALTER SEQUENCE saved_queries_id_seq OWNED BY saved_queries.id;


--
-- Name: user_sessions; Type: TABLE; Schema: public; Owner: bhe
--

CREATE TABLE IF NOT EXISTS user_sessions (
    user_id text,
    auth_provider_type bigint,
    auth_provider_id integer,
    expires_at timestamp with time zone,
    id bigint NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone
);

--
-- Name: user_sessions_id_seq; Type: SEQUENCE; Schema: public; Owner: bhe
--

CREATE SEQUENCE IF NOT EXISTS user_sessions_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;

--
-- Name: user_sessions_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: bhe
--
-- Unsure if necessary
-- ALTER SEQUENCE user_sessions_id_seq OWNED BY user_sessions.id;


--
-- Name: users; Type: TABLE; Schema: public; Owner: bhe
--

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

--
-- Name: users_roles; Type: TABLE; Schema: public; Owner: bhe
--

CREATE TABLE IF NOT EXISTS users_roles (
    user_id text NOT NULL,
    role_id integer NOT NULL
);


ALTER TABLE users_roles OWNER TO bhe;

--
-- Name: ad_data_quality_aggregations id; Type: DEFAULT; Schema: public; Owner: bhe
--

ALTER TABLE ONLY ad_data_quality_aggregations ALTER COLUMN id SET DEFAULT nextval('ad_data_quality_aggregations_id_seq'::regclass);


--
-- Name: ad_data_quality_stats id; Type: DEFAULT; Schema: public; Owner: bhe
--

ALTER TABLE ONLY ad_data_quality_stats ALTER COLUMN id SET DEFAULT nextval('ad_data_quality_stats_id_seq'::regclass);


--
-- Name: asset_group_collection_entries id; Type: DEFAULT; Schema: public; Owner: bhe
--

ALTER TABLE ONLY asset_group_collection_entries ALTER COLUMN id SET DEFAULT nextval('asset_group_collection_entries_id_seq'::regclass);


--
-- Name: asset_group_collections id; Type: DEFAULT; Schema: public; Owner: bhe
--

ALTER TABLE ONLY asset_group_collections ALTER COLUMN id SET DEFAULT nextval('asset_group_collections_id_seq'::regclass);


--
-- Name: asset_group_selectors id; Type: DEFAULT; Schema: public; Owner: bhe
--

ALTER TABLE ONLY asset_group_selectors ALTER COLUMN id SET DEFAULT nextval('asset_group_selectors_id_seq'::regclass);


--
-- Name: asset_groups id; Type: DEFAULT; Schema: public; Owner: bhe
--

ALTER TABLE ONLY asset_groups ALTER COLUMN id SET DEFAULT nextval('asset_groups_id_seq'::regclass);


--
-- Name: audit_logs id; Type: DEFAULT; Schema: public; Owner: bhe
--

ALTER TABLE ONLY audit_logs ALTER COLUMN id SET DEFAULT nextval('audit_logs_id_seq'::regclass);


--
-- Name: auth_secrets id; Type: DEFAULT; Schema: public; Owner: bhe
--

ALTER TABLE ONLY auth_secrets ALTER COLUMN id SET DEFAULT nextval('auth_secrets_id_seq'::regclass);


--
-- Name: azure_data_quality_aggregations id; Type: DEFAULT; Schema: public; Owner: bhe
--

ALTER TABLE ONLY azure_data_quality_aggregations ALTER COLUMN id SET DEFAULT nextval('azure_data_quality_aggregations_id_seq'::regclass);


--
-- Name: azure_data_quality_stats id; Type: DEFAULT; Schema: public; Owner: bhe
--

ALTER TABLE ONLY azure_data_quality_stats ALTER COLUMN id SET DEFAULT nextval('azure_data_quality_stats_id_seq'::regclass);


--
-- Name: client_ingest_tasks id; Type: DEFAULT; Schema: public; Owner: bhe
--

ALTER TABLE ONLY client_ingest_tasks ALTER COLUMN id SET DEFAULT nextval('client_ingest_tasks_id_seq'::regclass);


--
-- Name: client_scheduled_jobs id; Type: DEFAULT; Schema: public; Owner: bhe
--

ALTER TABLE ONLY client_scheduled_jobs ALTER COLUMN id SET DEFAULT nextval('client_scheduled_jobs_id_seq'::regclass);


--
-- Name: client_schedules id; Type: DEFAULT; Schema: public; Owner: bhe
--

ALTER TABLE ONLY client_schedules ALTER COLUMN id SET DEFAULT nextval('client_schedules_id_seq'::regclass);


--
-- Name: domain_collection_results id; Type: DEFAULT; Schema: public; Owner: bhe
--

ALTER TABLE ONLY domain_collection_results ALTER COLUMN id SET DEFAULT nextval('domain_collection_results_id_seq'::regclass);


--
-- Name: feature_flags id; Type: DEFAULT; Schema: public; Owner: bhe
--

ALTER TABLE ONLY feature_flags ALTER COLUMN id SET DEFAULT nextval('feature_flags_id_seq'::regclass);


--
-- Name: file_upload_jobs id; Type: DEFAULT; Schema: public; Owner: bhe
--

ALTER TABLE ONLY file_upload_jobs ALTER COLUMN id SET DEFAULT nextval('file_upload_jobs_id_seq'::regclass);


--
-- Name: ingest_tasks id; Type: DEFAULT; Schema: public; Owner: bhe
--

ALTER TABLE ONLY ingest_tasks ALTER COLUMN id SET DEFAULT nextval('ingest_tasks_id_seq'::regclass);


--
-- Name: list_findings id; Type: DEFAULT; Schema: public; Owner: bhe
--

ALTER TABLE ONLY list_findings ALTER COLUMN id SET DEFAULT nextval('list_findings_id_seq'::regclass);


--
-- Name: migrations id; Type: DEFAULT; Schema: public; Owner: bhe
--

ALTER TABLE ONLY migrations ALTER COLUMN id SET DEFAULT nextval('migrations_id_seq'::regclass);


--
-- Name: parameters id; Type: DEFAULT; Schema: public; Owner: bhe
--

ALTER TABLE ONLY parameters ALTER COLUMN id SET DEFAULT nextval('parameters_id_seq'::regclass);


--
-- Name: permissions id; Type: DEFAULT; Schema: public; Owner: bhe
--

ALTER TABLE ONLY permissions ALTER COLUMN id SET DEFAULT nextval('permissions_id_seq'::regclass);


--
-- Name: relationship_findings id; Type: DEFAULT; Schema: public; Owner: bhe
--

ALTER TABLE ONLY relationship_findings ALTER COLUMN id SET DEFAULT nextval('relationship_findings_id_seq'::regclass);


--
-- Name: risk_counts id; Type: DEFAULT; Schema: public; Owner: bhe
--

ALTER TABLE ONLY risk_counts ALTER COLUMN id SET DEFAULT nextval('risk_counts_id_seq'::regclass);


--
-- Name: risk_posture_stats id; Type: DEFAULT; Schema: public; Owner: bhe
--

ALTER TABLE ONLY risk_posture_stats ALTER COLUMN id SET DEFAULT nextval('risk_posture_stats_id_seq'::regclass);


--
-- Name: roles id; Type: DEFAULT; Schema: public; Owner: bhe
--

ALTER TABLE ONLY roles ALTER COLUMN id SET DEFAULT nextval('roles_id_seq'::regclass);


--
-- Name: saml_providers id; Type: DEFAULT; Schema: public; Owner: bhe
--

ALTER TABLE ONLY saml_providers ALTER COLUMN id SET DEFAULT nextval('saml_providers_id_seq'::regclass);


--
-- Name: saved_queries id; Type: DEFAULT; Schema: public; Owner: bhe
--

ALTER TABLE ONLY saved_queries ALTER COLUMN id SET DEFAULT nextval('saved_queries_id_seq'::regclass);


--
-- Name: user_sessions id; Type: DEFAULT; Schema: public; Owner: bhe
--

ALTER TABLE ONLY user_sessions ALTER COLUMN id SET DEFAULT nextval('user_sessions_id_seq'::regclass);


--
-- Name: ad_data_quality_aggregations ad_data_quality_aggregations_pkey; Type: CONSTRAINT; Schema: public; Owner: bhe
--

ALTER TABLE ONLY ad_data_quality_aggregations
    ADD CONSTRAINT ad_data_quality_aggregations_pkey PRIMARY KEY (id);


--
-- Name: ad_data_quality_stats ad_data_quality_stats_pkey; Type: CONSTRAINT; Schema: public; Owner: bhe
--

ALTER TABLE ONLY ad_data_quality_stats
    ADD CONSTRAINT ad_data_quality_stats_pkey PRIMARY KEY (id);


--
-- Name: asset_group_collection_entries asset_group_collection_entries_pkey; Type: CONSTRAINT; Schema: public; Owner: bhe
--

ALTER TABLE ONLY asset_group_collection_entries
    ADD CONSTRAINT asset_group_collection_entries_pkey PRIMARY KEY (id);


--
-- Name: asset_group_collections asset_group_collections_pkey; Type: CONSTRAINT; Schema: public; Owner: bhe
--

ALTER TABLE ONLY asset_group_collections
    ADD CONSTRAINT asset_group_collections_pkey PRIMARY KEY (id);


--
-- Name: asset_group_selectors asset_group_selectors_name_key; Type: CONSTRAINT; Schema: public; Owner: bhe
--

ALTER TABLE ONLY asset_group_selectors
    ADD CONSTRAINT asset_group_selectors_name_key UNIQUE (name);


--
-- Name: asset_group_selectors asset_group_selectors_pkey; Type: CONSTRAINT; Schema: public; Owner: bhe
--

ALTER TABLE ONLY asset_group_selectors
    ADD CONSTRAINT asset_group_selectors_pkey PRIMARY KEY (id);


--
-- Name: asset_groups asset_groups_pkey; Type: CONSTRAINT; Schema: public; Owner: bhe
--

ALTER TABLE ONLY asset_groups
    ADD CONSTRAINT asset_groups_pkey PRIMARY KEY (id);


--
-- Name: audit_logs audit_logs_pkey; Type: CONSTRAINT; Schema: public; Owner: bhe
--

ALTER TABLE ONLY audit_logs
    ADD CONSTRAINT audit_logs_pkey PRIMARY KEY (id);


--
-- Name: auth_secrets auth_secrets_pkey; Type: CONSTRAINT; Schema: public; Owner: bhe
--

ALTER TABLE ONLY auth_secrets
    ADD CONSTRAINT auth_secrets_pkey PRIMARY KEY (id);


--
-- Name: auth_tokens auth_tokens_pkey; Type: CONSTRAINT; Schema: public; Owner: bhe
--

ALTER TABLE ONLY auth_tokens
    ADD CONSTRAINT auth_tokens_pkey PRIMARY KEY (id);


--
-- Name: azure_data_quality_aggregations azure_data_quality_aggregations_pkey; Type: CONSTRAINT; Schema: public; Owner: bhe
--

ALTER TABLE ONLY azure_data_quality_aggregations
    ADD CONSTRAINT azure_data_quality_aggregations_pkey PRIMARY KEY (id);


--
-- Name: azure_data_quality_stats azure_data_quality_stats_pkey; Type: CONSTRAINT; Schema: public; Owner: bhe
--

ALTER TABLE ONLY azure_data_quality_stats
    ADD CONSTRAINT azure_data_quality_stats_pkey PRIMARY KEY (id);


--
-- Name: client_ingest_tasks client_ingest_tasks_pkey; Type: CONSTRAINT; Schema: public; Owner: bhe
--

ALTER TABLE ONLY client_ingest_tasks
    ADD CONSTRAINT client_ingest_tasks_pkey PRIMARY KEY (id);


--
-- Name: client_scheduled_jobs client_scheduled_jobs_pkey; Type: CONSTRAINT; Schema: public; Owner: bhe
--

ALTER TABLE ONLY client_scheduled_jobs
    ADD CONSTRAINT client_scheduled_jobs_pkey PRIMARY KEY (id);


--
-- Name: client_schedules client_schedules_pkey; Type: CONSTRAINT; Schema: public; Owner: bhe
--

ALTER TABLE ONLY client_schedules
    ADD CONSTRAINT client_schedules_pkey PRIMARY KEY (id);


--
-- Name: clients clients_pkey; Type: CONSTRAINT; Schema: public; Owner: bhe
--

ALTER TABLE ONLY clients
    ADD CONSTRAINT clients_pkey PRIMARY KEY (id);


--
-- Name: domain_collection_results domain_collection_results_pkey; Type: CONSTRAINT; Schema: public; Owner: bhe
--

ALTER TABLE ONLY domain_collection_results
    ADD CONSTRAINT domain_collection_results_pkey PRIMARY KEY (id);


--
-- Name: domain_details domain_details_pkey; Type: CONSTRAINT; Schema: public; Owner: bhe
--

ALTER TABLE ONLY domain_details
    ADD CONSTRAINT domain_details_pkey PRIMARY KEY (object_id);


--
-- Name: feature_flags feature_flags_key_key; Type: CONSTRAINT; Schema: public; Owner: bhe
--

ALTER TABLE ONLY feature_flags
    ADD CONSTRAINT feature_flags_key_key UNIQUE (key);


--
-- Name: feature_flags feature_flags_pkey; Type: CONSTRAINT; Schema: public; Owner: bhe
--

ALTER TABLE ONLY feature_flags
    ADD CONSTRAINT feature_flags_pkey PRIMARY KEY (id);


--
-- Name: file_upload_jobs file_upload_jobs_pkey; Type: CONSTRAINT; Schema: public; Owner: bhe
--

ALTER TABLE ONLY file_upload_jobs
    ADD CONSTRAINT file_upload_jobs_pkey PRIMARY KEY (id);


--
-- Name: asset_group_selectors idx_asset_group_selectors_name; Type: CONSTRAINT; Schema: public; Owner: bhe
--

ALTER TABLE ONLY asset_group_selectors
    ADD CONSTRAINT idx_asset_group_selectors_name UNIQUE (name);


--
-- Name: ingest_tasks ingest_tasks_pkey; Type: CONSTRAINT; Schema: public; Owner: bhe
--

ALTER TABLE ONLY ingest_tasks
    ADD CONSTRAINT ingest_tasks_pkey PRIMARY KEY (id);


--
-- Name: installations installations_pkey; Type: CONSTRAINT; Schema: public; Owner: bhe
--

ALTER TABLE ONLY installations
    ADD CONSTRAINT installations_pkey PRIMARY KEY (id);


--
-- Name: list_findings list_findings_pkey; Type: CONSTRAINT; Schema: public; Owner: bhe
--

ALTER TABLE ONLY list_findings
    ADD CONSTRAINT list_findings_pkey PRIMARY KEY (id);


--
-- Name: migrations migrations_pkey; Type: CONSTRAINT; Schema: public; Owner: bhe
--

ALTER TABLE ONLY migrations
    ADD CONSTRAINT migrations_pkey PRIMARY KEY (id);


--
-- Name: ou_details ou_details_pkey; Type: CONSTRAINT; Schema: public; Owner: bhe
--

ALTER TABLE ONLY ou_details
    ADD CONSTRAINT ou_details_pkey PRIMARY KEY (object_id);


--
-- Name: parameters parameters_key_key; Type: CONSTRAINT; Schema: public; Owner: bhe
--

ALTER TABLE ONLY parameters
    ADD CONSTRAINT parameters_key_key UNIQUE (key);


--
-- Name: parameters parameters_pkey; Type: CONSTRAINT; Schema: public; Owner: bhe
--

ALTER TABLE ONLY parameters
    ADD CONSTRAINT parameters_pkey PRIMARY KEY (id);


--
-- Name: permissions permissions_pkey; Type: CONSTRAINT; Schema: public; Owner: bhe
--

ALTER TABLE ONLY permissions
    ADD CONSTRAINT permissions_pkey PRIMARY KEY (id);


--
-- Name: relationship_findings relationship_findings_pkey; Type: CONSTRAINT; Schema: public; Owner: bhe
--

ALTER TABLE ONLY relationship_findings
    ADD CONSTRAINT relationship_findings_pkey PRIMARY KEY (id);


--
-- Name: risk_counts risk_counts_pkey; Type: CONSTRAINT; Schema: public; Owner: bhe
--

ALTER TABLE ONLY risk_counts
    ADD CONSTRAINT risk_counts_pkey PRIMARY KEY (id);


--
-- Name: risk_posture_stats risk_posture_stats_pkey; Type: CONSTRAINT; Schema: public; Owner: bhe
--

ALTER TABLE ONLY risk_posture_stats
    ADD CONSTRAINT risk_posture_stats_pkey PRIMARY KEY (id);


--
-- Name: roles_permissions roles_permissions_pkey; Type: CONSTRAINT; Schema: public; Owner: bhe
--

ALTER TABLE ONLY roles_permissions
    ADD CONSTRAINT roles_permissions_pkey PRIMARY KEY (role_id, permission_id);


--
-- Name: roles roles_pkey; Type: CONSTRAINT; Schema: public; Owner: bhe
--

ALTER TABLE ONLY roles
    ADD CONSTRAINT roles_pkey PRIMARY KEY (id);


--
-- Name: saml_providers saml_providers_name_key; Type: CONSTRAINT; Schema: public; Owner: bhe
--

ALTER TABLE ONLY saml_providers
    ADD CONSTRAINT saml_providers_name_key UNIQUE (name);


--
-- Name: saml_providers saml_providers_pkey; Type: CONSTRAINT; Schema: public; Owner: bhe
--

ALTER TABLE ONLY saml_providers
    ADD CONSTRAINT saml_providers_pkey PRIMARY KEY (id);


--
-- Name: saved_queries saved_queries_pkey; Type: CONSTRAINT; Schema: public; Owner: bhe
--

ALTER TABLE ONLY saved_queries
    ADD CONSTRAINT saved_queries_pkey PRIMARY KEY (id);


--
-- Name: user_sessions user_sessions_pkey; Type: CONSTRAINT; Schema: public; Owner: bhe
--

ALTER TABLE ONLY user_sessions
    ADD CONSTRAINT user_sessions_pkey PRIMARY KEY (id);


--
-- Name: users users_pkey; Type: CONSTRAINT; Schema: public; Owner: bhe
--

ALTER TABLE ONLY users
    ADD CONSTRAINT users_pkey PRIMARY KEY (id);


--
-- Name: users users_principal_name_key; Type: CONSTRAINT; Schema: public; Owner: bhe
--

ALTER TABLE ONLY users
    ADD CONSTRAINT users_principal_name_key UNIQUE (principal_name);


--
-- Name: users_roles users_roles_pkey; Type: CONSTRAINT; Schema: public; Owner: bhe
--

ALTER TABLE ONLY users_roles
    ADD CONSTRAINT users_roles_pkey PRIMARY KEY (user_id, role_id);


--
-- Name: idx_ad_data_quality_aggregations_run_id; Type: INDEX; Schema: public; Owner: bhe
--

CREATE INDEX IF NOT EXISTS idx_ad_data_quality_aggregations_run_id ON ad_data_quality_aggregations USING btree (run_id);


--
-- Name: idx_ad_data_quality_stats_run_id; Type: INDEX; Schema: public; Owner: bhe
--

CREATE INDEX IF NOT EXISTS idx_ad_data_quality_stats_run_id ON ad_data_quality_stats USING btree (run_id);


--
-- Name: idx_audit_logs_action; Type: INDEX; Schema: public; Owner: bhe
--

CREATE INDEX IF NOT EXISTS idx_audit_logs_action ON audit_logs USING btree (action);


--
-- Name: idx_audit_logs_actor_id; Type: INDEX; Schema: public; Owner: bhe
--

CREATE INDEX IF NOT EXISTS idx_audit_logs_actor_id ON audit_logs USING btree (actor_id);


--
-- Name: idx_audit_logs_created_at; Type: INDEX; Schema: public; Owner: bhe
--

CREATE INDEX IF NOT EXISTS idx_audit_logs_created_at ON audit_logs USING btree (created_at);


--
-- Name: idx_azure_data_quality_aggregations_run_id; Type: INDEX; Schema: public; Owner: bhe
--

CREATE INDEX IF NOT EXISTS idx_azure_data_quality_aggregations_run_id ON azure_data_quality_aggregations USING btree (run_id);


--
-- Name: idx_azure_data_quality_stats_run_id; Type: INDEX; Schema: public; Owner: bhe
--

CREATE INDEX IF NOT EXISTS idx_azure_data_quality_stats_run_id ON azure_data_quality_stats USING btree (run_id);


--
-- Name: idx_client_schedules_deleted_at; Type: INDEX; Schema: public; Owner: bhe
--

CREATE INDEX IF NOT EXISTS idx_client_schedules_deleted_at ON client_schedules USING btree (deleted_at);


--
-- Name: idx_clients_deleted_at; Type: INDEX; Schema: public; Owner: bhe
--

CREATE INDEX IF NOT EXISTS idx_clients_deleted_at ON clients USING btree (deleted_at);


--
-- Name: idx_list_findings_domain_s_id; Type: INDEX; Schema: public; Owner: bhe
--

CREATE INDEX IF NOT EXISTS idx_list_findings_domain_s_id ON list_findings USING btree (domain_sid);


--
-- Name: idx_list_findings_finding; Type: INDEX; Schema: public; Owner: bhe
--

CREATE INDEX IF NOT EXISTS idx_list_findings_finding ON list_findings USING btree (finding);


--
-- Name: idx_list_findings_principal; Type: INDEX; Schema: public; Owner: bhe
--

CREATE INDEX IF NOT EXISTS idx_list_findings_principal ON list_findings USING btree (principal);


--
-- Name: idx_relationship_findings_domain_s_id; Type: INDEX; Schema: public; Owner: bhe
--

CREATE INDEX IF NOT EXISTS idx_relationship_findings_domain_s_id ON relationship_findings USING btree (domain_sid);


--
-- Name: idx_relationship_findings_finding; Type: INDEX; Schema: public; Owner: bhe
--

CREATE INDEX IF NOT EXISTS idx_relationship_findings_finding ON relationship_findings USING btree (finding);


--
-- Name: idx_relationship_findings_principals_hash; Type: INDEX; Schema: public; Owner: bhe
--

CREATE INDEX IF NOT EXISTS idx_relationship_findings_principals_hash ON relationship_findings USING btree (principals_hash);


--
-- Name: idx_risk_counts_domain_s_id; Type: INDEX; Schema: public; Owner: bhe
--

CREATE INDEX IF NOT EXISTS idx_risk_counts_domain_s_id ON risk_counts USING btree (domain_sid);


--
-- Name: idx_risk_counts_finding; Type: INDEX; Schema: public; Owner: bhe
--

CREATE INDEX IF NOT EXISTS idx_risk_counts_finding ON risk_counts USING btree (finding);


--
-- Name: idx_saml_providers_name; Type: INDEX; Schema: public; Owner: bhe
--

CREATE INDEX IF NOT EXISTS idx_saml_providers_name ON saml_providers USING btree (name);


--
-- Name: idx_saved_queries_composite_index; Type: INDEX; Schema: public; Owner: bhe
--

CREATE UNIQUE INDEX IF NOT EXISTS idx_saved_queries_composite_index ON saved_queries USING btree (user_id, name);


--
-- Name: idx_users_principal_name; Type: INDEX; Schema: public; Owner: bhe
--

CREATE INDEX IF NOT EXISTS idx_users_principal_name ON users USING btree (principal_name);


--
-- Name: asset_group_collection_entries fk_asset_group_collections_entries; Type: FK CONSTRAINT; Schema: public; Owner: bhe
--

ALTER TABLE ONLY asset_group_collection_entries
    ADD CONSTRAINT fk_asset_group_collections_entries FOREIGN KEY (asset_group_collection_id) REFERENCES asset_group_collections(id) ON DELETE CASCADE;


--
-- Name: asset_group_collections fk_asset_groups_collections; Type: FK CONSTRAINT; Schema: public; Owner: bhe
--

ALTER TABLE ONLY asset_group_collections
    ADD CONSTRAINT fk_asset_groups_collections FOREIGN KEY (asset_group_id) REFERENCES asset_groups(id) ON DELETE CASCADE;


--
-- Name: asset_group_selectors fk_asset_groups_selectors; Type: FK CONSTRAINT; Schema: public; Owner: bhe
--

ALTER TABLE ONLY asset_group_selectors
    ADD CONSTRAINT fk_asset_groups_selectors FOREIGN KEY (asset_group_id) REFERENCES asset_groups(id) ON DELETE CASCADE;


--
-- Name: client_schedules fk_clients_schedules; Type: FK CONSTRAINT; Schema: public; Owner: bhe
--

ALTER TABLE ONLY client_schedules
    ADD CONSTRAINT fk_clients_schedules FOREIGN KEY (client_id) REFERENCES clients(id);


--
-- Name: file_upload_jobs fk_file_upload_jobs_user; Type: FK CONSTRAINT; Schema: public; Owner: bhe
--

ALTER TABLE ONLY file_upload_jobs
    ADD CONSTRAINT fk_file_upload_jobs_user FOREIGN KEY (user_id) REFERENCES users(id);


--
-- Name: roles_permissions fk_roles_permissions_permission; Type: FK CONSTRAINT; Schema: public; Owner: bhe
--

ALTER TABLE ONLY roles_permissions
    ADD CONSTRAINT fk_roles_permissions_permission FOREIGN KEY (permission_id) REFERENCES permissions(id);


--
-- Name: roles_permissions fk_roles_permissions_role; Type: FK CONSTRAINT; Schema: public; Owner: bhe
--

ALTER TABLE ONLY roles_permissions
    ADD CONSTRAINT fk_roles_permissions_role FOREIGN KEY (role_id) REFERENCES roles(id);


--
-- Name: user_sessions fk_user_sessions_user; Type: FK CONSTRAINT; Schema: public; Owner: bhe
--

ALTER TABLE ONLY user_sessions
    ADD CONSTRAINT fk_user_sessions_user FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE;


--
-- Name: auth_secrets fk_users_auth_secret; Type: FK CONSTRAINT; Schema: public; Owner: bhe
--

ALTER TABLE ONLY auth_secrets
    ADD CONSTRAINT fk_users_auth_secret FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE;


--
-- Name: auth_tokens fk_users_auth_tokens; Type: FK CONSTRAINT; Schema: public; Owner: bhe
--

ALTER TABLE ONLY auth_tokens
    ADD CONSTRAINT fk_users_auth_tokens FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE;


--
-- Name: users_roles fk_users_roles_role; Type: FK CONSTRAINT; Schema: public; Owner: bhe
--

ALTER TABLE ONLY users_roles
    ADD CONSTRAINT fk_users_roles_role FOREIGN KEY (role_id) REFERENCES roles(id);


--
-- Name: users_roles fk_users_roles_user; Type: FK CONSTRAINT; Schema: public; Owner: bhe
--

ALTER TABLE ONLY users_roles
    ADD CONSTRAINT fk_users_roles_user FOREIGN KEY (user_id) REFERENCES users(id);


--
-- Name: users fk_users_saml_provider; Type: FK CONSTRAINT; Schema: public; Owner: bhe
--

ALTER TABLE ONLY users
    ADD CONSTRAINT fk_users_saml_provider FOREIGN KEY (saml_provider_id) REFERENCES saml_providers(id);

--
-- PostgreSQL database dump complete
--

