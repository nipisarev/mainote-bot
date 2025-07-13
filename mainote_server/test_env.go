package main
import ("fmt"; "os")
func main() {
    key := os.Getenv("INTERNAL_API_KEY")
    fmt.Printf("INTERNAL_API_KEY in Go: %s
", key)
    fmt.Printf("Empty: %t
", key == "")
}
