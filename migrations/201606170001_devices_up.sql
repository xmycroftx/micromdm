CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE TABLE IF NOT EXISTS devices (
  device_uuid uuid PRIMARY KEY DEFAULT uuid_generate_v4(),
  udid text NOT NULL DEFAULT '',
  serial_number text,
  os_version text,
  model text NOT NULL DEFAULT '',
  color text,
  asset_tag text,
  dep_profile_status text,
  dep_profile_uuid text,
  dep_profile_assign_time date,
  dep_profile_push_time date,
  dep_profile_assigned_date date,
  dep_profile_assigned_by text,
  description text NOT NULL DEFAULT '',
  build_version text,
  product_name text,
  imei text NOT NULL DEFAULT '',
  meid text NOT NULL DEFAULT '',
  apple_mdm_token text,
  apple_mdm_topic text,
  apple_push_magic text,
  mdm_enrolled boolean,
  workflow_uuid text NOT NULL DEFAULT '',
  dep_device boolean,
  awaiting_configuration boolean
);

CREATE UNIQUE INDEX IF NOT EXISTS serial_idx ON devices (serial_number);
