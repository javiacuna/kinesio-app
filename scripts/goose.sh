#!/usr/bin/env bash
set -euo pipefail

# Requiere goose instalado: go install github.com/pressly/goose/v3/cmd/goose@latest
: "${DB_HOST:=localhost}"
: "${DB_PORT:=5432}"
: "${DB_NAME:=kinesio}"
: "${DB_USER:=kinesio}"
: "${DB_PASSWORD:=kinesio}"
: "${DB_SSLMODE:=disable}"

export GOOSE_DRIVER=postgres
export GOOSE_DBSTRING="host=$DB_HOST port=$DB_PORT user=$DB_USER password=$DB_PASSWORD dbname=$DB_NAME sslmode=$DB_SSLMODE"

goose -dir ./migrations "$@"
