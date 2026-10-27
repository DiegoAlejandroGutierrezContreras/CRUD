// Módulo centralizado de comunicación HTTP (Fetch API)
import { mostrarToast } from './ui.js';

const TOKEN_KEY = 'facturacion_jwt_token';
const USER_KEY = 'facturacion_usuario';

/**
 * Obtiene el token JWT almacenado.
 * @returns {string|null}
 */
export function getToken() {
  return localStorage.getItem(TOKEN_KEY);
}

/**
 * Guarda el token y datos del usuario en localStorage.
 * @param {string} token
 * @param {object} usuario
 */
export function setSession(token, usuario) {
  localStorage.setItem(TOKEN_KEY, token);
  localStorage.setItem(USER_KEY, JSON.stringify(usuario));
}

/**
 * Obtiene los datos del usuario de la sesión actual.
 * @returns {object|null}
 */
export function getUsuario() {
  const data = localStorage.getItem(USER_KEY);
  if (!data) return null;
  try {
    return JSON.parse(data);
  } catch {
    return null;
  }
}

/**
 * Cierra la sesión activa y redirige al login si es necesario.
 * @param {boolean} [redirect=true]
 */
export function clearSession(redirect = true) {
  localStorage.removeItem(TOKEN_KEY);
  localStorage.removeItem(USER_KEY);
  if (redirect && !window.location.pathname.endsWith('index.html') && window.location.pathname !== '/') {
    window.location.href = 'index.html';
  }
}

/**
 * Realiza una petición HTTP con manejo centralizado de headers, autenticación y errores.
 * @param {string} baseUrl
 * @param {string} path
 * @param {RequestInit} [options={}]
 * @returns {Promise<any>}
 */
export async function apiFetch(baseUrl, path, options = {}) {
  const url = `${baseUrl.replace(/\/$/, '')}/${path.replace(/^\//, '')}`;
  const headers = {
    'Content-Type': 'application/json',
    ...(options.headers || {}),
  };

  const token = getToken();
  if (token) {
    headers['Authorization'] = `Bearer ${token}`;
  }

  const fetchOptions = {
    ...options,
    headers,
  };

  let response;
  try {
    response = await fetch(url, fetchOptions);
  } catch (err) {
    const errorMsg = 'Error de conexión: No se pudo contactar al servicio (' + url + ')';
    mostrarToast(errorMsg, 'danger');
    throw new Error(errorMsg);
  }

  // Si responde 401, cerrar sesión automáticamente
  if (response.status === 401) {
    mostrarToast('Sesión expirada o no autorizada. Redirigiendo...', 'warning');
    clearSession(true);
    throw new Error('No autorizado');
  }

  // Manejo de 204 No Content
  if (response.status === 204) {
    return null;
  }

  // Manejar descargas directas de archivos
  const disposition = response.headers.get('Content-Disposition');
  if (disposition && disposition.includes('attachment')) {
    const blob = await response.blob();
    return { blob, filename: extractFilename(disposition) };
  }

  let data;
  try {
    data = await response.json();
  } catch (e) {
    if (!response.ok) {
      const errorMsg = `Error HTTP ${response.status}: ${response.statusText}`;
      mostrarToast(errorMsg, 'danger');
      throw new Error(errorMsg);
    }
    return null;
  }

  if (!response.ok) {
    const errorMsg = data && data.error ? data.error : `Error HTTP ${response.status}`;
    mostrarToast(errorMsg, 'danger');
    const error = new Error(errorMsg);
    error.status = response.status;
    error.data = data;
    throw error;
  }

  return data;
}

function extractFilename(disposition) {
  const match = disposition.match(/filename=(?:["']?)(.*?)(?:["']?)(?:;|$)/);
  return match ? match[1] : 'descarga.json';
}
