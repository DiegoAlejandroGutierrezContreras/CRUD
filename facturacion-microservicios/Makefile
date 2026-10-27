# Makefile para el Sistema de Facturación Digital con Microservicios
.PHONY: help run stop test seed clean build

SHELL := /bin/bash

help:
	@echo "Comandos disponibles:"
	@echo "  make run       - Iniciar todos los microservicios y el servidor web"
	@echo "  make stop      - Detener todos los servicios en ejecución"
	@echo "  make test      - Ejecutar todas las pruebas unitarias en Go"
	@echo "  make seed      - Restaurar los datos semilla iniciales"
	@echo "  make build     - Compilar todos los binarios en la carpeta bin/"

test:
	go test -v ./...

seed:
	@echo "Restaurando datos semilla..."
	@mkdir -p data
	@echo '[{"id":1,"nombre":"Ricardo Gómez","email":"ricardo@example.com","password":"$$2a$$10$$cFnMVBoqOq8mml/szXxjceFo1/MyUpKzHxrnWHc5UALo6SLDa9vHm","rol":"admin"},{"id":2,"nombre":"María López","email":"maria@example.com","password":"$$2a$$10$$5cC9PsHwnqkf6D0kqKqzduIpAUplQt26l.JJKMMblCuCcnGg.fhvW","rol":"cliente"}]' > data/usuarios.json
	@echo '[{"id":101,"nombre":"Laptop Lenovo ThinkPad","precio":3500000.00,"stock":15},{"id":102,"nombre":"Mouse inalámbrico Logitech","precio":12000.00,"stock":50},{"id":103,"nombre":"Monitor Samsung 24\'\'","precio":80000.00,"stock":20}]' > data/productos.json
	@echo '[{"id":201,"nombre":"Carlos Pérez","cc":"1234567","email":"carlos@example.com","direccion":"Calle 123 #45-67, Bogotá"},{"id":202,"nombre":"Ana Torres","cc":"56789101","email":"ana@example.com","direccion":"Carrera 10 #20-30, Medellín"}]' > data/clientes.json
	@echo '[{"factura_id":5001,"cliente":{"id":201,"nombre":"Carlos Pérez","cc":"1234567","email":"carlos@example.com","direccion":"Calle 123 #45-67, Bogotá"},"usuario":{"id":1,"nombre":"Ricardo Gómez","rol":"admin"},"productos":[{"id":101,"nombre":"Laptop Lenovo ThinkPad","cantidad":1,"precio_unitario":3500000.00,"subtotal":3500000.00},{"id":102,"nombre":"Mouse inalámbrico Logitech","cantidad":2,"precio_unitario":12000.00,"subtotal":24000.00}],"total":3524000.00,"fecha":"2026-10-18"},{"factura_id":5002,"cliente":{"id":202,"nombre":"Ana Torres","cc":"56789101","email":"ana@example.com","direccion":"Carrera 10 #20-30, Medellín"},"usuario":{"id":1,"nombre":"Ricardo Gómez","rol":"admin"},"productos":[{"id":103,"nombre":"Monitor Samsung 24\'\'","cantidad":2,"precio_unitario":80000.00,"subtotal":160000.00},{"id":102,"nombre":"Mouse inalámbrico Logitech","cantidad":1,"precio_unitario":12000.00,"subtotal":12000.00}],"total":172000.00,"fecha":"2026-10-19"}]' > data/facturas.json
	@echo "Datos semilla restaurados con éxito."

build:
	@mkdir -p bin
	go build -o bin/auth cmd/auth/main.go
	go build -o bin/productos cmd/productos/main.go
	go build -o bin/clientes cmd/clientes/main.go
	go build -o bin/facturacion cmd/facturacion/main.go
	go build -o bin/web cmd/web/main.go
	go build -o bin/hashpass cmd/tools/hashpass/main.go
	@echo "Binarios compilados en bin/"

run:
	@bash ./run-all.sh

stop:
	@pkill -f "cmd/auth" || true
	@pkill -f "cmd/productos" || true
	@pkill -f "cmd/clientes" || true
	@pkill -f "cmd/facturacion" || true
	@pkill -f "cmd/web" || true
	@echo "Servicios detenidos."
