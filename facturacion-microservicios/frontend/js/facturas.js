// Módulo de Facturación: Carrito, Generación, Visualización e Impresión en PDF
import { CONFIG } from './config.js';
import { apiFetch, getUsuario } from './api.js';
import { mostrarToast, formatearMoneda, getByHook, setButtonLoading } from './ui.js';
import { getProductosCache, cargarProductos } from './productos.js';
import { getClientesCache, cargarClientes } from './clientes.js';

let carrito = [];
let facturaActual = null;
let jsonModal = null;

/**
 * Inicializa la lógica de facturación.
 */
export function initFacturas() {
  const modalEl = document.getElementById('modal-ver-json');
  if (modalEl && window.bootstrap) {
    jsonModal = new window.bootstrap.Modal(modalEl);
  }

  const btnAgregarCarrito = getByHook('btn-agregar-carrito');
  if (btnAgregarCarrito) {
    btnAgregarCarrito.addEventListener('click', agregarAlCarrito);
  }

  const btnGenerarFactura = getByHook('btn-generar-factura');
  if (btnGenerarFactura) {
    btnGenerarFactura.addEventListener('click', generarFactura);
  }

  const btnVerJSON = getByHook('btn-ver-json');
  if (btnVerJSON) {
    btnVerJSON.addEventListener('click', mostrarModalJSON);
  }

  const btnImprimirPDF = getByHook('btn-imprimir-pdf');
  if (btnImprimirPDF) {
    btnImprimirPDF.addEventListener('click', () => {
      window.print();
    });
  }

  poblarSelectoresNuevaFactura();
}

/**
 * Llena los selectores de clientes y productos.
 */
export async function poblarSelectoresNuevaFactura() {
  const selectCliente = getByHook('select-cliente');
  const selectProducto = getByHook('select-producto');
  const userFacturador = getByHook('factura-usuario-actual');

  const usuario = getUsuario();
  if (userFacturador && usuario) {
    userFacturador.textContent = `${usuario.nombre} (${usuario.rol})`;
  }

  if (selectCliente) {
    let clientes = getClientesCache();
    if (clientes.length === 0) {
      await cargarClientes();
      clientes = getClientesCache();
    }

    const valorPrevio = selectCliente.value;
    selectCliente.innerHTML = '<option value="">-- Seleccione un cliente --</option>';
    clientes.forEach((c) => {
      const opt = document.createElement('option');
      opt.value = c.id;
      opt.textContent = `${c.nombre} (CC: ${c.cc})`;
      selectCliente.appendChild(opt);
    });
    if (valorPrevio) selectCliente.value = valorPrevio;
  }

  if (selectProducto) {
    let productos = getProductosCache();
    if (productos.length === 0) {
      await cargarProductos();
      productos = getProductosCache();
    }

    const valorPrevio = selectProducto.value;
    selectProducto.innerHTML = '<option value="">-- Seleccione un producto --</option>';
    productos.forEach((p) => {
      const opt = document.createElement('option');
      opt.value = p.id;
      opt.textContent = `${p.nombre} - ${formatearMoneda(p.precio)} (Stock: ${p.stock})`;
      selectProducto.appendChild(opt);
    });
    if (valorPrevio) selectProducto.value = valorPrevio;
  }
}

/**
 * Agrega un producto seleccionado al carrito.
 */
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

  const itemExistente = carrito.find((i) => i.producto.id === productoId);
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
  mostrarToast(`"${producto.nombre}" agregado al pedido`, 'success');
}

/**
 * Renderiza el carrito de compras en la tabla.
 */
