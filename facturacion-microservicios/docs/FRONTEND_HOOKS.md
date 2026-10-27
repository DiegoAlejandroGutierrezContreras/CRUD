# Contrato de Interfaz para Google Stitch (FRONTEND_HOOKS.md)

Este documento describe la especificación técnica de integración entre el frontend JavaScript y las plantillas HTML. Para permitir que el diseño visual sea reemplazado o enriquecido mediante **Google Stitch** sin romper la lógica del sistema, **la lógica JavaScript no depende de nombres de clases CSS ni de posiciones relativas en el DOM**. Toda la vinculación se realiza mediante atributos `data-hook`.

---

## 1. Convenciones Generales
- **Atributo Hook**: `data-hook="nombre-del-hook"`.
- **Estructuras Repetitivas**: Se instancian clonando elementos `<template data-hook="...">` donde cada celda o botón interno posee su respectivo `data-hook`.
- **Notificaciones**: Los mensajes emergentes se insertan dentro del contenedor con `data-hook="toast-container"`.

---

## 2. Pantalla: Acceso y Registro (`index.html`)

| `data-hook` | Tipo de Elemento Esperado | Descripción / Comportamiento en JavaScript |
| :--- | :--- | :--- |
| `toast-container` | `div` | Contenedor donde se insertan las alertas emergentes (toasts). |
| `tab-login-btn` | `button` / `a` | Botón/Pestaña para conmutar a la vista de Iniciar Sesión. |
| `tab-register-btn` | `button` / `a` | Botón/Pestaña para conmutar a la vista de Registro de Usuario. |
| `login-form` | `form` | Formulario de inicio de sesión. Captura evento `submit`. |
| `login-email` | `input[type="email"]` | Campo de texto para el correo del usuario. |
| `login-password` | `input[type="password"]` | Campo de texto para la contraseña. |
| `btn-login-submit` | `button[type="submit"]` | Botón de envío del formulario de login; muestra estado de carga. |
| `register-form` | `form` | Formulario de registro de cuenta. Captura evento `submit`. |
| `register-nombre` | `input[type="text"]` | Campo de texto para el nombre completo del usuario. |
| `register-email` | `input[type="email"]` | Campo de texto para el correo a registrar. |
| `register-password` | `input[type="password"]` | Campo para la contraseña (mínimo 6 caracteres). |
| `btn-register-submit` | `button[type="submit"]` | Botón de envío del formulario de registro; muestra estado de carga. |

---

## 3. Barra de Navegación y Sesión Global (`app.html`)

| `data-hook` | Tipo de Elemento Esperado | Descripción / Comportamiento en JavaScript |
| :--- | :--- | :--- |
| `nav-brand` | `a` / `div` | Logotipo o título de la aplicación. |
| `nav-tab-productos` | `button` / `a` | Pestaña para abrir la vista de Productos. |
| `nav-tab-clientes` | `button` / `a` | Pestaña para abrir la vista de Clientes. |
| `nav-tab-nueva-factura`| `button` / `a` | Pestaña para abrir la vista de Generar Factura. |
| `nav-tab-historial` | `button` / `a` | Pestaña para abrir la vista de Historial de Facturas. |
| `nav-tab-detalle` | `button` / `a` | Pestaña para abrir la vista de Detalle de Factura. |
| `nav-user-nombre` | `span` / `div` | Muestra el nombre del usuario autenticado en la sesión. |
| `nav-user-rol` | `span` / `div` | Muestra el rol (`admin` o `cliente`) del usuario autenticado. |
| `btn-logout` | `button` | Cierra la sesión activa (borra JWT de almacenamiento y redirige). |

---

## 4. Pantalla: Catálogo de Productos (`app.html` - `#view-productos`)

