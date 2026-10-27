// Módulo de Autenticación en el Frontend
import { CONFIG } from './config.js';
import { apiFetch, setSession, clearSession, getToken, getUsuario } from './api.js';
import { mostrarToast, getByHook, setButtonLoading } from './ui.js';

/**
 * Inicializa la lógica de la página de Login y Registro (index.html).
 */
export function initAuthPage() {
  // Si ya tiene sesión activa, redirigir a la app principal
  if (getToken()) {
    window.location.href = 'app.html';
    return;
  }

  const loginForm = getByHook('login-form');
  const registerForm = getByHook('register-form');

  if (loginForm) {
    loginForm.addEventListener('submit', async (e) => {
      e.preventDefault();
      const emailInput = getByHook('login-email');
      const passwordInput = getByHook('login-password');
      const submitBtn = getByHook('btn-login-submit');

      const email = emailInput ? emailInput.value.trim() : '';
      const password = passwordInput ? passwordInput.value : '';

      if (!email || !password) {
        mostrarToast('Por favor ingrese email y contraseña', 'warning');
        return;
      }

      setButtonLoading(submitBtn, true, 'Iniciando sesión...');

      try {
        const resp = await apiFetch(CONFIG.AUTH_URL, '/auth/login', {
          method: 'POST',
          body: JSON.stringify({ email, password }),
        });

        if (resp && resp.token) {
          setSession(resp.token, resp.usuario);
          mostrarToast(`¡Bienvenido, ${resp.usuario.nombre}!`, 'success');
          setTimeout(() => {
            window.location.href = 'app.html';
          }, 500);
        }
      } catch (err) {
        // Error ya mostrado por apiFetch en toast
      } finally {
        setButtonLoading(submitBtn, false);
      }
    });
  }

  if (registerForm) {
    registerForm.addEventListener('submit', async (e) => {
      e.preventDefault();
      const nombreInput = getByHook('register-nombre');
      const emailInput = getByHook('register-email');
      const passwordInput = getByHook('register-password');
      const submitBtn = getByHook('btn-register-submit');

      const nombre = nombreInput ? nombreInput.value.trim() : '';
      const email = emailInput ? emailInput.value.trim() : '';
      const password = passwordInput ? passwordInput.value : '';

      if (!nombre || !email || !password) {
        mostrarToast('Todos los campos son obligatorios', 'warning');
        return;
      }

      if (password.length < 6) {
        mostrarToast('La contraseña debe tener al menos 6 caracteres', 'warning');
        return;
      }

      setButtonLoading(submitBtn, true, 'Registrando...');

      try {
        const resp = await apiFetch(CONFIG.AUTH_URL, '/auth/register', {
          method: 'POST',
          body: JSON.stringify({ nombre, email, password }),
        });

        if (resp && resp.id) {
          mostrarToast('Registro exitoso. Ahora puede iniciar sesión con sus credenciales.', 'success');
          registerForm.reset();
          // Cambiar a la pestaña de login si existe botón de tab
          const loginTabBtn = getByHook('tab-login-btn');
          if (loginTabBtn) {
            loginTabBtn.click();
          }
        }
      } catch (err) {
        // Error ya mostrado por apiFetch
      } finally {
        setButtonLoading(submitBtn, false);
      }
    });
  }
}

/**
 * Inicializa el guardia de sesión y controles de usuario en la aplicación principal (app.html).
 */
export function initAppSession() {
  const token = getToken();
  if (!token) {
    window.location.href = 'index.html';
    return null;
  }

  const usuario = getUsuario();
  if (!usuario) {
    clearSession(true);
    return null;
  }

  // Pintar datos del usuario en la barra de navegación
  const nombreEl = getByHook('nav-user-nombre');
  const rolEl = getByHook('nav-user-rol');

  if (nombreEl) nombreEl.textContent = usuario.nombre || '';
  if (rolEl) rolEl.textContent = `(${usuario.rol || 'cliente'})`;

  // Botón de cerrar sesión
  const btnLogout = getByHook('btn-logout');
  if (btnLogout) {
    btnLogout.addEventListener('click', (e) => {
      e.preventDefault();
      clearSession(true);
    });
  }

  return usuario;
}
