# Graphify

Graphify is a tool used to streamline the process of working with graph data.

This tool ingests user-specified JSON files into PostgreSQL, then performs graph analysis on the ingested data. It outputs two graph-ready files: one representing the raw ingested graph, and another containing the analyzed graph.

The following environment variables are required:

`SB_PG_CONNECTION`: This environment variable should contain the Postgres connection string for the database you want to interact with.

The following flags are required:

- `-path`: Specifies the input directory for the consumed files.

The following flags are supported:

- `-outpath`: Specifies the output directory for generic ingestible graph files. These files will output with static names e.g. ingested.json & analyzed.json. Default is `{root}/tmp/`.

## Usage

Example: `just bh-graphify cmd/api/src/test/fixtures/fixtures/v6/ingest /tmp/`
