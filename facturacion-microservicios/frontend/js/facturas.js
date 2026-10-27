// Módulo de Facturación: Creación, Visualización, Historial y Exportación
import { CONFIG } from './config.js';
import { apiFetch, getUsuario } from './api.js';
import { mostrarToast, formatearMoneda, getByHook, setButtonLoading } from './ui.js';
import { descargarFacturaPDF } from './pdf.js';
import { getProductosCache, cargarProductos } from './productos.js';
import { getClientesCache, cargarClientes } from './clientes.js';

let carrito = []; // Arreglo de items en el carrito: { producto, cantidad, subtotal }
let facturaActual = null;
let jsonModal = null;

/**
 * Inicializa la lógica de facturación (nueva factura, detalle e historial).
 */
export function initFacturas() {
  const modalEl = document.getElementById('modal-ver-json');
  if (modalEl && window.bootstrap) {
    jsonModal = new window.bootstrap.Modal(modalEl);
  }

  // --- Nueva Factura ---
  const btnAgregarCarrito = getByHook('btn-agregar-carrito');
  if (btnAgregarCarrito) {
    btnAgregarCarrito.addEventListener('click', agregarAlCarrito);
  }

  const btnGenerarFactura = getByHook('btn-generar-factura');
  if (btnGenerarFactura) {
    btnGenerarFactura.addEventListener('click', generarFactura);
  }

  // --- Detalle de Factura ---
  const btnVerJSON = getByHook('btn-ver-json');
  if (btnVerJSON) {
    btnVerJSON.addEventListener('click', mostrarModalJSON);
  }

  const btnCopiarJSON = getByHook('btn-copiar-json');
  if (btnCopiarJSON) {
    btnCopiarJSON.addEventListener('click', copiarJSONAlPortapapeles);
  }

  const btnDescargarJSON = getByHook('btn-descargar-json');
  if (btnDescargarJSON) {
    btnDescargarJSON.addEventListener('click', descargarJSONFactura);
  }

  const btnImprimirPDF = getByHook('btn-imprimir-pdf');
  if (btnImprimirPDF) {
    btnImprimirPDF.addEventListener('click', () => {
      window.print();
    });
  }

  const btnDescargarPDF = getByHook('btn-descargar-pdf');
  if (btnDescargarPDF) {
    btnDescargarPDF.addEventListener('click', () => {
      if (facturaActual) {
        descargarFacturaPDF(facturaActual);
      }
    });
  }

  // Cargar datos necesarios
  poblarSelectoresNuevaFactura();
  cargarHistorialFacturas();
}

/**
 * Llena los selects de cliente y productos para la pantalla de Nueva Factura.
 */
export async function poblarSelectoresNuevaFactura() {
  const selectCliente = getByHook('select-cliente');
  const selectProducto = getByHook('select-producto');
  const userFacturador = getByHook('factura-usuario-actual');

  // Mostrar usuario logueado
  const usuario = getUsuario();
  if (userFacturador && usuario) {
    userFacturador.textContent = `${usuario.nombre} (${usuario.rol})`;
  }

  // Clientes
  if (selectCliente) {
    let clientes = getClientesCache();
    if (clientes.length === 0) {
      await cargarClientes();
      clientes = getClientesCache();
    }

    selectCliente.innerHTML = '<option value="">-- Seleccione un cliente --</option>';
    clientes.forEach((c) => {
      const opt = document.createElement('option');
      opt.value = c.id;
      opt.textContent = `${c.nombre} (CC: ${c.cc})`;
      selectCliente.appendChild(opt);
    });
  }

  // Productos
  if (selectProducto) {
    let productos = getProductosCache();
    if (productos.length === 0) {
      await cargarProductos();
      productos = getProductosCache();
    }

    selectProducto.innerHTML = '<option value="">-- Seleccione un producto --</option>';
    productos.forEach((p) => {
      const opt = document.createElement('option');
      opt.value = p.id;
      opt.textContent = `${p.nombre} - ${formatearMoneda(p.precio)} (Stock: ${p.stock})`;
      if (p.stock <= 0) {
        opt.disabled = true;
        opt.textContent += ' [SIN STOCK]';
      }
      selectProducto.appendChild(opt);
    });
  }
}

