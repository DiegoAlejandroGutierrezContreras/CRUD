package facturacion

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"

	"facturacion-microservicios/internal/jsondb"
)

func newTestFacturaStore(t *testing.T) *jsondb.Store[Factura] {
	t.Helper()
	dir := t.TempDir()
	fp := filepath.Join(dir, "facturas.json")
	store, err := jsondb.New[Factura](fp, 5002,
		func(f *Factura) int { return f.FacturaID },
		func(f *Factura, id int) { f.FacturaID = id })
	if err != nil {
		t.Fatalf("Error creando store de facturas: %v", err)
	}
	return store
}

func TestCombineItems(t *testing.T) {
	items := []CreateFacturaItem{
		{ProductoID: 101, Cantidad: 1},
		{ProductoID: 102, Cantidad: 2},
		{ProductoID: 101, Cantidad: 3},
		{ProductoID: 103, Cantidad: 1},
		{ProductoID: 102, Cantidad: 1},
	}

	combined := combineItems(items)

	if len(combined) != 3 {
		t.Fatalf("Se esperaban 3 items combinados, se obtuvieron %d", len(combined))
	}

	// Verificar orden y cantidades
	expected := map[int]int{101: 4, 102: 3, 103: 1}
	for _, item := range combined {
		exp, ok := expected[item.ProductoID]
		if !ok {
			t.Errorf("Producto inesperado: %d", item.ProductoID)
		}
		if item.Cantidad != exp {
			t.Errorf("Producto %d: se esperaban %d unidades, se obtuvieron %d", item.ProductoID, exp, item.Cantidad)
		}
	}

	// Verificar que se mantiene el orden de aparición
	if combined[0].ProductoID != 101 {
		t.Errorf("Primer producto debería ser 101, se obtuvo %d", combined[0].ProductoID)
	}
	if combined[1].ProductoID != 102 {
		t.Errorf("Segundo producto debería ser 102, se obtuvo %d", combined[1].ProductoID)
	}
	if combined[2].ProductoID != 103 {
		t.Errorf("Tercer producto debería ser 103, se obtuvo %d", combined[2].ProductoID)
	}
}

func TestCalculateTotal(t *testing.T) {
	items := []FacturaItem{
		{Subtotal: 3500000.00},
		{Subtotal: 24000.00},
		{Subtotal: 80000.50},
	}

	total := CalculateTotal(items)
	expected := 3604000.50

	if total != expected {
		t.Errorf("Se esperaba total %.2f, se obtuvo %.2f", expected, total)
	}
}

func TestCalculateTotalVacio(t *testing.T) {
	total := CalculateTotal([]FacturaItem{})
	if total != 0 {
		t.Errorf("Se esperaba total 0, se obtuvo %.2f", total)
	}
}

func TestRoundMoney(t *testing.T) {
	tests := []struct {
		input    float64
		expected float64
	}{
		{3500000.005, 3500000.01},
		{12000.004, 12000.00},
		{0.1 + 0.2, 0.30},
		{99999.999, 100000.00},
	}

	for _, test := range tests {
		result := roundMoney(test.input)
		if result != test.expected {
			t.Errorf("roundMoney(%.10f) = %.2f, esperado %.2f", test.input, result, test.expected)
		}
	}
}

