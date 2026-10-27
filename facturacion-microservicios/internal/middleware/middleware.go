// Package middleware provee middlewares HTTP comunes: CORS, logging, JWT y recover.
package middleware

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"facturacion-microservicios/internal/httpx"

	"github.com/golang-jwt/jwt/v5"
)

// contextKey es un tipo para claves de contexto.
type contextKey string

const (
	// ClaimsKey es la clave de contexto para los claims JWT.
	ClaimsKey contextKey = "claims"
)

// JWTClaims son los claims del token JWT.
type JWTClaims struct {
	jwt.RegisteredClaims
	Nombre string `json:"nombre"`
	Email  string `json:"email"`
	Rol    string `json:"rol"`
}

// CORS agrega headers de CORS y maneja preflight OPTIONS.
func CORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Authorization, Content-Type")
		w.Header().Set("Access-Control-Max-Age", "86400")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// Logging registra cada request con método, ruta, status y duración.
func Logging(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		sw := &statusWriter{ResponseWriter: w, status: http.StatusOK}

		next.ServeHTTP(sw, r)

		log.Printf("%s %s → %d (%s)", r.Method, r.URL.Path, sw.status, time.Since(start).Round(time.Millisecond))
	})
}

// statusWriter es un wrapper de ResponseWriter que captura el código de estado.
type statusWriter struct {
	http.ResponseWriter
	status int
}

func (sw *statusWriter) WriteHeader(code int) {
	sw.status = code
	sw.ResponseWriter.WriteHeader(code)
}

// Recover recupera panics en handlers y responde con error 500.
func Recover(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if rec := recover(); rec != nil {
				log.Printf("PANIC recuperado: %v", rec)
				httpx.RespondError(w, http.StatusInternalServerError, "Error interno del servidor")
			}
		}()
		next.ServeHTTP(w, r)
	})
}

// RequireJWT valida el token JWT y agrega los claims al contexto.
func RequireJWT(secret string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				httpx.RespondError(w, http.StatusUnauthorized, "Token de autorización requerido")
				return
			}

			parts := strings.SplitN(authHeader, " ", 2)
			if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
				httpx.RespondError(w, http.StatusUnauthorized, "Formato de autorización inválido. Use: Bearer <token>")
				return
			}

			tokenString := parts[1]

			claims := &JWTClaims{}
			token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (any, error) {
				if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
					return nil, fmt.Errorf("método de firma inesperado: %v", token.Header["alg"])
				}
				return []byte(secret), nil
			})

			if err != nil || !token.Valid {
				httpx.RespondError(w, http.StatusUnauthorized, "Token inválido o expirado")
				return
			}

			ctx := context.WithValue(r.Context(), ClaimsKey, claims)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// GetClaims extrae los claims JWT del contexto.
func GetClaims(r *http.Request) *JWTClaims {
	claims, _ := r.Context().Value(ClaimsKey).(*JWTClaims)
	return claims
}

// Chain aplica una cadena de middlewares a un handler.
func Chain(h http.Handler, middlewares ...func(http.Handler) http.Handler) http.Handler {
	for i := len(middlewares) - 1; i >= 0; i-- {
		h = middlewares[i](h)
	}
	return h
}
