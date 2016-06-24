UPDATE devices SET build_version = '' WHERE build_version IS NULL;
UPDATE devices SET product_name = '' WHERE product_name IS NULL;
UPDATE devices SET os_version = '' WHERE os_version IS NULL;

ALTER TABLE devices
    ALTER COLUMN build_version SET DEFAULT '',
    ALTER COLUMN product_name SET DEFAULT '',
    ALTER COLUMN os_version SET DEFAULT '';

ALTER TABLE devices
  ALTER COLUMN build_version SET NOT NULL,
  ALTER COLUMN product_name SET NOT NULL,
  ALTER COLUMN os_version SET NOT NULL;
