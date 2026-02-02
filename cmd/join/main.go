package main

import (
    "0TLink/internal/auth"
    "log"
)

func main() {
    err := auth.JoinMesh("http://localhost:8081", "your-secret-token", "miku-laptop")
    if err != nil {
        log.Fatalf("Join failed: %v", err)
    }
    log.Println("Join successful! Check ~/.local/share/0TLink")
}