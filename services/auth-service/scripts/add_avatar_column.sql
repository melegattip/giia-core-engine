-- =====================================================
-- MIGRACIÓN: AGREGAR COLUMNA AVATAR A LA TABLA USERS
-- Script para agregar soporte de avatares al microservicio de usuarios
-- =====================================================

-- Agregar columna avatar a la tabla users
ALTER TABLE users ADD COLUMN IF NOT EXISTS avatar VARCHAR(500);

-- Agregar comentario a la columna
COMMENT ON COLUMN users.avatar IS 'URL o ruta del archivo de avatar del usuario';

-- Crear directorio de uploads si no existe (esto se hace a nivel de aplicación)
-- La aplicación creará el directorio 'uploads' automáticamente

-- Verificar que la columna se agregó correctamente
SELECT 
    column_name, 
    data_type, 
    is_nullable, 
    column_default
FROM information_schema.columns 
WHERE table_name = 'users' 
    AND column_name = 'avatar';

-- Mostrar la estructura actualizada de la tabla users
SELECT 
    column_name, 
    data_type, 
    is_nullable
FROM information_schema.columns 
WHERE table_name = 'users' 
ORDER BY ordinal_position;

NOTIFY avatar_migration_complete, 'Avatar column migration completed successfully!'; 