# Graphify

Graphify is a tool used to streamline the process of working with graph data.

This tool ingests user-specified JSON files into PostgreSQL, then performs graph analysis on the ingested data. It outputs two graph-ready files: one representing the raw ingested graph, and another containing the analyzed graph.

The following environment variables are required:

`SB_PG_CONNECTION`: This environment variable should contain the Postgres connection string for the database you want to interact with.

The following flags are required:

-   `-path`: Specifies the directory where files should be consumed from and written to
    -   This path should include a `raw` directory containing raw sharphound files, and will write out a directory each for `ingested` and `analyzed` files.

## Usage

Example: `just bh-graphify cmd/api/src/services/graphify/fixtures/Version6JSON`
