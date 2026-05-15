# Goose Migrations

Goose is a database migration tool that manages schema changes through versioned SQL files. Migrations are tracked via the
`goose_db_version` table, which serves as the source of truth. Every migration should have its own file. These files have a timestamp
based naming convention(`YYYYMMDDHHMMSS_v#_description.sql`) which helps ensure that migrations are applied in the correct order and prevents
conflicts when multiple engineers are working simultaneously. Within each file, schema changes are split into two sections: `-- +goose Up`
for applying a new change and `-- +goose Down` for rolling back. On the application startup, goose will compare what is currently in the
table against any new available migration files and automatically applies pending migrations in the correct order.

## Migration Guidelines

**File Naming Convention**

-   Files can be created via the command `just goose-create <name>`. This will create a new file for you in the `migrations` folder in
    the timestamp format: `YYYYMMDDHHMMSS_name.sql`
    -   e.g. `20060102162005_v9_add_column_users.sql`
-   Add the current major version at the beginning of the name chosen. e.g, `just goose-create v9_add_column_users`
-   Choose a migration name that clearly describes the schema change being made
-   Use underscores instead of hyphens when picking a description/name

**Required Annotations**

-   All migration files must have an Up and Down command. It should look like:

    ```sql
    -- +goose Up
    CREATE TABLE example (id INT);
    -- +goose Down
    DROP TABLE IF EXISTS example;
    ```

**Statement Delimiters**

-   Statement delimiters are needed for functions, triggers and DO blocks

    ```sql
        -- +goose StatementBegin
        CREATE FUNCTION ...
        -- +goose StatementEnd
    ```

## Common Rules

-   Never modify the baseline migration(`00000000000001_init.sql`). This is a migration file all of the previous schemas before switching to goose.
-   Never modify migrations already merged into main or staging branches. Create a new migration if a fix is needed
-   Never modify the `goose_db_version` table
-   Always include a `Down` migration when possible
-   `Down` should safely reverse an `Up`, e.g., use `IF EXISTS`, `ON CONFLICT`, etc.
-   Always pull main before creating a new migration
-   If you are editing an existing migration file that has already been applied locally (but not yet merged to main), you must run `just goose-down` followed by `just goose-up` to roll back and re-apply the migration with your changes, since goose only tracks version numbers and will not detect file content changes.

## How to use in the codebase

**Local commands:**

-   `just goose-create <name>` - create a new migration file
-   `just goose-status` - see applied migrations
-   `just goose-up` - apply pending migrations
-   `just goose-up-by-one` - apply next pending migration
-   `just goose-up-to version` - apply up to a specific version
-   `just goose-down` - rollback last migration
-   `just goose-down-to version` - rollback to specified version (non-inclusive)
-   `just goose-down-all` - rollback all versions

**Common Workflow:**

-   `just goose-create <name>` - This creates a file with boilerplate SQL
-   Make edits to the desired migration file
-   `just goose-up` - Applies migrations
-   If modifications are needed to an already-applied local migration:
    -   Make edits
    -   `just goose-down` - Rolls back the most recent migration (goose tracks versions, not file contents)
        -   `just goose-down-to <timestamp>` - Use this if you need to roll back multiple local migrations
    -   `just goose-up` - Re-applies the modified migration/s

## Troubleshooting

### Migration Fails Mid Run

**Symptom:** `goose up` errors out partway through, leaving the database in a partial state.

**Cause:** A SQL error in the migration file — bad syntax, constraint violation, duplicate column etc.

**Outcome:** Because goose wraps each migration in a transaction, the failed migration is rolled back
automatically. The database is left at the last successfully applied version. Fix the migration file
and run `just goose-up` again.

### Checksum Mismatch / Accidentally Modified a Merged Migration

**Symptom:**

```sql
error: checksum mismatch for migration 202601010000000
```

**Cause:** A migration file that was already applied has been modified on disk. Goose detects the file
no longer matches what was applied.

**Outcome:** Goose refuses to run. Revert the file to its original state. Then create a new migration for the change you intended to make. Never modify a migration that has already been merged into main.

## Configuration Notes

Goose is configured with two non-default behaviors:

-   **Allow Missing** — goose will not error if migration files are missing
    from disk but exist in `goose_db_version`.

-   **Allow Out of Order** — goose will apply migrations even if they are
    out of sequence. This reduces friction when multiple developers are
    working on migrations simultaneously. Note that migrations which are
    logically dependent on each other should still be coordinated manually.
