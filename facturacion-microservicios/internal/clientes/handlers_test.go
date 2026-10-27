package clientes

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"

	"facturacion-microservicios/internal/jsondb"
)

func newTestClientesHandler(t *testing.T) (*Handler, *jsondb.Store[Cliente]) {
	t.Helper()
	dir := t.TempDir()
	store, err := jsondb.New[Cliente](filepath.Join(dir, "clientes.json"), 201,
		func(c *Cliente) int { return c.ID },
		func(c *Cliente, id int) { c.ID = id })
	if err != nil {
		t.Fatalf("Error creando store de clientes: %v", err)
	}

	handler := NewHandler(store)
	return handler, store
}

func TestClientesCRUDYValidaciones(t *testing.T) {
	handler, _ := newTestClientesHandler(t)
	mux := http.NewServeMux()
	handler.RegisterRoutes(mux)

	// 1. Crear cliente con CC no numérico -> 400
	badCC := `{"nombre":"Pedro","cc":"abc123","email":"pedro@test.com","direccion":"Calle 1"}`
	reqBadCC := httptest.NewRequest("POST", "/clientes", strings.NewReader(badCC))
	reqBadCC.Header.Set("Content-Type", "application/json")
	wBadCC := httptest.NewRecorder()
	mux.ServeHTTP(wBadCC, reqBadCC)
	if wBadCC.Code != http.StatusBadRequest {
		t.Errorf("Se esperaba 400 para CC no numérico, se obtuvo %d", wBadCC.Code)
	}

	// 2. Crear cliente válido
	validBody := `{"nombre":"Pedro Gomez","cc":"987654321","email":"pedro@test.com","direccion":"Calle 1 #2-3, Cali"}`
	reqCreate := httptest.NewRequest("POST", "/clientes", strings.NewReader(validBody))
	reqCreate.Header.Set("Content-Type", "application/json")
	wCreate := httptest.NewRecorder()
	mux.ServeHTTP(wCreate, reqCreate)
	if wCreate.Code != http.StatusCreated {
		t.Fatalf("Se esperaba 201 en crear cliente, se obtuvo %d: %s", wCreate.Code, wCreate.Body.String())
	}

	var cliente Cliente
	json.NewDecoder(wCreate.Body).Decode(&cliente)
	if cliente.ID != 201 || cliente.CC != "987654321" {
		t.Fatalf("Cliente creado con datos incorrectos: %+v", cliente)
	}

	// 3. Crear cliente con CC duplicado -> 409
	dupCC := `{"nombre":"Otro Nombre","cc":"987654321","email":"otro@test.com","direccion":"Calle 2"}`
	reqDupCC := httptest.NewRequest("POST", "/clientes", strings.NewReader(dupCC))
	reqDupCC.Header.Set("Content-Type", "application/json")
	wDupCC := httptest.NewRecorder()
	mux.ServeHTTP(wDupCC, reqDupCC)
	if wDupCC.Code != http.StatusConflict {
		t.Errorf("Se esperaba 409 para CC duplicado, se obtuvo %d", wDupCC.Code)
	}

	// 4. Crear cliente con Email duplicado -> 409
	dupEmail := `{"nombre":"Otro Nombre","cc":"11223344","email":"pedro@test.com","direccion":"Calle 2"}`
	reqDupEmail := httptest.NewRequest("POST", "/clientes", strings.NewReader(dupEmail))
	reqDupEmail.Header.Set("Content-Type", "application/json")
	wDupEmail := httptest.NewRecorder()
	mux.ServeHTTP(wDupEmail, reqDupEmail)
	if wDupEmail.Code != http.StatusConflict {
		t.Errorf("Se esperaba 409 para Email duplicado, se obtuvo %d", wDupEmail.Code)
	}

	// 5. Buscar por query ?q=pedro
	reqSearch := httptest.NewRequest("GET", "/clientes?q=pedro", nil)
	wSearch := httptest.NewRecorder()
	mux.ServeHTTP(wSearch, reqSearch)
	var listResp []Cliente
	json.NewDecoder(wSearch.Body).Decode(&listResp)
	if len(listResp) != 1 {
		t.Errorf("Se esperaba 1 resultado, se obtuvieron %d", len(listResp))
	}

	// 6. Actualizar cliente
	upBody := `{"nombre":"Pedro Gomez Actualizado","cc":"987654321","email":"pedro_nuevo@test.com","direccion":"Nueva Dir"}`
	reqUp := httptest.NewRequest("PUT", "/clientes/201", strings.NewReader(upBody))
	reqUp.Header.Set("Content-Type", "application/json")
	wUp := httptest.NewRecorder()
	mux.ServeHTTP(wUp, reqUp)
	if wUp.Code != http.StatusOK {
		t.Errorf("Se esperaba 200 en update cliente, se obtuvo %d", wUp.Code)
	}

	// 7. Eliminar cliente
	reqDel := httptest.NewRequest("DELETE", "/clientes/201", nil)
	wDel := httptest.NewRecorder()
	mux.ServeHTTP(wDel, reqDel)
	if wDel.Code != http.StatusNoContent {
		t.Errorf("Se esperaba 204 en delete cliente, se obtuvo %d", wDel.Code)
	}
}
