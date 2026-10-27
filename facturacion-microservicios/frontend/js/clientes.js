// Módulo de Clientes (Registro y consulta para facturación)
import { CONFIG } from './config.js';
import { apiFetch } from './api.js';
import { mostrarToast, getByHook, setButtonLoading } from './ui.js';
import { poblarSelectoresNuevaFactura } from './facturas.js';

let clientesCache = [];

/**
 * Inicializa la lógica de clientes.
 */
export function initClientes() {
  const form = getByHook('cliente-form');
  if (form) {
    form.addEventListener('submit', guardarCliente);
  }

  cargarClientes();
}

/**
 * Carga el listado de clientes desde el microservicio.
 */
export async function cargarClientes() {
  const tbody = getByHook('clientes-tabla-body');
  const template = getByHook('template-cliente-fila');
  if (!tbody || !template) return;

  try {
    const clientes = await apiFetch(CONFIG.CLIENTES_URL, '/clientes');
    clientesCache = clientes || [];
    renderizarClientes(clientesCache);
  } catch (err) {
    // Error mostrado por apiFetch
  }
}

/**
 * Renderiza la lista de clientes en la tabla.
 * @param {Array} clientes
 */
function renderizarClientes(clientes) {
  const tbody = getByHook('clientes-tabla-body');
  const template = getByHook('template-cliente-fila');
  if (!tbody || !template) return;

  tbody.innerHTML = '';

  if (clientes.length === 0) {
    const tr = document.createElement('tr');
    tr.innerHTML = `<td colspan="5" class="text-center text-muted py-3">No hay clientes registrados</td>`;
    tbody.appendChild(tr);
    return;
  }

  clientes.forEach((cli) => {
    const clone = template.content.cloneNode(true);

    const elId = clone.querySelector('[data-hook="item-cliente-id"]');
    const elNombre = clone.querySelector('[data-hook="item-cliente-nombre"]');
    const elCC = clone.querySelector('[data-hook="item-cliente-cc"]');
    const elEmail = clone.querySelector('[data-hook="item-cliente-email"]');
    const elDireccion = clone.querySelector('[data-hook="item-cliente-direccion"]');

    if (elId) elId.textContent = cli.id;
    if (elNombre) elNombre.textContent = cli.nombre;
    if (elCC) elCC.textContent = cli.cc;
    if (elEmail) elEmail.textContent = cli.email;
    if (elDireccion) elDireccion.textContent = cli.direccion;

    tbody.appendChild(clone);
  });
}

/**
 * Registra un nuevo cliente mediante POST /clientes.
 * @param {Event} e
 */
async function guardarCliente(e) {
  e.preventDefault();

  const nombreInput = getByHook('cliente-nombre-input');
  const ccInput = getByHook('cliente-cc-input');
  const emailInput = getByHook('cliente-email-input');
  const direccionInput = getByHook('cliente-direccion-input');
  const btnGuardar = getByHook('btn-guardar-cliente');

  const nombre = nombreInput ? nombreInput.value.trim() : '';
  const cc = ccInput ? ccInput.value.trim() : '';
  const email = emailInput ? emailInput.value.trim() : '';
  const direccion = direccionInput ? direccionInput.value.trim() : '';

  if (!nombre || !cc || !email || !direccion) {
    mostrarToast('Todos los campos son obligatorios', 'warning');
    return;
  }

  setButtonLoading(btnGuardar, true, 'Guardando...');

  try {
    await apiFetch(CONFIG.CLIENTES_URL, '/clientes', {
      method: 'POST',
      body: JSON.stringify({ nombre, cc, email, direccion }),
    });

    mostrarToast('Cliente registrado exitosamente', 'success');

    // Limpiar formulario
    const form = getByHook('cliente-form');
    if (form) form.reset();

    // Recargar lista de clientes y selector de factura
    await cargarClientes();
    await poblarSelectoresNuevaFactura();
  } catch (err) {
    // Error capturado y mostrado por apiFetch
  } finally {
    setButtonLoading(btnGuardar, false);
  }
}

/**
 * Retorna los clientes en memoria.
 * @returns {Array}
 */
export function getClientesCache() {
  return clientesCache;
}
