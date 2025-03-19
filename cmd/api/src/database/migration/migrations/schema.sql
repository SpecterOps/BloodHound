--
-- PostgreSQL database dump
--

-- Dumped from database version 13.2 (Debian 13.2-1.pgdg100+1)
-- Dumped by pg_dump version 14.17 (Homebrew)

SET statement_timeout = 0;
SET lock_timeout = 0;
SET idle_in_transaction_session_timeout = 0;
SET client_encoding = 'UTF8';
SET standard_conforming_strings = on;
SELECT pg_catalog.set_config('search_path', '', false);
SET check_function_bodies = false;
SET xmloption = content;
SET client_min_messages = warning;
SET row_security = off;

--
-- Name: pg_trgm; Type: EXTENSION; Schema: -; Owner: -
--

CREATE EXTENSION IF NOT EXISTS pg_trgm WITH SCHEMA public;


--
-- Name: EXTENSION pg_trgm; Type: COMMENT; Schema: -; Owner: 
--

COMMENT ON EXTENSION pg_trgm IS 'text similarity measurement and index searching based on trigrams';


SET default_tablespace = '';

SET default_table_access_method = heap;

--
-- Name: ad_data_quality_aggregations; Type: TABLE; Schema: public; Owner: bloodhound
--

CREATE TABLE public.ad_data_quality_aggregations (
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


ALTER TABLE public.ad_data_quality_aggregations OWNER TO bloodhound;

--
-- Name: ad_data_quality_aggregations_id_seq; Type: SEQUENCE; Schema: public; Owner: bloodhound
--

CREATE SEQUENCE public.ad_data_quality_aggregations_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.ad_data_quality_aggregations_id_seq OWNER TO bloodhound;

--
-- Name: ad_data_quality_aggregations_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: bloodhound
--

ALTER SEQUENCE public.ad_data_quality_aggregations_id_seq OWNED BY public.ad_data_quality_aggregations.id;


--
-- Name: ad_data_quality_stats; Type: TABLE; Schema: public; Owner: bloodhound
--

CREATE TABLE public.ad_data_quality_stats (
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


ALTER TABLE public.ad_data_quality_stats OWNER TO bloodhound;

--
-- Name: ad_data_quality_stats_id_seq; Type: SEQUENCE; Schema: public; Owner: bloodhound
--

CREATE SEQUENCE public.ad_data_quality_stats_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.ad_data_quality_stats_id_seq OWNER TO bloodhound;

--
-- Name: ad_data_quality_stats_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: bloodhound
--

ALTER SEQUENCE public.ad_data_quality_stats_id_seq OWNED BY public.ad_data_quality_stats.id;


--
-- Name: analysis_request_switch; Type: TABLE; Schema: public; Owner: bloodhound
--

CREATE TABLE public.analysis_request_switch (
    singleton boolean DEFAULT true NOT NULL,
    request_type text NOT NULL,
    requested_by text NOT NULL,
    requested_at timestamp with time zone NOT NULL,
    CONSTRAINT singleton_uni CHECK (singleton)
);


ALTER TABLE public.analysis_request_switch OWNER TO bloodhound;

--
-- Name: asset_group_collection_entries; Type: TABLE; Schema: public; Owner: bloodhound
--

CREATE TABLE public.asset_group_collection_entries (
    asset_group_collection_id bigint,
    object_id text,
    node_label text,
    properties jsonb,
    id bigint NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone
);


ALTER TABLE public.asset_group_collection_entries OWNER TO bloodhound;

--
-- Name: asset_group_collection_entries_id_seq; Type: SEQUENCE; Schema: public; Owner: bloodhound
--

CREATE SEQUENCE public.asset_group_collection_entries_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.asset_group_collection_entries_id_seq OWNER TO bloodhound;

--
-- Name: asset_group_collection_entries_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: bloodhound
--

ALTER SEQUENCE public.asset_group_collection_entries_id_seq OWNED BY public.asset_group_collection_entries.id;


--
-- Name: asset_group_collections; Type: TABLE; Schema: public; Owner: bloodhound
--

CREATE TABLE public.asset_group_collections (
    asset_group_id integer,
    id bigint NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone
);


ALTER TABLE public.asset_group_collections OWNER TO bloodhound;

--
-- Name: asset_group_collections_id_seq; Type: SEQUENCE; Schema: public; Owner: bloodhound
--

CREATE SEQUENCE public.asset_group_collections_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.asset_group_collections_id_seq OWNER TO bloodhound;

--
-- Name: asset_group_collections_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: bloodhound
--

ALTER SEQUENCE public.asset_group_collections_id_seq OWNED BY public.asset_group_collections.id;


--
-- Name: asset_group_selectors; Type: TABLE; Schema: public; Owner: bloodhound
--

CREATE TABLE public.asset_group_selectors (
    asset_group_id integer,
    name text,
    selector text,
    system_selector boolean,
    id integer NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone
);


ALTER TABLE public.asset_group_selectors OWNER TO bloodhound;

--
-- Name: asset_group_selectors_id_seq; Type: SEQUENCE; Schema: public; Owner: bloodhound
--

CREATE SEQUENCE public.asset_group_selectors_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.asset_group_selectors_id_seq OWNER TO bloodhound;

--
-- Name: asset_group_selectors_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: bloodhound
--

ALTER SEQUENCE public.asset_group_selectors_id_seq OWNED BY public.asset_group_selectors.id;


--
-- Name: asset_groups; Type: TABLE; Schema: public; Owner: bloodhound
--

CREATE TABLE public.asset_groups (
    name text NOT NULL,
    tag text NOT NULL,
    system_group boolean,
    id integer NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone
);


ALTER TABLE public.asset_groups OWNER TO bloodhound;

--
-- Name: asset_groups_id_seq; Type: SEQUENCE; Schema: public; Owner: bloodhound
--

CREATE SEQUENCE public.asset_groups_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.asset_groups_id_seq OWNER TO bloodhound;

--
-- Name: asset_groups_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: bloodhound
--

ALTER SEQUENCE public.asset_groups_id_seq OWNED BY public.asset_groups.id;


--
-- Name: audit_logs; Type: TABLE; Schema: public; Owner: bloodhound
--

CREATE TABLE public.audit_logs (
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


ALTER TABLE public.audit_logs OWNER TO bloodhound;

--
-- Name: audit_logs_id_seq; Type: SEQUENCE; Schema: public; Owner: bloodhound
--

CREATE SEQUENCE public.audit_logs_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.audit_logs_id_seq OWNER TO bloodhound;

--
-- Name: audit_logs_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: bloodhound
--

ALTER SEQUENCE public.audit_logs_id_seq OWNED BY public.audit_logs.id;


--
-- Name: auth_secrets; Type: TABLE; Schema: public; Owner: bloodhound
--

CREATE TABLE public.auth_secrets (
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


ALTER TABLE public.auth_secrets OWNER TO bloodhound;

--
-- Name: auth_secrets_id_seq; Type: SEQUENCE; Schema: public; Owner: bloodhound
--

CREATE SEQUENCE public.auth_secrets_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.auth_secrets_id_seq OWNER TO bloodhound;

--
-- Name: auth_secrets_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: bloodhound
--

ALTER SEQUENCE public.auth_secrets_id_seq OWNED BY public.auth_secrets.id;


--
-- Name: auth_tokens; Type: TABLE; Schema: public; Owner: bloodhound
--

CREATE TABLE public.auth_tokens (
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


ALTER TABLE public.auth_tokens OWNER TO bloodhound;

--
-- Name: azure_data_quality_aggregations; Type: TABLE; Schema: public; Owner: bloodhound
--

CREATE TABLE public.azure_data_quality_aggregations (
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


ALTER TABLE public.azure_data_quality_aggregations OWNER TO bloodhound;

--
-- Name: azure_data_quality_aggregations_id_seq; Type: SEQUENCE; Schema: public; Owner: bloodhound
--

CREATE SEQUENCE public.azure_data_quality_aggregations_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.azure_data_quality_aggregations_id_seq OWNER TO bloodhound;

--
-- Name: azure_data_quality_aggregations_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: bloodhound
--

ALTER SEQUENCE public.azure_data_quality_aggregations_id_seq OWNED BY public.azure_data_quality_aggregations.id;


--
-- Name: azure_data_quality_stats; Type: TABLE; Schema: public; Owner: bloodhound
--

CREATE TABLE public.azure_data_quality_stats (
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


ALTER TABLE public.azure_data_quality_stats OWNER TO bloodhound;

--
-- Name: azure_data_quality_stats_id_seq; Type: SEQUENCE; Schema: public; Owner: bloodhound
--

CREATE SEQUENCE public.azure_data_quality_stats_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.azure_data_quality_stats_id_seq OWNER TO bloodhound;

--
-- Name: azure_data_quality_stats_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: bloodhound
--

ALTER SEQUENCE public.azure_data_quality_stats_id_seq OWNED BY public.azure_data_quality_stats.id;


--
-- Name: database_switch; Type: TABLE; Schema: public; Owner: bloodhound
--

CREATE TABLE public.database_switch (
    driver text NOT NULL
);


ALTER TABLE public.database_switch OWNER TO bloodhound;

--
-- Name: datapipe_status; Type: TABLE; Schema: public; Owner: bloodhound
--

CREATE TABLE public.datapipe_status (
    singleton boolean DEFAULT true NOT NULL,
    status text NOT NULL,
    updated_at timestamp with time zone NOT NULL,
    last_complete_analysis_at timestamp with time zone,
    CONSTRAINT singleton_uni CHECK (singleton)
);


ALTER TABLE public.datapipe_status OWNER TO bloodhound;

--
-- Name: domain_collection_results; Type: TABLE; Schema: public; Owner: bloodhound
--

CREATE TABLE public.domain_collection_results (
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


ALTER TABLE public.domain_collection_results OWNER TO bloodhound;

--
-- Name: domain_collection_results_id_seq; Type: SEQUENCE; Schema: public; Owner: bloodhound
--

CREATE SEQUENCE public.domain_collection_results_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.domain_collection_results_id_seq OWNER TO bloodhound;

--
-- Name: domain_collection_results_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: bloodhound
--

ALTER SEQUENCE public.domain_collection_results_id_seq OWNED BY public.domain_collection_results.id;


--
-- Name: feature_flags; Type: TABLE; Schema: public; Owner: bloodhound
--

CREATE TABLE public.feature_flags (
    id integer NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    key text,
    name text,
    description text,
    enabled boolean,
    user_updatable boolean
);


ALTER TABLE public.feature_flags OWNER TO bloodhound;

--
-- Name: feature_flags_id_seq; Type: SEQUENCE; Schema: public; Owner: bloodhound
--

CREATE SEQUENCE public.feature_flags_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.feature_flags_id_seq OWNER TO bloodhound;

--
-- Name: feature_flags_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: bloodhound
--

ALTER SEQUENCE public.feature_flags_id_seq OWNED BY public.feature_flags.id;


--
-- Name: file_upload_jobs; Type: TABLE; Schema: public; Owner: bloodhound
--

CREATE TABLE public.file_upload_jobs (
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


ALTER TABLE public.file_upload_jobs OWNER TO bloodhound;

--
-- Name: file_upload_jobs_id_seq; Type: SEQUENCE; Schema: public; Owner: bloodhound
--

CREATE SEQUENCE public.file_upload_jobs_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.file_upload_jobs_id_seq OWNER TO bloodhound;

--
-- Name: file_upload_jobs_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: bloodhound
--

ALTER SEQUENCE public.file_upload_jobs_id_seq OWNED BY public.file_upload_jobs.id;


--
-- Name: ingest_tasks; Type: TABLE; Schema: public; Owner: bloodhound
--

CREATE TABLE public.ingest_tasks (
    file_name text,
    request_guid text,
    task_id bigint,
    id bigint NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    file_type integer DEFAULT 0
);


ALTER TABLE public.ingest_tasks OWNER TO bloodhound;

--
-- Name: ingest_tasks_id_seq; Type: SEQUENCE; Schema: public; Owner: bloodhound
--

CREATE SEQUENCE public.ingest_tasks_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.ingest_tasks_id_seq OWNER TO bloodhound;

--
-- Name: ingest_tasks_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: bloodhound
--

ALTER SEQUENCE public.ingest_tasks_id_seq OWNED BY public.ingest_tasks.id;


--
-- Name: installations; Type: TABLE; Schema: public; Owner: bloodhound
--

CREATE TABLE public.installations (
    id text NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone
);


ALTER TABLE public.installations OWNER TO bloodhound;

--
-- Name: migrations; Type: TABLE; Schema: public; Owner: bloodhound
--

CREATE TABLE public.migrations (
    id integer NOT NULL,
    updated_at timestamp with time zone,
    major integer,
    minor integer,
    patch integer
);


ALTER TABLE public.migrations OWNER TO bloodhound;

--
-- Name: migrations_id_seq; Type: SEQUENCE; Schema: public; Owner: bloodhound
--

CREATE SEQUENCE public.migrations_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.migrations_id_seq OWNER TO bloodhound;

--
-- Name: migrations_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: bloodhound
--

ALTER SEQUENCE public.migrations_id_seq OWNED BY public.migrations.id;


--
-- Name: parameters; Type: TABLE; Schema: public; Owner: bloodhound
--

CREATE TABLE public.parameters (
    key text,
    name text,
    description text,
    value jsonb,
    id integer NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone
);


ALTER TABLE public.parameters OWNER TO bloodhound;

--
-- Name: parameters_id_seq; Type: SEQUENCE; Schema: public; Owner: bloodhound
--

CREATE SEQUENCE public.parameters_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.parameters_id_seq OWNER TO bloodhound;

--
-- Name: parameters_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: bloodhound
--

ALTER SEQUENCE public.parameters_id_seq OWNED BY public.parameters.id;


--
-- Name: permissions; Type: TABLE; Schema: public; Owner: bloodhound
--

CREATE TABLE public.permissions (
    authority text,
    name text,
    id integer NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone
);


ALTER TABLE public.permissions OWNER TO bloodhound;

--
-- Name: permissions_id_seq; Type: SEQUENCE; Schema: public; Owner: bloodhound
--

CREATE SEQUENCE public.permissions_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.permissions_id_seq OWNER TO bloodhound;

--
-- Name: permissions_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: bloodhound
--

ALTER SEQUENCE public.permissions_id_seq OWNED BY public.permissions.id;


--
-- Name: roles; Type: TABLE; Schema: public; Owner: bloodhound
--

CREATE TABLE public.roles (
    name text,
    description text,
    id integer NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone
);


ALTER TABLE public.roles OWNER TO bloodhound;

--
-- Name: roles_id_seq; Type: SEQUENCE; Schema: public; Owner: bloodhound
--

CREATE SEQUENCE public.roles_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.roles_id_seq OWNER TO bloodhound;

--
-- Name: roles_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: bloodhound
--

ALTER SEQUENCE public.roles_id_seq OWNED BY public.roles.id;


--
-- Name: roles_permissions; Type: TABLE; Schema: public; Owner: bloodhound
--

CREATE TABLE public.roles_permissions (
    role_id integer NOT NULL,
    permission_id integer NOT NULL
);


ALTER TABLE public.roles_permissions OWNER TO bloodhound;

--
-- Name: saml_providers; Type: TABLE; Schema: public; Owner: bloodhound
--

CREATE TABLE public.saml_providers (
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


ALTER TABLE public.saml_providers OWNER TO bloodhound;

--
-- Name: saml_providers_id_seq; Type: SEQUENCE; Schema: public; Owner: bloodhound
--

CREATE SEQUENCE public.saml_providers_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.saml_providers_id_seq OWNER TO bloodhound;

--
-- Name: saml_providers_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: bloodhound
--

ALTER SEQUENCE public.saml_providers_id_seq OWNED BY public.saml_providers.id;


--
-- Name: saved_queries; Type: TABLE; Schema: public; Owner: bloodhound
--

CREATE TABLE public.saved_queries (
    user_id text,
    name text,
    query text,
    id bigint NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    description text DEFAULT ''::text
);


ALTER TABLE public.saved_queries OWNER TO bloodhound;

--
-- Name: saved_queries_id_seq; Type: SEQUENCE; Schema: public; Owner: bloodhound
--

CREATE SEQUENCE public.saved_queries_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.saved_queries_id_seq OWNER TO bloodhound;

--
-- Name: saved_queries_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: bloodhound
--

ALTER SEQUENCE public.saved_queries_id_seq OWNED BY public.saved_queries.id;


--
-- Name: saved_queries_permissions; Type: TABLE; Schema: public; Owner: bloodhound
--

CREATE TABLE public.saved_queries_permissions (
    id bigint NOT NULL,
    shared_to_user_id text,
    query_id bigint NOT NULL,
    public boolean DEFAULT false NOT NULL,
    created_at timestamp with time zone DEFAULT now(),
    updated_at timestamp with time zone DEFAULT now()
);


ALTER TABLE public.saved_queries_permissions OWNER TO bloodhound;

--
-- Name: saved_queries_permissions_id_seq; Type: SEQUENCE; Schema: public; Owner: bloodhound
--

CREATE SEQUENCE public.saved_queries_permissions_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.saved_queries_permissions_id_seq OWNER TO bloodhound;

--
-- Name: saved_queries_permissions_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: bloodhound
--

ALTER SEQUENCE public.saved_queries_permissions_id_seq OWNED BY public.saved_queries_permissions.id;


--
-- Name: saved_queries_permissions_query_id_seq; Type: SEQUENCE; Schema: public; Owner: bloodhound
--

CREATE SEQUENCE public.saved_queries_permissions_query_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.saved_queries_permissions_query_id_seq OWNER TO bloodhound;

--
-- Name: saved_queries_permissions_query_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: bloodhound
--

ALTER SEQUENCE public.saved_queries_permissions_query_id_seq OWNED BY public.saved_queries_permissions.query_id;


--
-- Name: user_sessions; Type: TABLE; Schema: public; Owner: bloodhound
--

CREATE TABLE public.user_sessions (
    user_id text,
    auth_provider_type bigint,
    auth_provider_id integer,
    expires_at timestamp with time zone,
    id bigint NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    flags jsonb
);


ALTER TABLE public.user_sessions OWNER TO bloodhound;

--
-- Name: user_sessions_id_seq; Type: SEQUENCE; Schema: public; Owner: bloodhound
--

CREATE SEQUENCE public.user_sessions_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.user_sessions_id_seq OWNER TO bloodhound;

--
-- Name: user_sessions_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: bloodhound
--

ALTER SEQUENCE public.user_sessions_id_seq OWNED BY public.user_sessions.id;


--
-- Name: users; Type: TABLE; Schema: public; Owner: bloodhound
--

CREATE TABLE public.users (
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


ALTER TABLE public.users OWNER TO bloodhound;

--
-- Name: users_roles; Type: TABLE; Schema: public; Owner: bloodhound
--

CREATE TABLE public.users_roles (
    user_id text NOT NULL,
    role_id integer NOT NULL
);


ALTER TABLE public.users_roles OWNER TO bloodhound;

--
-- Name: ad_data_quality_aggregations id; Type: DEFAULT; Schema: public; Owner: bloodhound
--

ALTER TABLE ONLY public.ad_data_quality_aggregations ALTER COLUMN id SET DEFAULT nextval('public.ad_data_quality_aggregations_id_seq'::regclass);


--
-- Name: ad_data_quality_stats id; Type: DEFAULT; Schema: public; Owner: bloodhound
--

ALTER TABLE ONLY public.ad_data_quality_stats ALTER COLUMN id SET DEFAULT nextval('public.ad_data_quality_stats_id_seq'::regclass);


--
-- Name: asset_group_collection_entries id; Type: DEFAULT; Schema: public; Owner: bloodhound
--

ALTER TABLE ONLY public.asset_group_collection_entries ALTER COLUMN id SET DEFAULT nextval('public.asset_group_collection_entries_id_seq'::regclass);


--
-- Name: asset_group_collections id; Type: DEFAULT; Schema: public; Owner: bloodhound
--

ALTER TABLE ONLY public.asset_group_collections ALTER COLUMN id SET DEFAULT nextval('public.asset_group_collections_id_seq'::regclass);


--
-- Name: asset_group_selectors id; Type: DEFAULT; Schema: public; Owner: bloodhound
--

ALTER TABLE ONLY public.asset_group_selectors ALTER COLUMN id SET DEFAULT nextval('public.asset_group_selectors_id_seq'::regclass);


--
-- Name: asset_groups id; Type: DEFAULT; Schema: public; Owner: bloodhound
--

ALTER TABLE ONLY public.asset_groups ALTER COLUMN id SET DEFAULT nextval('public.asset_groups_id_seq'::regclass);


--
-- Name: audit_logs id; Type: DEFAULT; Schema: public; Owner: bloodhound
--

ALTER TABLE ONLY public.audit_logs ALTER COLUMN id SET DEFAULT nextval('public.audit_logs_id_seq'::regclass);


--
-- Name: auth_secrets id; Type: DEFAULT; Schema: public; Owner: bloodhound
--

ALTER TABLE ONLY public.auth_secrets ALTER COLUMN id SET DEFAULT nextval('public.auth_secrets_id_seq'::regclass);


--
-- Name: azure_data_quality_aggregations id; Type: DEFAULT; Schema: public; Owner: bloodhound
--

ALTER TABLE ONLY public.azure_data_quality_aggregations ALTER COLUMN id SET DEFAULT nextval('public.azure_data_quality_aggregations_id_seq'::regclass);


--
-- Name: azure_data_quality_stats id; Type: DEFAULT; Schema: public; Owner: bloodhound
--

ALTER TABLE ONLY public.azure_data_quality_stats ALTER COLUMN id SET DEFAULT nextval('public.azure_data_quality_stats_id_seq'::regclass);


--
-- Name: domain_collection_results id; Type: DEFAULT; Schema: public; Owner: bloodhound
--

ALTER TABLE ONLY public.domain_collection_results ALTER COLUMN id SET DEFAULT nextval('public.domain_collection_results_id_seq'::regclass);


--
-- Name: feature_flags id; Type: DEFAULT; Schema: public; Owner: bloodhound
--

ALTER TABLE ONLY public.feature_flags ALTER COLUMN id SET DEFAULT nextval('public.feature_flags_id_seq'::regclass);


--
-- Name: file_upload_jobs id; Type: DEFAULT; Schema: public; Owner: bloodhound
--

ALTER TABLE ONLY public.file_upload_jobs ALTER COLUMN id SET DEFAULT nextval('public.file_upload_jobs_id_seq'::regclass);


--
-- Name: ingest_tasks id; Type: DEFAULT; Schema: public; Owner: bloodhound
--

ALTER TABLE ONLY public.ingest_tasks ALTER COLUMN id SET DEFAULT nextval('public.ingest_tasks_id_seq'::regclass);


--
-- Name: migrations id; Type: DEFAULT; Schema: public; Owner: bloodhound
--

ALTER TABLE ONLY public.migrations ALTER COLUMN id SET DEFAULT nextval('public.migrations_id_seq'::regclass);


--
-- Name: parameters id; Type: DEFAULT; Schema: public; Owner: bloodhound
--

ALTER TABLE ONLY public.parameters ALTER COLUMN id SET DEFAULT nextval('public.parameters_id_seq'::regclass);


--
-- Name: permissions id; Type: DEFAULT; Schema: public; Owner: bloodhound
--

ALTER TABLE ONLY public.permissions ALTER COLUMN id SET DEFAULT nextval('public.permissions_id_seq'::regclass);


--
-- Name: roles id; Type: DEFAULT; Schema: public; Owner: bloodhound
--

ALTER TABLE ONLY public.roles ALTER COLUMN id SET DEFAULT nextval('public.roles_id_seq'::regclass);


--
-- Name: saml_providers id; Type: DEFAULT; Schema: public; Owner: bloodhound
--

ALTER TABLE ONLY public.saml_providers ALTER COLUMN id SET DEFAULT nextval('public.saml_providers_id_seq'::regclass);


--
-- Name: saved_queries id; Type: DEFAULT; Schema: public; Owner: bloodhound
--

ALTER TABLE ONLY public.saved_queries ALTER COLUMN id SET DEFAULT nextval('public.saved_queries_id_seq'::regclass);


--
-- Name: saved_queries_permissions id; Type: DEFAULT; Schema: public; Owner: bloodhound
--

ALTER TABLE ONLY public.saved_queries_permissions ALTER COLUMN id SET DEFAULT nextval('public.saved_queries_permissions_id_seq'::regclass);


--
-- Name: saved_queries_permissions query_id; Type: DEFAULT; Schema: public; Owner: bloodhound
--

ALTER TABLE ONLY public.saved_queries_permissions ALTER COLUMN query_id SET DEFAULT nextval('public.saved_queries_permissions_query_id_seq'::regclass);


--
-- Name: user_sessions id; Type: DEFAULT; Schema: public; Owner: bloodhound
--

ALTER TABLE ONLY public.user_sessions ALTER COLUMN id SET DEFAULT nextval('public.user_sessions_id_seq'::regclass);


--
-- Data for Name: ad_data_quality_aggregations; Type: TABLE DATA; Schema: public; Owner: bloodhound
--

COPY public.ad_data_quality_aggregations (domains, users, groups, computers, ous, containers, gpos, acls, sessions, relationships, session_completeness, local_group_completeness, run_id, id, created_at, updated_at, aiacas, rootcas, enterprisecas, ntauthstores, certtemplates, issuancepolicies) FROM stdin;
\.


--
-- Data for Name: ad_data_quality_stats; Type: TABLE DATA; Schema: public; Owner: bloodhound
--

COPY public.ad_data_quality_stats (domain_sid, users, groups, computers, ous, containers, gpos, acls, sessions, relationships, session_completeness, local_group_completeness, run_id, id, created_at, updated_at, aiacas, rootcas, enterprisecas, ntauthstores, certtemplates, issuancepolicies) FROM stdin;
\.


--
-- Data for Name: analysis_request_switch; Type: TABLE DATA; Schema: public; Owner: bloodhound
--

COPY public.analysis_request_switch (singleton, request_type, requested_by, requested_at) FROM stdin;
\.


--
-- Data for Name: asset_group_collection_entries; Type: TABLE DATA; Schema: public; Owner: bloodhound
--

COPY public.asset_group_collection_entries (asset_group_collection_id, object_id, node_label, properties, id, created_at, updated_at) FROM stdin;
\.


--
-- Data for Name: asset_group_collections; Type: TABLE DATA; Schema: public; Owner: bloodhound
--

COPY public.asset_group_collections (asset_group_id, id, created_at, updated_at) FROM stdin;
\.


--
-- Data for Name: asset_group_selectors; Type: TABLE DATA; Schema: public; Owner: bloodhound
--

COPY public.asset_group_selectors (asset_group_id, name, selector, system_selector, id, created_at, updated_at) FROM stdin;
\.


--
-- Data for Name: asset_groups; Type: TABLE DATA; Schema: public; Owner: bloodhound
--

COPY public.asset_groups (name, tag, system_group, id, created_at, updated_at) FROM stdin;
Admin Tier Zero	admin_tier_0	t	1	2025-03-19 18:14:12.155626+00	2025-03-19 18:14:12.155626+00
Owned	owned	t	2	2025-03-19 18:14:12.219797+00	2025-03-19 18:14:12.219797+00
\.


--
-- Data for Name: audit_logs; Type: TABLE DATA; Schema: public; Owner: bloodhound
--

COPY public.audit_logs (id, created_at, actor_id, actor_name, action, fields, request_id, actor_email, source_ip_address, status, commit_id) FROM stdin;
\.


--
-- Data for Name: auth_secrets; Type: TABLE DATA; Schema: public; Owner: bloodhound
--

COPY public.auth_secrets (user_id, digest, digest_method, expires_at, totp_secret, totp_activated, id, created_at, updated_at) FROM stdin;
47dd420a-a7a1-4915-a66b-b07bdee6e713	$argon2id$v=19$m=1048576,t=1,p=8$LUYDDVTmiGSxuxovC0e+Vw==$ZU1+fl/8PdNIDxYQgLALvA==	argon2	2025-06-17 18:14:12.465698+00		f	1	2025-03-19 18:14:12.468264+00	2025-03-19 18:14:12.468264+00
\.


--
-- Data for Name: auth_tokens; Type: TABLE DATA; Schema: public; Owner: bloodhound
--

COPY public.auth_tokens (user_id, client_id, name, key, hmac_method, last_access, id, created_at, updated_at) FROM stdin;
\.


--
-- Data for Name: azure_data_quality_aggregations; Type: TABLE DATA; Schema: public; Owner: bloodhound
--

COPY public.azure_data_quality_aggregations (tenants, users, groups, apps, service_principals, devices, management_groups, subscriptions, resource_groups, vms, key_vaults, relationships, run_id, id, created_at, updated_at, automation_accounts, container_registries, function_apps, logic_apps, managed_clusters, vm_scale_sets, web_apps) FROM stdin;
\.


--
-- Data for Name: azure_data_quality_stats; Type: TABLE DATA; Schema: public; Owner: bloodhound
--

COPY public.azure_data_quality_stats (tenant_id, users, groups, apps, service_principals, devices, management_groups, subscriptions, resource_groups, vms, key_vaults, relationships, run_id, id, created_at, updated_at, automation_accounts, container_registries, function_apps, logic_apps, managed_clusters, vm_scale_sets, web_apps) FROM stdin;
\.


--
-- Data for Name: database_switch; Type: TABLE DATA; Schema: public; Owner: bloodhound
--

COPY public.database_switch (driver) FROM stdin;
\.


--
-- Data for Name: datapipe_status; Type: TABLE DATA; Schema: public; Owner: bloodhound
--

COPY public.datapipe_status (singleton, status, updated_at, last_complete_analysis_at) FROM stdin;
t	idle	2025-03-19 18:14:12.247995+00	\N
\.


--
-- Data for Name: domain_collection_results; Type: TABLE DATA; Schema: public; Owner: bloodhound
--

COPY public.domain_collection_results (job_id, domain_name, success, message, user_count, group_count, computer_count, gpo_count, ou_count, container_count, deleted_count, id, created_at, updated_at) FROM stdin;
\.


--
-- Data for Name: feature_flags; Type: TABLE DATA; Schema: public; Owner: bloodhound
--

COPY public.feature_flags (id, created_at, updated_at, key, name, description, enabled, user_updatable) FROM stdin;
1	2025-03-19 18:14:12.155626+00	2025-03-19 18:14:12.155626+00	butterfly_analysis	Enhanced Asset Inbound-Outbound Exposure Analysis	Enables more extensive analysis of attack path findings that allows BloodHound to help the user prioritize remediation of the most exposed assets.	f	f
2	2025-03-19 18:14:12.155626+00	2025-03-19 18:14:12.155626+00	enable_saml_sso	SAML Single Sign-On Support	Enables SSO authentication flows and administration panels to third party SAML identity providers.	t	f
3	2025-03-19 18:14:12.155626+00	2025-03-19 18:14:12.155626+00	scope_collection_by_ou	Enable SharpHound OU Scoped Collections	Enables scoping SharpHound collections to specific lists of OUs.	t	f
4	2025-03-19 18:14:12.155626+00	2025-03-19 18:14:12.155626+00	azure_support	Enable Azure Support	Enables Azure support.	t	f
6	2025-03-19 18:14:12.155626+00	2025-03-19 18:14:12.155626+00	entity_panel_cache	Enable application level caching	Enables the use of application level caching for entity panel queries	t	f
7	2025-03-19 18:14:12.155626+00	2025-03-19 18:14:12.155626+00	adcs	Enable collection and processing of Active Directory Certificate Services Data	Enables the ability to collect, analyze, and explore Active Directory Certificate Services data and previews new attack paths.	f	f
9	2025-03-19 18:14:12.255937+00	2025-03-19 18:14:12.255937+00	pg_migration_dual_ingest	PostgreSQL Migration Dual Ingest	Enables dual ingest pathing for both Neo4j and PostgreSQL.	f	f
10	2025-03-19 18:14:12.256348+00	2025-03-19 18:14:12.256348+00	clear_graph_data	Clear Graph Data	Enables the ability to delete all nodes and edges from the graph database.	t	f
11	2025-03-19 18:14:12.256348+00	2025-03-19 18:14:12.256348+00	risk_exposure_new_calculation	Use new tier zero risk exposure calculation	Enables the use of new tier zero risk exposure metatree metrics.	f	f
12	2025-03-19 18:14:12.256348+00	2025-03-19 18:14:12.256348+00	fedramp_eula	FedRAMP EULA	Enables showing the FedRAMP EULA on every login. (Enterprise only)	f	f
13	2025-03-19 18:14:12.256348+00	2025-03-19 18:14:12.256348+00	auto_tag_t0_parent_objects	Automatically add parent OUs and containers of Tier Zero AD objects to Tier Zero	Parent OUs and containers of Tier Zero AD objects are automatically added to Tier Zero during analysis. Containers are only added if they have a Tier Zero child object with ACL inheritance enabled.	t	t
14	2025-03-19 18:14:12.263971+00	2025-03-19 18:14:12.263971+00	oidc_support	OIDC Support	Enables OpenID Connect authentication support for SSO Authentication.	f	f
8	2025-03-19 18:14:12.252443+00	2025-03-19 18:14:12.252443+00	dark_mode	Dark Mode	Allows users to enable or disable dark mode via a toggle in the settings menu	t	f
\.


--
-- Data for Name: file_upload_jobs; Type: TABLE DATA; Schema: public; Owner: bloodhound
--

COPY public.file_upload_jobs (user_id, user_email_address, status, status_message, start_time, end_time, last_ingest, id, created_at, updated_at, total_files, failed_files) FROM stdin;
\.


--
-- Data for Name: ingest_tasks; Type: TABLE DATA; Schema: public; Owner: bloodhound
--

COPY public.ingest_tasks (file_name, request_guid, task_id, id, created_at, updated_at, file_type) FROM stdin;
\.


--
-- Data for Name: installations; Type: TABLE DATA; Schema: public; Owner: bloodhound
--

COPY public.installations (id, created_at, updated_at) FROM stdin;
74fab259-7829-4e06-85d4-b79ff89fe165	2025-03-19 18:14:12.466349+00	2025-03-19 18:14:12.466349+00
\.


--
-- Data for Name: migrations; Type: TABLE DATA; Schema: public; Owner: bloodhound
--

COPY public.migrations (id, updated_at, major, minor, patch) FROM stdin;
1	2025-03-19 18:14:12.218348+00	0	0	0
2	2025-03-19 18:14:12.219994+00	5	1	1
3	2025-03-19 18:14:12.221431+00	5	3	0
4	2025-03-19 18:14:12.228431+00	5	3	1
5	2025-03-19 18:14:12.230872+00	5	4	0
6	2025-03-19 18:14:12.23187+00	5	5	0
7	2025-03-19 18:14:12.234046+00	5	6	0
8	2025-03-19 18:14:12.234631+00	5	8	0
9	2025-03-19 18:14:12.235725+00	5	8	1
10	2025-03-19 18:14:12.245783+00	5	8	2
11	2025-03-19 18:14:12.247621+00	5	11	0
12	2025-03-19 18:14:12.252176+00	5	12	0
13	2025-03-19 18:14:12.252602+00	5	13	0
14	2025-03-19 18:14:12.2555+00	5	13	1
15	2025-03-19 18:14:12.256079+00	5	14	0
16	2025-03-19 18:14:12.263026+00	5	15	0
17	2025-03-19 18:14:12.265713+00	6	0	0
\.


--
-- Data for Name: parameters; Type: TABLE DATA; Schema: public; Owner: bloodhound
--

COPY public.parameters (key, name, description, value, id, created_at, updated_at) FROM stdin;
auth.password_expiration_window	Local Auth Password Expiry Window	This configuration parameter sets the local auth password expiry window for users that have valid auth secrets. Values for this configuration must follow the duration specification of ISO-8601.	{"duration": "P90D"}	1	2025-03-19 18:14:12.155626+00	2025-03-19 18:14:12.155626+00
neo4j.configuration	Neo4j Configuration Parameters	This configuration parameter sets the BatchWriteSize and the BatchFlushSize for Neo4J.	{"batch_write_size": 20000, "write_flush_size": 100000}	2	2025-03-19 18:14:12.155626+00	2025-03-19 18:14:12.155626+00
analysis.citrix_rdp_support	Citrix RDP Support	This configuration parameter toggles Citrix support during post-processing. When enabled, computers identified with a 'Direct Access Users' local group will assume that Citrix is installed and CanRDP edges will require membership of both 'Direct Access Users' and 'Remote Desktop Users' local groups on the computer.	{"enabled": false}	3	2025-03-19 18:14:12.263971+00	2025-03-19 18:14:12.263971+00
prune.ttl	Prune Retention TTL Configuration Parameters	This configuration parameter sets the retention TTLs during analysis pruning.	{"base_ttl": "P7D", "has_session_edge_ttl": "P3D"}	4	2025-03-19 18:14:12.263971+00	2025-03-19 18:14:12.263971+00
analysis.reconciliation	Reconciliation	This configuration parameter enables / disables reconciliation during analysis.	{"enabled": true}	5	2025-03-19 18:14:12.263971+00	2025-03-19 18:14:12.263971+00
\.


--
-- Data for Name: permissions; Type: TABLE DATA; Schema: public; Owner: bloodhound
--

COPY public.permissions (authority, name, id, created_at, updated_at) FROM stdin;
app	ReadAppConfig	1	2025-03-19 18:14:12.155626+00	2025-03-19 18:14:12.155626+00
app	WriteAppConfig	2	2025-03-19 18:14:12.155626+00	2025-03-19 18:14:12.155626+00
risks	GenerateReport	3	2025-03-19 18:14:12.155626+00	2025-03-19 18:14:12.155626+00
risks	ManageRisks	4	2025-03-19 18:14:12.155626+00	2025-03-19 18:14:12.155626+00
auth	CreateToken	5	2025-03-19 18:14:12.155626+00	2025-03-19 18:14:12.155626+00
auth	ManageAppConfig	6	2025-03-19 18:14:12.155626+00	2025-03-19 18:14:12.155626+00
auth	ManageProviders	7	2025-03-19 18:14:12.155626+00	2025-03-19 18:14:12.155626+00
auth	ManageSelf	8	2025-03-19 18:14:12.155626+00	2025-03-19 18:14:12.155626+00
auth	ManageUsers	9	2025-03-19 18:14:12.155626+00	2025-03-19 18:14:12.155626+00
clients	Manage	10	2025-03-19 18:14:12.155626+00	2025-03-19 18:14:12.155626+00
clients	Tasking	11	2025-03-19 18:14:12.155626+00	2025-03-19 18:14:12.155626+00
collection	ManageJobs	12	2025-03-19 18:14:12.155626+00	2025-03-19 18:14:12.155626+00
graphdb	Read	13	2025-03-19 18:14:12.155626+00	2025-03-19 18:14:12.155626+00
graphdb	Write	14	2025-03-19 18:14:12.155626+00	2025-03-19 18:14:12.155626+00
saved_queries	Read	15	2025-03-19 18:14:12.256348+00	2025-03-19 18:14:12.256348+00
saved_queries	Write	16	2025-03-19 18:14:12.256348+00	2025-03-19 18:14:12.256348+00
clients	Read	17	2025-03-19 18:14:12.256348+00	2025-03-19 18:14:12.256348+00
db	Wipe	18	2025-03-19 18:14:12.256348+00	2025-03-19 18:14:12.256348+00
graphdb	Mutate	19	2025-03-19 18:14:12.256348+00	2025-03-19 18:14:12.256348+00
\.


--
-- Data for Name: roles; Type: TABLE DATA; Schema: public; Owner: bloodhound
--

COPY public.roles (name, description, id, created_at, updated_at) FROM stdin;
Administrator	Can manage users, clients, and application configuration	1	2025-03-19 18:14:12.155626+00	2025-03-19 18:14:12.155626+00
User	Can read data, modify asset group memberships	2	2025-03-19 18:14:12.155626+00	2025-03-19 18:14:12.155626+00
Read-Only	Used for integrations	3	2025-03-19 18:14:12.155626+00	2025-03-19 18:14:12.155626+00
Upload-Only	Used for data collection clients, can post data but cannot read data	4	2025-03-19 18:14:12.155626+00	2025-03-19 18:14:12.155626+00
Power User	Can upload data, manage clients, and perform any action a User can	5	2025-03-19 18:14:12.256348+00	2025-03-19 18:14:12.256348+00
\.


--
-- Data for Name: roles_permissions; Type: TABLE DATA; Schema: public; Owner: bloodhound
--

COPY public.roles_permissions (role_id, permission_id) FROM stdin;
1	1
1	2
1	3
1	4
1	5
1	6
1	7
1	8
1	9
1	10
1	11
1	12
1	13
1	14
2	1
2	3
2	5
2	8
2	13
3	1
3	3
3	8
3	13
4	11
4	14
1	15
1	16
1	17
1	18
1	19
2	15
2	16
2	17
3	5
5	1
5	2
5	3
5	4
5	5
5	8
5	10
5	17
5	11
5	12
5	14
5	13
5	15
5	16
5	19
3	15
\.


--
-- Data for Name: saml_providers; Type: TABLE DATA; Schema: public; Owner: bloodhound
--

COPY public.saml_providers (name, display_name, issuer_uri, single_sign_on_uri, metadata_xml, ous, id, created_at, updated_at) FROM stdin;
\.


--
-- Data for Name: saved_queries; Type: TABLE DATA; Schema: public; Owner: bloodhound
--

COPY public.saved_queries (user_id, name, query, id, created_at, updated_at, description) FROM stdin;
\.


--
-- Data for Name: saved_queries_permissions; Type: TABLE DATA; Schema: public; Owner: bloodhound
--

COPY public.saved_queries_permissions (id, shared_to_user_id, query_id, public, created_at, updated_at) FROM stdin;
\.


--
-- Data for Name: user_sessions; Type: TABLE DATA; Schema: public; Owner: bloodhound
--

COPY public.user_sessions (user_id, auth_provider_type, auth_provider_id, expires_at, id, created_at, updated_at, flags) FROM stdin;
\.


--
-- Data for Name: users; Type: TABLE DATA; Schema: public; Owner: bloodhound
--

COPY public.users (saml_provider_id, first_name, last_name, email_address, principal_name, last_login, is_disabled, eula_accepted, id, created_at, updated_at) FROM stdin;
\N	BloodHound	Dev	spam@example.com	admin	0001-01-01 00:00:00+00	f	f	47dd420a-a7a1-4915-a66b-b07bdee6e713	2025-03-19 18:14:12.466972+00	2025-03-19 18:14:12.466972+00
\.


--
-- Data for Name: users_roles; Type: TABLE DATA; Schema: public; Owner: bloodhound
--

COPY public.users_roles (user_id, role_id) FROM stdin;
47dd420a-a7a1-4915-a66b-b07bdee6e713	1
\.


--
-- Name: ad_data_quality_aggregations_id_seq; Type: SEQUENCE SET; Schema: public; Owner: bloodhound
--

SELECT pg_catalog.setval('public.ad_data_quality_aggregations_id_seq', 1, false);


--
-- Name: ad_data_quality_stats_id_seq; Type: SEQUENCE SET; Schema: public; Owner: bloodhound
--

SELECT pg_catalog.setval('public.ad_data_quality_stats_id_seq', 1, false);


--
-- Name: asset_group_collection_entries_id_seq; Type: SEQUENCE SET; Schema: public; Owner: bloodhound
--

SELECT pg_catalog.setval('public.asset_group_collection_entries_id_seq', 1, false);


--
-- Name: asset_group_collections_id_seq; Type: SEQUENCE SET; Schema: public; Owner: bloodhound
--

SELECT pg_catalog.setval('public.asset_group_collections_id_seq', 1, false);


--
-- Name: asset_group_selectors_id_seq; Type: SEQUENCE SET; Schema: public; Owner: bloodhound
--

SELECT pg_catalog.setval('public.asset_group_selectors_id_seq', 1, false);


--
-- Name: asset_groups_id_seq; Type: SEQUENCE SET; Schema: public; Owner: bloodhound
--

SELECT pg_catalog.setval('public.asset_groups_id_seq', 2, true);


--
-- Name: audit_logs_id_seq; Type: SEQUENCE SET; Schema: public; Owner: bloodhound
--

SELECT pg_catalog.setval('public.audit_logs_id_seq', 1, false);


--
-- Name: auth_secrets_id_seq; Type: SEQUENCE SET; Schema: public; Owner: bloodhound
--

SELECT pg_catalog.setval('public.auth_secrets_id_seq', 1, true);


--
-- Name: azure_data_quality_aggregations_id_seq; Type: SEQUENCE SET; Schema: public; Owner: bloodhound
--

SELECT pg_catalog.setval('public.azure_data_quality_aggregations_id_seq', 1, false);


--
-- Name: azure_data_quality_stats_id_seq; Type: SEQUENCE SET; Schema: public; Owner: bloodhound
--

SELECT pg_catalog.setval('public.azure_data_quality_stats_id_seq', 1, false);


--
-- Name: domain_collection_results_id_seq; Type: SEQUENCE SET; Schema: public; Owner: bloodhound
--

SELECT pg_catalog.setval('public.domain_collection_results_id_seq', 1, false);


--
-- Name: feature_flags_id_seq; Type: SEQUENCE SET; Schema: public; Owner: bloodhound
--

SELECT pg_catalog.setval('public.feature_flags_id_seq', 14, true);


--
-- Name: file_upload_jobs_id_seq; Type: SEQUENCE SET; Schema: public; Owner: bloodhound
--

SELECT pg_catalog.setval('public.file_upload_jobs_id_seq', 1, false);


--
-- Name: ingest_tasks_id_seq; Type: SEQUENCE SET; Schema: public; Owner: bloodhound
--

SELECT pg_catalog.setval('public.ingest_tasks_id_seq', 1, false);


--
-- Name: migrations_id_seq; Type: SEQUENCE SET; Schema: public; Owner: bloodhound
--

SELECT pg_catalog.setval('public.migrations_id_seq', 17, true);


--
-- Name: parameters_id_seq; Type: SEQUENCE SET; Schema: public; Owner: bloodhound
--

SELECT pg_catalog.setval('public.parameters_id_seq', 5, true);


--
-- Name: permissions_id_seq; Type: SEQUENCE SET; Schema: public; Owner: bloodhound
--

SELECT pg_catalog.setval('public.permissions_id_seq', 19, true);


--
-- Name: roles_id_seq; Type: SEQUENCE SET; Schema: public; Owner: bloodhound
--

SELECT pg_catalog.setval('public.roles_id_seq', 5, true);


--
-- Name: saml_providers_id_seq; Type: SEQUENCE SET; Schema: public; Owner: bloodhound
--

SELECT pg_catalog.setval('public.saml_providers_id_seq', 1, false);


--
-- Name: saved_queries_id_seq; Type: SEQUENCE SET; Schema: public; Owner: bloodhound
--

SELECT pg_catalog.setval('public.saved_queries_id_seq', 1, false);


--
-- Name: saved_queries_permissions_id_seq; Type: SEQUENCE SET; Schema: public; Owner: bloodhound
--

SELECT pg_catalog.setval('public.saved_queries_permissions_id_seq', 1, false);


--
-- Name: saved_queries_permissions_query_id_seq; Type: SEQUENCE SET; Schema: public; Owner: bloodhound
--

SELECT pg_catalog.setval('public.saved_queries_permissions_query_id_seq', 1, false);


--
-- Name: user_sessions_id_seq; Type: SEQUENCE SET; Schema: public; Owner: bloodhound
--

SELECT pg_catalog.setval('public.user_sessions_id_seq', 1, false);


--
-- Name: ad_data_quality_aggregations ad_data_quality_aggregations_pkey; Type: CONSTRAINT; Schema: public; Owner: bloodhound
--

ALTER TABLE ONLY public.ad_data_quality_aggregations
    ADD CONSTRAINT ad_data_quality_aggregations_pkey PRIMARY KEY (id);


--
-- Name: ad_data_quality_stats ad_data_quality_stats_pkey; Type: CONSTRAINT; Schema: public; Owner: bloodhound
--

ALTER TABLE ONLY public.ad_data_quality_stats
    ADD CONSTRAINT ad_data_quality_stats_pkey PRIMARY KEY (id);


--
-- Name: analysis_request_switch analysis_request_switch_pkey; Type: CONSTRAINT; Schema: public; Owner: bloodhound
--

ALTER TABLE ONLY public.analysis_request_switch
    ADD CONSTRAINT analysis_request_switch_pkey PRIMARY KEY (singleton);


--
-- Name: asset_group_collection_entries asset_group_collection_entries_pkey; Type: CONSTRAINT; Schema: public; Owner: bloodhound
--

ALTER TABLE ONLY public.asset_group_collection_entries
    ADD CONSTRAINT asset_group_collection_entries_pkey PRIMARY KEY (id);


--
-- Name: asset_group_collections asset_group_collections_pkey; Type: CONSTRAINT; Schema: public; Owner: bloodhound
--

ALTER TABLE ONLY public.asset_group_collections
    ADD CONSTRAINT asset_group_collections_pkey PRIMARY KEY (id);


--
-- Name: asset_group_selectors asset_group_selectors_name_assetgroupid_key; Type: CONSTRAINT; Schema: public; Owner: bloodhound
--

ALTER TABLE ONLY public.asset_group_selectors
    ADD CONSTRAINT asset_group_selectors_name_assetgroupid_key UNIQUE (name, asset_group_id);


--
-- Name: asset_group_selectors asset_group_selectors_pkey; Type: CONSTRAINT; Schema: public; Owner: bloodhound
--

ALTER TABLE ONLY public.asset_group_selectors
    ADD CONSTRAINT asset_group_selectors_pkey PRIMARY KEY (id);


--
-- Name: asset_groups asset_groups_name_key; Type: CONSTRAINT; Schema: public; Owner: bloodhound
--

ALTER TABLE ONLY public.asset_groups
    ADD CONSTRAINT asset_groups_name_key UNIQUE (name);


--
-- Name: asset_groups asset_groups_pkey; Type: CONSTRAINT; Schema: public; Owner: bloodhound
--

ALTER TABLE ONLY public.asset_groups
    ADD CONSTRAINT asset_groups_pkey PRIMARY KEY (id);


--
-- Name: asset_groups asset_groups_tag_key; Type: CONSTRAINT; Schema: public; Owner: bloodhound
--

ALTER TABLE ONLY public.asset_groups
    ADD CONSTRAINT asset_groups_tag_key UNIQUE (tag);


--
-- Name: audit_logs audit_logs_pkey; Type: CONSTRAINT; Schema: public; Owner: bloodhound
--

ALTER TABLE ONLY public.audit_logs
    ADD CONSTRAINT audit_logs_pkey PRIMARY KEY (id);


--
-- Name: auth_secrets auth_secrets_pkey; Type: CONSTRAINT; Schema: public; Owner: bloodhound
--

ALTER TABLE ONLY public.auth_secrets
    ADD CONSTRAINT auth_secrets_pkey PRIMARY KEY (id);


--
-- Name: auth_tokens auth_tokens_pkey; Type: CONSTRAINT; Schema: public; Owner: bloodhound
--

ALTER TABLE ONLY public.auth_tokens
    ADD CONSTRAINT auth_tokens_pkey PRIMARY KEY (id);


--
-- Name: azure_data_quality_aggregations azure_data_quality_aggregations_pkey; Type: CONSTRAINT; Schema: public; Owner: bloodhound
--

ALTER TABLE ONLY public.azure_data_quality_aggregations
    ADD CONSTRAINT azure_data_quality_aggregations_pkey PRIMARY KEY (id);


--
-- Name: azure_data_quality_stats azure_data_quality_stats_pkey; Type: CONSTRAINT; Schema: public; Owner: bloodhound
--

ALTER TABLE ONLY public.azure_data_quality_stats
    ADD CONSTRAINT azure_data_quality_stats_pkey PRIMARY KEY (id);


--
-- Name: database_switch database_switch_pkey; Type: CONSTRAINT; Schema: public; Owner: bloodhound
--

ALTER TABLE ONLY public.database_switch
    ADD CONSTRAINT database_switch_pkey PRIMARY KEY (driver);


--
-- Name: datapipe_status datapipe_status_pkey; Type: CONSTRAINT; Schema: public; Owner: bloodhound
--

ALTER TABLE ONLY public.datapipe_status
    ADD CONSTRAINT datapipe_status_pkey PRIMARY KEY (singleton);


--
-- Name: domain_collection_results domain_collection_results_pkey; Type: CONSTRAINT; Schema: public; Owner: bloodhound
--

ALTER TABLE ONLY public.domain_collection_results
    ADD CONSTRAINT domain_collection_results_pkey PRIMARY KEY (id);


--
-- Name: feature_flags feature_flags_key_key; Type: CONSTRAINT; Schema: public; Owner: bloodhound
--

ALTER TABLE ONLY public.feature_flags
    ADD CONSTRAINT feature_flags_key_key UNIQUE (key);


--
-- Name: feature_flags feature_flags_pkey; Type: CONSTRAINT; Schema: public; Owner: bloodhound
--

ALTER TABLE ONLY public.feature_flags
    ADD CONSTRAINT feature_flags_pkey PRIMARY KEY (id);


--
-- Name: file_upload_jobs file_upload_jobs_pkey; Type: CONSTRAINT; Schema: public; Owner: bloodhound
--

ALTER TABLE ONLY public.file_upload_jobs
    ADD CONSTRAINT file_upload_jobs_pkey PRIMARY KEY (id);


--
-- Name: ingest_tasks ingest_tasks_pkey; Type: CONSTRAINT; Schema: public; Owner: bloodhound
--

ALTER TABLE ONLY public.ingest_tasks
    ADD CONSTRAINT ingest_tasks_pkey PRIMARY KEY (id);


--
-- Name: installations installations_pkey; Type: CONSTRAINT; Schema: public; Owner: bloodhound
--

ALTER TABLE ONLY public.installations
    ADD CONSTRAINT installations_pkey PRIMARY KEY (id);


--
-- Name: migrations migrations_pkey; Type: CONSTRAINT; Schema: public; Owner: bloodhound
--

ALTER TABLE ONLY public.migrations
    ADD CONSTRAINT migrations_pkey PRIMARY KEY (id);


--
-- Name: parameters parameters_key_key; Type: CONSTRAINT; Schema: public; Owner: bloodhound
--

ALTER TABLE ONLY public.parameters
    ADD CONSTRAINT parameters_key_key UNIQUE (key);


--
-- Name: parameters parameters_pkey; Type: CONSTRAINT; Schema: public; Owner: bloodhound
--

ALTER TABLE ONLY public.parameters
    ADD CONSTRAINT parameters_pkey PRIMARY KEY (id);


--
-- Name: permissions permissions_authority_name_key; Type: CONSTRAINT; Schema: public; Owner: bloodhound
--

ALTER TABLE ONLY public.permissions
    ADD CONSTRAINT permissions_authority_name_key UNIQUE (authority, name);


--
-- Name: permissions permissions_pkey; Type: CONSTRAINT; Schema: public; Owner: bloodhound
--

ALTER TABLE ONLY public.permissions
    ADD CONSTRAINT permissions_pkey PRIMARY KEY (id);


--
-- Name: roles roles_name_key; Type: CONSTRAINT; Schema: public; Owner: bloodhound
--

ALTER TABLE ONLY public.roles
    ADD CONSTRAINT roles_name_key UNIQUE (name);


--
-- Name: roles_permissions roles_permissions_pkey; Type: CONSTRAINT; Schema: public; Owner: bloodhound
--

ALTER TABLE ONLY public.roles_permissions
    ADD CONSTRAINT roles_permissions_pkey PRIMARY KEY (role_id, permission_id);


--
-- Name: roles roles_pkey; Type: CONSTRAINT; Schema: public; Owner: bloodhound
--

ALTER TABLE ONLY public.roles
    ADD CONSTRAINT roles_pkey PRIMARY KEY (id);


--
-- Name: saml_providers saml_providers_name_key; Type: CONSTRAINT; Schema: public; Owner: bloodhound
--

ALTER TABLE ONLY public.saml_providers
    ADD CONSTRAINT saml_providers_name_key UNIQUE (name);


--
-- Name: saml_providers saml_providers_pkey; Type: CONSTRAINT; Schema: public; Owner: bloodhound
--

ALTER TABLE ONLY public.saml_providers
    ADD CONSTRAINT saml_providers_pkey PRIMARY KEY (id);


--
-- Name: saved_queries_permissions saved_queries_permissions_pkey; Type: CONSTRAINT; Schema: public; Owner: bloodhound
--

ALTER TABLE ONLY public.saved_queries_permissions
    ADD CONSTRAINT saved_queries_permissions_pkey PRIMARY KEY (id);


--
-- Name: saved_queries_permissions saved_queries_permissions_shared_to_user_id_query_id_key; Type: CONSTRAINT; Schema: public; Owner: bloodhound
--

ALTER TABLE ONLY public.saved_queries_permissions
    ADD CONSTRAINT saved_queries_permissions_shared_to_user_id_query_id_key UNIQUE (shared_to_user_id, query_id);


--
-- Name: saved_queries saved_queries_pkey; Type: CONSTRAINT; Schema: public; Owner: bloodhound
--

ALTER TABLE ONLY public.saved_queries
    ADD CONSTRAINT saved_queries_pkey PRIMARY KEY (id);


--
-- Name: user_sessions user_sessions_pkey; Type: CONSTRAINT; Schema: public; Owner: bloodhound
--

ALTER TABLE ONLY public.user_sessions
    ADD CONSTRAINT user_sessions_pkey PRIMARY KEY (id);


--
-- Name: users users_pkey; Type: CONSTRAINT; Schema: public; Owner: bloodhound
--

ALTER TABLE ONLY public.users
    ADD CONSTRAINT users_pkey PRIMARY KEY (id);


--
-- Name: users users_principal_name_key; Type: CONSTRAINT; Schema: public; Owner: bloodhound
--

ALTER TABLE ONLY public.users
    ADD CONSTRAINT users_principal_name_key UNIQUE (principal_name);


--
-- Name: users_roles users_roles_pkey; Type: CONSTRAINT; Schema: public; Owner: bloodhound
--

ALTER TABLE ONLY public.users_roles
    ADD CONSTRAINT users_roles_pkey PRIMARY KEY (user_id, role_id);


--
-- Name: idx_ad_asset_groups_created_at; Type: INDEX; Schema: public; Owner: bloodhound
--

CREATE INDEX idx_ad_asset_groups_created_at ON public.asset_groups USING btree (created_at);


--
-- Name: idx_ad_asset_groups_updated_at; Type: INDEX; Schema: public; Owner: bloodhound
--

CREATE INDEX idx_ad_asset_groups_updated_at ON public.asset_groups USING btree (updated_at);


--
-- Name: idx_ad_data_quality_aggregations_created_at; Type: INDEX; Schema: public; Owner: bloodhound
--

CREATE INDEX idx_ad_data_quality_aggregations_created_at ON public.ad_data_quality_aggregations USING btree (created_at);


--
-- Name: idx_ad_data_quality_aggregations_run_id; Type: INDEX; Schema: public; Owner: bloodhound
--

CREATE INDEX idx_ad_data_quality_aggregations_run_id ON public.ad_data_quality_aggregations USING btree (run_id);


--
-- Name: idx_ad_data_quality_aggregations_updated_at; Type: INDEX; Schema: public; Owner: bloodhound
--

CREATE INDEX idx_ad_data_quality_aggregations_updated_at ON public.ad_data_quality_aggregations USING btree (updated_at);


--
-- Name: idx_ad_data_quality_stats_run_id; Type: INDEX; Schema: public; Owner: bloodhound
--

CREATE INDEX idx_ad_data_quality_stats_run_id ON public.ad_data_quality_stats USING btree (run_id);


--
-- Name: idx_asset_group_collection_entries_asset_group_collection_id; Type: INDEX; Schema: public; Owner: bloodhound
--

CREATE INDEX idx_asset_group_collection_entries_asset_group_collection_id ON public.asset_group_collection_entries USING btree (asset_group_collection_id);


--
-- Name: idx_asset_group_collection_entries_created_at; Type: INDEX; Schema: public; Owner: bloodhound
--

CREATE INDEX idx_asset_group_collection_entries_created_at ON public.asset_group_collection_entries USING btree (created_at);


--
-- Name: idx_asset_group_collection_entries_updated_at; Type: INDEX; Schema: public; Owner: bloodhound
--

CREATE INDEX idx_asset_group_collection_entries_updated_at ON public.asset_group_collection_entries USING btree (updated_at);


--
-- Name: idx_asset_group_collections_asset_group_id; Type: INDEX; Schema: public; Owner: bloodhound
--

CREATE INDEX idx_asset_group_collections_asset_group_id ON public.asset_group_collections USING btree (asset_group_id);


--
-- Name: idx_asset_group_collections_created_at; Type: INDEX; Schema: public; Owner: bloodhound
--

CREATE INDEX idx_asset_group_collections_created_at ON public.asset_group_collections USING btree (created_at);


--
-- Name: idx_asset_group_collections_updated_at; Type: INDEX; Schema: public; Owner: bloodhound
--

CREATE INDEX idx_asset_group_collections_updated_at ON public.asset_group_collections USING btree (updated_at);


--
-- Name: idx_audit_logs_action; Type: INDEX; Schema: public; Owner: bloodhound
--

CREATE INDEX idx_audit_logs_action ON public.audit_logs USING btree (action);


--
-- Name: idx_audit_logs_actor_email; Type: INDEX; Schema: public; Owner: bloodhound
--

CREATE INDEX idx_audit_logs_actor_email ON public.audit_logs USING btree (actor_email);


--
-- Name: idx_audit_logs_actor_id; Type: INDEX; Schema: public; Owner: bloodhound
--

CREATE INDEX idx_audit_logs_actor_id ON public.audit_logs USING btree (actor_id);


--
-- Name: idx_audit_logs_created_at; Type: INDEX; Schema: public; Owner: bloodhound
--

CREATE INDEX idx_audit_logs_created_at ON public.audit_logs USING btree (created_at);


--
-- Name: idx_audit_logs_source_ip_address; Type: INDEX; Schema: public; Owner: bloodhound
--

CREATE INDEX idx_audit_logs_source_ip_address ON public.audit_logs USING btree (source_ip_address);


--
-- Name: idx_audit_logs_status; Type: INDEX; Schema: public; Owner: bloodhound
--

CREATE INDEX idx_audit_logs_status ON public.audit_logs USING btree (status);


--
-- Name: idx_azure_data_quality_aggregations_created_at; Type: INDEX; Schema: public; Owner: bloodhound
--

CREATE INDEX idx_azure_data_quality_aggregations_created_at ON public.azure_data_quality_aggregations USING btree (created_at);


--
-- Name: idx_azure_data_quality_aggregations_run_id; Type: INDEX; Schema: public; Owner: bloodhound
--

CREATE INDEX idx_azure_data_quality_aggregations_run_id ON public.azure_data_quality_aggregations USING btree (run_id);


--
-- Name: idx_azure_data_quality_stats_created_at; Type: INDEX; Schema: public; Owner: bloodhound
--

CREATE INDEX idx_azure_data_quality_stats_created_at ON public.azure_data_quality_stats USING btree (created_at);


--
-- Name: idx_azure_data_quality_stats_run_id; Type: INDEX; Schema: public; Owner: bloodhound
--

CREATE INDEX idx_azure_data_quality_stats_run_id ON public.azure_data_quality_stats USING btree (run_id);


--
-- Name: idx_azure_data_quality_stats_updated_at; Type: INDEX; Schema: public; Owner: bloodhound
--

CREATE INDEX idx_azure_data_quality_stats_updated_at ON public.azure_data_quality_stats USING btree (updated_at);


--
-- Name: idx_file_upload_jobs_created_at; Type: INDEX; Schema: public; Owner: bloodhound
--

CREATE INDEX idx_file_upload_jobs_created_at ON public.file_upload_jobs USING btree (created_at);


--
-- Name: idx_file_upload_jobs_end_time; Type: INDEX; Schema: public; Owner: bloodhound
--

CREATE INDEX idx_file_upload_jobs_end_time ON public.file_upload_jobs USING btree (end_time);


--
-- Name: idx_file_upload_jobs_start_time; Type: INDEX; Schema: public; Owner: bloodhound
--

CREATE INDEX idx_file_upload_jobs_start_time ON public.file_upload_jobs USING btree (start_time);


--
-- Name: idx_file_upload_jobs_status; Type: INDEX; Schema: public; Owner: bloodhound
--

CREATE INDEX idx_file_upload_jobs_status ON public.file_upload_jobs USING btree (status);


--
-- Name: idx_file_upload_jobs_updated_at; Type: INDEX; Schema: public; Owner: bloodhound
--

CREATE INDEX idx_file_upload_jobs_updated_at ON public.file_upload_jobs USING btree (updated_at);


--
-- Name: idx_ingest_tasks_task_id; Type: INDEX; Schema: public; Owner: bloodhound
--

CREATE INDEX idx_ingest_tasks_task_id ON public.ingest_tasks USING btree (task_id);


--
-- Name: idx_saml_providers_name; Type: INDEX; Schema: public; Owner: bloodhound
--

CREATE INDEX idx_saml_providers_name ON public.saml_providers USING btree (name);


--
-- Name: idx_saved_queries_composite_index; Type: INDEX; Schema: public; Owner: bloodhound
--

CREATE UNIQUE INDEX idx_saved_queries_composite_index ON public.saved_queries USING btree (user_id, name);


--
-- Name: idx_saved_queries_description; Type: INDEX; Schema: public; Owner: bloodhound
--

CREATE INDEX idx_saved_queries_description ON public.saved_queries USING gin (description public.gin_trgm_ops);


--
-- Name: idx_saved_queries_name; Type: INDEX; Schema: public; Owner: bloodhound
--

CREATE INDEX idx_saved_queries_name ON public.saved_queries USING gin (name public.gin_trgm_ops);


--
-- Name: idx_users_eula_accepted; Type: INDEX; Schema: public; Owner: bloodhound
--

CREATE INDEX idx_users_eula_accepted ON public.users USING btree (eula_accepted);


--
-- Name: idx_users_principal_name; Type: INDEX; Schema: public; Owner: bloodhound
--

CREATE INDEX idx_users_principal_name ON public.users USING btree (principal_name);


--
-- Name: asset_group_collection_entries fk_asset_group_collections_entries; Type: FK CONSTRAINT; Schema: public; Owner: bloodhound
--

ALTER TABLE ONLY public.asset_group_collection_entries
    ADD CONSTRAINT fk_asset_group_collections_entries FOREIGN KEY (asset_group_collection_id) REFERENCES public.asset_group_collections(id) ON DELETE CASCADE;


--
-- Name: asset_group_collections fk_asset_groups_collections; Type: FK CONSTRAINT; Schema: public; Owner: bloodhound
--

ALTER TABLE ONLY public.asset_group_collections
    ADD CONSTRAINT fk_asset_groups_collections FOREIGN KEY (asset_group_id) REFERENCES public.asset_groups(id) ON DELETE CASCADE;


--
-- Name: asset_group_selectors fk_asset_groups_selectors; Type: FK CONSTRAINT; Schema: public; Owner: bloodhound
--

ALTER TABLE ONLY public.asset_group_selectors
    ADD CONSTRAINT fk_asset_groups_selectors FOREIGN KEY (asset_group_id) REFERENCES public.asset_groups(id) ON DELETE CASCADE;


--
-- Name: file_upload_jobs fk_file_upload_jobs_user; Type: FK CONSTRAINT; Schema: public; Owner: bloodhound
--

ALTER TABLE ONLY public.file_upload_jobs
    ADD CONSTRAINT fk_file_upload_jobs_user FOREIGN KEY (user_id) REFERENCES public.users(id);


--
-- Name: roles_permissions fk_roles_permissions_permission; Type: FK CONSTRAINT; Schema: public; Owner: bloodhound
--

ALTER TABLE ONLY public.roles_permissions
    ADD CONSTRAINT fk_roles_permissions_permission FOREIGN KEY (permission_id) REFERENCES public.permissions(id);


--
-- Name: roles_permissions fk_roles_permissions_role; Type: FK CONSTRAINT; Schema: public; Owner: bloodhound
--

ALTER TABLE ONLY public.roles_permissions
    ADD CONSTRAINT fk_roles_permissions_role FOREIGN KEY (role_id) REFERENCES public.roles(id);


--
-- Name: user_sessions fk_user_sessions_user; Type: FK CONSTRAINT; Schema: public; Owner: bloodhound
--

ALTER TABLE ONLY public.user_sessions
    ADD CONSTRAINT fk_user_sessions_user FOREIGN KEY (user_id) REFERENCES public.users(id) ON DELETE CASCADE;


--
-- Name: auth_secrets fk_users_auth_secret; Type: FK CONSTRAINT; Schema: public; Owner: bloodhound
--

ALTER TABLE ONLY public.auth_secrets
    ADD CONSTRAINT fk_users_auth_secret FOREIGN KEY (user_id) REFERENCES public.users(id) ON DELETE CASCADE;


--
-- Name: auth_tokens fk_users_auth_tokens; Type: FK CONSTRAINT; Schema: public; Owner: bloodhound
--

ALTER TABLE ONLY public.auth_tokens
    ADD CONSTRAINT fk_users_auth_tokens FOREIGN KEY (user_id) REFERENCES public.users(id) ON DELETE CASCADE;


--
-- Name: users_roles fk_users_roles_role; Type: FK CONSTRAINT; Schema: public; Owner: bloodhound
--

ALTER TABLE ONLY public.users_roles
    ADD CONSTRAINT fk_users_roles_role FOREIGN KEY (role_id) REFERENCES public.roles(id);


--
-- Name: users_roles fk_users_roles_user; Type: FK CONSTRAINT; Schema: public; Owner: bloodhound
--

ALTER TABLE ONLY public.users_roles
    ADD CONSTRAINT fk_users_roles_user FOREIGN KEY (user_id) REFERENCES public.users(id);


--
-- Name: users fk_users_saml_provider; Type: FK CONSTRAINT; Schema: public; Owner: bloodhound
--

ALTER TABLE ONLY public.users
    ADD CONSTRAINT fk_users_saml_provider FOREIGN KEY (saml_provider_id) REFERENCES public.saml_providers(id);


--
-- Name: saved_queries_permissions saved_queries_permissions_query_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: bloodhound
--

ALTER TABLE ONLY public.saved_queries_permissions
    ADD CONSTRAINT saved_queries_permissions_query_id_fkey FOREIGN KEY (query_id) REFERENCES public.saved_queries(id) ON DELETE CASCADE;


--
-- Name: saved_queries_permissions saved_queries_permissions_shared_to_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: bloodhound
--

ALTER TABLE ONLY public.saved_queries_permissions
    ADD CONSTRAINT saved_queries_permissions_shared_to_user_id_fkey FOREIGN KEY (shared_to_user_id) REFERENCES public.users(id) ON DELETE CASCADE;


--
-- PostgreSQL database dump complete
--

