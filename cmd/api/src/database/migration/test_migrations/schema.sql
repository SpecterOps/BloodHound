-- Test Schema
CREATE TABLE IF NOT EXISTS migrations (
    id integer NOT NULL,
    updated_at timestamp with time zone,
    major integer,
    minor integer,
    patch integer
);

CREATE SEQUENCE IF NOT EXISTS migrations_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;
ALTER SEQUENCE migrations_id_seq OWNED BY migrations.id;

ALTER TABLE ONLY migrations ALTER COLUMN id SET DEFAULT nextval('migrations_id_seq'::regclass);
ALTER TABLE ONLY migrations
    ADD CONSTRAINT migrations_pkey PRIMARY KEY (id);

CREATE TABLE IF NOT EXISTS ingest_tasks (
    id integer
)