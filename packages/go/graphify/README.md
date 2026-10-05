# Graphify

Graphify is a fixtures testing tool that generates the expected graph files used
by BloodHound ingest and analysis integration tests. These are commonly referred
to as "golden files."

Graphify runs the data in a fixture's `raw/` directory through ingest and
analysis to produce golden files that integration tests compare against:

- `ingest/ingested.json` contains the expected graph after ingest.
- `analysis/analyzed.json` contains the expected graph after analysis.

Regenerate the golden files when the raw fixture data changes or when an ingest,
analysis, or post-processing change in the codebase affects the expected outcome.

## Fixture layout

The value passed to `-path` must be the root of a fixture set with this layout:

```text
<fixture>/
├── raw/                   # Input JSON collection files
├── ingest/
│   └── ingested.json     # Generated ingest expectation
└── analysis/
    └── analyzed.json     # Generated analysis expectation
```

Graphify creates the `ingest/` and `analysis/` directories if they do not exist
and overwrites the generated files when they do exist. Only `.json` files under
`raw/` are processed.

## Prerequisites

Graphify requires a PostgreSQL database and reads its connection string from
`SB_PG_CONNECTION`. The `just` recipes provide the connection string for the
repository's testing database by default.

> [!WARNING]
> Graphify wipes the configured database as part of its generation lifecycle.
> Use the local testing database when running it.

Start the testing database from the BloodHound repository root:

```sh
just bh-testing
```

## Run Graphify

From the `bhce` repository root, pass the fixture-set root (not its `raw/`
directory) to the `bh-graphify` recipe:

```sh
just bh-graphify cmd/api/src/services/graphify/fixtures/Version6JSON
```

The equivalent direct invocation is:

```sh
go run github.com/specterops/bloodhound/packages/go/graphify \
  -path=cmd/api/src/services/graphify/fixtures/Version6JSON
```

After generation, review the changes to `ingested.json` and `analyzed.json` to
ensure they reflect the intended ingest or analysis behavior, then run the
affected slow integration tests.

For the origin and intended ownership of individual fixture datasets, see the
[fixture provenance documentation](../../../cmd/api/src/services/graphify/fixtures/README.md).
