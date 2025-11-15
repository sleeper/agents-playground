package main

import (
    "flag"
    "log"
    "net/http"

    "github.com/example/potion-revamp/backend/internal/server"
    "github.com/example/potion-revamp/backend/internal/storage"
)

func main() {
    addr := flag.String("addr", ":8080", "HTTP listen address")
    dbPath := flag.String("db", "potion.db", "SQLite database path")
    flag.Parse()

    store, err := storage.NewStore(*dbPath)
    if err != nil {
        log.Fatalf("failed to initialize storage: %v", err)
    }
    defer store.Close()

    srv := server.New(store)
    log.Printf("Potion server listening on %s", *addr)
    if err := http.ListenAndServe(*addr, srv.Router()); err != nil {
        log.Fatalf("server exited: %v", err)
    }
}