| `data-hook` | Tipo de Elemento Esperado | Descripción / Comportamiento en JavaScript |
| :--- | :--- | :--- |
| `vista-productos` | `section` / `div` | Contenedor principal de la vista de productos. |
| `buscar-producto-input` | `input[type="text"]` | Input de búsqueda con debounce (consulta `GET /productos?q=`). |
| `btn-nuevo-producto` | `button` | Abre el modal para crear un nuevo producto. |
| `productos-tabla-body` | `tbody` / `div` | Contenedor donde se insertan las filas de los productos. |
| `template-producto-fila` | `template` | Plantilla HTML para renderizar cada producto individual. |
| *Dentro del template de producto:* | | |
| `item-producto-id` | `td` / `span` | Muestra el ID autoincremental del producto. |
| `item-producto-nombre` | `td` / `span` | Muestra el nombre comercial del producto. |
| `item-producto-precio` | `td` / `span` | Muestra el precio formateado en pesos colombianos (`COP`). |
| `item-producto-stock` | `td` / `span` | Muestra el stock disponible (resalta en rojo si es `<= 0`). |
| `btn-editar-producto` | `button` | Abre el modal de edición precargando los datos del producto. |
| `btn-eliminar-producto` | `button` | Solicita confirmación y ejecuta `DELETE /productos/{id}`. |
| *Modal de Producto:* | | |
| `producto-form` | `form` | Formulario de creación y edición de productos. |
| `producto-modal-titulo`| `h5` / `span` | Título dinámico ("Nuevo Producto" o "Editar Producto #ID"). |
| `producto-id-input` | `input[type="hidden"]`| Contiene el ID del producto al editar (o vacío al crear). |
| `producto-nombre-input`| `input[type="text"]` | Campo de entrada para el nombre del producto. |
| `producto-precio-input`| `input[type="number"]`| Campo de entrada para el precio unitario. |
| `producto-stock-input` | `input[type="number"]`| Campo de entrada para las existencias disponibles. |
| `btn-guardar-producto` | `button[type="submit"]`| Envía el formulario; ejecuta `POST /productos` o `PUT /productos/{id}`. |

---

## 5. Pantalla: Gestión de Clientes (`app.html` - `#view-clientes`)

| `data-hook` | Tipo de Elemento Esperado | Descripción / Comportamiento en JavaScript |
| :--- | :--- | :--- |
| `vista-clientes` | `section` / `div` | Contenedor principal de la vista de clientes. |
| `buscar-cliente-input` | `input[type="text"]` | Input de búsqueda con debounce por nombre o cédula. |
| `btn-nuevo-cliente` | `button` | Abre el modal para registrar un nuevo cliente. |
| `clientes-tabla-body` | `tbody` / `div` | Contenedor donde se insertan las filas de clientes. |
| `template-cliente-fila` | `template` | Plantilla HTML para renderizar cada cliente. |
| *Dentro del template de cliente:* | | |
| `item-cliente-id` | `td` / `span` | Muestra el ID del cliente. |
| `item-cliente-nombre` | `td` / `span` | Muestra el nombre completo del cliente. |
| `item-cliente-cc` | `td` / `span` | Muestra la cédula de ciudadanía (numérica). |
| `item-cliente-email` | `td` / `span` | Muestra la dirección de correo electrónico. |
| `item-cliente-direccion`| `td` / `span` | Muestra la dirección física de envío o residencia. |
| `btn-editar-cliente` | `button` | Abre el modal de edición precargando los datos del cliente. |
| `btn-eliminar-cliente` | `button` | Solicita confirmación y ejecuta `DELETE /clientes/{id}`. |
| *Modal de Cliente:* | | |
| `cliente-form` | `form` | Formulario de registro y edición de clientes. |
| `cliente-modal-titulo` | `h5` / `span` | Título dinámico ("Nuevo Cliente" o "Editar Cliente #ID"). |
| `cliente-id-input` | `input[type="hidden"]`| Contiene el ID del cliente al editar (o vacío al crear). |
| `cliente-nombre-input` | `input[type="text"]` | Campo de entrada para el nombre completo. |
| `cliente-cc-input` | `input[type="text"]` | Campo para la cédula (debe contener solo dígitos). |
| `cliente-email-input` | `input[type="email"]` | Campo para el correo electrónico. |
| `cliente-direccion-input`| `input[type="text"]` | Campo para la dirección física. |
| `btn-guardar-cliente` | `button[type="submit"]`| Envía el formulario; ejecuta `POST /clientes` o `PUT /clientes/{id}`. |

---

## 6. Pantalla: Nueva Factura (`app.html` - `#view-nueva-factura`)

