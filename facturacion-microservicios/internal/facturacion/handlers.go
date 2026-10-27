package facturacion

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net/http"
	"sort"
	"strconv"
	"time"

	"facturacion-microservicios/internal/httpx"
	"facturacion-microservicios/internal/jsondb"
)

// ServiceURLs contiene las URLs de los servicios externos.
type ServiceURLs struct {
	AuthURL      string
	ProductosURL string
	ClientesURL  string
}

// Handler maneja las rutas del servicio de facturación.
type Handler struct {
	store  *jsondb.Store[Factura]
	urls   ServiceURLs
	client *http.Client
}

// NewHandler crea un nuevo handler de facturación.
func NewHandler(store *jsondb.Store[Factura], urls ServiceURLs) *Handler {
	return &Handler{
		store: store,
		urls:  urls,
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// NewHandlerWithClient crea un handler con un cliente HTTP personalizado (para pruebas).
func NewHandlerWithClient(store *jsondb.Store[Factura], urls ServiceURLs, client *http.Client) *Handler {
	return &Handler{
		store:  store,
		urls:   urls,
		client: client,
	}
}

// RegisterRoutes registra las rutas del servicio.
func (h *Handler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("POST /facturas", h.handleCreate)
	mux.HandleFunc("GET /facturas", h.handleList)
	mux.HandleFunc("GET /facturas/{id}", h.handleGet)
	mux.HandleFunc("GET /facturas/{id}/descargar", h.handleDownload)
	mux.HandleFunc("GET /health", h.handleHealth)
}

// handleCreate maneja POST /facturas.
func (h *Handler) handleCreate(w http.ResponseWriter, r *http.Request) {
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		httpx.RespondError(w, http.StatusUnauthorized, "Token de autorización requerido")
		return
	}

	var req CreateFacturaRequest
	if !httpx.DecodeJSON(w, r, &req) {
		return
	}

	// Validar request básico
	if req.ClienteID <= 0 {
		httpx.RespondError(w, http.StatusBadRequest, "El cliente_id es obligatorio y debe ser mayor a 0")
		return
	}
	if len(req.Items) == 0 {
		httpx.RespondError(w, http.StatusBadRequest, "Debe incluir al menos un item en la factura")
		return
	}

	for i, item := range req.Items {
		if item.ProductoID <= 0 {
			httpx.RespondError(w, http.StatusBadRequest, fmt.Sprintf("El producto_id del item %d es inválido", i+1))
			return
		}
		if item.Cantidad <= 0 {
			httpx.RespondError(w, http.StatusBadRequest, fmt.Sprintf("La cantidad del item %d debe ser mayor a 0", i+1))
			return
		}
	}

	// Combinar items del mismo producto
	combinedItems := combineItems(req.Items)

	// 1. Obtener usuario autenticado
	usuario, err := h.getUsuario(authHeader)
	if err != nil {
		handleServiceError(w, err, "autenticación")
		return
	}

	// 2. Obtener cliente
	cliente, err := h.getCliente(req.ClienteID, authHeader)
	if err != nil {
		handleServiceError(w, err, "clientes")
		return
	}

	// 3. Obtener productos y validar stock
	var productosFactura []FacturaItem
	var productosExternos []ProductoExterno

	for _, item := range combinedItems {
		producto, err := h.getProducto(item.ProductoID, authHeader)
		if err != nil {
			handleServiceError(w, err, "productos")
			return
		}

		if producto.Stock < item.Cantidad {
			httpx.RespondError(w, http.StatusUnprocessableEntity,
				fmt.Sprintf("Stock insuficiente para '%s' (ID: %d). Disponible: %d, solicitado: %d",
					producto.Nombre, producto.ID, producto.Stock, item.Cantidad))
			return
		}

		subtotal := roundMoney(float64(item.Cantidad) * producto.Precio)
		productosFactura = append(productosFactura, FacturaItem{
			ID:             producto.ID,
			Nombre:         producto.Nombre,
			Cantidad:       item.Cantidad,
			PrecioUnitario: producto.Precio,
			Subtotal:       subtotal,
		})
		productosExternos = append(productosExternos, *producto)
	}

	// Calcular total
	total := CalculateTotal(productosFactura)

	// 4. Descontar stock de cada producto
	var stockDescontados []struct {
		productoID int
		delta      int
	}

	for i, item := range combinedItems {
		err := h.patchStock(productosExternos[i].ID, -item.Cantidad, authHeader)
		if err != nil {
			// Compensar los descuentos ya hechos
			for _, descontado := range stockDescontados {
				_ = h.patchStock(descontado.productoID, descontado.delta, authHeader)
			}
			httpx.RespondError(w, http.StatusUnprocessableEntity,
				fmt.Sprintf("Error descontando stock del producto '%s' (ID: %d): %v. Se revirtieron los cambios anteriores.",
					productosExternos[i].Nombre, productosExternos[i].ID, err))
			return
		}
		stockDescontados = append(stockDescontados, struct {
			productoID int
			delta      int
		}{productoID: productosExternos[i].ID, delta: item.Cantidad})
	}

	// 5. Crear la factura
	factura := Factura{
		Cliente: FacturaCliente{
			ID:        cliente.ID,
			Nombre:    cliente.Nombre,
			CC:        cliente.CC,
			Email:     cliente.Email,
			Direccion: cliente.Direccion,
		},
		Usuario: FacturaUsuario{
			ID:     usuario.ID,
			Nombre: usuario.Nombre,
			Rol:    usuario.Rol,
		},
		Productos: productosFactura,
		Total:     total,
		Fecha:     time.Now().Format("2006-01-02"),
	}

	factura, err = h.store.Create(factura)
	if err != nil {
		// Compensar stock si falla guardar la factura
		for _, descontado := range stockDescontados {
			_ = h.patchStock(descontado.productoID, descontado.delta, authHeader)
		}
		httpx.RespondError(w, http.StatusInternalServerError, "Error guardando la factura")
		return
	}

	httpx.RespondCreated(w, factura)
}

// handleList maneja GET /facturas.
func (h *Handler) handleList(w http.ResponseWriter, r *http.Request) {
	// Validar token (verificar header pero solo reenviar al servicio de auth si se quiere verificar)
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		httpx.RespondError(w, http.StatusUnauthorized, "Token de autorización requerido")
		return
	}

	// Verificar que el token sea válido consultando al servicio de auth
	_, err := h.getUsuario(authHeader)
	if err != nil {
		handleServiceError(w, err, "autenticación")
		return
	}

	facturas, err := h.store.GetAll()
	if err != nil {
		httpx.RespondError(w, http.StatusInternalServerError, "Error interno del servidor")
		return
	}

	if facturas == nil {
		facturas = []Factura{}
	}

	// Ordenar descendente por factura_id
	sort.Slice(facturas, func(i, j int) bool {
		return facturas[i].FacturaID > facturas[j].FacturaID
	})

	httpx.RespondJSON(w, http.StatusOK, facturas)
}

