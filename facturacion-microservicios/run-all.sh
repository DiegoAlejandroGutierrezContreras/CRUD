#!/usr/bin/env bash
# Script para iniciar todos los microservicios en Linux/macOS
# Uso: ./run-all.sh [stop]

PID_FILE=".running_pids"

stop_services() {
  if [ -f "$PID_FILE" ]; then
    echo "Deteniendo servicios en ejecución..."
    while read -r pid; do
      if kill -0 "$pid" 2>/dev/null; then
        kill "$pid" 2>/dev/null || true
      fi
    done < "$PID_FILE"
    rm -f "$PID_FILE"
    echo "Todos los servicios han sido detenidos."
  else
    echo "No se encontró archivo de PIDs ($PID_FILE)."
  fi
}

if [ "$1" == "stop" ]; then
  stop_services
  exit 0
fi

# Si ya hay servicios corriendo, detenerlos primero
if [ -f "$PID_FILE" ]; then
  stop_services
fi

echo "=========================================================="
echo "🚀 Levantando Sistema de Facturación Digital Microservicios"
echo "=========================================================="

mkdir -p logs

# Iniciar Servicio de Autenticación
PORT=8081 go run cmd/auth/main.go > logs/auth.log 2>&1 &
PID_AUTH=$!
echo "🔐 Servicio Auth iniciado (PID: $PID_AUTH) en http://localhost:8081"

# Iniciar Servicio de Productos
PORT=8082 go run cmd/productos/main.go > logs/productos.log 2>&1 &
PID_PROD=$!
echo "📦 Servicio Productos iniciado (PID: $PID_PROD) en http://localhost:8082"

# Iniciar Servicio de Clientes
PORT=8083 go run cmd/clientes/main.go > logs/clientes.log 2>&1 &
PID_CLI=$!
echo "👥 Servicio Clientes iniciado (PID: $PID_CLI) en http://localhost:8083"

# Iniciar Servicio de Facturación
PORT=8084 go run cmd/facturacion/main.go > logs/facturacion.log 2>&1 &
PID_FACT=$!
echo "🧾 Servicio Facturación iniciado (PID: $PID_FACT) en http://localhost:8084"

# Iniciar Servidor Web Estático
PORT=3000 go run cmd/web/main.go > logs/web.log 2>&1 &
PID_WEB=$!
echo "🌐 Servidor Web iniciado (PID: $PID_WEB) en http://localhost:3000"

# Guardar PIDs
echo "$PID_AUTH" > "$PID_FILE"
echo "$PID_PROD" >> "$PID_FILE"
echo "$PID_CLI" >> "$PID_FILE"
echo "$PID_FACT" >> "$PID_FILE"
echo "$PID_WEB" >> "$PID_FILE"

echo "----------------------------------------------------------"
echo "✔ Todos los servicios están corriendo en segundo plano."
echo "  Abra su navegador en: http://localhost:3000"
echo "  Para detener todo ejecute: ./run-all.sh stop"
echo "=========================================================="
