// Orquestador principal: inicia los 4 microservicios y el servidor web con un solo comando.
// Uso: go run .   (o go run main.go)
package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"facturacion-microservicios/internal/auth"
	"facturacion-microservicios/internal/clientes"
	"facturacion-microservicios/internal/facturacion"
	"facturacion-microservicios/internal/jsondb"
	"facturacion-microservicios/internal/middleware"
	"facturacion-microservicios/internal/productos"
)

func main() {
	jwtSecret := getEnv("JWT_SECRET", "mi-secreto-desarrollo-2026")
	authPort := getEnv("AUTH_PORT", "8081")
	prodPort := getEnv("PROD_PORT", "8082")
	cliPort := getEnv("CLI_PORT", "8083")
	factPort := getEnv("FACT_PORT", "8084")
	webPort := getEnv("WEB_PORT", "3000")

	log.Println("=============================================================")
	log.Println("🚀 INICIANDO SISTEMA DE FACTURACIÓN DIGITAL (MICROSERVICIOS)")
	log.Println("=============================================================")

	// -------------------------------------------------------------
	// 1. Servicio de Autenticación (:8081)
	// -------------------------------------------------------------
	authStore, err := jsondb.New[auth.Usuario]("data/usuarios.json", 3,
		func(u *auth.Usuario) int { return u.ID },
		func(u *auth.Usuario, id int) { u.ID = id })
	if err != nil {
		log.Fatalf("Error inicializando store de usuarios: %v", err)
	}

	authHandler := auth.NewHandler(authStore, jwtSecret)
	authMux := http.NewServeMux()
	authHandler.RegisterRoutes(authMux)
	authFinal := middleware.Chain(authMux, middleware.CORS, middleware.Logging, middleware.Recover)

	serverAuth := &http.Server{
		Addr:         ":" + authPort,
		Handler:      authFinal,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
	}

	// -------------------------------------------------------------
	// 2. Servicio de Productos (:8082)
	// -------------------------------------------------------------
	prodStore, err := jsondb.New[productos.Producto]("data/productos.json", 104,
		func(p *productos.Producto) int { return p.ID },
		func(p *productos.Producto, id int) { p.ID = id })
	if err != nil {
		log.Fatalf("Error inicializando store de productos: %v", err)
	}

	prodHandler := productos.NewHandler(prodStore)
	prodMux := http.NewServeMux()
	prodHandler.RegisterRoutes(prodMux)

	prodProtectedMux := http.NewServeMux()
	jwtMiddleware := middleware.RequireJWT(jwtSecret)
	prodProtectedMux.Handle("GET /health", prodMux)
	prodProtectedMux.Handle("/", jwtMiddleware(prodMux))
	prodFinal := middleware.Chain(prodProtectedMux, middleware.CORS, middleware.Logging, middleware.Recover)

	serverProd := &http.Server{
		Addr:         ":" + prodPort,
		Handler:      prodFinal,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
	}

	// -------------------------------------------------------------
	// 3. Servicio de Clientes (:8083)
	// -------------------------------------------------------------
	cliStore, err := jsondb.New[clientes.Cliente]("data/clientes.json", 203,
		func(c *clientes.Cliente) int { return c.ID },
		func(c *clientes.Cliente, id int) { c.ID = id })
	if err != nil {
		log.Fatalf("Error inicializando store de clientes: %v", err)
	}

	cliHandler := clientes.NewHandler(cliStore)
	cliMux := http.NewServeMux()
	cliHandler.RegisterRoutes(cliMux)

	cliProtectedMux := http.NewServeMux()
	cliProtectedMux.Handle("GET /health", cliMux)
	cliProtectedMux.Handle("/", jwtMiddleware(cliMux))
	cliFinal := middleware.Chain(cliProtectedMux, middleware.CORS, middleware.Logging, middleware.Recover)

	serverCli := &http.Server{
		Addr:         ":" + cliPort,
		Handler:      cliFinal,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
	}

	// -------------------------------------------------------------
	// 4. Servicio de Facturación (:8084)
	// -------------------------------------------------------------
	factStore, err := jsondb.New[facturacion.Factura]("data/facturas.json", 5003,
		func(f *facturacion.Factura) int { return f.FacturaID },
		func(f *facturacion.Factura, id int) { f.FacturaID = id })
	if err != nil {
		log.Fatalf("Error inicializando store de facturas: %v", err)
	}

	factUrls := facturacion.ServiceURLs{
		AuthURL:      "http://localhost:" + authPort,
		ProductosURL: "http://localhost:" + prodPort,
		ClientesURL:  "http://localhost:" + cliPort,
	}

	factHandler := facturacion.NewHandler(factStore, factUrls)
	factMux := http.NewServeMux()
	factHandler.RegisterRoutes(factMux)
	factFinal := middleware.Chain(factMux, middleware.CORS, middleware.Logging, middleware.Recover)

	serverFact := &http.Server{
		Addr:         ":" + factPort,
		Handler:      factFinal,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 30 * time.Second,
	}

	// -------------------------------------------------------------
	// 5. Servidor Web Estático del Frontend (:3000)
	// -------------------------------------------------------------
	frontendDir := getEnv("FRONTEND_DIR", "frontend")
	webMux := http.NewServeMux()
	webMux.Handle("/", http.FileServer(http.Dir(frontendDir)))

	serverWeb := &http.Server{
		Addr:         ":" + webPort,
		Handler:      webMux,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
	}

	// Iniciar cada servidor en su propia goroutine
	servers := []*http.Server{serverAuth, serverProd, serverCli, serverFact, serverWeb}
	names := []string{"🔐 Auth (:8081)", "📦 Productos (:8082)", "👥 Clientes (:8083)", "🧾 Facturación (:8084)", "🌐 Frontend Web (:3000)"}

	for i, s := range servers {
		go func(srv *http.Server, name string) {
			log.Printf("✔ %s escuchando...", name)
			if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
				log.Printf("❌ Error en %s: %v", name, err)
			}
		}(s, names[i])
	}

	log.Println("-------------------------------------------------------------")
	log.Printf("✨ ¡SISTEMA LISTO! Abre tu navegador en: http://localhost:%s", webPort)
	log.Println("   Presiona Ctrl+C en esta terminal para detener todo.")
	log.Println("=============================================================")

	// Manejo de apagado ordenado con Ctrl+C
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	log.Println("\nDeteniendo todos los microservicios...")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var wg sync.WaitGroup
	for _, s := range servers {
		wg.Add(1)
		go func(srv *http.Server) {
			defer wg.Done()
			_ = srv.Shutdown(ctx)
		}(s)
	}
	wg.Wait()

	log.Println("Todos los servicios se han detenido correctamente. ¡Hasta luego!")
}

func getEnv(key, defaultVal string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return defaultVal
}
