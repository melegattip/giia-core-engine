-- =====================================================
-- USERS SERVICE - ESTRUCTURA DE BASE DE DATOS
-- Script que crea tablas, índices y relaciones para el microservicio de usuarios
-- Puerto: 5434 | Base de datos: users_db
-- =====================================================

-- Eliminar tablas existentes si existen (en orden correcto por dependencias)
DROP TABLE IF EXISTS login_attempts CASCADE;
DROP TABLE IF EXISTS user_tokens CASCADE;
DROP TABLE IF EXISTS user_two_fa CASCADE;
DROP TABLE IF EXISTS user_notification_settings CASCADE;
DROP TABLE IF EXISTS user_preferences CASCADE;
DROP TABLE IF EXISTS users CASCADE;

-- =====================================================
-- TABLA PRINCIPAL DE USUARIOS
-- =====================================================

CREATE TABLE IF NOT EXISTS users (
    id SERIAL PRIMARY KEY,
    email VARCHAR(255) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    first_name VARCHAR(100) NOT NULL,
    last_name VARCHAR(100) NOT NULL,
    phone VARCHAR(20),
    avatar VARCHAR(500),
    is_active BOOLEAN DEFAULT true,
    is_verified BOOLEAN DEFAULT false,
    email_verification_token VARCHAR(255),
    email_verification_expires TIMESTAMP,
    password_reset_token VARCHAR(255),
    password_reset_expires TIMESTAMP,
    last_login TIMESTAMP,
    failed_login_attempts INTEGER DEFAULT 0,
    locked_until TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- =====================================================
-- TABLA DE PREFERENCIAS DE USUARIO
-- =====================================================

CREATE TABLE IF NOT EXISTS user_preferences (
    user_id INTEGER PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE,
    currency VARCHAR(3) DEFAULT 'USD',
    language VARCHAR(5) DEFAULT 'en',
    theme VARCHAR(10) DEFAULT 'light',
    date_format VARCHAR(20) DEFAULT 'YYYY-MM-DD',
    timezone VARCHAR(50) DEFAULT 'UTC',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- =====================================================
-- TABLA DE CONFIGURACIÓN DE NOTIFICACIONES
-- =====================================================

CREATE TABLE IF NOT EXISTS user_notification_settings (
    user_id INTEGER PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE,
    email_notifications BOOLEAN DEFAULT true,
    push_notifications BOOLEAN DEFAULT true,
    weekly_reports BOOLEAN DEFAULT true,
    expense_alerts BOOLEAN DEFAULT true,
    budget_alerts BOOLEAN DEFAULT true,
    achievement_notifications BOOLEAN DEFAULT true,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- =====================================================
-- TABLA DE CONFIGURACIÓN 2FA
-- =====================================================

CREATE TABLE IF NOT EXISTS user_two_fa (
    user_id INTEGER PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE,
    secret VARCHAR(255) NOT NULL,
    enabled BOOLEAN DEFAULT false,
    backup_codes TEXT[], -- Array of backup codes
    last_used_code VARCHAR(10),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- =====================================================
-- TABLA DE GESTIÓN DE TOKENS JWT
-- =====================================================

CREATE TABLE IF NOT EXISTS user_tokens (
    id SERIAL PRIMARY KEY,
    user_id INTEGER REFERENCES users(id) ON DELETE CASCADE,
    token_hash VARCHAR(255) NOT NULL,
    token_type VARCHAR(20) NOT NULL, -- 'refresh', 'access', 'email_verification', 'password_reset'
    expires_at TIMESTAMP NOT NULL,
    is_revoked BOOLEAN DEFAULT false,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- =====================================================
-- TABLA DE INTENTOS DE LOGIN (SEGURIDAD)
-- =====================================================

CREATE TABLE IF NOT EXISTS login_attempts (
    id SERIAL PRIMARY KEY,
    email VARCHAR(255) NOT NULL,
    ip_address INET,
    user_agent TEXT,
    success BOOLEAN NOT NULL,
    failure_reason VARCHAR(100),
    attempted_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- =====================================================
-- ÍNDICES PARA OPTIMIZACIÓN DE CONSULTAS
-- =====================================================

-- Índices en tabla users
CREATE INDEX IF NOT EXISTS idx_users_email ON users(email);
CREATE INDEX IF NOT EXISTS idx_users_email_verification_token ON users(email_verification_token);
CREATE INDEX IF NOT EXISTS idx_users_password_reset_token ON users(password_reset_token);
CREATE INDEX IF NOT EXISTS idx_users_is_active ON users(is_active);
CREATE INDEX IF NOT EXISTS idx_users_is_verified ON users(is_verified);
CREATE INDEX IF NOT EXISTS idx_users_created_at ON users(created_at);

-- Índices en tabla user_tokens
CREATE INDEX IF NOT EXISTS idx_user_tokens_user_id ON user_tokens(user_id);
CREATE INDEX IF NOT EXISTS idx_user_tokens_token_hash ON user_tokens(token_hash);
CREATE INDEX IF NOT EXISTS idx_user_tokens_token_type ON user_tokens(token_type);
CREATE INDEX IF NOT EXISTS idx_user_tokens_expires_at ON user_tokens(expires_at);
CREATE INDEX IF NOT EXISTS idx_user_tokens_is_revoked ON user_tokens(is_revoked);

-- Índices en tabla login_attempts
CREATE INDEX IF NOT EXISTS idx_login_attempts_email ON login_attempts(email);
CREATE INDEX IF NOT EXISTS idx_login_attempts_attempted_at ON login_attempts(attempted_at);
CREATE INDEX IF NOT EXISTS idx_login_attempts_success ON login_attempts(success);
CREATE INDEX IF NOT EXISTS idx_login_attempts_ip_address ON login_attempts(ip_address);

-- =====================================================
-- FUNCIONES DE TRIGGER PARA UPDATED_AT
-- =====================================================

-- Función para actualizar timestamp automáticamente
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ language 'plpgsql';

-- Triggers para actualizar updated_at automáticamente
CREATE TRIGGER update_users_updated_at 
    BEFORE UPDATE ON users 
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_user_preferences_updated_at 
    BEFORE UPDATE ON user_preferences 
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_user_notification_settings_updated_at 
    BEFORE UPDATE ON user_notification_settings 
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_user_two_fa_updated_at 
    BEFORE UPDATE ON user_two_fa 
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- =====================================================
-- COMENTARIOS DOCUMENTALES
-- =====================================================

COMMENT ON TABLE users IS 'Tabla principal de usuarios del sistema';
COMMENT ON TABLE user_preferences IS 'Preferencias personalizables por usuario (idioma, moneda, tema, etc.)';
COMMENT ON TABLE user_notification_settings IS 'Configuración de notificaciones por usuario';
COMMENT ON TABLE user_two_fa IS 'Configuración de autenticación de dos factores';
COMMENT ON TABLE user_tokens IS 'Gestión de tokens JWT (access, refresh, email verification, etc.)';
COMMENT ON TABLE login_attempts IS 'Registro de intentos de login para análisis de seguridad';

-- =====================================================
-- VERIFICACIÓN DE INSTALACIÓN
-- =====================================================

-- Mostrar resumen de tablas creadas
SELECT 
    schemaname, 
    tablename, 
    tableowner 
FROM pg_tables 
WHERE schemaname = 'public' 
    AND tablename LIKE 'user%' 
    OR tablename = 'login_attempts'
ORDER BY tablename;

-- Mostrar índices creados
SELECT 
    indexname, 
    tablename 
FROM pg_indexes 
WHERE schemaname = 'public' 
    AND tablename LIKE 'user%' 
    OR tablename = 'login_attempts'
ORDER BY tablename, indexname;

NOTIFY users_db_ready, 'Users Service database initialization completed successfully!'; 