-- =====================================================
-- USERS SERVICE - DATOS DE PRUEBA
-- Script con usuarios mock para testing y desarrollo
-- Basado en usuarios existentes de la API principal
-- =====================================================
-- 
-- 🔑 CREDENCIALES DE USUARIOS PARA TESTING:
-- 📧 nivel1@test.com     🔑 password123    🎮 Usuario Principiante
-- 📧 nivel3@test.com     🔑 password123    🎮 Usuario Intermedio  
-- 📧 nivel5@test.com     🔑 password123    🎮 Usuario Avanzado
-- 📧 nivel10@test.com    🔑 password123    🎮 Usuario Experto
-- 📧 pablo@niloft.com    🔑 password123    🎮 Usuario Principal
-- 📧 admin@test.com      🔑 password123    🎮 Usuario Admin
-- =====================================================

-- Limpiar datos existentes si los hay
DELETE FROM login_attempts;
DELETE FROM user_tokens;
DELETE FROM user_two_fa;
DELETE FROM user_notification_settings;
DELETE FROM user_preferences;
DELETE FROM users;

-- Reiniciar secuencia
ALTER SEQUENCE users_id_seq RESTART WITH 1;

-- =====================================================
-- USUARIOS DE PRUEBA CON CREDENCIALES REALES
-- =====================================================
-- Contraseña para todos: "password123" 
-- Hashes generados con bcrypt cost 10

-- USUARIO NIVEL 1 - PRINCIPIANTE
INSERT INTO users (
    id, email, password_hash, first_name, last_name, phone, 
    is_active, is_verified, created_at, updated_at
) VALUES (
    1, 
    'nivel1@test.com', 
    '$2a$10$j7nQDMUmfSfPjntQDxO/YeJCsaP6d7P3DTX1aN.zqeEOgFgH0XSu.', -- password123
    'Usuario', 
    'Principiante',
    '+1234567001',
    true, 
    true, 
    CURRENT_TIMESTAMP - INTERVAL '30 days', 
    CURRENT_TIMESTAMP - INTERVAL '1 day'
);

-- USUARIO NIVEL 3 - INTERMEDIO (SMART SAVER)
INSERT INTO users (
    id, email, password_hash, first_name, last_name, phone,
    is_active, is_verified, created_at, updated_at
) VALUES (
    2, 
    'nivel3@test.com', 
    '$2a$10$j7nQDMUmfSfPjntQDxO/YeJCsaP6d7P3DTX1aN.zqeEOgFgH0XSu.', -- password123
    'Usuario', 
    'Intermedio',
    '+1234567003',
    true, 
    true, 
    CURRENT_TIMESTAMP - INTERVAL '25 days', 
    CURRENT_TIMESTAMP - INTERVAL '2 hours'
);

-- USUARIO NIVEL 5 - AVANZADO (FINANCIAL PLANNER)
INSERT INTO users (
    id, email, password_hash, first_name, last_name, phone,
    is_active, is_verified, created_at, updated_at
) VALUES (
    3, 
    'nivel5@test.com', 
    '$2a$10$j7nQDMUmfSfPjntQDxO/YeJCsaP6d7P3DTX1aN.zqeEOgFgH0XSu.', -- password123
    'Usuario', 
    'Avanzado',
    '+1234567005',
    true, 
    true, 
    CURRENT_TIMESTAMP - INTERVAL '20 days', 
    CURRENT_TIMESTAMP - INTERVAL '1 hour'
);

-- USUARIO NIVEL 10 - EXPERTO (FINANCIAL MAGNATE)
INSERT INTO users (
    id, email, password_hash, first_name, last_name, phone,
    is_active, is_verified, last_login, created_at, updated_at
) VALUES (
    4, 
    'nivel10@test.com', 
    '$2a$10$j7nQDMUmfSfPjntQDxO/YeJCsaP6d7P3DTX1aN.zqeEOgFgH0XSu.', -- password123
    'Usuario', 
    'Experto',
    '+1234567010',
    true, 
    true, 
    CURRENT_TIMESTAMP - INTERVAL '30 minutes',
    CURRENT_TIMESTAMP - INTERVAL '15 days', 
    CURRENT_TIMESTAMP - INTERVAL '30 minutes'
);

-- USUARIO PRINCIPAL - PABLO MELEGATTI (FOUNDER)
INSERT INTO users (
    id, email, password_hash, first_name, last_name, phone,
    is_active, is_verified, last_login, created_at, updated_at
) VALUES (
    5, 
    'pablo@niloft.com', 
    '$2a$10$j7nQDMUmfSfPjntQDxO/YeJCsaP6d7P3DTX1aN.zqeEOgFgH0XSu.', -- password123
    'Pablo', 
    'Melegatti',
    '+541234567890',
    true, 
    true, 
    CURRENT_TIMESTAMP - INTERVAL '15 minutes',
    CURRENT_TIMESTAMP - INTERVAL '60 days', 
    CURRENT_TIMESTAMP - INTERVAL '15 minutes'
);

