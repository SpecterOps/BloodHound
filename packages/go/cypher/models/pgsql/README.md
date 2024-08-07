# BloodHound PgSQL Model

This package contains a syntax model for the PostgreSQL SQL dialect. It also contains a translation implementation to
take openCpyher input and output valid PostgreSQL SQL. This model is not intended to be a complete implementation of all
available SQL dialect features but rather the subset of the dialect required to perform openCypher to pgsql translation.

**Expected PostgreSQL SQL dialect version**: `16.X`

## Formatting

The `format` package contains the string rendering logic for the PgSQL syntax model.

## Translation

The `translate` package contains the openCypher translation implementation.

## Visualization

The `visualization` package contains a PUML digraph formatter for the PgSQL syntax model.

## Test Cases

The `test` package contains the test cases used to validate translation.
