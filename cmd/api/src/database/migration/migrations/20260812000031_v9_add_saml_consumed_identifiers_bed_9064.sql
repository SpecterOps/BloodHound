-- +goose Up
CREATE TABLE IF NOT EXISTS saml_consumed_identifiers
(
  sso_provider_id INTEGER                  NOT NULL REFERENCES sso_providers (id) ON DELETE CASCADE,
  idp_issuer      VARCHAR(1024)            NOT NULL,
  identifier_type TEXT                     NOT NULL CHECK (identifier_type IN ('response', 'assertion')),
  identifier      VARCHAR(255)             NOT NULL,
  expires_at      TIMESTAMP WITH TIME ZONE NOT NULL,
  consumed_at     TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),

  PRIMARY KEY (sso_provider_id, idp_issuer, identifier_type, identifier)
);

CREATE INDEX IF NOT EXISTS idx_saml_consumed_identifiers_expires_at
  ON saml_consumed_identifiers (expires_at);

-- +goose Down
DROP TABLE IF EXISTS saml_consumed_identifiers;
