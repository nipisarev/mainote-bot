package main

import (
	"fmt"
	"os"
)

func main() {
	key := os.Getenv("INTERNAL_API_KEY")
	fmt.Printf("INTERNAL_API_KEY in Go: %s\n", key)
	fmt.Printf("Empty: %t\n", key == "")
}
