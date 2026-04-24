FROM golang:1.25-bookworm

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go install github.com/pressly/goose/v3/cmd/goose@latest \
  && go build -o /usr/local/bin/kinesio-api ./cmd/api

EXPOSE 8080

CMD ["sh", "-c", "/go/bin/goose -dir ./migrations up && /usr/local/bin/kinesio-api"]
