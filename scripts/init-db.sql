-- =============================================================================
-- GIIA Database Initialization Script
-- =============================================================================
-- This script creates separate schemas within the giia_dev database for each
-- microservice, following the multi-schema approach for logical data isolation.
-- =============================================================================

-- Create schemas for each microservice
CREATE SCHEMA IF NOT EXISTS auth;
CREATE SCHEMA IF NOT EXISTS catalog;
CREATE SCHEMA IF NOT EXISTS ddmrp;
CREATE SCHEMA IF NOT EXISTS execution;
CREATE SCHEMA IF NOT EXISTS analytics;
CREATE SCHEMA IF NOT EXISTS ai_agent;

-- Enable required PostgreSQL extensions
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE EXTENSION IF NOT EXISTS "pgcrypto";

-- Enable vector extension for AI service (if available)
-- Note: This may require pg_vector to be installed
-- Uncomment if using vector embeddings:
-- CREATE EXTENSION IF NOT EXISTS "vector";

-- Set default search_path to include all schemas (optional)
-- Each service should specify its schema in connection string
ALTER DATABASE giia_dev SET search_path TO auth, catalog, ddmrp, execution, analytics, ai_agent, public;

-- Grant schema usage to the giia user
GRANT USAGE ON SCHEMA auth TO giia;
GRANT USAGE ON SCHEMA catalog TO giia;
GRANT USAGE ON SCHEMA ddmrp TO giia;
GRANT USAGE ON SCHEMA execution TO giia;
GRANT USAGE ON SCHEMA analytics TO giia;
GRANT USAGE ON SCHEMA ai_agent TO giia;

-- Grant all privileges on all tables in schemas to giia user
GRANT ALL PRIVILEGES ON ALL TABLES IN SCHEMA auth TO giia;
GRANT ALL PRIVILEGES ON ALL TABLES IN SCHEMA catalog TO giia;
GRANT ALL PRIVILEGES ON ALL TABLES IN SCHEMA ddmrp TO giia;
GRANT ALL PRIVILEGES ON ALL TABLES IN SCHEMA execution TO giia;
GRANT ALL PRIVILEGES ON ALL TABLES IN SCHEMA analytics TO giia;
GRANT ALL PRIVILEGES ON ALL TABLES IN SCHEMA ai_agent TO giia;

-- Grant all privileges on all sequences in schemas to giia user
GRANT ALL PRIVILEGES ON ALL SEQUENCES IN SCHEMA auth TO giia;
GRANT ALL PRIVILEGES ON ALL SEQUENCES IN SCHEMA catalog TO giia;
GRANT ALL PRIVILEGES ON ALL SEQUENCES IN SCHEMA ddmrp TO giia;
GRANT ALL PRIVILEGES ON ALL SEQUENCES IN SCHEMA execution TO giia;
GRANT ALL PRIVILEGES ON ALL SEQUENCES IN SCHEMA analytics TO giia;
GRANT ALL PRIVILEGES ON ALL SEQUENCES IN SCHEMA ai_agent TO giia;

-- Set default privileges for future tables
ALTER DEFAULT PRIVILEGES IN SCHEMA auth GRANT ALL ON TABLES TO giia;
ALTER DEFAULT PRIVILEGES IN SCHEMA catalog GRANT ALL ON TABLES TO giia;
ALTER DEFAULT PRIVILEGES IN SCHEMA ddmrp GRANT ALL ON TABLES TO giia;
ALTER DEFAULT PRIVILEGES IN SCHEMA execution GRANT ALL ON TABLES TO giia;
ALTER DEFAULT PRIVILEGES IN SCHEMA analytics GRANT ALL ON TABLES TO giia;
ALTER DEFAULT PRIVILEGES IN SCHEMA ai_agent GRANT ALL ON TABLES TO giia;

-- Set default privileges for future sequences
ALTER DEFAULT PRIVILEGES IN SCHEMA auth GRANT ALL ON SEQUENCES TO giia;
ALTER DEFAULT PRIVILEGES IN SCHEMA catalog GRANT ALL ON SEQUENCES TO giia;
ALTER DEFAULT PRIVILEGES IN SCHEMA ddmrp GRANT ALL ON SEQUENCES TO giia;
ALTER DEFAULT PRIVILEGES IN SCHEMA execution GRANT ALL ON SEQUENCES TO giia;
ALTER DEFAULT PRIVILEGES IN SCHEMA analytics GRANT ALL ON SEQUENCES TO giia;
ALTER DEFAULT PRIVILEGES IN SCHEMA ai_agent GRANT ALL ON SEQUENCES TO giia;

-- Log initialization complete
DO $$
BEGIN
    RAISE NOTICE 'GIIA database schemas initialized successfully!';
    RAISE NOTICE 'Schemas created: auth, catalog, ddmrp, execution, analytics, ai_agent';
END $$;