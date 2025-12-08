# 🌐 **SOLUCIÓN AL PROBLEMA DE CORS**

## 📋 **PROBLEMA IDENTIFICADO**

El frontend está intentando hacer peticiones al backend de usuarios desde `localhost:3000` hacia `localhost:8083`, pero el backend no está configurado para permitir el header `x-caller-id` que está enviando el frontend.

### **Error Original:**
```
Access to XMLHttpRequest at 'http://localhost:8083/api/v1/users/profile' 
from origin 'http://localhost:3000' has been blocked by CORS policy: 
Request header field x-caller-id is not allowed by Access-Control-Allow-Headers 
in preflight response.
```

---

## 🔧 **SOLUCIÓN IMPLEMENTADA**

### **1. Actualización del Middleware CORS**
```go
// CORS middleware (basic)
r.Use(func(c *gin.Context) {
    c.Header("Access-Control-Allow-Origin", "*")
    c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
    c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Authorization, x-caller-id")

    if c.Request.Method == "OPTIONS" {
        c.AbortWithStatus(204)
        return
    }

    c.Next()
})
```

### **2. Cambio Realizado**
- **Antes**: `"Origin, Content-Type, Authorization"`
- **Después**: `"Origin, Content-Type, Authorization, x-caller-id"`

---

## 🎯 **EXPLICACIÓN DEL PROBLEMA**

### **¿Qué es CORS?**
CORS (Cross-Origin Resource Sharing) es un mecanismo de seguridad que controla cómo los navegadores permiten peticiones entre diferentes orígenes (dominios, puertos, protocolos).

### **¿Por qué ocurre el error?**
1. **Frontend**: `http://localhost:3000` (React dev server)
2. **Backend**: `http://localhost:8083` (Users service)
3. **Header enviado**: `x-caller-id` (para identificación de servicios)
4. **Problema**: El backend no permitía este header en la configuración CORS

### **¿Qué es el header x-caller-id?**
Es un header personalizado que usa el frontend para identificar qué servicio está haciendo la petición, útil para:
- **Logging**: Identificar el origen de las peticiones
- **Debugging**: Rastrear peticiones entre servicios
- **Monitoreo**: Analizar patrones de uso

---

## 🚀 **PASOS PARA APLICAR LA SOLUCIÓN**

### **1. Recompilar el Servicio**
```bash
cd users-service
go build -o bin/users-service cmd/api/main.go
```

### **2. Reiniciar el Servicio**
```bash
# Si usas Docker
docker-compose restart users-service

# Si usas binario directo
./bin/users-service
```

### **3. Verificar que Funciona**
```bash
# Probar el endpoint directamente
curl -X OPTIONS http://localhost:8083/api/v1/users/profile \
  -H "Origin: http://localhost:3000" \
  -H "Access-Control-Request-Method: GET" \
  -H "Access-Control-Request-Headers: x-caller-id" \
  -v
```

---

## 🔍 **VERIFICACIÓN DE LA SOLUCIÓN**

### **1. Headers de Respuesta Esperados**
```http
Access-Control-Allow-Origin: *
Access-Control-Allow-Methods: GET, POST, PUT, DELETE, OPTIONS
Access-Control-Allow-Headers: Origin, Content-Type, Authorization, x-caller-id
```

### **2. Petición Preflight Exitosa**
```javascript
// El navegador debería poder hacer la petición preflight
fetch('http://localhost:8083/api/v1/users/profile', {
  method: 'GET',
  headers: {
    'Authorization': 'Bearer token',
    'x-caller-id': 'frontend'
  }
})
.then(response => response.json())
.then(data => console.log(data));
```

---

## 🚨 **POSIBLES PROBLEMAS ADICIONALES**

### **1. Si el error persiste**
- **Verificar**: Que el servicio se reinició correctamente
- **Checkear**: Logs del servicio para errores de compilación
- **Solución**: Reiniciar completamente el contenedor Docker

### **2. Si aparecen otros headers faltantes**
- **Síntoma**: Error similar con otros headers
- **Solución**: Agregar el header faltante a `Access-Control-Allow-Headers`

### **3. Si hay problemas de autenticación**
- **Verificar**: Que el token JWT sea válido
- **Checkear**: Que el endpoint requiera autenticación
- **Solución**: Asegurar que el token se envía correctamente

---

## ✅ **RESULTADO ESPERADO**

Después de aplicar la solución:

1. **✅ CORS**: Las peticiones del frontend al backend funcionan
2. **✅ Headers**: El header `x-caller-id` es permitido
3. **✅ Autenticación**: Las peticiones autenticadas funcionan
4. **✅ Perfil**: La actualización de perfil funciona correctamente
5. **✅ Avatar**: La subida de avatares funciona

**¡El problema de CORS está resuelto y el frontend puede comunicarse correctamente con el backend de usuarios!** 🎉 