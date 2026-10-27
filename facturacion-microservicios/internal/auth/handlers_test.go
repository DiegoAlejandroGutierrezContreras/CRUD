package auth

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"

	"facturacion-microservicios/internal/jsondb"
)

func newTestAuthHandler(t *testing.T) (*Handler, *jsondb.Store[Usuario]) {
	t.Helper()
	dir := t.TempDir()
	store, err := jsondb.New[Usuario](filepath.Join(dir, "usuarios.json"), 1,
		func(u *Usuario) int { return u.ID },
		func(u *Usuario, id int) { u.ID = id })
	if err != nil {
		t.Fatalf("Error creando store de usuarios: %v", err)
	}

	handler := NewHandler(store, "secret-test-key")
	return handler, store
}

func TestRegisterYLogin(t *testing.T) {
	handler, _ := newTestAuthHandler(t)

	// 1. Registro exitoso
	regBody := `{"nombre":"Laura Gómez","email":"laura@example.com","password":"secretpassword"}`
	reqReg := httptest.NewRequest("POST", "/auth/register", strings.NewReader(regBody))
	reqReg.Header.Set("Content-Type", "application/json")
	wReg := httptest.NewRecorder()

	handler.handleRegister(wReg, reqReg)

	if wReg.Code != http.StatusCreated {
		t.Fatalf("Se esperaba 201 en registro, se obtuvo %d: %s", wReg.Code, wReg.Body.String())
	}

	var regResp UsuarioSinPassword
	json.NewDecoder(wReg.Body).Decode(&regResp)
	if regResp.Nombre != "Laura Gómez" || regResp.Email != "laura@example.com" || regResp.Rol != "cliente" {
		t.Errorf("Datos de usuario registrado incorrectos: %+v", regResp)
	}

	// 2. Registro con email duplicado
	wDup := httptest.NewRecorder()
	reqDup := httptest.NewRequest("POST", "/auth/register", strings.NewReader(regBody))
	reqDup.Header.Set("Content-Type", "application/json")
	handler.handleRegister(wDup, reqDup)
	if wDup.Code != http.StatusConflict {
		t.Errorf("Se esperaba 409 para email duplicado, se obtuvo %d", wDup.Code)
	}

	// 3. Login exitoso
	loginBody := `{"email":"laura@example.com","password":"secretpassword"}`
	reqLogin := httptest.NewRequest("POST", "/auth/login", strings.NewReader(loginBody))
	reqLogin.Header.Set("Content-Type", "application/json")
	wLogin := httptest.NewRecorder()

	handler.handleLogin(wLogin, reqLogin)
	if wLogin.Code != http.StatusOK {
		t.Fatalf("Se esperaba 200 en login, se obtuvo %d: %s", wLogin.Code, wLogin.Body.String())
	}

	var loginResp LoginResponse
	json.NewDecoder(wLogin.Body).Decode(&loginResp)
	if loginResp.Token == "" {
		t.Fatal("Token no retornado en respuesta de login")
	}

	// 4. Login con contraseña errónea
	badLogin := `{"email":"laura@example.com","password":"wrongpassword"}`
	reqBad := httptest.NewRequest("POST", "/auth/login", strings.NewReader(badLogin))
	reqBad.Header.Set("Content-Type", "application/json")
	wBad := httptest.NewRecorder()
	handler.handleLogin(wBad, reqBad)
	if wBad.Code != http.StatusUnauthorized {
		t.Errorf("Se esperaba 401 con contraseña incorrecta, se obtuvo %d", wBad.Code)
	}

	// 5. GET /auth/me con token generado
	reqMe := httptest.NewRequest("GET", "/auth/me", nil)
	reqMe.Header.Set("Authorization", "Bearer "+loginResp.Token)
	wMe := httptest.NewRecorder()
	handler.handleMe(wMe, reqMe)
	if wMe.Code != http.StatusOK {
		t.Fatalf("Se esperaba 200 en /auth/me, se obtuvo %d: %s", wMe.Code, wMe.Body.String())
	}

	var meResp UsuarioSinPassword
	json.NewDecoder(wMe.Body).Decode(&meResp)
	if meResp.Email != "laura@example.com" {
		t.Errorf("Email en /me incorrecto: %s", meResp.Email)
	}
}

func TestHealth(t *testing.T) {
	handler, _ := newTestAuthHandler(t)
	req := httptest.NewRequest("GET", "/health", nil)
	w := httptest.NewRecorder()
	handler.handleHealth(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("Se esperaba 200 en health, se obtuvo %d", w.Code)
	}
}
