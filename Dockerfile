FROM node:22-bookworm AS frontend-build

WORKDIR /app/frontend

COPY frontend/package.json frontend/package-lock.json ./
RUN npm ci

COPY frontend/ .
RUN npm run build

FROM golang:1.25-bookworm AS backend-build

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go install github.com/pressly/goose/v3/cmd/goose@latest \
  && go build -o /usr/local/bin/kinesio-api ./cmd/api

FROM debian:bookworm-slim

RUN apt-get update && apt-get install -y --no-install-recommends ca-certificates && rm -rf /var/lib/apt/lists/*

WORKDIR /app

COPY --from=backend-build /usr/local/bin/kinesio-api /usr/local/bin/kinesio-api
COPY --from=backend-build /go/bin/goose /usr/local/bin/goose
COPY --from=backend-build /app/migrations ./migrations
COPY --from=frontend-build /app/frontend/dist ./web

ENV STATIC_DIR=/app/web

EXPOSE 8080

CMD ["sh", "-c", "goose -dir ./migrations up && kinesio-api"]
