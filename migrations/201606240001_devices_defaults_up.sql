ALTER TABLE devices
    ALTER COLUMN last_query_response SET DEFAULT '{}',
    ALTER COLUMN build_version SET DEFAULT '',
    ALTER COLUMN product_name SET DEFAULT '',
    ALTER COLUMN os_version SET DEFAULT '';

ALTER TABLE devices
  ALTER COLUMN last_query_response SET NOT NULL,
  ALTER COLUMN build_version SET NOT NULL,
  ALTER COLUMN product_name SET NOT NULL,
  ALTER COLUMN os_version SET NOT NULL;

