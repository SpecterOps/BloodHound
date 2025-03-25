--
-- PostgreSQL database dump
--

-- Dumped from database version 14.17 (Homebrew)
-- Dumped by pg_dump version 14.17 (Homebrew)

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
-- Data for Name: datapipe_status; Type: TABLE DATA; Schema: public; Owner: bloodhound
--

INSERT INTO public.datapipe_status VALUES (true, 'idle', '2025-03-21 12:58:42.971311-05', NULL);


--
-- Data for Name: feature_flags; Type: TABLE DATA; Schema: public; Owner: bloodhound
--

INSERT INTO public.feature_flags VALUES (1, '2025-03-21 12:56:17.075059-05', '2025-03-21 12:56:17.075059-05', 'butterfly_analysis', 'Enhanced Asset Inbound-Outbound Exposure Analysis', 'Enables more extensive analysis of attack path findings that allows BloodHound to help the user prioritize remediation of the most exposed assets.', false, false);
INSERT INTO public.feature_flags VALUES (2, '2025-03-21 12:56:17.075474-05', '2025-03-21 12:56:17.075474-05', 'enable_saml_sso', 'SAML Single Sign-On Support', 'Enables SSO authentication flows and administration panels to third party SAML identity providers.', true, false);
INSERT INTO public.feature_flags VALUES (3, '2025-03-21 12:56:17.07564-05', '2025-03-21 12:56:17.07564-05', 'scope_collection_by_ou', 'Enable SharpHound OU Scoped Collections', 'Enables scoping SharpHound collections to specific lists of OUs.', true, false);
INSERT INTO public.feature_flags VALUES (4, '2025-03-21 12:56:17.075794-05', '2025-03-21 12:56:17.075794-05', 'azure_support', 'Enable Azure Support', 'Enables Azure support.', true, false);
INSERT INTO public.feature_flags VALUES (5, '2025-03-21 12:56:17.075951-05', '2025-03-21 12:56:17.075951-05', 'reconciliation', 'Reconciliation', 'Enables Reconciliation', true, false);
INSERT INTO public.feature_flags VALUES (6, '2025-03-21 12:56:17.076129-05', '2025-03-21 12:56:17.076129-05', 'entity_panel_cache', 'Enable application level caching', 'Enables the use of application level caching for entity panel queries', true, false);
INSERT INTO public.feature_flags VALUES (7, '2025-03-21 12:56:17.076293-05', '2025-03-21 12:56:17.076293-05', 'adcs', 'Enable collection and processing of Active Directory Certificate Services Data', 'Enables the ability to collect, analyze, and explore Active Directory Certificate Services data and previews new attack paths.', false, false);
INSERT INTO public.feature_flags VALUES (8, '2025-03-21 12:58:42.987985-05', '2025-03-21 12:58:42.987985-05', 'dark_mode', 'Dark Mode', 'Allows users to enable or disable dark mode via a toggle in the settings menu', false, true);
INSERT INTO public.feature_flags VALUES (9, '2025-03-21 12:58:43.022829-05', '2025-03-21 12:58:43.022829-05', 'pg_migration_dual_ingest', 'PostgreSQL Migration Dual Ingest', 'Enables dual ingest pathing for both Neo4j and PostgreSQL.', false, false);
INSERT INTO public.feature_flags VALUES (10, '2025-03-21 12:58:43.038933-05', '2025-03-21 12:58:43.038933-05', 'clear_graph_data', 'Clear Graph Data', 'Enables the ability to delete all nodes and edges from the graph database.', true, false);
INSERT INTO public.feature_flags VALUES (11, '2025-03-21 12:58:43.039466-05', '2025-03-21 12:58:43.039466-05', 'risk_exposure_new_calculation', 'Use new tier zero risk exposure calculation', 'Enables the use of new tier zero risk exposure metatree metrics.', false, false);
INSERT INTO public.feature_flags VALUES (12, '2025-03-21 12:58:43.039655-05', '2025-03-21 12:58:43.039655-05', 'fedramp_eula', 'FedRAMP EULA', 'Enables showing the FedRAMP EULA on every login. (Enterprise only)', false, false);
INSERT INTO public.feature_flags VALUES (13, '2025-03-21 12:58:43.039915-05', '2025-03-21 12:58:43.039915-05', 'auto_tag_t0_parent_objects', 'Automatically add parent OUs and containers of Tier Zero AD objects to Tier Zero', 'Parent OUs and containers of Tier Zero AD objects are automatically added to Tier Zero during analysis. Containers are only added if they have a Tier Zero child object with ACL inheritance enabled.', true, true);


--
-- Data for Name: parameters; Type: TABLE DATA; Schema: public; Owner: bloodhound
--

