# Sistema de Facturación Digital (Microservicios)

Proyecto de Programación Avanzada 2026_2. Sistema de facturación basado en arquitectura de microservicios desarrollado en Go utilizando exclusivamente la librería estándar `net/http`, base de datos persistente en archivos JSON y una interfaz de usuario web desarrollada en JavaScript.

---

## Requisitos

- **Go 1.22** o superior instalado en el sistema.
- Navegador web moderno (Chrome, Edge, Firefox, Safari).

---

## Puertos y Servicios

| Servicio | Puerto | Descripción | Archivo de Persistencia |
| :--- | :---: | :--- | :--- |
| **Autenticación** | `8081` | Registro, login y validación de usuarios (JWT + bcrypt) | `data/usuarios.json` |
| **Productos** | `8082` | CRUD de catálogo de productos, precios y stock | `data/productos.json` |
| **Clientes** | `8083` | CRUD de clientes con nombre, cédula, correo y dirección | `data/clientes.json` |
| **Facturación** | `8084` | Orquestador que consume los 3 anteriores y emite facturas | `data/facturas.json` |
| **Frontend Web** | `3000` | Servidor web estático que sirve la interfaz de usuario | Carpeta `frontend/` |

---

## Credenciales de Prueba

Usuarios registrados en `data/usuarios.json` con contraseña cifrada mediante bcrypt:

- **Administrador:** `ricardo@example.com` / `admin123`
- **Cliente:** `maria@example.com` / `cliente123`

---

## Ejecución

### Opción 1: Iniciar todos los servicios con el script

**En Windows (PowerShell):**
```powershell
# Iniciar los 5 servicios en segundo plano
.\run-all.ps1

# Detener todos los servicios
.\run-all.ps1 -Stop
```

**En Linux / macOS (Bash):**
```bash
# Iniciar los 5 servicios en segundo plano
chmod +x run-all.sh
./run-all.sh

# Detener todos los servicios
./run-all.sh stop
```

Una vez iniciados, abrir en el navegador: **http://localhost:3000**

---

### Opción 2: Iniciar cada servicio de forma individual

Abrir una terminal para cada microservicio y ejecutar en la carpeta del proyecto:

```bash
# Terminal 1 - Autenticación
go run cmd/auth/main.go

# Terminal 2 - Productos
go run cmd/productos/main.go

# Terminal 3 - Clientes
go run cmd/clientes/main.go

# Terminal 4 - Facturación
go run cmd/facturacion/main.go

# Terminal 5 - Frontend Web
go run cmd/web/main.go
```

---

## Estructura del Proyecto

```text
facturacion-microservicios/
├── cmd/
│   ├── auth/         # Punto de entrada del microservicio de autenticación (:8081)
│   ├── productos/    # Punto de entrada del microservicio de productos (:8082)
│   ├── clientes/     # Punto de entrada del microservicio de clientes (:8083)
│   ├── facturacion/  # Punto de entrada del microservicio de facturación (:8084)
│   └── web/          # Servidor HTTP para archivos estáticos del frontend (:3000)
├── data/
│   ├── usuarios.json   # Base de datos de usuarios
│   ├── productos.json  # Base de datos de catálogo de productos
│   ├── clientes.json   # Base de datos de clientes
│   └── facturas.json   # Base de datos de facturas emitidas
├── docs/
│   ├── DOCUMENTO_TECNICO.md   # Documento técnico de arquitectura y endpoints
│   └── DOCUMENTO_TECNICO.pdf  # Versión PDF del documento técnico
├── frontend/
│   ├── index.html    # Página de acceso (login y registro)
│   ├── app.html      # Aplicación web (pedido, clientes, visualización de factura)
│   ├── css/          # Hojas de estilo y reglas @media print
│   └── js/           # Módulos JavaScript (api, auth, productos, clientes, facturas, ui)
├── internal/
│   ├── auth/         # Lógica, modelos y handlers de autenticación
│   ├── productos/    # Lógica, modelos y handlers de productos
│   ├── clientes/     # Lógica, modelos y handlers de clientes
│   ├── facturacion/  # Lógica, modelos y handlers de facturación
│   ├── httpx/        # Utilidades HTTP (JSON decoder, respuestas y errores)
│   ├── jsondb/       # Almacén atómico y concurrente sobre archivos JSON
│   └── middleware/   # Middlewares (CORS y validación JWT)
├── go.mod            # Definición del módulo Go
├── go.sum            # Checksums de dependencias
├── run-all.ps1       # Script para Windows
└── run-all.sh        # Script para Linux/macOS
```
