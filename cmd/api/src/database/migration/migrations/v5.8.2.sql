ALTER TABLE ad_data_quality_stats
  ADD COLUMN IF NOT EXISTS issuancepolicies BIGINT DEFAULT 0;

ALTER TABLE ad_data_quality_aggregations
  ADD COLUMN IF NOT EXISTS issuancepolicies BIGINT DEFAULT 0;
