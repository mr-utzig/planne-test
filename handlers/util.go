package handlers

import (
	"encoding/json"
	"net/http"
)

// respondWithError envia uma resposta de erro JSON padronizada.
func respondWithError(w http.ResponseWriter, code int, message string) {
	respondWithJSON(w, code, map[string]string{"error": message})
}

// respondWithJSON envia uma resposta JSON com o status code e payload fornecidos.
func respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	response, _ := json.Marshal(payload)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	w.Write(response)
}