function agregarAlCarrito() {
  const selectProducto = getByHook('select-producto');
  const inputCantidad = getByHook('input-cantidad-producto');

  if (!selectProducto || !inputCantidad) return;

  const productoId = parseInt(selectProducto.value, 10);
  const cantidad = parseInt(inputCantidad.value, 10);

  if (isNaN(productoId) || productoId <= 0) {
    mostrarToast('Por favor seleccione un producto', 'warning');
    return;
  }

  if (isNaN(cantidad) || cantidad <= 0) {
    mostrarToast('La cantidad debe ser mayor a 0', 'warning');
    return;
  }

  const productos = getProductosCache();
  const producto = productos.find((p) => p.id === productoId);
  if (!producto) {
    mostrarToast('Producto no encontrado', 'danger');
    return;
  }

  // Verificar cantidad acumulada en carrito vs stock disponible
  const itemExistente = carrito.find((i) => i.producto.id === productoId);
  const cantidadTotal = (itemExistente ? itemExistente.cantidad : 0) + cantidad;

  if (cantidadTotal > producto.stock) {
    mostrarToast(
      `No puede agregar ${cantidad} unidades. Stock disponible: ${producto.stock} (en carrito: ${itemExistente ? itemExistente.cantidad : 0})`,
      'warning'
    );
    return;
  }

  if (itemExistente) {
    itemExistente.cantidad += cantidad;
    itemExistente.subtotal = itemExistente.cantidad * producto.precio;
  } else {
    carrito.push({
      producto,
      cantidad,
      subtotal: cantidad * producto.precio,
    });
  }

  inputCantidad.value = '1';
  renderizarCarrito();
  mostrarToast(`"${producto.nombre}" agregado al carrito`, 'success');
}

function renderizarCarrito() {
  const tbody = getByHook('carrito-tabla-body');
  const template = getByHook('template-carrito-fila');
  const totalEl = getByHook('carrito-total');

  if (!tbody || !template) return;

  tbody.innerHTML = '';

  if (carrito.length === 0) {
    const tr = document.createElement('tr');
    tr.innerHTML = `<td colspan="5" class="text-center text-muted py-3">El carrito está vacío. Agregue productos arriba.</td>`;
    tbody.appendChild(tr);
    if (totalEl) totalEl.textContent = formatearMoneda(0);
    return;
  }

  let total = 0;

  carrito.forEach((item, index) => {
    total += item.subtotal;
    const clone = template.content.cloneNode(true);

    const elNombre = clone.querySelector('[data-hook="item-carrito-nombre"]');
    const elPrecio = clone.querySelector('[data-hook="item-carrito-precio"]');
    const elCantidad = clone.querySelector('[data-hook="item-carrito-cantidad"]');
    const elSubtotal = clone.querySelector('[data-hook="item-carrito-subtotal"]');
    const btnQuitar = clone.querySelector('[data-hook="btn-quitar-carrito"]');

    if (elNombre) elNombre.textContent = item.producto.nombre;
    if (elPrecio) elPrecio.textContent = formatearMoneda(item.producto.precio);
    if (elCantidad) elCantidad.textContent = item.cantidad;
    if (elSubtotal) elSubtotal.textContent = formatearMoneda(item.subtotal);

    if (btnQuitar) {
      btnQuitar.addEventListener('click', () => {
        carrito.splice(index, 1);
        renderizarCarrito();
      });
    }

    tbody.appendChild(clone);
  });

  if (totalEl) {
    totalEl.textContent = formatearMoneda(total);
  }
}

