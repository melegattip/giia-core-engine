package main

import (
	"fmt"
	"log"

	"golang.org/x/crypto/bcrypt"
)

func main() {
	password := "MySecurePass2024!"

	// Hash que está en la base de datos
	storedHash := "$2a$10$rlIZSv1HkoqX69hslbFrHeqBPGX3/ZSl1VGp3L4beATf/2G/n4C6G"

	fmt.Printf("Password: %s\n", password)
	fmt.Printf("Stored Hash: %s\n", storedHash)

	// Verificar si el hash coincide con la contraseña
	err := bcrypt.CompareHashAndPassword([]byte(storedHash), []byte(password))
	if err != nil {
		fmt.Printf("❌ Hash verification failed: %v\n", err)
	} else {
		fmt.Println("✅ Hash verification successful!")
	}

	// Generar un nuevo hash para comparar
	newHash, err := bcrypt.GenerateFromPassword([]byte(password), 10)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("New Hash: %s\n", string(newHash))

	// Verificar el nuevo hash
	err = bcrypt.CompareHashAndPassword(newHash, []byte(password))
	if err != nil {
		fmt.Printf("❌ New hash verification failed: %v\n", err)
	} else {
		fmt.Println("✅ New hash verification successful!")
	}
}
