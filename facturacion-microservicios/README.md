# Sistema de Facturación Digital con Microservicios

Proyecto para la asignatura **Programación Avanzada 2026_2**.  
Implementa una arquitectura de microservicios en Go (1.22+) con persistencia en archivos JSON atómicos y concurrentes, comunicación REST sobre HTTP, seguridad mediante JWT (`HS256`) y contraseñas hasheadas con `bcrypt`. La interfaz de usuario es una aplicación web desacoplada en JavaScript Vanilla con Bootstrap 5, preparada para integración con **Google Stitch** mediante el estándar `data-hook`.

---

## 📋 Requisitos Previos

- **Go:** Versión 1.22 o superior ([Descargar Go](https://go.dev/dl/)).
- **Navegador Web:** Chrome, Firefox, Edge o Safari.
- *(Opcional)* **Make** en entornos Linux/macOS o **Docker / Docker Compose**.

---

## 🚀 Inicio Rápido

### Método 1: Un Solo Comando con Go (¡El más rápido!)
Desde la carpeta raíz `CRUD/` (o entrando a `facturacion-microservicios/`), simplemente ejecuta:
```bash
go run .
```
Esto inicia automáticamente los 4 microservicios y el servidor web en una sola consola:
- 🔐 Auth: `http://localhost:8081`
- 📦 Productos: `http://localhost:8082`
- 👥 Clientes: `http://localhost:8083`
- 🧾 Facturación: `http://localhost:8084`
- 🌐 Frontend: `http://localhost:3000`

Para detener todo ordenadamente, solo presiona `Ctrl + C`.

---

### Método 2: Scripts en Segundo Plano (Windows PowerShell)
```powershell
# 1. Iniciar todos los servicios en segundo plano:
.\run-all.ps1

# 2. Abrir en el navegador:
Start-Process "http://localhost:3000"

# 3. Para detener todos los servicios:
.\run-all.ps1 -Stop
```

### En Linux / macOS
```bash
# 1. Dar permisos de ejecución:
chmod +x run-all.sh docs/pruebas/flujo-completo.sh

# 2. Iniciar todos los servicios:
./run-all.sh
# O usando Make:
make run

# 3. Para detener los servicios:
./run-all.sh stop
# O usando Make:
make stop
```

### Usando Docker Compose (Bonus)
```bash
docker-compose up --build
```

---

## 🔑 Credenciales de Prueba (Datos Semilla)

El sistema incluye usuarios precargados con hashes **bcrypt reales**:

| Rol | Correo Electrónico | Contraseña | Permisos |
| :--- | :--- | :--- | :--- |
| **Administrador** | `ricardo@example.com` | `admin123` | Emisión y gestión completa |
| **Cliente** | `maria@example.com` | `cliente123` | Emisión y gestión de compras |

*(Para generar nuevos hashes de contraseña se incluye la herramienta: `go run cmd/tools/hashpass/main.go <contraseña>`)*

---

## 🌐 Puertos y Servicios

| Servicio | Puerto | Directorio / Binario | Archivo de Persistencia |
| :--- | :---: | :--- | :--- |
| **Autenticación** | `8081` | `cmd/auth/main.go` | `data/usuarios.json` |
| **Productos** | `8082` | `cmd/productos/main.go` | `data/productos.json` |
| **Clientes** | `8083` | `cmd/clientes/main.go` | `data/clientes.json` |
| **Facturación** | `8084` | `cmd/facturacion/main.go` | `data/facturas.json` |
| **Frontend Web** | `3000` | `cmd/web/main.go` | Carpeta `frontend/` |

---

## ⚙️ Variables de Entorno

Cada servicio es configurable independientemente:

| Variable | Valor por Defecto | Descripción |
| :--- | :--- | :--- |
| `PORT` | `8081`, `8082`, `8083`, `8084`, `3000` | Puerto TCP en el que escucha el servicio |
| `DATA_FILE` | `data/<servicio>.json` | Ruta al archivo JSON de persistencia |
| `JWT_SECRET` | `mi-secreto-desarrollo-2026` | Clave secreta para firma y verificación HS256 |
| `AUTH_URL` | `http://localhost:8081` | URL base del servicio de Autenticación |
| `PRODUCTOS_URL` | `http://localhost:8082` | URL base del servicio de Productos |
| `CLIENTES_URL` | `http://localhost:8083` | URL base del servicio de Clientes |
| `FRONTEND_DIR` | `frontend` | Directorio con los archivos estáticos de la UI |

---

## 📂 Estructura del Proyecto

```
facturacion-microservicios/
├── go.mod                     → Dependencias Go (sólo jwt/v5 y crypto/bcrypt)
├── go.sum
├── Makefile                   → Automatización para Linux/macOS
├── run-all.sh                 → Script para levantar/detener todo en Unix
├── run-all.ps1                → Script para levantar/detener todo en Windows
├── Dockerfile                 → Definición multi-binario para contenedores
├── docker-compose.yml         → Orquestación de contenedores (Bonus)
├── README.md                  → Documentación general y guía de uso
├── cmd/
│   ├── auth/main.go           → Servicio de Autenticación (:8081)
│   ├── productos/main.go      → Servicio de Productos (:8082)
│   ├── clientes/main.go       → Servicio de Clientes (:8083)
│   ├── facturacion/main.go    → Servicio de Facturación (:8084)
│   ├── web/main.go            → Servidor estático de frontend (:3000)
│   └── tools/hashpass/main.go → Utilidad CLI para generar hashes bcrypt
├── internal/
│   ├── jsondb/                → Almacén genérico con sync.RWMutex y escrituras atómicas
│   ├── middleware/            → CORS, Logging, JWT HS256, Recover
│   ├── httpx/                 → Respuestas JSON y manejo uniforme de errores
│   ├── auth/                  → Lógica, modelos y endpoints de autenticación
│   ├── productos/             → Lógica, modelos y control de stock atómico
│   ├── clientes/              → Lógica, modelos y validaciones de clientes
│   └── facturacion/           → Orquestador de facturación y compensación
├── data/
│   ├── usuarios.json          → Usuarios semilla (bcrypt)
│   ├── productos.json         → Catálogo semilla
│   ├── clientes.json          → Clientes semilla
│   └── facturas.json          → Facturas históricas semilla
├── frontend/
│   ├── index.html             → Vista de Login y Registro
│   ├── app.html               → SPA con navegación: Productos, Clientes, Nueva Factura, etc.
│   ├── css/styles.css         → Estilos neutros y reglas @media print
│   └── js/
│       ├── config.js          → Centralización de URLs de microservicios
│       ├── api.js             → Fetch centralizado con Bearer token y auto-logout
│       ├── auth.js            → Lógica de sesión y autenticación
│       ├── productos.js       → CRUD y renderizado de productos con templates
│       ├── clientes.js        → CRUD y registro de clientes
│       ├── facturas.js        → Carrito, cálculo de totales, emisión y vistas
│       ├── pdf.js             → Generación de PDF en cliente (jsPDF)
│       └── ui.js              → Toasts, formato COP (Intl) y escape HTML (anti-XSS)
└── docs/
    ├── DOCUMENTO_TECNICO.md   → Especificación técnica, tablas de endpoints y diagramas Mermaid
    ├── FRONTEND_HOOKS.md      → Contrato de selectores data-hook para Google Stitch
    └── pruebas/
        ├── flujo-completo.sh  → Script automatizado en Bash/cURL
        └── flujo-completo.ps1 → Script automatizado en PowerShell
```

---

## 🧪 Ejecución de Pruebas

### 1. Pruebas Unitarias de Go
Ejecuta todas las pruebas en los paquetes internos (`jsondb`, `middleware`, `auth`, `productos`, `clientes`, `facturacion`):
```bash
go test -v ./...
# o con Makefile:
make test
```

### 2. Prueba Integral Automatizada (End-to-End)
Con los servicios corriendo (vía `.\run-all.ps1` o `./run-all.sh`), ejecute:

- **En Windows:**
  ```powershell
  .\docs\pruebas\flujo-completo.ps1
  ```
- **En Linux / macOS:**
  ```bash
  ./docs/pruebas/flujo-completo.sh
  ```

Esta prueba verifica de forma automática:
1. Health check de los 4 microservicios.
2. Login y obtención del token JWT.
3. Validación del usuario en `/auth/me`.
4. Creación de un cliente.
5. Emisión de una factura oficial.
6. **Comprobación de que el stock del producto disminuyó atómicamente**.
7. Consulta del historial de facturación.

### 3. Restaurar Datos Semilla
Si realizó pruebas de modificación o eliminación de datos y desea restaurar los datos semilla originales:
```bash
# En Linux/macOS:
make seed

# En Windows:
# Basta con volver a copiar los archivos desde data/ o git checkout data/
```

---

## 🎨 Contrato de Diseño con Google Stitch

La lógica JavaScript no depende de clases CSS. Toda la interactividad utiliza atributos `data-hook`. Puede reemplazar completamente el diseño visual en HTML y CSS guiándose con la tabla documentada en:
👉 [docs/FRONTEND_HOOKS.md](file:///c:/Users/Kiba/Desktop/CRUD/facturacion-microservicios/docs/FRONTEND_HOOKS.md)