function renderizarCarrito() {
  const tbody = getByHook('carrito-tabla-body');
  const template = getByHook('template-carrito-fila');
  const totalEl = getByHook('carrito-total');

  if (!tbody || !template) return;

  tbody.innerHTML = '';

  if (carrito.length === 0) {
    const tr = document.createElement('tr');
    tr.innerHTML = `<td colspan="5" class="text-center text-muted py-3">El carrito está vacío. Agregue productos.</td>`;
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

/**
 * Envía la factura al servicio de facturación (POST /facturas).
 */
async function generarFactura() {
  const selectCliente = getByHook('select-cliente');
  const btnGenerar = getByHook('btn-generar-factura');

  if (!selectCliente) return;

  const clienteId = parseInt(selectCliente.value, 10);
  if (isNaN(clienteId) || clienteId <= 0) {
    mostrarToast('Por favor seleccione un cliente destinatario', 'warning');
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
      carrito = [];
      renderizarCarrito();

      mostrarDetalleFactura(factura);

      const tabDetalleBtn = getByHook('nav-tab-detalle');
      if (tabDetalleBtn) {
        tabDetalleBtn.click();
      }
    }
  } catch (err) {
    // Error manejado por apiFetch
  } finally {
    setButtonLoading(btnGenerar, false);
  }
}

/**
 * Pinta los datos de la factura generada con el esquema exacto.
 * (Cliente SOLO muestra id, nombre, email. No cc, no direccion).
 * @param {object} factura
 */
export function mostrarDetalleFactura(factura) {
  facturaActual = factura;

  const elId = getByHook('detalle-factura-id');
  const elFecha = getByHook('detalle-factura-fecha');
  const elCliId = getByHook('detalle-cliente-id');
  const elCliNombre = getByHook('detalle-cliente-nombre');
  const elCliEmail = getByHook('detalle-cliente-email');
  const elUsuId = getByHook('detalle-usuario-id');
  const elUsuNombre = getByHook('detalle-usuario-nombre');
  const elUsuRol = getByHook('detalle-usuario-rol');
  const elTotal = getByHook('detalle-factura-total');
  const tbody = getByHook('detalle-productos-body');
  const template = getByHook('template-detalle-producto-fila');

  if (elId) elId.textContent = factura.factura_id;
  if (elFecha) elFecha.textContent = factura.fecha;

  if (factura.cliente) {
    if (elCliId) elCliId.textContent = factura.cliente.id;
    if (elCliNombre) elCliNombre.textContent = factura.cliente.nombre || '';
    if (elCliEmail) elCliEmail.textContent = factura.cliente.email || '';
  }

  if (factura.usuario) {
    if (elUsuId) elUsuId.textContent = factura.usuario.id;
    if (elUsuNombre) elUsuNombre.textContent = factura.usuario.nombre || '';
    if (elUsuRol) elUsuRol.textContent = factura.usuario.rol || '';
  }

  if (elTotal) elTotal.textContent = formatearMoneda(factura.total);

  if (tbody && template) {
    tbody.innerHTML = '';
    (factura.productos || []).forEach((prod) => {
      const clone = template.content.cloneNode(true);
      const rowId = clone.querySelector('[data-hook="item-detalle-prod-id"]');
      const rowNombre = clone.querySelector('[data-hook="item-detalle-prod-nombre"]');
      const rowCant = clone.querySelector('[data-hook="item-detalle-prod-cantidad"]');
      const rowPrecio = clone.querySelector('[data-hook="item-detalle-prod-precio"]');
      const rowSubtotal = clone.querySelector('[data-hook="item-detalle-prod-subtotal"]');

      if (rowId) rowId.textContent = prod.id;
      if (rowNombre) rowNombre.textContent = prod.nombre;
      if (rowCant) rowCant.textContent = prod.cantidad;
      if (rowPrecio) rowPrecio.textContent = formatearMoneda(prod.precio_unitario);
      if (rowSubtotal) rowSubtotal.textContent = formatearMoneda(prod.subtotal);

      tbody.appendChild(clone);
    });
  }
}

/**
 * Muestra el modal con el JSON exacto de la factura.
 */
function mostrarModalJSON() {
  if (!facturaActual) {
    mostrarToast('No hay una factura activa para visualizar', 'warning');
    return;
  }
  const jsonPre = getByHook('modal-json-pre');
  if (jsonPre) {
    jsonPre.textContent = JSON.stringify(facturaActual, null, 2);
  }
  if (jsonModal) jsonModal.show();
}
