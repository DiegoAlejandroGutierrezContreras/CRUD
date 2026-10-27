// Módulo de Productos (solo consulta para el catálogo y selección de pedidos)
import { CONFIG } from './config.js';
import { apiFetch } from './api.js';
import { formatearMoneda, getByHook } from './ui.js';

let productosCache = [];

/**
 * Inicializa la carga y visualización de productos.
 */
export function initProductos() {
  cargarProductos();
}

/**
 * Carga el catálogo de productos desde el microservicio.
 */
export async function cargarProductos() {
  const tbody = getByHook('productos-tabla-body');
  const template = getByHook('template-producto-fila');
  if (!tbody || !template) return;

  try {
    const productos = await apiFetch(CONFIG.PRODUCTOS_URL, '/productos');
    productosCache = productos || [];
    renderizarProductos(productosCache);
  } catch (err) {
    // Error notificado por apiFetch
  }
}

/**
 * Renderiza la lista de productos en la tabla.
 * @param {Array} productos
 */
function renderizarProductos(productos) {
  const tbody = getByHook('productos-tabla-body');
  const template = getByHook('template-producto-fila');
  if (!tbody || !template) return;

  tbody.innerHTML = '';

  if (productos.length === 0) {
    const tr = document.createElement('tr');
    tr.innerHTML = `<td colspan="4" class="text-center text-muted py-3">No hay productos en catálogo</td>`;
    tbody.appendChild(tr);
    return;
  }

  productos.forEach((prod) => {
    const clone = template.content.cloneNode(true);

    const elId = clone.querySelector('[data-hook="item-producto-id"]');
    const elNombre = clone.querySelector('[data-hook="item-producto-nombre"]');
    const elPrecio = clone.querySelector('[data-hook="item-producto-precio"]');
    const elStock = clone.querySelector('[data-hook="item-producto-stock"]');

    if (elId) elId.textContent = prod.id;
    if (elNombre) elNombre.textContent = prod.nombre;
    if (elPrecio) elPrecio.textContent = formatearMoneda(prod.precio);
    if (elStock) elStock.textContent = prod.stock;

    tbody.appendChild(clone);
  });
}

/**
 * Retorna los productos en memoria.
 * @returns {Array}
 */
export function getProductosCache() {
  return productosCache;
}
