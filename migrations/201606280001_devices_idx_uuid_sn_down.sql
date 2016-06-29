DROP INDEX IF EXISTS serial_idx;
DROP INDEX IF EXISTS udid_idx;

UPDATE devices SET udid = '' WHERE udid IS NULL;
UPDATE devices SET serial_number = '' WHERE serial_number IS NULL;

ALTER TABLE devices
  ALTER COLUMN udid SET NOT NULL;

CREATE UNIQUE INDEX IF NOT EXISTS udid_serial_idx ON devices (udid, serial_number);

ALTER TABLE devices
  ALTER COLUMN udid SET DEFAULT '';