// handleGet maneja GET /facturas/{id}.
func (h *Handler) handleGet(w http.ResponseWriter, r *http.Request) {
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		httpx.RespondError(w, http.StatusUnauthorized, "Token de autorización requerido")
		return
	}

	_, err := h.getUsuario(authHeader)
	if err != nil {
		handleServiceError(w, err, "autenticación")
		return
	}

	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		httpx.RespondError(w, http.StatusBadRequest, "ID de factura inválido")
		return
	}

	factura, err := h.store.GetByID(id)
	if err != nil {
		httpx.RespondError(w, http.StatusInternalServerError, "Error interno del servidor")
		return
	}
	if factura == nil {
		httpx.RespondError(w, http.StatusNotFound, fmt.Sprintf("Factura con ID %d no encontrada", id))
		return
	}

	httpx.RespondJSON(w, http.StatusOK, factura)
}

// handleDownload maneja GET /facturas/{id}/descargar.
func (h *Handler) handleDownload(w http.ResponseWriter, r *http.Request) {
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		httpx.RespondError(w, http.StatusUnauthorized, "Token de autorización requerido")
		return
	}

	_, err := h.getUsuario(authHeader)
	if err != nil {
		handleServiceError(w, err, "autenticación")
		return
	}

	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		httpx.RespondError(w, http.StatusBadRequest, "ID de factura inválido")
		return
	}

	factura, err := h.store.GetByID(id)
	if err != nil {
		httpx.RespondError(w, http.StatusInternalServerError, "Error interno del servidor")
		return
	}
	if factura == nil {
		httpx.RespondError(w, http.StatusNotFound, fmt.Sprintf("Factura con ID %d no encontrada", id))
		return
	}

	data, err := json.MarshalIndent(factura, "", "  ")
	if err != nil {
		httpx.RespondError(w, http.StatusInternalServerError, "Error generando el archivo")
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=factura-%d.json", id))
	w.WriteHeader(http.StatusOK)
	w.Write(data)
}

