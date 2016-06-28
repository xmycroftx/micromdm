-- Add the UUID and Serial Number back as individual UNIQUE constraints, along with others
DROP INDEX udid_serial_idx;

ALTER TABLE devices
    ALTER COLUMN udid DROP NOT NULL;

UPDATE devices SET udid = NULL WHERE udid = '';
UPDATE devices SET serial_number = NULL WHERE serial_number = '';

CREATE UNIQUE INDEX IF NOT EXISTS serial_idx ON devices (serial_number);
CREATE UNIQUE INDEX IF NOT EXISTS udid_idx ON devices (udid);