async function generarFactura() {
  const selectCliente = getByHook('select-cliente');
  const btnGenerar = getByHook('btn-generar-factura');

  if (!selectCliente) return;

  const clienteId = parseInt(selectCliente.value, 10);
  if (isNaN(clienteId) || clienteId <= 0) {
    mostrarToast('Por favor seleccione un cliente', 'warning');
    return;
  }

  if (carrito.length === 0) {
    mostrarToast('El carrito está vacío. Agregue al menos un producto.', 'warning');
    return;
  }

  const items = carrito.map((item) => ({
    producto_id: item.producto.id,
    cantidad: item.cantidad,
  }));

  setButtonLoading(btnGenerar, true, 'Generando factura...');

  try {
    const factura = await apiFetch(CONFIG.FACTURAS_URL, '/facturas', {
      method: 'POST',
      body: JSON.stringify({
        cliente_id: clienteId,
        items,
      }),
    });

    if (factura && factura.factura_id) {
      mostrarToast(`¡Factura #${factura.factura_id} generada con éxito!`, 'success');
      // Limpiar carrito
      carrito = [];
      renderizarCarrito();

      // Recargar catálogo de productos para ver el stock actualizado
      await cargarProductos();
      await poblarSelectoresNuevaFactura();

      // Mostrar la factura generada
      mostrarDetalleFactura(factura);

      // Cambiar de vista a Detalle de Factura
      const tabDetalleBtn = getByHook('nav-tab-detalle');
      if (tabDetalleBtn) {
        tabDetalleBtn.click();
      }

      // Actualizar historial
      cargarHistorialFacturas();
    }
  } catch (err) {
    // Error ya manejado por apiFetch
  } finally {
    setButtonLoading(btnGenerar, false);
  }
}

/**
 * Carga el historial de facturas.
 */
export async function cargarHistorialFacturas() {
  const tbody = getByHook('facturas-tabla-body');
  const template = getByHook('template-factura-fila');
  if (!tbody || !template) return;

  try {
    const facturas = await apiFetch(CONFIG.FACTURAS_URL, '/facturas');
    renderizarHistorialFacturas(facturas || []);
  } catch (err) {
    // Error manejado
  }
}

function renderizarHistorialFacturas(facturas) {
  const tbody = getByHook('facturas-tabla-body');
  const template = getByHook('template-factura-fila');
  if (!tbody || !template) return;

  tbody.innerHTML = '';

  if (facturas.length === 0) {
    const tr = document.createElement('tr');
    tr.innerHTML = `<td colspan="5" class="text-center text-muted py-4">No hay facturas registradas</td>`;
    tbody.appendChild(tr);
    return;
  }

  facturas.forEach((fac) => {
    const clone = template.content.cloneNode(true);

    const elId = clone.querySelector('[data-hook="item-factura-id"]');
    const elFecha = clone.querySelector('[data-hook="item-factura-fecha"]');
    const elCliente = clone.querySelector('[data-hook="item-factura-cliente"]');
    const elTotal = clone.querySelector('[data-hook="item-factura-total"]');
    const btnVer = clone.querySelector('[data-hook="btn-ver-factura"]');

    if (elId) elId.textContent = `#${fac.factura_id}`;
    if (elFecha) elFecha.textContent = fac.fecha;
    if (elCliente) elCliente.textContent = fac.cliente ? fac.cliente.nombre : 'N/A';
    if (elTotal) elTotal.textContent = formatearMoneda(fac.total);

    if (btnVer) {
      btnVer.addEventListener('click', () => {
        mostrarDetalleFactura(fac);
        const tabDetalleBtn = getByHook('nav-tab-detalle');
        if (tabDetalleBtn) {
          tabDetalleBtn.click();
        }
      });
    }

    tbody.appendChild(clone);
  });
}

/**
 * Pinta los datos de una factura en la pantalla de Detalle.
 * @param {object} factura
 */
