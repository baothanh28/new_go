// +build ignore

package main

import (
	"fmt"
	"os"
	"path/filepath"

	"myapp/internal/pkg/auth/keys"
)

func main() {
	// Get the directory where this file is located
	dir := filepath.Dir(os.Args[0])
	if len(os.Args) > 1 {
		dir = os.Args[1]
	}
	
	privateKeyPath := filepath.Join(dir, "private.pem")
	publicKeyPath := filepath.Join(dir, "public.pem")
	
	fmt.Printf("Generating RSA key pair...\n")
	fmt.Printf("Private key: %s\n", privateKeyPath)
	fmt.Printf("Public key: %s\n", publicKeyPath)
	
	if err := keys.GenerateAndSaveKeyPair(privateKeyPath, publicKeyPath, 2048); err != nil {
		fmt.Fprintf(os.Stderr, "Error generating keys: %v\n", err)
		os.Exit(1)
	}
	
	fmt.Println("RSA keys generated successfully!")
}
