package main

import (
	"encoding/json"
	"log"
	"net/http"
)

// respondWithError schreibt eine Fehlerantwort als JSON.
// Bei 5xx wird zusätzlich ein Hinweis ins Log geschrieben.
func respondWithError(w http.ResponseWriter, code int, msg string) {
	if code >= http.StatusInternalServerError {
		log.Printf("responding with %d error: %s", code, msg)
	}

	type errorResponse struct {
		Error string `json:"error"`
	}
	respondWithJSON(w, code, errorResponse{Error: msg})
}

// respondWithJSON serialisiert payload zu JSON und schreibt es in die Response.
// Alle Schreibfehler werden geprüft und geloggt (Fix für gosec G104).
func respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")

	data, err := json.Marshal(payload)
	if err != nil {
		log.Printf("marshal json failed: %v", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(code)
	if _, err := w.Write(data); err != nil {
		// WICHTIG: Fehler beim Schreiben nicht ignorieren (gosec G104)
		log.Printf("write response failed: %v", err)
	}
}
