// Package httpx provee helpers para respuestas JSON y errores estándar.
package httpx

import (
	"encoding/json"
	"log"
	"net/http"
)

// ErrorResponse es la estructura estándar de error.
type ErrorResponse struct {
	Error string `json:"error"`
}

// RespondJSON escribe una respuesta JSON con el código de estado dado.
func RespondJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	if data != nil {
		if err := json.NewEncoder(w).Encode(data); err != nil {
			log.Printf("Error codificando respuesta JSON: %v", err)
		}
	}
}

// RespondError escribe un error JSON con el código de estado dado.
func RespondError(w http.ResponseWriter, status int, mensaje string) {
	RespondJSON(w, status, ErrorResponse{Error: mensaje})
}

// DecodeJSON decodifica el body JSON de un request. Retorna false si falla
// (y ya envía la respuesta de error al cliente).
func DecodeJSON(w http.ResponseWriter, r *http.Request, dest any) bool {
	// Limitar tamaño del body a 1MB
	r.Body = http.MaxBytesReader(w, r.Body, 1048576)
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()

	if err := decoder.Decode(dest); err != nil {
		RespondError(w, http.StatusBadRequest, "JSON inválido: "+err.Error())
		return false
	}
	return true
}

// RespondCreated responde con 201 y el objeto creado.
func RespondCreated(w http.ResponseWriter, data any) {
	RespondJSON(w, http.StatusCreated, data)
}

// RespondNoContent responde con 204 sin body.
func RespondNoContent(w http.ResponseWriter) {
	w.WriteHeader(http.StatusNoContent)
}