-- USUARIO ADMIN - PARA TESTING AVANZADO
INSERT INTO users (
    id, email, password_hash, first_name, last_name, phone,
    is_active, is_verified, last_login, created_at, updated_at
) VALUES (
    6, 
    'admin@test.com', 
    '$2a$10$j7nQDMUmfSfPjntQDxO/YeJCsaP6d7P3DTX1aN.zqeEOgFgH0XSu.', -- password123
    'Admin', 
    'Sistema',
    '+1234567999',
    true, 
    true, 
    CURRENT_TIMESTAMP - INTERVAL '5 minutes',
    CURRENT_TIMESTAMP - INTERVAL '90 days', 
    CURRENT_TIMESTAMP - INTERVAL '5 minutes'
);

-- =====================================================
-- PREFERENCIAS PARA USUARIOS DE PRUEBA
-- =====================================================

-- Preferencias Usuario Nivel 1 - Configuración básica
INSERT INTO user_preferences (
    user_id, currency, language, theme, date_format, timezone
) VALUES (
    1, 'USD', 'en', 'light', 'MM/DD/YYYY', 'America/New_York'
);

-- Preferencias Usuario Nivel 3 - Personalizado
INSERT INTO user_preferences (
    user_id, currency, language, theme, date_format, timezone
) VALUES (
    2, 'EUR', 'es', 'light', 'DD/MM/YYYY', 'Europe/Madrid'
);

-- Preferencias Usuario Nivel 5 - Avanzado
INSERT INTO user_preferences (
    user_id, currency, language, theme, date_format, timezone
) VALUES (
    3, 'GBP', 'en', 'dark', 'YYYY-MM-DD', 'Europe/London'
);

-- Preferencias Usuario Nivel 10 - Experto
INSERT INTO user_preferences (
    user_id, currency, language, theme, date_format, timezone
) VALUES (
    4, 'JPY', 'en', 'dark', 'YYYY/MM/DD', 'Asia/Tokyo'
);

-- Preferencias Pablo - Argentina
INSERT INTO user_preferences (
    user_id, currency, language, theme, date_format, timezone
) VALUES (
    5, 'ARS', 'es', 'dark', 'DD/MM/YYYY', 'America/Argentina/Buenos_Aires'
);

-- Preferencias Admin - Sistema
INSERT INTO user_preferences (
    user_id, currency, language, theme, date_format, timezone
) VALUES (
    6, 'USD', 'en', 'light', 'YYYY-MM-DD', 'UTC'
);

-- =====================================================
-- CONFIGURACIÓN DE NOTIFICACIONES
-- =====================================================

-- Usuario Nivel 1 - Todas las notificaciones activas (principiante)
INSERT INTO user_notification_settings (
    user_id, email_notifications, push_notifications, weekly_reports, 
    expense_alerts, budget_alerts, achievement_notifications
) VALUES (
    1, true, true, true, true, true, true
);

-- Usuario Nivel 3 - Configuración intermedia
INSERT INTO user_notification_settings (
    user_id, email_notifications, push_notifications, weekly_reports, 
    expense_alerts, budget_alerts, achievement_notifications
) VALUES (
    2, true, true, true, true, false, true
);

-- Usuario Nivel 5 - Configuración selectiva
INSERT INTO user_notification_settings (
    user_id, email_notifications, push_notifications, weekly_reports, 
    expense_alerts, budget_alerts, achievement_notifications
) VALUES (
    3, true, false, true, false, true, false
);

-- Usuario Nivel 10 - Configuración mínima (experto)
INSERT INTO user_notification_settings (
    user_id, email_notifications, push_notifications, weekly_reports, 
    expense_alerts, budget_alerts, achievement_notifications
) VALUES (
    4, false, false, true, false, false, false
);

-- Pablo - Configuración personalizada
INSERT INTO user_notification_settings (
    user_id, email_notifications, push_notifications, weekly_reports, 
    expense_alerts, budget_alerts, achievement_notifications
) VALUES (
    5, true, true, false, true, true, true
);

-- Admin - Solo notificaciones esenciales
INSERT INTO user_notification_settings (
    user_id, email_notifications, push_notifications, weekly_reports, 
    expense_alerts, budget_alerts, achievement_notifications
) VALUES (
    6, true, false, false, false, false, false
);

-- =====================================================
-- CONFIGURACIÓN 2FA PARA ALGUNOS USUARIOS
-- =====================================================

