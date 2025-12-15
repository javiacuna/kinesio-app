# Kinesio App

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
