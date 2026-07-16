# Audit Logs Partitioning Migration Procedure

**Related RFC:** [bh-rfc-7.md](../../../../rfc/bh-rfc-7.md)
**Status:** Implementation Guidance
**Last Updated:** 2026-06-16

## Overview

This document provides the step-by-step SQL procedure for converting the `audit_logs` table from a regular table to a range-partitioned table. This is referenced by RFC 7 Section 7.2.

---

## Migration Constraints

### PostgreSQL Limitation

**PostgreSQL cannot convert a populated regular table into a partitioned table in place.** The table must be recreated.

### Strategy: Rename-and-Swap

1. Rename existing table aside
2. Create new partitioned table with same name
3. Backfill data from old table to new table
4. Drop old table after verification

---

## Migration File Location

```
cmd/api/src/database/migration/migrations/YYYYMMDDHHMMSS_v10_audit_logs_partitioning.sql
```

**Naming:**
- Timestamp: Current date/time in `YYYYMMDDHHMMSS` format
- Version: `v10` (or current major version)
- Name: `audit_logs_partitioning`

**Annotations:**
- `-- +goose Up` / `-- +goose Down` sections
- `-- +goose StatementBegin` / `-- +goose StatementEnd` around PL/pgSQL blocks

---

## Current Table Schema

From baseline `00000000000001_init.sql`:

```sql
CREATE TABLE audit_logs (
    id                  BIGINT PRIMARY KEY DEFAULT nextval('audit_logs_id_seq'),
    created_at          TIMESTAMPTZ,  -- Currently NULLABLE
    action              TEXT NOT NULL,
    actor_id            TEXT,
    actor_name          TEXT,
    actor_email         TEXT,
    request_id          TEXT,
    source_ip_address   TEXT,
    status              VARCHAR(15) DEFAULT 'intent' CHECK (status IN ('intent', 'success', 'failure')),
    commit_id           TEXT,
    fields              JSONB
);

CREATE INDEX idx_audit_logs_created_at ON audit_logs(created_at);
CREATE INDEX idx_audit_logs_actor_id ON audit_logs(actor_id);
CREATE INDEX idx_audit_logs_action ON audit_logs(action);
```

**GORM model also references these indexes.**

---

## Migration Procedure (`-- +goose Up`)

### Step 1: Rename Existing Table

```sql
ALTER TABLE audit_logs RENAME TO audit_logs_legacy;
```

### Step 2: Create Partitioned Table

```sql
CREATE TABLE audit_logs (
    id                  BIGINT NOT NULL DEFAULT nextval('audit_logs_id_seq'),
    created_at          TIMESTAMPTZ NOT NULL DEFAULT now(),  -- NOW NOT NULL with default
    action              TEXT NOT NULL,
    actor_id            TEXT,
    actor_name          TEXT,
    actor_email         TEXT,
    request_id          TEXT,
    source_ip_address   TEXT,
    status              VARCHAR(15) DEFAULT 'intent' CHECK (status IN ('intent', 'success', 'failure')),
    commit_id           TEXT,
    fields              JSONB,
    source              VARCHAR(20) DEFAULT 'middleware',  -- NEW: dual-write tagging
    PRIMARY KEY (id, created_at)  -- Partition key must be in PK
) PARTITION BY RANGE (created_at);
```

**Key changes:**
- `created_at` is now `NOT NULL` with `DEFAULT now()`
- Primary key is `(id, created_at)` (composite)
- New `source` column for dual-write tagging
- `PARTITION BY RANGE (created_at)` declaration

### Step 3: Re-point Sequence

```sql
ALTER SEQUENCE audit_logs_id_seq OWNED BY audit_logs.id;
```

Ensures `audit_logs_id_seq` continues to auto-assign `id` for the new table.

### Step 4: Declare Indexes on Parent

```sql
CREATE INDEX idx_audit_logs_created_at ON audit_logs(created_at);
CREATE INDEX idx_audit_logs_actor_id ON audit_logs(actor_id);
CREATE INDEX idx_audit_logs_action ON audit_logs(action);
CREATE INDEX idx_audit_logs_source ON audit_logs(source);  -- NEW
```

On PostgreSQL 11+, these propagate to every partition automatically.

### Step 5: Pre-create Partitions

**Determine date range of legacy data:**

```sql
SELECT MIN(created_at), MAX(created_at) FROM audit_logs_legacy WHERE created_at IS NOT NULL;
```

**Create partitions covering legacy range + current month + next month:**

