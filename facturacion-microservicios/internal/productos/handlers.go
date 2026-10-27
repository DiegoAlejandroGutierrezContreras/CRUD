package productos

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"facturacion-microservicios/internal/httpx"
	"facturacion-microservicios/internal/jsondb"
)

// Handler maneja las rutas del servicio de productos.
type Handler struct {
	store *jsondb.Store[Producto]
}

// NewHandler crea un nuevo handler de productos.
func NewHandler(store *jsondb.Store[Producto]) *Handler {
	return &Handler{store: store}
}

// RegisterRoutes registra las rutas del servicio.
func (h *Handler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /productos", h.handleList)
	mux.HandleFunc("GET /productos/{id}", h.handleGet)
	mux.HandleFunc("POST /productos", h.handleCreate)
	mux.HandleFunc("PUT /productos/{id}", h.handleUpdate)
	mux.HandleFunc("DELETE /productos/{id}", h.handleDelete)
}

// handleList maneja GET /productos.
func (h *Handler) handleList(w http.ResponseWriter, r *http.Request) {
	items, err := h.store.GetAll()
	if err != nil {
		httpx.RespondError(w, http.StatusInternalServerError, "Error interno del servidor")
		return
	}
	if items == nil {
		items = []Producto{}
	}
	httpx.RespondJSON(w, http.StatusOK, items)
}

// handleGet maneja GET /productos/{id}.
func (h *Handler) handleGet(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		httpx.RespondError(w, http.StatusBadRequest, "ID de producto inválido")
		return
	}

	producto, err := h.store.GetByID(id)
	if err != nil {
		httpx.RespondError(w, http.StatusInternalServerError, "Error interno del servidor")
		return
	}
	if producto == nil {
		httpx.RespondError(w, http.StatusNotFound, fmt.Sprintf("Producto con ID %d no encontrado", id))
		return
	}

	httpx.RespondJSON(w, http.StatusOK, producto)
}

// handleCreate maneja POST /productos.
func (h *Handler) handleCreate(w http.ResponseWriter, r *http.Request) {
	var req CreateRequest
	if !httpx.DecodeJSON(w, r, &req) {
		return
	}

	req.Nombre = strings.TrimSpace(req.Nombre)
	if req.Nombre == "" {
		httpx.RespondError(w, http.StatusBadRequest, "El nombre del producto es obligatorio")
		return
	}
	if req.Precio <= 0 {
		httpx.RespondError(w, http.StatusBadRequest, "El precio debe ser mayor a 0")
		return
	}
	if req.Stock < 0 {
		httpx.RespondError(w, http.StatusBadRequest, "El stock no puede ser negativo")
		return
	}

	producto := Producto{
		Nombre: req.Nombre,
		Precio: req.Precio,
		Stock:  req.Stock,
	}

	producto, err := h.store.Create(producto)
	if err != nil {
		httpx.RespondError(w, http.StatusInternalServerError, "Error interno del servidor")
		return
	}

	httpx.RespondCreated(w, producto)
}

// handleUpdate maneja PUT /productos/{id}.
func (h *Handler) handleUpdate(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		httpx.RespondError(w, http.StatusBadRequest, "ID de producto inválido")
		return
	}

	var req UpdateRequest
	if !httpx.DecodeJSON(w, r, &req) {
		return
	}

	req.Nombre = strings.TrimSpace(req.Nombre)
	if req.Nombre == "" {
		httpx.RespondError(w, http.StatusBadRequest, "El nombre del producto es obligatorio")
		return
	}
	if req.Precio <= 0 {
		httpx.RespondError(w, http.StatusBadRequest, "El precio debe ser mayor a 0")
		return
	}
	if req.Stock < 0 {
		httpx.RespondError(w, http.StatusBadRequest, "El stock no puede ser negativo")
		return
	}

	updated, ok, err := h.store.Update(id, func(p *Producto) {
		p.Nombre = req.Nombre
		p.Precio = req.Precio
		p.Stock = req.Stock
	})
	if err != nil {
		httpx.RespondError(w, http.StatusInternalServerError, "Error interno del servidor")
		return
	}
	if !ok {
		httpx.RespondError(w, http.StatusNotFound, fmt.Sprintf("Producto con ID %d no encontrado", id))
		return
	}

	httpx.RespondJSON(w, http.StatusOK, updated)
}

// handleDelete maneja DELETE /productos/{id}.
func (h *Handler) handleDelete(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		httpx.RespondError(w, http.StatusBadRequest, "ID de producto inválido")
		return
	}

	ok, err := h.store.Delete(id)
	if err != nil {
		httpx.RespondError(w, http.StatusInternalServerError, "Error interno del servidor")
		return
	}
	if !ok {
		httpx.RespondError(w, http.StatusNotFound, fmt.Sprintf("Producto con ID %d no encontrado", id))
		return
	}

	httpx.RespondNoContent(w)
}
