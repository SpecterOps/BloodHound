-- OIDC Provider
CREATE TABLE IF NOT EXISTS oidc_providers
(
  id         BIGSERIAL PRIMARY KEY,
  name       TEXT NOT NULL,
  client_id  TEXT NOT NULL,
  issuer     TEXT NOT NULL,

  updated_at TIMESTAMP WITH TIME ZONE DEFAULT now(),
  created_at TIMESTAMP WITH TIME ZONE DEFAULT now(),

  UNIQUE (name)
);
