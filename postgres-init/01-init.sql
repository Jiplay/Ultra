-- PostgreSQL initialization script
-- This script will be executed when the container starts

-- Ensure the database exists (should be created by POSTGRES_DB env var)
-- Creating extensions that might be useful
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE EXTENSION IF NOT EXISTS "pgcrypto";

-- Grant necessary permissions
GRANT ALL PRIVILEGES ON DATABASE ultra_db TO ultra_user;