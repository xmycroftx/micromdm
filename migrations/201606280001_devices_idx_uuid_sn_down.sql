DROP INDEX IF EXISTS serial_idx;
DROP INDEX IF EXISTS udid_idx;

CREATE UNIQUE INDEX IF NOT EXISTS udid_serial_idx ON devices (udid, serial_number);