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
)


-- Add Scheduled Analysis Configs
  INSERT INTO parameters (key, name, description, value, created_at, updated_at)
  VALUES ('analysis.scheduled', 'Scheduled Analysis', 'This configuration parameter allows setting a schedule for analysis. When enabled, analysis will only run when the scheduled time arrives','{"enabled": false, "rrule": "", "last_run": "2006-01-02T15:04:05Z"}',current_timestamp,current_timestamp) ON CONFLICT DO NOTHING;