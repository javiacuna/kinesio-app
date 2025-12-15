# Kinesio App

Kinesio App es una aplicación para la gestión integral de un consultorio de kinesiología. Permite administrar pacientes y turnos (agenda), y está preparada para incorporar módulos como planes de tratamiento, seguimiento/evolución clínica, materiales/ejercicios y reportes, según el alcance del Trabajo Final de Grado.

El backend está desarrollado en **Go** exponiendo una **API REST**, con persistencia en **PostgreSQL** y una arquitectura basada en **Clean Architecture** (separación de dominio, casos de uso, infraestructura y delivery HTTP). La autenticación/autorización se contempla mediante **Firebase** (en desarrollo puede utilizarse un modo demo para facilitar pruebas locales).

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
