# Graphify test fixtures

This directory contains the inputs and golden files used by the Graphify ingest and analysis integration tests:

- `raw/` contains the test inputs and is the source of truth for each fixture scenario.
- `ingest/ingested.json` is generated from `raw/` and captures the expected graph immediately after ingest.
- `analysis/analyzed.json` is generated after analysis and post-processing and captures the final expected graph.

Most raw inputs were copied or adapted from the test data under `cmd/api/src/test/fixtures/fixtures/`. The copies in this directory are owned by the Graphify tests and do not automatically track changes to their original source.

Start the testing database and run Graphify from the BloodHound repository root:

```sh
just bh-testing
just bh-graphify cmd/api/src/services/graphify/fixtures/Version6JSON
```

See the [Graphify README](../../../../../../packages/go/graphify/README.md) for more information. Review the resulting golden-file changes and run the affected slow integration tests instead of editing generated files by hand.
