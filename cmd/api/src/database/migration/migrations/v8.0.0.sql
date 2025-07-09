create table if not exists source_kinds
(
  id   smallserial,
  name varchar(256) not null,
  primary key (id),
  unique (name)
);

INSERT INTO source_kinds (name)
VALUES 
  ('Base'),
  ('AZBase')
ON CONFLICT (name) DO NOTHING;