// handleHealth maneja GET /health.
func (h *Handler) handleHealth(w http.ResponseWriter, r *http.Request) {
	httpx.RespondJSON(w, http.StatusOK, map[string]string{
		"status":   "ok",
		"servicio": "facturacion",
	})
}

// --- Clientes HTTP para servicios externos ---

type serviceError struct {
	StatusCode int
	Message    string
	IsNetwork  bool
}

func (e *serviceError) Error() string {
	return e.Message
}

func handleServiceError(w http.ResponseWriter, err error, servicio string) {
	se, ok := err.(*serviceError)
	if !ok {
		httpx.RespondError(w, http.StatusInternalServerError, "Error interno del servidor")
		return
	}

	if se.IsNetwork {
		httpx.RespondError(w, http.StatusBadGateway,
			fmt.Sprintf("No se pudo conectar con el servicio de %s. Intente más tarde.", servicio))
		return
	}

	// Mapear códigos del servicio externo
	switch {
	case se.StatusCode == http.StatusNotFound:
		httpx.RespondError(w, http.StatusNotFound, se.Message)
	case se.StatusCode == http.StatusUnauthorized:
		httpx.RespondError(w, http.StatusUnauthorized, se.Message)
	case se.StatusCode == http.StatusUnprocessableEntity:
		httpx.RespondError(w, http.StatusUnprocessableEntity, se.Message)
	default:
		httpx.RespondError(w, http.StatusBadGateway,
			fmt.Sprintf("Error en el servicio de %s: %s", servicio, se.Message))
	}
}

func (h *Handler) doRequest(method, url, authHeader string, body any) ([]byte, int, error) {
	var reqBody io.Reader
	if body != nil {
		data, err := json.Marshal(body)
		if err != nil {
			return nil, 0, &serviceError{Message: "Error interno", IsNetwork: false}
		}
		reqBody = bytes.NewReader(data)
	}

	req, err := http.NewRequest(method, url, reqBody)
	if err != nil {
		return nil, 0, &serviceError{Message: "Error creando solicitud", IsNetwork: true}
	}

	req.Header.Set("Content-Type", "application/json")
	if authHeader != "" {
		req.Header.Set("Authorization", authHeader)
	}

	resp, err := h.client.Do(req)
	if err != nil {
		return nil, 0, &serviceError{Message: "Servicio no disponible", IsNetwork: true}
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, resp.StatusCode, &serviceError{Message: "Error leyendo respuesta", IsNetwork: false}
	}

	return respBody, resp.StatusCode, nil
}

