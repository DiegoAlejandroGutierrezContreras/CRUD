// Módulo de Gestión de Clientes
import { CONFIG } from './config.js';
import { apiFetch } from './api.js';
import { mostrarToast, getByHook, setButtonLoading } from './ui.js';

let clientesCache = [];
let clienteModal = null;

/**
 * Inicializa la vista de clientes.
 */
export function initClientes() {
  const modalEl = document.getElementById('modal-cliente');
  if (modalEl && window.bootstrap) {
    clienteModal = new window.bootstrap.Modal(modalEl);
  }

  const btnNuevo = getByHook('btn-nuevo-cliente');
  if (btnNuevo) {
    btnNuevo.addEventListener('click', () => abrirModalCrear());
  }

  const inputBuscar = getByHook('buscar-cliente-input');
  if (inputBuscar) {
    let debounceTimer;
    inputBuscar.addEventListener('input', (e) => {
      clearTimeout(debounceTimer);
      debounceTimer = setTimeout(() => {
        cargarClientes(e.target.value.trim());
      }, 300);
    });
  }

  const form = getByHook('cliente-form');
  if (form) {
    form.addEventListener('submit', guardarCliente);
  }

  cargarClientes();
}

/**
 * Carga y renderiza el listado de clientes desde el microservicio.
 * @param {string} [query='']
 */
export async function cargarClientes(query = '') {
  const tbody = getByHook('clientes-tabla-body');
  const template = getByHook('template-cliente-fila');
  if (!tbody || !template) return;

  const path = query ? `/clientes?q=${encodeURIComponent(query)}` : '/clientes';

  try {
    const clientes = await apiFetch(CONFIG.CLIENTES_URL, path);
    clientesCache = clientes || [];
    renderizarClientes(clientesCache);
  } catch (err) {
    // Error capturado y mostrado por apiFetch
  }
}

/**
 * Renderiza la lista de clientes usando el template HTML y sus data-hooks.
 * @param {Array} clientes
 */
function renderizarClientes(clientes) {
  const tbody = getByHook('clientes-tabla-body');
  const template = getByHook('template-cliente-fila');
  if (!tbody || !template) return;

  tbody.innerHTML = '';

  if (clientes.length === 0) {
    const tr = document.createElement('tr');
    tr.innerHTML = `<td colspan="6" class="text-center text-muted py-4">No se encontraron clientes</td>`;
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
    const btnEditar = clone.querySelector('[data-hook="btn-editar-cliente"]');
    const btnEliminar = clone.querySelector('[data-hook="btn-eliminar-cliente"]');

    if (elId) elId.textContent = cli.id;
    if (elNombre) elNombre.textContent = cli.nombre;
    if (elCC) elCC.textContent = cli.cc;
    if (elEmail) elEmail.textContent = cli.email;
    if (elDireccion) elDireccion.textContent = cli.direccion;

    if (btnEditar) {
      btnEditar.addEventListener('click', () => abrirModalEditar(cli));
    }
    if (btnEliminar) {
      btnEliminar.addEventListener('click', () => confirmarEliminarCliente(cli));
    }

    tbody.appendChild(clone);
  });
}

function abrirModalCrear() {
  const titulo = getByHook('cliente-modal-titulo');
  const idInput = getByHook('cliente-id-input');
  const nombreInput = getByHook('cliente-nombre-input');
  const ccInput = getByHook('cliente-cc-input');
  const emailInput = getByHook('cliente-email-input');
  const direccionInput = getByHook('cliente-direccion-input');

  if (titulo) titulo.textContent = 'Registrar Nuevo Cliente';
  if (idInput) idInput.value = '';
  if (nombreInput) nombreInput.value = '';
  if (ccInput) ccInput.value = '';
  if (emailInput) emailInput.value = '';
  if (direccionInput) direccionInput.value = '';

  if (clienteModal) clienteModal.show();
}

function abrirModalEditar(cli) {
  const titulo = getByHook('cliente-modal-titulo');
  const idInput = getByHook('cliente-id-input');
  const nombreInput = getByHook('cliente-nombre-input');
  const ccInput = getByHook('cliente-cc-input');
  const emailInput = getByHook('cliente-email-input');
  const direccionInput = getByHook('cliente-direccion-input');

  if (titulo) titulo.textContent = `Editar Cliente #${cli.id}`;
  if (idInput) idInput.value = cli.id;
  if (nombreInput) nombreInput.value = cli.nombre;
  if (ccInput) ccInput.value = cli.cc;
  if (emailInput) emailInput.value = cli.email;
  if (direccionInput) direccionInput.value = cli.direccion;

  if (clienteModal) clienteModal.show();
}

async function guardarCliente(e) {
  e.preventDefault();

  const idInput = getByHook('cliente-id-input');
  const nombreInput = getByHook('cliente-nombre-input');
  const ccInput = getByHook('cliente-cc-input');
  const emailInput = getByHook('cliente-email-input');
  const direccionInput = getByHook('cliente-direccion-input');
  const btnGuardar = getByHook('btn-guardar-cliente');

  const id = idInput ? idInput.value.trim() : '';
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
    if (id) {
      // Actualizar
      await apiFetch(CONFIG.CLIENTES_URL, `/clientes/${id}`, {
        method: 'PUT',
        body: JSON.stringify({ nombre, cc, email, direccion }),
      });
      mostrarToast('Cliente actualizado exitosamente', 'success');
    } else {
      // Crear
      await apiFetch(CONFIG.CLIENTES_URL, '/clientes', {
        method: 'POST',
        body: JSON.stringify({ nombre, cc, email, direccion }),
      });
      mostrarToast('Cliente registrado exitosamente', 'success');
    }

    if (clienteModal) clienteModal.hide();
    const inputBuscar = getByHook('buscar-cliente-input');
    cargarClientes(inputBuscar ? inputBuscar.value.trim() : '');
  } catch (err) {
    // Error capturado y mostrado por apiFetch
  } finally {
    setButtonLoading(btnGuardar, false);
  }
}

async function confirmarEliminarCliente(cli) {
  const confirmacion = window.confirm(`¿Está seguro de eliminar al cliente "${cli.nombre}" (CC: ${cli.cc})?`);
  if (!confirmacion) return;

  try {
    await apiFetch(CONFIG.CLIENTES_URL, `/clientes/${cli.id}`, {
      method: 'DELETE',
    });
    mostrarToast('Cliente eliminado exitosamente', 'success');
    const inputBuscar = getByHook('buscar-cliente-input');
    cargarClientes(inputBuscar ? inputBuscar.value.trim() : '');
  } catch (err) {
    // Error mostrado por apiFetch
  }
}

/**
 * Retorna la lista de clientes en caché (para el selector en nueva factura).
 * @returns {Array}
 */
export function getClientesCache() {
  return clientesCache;
}
