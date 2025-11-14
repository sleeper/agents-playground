package main

import (
    "flag"
    "log"
    "net/http"

    "github.com/example/potion/internal/app"
)

func main() {
    var addr string
    var dbPath string
    flag.StringVar(&addr, "addr", ":8080", "HTTP listen address")
    flag.StringVar(&dbPath, "db", "./potion.db", "Path to SQLite database")
    flag.Parse()

    server, err := app.NewServer(app.ServerConfig{DBPath: dbPath})
    if err != nil {
        log.Fatalf("unable to initialise server: %v", err)
    }
    defer server.Close()

    log.Printf("Potion server listening on %s with database %s", addr, dbPath)
    if err := server.Run(addr); err != nil && err != http.ErrServerClosed {
        log.Fatalf("server stopped: %v", err)
    }
}
