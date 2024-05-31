ALTER TABLE relationship_findings ADD COLUMN inbound_cardinality integer;
ALTER TABLE relationship_findings ADD COLUMN outbound_cardinality integer;

ALTER TABLE list_findings ADD COLUMN outbound_cardinality integer;