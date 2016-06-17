CREATE TABLE IF NOT EXISTS profiles (
  profile_uuid uuid PRIMARY KEY DEFAULT uuid_generate_v4(),
  payload_identifier text UNIQUE NOT NULL CHECK (payload_identifier <> ''),
  profile_data bytea
);
