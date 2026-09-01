-- +goose Up
UPDATE feature_flags
SET enabled = TRUE
WHERE
  key = 'collector_support_bundle'
  OR key = 'artifact_expiration_cleanup';

-- +goose Down
UPDATE feature_flags
SET enabled = FALSE
WHERE
  key = 'collector_support_bundle'
  OR key = 'artifact_expiration_cleanup';