```sql
-- Example: Legacy data spans 2024-01 to 2026-05, current month is 2026-06

-- +goose StatementBegin
DO $$
DECLARE
    start_date DATE := '2024-01-01';
    end_date DATE := '2026-08-01';  -- Current + next month
    current_month DATE;
BEGIN
    current_month := start_date;
    WHILE current_month < end_date LOOP
        EXECUTE format(
            'CREATE TABLE IF NOT EXISTS audit_logs_%s PARTITION OF audit_logs
             FOR VALUES FROM (%L) TO (%L)',
            to_char(current_month, 'YYYY_MM'),
            current_month,
            current_month + interval '1 month'
        );
        current_month := current_month + interval '1 month';
    END LOOP;
END $$;
-- +goose StatementEnd
```

**Create DEFAULT partition for operational safety:**

```sql
CREATE TABLE audit_logs_default PARTITION OF audit_logs DEFAULT;
```

Prevents inserts from failing on missing partitions. The maintenance worker pre-creates bounded partitions, but DEFAULT provides fallback.

### Step 6: Backfill Data

```sql
INSERT INTO audit_logs (
    id, created_at, action, actor_id, actor_name, actor_email,
    request_id, source_ip_address, status, commit_id, fields, source
)
SELECT
    id,
    COALESCE(created_at, '2020-01-01'::timestamptz) AS created_at,  -- Handle nulls
    action,
    actor_id,
    actor_name,
    actor_email,
    request_id,
    source_ip_address,
    status,
    commit_id,
    fields,
    'legacy' AS source  -- Tag all legacy rows
FROM audit_logs_legacy;
```

**For large tables:** Batch the copy in chunks (e.g., 10k rows at a time) to avoid long-running transaction locks.

### Step 7: Verify Backfill

```sql
SELECT COUNT(*) FROM audit_logs;
SELECT COUNT(*) FROM audit_logs_legacy;
-- Counts should match
```

### Step 8: Drop Legacy Table

```sql
DROP TABLE audit_logs_legacy;
```

**Alternative:** Defer this to a follow-up migration if a verification window is desired.

---

## Migration Rollback (`-- +goose Down`)

### Recreate Flat Table

```sql
DROP TABLE IF EXISTS audit_logs CASCADE;

CREATE TABLE audit_logs (
    id                  BIGINT PRIMARY KEY DEFAULT nextval('audit_logs_id_seq'),
    created_at          TIMESTAMPTZ,
    action              TEXT NOT NULL,
    actor_id            TEXT,
    actor_name          TEXT,
    actor_email         TEXT,
    request_id          TEXT,
    source_ip_address   TEXT,
    status              VARCHAR(15) DEFAULT 'intent' CHECK (status IN ('intent', 'success', 'failure')),
    commit_id           TEXT,
    fields              JSONB
);
```

**Note:** `source` column is dropped in the rollback. Any rows written after `Up` with `source = 'middleware'` will lose that metadata.

### Recreate Indexes

```sql
CREATE INDEX idx_audit_logs_created_at ON audit_logs(created_at);
CREATE INDEX idx_audit_logs_actor_id ON audit_logs(actor_id);
CREATE INDEX idx_audit_logs_action ON audit_logs(action);
```

### Copy Data Back (Best Effort)

```sql
-- If audit_logs_legacy still exists (Step 8 was deferred)
INSERT INTO audit_logs SELECT id, created_at, action, actor_id, actor_name, actor_email, 
    request_id, source_ip_address, status, commit_id, fields 
FROM audit_logs_legacy;
```

**Note:** If Step 8 already dropped `audit_logs_legacy`, this rollback only recovers data that existed at migration time. New rows written after migration are preserved.

---

## Post-Migration Validation

1. **Row count matches:** `SELECT COUNT(*) FROM audit_logs;`
2. **Partitions created:** `SELECT tablename FROM pg_tables WHERE tablename LIKE 'audit_logs_%';`
3. **Indexes exist:** `SELECT indexname FROM pg_indexes WHERE tablename = 'audit_logs';`
4. **New writes succeed:** Insert a test row and verify it lands in correct partition
5. **Query performance:** Run `EXPLAIN` on typical audit log queries to verify partition pruning

---

## References

- **RFC 7 Section 7.2:** Migration mechanics and design rationale
- **Partition lifecycle:** Managed by `audit.Maintainer` interface (see `module-structure.md`)
- **Goose migration docs:** `bhce/AGENTS.md` (migration rules)
