# 👤 Users Service

Microservicio completo para gestión de usuarios, autenticación, seguridad y preferencias del ecosistema Financial Resume.

## 🚀 Características

### ✅ **Funcionalidades Implementadas**
- **Autenticación JWT completa** - Access tokens + Refresh tokens
- **Registro y login seguro** - Con validación de contraseñas robusta
- **2FA (TOTP)** - Compatible con Google Authenticator + códigos de backup
- **Gestión de perfiles** - CRUD completo de usuarios
- **Preferencias personalizables** - Moneda, idioma, tema, timezone
- **Configuración de notificaciones** - Email, push, alertas
- **Seguridad avanzada** - Bloqueo de cuentas, rate limiting
- **Recuperación de contraseña** - Tokens JWT seguros
- **Verificación de email** - Tokens con expiración
- **Exportación de datos** - GDPR compliance
- **Eliminación de cuenta** - Cascada automática

### 🔐 **Seguridad**
- **Hashing bcrypt** para contraseñas
- **JWT firmados** con algoritmo HS256
- **Validación de fortaleza** de contraseñas
- **Rate limiting** por IP
- **Account lockout** tras intentos fallidos
- **2FA TOTP** con backup codes
- **Token expiration** configurable

## 🏗️ **Arquitectura**

Implementa **Clean Architecture** con separación clara de responsabilidades:

```
users-service/
├── cmd/api/                    # Entry point
│   └── main.go                 # Server setup & dependency injection
├── internal/
│   ├── domain/                 # Entidades y DTOs
│   │   └── user.go             # User, Preferences, NotificationSettings
│   ├── usecases/               # Lógica de negocio
│   │   └── user_service.go     # Casos de uso completos
│   ├── repository/             # Capa de datos
│   │   └── user_repository.go  # PostgreSQL implementation
│   ├── handlers/               # Capa de presentación
│   │   └── user_handler.go     # HTTP handlers
│   └── infrastructure/         # Servicios externos
│       ├── auth/               # JWT, Password, 2FA services
│       ├── config/             # Configuración
│       └── middleware/         # Auth middleware
├── pkg/database/               # Database connection & migrations
├── Dockerfile                  # Container setup
├── docker-compose.yml          # Local development
└── env.example                 # Environment variables
```

## 📡 **API Endpoints**

### **Públicos (Sin autenticación)**
```bash
POST   /api/v1/users/register                    # Registro de usuario
POST   /api/v1/users/login                       # Login con opcional 2FA
POST   /api/v1/users/refresh                     # Renovar tokens
GET    /api/v1/users/verify-email/:token         # Verificar email
POST   /api/v1/users/request-password-reset      # Solicitar reset password
POST   /api/v1/users/reset-password              # Reset password con token

# Auth endpoints (compatibilidad con API principal)
POST   /api/v1/auth/login                        # Login alternativo
POST   /api/v1/auth/register                     # Registro alternativo
POST   /api/v1/auth/refresh                      # Refresh alternativo
PUT    /api/v1/auth/change-password              # Cambiar contraseña

GET    /health                                   # Health check (sin prefijo)
```

### **Protegidos (Requieren JWT)**
```bash
# Perfil
GET    /api/v1/users/profile                     # Obtener perfil
PUT    /api/v1/users/profile                     # Actualizar perfil
POST   /api/v1/users/logout                      # Logout

# Preferencias
GET    /api/v1/users/preferences                 # Obtener preferencias
PUT    /api/v1/users/preferences                 # Actualizar preferencias

# Notificaciones
GET    /api/v1/users/notifications/settings      # Configuración notificaciones
PUT    /api/v1/users/notifications/settings      # Actualizar notificaciones

# Seguridad
PUT    /api/v1/users/security/change-password    # Cambiar contraseña

# 2FA
POST   /api/v1/users/security/2fa/setup          # Configurar 2FA
POST   /api/v1/users/security/2fa/enable         # Activar 2FA
POST   /api/v1/users/security/2fa/disable        # Desactivar 2FA
POST   /api/v1/users/security/2fa/verify         # Verificar código 2FA

# Gestión de datos
POST   /api/v1/users/export                      # Exportar datos usuario
DELETE /api/v1/users                             # Eliminar cuenta
```

