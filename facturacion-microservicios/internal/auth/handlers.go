package auth

import (
	"fmt"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"

	"facturacion-microservicios/internal/httpx"
	"facturacion-microservicios/internal/jsondb"
	"facturacion-microservicios/internal/middleware"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

// emailRegex valida el formato básico de un email.
var emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)

// Handler maneja las rutas del servicio de autenticación.
type Handler struct {
	store     *jsondb.Store[Usuario]
	jwtSecret string
}

// NewHandler crea un nuevo handler de autenticación.
func NewHandler(store *jsondb.Store[Usuario], jwtSecret string) *Handler {
	return &Handler{store: store, jwtSecret: jwtSecret}
}

// RegisterRoutes registra las rutas del servicio.
func (h *Handler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("POST /auth/register", h.handleRegister)
	mux.HandleFunc("POST /auth/login", h.handleLogin)
	mux.HandleFunc("GET /auth/me", h.handleMe)
}

// handleRegister maneja POST /auth/register.
func (h *Handler) handleRegister(w http.ResponseWriter, r *http.Request) {
	var req RegisterRequest
	if !httpx.DecodeJSON(w, r, &req) {
		return
	}

	// Validaciones
	req.Nombre = strings.TrimSpace(req.Nombre)
	req.Email = strings.TrimSpace(strings.ToLower(req.Email))
	req.Password = strings.TrimSpace(req.Password)

	if req.Nombre == "" {
		httpx.RespondError(w, http.StatusBadRequest, "El nombre es obligatorio")
		return
	}

	if !emailRegex.MatchString(req.Email) {
		httpx.RespondError(w, http.StatusBadRequest, "El formato del email es inválido")
		return
	}

	if len(req.Password) < 6 {
		httpx.RespondError(w, http.StatusBadRequest, "La contraseña debe tener al menos 6 caracteres")
		return
	}

	// Verificar email único
	existing, err := h.store.FindOne(func(u *Usuario) bool {
		return strings.EqualFold(u.Email, req.Email)
	})
	if err != nil {
		httpx.RespondError(w, http.StatusInternalServerError, "Error interno del servidor")
		return
	}
	if existing != nil {
		httpx.RespondError(w, http.StatusConflict, "Ya existe un usuario con ese email")
		return
	}

	// Generar hash bcrypt
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		httpx.RespondError(w, http.StatusInternalServerError, "Error interno del servidor")
		return
	}

	usuario := Usuario{
		Nombre:   req.Nombre,
		Email:    req.Email,
		Password: string(hashedPassword),
		Rol:      "cliente",
	}

	usuario, err = h.store.Create(usuario)
	if err != nil {
		httpx.RespondError(w, http.StatusInternalServerError, "Error interno del servidor")
		return
	}

	httpx.RespondCreated(w, usuario.SinPassword())
}

// handleLogin maneja POST /auth/login.
func (h *Handler) handleLogin(w http.ResponseWriter, r *http.Request) {
	var req LoginRequest
	if !httpx.DecodeJSON(w, r, &req) {
		return
	}

	req.Email = strings.TrimSpace(strings.ToLower(req.Email))

	if req.Email == "" || req.Password == "" {
		httpx.RespondError(w, http.StatusBadRequest, "Email y contraseña son obligatorios")
		return
	}

	// Buscar usuario por email
	usuario, err := h.store.FindOne(func(u *Usuario) bool {
		return strings.EqualFold(u.Email, req.Email)
	})
	if err != nil {
		httpx.RespondError(w, http.StatusInternalServerError, "Error interno del servidor")
		return
	}

	if usuario == nil {
		httpx.RespondError(w, http.StatusUnauthorized, "Credenciales incorrectas")
		return
	}

	// Validar exclusivamente con hash bcrypt (sin puertas traseras)
	if err := bcrypt.CompareHashAndPassword([]byte(usuario.Password), []byte(req.Password)); err != nil {
		httpx.RespondError(w, http.StatusUnauthorized, "Credenciales incorrectas")
		return
	}

	// Generar token JWT
	token, err := h.generateToken(usuario)
	if err != nil {
		httpx.RespondError(w, http.StatusInternalServerError, "Error generando token de acceso")
		return
	}

	httpx.RespondJSON(w, http.StatusOK, LoginResponse{
		Token:   token,
		Usuario: usuario.SinPassword(),
	})
}

// handleMe maneja GET /auth/me.
func (h *Handler) handleMe(w http.ResponseWriter, r *http.Request) {
	// Validar token JWT
	claims, err := h.extractClaims(r)
	if err != nil {
		httpx.RespondError(w, http.StatusUnauthorized, "Token inválido o expirado")
		return
	}

	userID, err := strconv.Atoi(claims.Subject)
	if err != nil {
		httpx.RespondError(w, http.StatusUnauthorized, "Token inválido")
		return
	}

	usuario, err := h.store.GetByID(userID)
	if err != nil {
		httpx.RespondError(w, http.StatusInternalServerError, "Error interno del servidor")
		return
	}
	if usuario == nil {
		httpx.RespondError(w, http.StatusNotFound, "Usuario no encontrado")
		return
	}

	httpx.RespondJSON(w, http.StatusOK, usuario.SinPassword())
}

// generateToken genera un token JWT para un usuario.
func (h *Handler) generateToken(u *Usuario) (string, error) {
	claims := middleware.JWTClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   fmt.Sprintf("%d", u.ID),
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(8 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
		Nombre: u.Nombre,
		Email:  u.Email,
		Rol:    u.Rol,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(h.jwtSecret))
}

// extractClaims extrae y valida los claims JWT del header Authorization.
func (h *Handler) extractClaims(r *http.Request) (*middleware.JWTClaims, error) {
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		return nil, fmt.Errorf("token requerido")
	}

	parts := strings.SplitN(authHeader, " ", 2)
	if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
		return nil, fmt.Errorf("formato inválido")
	}

	claims := &middleware.JWTClaims{}
	token, err := jwt.ParseWithClaims(parts[1], claims, func(token *jwt.Token) (any, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("método de firma inesperado")
		}
		return []byte(h.jwtSecret), nil
	})

	if err != nil || !token.Valid {
		return nil, fmt.Errorf("token inválido")
	}

	return claims, nil
}
