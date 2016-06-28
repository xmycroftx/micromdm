-- Add the UUID and Serial Number back as individual UNIQUE constraints, along with others
DROP INDEX udid_serial_idx;

CREATE UNIQUE INDEX IF NOT EXISTS serial_idx ON devices (serial_number);
CREATE UNIQUE INDEX IF NOT EXISTS udid_idx ON devices (udid);
