# Graphify

Graphify is a tool used to ingest json files into Postgres and runs analysis on the nodes and edges ingested. It creates two generic ingestible output files -- one for ingest, and another for analysis.

The following environment variables are required:

`SB_PG_CONNECTION`: This environment variable should contain the Postgres connection string for the database you want to interact with.

- Example: `SB_PG_CONNECTION="user=XYZ password=XYZ dbname=XYZ host=XYZ port=XYZ" just bhe-graphify lib/go/daemons/datapipe/fixtures/AzureJSON/raw tmp`

The following flags are required:

- `--path`: Specifies the input directory for the consumed files.

The following flags are supported:

- `--outpath`: Specifies the output directory for generic ingestible graph file.
