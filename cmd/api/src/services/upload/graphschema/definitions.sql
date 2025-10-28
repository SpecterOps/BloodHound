-- CORE TABLES
create table if not exists collectors
    (
        id serial not null,
        name text unique not null, -- sharp-hound
        display_name text not null, -- SharpHound
        version text not null, -- v0.0.0
        created_at timestamp with time zone default current_timestamp,

        primary key (id)
    );

create table if not exists graph_schemas
    (
        id serial not null,
        collector_id integer not null references collectors (id) on delete cascade,
        version text not null, -- v0.0.0, different from collector version
        description text,
        created_at timestamp with time zone default current_timestamp,
        
        primary key (id)
    );

create table if not exists schema_properties (
    id serial not null,
    schema_id integer not null references graph_schemas (id) on delete cascade,
    symbol text not null, -- used as variable name in generated code
    display_name text not null, -- human-readable
    representation text not null, -- actual string stored in property bag
    data_type text not null, -- string, int, float, bool, array
    description text,
    created_at timestamp with time zone default current_timestamp,

    primary key (id)
    unique(schema_id, symbol)
);

create table if not exists schema_node_kinds (
    id serial not null,
    schema_id integer not null references graph_schemas (id) on delete cascade,
    symbol text not null,
    display_name text not null,
    description text,
    is_source boolean not null default false,
    icon text,
    color text,
    created_at timestamp with time zone default current_timestamp,

    primary key (id),
    unique (schema_id, symbol)
);

create table if not exists schema_relationship_kinds (
    id serial not null,
    schema_id integer not null references graph_schemas (id) on delete cascade,
    symbol text not null,
    display_name text not null,
    description text,
    created_at timestamp with time zone default current_timestamp,

    primary key (id),
    unique (schema_id, symbol)
);

-- CATEGORIZATION TABLES
create table if not exists schema_categories (
    id serial not null,
    schema_id integer not null references graph_schemas (id) on delete cascade,
    name text not null,
    description text,
    category_type text not null,
    created_at timestamp with time zone default current_timestamp,
    
    primary key (id),
    unique (schema_id, name)
);

create table if not exists schema_relationship_categories (
    category_id integer not null references schema_categories (id) on delete cascade,
    relationship_id integer not null references schema_relationship_kinds (id) on delete cascade,
    
    created_at timestamp with time zone default current_timestamp,
    primary key (category_id, relationship_id)
);

-- CONSTRAINT TABLES
create table if not exists schema_node_properties (
    node_id integer not null references schema_node_kinds (id) on delete cascade,
    property_id integer not null references schema_properties (id) on delete cascade,
    is_required boolean not null default false,
    created_at timestamp with time zone default current_timestamp,
    
    primary key (node_id, property_id)
);

create table if not exists schema_relationship_constraints (
    id serial not null,
    relationship_id integer not null references schema_relationship_kinds (id) on delete cascade,
    source_id integer not null references schema_node_kinds (id) on delete cascade,
    target_id integer not null references schema_node_kinds (id) on delete cascade,
    created_at timestamp with time zone default current_timestamp,
    primary key (id),
    unique (relationship_id, source_id, target_id)
);
