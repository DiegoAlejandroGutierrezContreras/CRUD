# Script de prueba integral en PowerShell para Windows
$ErrorActionPreference = "Stop"

$AuthUrl = "http://localhost:8081"
$ProdUrl = "http://localhost:8082"
$CliUrl  = "http://localhost:8083"
$FactUrl = "http://localhost:8084"

Write-Host "========================================================" -ForegroundColor Cyan
Write-Host "INICIANDO PRUEBA INTEGRAL DEL SISTEMA DE FACTURACION" -ForegroundColor Cyan
Write-Host "========================================================" -ForegroundColor Cyan

# 1. Health checks
Write-Host ""
Write-Host "[1/7] Verificando salud de los microservicios..." -ForegroundColor Yellow
$hAuth = Invoke-RestMethod -Uri "$AuthUrl/health" -Method Get
$hProd = Invoke-RestMethod -Uri "$ProdUrl/health" -Method Get
$hCli  = Invoke-RestMethod -Uri "$CliUrl/health" -Method Get
$hFact = Invoke-RestMethod -Uri "$FactUrl/health" -Method Get

Write-Host "  [OK] Auth: $($hAuth.status) - $($hAuth.servicio)" -ForegroundColor Green
Write-Host "  [OK] Productos: $($hProd.status) - $($hProd.servicio)" -ForegroundColor Green
Write-Host "  [OK] Clientes: $($hCli.status) - $($hCli.servicio)" -ForegroundColor Green
Write-Host "  [OK] Facturacion: $($hFact.status) - $($hFact.servicio)" -ForegroundColor Green

# 2. Login
Write-Host ""
Write-Host "[2/7] Iniciando sesion con ricardo@example.com / admin123..." -ForegroundColor Yellow
$loginBody = '{"email":"ricardo@example.com","password":"admin123"}'
$loginResp = Invoke-RestMethod -Uri "$AuthUrl/auth/login" -Method Post -Body $loginBody -ContentType "application/json"

$Token = $loginResp.token
$Headers = @{ Authorization = "Bearer $Token" }
Write-Host "  [OK] Token JWT obtenido para: $($loginResp.usuario.nombre)" -ForegroundColor Green

# 3. GET /auth/me
Write-Host ""
Write-Host "[3/7] Consultando /auth/me con token..." -ForegroundColor Yellow
$me = Invoke-RestMethod -Uri "$AuthUrl/auth/me" -Method Get -Headers $Headers
Write-Host "  [OK] Sesion verificada: $($me.nombre) (Rol: $($me.rol))" -ForegroundColor Green

# 4. Consultar stock inicial del producto 101
Write-Host ""
Write-Host "[4/7] Consultando stock inicial de Laptop Lenovo (101)..." -ForegroundColor Yellow
$prod101 = Invoke-RestMethod -Uri "$ProdUrl/productos/101" -Method Get -Headers $Headers
$stockInicial = [int]$prod101.stock
Write-Host "  [OK] Stock actual de $($prod101.nombre): $stockInicial unidades" -ForegroundColor Green

# 5. Crear nuevo cliente
Write-Host ""
Write-Host "[5/7] Creando cliente de prueba..." -ForegroundColor Yellow
$randCC = Get-Random -Minimum 10000000 -Maximum 99999999
$cliBody = "{""nombre"":""Cliente PowerShell $randCC"",""cc"":""$randCC"",""email"":""ps_$randCC@test.com"",""direccion"":""Calle 123""}"

$nuevoCli = Invoke-RestMethod -Uri "$CliUrl/clientes" -Method Post -Body $cliBody -ContentType "application/json" -Headers $Headers
Write-Host "  [OK] Cliente creado con ID: $($nuevoCli.id), CC: $($nuevoCli.cc)" -ForegroundColor Green

# 6. Generar Factura descontando 2 unidades del producto 101
Write-Host ""
Write-Host "[6/7] Generando factura con 2 unidades de $($prod101.nombre)..." -ForegroundColor Yellow
$facturaBody = "{""cliente_id"":$($nuevoCli.id),""items"":[{""producto_id"":101,""cantidad"":2}]}"

$nuevaFactura = Invoke-RestMethod -Uri "$FactUrl/facturas" -Method Post -Body $facturaBody -ContentType "application/json" -Headers $Headers
Write-Host "  [OK] Factura No. $($nuevaFactura.factura_id) generada exitosamente!" -ForegroundColor Green
Write-Host "       Total: $($nuevaFactura.total)" -ForegroundColor Green

# 7. Verificar que el stock bajó en 2 unidades
Write-Host ""
Write-Host "[7/7] Verificando descuento de stock en el microservicio de productos..." -ForegroundColor Yellow
$prod101Post = Invoke-RestMethod -Uri "$ProdUrl/productos/101" -Method Get -Headers $Headers
$stockFinal = [int]$prod101Post.stock
Write-Host "  [OK] Stock anterior: $stockInicial, Stock nuevo: $stockFinal" -ForegroundColor Green

if ($stockFinal -ne ($stockInicial - 2)) {
    Write-Host "  [ERROR] El stock no disminuyo en 2 unidades" -ForegroundColor Red
    exit 1
} else {
    Write-Host "  [COMPROBADO] El stock se desconto atomicamente de forma correcta." -ForegroundColor Green
}

# 8. Consultar historial
Write-Host ""
Write-Host "Consultando historial de facturas..." -ForegroundColor Yellow
$historial = Invoke-RestMethod -Uri "$FactUrl/facturas" -Method Get -Headers $Headers
Write-Host "  [OK] Se obtuvieron $($historial.Count) facturas del historial." -ForegroundColor Green

Write-Host ""
Write-Host "========================================================" -ForegroundColor Cyan
Write-Host "PRUEBA INTEGRAL EN POWERSHELL COMPLETADA CON EXITO!" -ForegroundColor Cyan
Write-Host "========================================================" -ForegroundColor Cyan
