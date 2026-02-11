package keys

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"os"
	"path/filepath"
)

// GenerateRSAKeyPair generates a new RSA key pair
// bits: key size in bits (recommended: 2048 or 4096)
func GenerateRSAKeyPair(bits int) (*rsa.PrivateKey, error) {
	if bits < 2048 {
		return nil, fmt.Errorf("RSA key size must be at least 2048 bits for security")
	}
	
	privateKey, err := rsa.GenerateKey(rand.Reader, bits)
	if err != nil {
		return nil, fmt.Errorf("generate RSA key pair: %w", err)
	}
	
	return privateKey, nil
}

// SavePrivateKeyPEM saves an RSA private key to a PEM file
func SavePrivateKeyPEM(filename string, key *rsa.PrivateKey) error {
	// Create directory if it doesn't exist
	dir := filepath.Dir(filename)
	if dir != "." && dir != "" {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("create directory: %w", err)
		}
	}
	
	// Encode private key to PKCS#1 format
	privateKeyBytes := x509.MarshalPKCS1PrivateKey(key)
	
	// Create PEM block
	privateKeyPEM := &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: privateKeyBytes,
	}
	
	// Write to file
	file, err := os.OpenFile(filename, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return fmt.Errorf("open file for writing: %w", err)
	}
	defer file.Close()
	
	if err := pem.Encode(file, privateKeyPEM); err != nil {
		return fmt.Errorf("encode PEM: %w", err)
	}
	
	return nil
}

// SavePublicKeyPEM saves an RSA public key to a PEM file
func SavePublicKeyPEM(filename string, key *rsa.PublicKey) error {
	// Create directory if it doesn't exist
	dir := filepath.Dir(filename)
	if dir != "." && dir != "" {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("create directory: %w", err)
		}
	}
	
	// Encode public key to PKIX format
	publicKeyBytes, err := x509.MarshalPKIXPublicKey(key)
	if err != nil {
		return fmt.Errorf("marshal public key: %w", err)
	}
	
	// Create PEM block
	publicKeyPEM := &pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: publicKeyBytes,
	}
	
	// Write to file
	file, err := os.OpenFile(filename, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		return fmt.Errorf("open file for writing: %w", err)
	}
	defer file.Close()
	
	if err := pem.Encode(file, publicKeyPEM); err != nil {
		return fmt.Errorf("encode PEM: %w", err)
	}
	
	return nil
}

// LoadPrivateKeyPEM loads an RSA private key from a PEM file
func LoadPrivateKeyPEM(filename string) (*rsa.PrivateKey, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("read private key file: %w", err)
	}
	
	block, _ := pem.Decode(data)
	if block == nil {
		return nil, fmt.Errorf("failed to decode PEM block")
	}
	
	if block.Type != "RSA PRIVATE KEY" {
		return nil, fmt.Errorf("invalid PEM block type: expected RSA PRIVATE KEY, got %s", block.Type)
	}
	
	privateKey, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("parse private key: %w", err)
	}
	
	return privateKey, nil
}

// LoadPublicKeyPEM loads an RSA public key from a PEM file
func LoadPublicKeyPEM(filename string) (*rsa.PublicKey, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("read public key file: %w", err)
	}
	
	block, _ := pem.Decode(data)
	if block == nil {
		return nil, fmt.Errorf("failed to decode PEM block")
	}
	
	if block.Type != "PUBLIC KEY" {
		return nil, fmt.Errorf("invalid PEM block type: expected PUBLIC KEY, got %s", block.Type)
	}
	
	publicKeyInterface, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("parse public key: %w", err)
	}
	
	publicKey, ok := publicKeyInterface.(*rsa.PublicKey)
	if !ok {
		return nil, fmt.Errorf("not an RSA public key")
	}
	
	return publicKey, nil
}

// GenerateAndSaveKeyPair generates RSA key pair and saves both keys to files
func GenerateAndSaveKeyPair(privateKeyPath, publicKeyPath string, bits int) error {
	// Generate key pair
	privateKey, err := GenerateRSAKeyPair(bits)
	if err != nil {
		return fmt.Errorf("generate key pair: %w", err)
	}
	
	// Save private key
	if err := SavePrivateKeyPEM(privateKeyPath, privateKey); err != nil {
		return fmt.Errorf("save private key: %w", err)
	}
	
	// Save public key
	if err := SavePublicKeyPEM(publicKeyPath, &privateKey.PublicKey); err != nil {
		return fmt.Errorf("save public key: %w", err)
	}
	
	return nil
}
