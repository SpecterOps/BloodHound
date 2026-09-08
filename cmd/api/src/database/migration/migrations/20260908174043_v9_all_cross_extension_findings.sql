-- +goose Up
-- schema_findings.environment_id now stores the environment kind ID from dawgs rather than
-- the schema_environments ID. This permits findings to reference environment kinds defined by another extension.
ALTER TABLE schema_findings
    DROP CONSTRAINT IF EXISTS schema_findings_environment_id_fkey;

UPDATE schema_findings sf
SET environment_id = se.environment_kind_id
FROM schema_environments se
WHERE sf.environment_id = se.id;

ALTER TABLE schema_findings
    ADD CONSTRAINT schema_findings_environment_id_fkey
    FOREIGN KEY (environment_id) REFERENCES kind(id) ON DELETE RESTRICT;

-- +goose Down
-- A rollback would only work while every referenced environment kind still has a
-- schema_environments mapping to restore the original foreign key.
-- +goose StatementBegin
DO $$
BEGIN
    IF EXISTS (
        SELECT 1
        FROM schema_findings sf
        LEFT JOIN schema_environments se ON se.environment_kind_id = sf.environment_id
        WHERE se.id IS NULL
    ) THEN
        RAISE EXCEPTION 'cannot restore schema_findings.environment_id to schema_environments: referenced environment kinds have no schema environment mapping';
    END IF;
END
$$;
-- +goose StatementEnd

ALTER TABLE schema_findings
    DROP CONSTRAINT IF EXISTS schema_findings_environment_id_fkey;

UPDATE schema_findings sf
SET environment_id = se.id
FROM schema_environments se
WHERE sf.environment_id = se.environment_kind_id;

ALTER TABLE schema_findings
    ADD CONSTRAINT schema_findings_environment_id_fkey
    FOREIGN KEY (environment_id) REFERENCES schema_environments(id) ON DELETE CASCADE;
