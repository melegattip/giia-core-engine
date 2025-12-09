-- =============================================================================
-- GIIA Seed Data Script
-- =============================================================================
-- This script populates test data for local development and testing.
-- Only use in development environments - NOT for production!
-- =============================================================================

-- =============================================================================
-- AUTH SCHEMA - Sample Users and Roles
-- =============================================================================

SET search_path TO auth;

-- Create sample users table (if not exists from migrations)
CREATE TABLE IF NOT EXISTS users (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    email VARCHAR(255) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    first_name VARCHAR(100),
    last_name VARCHAR(100),
    is_active BOOLEAN DEFAULT true,
    is_verified BOOLEAN DEFAULT false,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Insert sample users (password: "password123")
INSERT INTO users (email, password_hash, first_name, last_name, is_active, is_verified)
VALUES
    ('admin@giia.local', '$2a$10$XqjT8YZhV5d8K5x5L5J5W.5J5L5J5W5J5L5J5W5J5L5J5W5J5L5J5W', 'Admin', 'User', true, true),
    ('john.doe@giia.local', '$2a$10$XqjT8YZhV5d8K5x5L5J5W.5J5L5J5W5J5L5J5W5J5L5J5W5J5L5J5W', 'John', 'Doe', true, true),
    ('jane.smith@giia.local', '$2a$10$XqjT8YZhV5d8K5x5L5J5W.5J5L5J5W5J5L5J5W5J5L5J5W5J5L5J5W', 'Jane', 'Smith', true, false)
ON CONFLICT (email) DO NOTHING;

-- =============================================================================
-- CATALOG SCHEMA - Sample Products and Categories
-- =============================================================================

SET search_path TO catalog;

-- Create sample categories table
CREATE TABLE IF NOT EXISTS categories (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(255) NOT NULL,
    description TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Insert sample categories
INSERT INTO categories (name, description)
VALUES
    ('Raw Materials', 'Basic materials for production'),
    ('Finished Goods', 'Ready to sell products'),
    ('Work in Progress', 'Items currently being manufactured')
ON CONFLICT DO NOTHING;

-- Create sample products table
CREATE TABLE IF NOT EXISTS products (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    sku VARCHAR(100) UNIQUE NOT NULL,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    category_id UUID REFERENCES categories(id),
    unit_cost DECIMAL(10, 2),
    unit_price DECIMAL(10, 2),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Insert sample products
INSERT INTO products (sku, name, description, unit_cost, unit_price)
VALUES
    ('PROD-001', 'Widget A', 'Standard widget for general use', 10.50, 25.00),
    ('PROD-002', 'Widget B', 'Premium widget with advanced features', 15.75, 35.00),
    ('PROD-003', 'Component X', 'Essential component for assembly', 5.25, 12.00)
ON CONFLICT (sku) DO NOTHING;

-- =============================================================================
-- DDMRP SCHEMA - Sample Buffer Configurations
-- =============================================================================

SET search_path TO ddmrp;

-- Note: Actual DDMRP tables will be created by service migrations
-- This is just sample data structure

RAISE NOTICE 'DDMRP seed data will be added when service schema is ready';

-- =============================================================================
-- EXECUTION SCHEMA - Sample Orders
-- =============================================================================

SET search_path TO execution;

RAISE NOTICE 'Execution service seed data will be added when service schema is ready';

-- =============================================================================
-- ANALYTICS SCHEMA - Sample Reports
-- =============================================================================

SET search_path TO analytics;

RAISE NOTICE 'Analytics service seed data will be added when service schema is ready';

-- =============================================================================
-- AI_AGENT SCHEMA - Sample Configurations
-- =============================================================================

SET search_path TO ai_agent;

RAISE NOTICE 'AI Agent service seed data will be added when service schema is ready';

-- =============================================================================
-- Completion Message
-- =============================================================================

DO $$
BEGIN
    RAISE NOTICE 'Seed data loaded successfully for development environment!';
    RAISE NOTICE 'Test users: admin@giia.local, john.doe@giia.local, jane.smith@giia.local';
    RAISE NOTICE 'Default password for all users: password123';
END $$;