-- Usuario Nivel 10 - 2FA Activado (usuario experto)
INSERT INTO user_two_fa (
    user_id, secret, enabled, backup_codes
) VALUES (
    4, 
    'JBSWY3DPEHPK3PXP', -- Secret TOTP de ejemplo
    true,
    ARRAY['ABCD-1234', 'EFGH-5678', 'IJKL-9012', 'MNOP-3456', 'QRST-7890', 'UVWX-2345', 'YZAB-6789', 'CDEF-0123']
);

-- Pablo - 2FA Configurado pero no activado aún
INSERT INTO user_two_fa (
    user_id, secret, enabled, backup_codes
) VALUES (
    5, 
    'KRUGKIDROVUWG2ZAMJZG653OEBTG66BANJ2W24DTEBXXMZLSEB2GQZJANRQXU6JAMRXWOLQ', 
    false,
    ARRAY['WXYZ-4567', 'ABCD-8901', 'EFGH-2345', 'IJKL-6789', 'MNOP-0123', 'QRST-4567', 'UVWX-8901', 'YZAB-2345']
);

-- Admin - 2FA Activado (seguridad máxima)
INSERT INTO user_two_fa (
    user_id, secret, enabled, backup_codes
) VALUES (
    6, 
    'MFRGG2LTMFZXI4TJNZ2HS4LTORSW24A', 
    true,
    ARRAY['ADMIN-001', 'ADMIN-002', 'ADMIN-003', 'ADMIN-004', 'ADMIN-005', 'ADMIN-006', 'ADMIN-007', 'ADMIN-008']
);

-- =====================================================
-- HISTORIAL DE LOGIN ATTEMPTS PARA ANÁLISIS
-- =====================================================

-- Login exitosos recientes
INSERT INTO login_attempts (email, ip_address, user_agent, success, attempted_at) VALUES 
('pablo@niloft.com', '192.168.1.100', 'Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7)', true, CURRENT_TIMESTAMP - INTERVAL '15 minutes'),
('nivel10@test.com', '192.168.1.101', 'Mozilla/5.0 (Windows NT 10.0; Win64; x64)', true, CURRENT_TIMESTAMP - INTERVAL '30 minutes'),
('nivel5@test.com', '192.168.1.102', 'Mozilla/5.0 (X11; Linux x86_64)', true, CURRENT_TIMESTAMP - INTERVAL '1 hour'),
('admin@test.com', '192.168.1.103', 'Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7)', true, CURRENT_TIMESTAMP - INTERVAL '5 minutes');

-- Algunos intentos fallidos para demo de seguridad
INSERT INTO login_attempts (email, ip_address, user_agent, success, failure_reason, attempted_at) VALUES 
('nivel1@test.com', '192.168.1.200', 'Mozilla/5.0 (Windows NT 10.0; Win64; x64)', false, 'invalid_password', CURRENT_TIMESTAMP - INTERVAL '2 hours'),
('test@hacker.com', '192.168.1.66', 'curl/7.68.0', false, 'user_not_found', CURRENT_TIMESTAMP - INTERVAL '3 hours'),
('pablo@niloft.com', '192.168.1.200', 'Mozilla/5.0 (Windows NT 10.0; Win64; x64)', false, 'invalid_password', CURRENT_TIMESTAMP - INTERVAL '4 hours');

-- =====================================================
-- ACTUALIZAR SECUENCIA DE IDs
-- =====================================================

-- Asegurar que la secuencia esté sincronizada
SELECT setval('users_id_seq', (SELECT MAX(id) FROM users));

-- =====================================================
-- VERIFICACIÓN DE DATOS INSERTADOS
-- =====================================================

-- Mostrar resumen de usuarios creados
SELECT 
    id,
    email,
    first_name,
    last_name,
    is_active,
    is_verified,
    CASE WHEN last_login IS NOT NULL THEN 'Yes' ELSE 'No' END as has_logged_in,
    created_at::date as created_date
FROM users 
ORDER BY id;

-- Mostrar configuración 2FA
SELECT 
    u.email,
    CASE WHEN t.enabled THEN 'Enabled' ELSE 'Disabled' END as twofa_status,
    array_length(t.backup_codes, 1) as backup_codes_count
FROM users u
LEFT JOIN user_two_fa t ON u.id = t.user_id
ORDER BY u.id;

-- Mostrar preferencias por usuario
SELECT 
    u.email,
    p.currency,
    p.language,
    p.theme,
    p.timezone
FROM users u
JOIN user_preferences p ON u.id = p.user_id
ORDER BY u.id;

-- Estadísticas de login attempts
SELECT 
    success,
    COUNT(*) as attempts_count,
    COUNT(DISTINCT email) as unique_emails
FROM login_attempts 
GROUP BY success;

NOTIFY users_test_data_ready, 'Users Service test data loaded successfully!'; 