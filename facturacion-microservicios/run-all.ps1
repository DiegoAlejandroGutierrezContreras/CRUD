# Script PowerShell para iniciar/detener todos los microservicios en Windows
# Uso: .\run-all.ps1 [-Stop] [-Foreground]

param(
    [switch]$Stop,
    [switch]$Foreground
)

$scriptDir = Split-Path -Parent $MyInvocation.MyCommand.Path
if (-not $scriptDir) { $scriptDir = (Get-Location).Path }

$PidFile = Join-Path $scriptDir ".running_pids.json"

function Stop-AllServices {
    if (Test-Path $PidFile) {
        Write-Host "Deteniendo servicios en ejecucion..." -ForegroundColor Yellow
        try {
            $pidsList = Get-Content $PidFile | ConvertFrom-Json
            foreach ($p in $pidsList) {
                if ($p.Id) {
                    try {
                        Stop-Process -Id $p.Id -Force -ErrorAction SilentlyContinue
                        Write-Host "  Detenido: $($p.Name) (PID: $($p.Id))" -ForegroundColor Gray
                    } catch {
                    }
                }
            }
        } catch {
        }
        Remove-Item $PidFile -Force -ErrorAction SilentlyContinue
    }
    
    # Asegurar liberacion de puertos 8081, 8082, 8083, 8084, 3000
    $ports = @(8081, 8082, 8083, 8084, 3000)
    foreach ($port in $ports) {
        $connections = Get-NetTCPConnection -LocalPort $port -ErrorAction SilentlyContinue
        if ($connections) {
            foreach ($c in $connections) {
                Stop-Process -Id $c.OwningProcess -Force -ErrorAction SilentlyContinue
            }
        }
    }
    Write-Host "Todos los servicios han sido detenidos correctamente." -ForegroundColor Green
}

if ($Stop) {
    Stop-AllServices
    exit 0
}

# Si ya hay servicios corriendo, detenerlos primero
Stop-AllServices

Write-Host "==========================================================" -ForegroundColor Cyan
Write-Host "Levantando Sistema de Facturacion Digital Microservicios" -ForegroundColor Cyan
Write-Host "==========================================================" -ForegroundColor Cyan

$binDir = Join-Path $scriptDir "bin"
if (-not (Test-Path $binDir)) {
    New-Item -ItemType Directory -Path $binDir -Force | Out-Null
}

Write-Host "Compilando binarios de microservicios..." -ForegroundColor Yellow
& go build -o (Join-Path $binDir "auth.exe") cmd/auth/main.go
& go build -o (Join-Path $binDir "productos.exe") cmd/productos/main.go
& go build -o (Join-Path $binDir "clientes.exe") cmd/clientes/main.go
& go build -o (Join-Path $binDir "facturacion.exe") cmd/facturacion/main.go
& go build -o (Join-Path $binDir "web.exe") cmd/web/main.go

function Start-DaemonProcess($name, $exePath, $port) {
    $p = Start-Process -FilePath $exePath -WorkingDirectory $scriptDir -WindowStyle Hidden -PassThru
    Write-Host "Servicio $name iniciado (PID: $($p.Id)) en http://localhost:$port" -ForegroundColor Green
    return [PSCustomObject]@{
        Name = $name
        Id   = $p.Id
        Port = $port
    }
}

$services = @()
$services += Start-DaemonProcess "Auth" (Join-Path $binDir "auth.exe") 8081
$services += Start-DaemonProcess "Productos" (Join-Path $binDir "productos.exe") 8082
$services += Start-DaemonProcess "Clientes" (Join-Path $binDir "clientes.exe") 8083
$services += Start-DaemonProcess "Facturacion" (Join-Path $binDir "facturacion.exe") 8084
$services += Start-DaemonProcess "Web" (Join-Path $binDir "web.exe") 3000

# Guardar lista de PIDs correctamente formateada
$services | ConvertTo-Json | Set-Content $PidFile

Write-Host "----------------------------------------------------------" -ForegroundColor Yellow
Write-Host "Todos los servicios estan corriendo." -ForegroundColor Yellow
Write-Host "  Abra su navegador en: http://localhost:3000" -ForegroundColor Cyan
Write-Host "  Para detener todo ejecute: .\run-all.ps1 -Stop" -ForegroundColor Yellow
Write-Host "==========================================================" -ForegroundColor Cyan

if ($Foreground) {
    Write-Host "Presione Ctrl+C para detener todos los servicios..." -ForegroundColor Yellow
    try {
        while ($true) {
            Start-Sleep -Seconds 1
        }
    } finally {
        Stop-AllServices
    }
}