INSERT INTO public.parameters VALUES ('auth.password_expiration_window', 'Local Auth Password Expiry Window', 'This configuration parameter sets the local auth password expiry window for users that have valid auth secrets. Values for this configuration must follow the duration specification of ISO-8601.', '{"duration": "P90D"}', 1, '2025-03-21 12:56:17.07645-05', '2025-03-21 12:56:17.07645-05');
INSERT INTO public.parameters VALUES ('neo4j.configuration', 'Neo4j Configuration Parameters', 'This configuration parameter sets the BatchWriteSize and the BatchFlushSize for Neo4J.', '{"batch_write_size": 20000, "write_flush_size": 100000}', 2, '2025-03-21 12:56:17.077247-05', '2025-03-21 12:56:17.077247-05');


--
-- Data for Name: permissions; Type: TABLE DATA; Schema: public; Owner: bloodhound
--

INSERT INTO public.permissions VALUES ('app', 'ReadAppConfig', 1, '2025-03-21 12:56:17.066242-05', '2025-03-21 12:56:17.066242-05');
INSERT INTO public.permissions VALUES ('app', 'WriteAppConfig', 2, '2025-03-21 12:56:17.066532-05', '2025-03-21 12:56:17.066532-05');
INSERT INTO public.permissions VALUES ('risks', 'GenerateReport', 3, '2025-03-21 12:56:17.066707-05', '2025-03-21 12:56:17.066707-05');
INSERT INTO public.permissions VALUES ('risks', 'ManageRisks', 4, '2025-03-21 12:56:17.066878-05', '2025-03-21 12:56:17.066878-05');
INSERT INTO public.permissions VALUES ('auth', 'CreateToken', 5, '2025-03-21 12:56:17.067068-05', '2025-03-21 12:56:17.067068-05');
INSERT INTO public.permissions VALUES ('auth', 'ManageAppConfig', 6, '2025-03-21 12:56:17.067239-05', '2025-03-21 12:56:17.067239-05');
INSERT INTO public.permissions VALUES ('auth', 'ManageProviders', 7, '2025-03-21 12:56:17.067393-05', '2025-03-21 12:56:17.067393-05');
INSERT INTO public.permissions VALUES ('auth', 'ManageSelf', 8, '2025-03-21 12:56:17.067537-05', '2025-03-21 12:56:17.067537-05');
INSERT INTO public.permissions VALUES ('auth', 'ManageUsers', 9, '2025-03-21 12:56:17.067704-05', '2025-03-21 12:56:17.067704-05');
INSERT INTO public.permissions VALUES ('clients', 'Manage', 10, '2025-03-21 12:56:17.067872-05', '2025-03-21 12:56:17.067872-05');
INSERT INTO public.permissions VALUES ('clients', 'Tasking', 11, '2025-03-21 12:56:17.068033-05', '2025-03-21 12:56:17.068033-05');
INSERT INTO public.permissions VALUES ('collection', 'ManageJobs', 12, '2025-03-21 12:56:17.068194-05', '2025-03-21 12:56:17.068194-05');
INSERT INTO public.permissions VALUES ('graphdb', 'Read', 13, '2025-03-21 12:56:17.068354-05', '2025-03-21 12:56:17.068354-05');
INSERT INTO public.permissions VALUES ('graphdb', 'Write', 14, '2025-03-21 12:56:17.068512-05', '2025-03-21 12:56:17.068512-05');
INSERT INTO public.permissions VALUES ('saved_queries', 'Read', 15, '2025-03-21 12:58:43.040108-05', '2025-03-21 12:58:43.040108-05');
INSERT INTO public.permissions VALUES ('saved_queries', 'Write', 16, '2025-03-21 12:58:43.040475-05', '2025-03-21 12:58:43.040475-05');
INSERT INTO public.permissions VALUES ('clients', 'Read', 17, '2025-03-21 12:58:43.040655-05', '2025-03-21 12:58:43.040655-05');
INSERT INTO public.permissions VALUES ('db', 'Wipe', 18, '2025-03-21 12:58:43.04082-05', '2025-03-21 12:58:43.04082-05');
INSERT INTO public.permissions VALUES ('graphdb', 'Mutate', 19, '2025-03-21 12:58:43.041001-05', '2025-03-21 12:58:43.041001-05');


--
-- Data for Name: roles; Type: TABLE DATA; Schema: public; Owner: bloodhound
--

