// Módulo de utilidades UI: Toasts, formateadores y prevención de XSS

/**
 * Escapa caracteres HTML para prevenir inyección XSS.
 * @param {string} str
 * @returns {string}
 */
export function escapeHTML(str) {
  if (str === null || str === undefined) return '';
  const div = document.createElement('div');
  div.textContent = String(str);
  return div.innerHTML;
}

/**
 * Formatea un valor numérico como moneda en pesos colombianos (COP).
 * @param {number} valor
 * @returns {string}
 */
export function formatearMoneda(valor) {
  const numero = typeof valor === 'number' ? valor : parseFloat(valor) || 0;
  return new Intl.NumberFormat('es-CO', {
    style: 'currency',
    currency: 'COP',
    minimumFractionDigits: 2,
    maximumFractionDigits: 2,
  }).format(numero);
}

/**
 * Muestra un mensaje toast en la interfaz.
 * @param {string} mensaje
 * @param {'success'|'danger'|'warning'|'info'} tipo
 */
export function mostrarToast(mensaje, tipo = 'info') {
  let container = document.querySelector('[data-hook="toast-container"]');
  if (!container) {
    container = document.createElement('div');
    container.setAttribute('data-hook', 'toast-container');
    container.id = 'toast-container';
    document.body.appendChild(container);
  }

  const toastId = 'toast-' + Date.now();
  const alertClass = {
    success: 'bg-success text-white',
    danger: 'bg-danger text-white',
    warning: 'bg-warning text-dark',
    info: 'bg-primary text-white',
  }[tipo] || 'bg-secondary text-white';

  const toastEl = document.createElement('div');
  toastEl.className = `toast align-items-center ${alertClass} border-0 show mb-2`;
  toastEl.setAttribute('role', 'alert');
  toastEl.setAttribute('aria-live', 'assertive');
  toastEl.setAttribute('aria-atomic', 'true');
  toastEl.id = toastId;

  toastEl.innerHTML = `
    <div class="d-flex">
      <div class="toast-body">
        ${escapeHTML(mensaje)}
      </div>
      <button type="button" class="btn-close btn-close-white me-2 m-auto" data-bs-dismiss="toast" aria-label="Close"></button>
    </div>
  `;

  container.appendChild(toastEl);

  // Auto eliminar después de 4 segundos
  setTimeout(() => {
    toastEl.classList.remove('show');
    setTimeout(() => toastEl.remove(), 300);
  }, 4000);

  const btnClose = toastEl.querySelector('.btn-close');
  if (btnClose) {
    btnClose.addEventListener('click', () => toastEl.remove());
  }
}

/**
 * Helper para buscar un elemento por data-hook dentro de un contenedor o document.
 * @param {string} hookName
 * @param {ParentNode} [root=document]
 * @returns {HTMLElement|null}
 */
export function getByHook(hookName, root = document) {
  return root.querySelector(`[data-hook="${hookName}"]`);
}

/**
 * Helper para buscar todos los elementos por data-hook.
 * @param {string} hookName
 * @param {ParentNode} [root=document]
 * @returns {NodeListOf<HTMLElement>}
 */
export function getAllByHook(hookName, root = document) {
  return root.querySelectorAll(`[data-hook="${hookName}"]`);
}

/**
 * Activa o desactiva el estado de carga en un botón.
 * @param {HTMLButtonElement} btn
 * @param {boolean} cargando
 * @param {string} [textoCargando='Cargando...']
 */
export function setButtonLoading(btn, cargando, textoCargando = 'Procesando...') {
  if (!btn) return;
  if (cargando) {
    btn.dataset.originalText = btn.innerHTML;
    btn.disabled = true;
    btn.innerHTML = `<span class="spinner-border spinner-border-sm me-1" role="status" aria-hidden="true"></span>${textoCargando}`;
  } else {
    btn.disabled = false;
    if (btn.dataset.originalText) {
      btn.innerHTML = btn.dataset.originalText;
    }
  }
}
