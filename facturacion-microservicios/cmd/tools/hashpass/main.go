// Utilidad para generar hashes bcrypt de contraseñas.
// Uso: go run cmd/tools/hashpass/main.go <contraseña>
package main

import (
	"fmt"
	"os"

	"golang.org/x/crypto/bcrypt"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintf(os.Stderr, "Uso: hashpass <contraseña>\n")
		os.Exit(1)
	}

	password := os.Args[1]
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error generando hash: %v\n", err)
		os.Exit(1)
	}

	fmt.Println(string(hash))
}
