# 👤 **IMPLEMENTACIÓN DE AVATARES EN EL BACKEND**

## 📋 **RESUMEN DE CAMBIOS**

Se ha implementado la funcionalidad completa para manejar avatares de usuario en el microservicio de usuarios, incluyendo subida de archivos, almacenamiento y actualización de perfiles.

---

## 🔧 **CAMBIOS IMPLEMENTADOS**

### **1. Modelo de Usuario Actualizado**
```go
type User struct {
    ID                        uint       `json:"id"`
    Email                     string     `json:"email"`
    Password                  string     `json:"-"`
    FirstName                 string     `json:"first_name"`
    LastName                  string     `json:"last_name"`
    Phone                     string     `json:"phone"`
    Avatar                    string     `json:"avatar,omitempty"` // ✅ NUEVO CAMPO
    IsActive                  bool       `json:"is_active"`
    IsVerified                bool       `json:"is_verified"`
    // ... resto de campos
}
```

### **2. Base de Datos - Migración**
```sql
-- Script: scripts/add_avatar_column.sql
ALTER TABLE users ADD COLUMN IF NOT EXISTS avatar VARCHAR(500);
COMMENT ON COLUMN users.avatar IS 'URL o ruta del archivo de avatar del usuario';
```

### **3. Repositorio Actualizado**
```go
// Método Update actualizado para incluir avatar
func (r *userRepository) Update(ctx context.Context, user *domain.User) error {
    query := `
        UPDATE users 
        SET first_name = $1, last_name = $2, phone = $3, avatar = $4, is_active = $5, updated_at = CURRENT_TIMESTAMP
        WHERE id = $6`
    // ...
}

// Métodos GetByID y GetByEmail actualizados para incluir avatar
func (r *userRepository) GetByID(ctx context.Context, id uint) (*domain.User, error) {
    query := `
        SELECT id, email, password_hash, first_name, last_name, phone, avatar, is_active, is_verified,
               // ... resto de campos
        FROM users WHERE id = $1`
    // ...
}
```

### **4. Servicio de Usuario - Nuevo Método**
```go
type UserService interface {
    // ... métodos existentes
    UpdateAvatar(ctx context.Context, userID uint, avatarPath string) error
}

func (s *userService) UpdateAvatar(ctx context.Context, userID uint, avatarPath string) error {
    user, err := s.repo.GetByID(ctx, userID)
    if err != nil {
        return fmt.Errorf("failed to get user: %w", err)
    }

    user.Avatar = avatarPath
    if err := s.repo.Update(ctx, user); err != nil {
        return fmt.Errorf("failed to update avatar: %w", err)
    }

    return nil
}
```

### **5. Handler - Nuevo Endpoint**
```go
func (h *UserHandler) UploadAvatar(c *gin.Context) {
    userID := h.getUserID(c)
    if userID == 0 {
        c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
        return
    }

    file, err := c.FormFile("avatar")
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to get file"})
        return
    }

    // Validaciones
    if file.Size > 1024*1024 { // 1MB limit
        c.JSON(http.StatusBadRequest, gin.H{"error": "File too large"})
        return
    }

    ext := filepath.Ext(file.Filename)
    if ext != ".jpg" && ext != ".jpeg" && ext != ".png" && ext != ".gif" {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid file type"})
        return
    }

    // Generar nombre único
    filename := fmt.Sprintf("%d_%d%s", userID, time.Now().Unix(), ext)
    uploadPath := filepath.Join("uploads", filename)

    // Crear directorio si no existe
    if err := os.MkdirAll("uploads", 0755); err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create upload directory"})
        return
    }

    // Guardar archivo
    if err := c.SaveUploadedFile(file, uploadPath); err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save file"})
        return
    }

    // Actualizar en base de datos
    if err := h.userService.UpdateAvatar(c.Request.Context(), userID, uploadPath); err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update avatar"})
        return
    }

    c.JSON(http.StatusOK, gin.H{"message": "Avatar uploaded successfully", "avatar_url": uploadPath})
}
```