// TestCompensacionAnteError verifica que si falla el descuento de stock de un producto,
// se reviertan los descuentos previos.
func TestCompensacionAnteError(t *testing.T) {
	// Simular servicios externos
	var stockChanges []struct {
		id    int
		delta int
	}

	// Servicio de auth
	authServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(UsuarioExterno{ID: 1, Nombre: "Test", Email: "test@test.com", Rol: "admin"})
	}))
	defer authServer.Close()

	// Servicio de clientes
	clientesServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(ClienteExterno{ID: 201, Nombre: "Test", CC: "123", Email: "c@test.com", Direccion: "Dir"})
	}))
	defer clientesServer.Close()

	// Servicio de productos - el segundo PATCH falla
	patchCount := 0
	productosServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "GET" {
			// Determinar qué producto se pide
			path := r.URL.Path
			if strings.HasSuffix(path, "/101") {
				json.NewEncoder(w).Encode(ProductoExterno{ID: 101, Nombre: "Laptop", Precio: 3500000, Stock: 10})
			} else if strings.HasSuffix(path, "/102") {
				json.NewEncoder(w).Encode(ProductoExterno{ID: 102, Nombre: "Mouse", Precio: 12000, Stock: 5})
			}
			return
		}

		if r.Method == "PATCH" {
			patchCount++
			var req StockDeltaRequest
			json.NewDecoder(r.Body).Decode(&req)

			if patchCount == 2 && req.Delta < 0 {
				// Segundo descuento falla
				w.WriteHeader(http.StatusUnprocessableEntity)
				json.NewEncoder(w).Encode(ErrorExterno{Error: "Stock insuficiente"})
				return
			}

			// Registrar el cambio de stock
			id := 0
			parts := strings.Split(r.URL.Path, "/")
			for i, p := range parts {
				if p == "productos" && i+1 < len(parts) {
					fmt.Sscanf(parts[i+1], "%d", &id)
					break
				}
			}
			stockChanges = append(stockChanges, struct {
				id    int
				delta int
			}{id: id, delta: req.Delta})

			json.NewEncoder(w).Encode(ProductoExterno{ID: id, Nombre: "Test", Precio: 1000, Stock: 5})
			return
		}
	}))
	defer productosServer.Close()

	store := newTestFacturaStore(t)
	handler := NewHandlerWithClient(store, ServiceURLs{
		AuthURL:      authServer.URL,
		ProductosURL: productosServer.URL,
		ClientesURL:  clientesServer.URL,
	}, productosServer.Client())

	// Crear request
	body := `{"cliente_id": 201, "items": [{"producto_id": 101, "cantidad": 1}, {"producto_id": 102, "cantidad": 3}]}`
	req := httptest.NewRequest("POST", "/facturas", strings.NewReader(body))
	req.Header.Set("Authorization", "Bearer test-token")
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.handleCreate(w, req)

	// Verificar que falló
	if w.Code == http.StatusCreated {
		t.Fatal("Se esperaba un error, pero la factura se creó exitosamente")
	}

	// Verificar que se intentó compensar (hay un PATCH con delta positivo después del fallo)
	hasCompensation := false
	for _, change := range stockChanges {
		if change.delta > 0 {
			hasCompensation = true
			break
		}
	}

	if !hasCompensation {
		t.Error("No se encontró compensación de stock después del fallo")
	}

	// Verificar que no se guardó la factura
	facturas, _ := store.GetAll()
	if len(facturas) != 0 {
		t.Errorf("Se esperaban 0 facturas guardadas, se encontraron %d", len(facturas))
	}
}

func TestCrearFacturaExitosa(t *testing.T) {
	// Simular servicios externos
	authServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(UsuarioExterno{ID: 1, Nombre: "Ricardo", Email: "ricardo@test.com", Rol: "admin"})
	}))
	defer authServer.Close()

	clientesServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(ClienteExterno{ID: 201, Nombre: "Carlos", CC: "123", Email: "carlos@test.com", Direccion: "Calle 123"})
	}))
	defer clientesServer.Close()

	productosServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "GET" {
			json.NewEncoder(w).Encode(ProductoExterno{ID: 101, Nombre: "Laptop", Precio: 3500000, Stock: 10})
			return
		}
		if r.Method == "PATCH" {
			json.NewEncoder(w).Encode(ProductoExterno{ID: 101, Nombre: "Laptop", Precio: 3500000, Stock: 9})
		}
	}))
	defer productosServer.Close()

	store := newTestFacturaStore(t)
	handler := NewHandlerWithClient(store, ServiceURLs{
		AuthURL:      authServer.URL,
		ProductosURL: productosServer.URL,
		ClientesURL:  clientesServer.URL,
	}, productosServer.Client())

	body := `{"cliente_id": 201, "items": [{"producto_id": 101, "cantidad": 1}]}`
	req := httptest.NewRequest("POST", "/facturas", strings.NewReader(body))
	req.Header.Set("Authorization", "Bearer test-token")
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.handleCreate(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("Se esperaba 201, se obtuvo %d: %s", w.Code, w.Body.String())
	}

	var factura Factura
	json.NewDecoder(w.Body).Decode(&factura)

	if factura.FacturaID != 5002 {
		t.Errorf("Se esperaba factura_id 5002, se obtuvo %d", factura.FacturaID)
	}
	if factura.Total != 3500000.00 {
		t.Errorf("Se esperaba total 3500000.00, se obtuvo %.2f", factura.Total)
	}
	if factura.Cliente.Nombre != "Carlos" {
		t.Errorf("Se esperaba cliente 'Carlos', se obtuvo '%s'", factura.Cliente.Nombre)
	}
	if factura.Usuario.Nombre != "Ricardo" {
		t.Errorf("Se esperaba usuario 'Ricardo', se obtuvo '%s'", factura.Usuario.Nombre)
	}
	if len(factura.Productos) != 1 {
		t.Fatalf("Se esperaba 1 producto, se obtuvieron %d", len(factura.Productos))
	}
	if factura.Productos[0].PrecioUnitario != 3500000 {
		t.Errorf("Se esperaba precio unitario 3500000, se obtuvo %.2f", factura.Productos[0].PrecioUnitario)
	}
}