export function mostrarDetalleFactura(factura) {
  facturaActual = factura;

  const elId = getByHook('detalle-factura-id');
  const elFecha = getByHook('detalle-factura-fecha');
  const elCliNombre = getByHook('detalle-cliente-nombre');
  const elCliCC = getByHook('detalle-cliente-cc');
  const elCliEmail = getByHook('detalle-cliente-email');
  const elCliDir = getByHook('detalle-cliente-direccion');
  const elUsuNombre = getByHook('detalle-usuario-nombre');
  const elUsuRol = getByHook('detalle-usuario-rol');
  const elTotal = getByHook('detalle-factura-total');
  const tbody = getByHook('detalle-productos-body');
  const template = getByHook('template-detalle-producto-fila');

  if (elId) elId.textContent = factura.factura_id;
  if (elFecha) elFecha.textContent = factura.fecha;

  if (factura.cliente) {
    if (elCliNombre) elCliNombre.textContent = factura.cliente.nombre || '';
    if (elCliCC) elCliCC.textContent = factura.cliente.cc || '';
    if (elCliEmail) elCliEmail.textContent = factura.cliente.email || '';
    if (elCliDir) elCliDir.textContent = factura.cliente.direccion || '';
  }

  if (factura.usuario) {
    if (elUsuNombre) elUsuNombre.textContent = factura.usuario.nombre || '';
    if (elUsuRol) elUsuRol.textContent = factura.usuario.rol || '';
  }

  if (elTotal) elTotal.textContent = formatearMoneda(factura.total);

  // Tabla de productos de la factura
  if (tbody && template) {
    tbody.innerHTML = '';
    (factura.productos || []).forEach((prod) => {
      const clone = template.content.cloneNode(true);
      const rowNombre = clone.querySelector('[data-hook="item-detalle-prod-nombre"]');
      const rowCant = clone.querySelector('[data-hook="item-detalle-prod-cantidad"]');
      const rowPrecio = clone.querySelector('[data-hook="item-detalle-prod-precio"]');
      const rowSubtotal = clone.querySelector('[data-hook="item-detalle-prod-subtotal"]');

      if (rowNombre) rowNombre.textContent = prod.nombre;
      if (rowCant) rowCant.textContent = prod.cantidad;
      if (rowPrecio) rowPrecio.textContent = formatearMoneda(prod.precio_unitario);
      if (rowSubtotal) rowSubtotal.textContent = formatearMoneda(prod.subtotal);

      tbody.appendChild(clone);
    });
  }
}

function mostrarModalJSON() {
  if (!facturaActual) return;
  const jsonPre = getByHook('modal-json-pre');
  if (jsonPre) {
    jsonPre.textContent = JSON.stringify(facturaActual, null, 2);
  }
  if (jsonModal) jsonModal.show();
}

function copiarJSONAlPortapapeles() {
  if (!facturaActual) return;
  const jsonText = JSON.stringify(facturaActual, null, 2);
  navigator.clipboard.writeText(jsonText).then(
    () => mostrarToast('¡JSON copiado al portapapeles!', 'success'),
    () => mostrarToast('No se pudo copiar el JSON', 'warning')
  );
}

async function descargarJSONFactura() {
  if (!facturaActual) return;
  try {
    const res = await apiFetch(CONFIG.FACTURAS_URL, `/facturas/${facturaActual.factura_id}/descargar`);
    if (res && res.blob) {
      const url = window.URL.createObjectURL(res.blob);
      const a = document.createElement('a');
      a.href = url;
      a.download = res.filename || `factura-${facturaActual.factura_id}.json`;
      document.body.appendChild(a);
      a.click();
      a.remove();
      window.URL.revokeObjectURL(url);
    } else {
      // Descarga alternativa directa
      const dataStr = 'data:text/json;charset=utf-8,' + encodeURIComponent(JSON.stringify(facturaActual, null, 2));
      const downloadAnchor = document.createElement('a');
      downloadAnchor.setAttribute('href', dataStr);
      downloadAnchor.setAttribute('download', `factura-${facturaActual.factura_id}.json`);
      document.body.appendChild(downloadAnchor);
      downloadAnchor.click();
      downloadAnchor.remove();
    }
    mostrarToast('Descarga iniciada', 'success');
  } catch (err) {
    // Si falla la descarga del endpoint, descargar con Blob local
    const dataStr = 'data:text/json;charset=utf-8,' + encodeURIComponent(JSON.stringify(facturaActual, null, 2));
    const downloadAnchor = document.createElement('a');
    downloadAnchor.setAttribute('href', dataStr);
    downloadAnchor.setAttribute('download', `factura-${facturaActual.factura_id}.json`);
    document.body.appendChild(downloadAnchor);
    downloadAnchor.click();
    downloadAnchor.remove();
  }
}
