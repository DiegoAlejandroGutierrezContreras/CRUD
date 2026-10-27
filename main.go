// Punto de entrada rápido desde la raíz de la carpeta CRUD.
// Permite ejecutar todo el sistema de microservicios con: go run .
package main

import (
	"log"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"syscall"
)

func main() {
	targetDir := "facturacion-microservicios"
	if _, err := os.Stat(targetDir); os.IsNotExist(err) {
		log.Fatalf("No se encontró la carpeta %s", targetDir)
	}

	cmd := exec.Command("go", "run", ".")
	cmd.Dir = filepath.Clean(targetDir)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin

	// Propagar señales de interrupción (Ctrl+C)
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		sig := <-sigChan
		if cmd.Process != nil {
			_ = cmd.Process.Signal(sig)
		}
	}()

	if err := cmd.Run(); err != nil {
		os.Exit(1)
	}
}
