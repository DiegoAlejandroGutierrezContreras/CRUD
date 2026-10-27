// Package productos implementa la lógica del servicio de productos.
package productos

// Producto representa un producto en el catálogo.
type Producto struct {
	ID     int     `json:"id"`
	Nombre string  `json:"nombre"`
	Precio float64 `json:"precio"`
	Stock  int     `json:"stock"`
}

// CreateRequest es el body de POST /productos.
type CreateRequest struct {
	Nombre string  `json:"nombre"`
	Precio float64 `json:"precio"`
	Stock  int     `json:"stock"`
}

// UpdateRequest es el body de PUT /productos/{id}.
type UpdateRequest struct {
	Nombre string  `json:"nombre"`
	Precio float64 `json:"precio"`
	Stock  int     `json:"stock"`
}

// StockDelta es el body de PATCH /productos/{id}/stock.
type StockDelta struct {
	Delta int `json:"delta"`
}
