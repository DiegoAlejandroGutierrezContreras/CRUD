// Package clientes implementa la lógica del servicio de clientes.
package clientes

// Cliente representa un cliente en el sistema.
type Cliente struct {
	ID        int    `json:"id"`
	Nombre    string `json:"nombre"`
	CC        string `json:"cc"`
	Email     string `json:"email"`
	Direccion string `json:"direccion"`
}

// CreateRequest es el body de POST /clientes.
type CreateRequest struct {
	Nombre    string `json:"nombre"`
	CC        string `json:"cc"`
	Email     string `json:"email"`
	Direccion string `json:"direccion"`
}

// UpdateRequest es el body de PUT /clientes/{id}.
type UpdateRequest struct {
	Nombre    string `json:"nombre"`
	CC        string `json:"cc"`
	Email     string `json:"email"`
	Direccion string `json:"direccion"`
}