## 🗄️ **Base de Datos**

### **PostgreSQL en puerto 5434**
- **users** - Datos principales del usuario
- **user_preferences** - Configuraciones personales
- **user_notification_settings** - Configuración de notificaciones
- **user_two_fa** - Configuración 2FA y backup codes
- **user_tokens** - Gestión de tokens JWT
- **login_attempts** - Registro de intentos de login

### **Migraciones Automáticas**
El servicio ejecuta automáticamente las migraciones al iniciar.

## 🚀 **Ejecución**

### **Desarrollo Local**

1. **Clonar y configurar**:
```bash
cd users-service
cp env.example .env
# Editar .env con tu configuración
```

2. **Con Docker Compose** (Recomendado):
```bash
docker-compose up -d
```

3. **Desarrollo nativo**:
```bash
# Iniciar PostgreSQL (puerto 5434)
docker run -d --name users_db -p 5434:5432 \
  -e POSTGRES_USER=postgres \
  -e POSTGRES_PASSWORD=postgres \
  -e POSTGRES_DB=users_db \
  postgres:15

# Ejecutar servicio
go run cmd/api/main.go
```

### **Testing**
```bash
# Ejecutar tests
go test ./...

# Test con coverage
go test -cover ./...

# Test de endpoints
curl http://localhost:8083/health
```

## 🔧 **Configuración**

### **Variables de Entorno**

```bash
# Servidor
PORT=8083
ENVIRONMENT=development

# Base de datos
DB_HOST=localhost
DB_PORT=5434
DB_USER=postgres
DB_PASSWORD=postgres
DB_NAME=users_db

# JWT
JWT_SECRET=your-secret-key
JWT_ACCESS_EXPIRY_HOURS=24
JWT_REFRESH_EXPIRY_DAYS=7

# Seguridad
PASSWORD_MIN_LENGTH=8
MAX_LOGIN_ATTEMPTS=5
LOCKOUT_DURATION_MINUTES=15
```

## 🧪 **Testing de API**

### **Registro de Usuario**
```bash
curl -X POST http://localhost:8083/users/register \
  -H "Content-Type: application/json" \
  -d '{
    "email": "test@example.com",
    "password": "SecurePass123!",
    "first_name": "Test",
    "last_name": "User",
    "phone": "+1234567890"
  }'
```

### **Login**
```bash
curl -X POST http://localhost:8083/users/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "test@example.com",
    "password": "SecurePass123!"
  }'
```

### **Acceso con JWT**
```bash
# Usar el access_token del login
curl -X GET http://localhost:8083/users/profile \
  -H "Authorization: Bearer YOUR_ACCESS_TOKEN"
```

## 🔗 **Integración con Financial Resume Engine**

El users-service está diseñado para integrarse con el engine principal:

```bash
# En financial-resume-engine, agregar proxy:
USERS_SERVICE_URL=http://localhost:8083
```

Los endpoints se accederán vía proxy: `/api/v1/users/*`

## 📊 **Estado del Proyecto**

### ✅ **Completado (100%)**
- [x] Clean Architecture implementada
- [x] JWT Authentication completo
- [x] Password security con bcrypt
- [x] 2FA con Google Authenticator
- [x] Database PostgreSQL con migraciones
- [x] CRUD completo de usuarios
- [x] Gestión de preferencias y notificaciones
- [x] Security features (lockout, rate limiting)
- [x] Email verification y password reset
- [x] Docker y docker-compose
- [x] Health checks
- [x] Graceful shutdown
- [x] Middleware de autenticación
- [x] Error handling robusto
- [x] Documentación completa

### 🎯 **Listo para Producción**
El users-service está **100% funcional** y listo para integración con el ecosistema Financial Resume.

**Total**: ~3,400 líneas de código implementadas ✅
# Deployment trigger Wed Aug 20 14:23:57 -03 2025
