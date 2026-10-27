// Módulo de Gestión de Productos
import { CONFIG } from './config.js';
import { apiFetch } from './api.js';
import { mostrarToast, formatearMoneda, getByHook, setButtonLoading } from './ui.js';

let productosCache = [];
let productoModal = null;

/**
 * Inicializa la vista de productos.
 */
export function initProductos() {
  const modalEl = document.getElementById('modal-producto');
  if (modalEl && window.bootstrap) {
    productoModal = new window.bootstrap.Modal(modalEl);
  }

  const btnNuevo = getByHook('btn-nuevo-producto');
  if (btnNuevo) {
    btnNuevo.addEventListener('click', () => abrirModalCrear());
  }

  const inputBuscar = getByHook('buscar-producto-input');
  if (inputBuscar) {
    let debounceTimer;
    inputBuscar.addEventListener('input', (e) => {
      clearTimeout(debounceTimer);
      debounceTimer = setTimeout(() => {
        cargarProductos(e.target.value.trim());
      }, 300);
    });
  }

  const form = getByHook('producto-form');
  if (form) {
    form.addEventListener('submit', guardarProducto);
  }

  cargarProductos();
}

/**
 * Carga y renderiza el listado de productos desde el microservicio.
 * @param {string} [query='']
 */
export async function cargarProductos(query = '') {
  const tbody = getByHook('productos-tabla-body');
  const template = getByHook('template-producto-fila');
  if (!tbody || !template) return;

  const path = query ? `/productos?q=${encodeURIComponent(query)}` : '/productos';

  try {
    const productos = await apiFetch(CONFIG.PRODUCTOS_URL, path);
    productosCache = productos || [];
    renderizarProductos(productosCache);
  } catch (err) {
    // Error mostrado por apiFetch
  }
}

/**
 * Renderiza la lista de productos usando el template HTML y sus data-hooks.
 * @param {Array} productos
 */
function renderizarProductos(productos) {
  const tbody = getByHook('productos-tabla-body');
  const template = getByHook('template-producto-fila');
  if (!tbody || !template) return;

  tbody.innerHTML = '';

  if (productos.length === 0) {
    const tr = document.createElement('tr');
    tr.innerHTML = `<td colspan="5" class="text-center text-muted py-4">No se encontraron productos</td>`;
    tbody.appendChild(tr);
    return;
  }

  productos.forEach((prod) => {
    const clone = template.content.cloneNode(true);

    const elId = clone.querySelector('[data-hook="item-producto-id"]');
    const elNombre = clone.querySelector('[data-hook="item-producto-nombre"]');
    const elPrecio = clone.querySelector('[data-hook="item-producto-precio"]');
    const elStock = clone.querySelector('[data-hook="item-producto-stock"]');
    const btnEditar = clone.querySelector('[data-hook="btn-editar-producto"]');
    const btnEliminar = clone.querySelector('[data-hook="btn-eliminar-producto"]');

    if (elId) elId.textContent = prod.id;
    if (elNombre) elNombre.textContent = prod.nombre;
    if (elPrecio) elPrecio.textContent = formatearMoneda(prod.precio);
    if (elStock) {
      elStock.textContent = prod.stock;
      if (prod.stock <= 0) {
        elStock.className += ' text-danger fw-bold';
      }
    }

    if (btnEditar) {
      btnEditar.addEventListener('click', () => abrirModalEditar(prod));
    }
    if (btnEliminar) {
      btnEliminar.addEventListener('click', () => confirmarEliminarProducto(prod));
    }

    tbody.appendChild(clone);
  });
}

function abrirModalCrear() {
  const titulo = getByHook('producto-modal-titulo');
  const idInput = getByHook('producto-id-input');
  const nombreInput = getByHook('producto-nombre-input');
  const precioInput = getByHook('producto-precio-input');
  const stockInput = getByHook('producto-stock-input');

  if (titulo) titulo.textContent = 'Nuevo Producto';
  if (idInput) idInput.value = '';
  if (nombreInput) nombreInput.value = '';
  if (precioInput) precioInput.value = '';
  if (stockInput) stockInput.value = '';

  if (productoModal) productoModal.show();
}

function abrirModalEditar(prod) {
  const titulo = getByHook('producto-modal-titulo');
  const idInput = getByHook('producto-id-input');
  const nombreInput = getByHook('producto-nombre-input');
  const precioInput = getByHook('producto-precio-input');
  const stockInput = getByHook('producto-stock-input');

  if (titulo) titulo.textContent = `Editar Producto #${prod.id}`;
  if (idInput) idInput.value = prod.id;
  if (nombreInput) nombreInput.value = prod.nombre;
  if (precioInput) precioInput.value = prod.precio;
  if (stockInput) stockInput.value = prod.stock;

  if (productoModal) productoModal.show();
}

async function guardarProducto(e) {
  e.preventDefault();

  const idInput = getByHook('producto-id-input');
  const nombreInput = getByHook('producto-nombre-input');
  const precioInput = getByHook('producto-precio-input');
  const stockInput = getByHook('producto-stock-input');
  const btnGuardar = getByHook('btn-guardar-producto');

  const id = idInput ? idInput.value.trim() : '';
  const nombre = nombreInput ? nombreInput.value.trim() : '';
  const precio = precioInput ? parseFloat(precioInput.value) : 0;
  const stock = stockInput ? parseInt(stockInput.value, 10) : 0;

  if (!nombre) {
    mostrarToast('El nombre del producto es obligatorio', 'warning');
    return;
  }
  if (isNaN(precio) || precio <= 0) {
    mostrarToast('El precio debe ser un número mayor a 0', 'warning');
    return;
  }
  if (isNaN(stock) || stock < 0) {
    mostrarToast('El stock no puede ser un número negativo', 'warning');
    return;
  }

  setButtonLoading(btnGuardar, true, 'Guardando...');

  try {
    if (id) {
      // Actualizar
      await apiFetch(CONFIG.PRODUCTOS_URL, `/productos/${id}`, {
        method: 'PUT',
        body: JSON.stringify({ nombre, precio, stock }),
      });
      mostrarToast('Producto actualizado exitosamente', 'success');
    } else {
      // Crear
      await apiFetch(CONFIG.PRODUCTOS_URL, '/productos', {
        method: 'POST',
        body: JSON.stringify({ nombre, precio, stock }),
      });
      mostrarToast('Producto creado exitosamente', 'success');
    }

    if (productoModal) productoModal.hide();
    const inputBuscar = getByHook('buscar-producto-input');
    cargarProductos(inputBuscar ? inputBuscar.value.trim() : '');
  } catch (err) {
    // Error capturado y mostrado por apiFetch
  } finally {
    setButtonLoading(btnGuardar, false);
  }
}

async function confirmarEliminarProducto(prod) {
  const confirmacion = window.confirm(`¿Está seguro de eliminar el producto "${prod.nombre}" (ID: ${prod.id})?`);
  if (!confirmacion) return;

  try {
    await apiFetch(CONFIG.PRODUCTOS_URL, `/productos/${prod.id}`, {
      method: 'DELETE',
    });
    mostrarToast('Producto eliminado exitosamente', 'success');
    const inputBuscar = getByHook('buscar-producto-input');
    cargarProductos(inputBuscar ? inputBuscar.value.trim() : '');
  } catch (err) {
    // Error mostrado por apiFetch
  }
}

/**
 * Retorna la lista de productos en caché (para el carrito de compras).
 * @returns {Array}
 */
export function getProductosCache() {
  return productosCache;
}
