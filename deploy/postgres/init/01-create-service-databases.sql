-- Keep one PostgreSQL instance for local development while preserving
-- schema ownership boundaries between device_platform and sub2api.
CREATE DATABASE sprout_device_platform;
CREATE DATABASE sprout_sub2api;
