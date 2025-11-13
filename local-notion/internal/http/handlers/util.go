package handlers

import (
	"encoding/json"
	"net/http"
)

// respondJSON writes JSON responses with envelope structure.
func respondJSON(w http.ResponseWriter, status int, envelope Envelope) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(envelope)
}