func (h *Handler) getUsuario(authHeader string) (*UsuarioExterno, error) {
	body, status, err := h.doRequest("GET", h.urls.AuthURL+"/auth/me", authHeader, nil)
	if err != nil {
		return nil, err
	}

	if status != http.StatusOK {
		var errResp ErrorExterno
		json.Unmarshal(body, &errResp)
		msg := errResp.Error
		if msg == "" {
			msg = "Error de autenticación"
		}
		return nil, &serviceError{StatusCode: status, Message: msg}
	}

	var usuario UsuarioExterno
	if err := json.Unmarshal(body, &usuario); err != nil {
		return nil, &serviceError{Message: "Error decodificando respuesta del servicio de autenticación"}
	}

	return &usuario, nil
}

func (h *Handler) getCliente(id int, authHeader string) (*ClienteExterno, error) {
	body, status, err := h.doRequest("GET", fmt.Sprintf("%s/clientes/%d", h.urls.ClientesURL, id), authHeader, nil)
	if err != nil {
		return nil, err
	}

	if status != http.StatusOK {
		var errResp ErrorExterno
		json.Unmarshal(body, &errResp)
		msg := errResp.Error
		if msg == "" {
			msg = fmt.Sprintf("Cliente con ID %d no encontrado", id)
		}
		return nil, &serviceError{StatusCode: status, Message: msg}
	}

	var cliente ClienteExterno
	if err := json.Unmarshal(body, &cliente); err != nil {
		return nil, &serviceError{Message: "Error decodificando respuesta del servicio de clientes"}
	}

	return &cliente, nil
}

func (h *Handler) getProducto(id int, authHeader string) (*ProductoExterno, error) {
	body, status, err := h.doRequest("GET", fmt.Sprintf("%s/productos/%d", h.urls.ProductosURL, id), authHeader, nil)
	if err != nil {
		return nil, err
	}

	if status != http.StatusOK {
		var errResp ErrorExterno
		json.Unmarshal(body, &errResp)
		msg := errResp.Error
		if msg == "" {
			msg = fmt.Sprintf("Producto con ID %d no encontrado", id)
		}
		return nil, &serviceError{StatusCode: status, Message: msg}
	}

	var producto ProductoExterno
	if err := json.Unmarshal(body, &producto); err != nil {
		return nil, &serviceError{Message: "Error decodificando respuesta del servicio de productos"}
	}

	return &producto, nil
}

func (h *Handler) patchStock(productoID, delta int, authHeader string) error {
	body, status, err := h.doRequest("PATCH",
		fmt.Sprintf("%s/productos/%d/stock", h.urls.ProductosURL, productoID),
		authHeader,
		StockDeltaRequest{Delta: delta})
	if err != nil {
		return err
	}

	if status != http.StatusOK {
		var errResp ErrorExterno
		json.Unmarshal(body, &errResp)
		msg := errResp.Error
		if msg == "" {
			msg = "Error actualizando stock"
		}
		return &serviceError{StatusCode: status, Message: msg}
	}

	return nil
}

// --- Funciones auxiliares ---

// combineItems combina items con el mismo producto_id sumando cantidades.
func combineItems(items []CreateFacturaItem) []CreateFacturaItem {
	combined := make(map[int]int) // producto_id -> cantidad total
	order := []int{}              // mantener orden de aparición

	for _, item := range items {
		if _, exists := combined[item.ProductoID]; !exists {
			order = append(order, item.ProductoID)
		}
		combined[item.ProductoID] += item.Cantidad
	}

	var result []CreateFacturaItem
	for _, pid := range order {
		result = append(result, CreateFacturaItem{
			ProductoID: pid,
			Cantidad:   combined[pid],
		})
	}

	return result
}

// CalculateTotal calcula el total de una factura sumando los subtotales.
func CalculateTotal(items []FacturaItem) float64 {
	var total float64
	for _, item := range items {
		total += item.Subtotal
	}
	return roundMoney(total)
}

// roundMoney redondea un valor monetario a 2 decimales.
func roundMoney(amount float64) float64 {
	return math.Round(amount*100) / 100
}
