package clientes

import (
	"fmt"
	"net/http"
	"regexp"
	"strconv"
	"strings"

	"facturacion-microservicios/internal/httpx"
	"facturacion-microservicios/internal/jsondb"
)

var emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)
var ccRegex = regexp.MustCompile(`^\d+$`)

// Handler maneja las rutas del servicio de clientes.
type Handler struct {
	store *jsondb.Store[Cliente]
}

// NewHandler crea un nuevo handler de clientes.
func NewHandler(store *jsondb.Store[Cliente]) *Handler {
	return &Handler{store: store}
}

// RegisterRoutes registra las rutas del servicio.
func (h *Handler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /clientes", h.handleList)
	mux.HandleFunc("GET /clientes/{id}", h.handleGet)
	mux.HandleFunc("POST /clientes", h.handleCreate)
	mux.HandleFunc("PUT /clientes/{id}", h.handleUpdate)
	mux.HandleFunc("DELETE /clientes/{id}", h.handleDelete)
}

// handleList maneja GET /clientes.
func (h *Handler) handleList(w http.ResponseWriter, r *http.Request) {
	items, err := h.store.GetAll()
	if err != nil {
		httpx.RespondError(w, http.StatusInternalServerError, "Error interno del servidor")
		return
	}
	if items == nil {
		items = []Cliente{}
	}
	httpx.RespondJSON(w, http.StatusOK, items)
}

// handleGet maneja GET /clientes/{id}.
func (h *Handler) handleGet(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		httpx.RespondError(w, http.StatusBadRequest, "ID de cliente inválido")
		return
	}

	cliente, err := h.store.GetByID(id)
	if err != nil {
		httpx.RespondError(w, http.StatusInternalServerError, "Error interno del servidor")
		return
	}
	if cliente == nil {
		httpx.RespondError(w, http.StatusNotFound, fmt.Sprintf("Cliente con ID %d no encontrado", id))
		return
	}

	httpx.RespondJSON(w, http.StatusOK, cliente)
}

// validateCliente valida los campos de un cliente.
func validateCliente(nombre, cc, email, direccion string) string {
	if nombre == "" {
		return "El nombre es obligatorio"
	}
	if cc == "" {
		return "La cédula (cc) es obligatoria"
	}
	if !ccRegex.MatchString(cc) {
		return "La cédula (cc) debe ser numérica"
	}
	if !emailRegex.MatchString(email) {
		return "El formato del email es inválido"
	}
	if direccion == "" {
		return "La dirección es obligatoria"
	}
	return ""
}

// handleCreate maneja POST /clientes.
func (h *Handler) handleCreate(w http.ResponseWriter, r *http.Request) {
	var req CreateRequest
	if !httpx.DecodeJSON(w, r, &req) {
		return
	}

	req.Nombre = strings.TrimSpace(req.Nombre)
	req.CC = strings.TrimSpace(req.CC)
	req.Email = strings.TrimSpace(strings.ToLower(req.Email))
	req.Direccion = strings.TrimSpace(req.Direccion)

	if msg := validateCliente(req.Nombre, req.CC, req.Email, req.Direccion); msg != "" {
		httpx.RespondError(w, http.StatusBadRequest, msg)
		return
	}

	// Verificar email único
	existing, err := h.store.FindOne(func(c *Cliente) bool {
		return strings.EqualFold(c.Email, req.Email)
	})
	if err != nil {
		httpx.RespondError(w, http.StatusInternalServerError, "Error interno del servidor")
		return
	}
	if existing != nil {
		httpx.RespondError(w, http.StatusConflict, "Ya existe un cliente con ese email")
		return
	}

	// Verificar CC único
	existing, err = h.store.FindOne(func(c *Cliente) bool {
		return c.CC == req.CC
	})
	if err != nil {
		httpx.RespondError(w, http.StatusInternalServerError, "Error interno del servidor")
		return
	}
	if existing != nil {
		httpx.RespondError(w, http.StatusConflict, "Ya existe un cliente con esa cédula")
		return
	}

	cliente := Cliente{
		Nombre:    req.Nombre,
		CC:        req.CC,
		Email:     req.Email,
		Direccion: req.Direccion,
	}

	cliente, err = h.store.Create(cliente)
	if err != nil {
		httpx.RespondError(w, http.StatusInternalServerError, "Error interno del servidor")
		return
	}

	httpx.RespondCreated(w, cliente)
}

// handleUpdate maneja PUT /clientes/{id}.
func (h *Handler) handleUpdate(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		httpx.RespondError(w, http.StatusBadRequest, "ID de cliente inválido")
		return
	}

	var req UpdateRequest
	if !httpx.DecodeJSON(w, r, &req) {
		return
	}

	req.Nombre = strings.TrimSpace(req.Nombre)
	req.CC = strings.TrimSpace(req.CC)
	req.Email = strings.TrimSpace(strings.ToLower(req.Email))
	req.Direccion = strings.TrimSpace(req.Direccion)

	if msg := validateCliente(req.Nombre, req.CC, req.Email, req.Direccion); msg != "" {
		httpx.RespondError(w, http.StatusBadRequest, msg)
		return
	}

	// Verificar email único (excluyendo el cliente actual)
	existing, err := h.store.FindOne(func(c *Cliente) bool {
		return strings.EqualFold(c.Email, req.Email) && c.ID != id
	})
	if err != nil {
		httpx.RespondError(w, http.StatusInternalServerError, "Error interno del servidor")
		return
	}
	if existing != nil {
		httpx.RespondError(w, http.StatusConflict, "Ya existe un cliente con ese email")
		return
	}

	// Verificar CC único (excluyendo el cliente actual)
	existing, err = h.store.FindOne(func(c *Cliente) bool {
		return c.CC == req.CC && c.ID != id
	})
	if err != nil {
		httpx.RespondError(w, http.StatusInternalServerError, "Error interno del servidor")
		return
	}
	if existing != nil {
		httpx.RespondError(w, http.StatusConflict, "Ya existe un cliente con esa cédula")
		return
	}

	updated, ok, err := h.store.Update(id, func(c *Cliente) {
		c.Nombre = req.Nombre
		c.CC = req.CC
		c.Email = req.Email
		c.Direccion = req.Direccion
	})
	if err != nil {
		httpx.RespondError(w, http.StatusInternalServerError, "Error interno del servidor")
		return
	}
	if !ok {
		httpx.RespondError(w, http.StatusNotFound, fmt.Sprintf("Cliente con ID %d no encontrado", id))
		return
	}

	httpx.RespondJSON(w, http.StatusOK, updated)
}

// handleDelete maneja DELETE /clientes/{id}.
func (h *Handler) handleDelete(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		httpx.RespondError(w, http.StatusBadRequest, "ID de cliente inválido")
		return
	}

	ok, err := h.store.Delete(id)
	if err != nil {
		httpx.RespondError(w, http.StatusInternalServerError, "Error interno del servidor")
		return
	}
	if !ok {
		httpx.RespondError(w, http.StatusNotFound, fmt.Sprintf("Cliente con ID %d no encontrado", id))
		return
	}

	httpx.RespondNoContent(w)
}