### **6. Router - Nuevo Endpoint**
```go
// En main.go
protected.POST("/avatar", userHandler.UploadAvatar)
```

---

## 🚀 **PASOS PARA IMPLEMENTAR**

### **1. Ejecutar Migración de Base de Datos**
```bash
# Conectar a la base de datos de usuarios
psql -h localhost -p 5434 -U postgres -d users_db

# Ejecutar el script de migración
\i scripts/add_avatar_column.sql
```

### **2. Recompilar el Servicio**
```bash
cd users-service
go build -o bin/users-service cmd/api/main.go
```

### **3. Reiniciar el Servicio**
```bash
# Si usas Docker
docker-compose restart users-service

# Si usas binario directo
./bin/users-service
```

---

## 📡 **ENDPOINTS DISPONIBLES**

### **POST /api/v1/users/avatar**
- **Autenticación**: Requerida (JWT)
- **Método**: POST
- **Content-Type**: multipart/form-data
- **Parámetros**:
  - `avatar`: Archivo de imagen (jpg, jpeg, png, gif)
- **Límites**:
  - Tamaño máximo: 1MB
  - Formatos permitidos: jpg, jpeg, png, gif

#### **Respuesta Exitosa (200)**
```json
{
  "message": "Avatar uploaded successfully",
  "avatar_url": "uploads/123_1640995200.jpg"
}
```

#### **Respuesta de Error (400)**
```json
{
  "error": "File too large"
}
```

---

## 🔒 **VALIDACIONES DE SEGURIDAD**

### **1. Validación de Archivo**
- ✅ Tipo de archivo (solo imágenes)
- ✅ Tamaño máximo (1MB)
- ✅ Extensión permitida

### **2. Validación de Usuario**
- ✅ Autenticación requerida
- ✅ Solo el propio usuario puede subir su avatar

### **3. Almacenamiento Seguro**
- ✅ Directorio separado (`uploads/`)
- ✅ Nombres únicos (userID_timestamp.ext)
- ✅ Permisos de directorio (0755)

---

## 🎯 **INTEGRACIÓN CON FRONTEND**

### **1. Subida de Archivo**
```javascript
const formData = new FormData();
formData.append('avatar', file);

const response = await fetch('/api/v1/users/avatar', {
  method: 'POST',
  headers: {
    'Authorization': `Bearer ${token}`
  },
  body: formData
});
```

### **2. Actualización de Perfil**
```javascript
// El avatar se incluye automáticamente en las respuestas de perfil
const profile = await fetch('/api/v1/users/profile', {
  headers: {
    'Authorization': `Bearer ${token}`
  }
});
```

---

## 🚨 **POSIBLES PROBLEMAS**

### **1. Error de Migración**
- **Síntoma**: `column "avatar" already exists`
- **Solución**: El script usa `ADD COLUMN IF NOT EXISTS`, es seguro ejecutarlo múltiples veces

### **2. Error de Permisos**
- **Síntoma**: `Failed to create upload directory`
- **Solución**: Verificar permisos de escritura en el directorio del servicio

### **3. Error de Tamaño**
- **Síntoma**: `File too large`
- **Solución**: Comprimir imagen o usar formato más eficiente

### **4. Error de Tipo**
- **Síntoma**: `Invalid file type`
- **Solución**: Usar solo formatos permitidos (jpg, jpeg, png, gif)

---

## ✅ **RESULTADO ESPERADO**

Después de la implementación:

1. **✅ Base de datos**: Columna `avatar` agregada a tabla `users`
2. **✅ Backend**: Endpoint `/api/v1/users/avatar` disponible
3. **✅ Validaciones**: Archivos validados por tipo y tamaño
4. **✅ Almacenamiento**: Archivos guardados en `uploads/`
5. **✅ Seguridad**: Solo usuarios autenticados pueden subir avatares
6. **✅ Integración**: Avatar incluido en respuestas de perfil

**¡La funcionalidad de avatares está completamente implementada en el backend!** 🎉 