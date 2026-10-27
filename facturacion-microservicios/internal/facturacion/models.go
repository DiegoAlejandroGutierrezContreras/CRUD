// Package facturacion implementa la lógica del servicio de facturación.
package facturacion

// Factura representa una factura generada.
type Factura struct {
	FacturaID int            `json:"factura_id"`
	Cliente   FacturaCliente `json:"cliente"`
	Usuario   FacturaUsuario `json:"usuario"`
	Productos []FacturaItem  `json:"productos"`
	Total     float64        `json:"total"`
	Fecha     string         `json:"fecha"`
}

// FacturaCliente es la información del cliente dentro de una factura.
type FacturaCliente struct {
	ID        int    `json:"id"`
	Nombre    string `json:"nombre"`
	CC        string `json:"cc"`
	Email     string `json:"email"`
	Direccion string `json:"direccion"`
}

// FacturaUsuario es la información del usuario que generó la factura.
type FacturaUsuario struct {
	ID     int    `json:"id"`
	Nombre string `json:"nombre"`
	Rol    string `json:"rol"`
}

// FacturaItem es un producto dentro de una factura.
type FacturaItem struct {
	ID             int     `json:"id"`
	Nombre         string  `json:"nombre"`
	Cantidad       int     `json:"cantidad"`
	PrecioUnitario float64 `json:"precio_unitario"`
	Subtotal       float64 `json:"subtotal"`
}

// CreateFacturaRequest es el body de POST /facturas.
type CreateFacturaRequest struct {
	ClienteID int                 `json:"cliente_id"`
	Items     []CreateFacturaItem `json:"items"`
}

// CreateFacturaItem es un item dentro del request de crear factura.
type CreateFacturaItem struct {
	ProductoID int `json:"producto_id"`
	Cantidad   int `json:"cantidad"`
}

// --- Modelos para consumir servicios externos ---

// UsuarioExterno es la respuesta de GET /auth/me.
type UsuarioExterno struct {
	ID     int    `json:"id"`
	Nombre string `json:"nombre"`
	Email  string `json:"email"`
	Rol    string `json:"rol"`
}

// ClienteExterno es la respuesta de GET /clientes/{id}.
type ClienteExterno struct {
	ID        int    `json:"id"`
	Nombre    string `json:"nombre"`
	CC        string `json:"cc"`
	Email     string `json:"email"`
	Direccion string `json:"direccion"`
}

// ProductoExterno es la respuesta de GET /productos/{id}.
type ProductoExterno struct {
	ID     int     `json:"id"`
	Nombre string  `json:"nombre"`
	Precio float64 `json:"precio"`
	Stock  int     `json:"stock"`
}

// StockDeltaRequest es el body de PATCH /productos/{id}/stock.
type StockDeltaRequest struct {
	Delta int `json:"delta"`
}

// ErrorExterno es la estructura de error de los servicios externos.
type ErrorExterno struct {
	Error string `json:"error"`
}