| `data-hook` | Tipo de Elemento Esperado | Descripción / Comportamiento en JavaScript |
| :--- | :--- | :--- |
| `vista-nueva-factura` | `section` / `div` | Contenedor principal de emisión de facturas. |
| `factura-usuario-actual`| `span` / `div` | Muestra el usuario logueado que generará la factura. |
| `select-cliente` | `select` | Desplegable de selección del cliente destinatario. |
| `select-producto` | `select` | Desplegable de productos disponibles con precio y stock. |
| `input-cantidad-producto`| `input[type="number"]`| Campo para definir la cantidad a agregar al carrito. |
| `btn-agregar-carrito` | `button` | Agrega el producto seleccionado con validación de stock local. |
| `carrito-tabla-body` | `tbody` / `div` | Contenedor donde se insertan los productos del carrito. |
| `template-carrito-fila` | `template` | Plantilla para cada ítem en el carrito de compras. |
| *Dentro del template de carrito:* | | |
| `item-carrito-nombre` | `td` / `span` | Nombre del producto en el carrito. |
| `item-carrito-precio` | `td` / `span` | Precio unitario formateado en `COP`. |
| `item-carrito-cantidad`| `td` / `span` | Cantidad de unidades agregadas. |
| `item-carrito-subtotal`| `td` / `span` | Subtotal calculado (`precio * cantidad`). |
| `btn-quitar-carrito` | `button` | Elimina la fila del carrito y recalcula el total. |
| *Totales y Acción:* | | |
| `carrito-total` | `span` / `div` | Total acumulado de la factura en tiempo real. |
| `btn-generar-factura` | `button` | Envía `POST /facturas` al microservicio y abre el detalle. |

---

## 7. Pantalla: Detalle de Factura (`app.html` - `#view-detalle`)

| `data-hook` | Tipo de Elemento Esperado | Descripción / Comportamiento en JavaScript |
| :--- | :--- | :--- |
| `vista-detalle` | `section` / `div` | Contenedor principal de la vista de detalle. |
| `factura-detalle-vista` | `div` | Caja de factura con estilos `@media print` para impresión limpia. |
| `detalle-factura-id` | `span` | Número oficial de la factura generada (`factura_id`). |
| `detalle-factura-fecha` | `span` | Fecha de expedición (`YYYY-MM-DD`). |
| `detalle-cliente-nombre`| `span` / `div` | Nombre del cliente receptor. |
| `detalle-cliente-cc` | `span` | Cédula del cliente. |
| `detalle-cliente-email` | `span` | Correo electrónico del cliente. |
| `detalle-cliente-direccion`| `span` | Dirección registrada del cliente. |
| `detalle-usuario-nombre`| `span` / `div` | Nombre del usuario/facturador que expidió la venta. |
| `detalle-usuario-rol` | `span` | Rol del facturador. |
| `detalle-productos-body`| `tbody` / `div` | Contenedor de las filas de productos facturados. |
| `template-detalle-producto-fila` | `template` | Plantilla para cada ítem facturado. |
| *Dentro del template detalle producto:* | | |
| `item-detalle-prod-nombre` | `td` / `span` | Nombre del producto. |
| `item-detalle-prod-cantidad`| `td` / `span` | Cantidad comprada. |
| `item-detalle-prod-precio` | `td` / `span` | Precio unitario oficial al momento de la venta. |
| `item-detalle-prod-subtotal`| `td` / `span` | Subtotal del renglón. |
| `detalle-factura-total` | `span` / `th` | Valor total de la factura formateado en `COP`. |
| *Botones de Exportación / Acción:* | | |
| `btn-ver-json` | `button` | Abre el modal de inspección JSON de la factura. |
| `btn-descargar-json` | `button` | Descarga directa del archivo JSON `factura-<id>.json`. |
| `btn-imprimir-pdf` | `button` | Invoca `window.print()` con estilos de impresión limpia. |
| `btn-descargar-pdf` | `button` | Genera y descarga el PDF en el cliente mediante `jsPDF`. |
| *Modal Ver JSON:* | | |
| `modal-json-pre` | `pre` | Elemento donde se dibuja el JSON formateado con sangría. |
| `btn-copiar-json` | `button` | Copia el contenido JSON de la factura al portapapeles. |

---

## 8. Pantalla: Historial de Facturas (`app.html` - `#view-historial`)

| `data-hook` | Tipo de Elemento Esperado | Descripción / Comportamiento en JavaScript |
| :--- | :--- | :--- |
| `vista-historial` | `section` / `div` | Contenedor principal de la tabla de historial. |
| `facturas-tabla-body` | `tbody` / `div` | Contenedor donde se insertan las filas del historial. |
| `template-factura-fila`| `template` | Plantilla para cada factura en el listado histórico. |
| *Dentro del template de factura:* | | |
| `item-factura-id` | `td` / `span` | Muestra el ID de la factura (ej: `#5001`). |
| `item-factura-fecha` | `td` / `span` | Muestra la fecha de emisión. |
| `item-factura-cliente` | `td` / `span` | Muestra el nombre del cliente. |
| `item-factura-total` | `td` / `span` | Muestra el total de la factura formateado. |
| `btn-ver-factura` | `button` | Carga los datos en la vista de Detalle y conmuta a ella. |
