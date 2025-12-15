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