INSERT INTO public.roles VALUES ('Administrator', 'Can manage users, clients, and application configuration', 1, '2025-03-21 12:56:17.0687-05', '2025-03-21 12:56:17.0687-05');
INSERT INTO public.roles VALUES ('User', 'Can read data, modify asset group memberships', 2, '2025-03-21 12:56:17.069018-05', '2025-03-21 12:56:17.069018-05');
INSERT INTO public.roles VALUES ('Read-Only', 'Used for integrations', 3, '2025-03-21 12:56:17.06919-05', '2025-03-21 12:56:17.06919-05');
INSERT INTO public.roles VALUES ('Upload-Only', 'Used for data collection clients, can post data but cannot read data', 4, '2025-03-21 12:56:17.069362-05', '2025-03-21 12:56:17.069362-05');
INSERT INTO public.roles VALUES ('Power User', 'Can upload data, manage clients, and perform any action a User can', 5, '2025-03-21 12:58:43.041242-05', '2025-03-21 12:58:43.041242-05');


--
-- Data for Name: roles_permissions; Type: TABLE DATA; Schema: public; Owner: bloodhound
--

INSERT INTO public.roles_permissions VALUES (1, 1);
INSERT INTO public.roles_permissions VALUES (1, 2);
INSERT INTO public.roles_permissions VALUES (1, 3);
INSERT INTO public.roles_permissions VALUES (1, 4);
INSERT INTO public.roles_permissions VALUES (1, 5);
INSERT INTO public.roles_permissions VALUES (1, 6);
INSERT INTO public.roles_permissions VALUES (1, 7);
INSERT INTO public.roles_permissions VALUES (1, 8);
INSERT INTO public.roles_permissions VALUES (1, 9);
INSERT INTO public.roles_permissions VALUES (1, 10);
INSERT INTO public.roles_permissions VALUES (1, 11);
INSERT INTO public.roles_permissions VALUES (1, 12);
INSERT INTO public.roles_permissions VALUES (1, 13);
INSERT INTO public.roles_permissions VALUES (1, 14);
INSERT INTO public.roles_permissions VALUES (2, 1);
INSERT INTO public.roles_permissions VALUES (2, 3);
INSERT INTO public.roles_permissions VALUES (2, 5);
INSERT INTO public.roles_permissions VALUES (2, 8);
INSERT INTO public.roles_permissions VALUES (2, 13);
INSERT INTO public.roles_permissions VALUES (3, 1);
INSERT INTO public.roles_permissions VALUES (3, 3);
INSERT INTO public.roles_permissions VALUES (3, 8);
INSERT INTO public.roles_permissions VALUES (3, 13);
INSERT INTO public.roles_permissions VALUES (4, 11);
INSERT INTO public.roles_permissions VALUES (4, 14);
INSERT INTO public.roles_permissions VALUES (1, 15);
INSERT INTO public.roles_permissions VALUES (1, 16);
INSERT INTO public.roles_permissions VALUES (1, 17);
INSERT INTO public.roles_permissions VALUES (1, 18);
INSERT INTO public.roles_permissions VALUES (1, 19);
INSERT INTO public.roles_permissions VALUES (2, 15);
INSERT INTO public.roles_permissions VALUES (2, 16);
INSERT INTO public.roles_permissions VALUES (2, 17);
INSERT INTO public.roles_permissions VALUES (3, 5);
INSERT INTO public.roles_permissions VALUES (5, 1);
INSERT INTO public.roles_permissions VALUES (5, 2);
INSERT INTO public.roles_permissions VALUES (5, 3);
INSERT INTO public.roles_permissions VALUES (5, 4);
INSERT INTO public.roles_permissions VALUES (5, 5);
INSERT INTO public.roles_permissions VALUES (5, 8);
INSERT INTO public.roles_permissions VALUES (5, 10);
INSERT INTO public.roles_permissions VALUES (5, 17);
INSERT INTO public.roles_permissions VALUES (5, 11);
INSERT INTO public.roles_permissions VALUES (5, 12);
INSERT INTO public.roles_permissions VALUES (5, 14);
INSERT INTO public.roles_permissions VALUES (5, 13);
INSERT INTO public.roles_permissions VALUES (5, 15);
INSERT INTO public.roles_permissions VALUES (5, 16);
INSERT INTO public.roles_permissions VALUES (5, 19);


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

SELECT pg_catalog.setval('public.auth_secrets_id_seq', 1, false);


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

SELECT pg_catalog.setval('public.feature_flags_id_seq', 13, true);


--
-- Name: file_upload_jobs_id_seq; Type: SEQUENCE SET; Schema: public; Owner: bloodhound
--

SELECT pg_catalog.setval('public.file_upload_jobs_id_seq', 1, false);


--
-- Name: ingest_tasks_id_seq; Type: SEQUENCE SET; Schema: public; Owner: bloodhound
--

SELECT pg_catalog.setval('public.ingest_tasks_id_seq', 1, false);


--
-- Name: parameters_id_seq; Type: SEQUENCE SET; Schema: public; Owner: bloodhound
--

SELECT pg_catalog.setval('public.parameters_id_seq', 2, true);


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

