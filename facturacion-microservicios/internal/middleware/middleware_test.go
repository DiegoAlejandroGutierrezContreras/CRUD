package middleware

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func TestCORS(t *testing.T) {
	handler := CORS(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	// Test OPTIONS preflight
	req := httptest.NewRequest("OPTIONS", "/test", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusNoContent {
		t.Errorf("Se esperaba status 204 para OPTIONS, se obtuvo %d", w.Code)
	}
	if w.Header().Get("Access-Control-Allow-Origin") != "*" {
		t.Errorf("CORS header incorrecto: %s", w.Header().Get("Access-Control-Allow-Origin"))
	}

	// Test regular GET
	reqGet := httptest.NewRequest("GET", "/test", nil)
	wGet := httptest.NewRecorder()
	handler.ServeHTTP(wGet, reqGet)

	if wGet.Code != http.StatusOK {
		t.Errorf("Se esperaba status 200, se obtuvo %d", wGet.Code)
	}
}

func TestRecover(t *testing.T) {
	panicHandler := Recover(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		panic("error inesperado")
	}))

	req := httptest.NewRequest("GET", "/panic", nil)
	w := httptest.NewRecorder()

	panicHandler.ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("Se esperaba 500 tras panic, se obtuvo %d", w.Code)
	}
}

func TestRequireJWT(t *testing.T) {
	secret := "test-secret-key"

	protectedHandler := RequireJWT(secret)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		claims := GetClaims(r)
		if claims == nil {
			t.Fatal("Claims no encontrados en el contexto")
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(claims.Nombre))
	}))

	// 1. Sin header
	reqNoHeader := httptest.NewRequest("GET", "/protected", nil)
	wNoHeader := httptest.NewRecorder()
	protectedHandler.ServeHTTP(wNoHeader, reqNoHeader)
	if wNoHeader.Code != http.StatusUnauthorized {
		t.Errorf("Se esperaba 401 sin header, se obtuvo %d", wNoHeader.Code)
	}

	// 2. Token inválido
	reqInvalid := httptest.NewRequest("GET", "/protected", nil)
	reqInvalid.Header.Set("Authorization", "Bearer invalid-token-str")
	wInvalid := httptest.NewRecorder()
	protectedHandler.ServeHTTP(wInvalid, reqInvalid)
	if wInvalid.Code != http.StatusUnauthorized {
		t.Errorf("Se esperaba 401 para token inválido, se obtuvo %d", wInvalid.Code)
	}

	// 3. Token válido
	tokenClaims := JWTClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   "1",
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(1 * time.Hour)),
		},
		Nombre: "Admin User",
		Email:  "admin@test.com",
		Rol:    "admin",
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, tokenClaims)
	tokenStr, _ := token.SignedString([]byte(secret))

	reqValid := httptest.NewRequest("GET", "/protected", nil)
	reqValid.Header.Set("Authorization", fmt.Sprintf("Bearer %s", tokenStr))
	wValid := httptest.NewRecorder()
	protectedHandler.ServeHTTP(wValid, reqValid)

	if wValid.Code != http.StatusOK {
		t.Errorf("Se esperaba 200 con token válido, se obtuvo %d", wValid.Code)
	}
	if wValid.Body.String() != "Admin User" {
		t.Errorf("Respuesta incorrecta: %s", wValid.Body.String())
	}
}
