-- GIIA Database Initialization Script
-- This script creates separate databases for each service

-- Create databases for each service
CREATE DATABASE IF NOT EXISTS giia_auth;
CREATE DATABASE IF NOT EXISTS giia_catalog;
CREATE DATABASE IF NOT EXISTS giia_ddmrp;
CREATE DATABASE IF NOT EXISTS giia_execution;
CREATE DATABASE IF NOT EXISTS giia_analytics;
CREATE DATABASE IF NOT EXISTS giia_ai;

-- Create users (optional - for service isolation)
DO
$$
BEGIN
    IF NOT EXISTS (SELECT FROM pg_catalog.pg_roles WHERE rolname = 'auth_service') THEN
        CREATE USER auth_service WITH PASSWORD 'auth_service_password';
    END IF;
    IF NOT EXISTS (SELECT FROM pg_catalog.pg_roles WHERE rolname = 'catalog_service') THEN
        CREATE USER catalog_service WITH PASSWORD 'catalog_service_password';
    END IF;
    IF NOT EXISTS (SELECT FROM pg_catalog.pg_roles WHERE rolname = 'ddmrp_service') THEN
        CREATE USER ddmrp_service WITH PASSWORD 'ddmrp_service_password';
    END IF;
    IF NOT EXISTS (SELECT FROM pg_catalog.pg_roles WHERE rolname = 'execution_service') THEN
        CREATE USER execution_service WITH PASSWORD 'execution_service_password';
    END IF;
    IF NOT EXISTS (SELECT FROM pg_catalog.pg_roles WHERE rolname = 'analytics_service') THEN
        CREATE USER analytics_service WITH PASSWORD 'analytics_service_password';
    END IF;
    IF NOT EXISTS (SELECT FROM pg_catalog.pg_roles WHERE rolname = 'ai_service') THEN
        CREATE USER ai_service WITH PASSWORD 'ai_service_password';
    END IF;
END
$$;

-- Grant privileges
GRANT ALL PRIVILEGES ON DATABASE giia_auth TO auth_service;
GRANT ALL PRIVILEGES ON DATABASE giia_catalog TO catalog_service;
GRANT ALL PRIVILEGES ON DATABASE giia_ddmrp TO ddmrp_service;
GRANT ALL PRIVILEGES ON DATABASE giia_execution TO execution_service;
GRANT ALL PRIVILEGES ON DATABASE giia_analytics TO analytics_service;
GRANT ALL PRIVILEGES ON DATABASE giia_ai TO ai_service;

-- Enable UUID extension
\c giia_auth
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE EXTENSION IF NOT EXISTS "pgcrypto";

\c giia_catalog
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

\c giia_ddmrp
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

\c giia_execution
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

\c giia_analytics
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

\c giia_ai
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE EXTENSION IF NOT EXISTS "vector";  -- For embeddings (if using RAG)

-- Switch back to default database
\c giia_dev;

-- Log initialization complete
SELECT 'GIIA databases initialized successfully!' AS status;
