CREATE TABLE IF NOT EXISTS organizations (
  id TEXT PRIMARY KEY,
  name TEXT NOT NULL,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE TABLE IF NOT EXISTS users (
  id TEXT PRIMARY KEY,
  org_id TEXT REFERENCES organizations(id),
  email TEXT NOT NULL,
  role TEXT NOT NULL,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE TABLE IF NOT EXISTS projects (
  id TEXT PRIMARY KEY,
  name TEXT NOT NULL,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE TABLE IF NOT EXISTS machine_profiles (
  name TEXT PRIMARY KEY,
  base_image TEXT NOT NULL,
  cpu INTEGER NOT NULL,
  memory_mb INTEGER NOT NULL,
  disk_gb INTEGER NOT NULL,
  network_mode TEXT NOT NULL,
  policy_class TEXT NOT NULL
);
CREATE TABLE IF NOT EXISTS virtual_machines (
  id TEXT PRIMARY KEY,
  profile_name TEXT NOT NULL REFERENCES machine_profiles(name),
  project_id TEXT REFERENCES projects(id),
  state TEXT NOT NULL,
  runtime_id TEXT,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE TABLE IF NOT EXISTS machine_networks (
  id TEXT PRIMARY KEY,
  machine_id TEXT REFERENCES virtual_machines(id),
  mode TEXT NOT NULL,
  metadata_json JSONB NOT NULL DEFAULT '{}'::jsonb
);
CREATE TABLE IF NOT EXISTS snapshots (
  id TEXT PRIMARY KEY,
  machine_id TEXT REFERENCES virtual_machines(id),
  parent_id TEXT,
  disk_ref TEXT NOT NULL,
  metadata_json JSONB NOT NULL DEFAULT '{}'::jsonb,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE TABLE IF NOT EXISTS tasks (
  id TEXT PRIMARY KEY,
  machine_id TEXT REFERENCES virtual_machines(id),
  goal TEXT NOT NULL,
  status TEXT NOT NULL,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE TABLE IF NOT EXISTS task_iterations (
  id TEXT PRIMARY KEY,
  task_id TEXT REFERENCES tasks(id),
  iteration_number INTEGER NOT NULL,
  result TEXT NOT NULL,
  logs TEXT NOT NULL,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE TABLE IF NOT EXISTS artifacts (
  id TEXT PRIMARY KEY,
  task_id TEXT,
  machine_id TEXT,
  object_key TEXT NOT NULL,
  content_type TEXT NOT NULL,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE TABLE IF NOT EXISTS services (
  id TEXT PRIMARY KEY,
  machine_id TEXT REFERENCES virtual_machines(id),
  name TEXT NOT NULL,
  image TEXT NOT NULL,
  status TEXT NOT NULL,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE TABLE IF NOT EXISTS audit_events (
  id TEXT PRIMARY KEY,
  event_type TEXT NOT NULL,
  subject_id TEXT,
  message TEXT NOT NULL,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE TABLE IF NOT EXISTS event_stream_offsets (
  consumer TEXT PRIMARY KEY,
  stream TEXT NOT NULL,
  last_offset BIGINT NOT NULL
);
CREATE TABLE IF NOT EXISTS policies (
  id TEXT PRIMARY KEY,
  profile_name TEXT,
  network_mode TEXT NOT NULL,
  secret_policy TEXT NOT NULL,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
