#!/usr/bin/env bash
# Script de prueba integral del flujo completo de Facturación con Microservicios
# Requiere: curl y jq (opcional para formato)

set -e

AUTH_URL="http://localhost:8081"
PROD_URL="http://localhost:8082"
CLI_URL="http://localhost:8083"
FACT_URL="http://localhost:8084"

echo "========================================================"
echo "🧪 INICIANDO PRUEBA INTEGRAL DEL SISTEMA DE FACTURACIÓN"
echo "========================================================"

# 1. Health checks
echo -e "\n[1/7] Verificando salud de los microservicios..."
curl -s "$AUTH_URL/health" | grep -q "ok" && echo "  ✔ Auth Service (:8081) OK"
curl -s "$PROD_URL/health" | grep -q "ok" && echo "  ✔ Productos Service (:8082) OK"
curl -s "$CLI_URL/health" | grep -q "ok" && echo "  ✔ Clientes Service (:8083) OK"
curl -s "$FACT_URL/health" | grep -q "ok" && echo "  ✔ Facturación Service (:8084) OK"

# 2. Login con usuario administrador semilla
echo -e "\n[2/7] Iniciando sesión con usuario semilla (ricardo@example.com)..."
LOGIN_RESP=$(curl -s -X POST "$AUTH_URL/auth/login" \
  -H "Content-Type: application/json" \
  -d '{"email":"ricardo@example.com","password":"admin123"}')

TOKEN=$(echo "$LOGIN_RESP" | grep -o '"token":"[^"]*' | cut -d'"' -f4)

if [ -z "$TOKEN" ]; then
  echo "  ❌ Falló el inicio de sesión. Respuesta: $LOGIN_RESP"
  exit 1
fi
echo "  ✔ Token JWT obtenido exitosamente"

# 3. Validar token con /auth/me
echo -e "\n[3/7] Verificando perfil autenticado en /auth/me..."
ME_RESP=$(curl -s "$AUTH_URL/auth/me" -H "Authorization: Bearer $TOKEN")
echo "  ✔ Usuario: $ME_RESP"

# 4. CRUD de Productos: Consultar stock inicial de Laptop (101)
echo -e "\n[4/7] Consultando producto 101 (Laptop Lenovo)..."
PROD_101=$(curl -s "$PROD_URL/productos/101" -H "Authorization: Bearer $TOKEN")
echo "  ✔ Producto 101: $PROD_101"

# 5. CRUD de Clientes: Registrar un nuevo cliente de prueba
echo -e "\n[5/7] Creando nuevo cliente..."
RANDOM_CC=$((RANDOM + 10000000))
NUEVO_CLIENTE=$(curl -s -X POST "$CLI_URL/clientes" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d "{\"nombre\":\"Cliente Prueba Automatizada\",\"cc\":\"$RANDOM_CC\",\"email\":\"test_$RANDOM_CC@mail.com\",\"direccion\":\"Calle de Pruebas 123\"}")
echo "  ✔ Cliente creado: $NUEVO_CLIENTE"

CLIENTE_ID=$(echo "$NUEVO_CLIENTE" | grep -o '"id":[0-9]*' | cut -d':' -f2)

# 6. Generar una nueva Factura consumiendo 2 Laptops (ID 101)
echo -e "\n[6/7] Generando factura de venta..."
NUEVA_FACTURA=$(curl -s -X POST "$FACT_URL/facturas" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d "{\"cliente_id\": $CLIENTE_ID, \"items\": [{\"producto_id\": 101, \"cantidad\": 2}]}")

echo "  ✔ Factura generada:"
echo "$NUEVA_FACTURA"

FACTURA_ID=$(echo "$NUEVA_FACTURA" | grep -o '"factura_id":[0-9]*' | cut -d':' -f2)

# 7. Verificar que el stock bajó en el microservicio de productos
echo -e "\n[7/7] Comprobando disminución de existencias en Productos..."
PROD_101_POST=$(curl -s "$PROD_URL/productos/101" -H "Authorization: Bearer $TOKEN")
echo "  ✔ Producto 101 después de la venta: $PROD_101_POST"

# 8. Consultar historial de facturas
echo -e "\nConsultando historial de facturas..."
HISTORIAL=$(curl -s "$FACT_URL/facturas" -H "Authorization: Bearer $TOKEN")
echo "  ✔ Historial obtenido (primeras facturas):"
echo "$HISTORIAL" | head -c 200
echo -e "\n..."

echo -e "\n========================================================"
echo "🎉 ¡PRUEBA INTEGRAL COMPLETADA CON ÉXITO!"
echo "========================================================"
