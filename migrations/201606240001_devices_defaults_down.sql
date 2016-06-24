ALTER TABLE devices
  ALTER COLUMN last_query_response DROP NOT NULL,
  ALTER COLUMN build_version DROP NOT NULL,
  ALTER COLUMN product_name DROP NOT NULL,
  ALTER COLUMN os_version DROP NOT NULL;

ALTER TABLE devices
  ALTER COLUMN last_query_response DROP DEFAULT,
  ALTER COLUMN build_version DROP DEFAULT,
  ALTER COLUMN product_name DROP DEFAULT,
  ALTER COLUMN os_version DROP DEFAULT;
