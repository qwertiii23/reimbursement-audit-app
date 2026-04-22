package main

import (
	"fmt"
	"log"

	"reimbursement-audit/internal/pkg/crypto"
)

func main() {
	passwords := map[string]string{
		"admin":   "admin123",
		"user":    "password123",
		"test":    "test123",
		"finance": "finance123",
	}

	for username, password := range passwords {
		hash, err := crypto.HashPassword(password)
		if err != nil {
			log.Fatalf("Failed to hash password for %s: %v", username, err)
		}
		fmt.Printf("Username: %s, Password: %s, Hash: %s\n", username, password, hash)
	}
}
