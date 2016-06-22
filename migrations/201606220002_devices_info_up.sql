-- Add last_checkin, assume timestamp will be UTC
ALTER TABLE devices
  ADD COLUMN last_checkin timestamp,
  ADD COLUMN device_name text NOT NULL DEFAULT '';
