# Kinesio App

Kinesio App es una aplicación web para la gestión integral de un consultorio de kinesiología. 
Permite administrar pacientes y gestionar turnos mediante una agenda diaria, centralizando la información clínica y operativa del consultorio en una única plataforma.

El sistema está compuesto por un backend desarrollado en Go, siguiendo principios de Clean Architecture y separación de responsabilidades, y un frontend desarrollado en React con TypeScript, orientado a una experiencia de usuario simple y eficiente para el personal administrativo y profesional.

Kinesio App facilita:
- El registro y consulta de pacientes.
- La asignación, visualización, reprogramación y cancelación de turnos.
- El control de solapamientos en la agenda.
- La organización diaria del trabajo del kinesiólogo.

La aplicación fue diseñada como solución tecnológica para un entorno real de consultorio, priorizando mantenibilidad, claridad en los flujos de negocio y una arquitectura escalable.

## Requisitos
- Go 1.22+
- Docker + Docker Compose

## Levantar en local

1) Copiar variables de entorno:
```bash
cp .env.example .env
```

2) Levantar Postgres:
```bash
docker compose up -d db
```

3) Ejecutar la API:
```bash
go run ./cmd/api
```

API:
- Health: `GET http://localhost:8080/health`
- Version: `GET http://localhost:8080/version`

### Levantar DB + API con Docker

Si querés correr Postgres y la API juntos con Docker:

```bash
docker compose up --build
```

La API va a:
- esperar a que Postgres esté saludable,
- correr las migraciones automáticamente,
- levantar en `http://localhost:8080`.

Para Firebase, `docker-compose.yml` monta el archivo indicado en `GOOGLE_APPLICATION_CREDENTIALS`
como credencial dentro del contenedor.

### Firebase Auth

Para usar autenticación real con Firebase:

1. Activar Email/Password en Firebase Authentication.
2. Configurar estas variables en `.env`:

```bash
FIREBASE_PROJECT_ID=proyecto-final-b67f5
FIREBASE_WEB_API_KEY=tu_web_api_key
GOOGLE_APPLICATION_CREDENTIALS=/ruta/absoluta/firebase-service-account.json
```

El login del backend queda disponible en:

```bash
POST /api/v1/auth/login
```

El resto de endpoints puede recibir el token con:

```bash
Authorization: Bearer <firebase_id_token>
```

### Notificaciones por email

Las notificaciones in-app pueden enviarse también por email usando SMTP. Si no se activa esta configuración,
la aplicación sigue guardando las notificaciones dentro del sistema.

```bash
NOTIFICATIONS_EMAIL_ENABLED=true
SMTP_HOST=smtp.example.com
SMTP_PORT=587
SMTP_USERNAME=usuario
SMTP_PASSWORD=password
SMTP_FROM_EMAIL=no-reply@example.com
SMTP_FROM_NAME=Kinesio App
```

En Docker local el proyecto incluye Mailpit:

- SMTP interno para la API: `mailpit:1025`
- Bandeja web: `http://localhost:8025`

Eventos incluidos en esta versión:
- turno creado, reprogramado y cancelado;
- paquete de turnos creado y actualizado;
- plan de ejercicios creado y actualizado.

### Videollamadas automáticas

Los turnos virtuales pueden usar link manual o generar una sala automáticamente desde el backend.
Para activar la generación automática, configurá un proveedor:

```bash
# daily o whereby
VIDEO_CALL_PROVIDER=daily
DAILY_API_KEY=tu_daily_api_key

# Alternativa Whereby
# VIDEO_CALL_PROVIDER=whereby
# WHEREBY_API_KEY=tu_whereby_api_key
# WHEREBY_ROOM_PREFIX=kinesio

# Minutos extra que la sala queda activa después del fin del turno
VIDEO_CALL_EXPIRY_MIN=60
```

Si no hay proveedor configurado, la aplicación mantiene el flujo manual: se puede pegar el link en el turno.

En el frontend:

- `/login`: formulario de ingreso con email y contraseña.
- `recepcionista` y `admin`: redirigen a `/agenda` y pueden entrar a `/patients`.
- `kinesiologo`: redirige a `/agenda`.
- `paciente`: redirige a `/portal`.
- La interfaz es responsive y cuenta con manifest + service worker para instalarse como PWA desde navegadores compatibles.

Para asignar roles a usuarios de Firebase:

```bash
go run ./cmd/firebase-role --email recepcion@test.com --role recepcionista
go run ./cmd/firebase-role --email admin@test.com --role admin
go run ./cmd/firebase-role --email kine@test.com --role kinesiologo
go run ./cmd/firebase-role --email paciente@test.com --role paciente
```

## Frontend

El frontend está desarrollado con React + TypeScript + Vite y se encuentra en la carpeta `frontend/`.  
Durante el desarrollo local, el frontend utiliza un proxy de Vite para comunicarse con el backend sin problemas de CORS.

### Requisitos
- Node.js 18+
- npm 9+

### Levantar el frontend en local

1) Ir a la carpeta del frontend:
```bash
cd frontend
```

2) Instalar dependencias:
```bash
npm install
```

3) Ejecutar el servidor de desarrollo:
```bash
npm run dev
```

4) El frontend quedará disponible en:
```bash
http://localhost:5173
```

Comunicación con el backend

Para desarrollo local, el frontend asume que el backend está corriendo en:
```bash
http://localhost:8080
```

Las llamadas a la API se realizan mediante rutas relativas (/api/v1/...) y son redirigidas automáticamente al backend a través del proxy configurado en Vite.
