package productos

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"

	"facturacion-microservicios/internal/jsondb"
)

func newTestProductosHandler(t *testing.T) (*Handler, *jsondb.Store[Producto]) {
	t.Helper()
	dir := t.TempDir()
	store, err := jsondb.New[Producto](filepath.Join(dir, "productos.json"), 101,
		func(p *Producto) int { return p.ID },
		func(p *Producto, id int) { p.ID = id })
	if err != nil {
		t.Fatalf("Error creando store de productos: %v", err)
	}

	handler := NewHandler(store)
	return handler, store
}

func TestProductosCRUDYStock(t *testing.T) {
	handler, _ := newTestProductosHandler(t)
	mux := http.NewServeMux()
	handler.RegisterRoutes(mux)

	// 1. Crear producto
	createBody := `{"nombre":"Teclado Mecánico","precio":150000.0,"stock":10}`
	reqCreate := httptest.NewRequest("POST", "/productos", strings.NewReader(createBody))
	reqCreate.Header.Set("Content-Type", "application/json")
	wCreate := httptest.NewRecorder()
	mux.ServeHTTP(wCreate, reqCreate)

	if wCreate.Code != http.StatusCreated {
		t.Fatalf("Se esperaba 201 en crear producto, se obtuvo %d: %s", wCreate.Code, wCreate.Body.String())
	}

	var prod Producto
	json.NewDecoder(wCreate.Body).Decode(&prod)
	if prod.ID != 101 || prod.Stock != 10 {
		t.Fatalf("Producto creado con datos incorrectos: %+v", prod)
	}

	// 2. Obtener producto por ID
	reqGet := httptest.NewRequest("GET", "/productos/101", nil)
	wGet := httptest.NewRecorder()
	mux.ServeHTTP(wGet, reqGet)
	if wGet.Code != http.StatusOK {
		t.Errorf("Se esperaba 200 en get producto, se obtuvo %d", wGet.Code)
	}

	// 3. Listar con filtro ?q=teclado
	reqList := httptest.NewRequest("GET", "/productos?q=teclado", nil)
	wList := httptest.NewRecorder()
	mux.ServeHTTP(wList, reqList)
	var listResp []Producto
	json.NewDecoder(wList.Body).Decode(&listResp)
	if len(listResp) != 1 {
		t.Errorf("Se esperaba 1 producto en búsqueda, se obtuvieron %d", len(listResp))
	}

	// 4. Actualizar producto
	upBody := `{"nombre":"Teclado Mecánico RGB","precio":180000.0,"stock":12}`
	reqUp := httptest.NewRequest("PUT", "/productos/101", strings.NewReader(upBody))
	reqUp.Header.Set("Content-Type", "application/json")
	wUp := httptest.NewRecorder()
	mux.ServeHTTP(wUp, reqUp)
	if wUp.Code != http.StatusOK {
		t.Errorf("Se esperaba 200 en update, se obtuvo %d", wUp.Code)
	}

	// 5. PATCH stock - descontar 5 unidades (delta: -5)
	patchBody := `{"delta": -5}`
	reqPatch := httptest.NewRequest("PATCH", "/productos/101/stock", strings.NewReader(patchBody))
	reqPatch.Header.Set("Content-Type", "application/json")
	wPatch := httptest.NewRecorder()
	mux.ServeHTTP(wPatch, reqPatch)
	if wPatch.Code != http.StatusOK {
		t.Fatalf("Se esperaba 200 en patch stock, se obtuvo %d: %s", wPatch.Code, wPatch.Body.String())
	}
	var patchedProd Producto
	json.NewDecoder(wPatch.Body).Decode(&patchedProd)
	if patchedProd.Stock != 7 {
		t.Errorf("Se esperaba stock 7 tras descontar 5 de 12, se obtuvo %d", patchedProd.Stock)
	}

	// 6. PATCH stock - intentar descontar más del disponible (delta: -10 cuando hay 7) -> 422
	patchFailBody := `{"delta": -10}`
	reqFailPatch := httptest.NewRequest("PATCH", "/productos/101/stock", strings.NewReader(patchFailBody))
	reqFailPatch.Header.Set("Content-Type", "application/json")
	wFailPatch := httptest.NewRecorder()
	mux.ServeHTTP(wFailPatch, reqFailPatch)
	if wFailPatch.Code != http.StatusUnprocessableEntity {
		t.Errorf("Se esperaba 422 para stock insuficiente, se obtuvo %d", wFailPatch.Code)
	}

	// 7. Eliminar producto
	reqDel := httptest.NewRequest("DELETE", "/productos/101", nil)
	wDel := httptest.NewRecorder()
	mux.ServeHTTP(wDel, reqDel)
	if wDel.Code != http.StatusNoContent {
		t.Errorf("Se esperaba 204 en delete, se obtuvo %d", wDel.Code)
	}
}
