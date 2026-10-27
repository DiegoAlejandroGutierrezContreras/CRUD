FROM golang:1.22-alpine AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /bin/auth cmd/auth/main.go
RUN CGO_ENABLED=0 GOOS=linux go build -o /bin/productos cmd/productos/main.go
RUN CGO_ENABLED=0 GOOS=linux go build -o /bin/clientes cmd/clientes/main.go
RUN CGO_ENABLED=0 GOOS=linux go build -o /bin/facturacion cmd/facturacion/main.go
RUN CGO_ENABLED=0 GOOS=linux go build -o /bin/web cmd/web/main.go

FROM alpine:3.19
WORKDIR /app
COPY --from=builder /bin/* /usr/local/bin/
COPY data/ /app/data/
COPY frontend/ /app/frontend/

EXPOSE 8081 8082 8083 8084 3000
