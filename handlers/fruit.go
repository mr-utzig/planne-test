package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/mr-utzig/planne-test/models"
)

// CreateFruit cria uma nova fruta.
func CreateFruit(w http.ResponseWriter, r *http.Request) {
	var payload models.CreateFruitRequest
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		respondWithError(w, http.StatusBadRequest, "Payload inválido")
		return
	}

	if payload.Name == "" || payload.Price <= 0 || payload.ExpiresInSeconds <= 0 {
		respondWithError(w, http.StatusBadRequest, "Campos 'name', 'price' e 'expires_in_seconds' são obrigatórios e devem ser positivos")
		return
	}

	fruit, err := payload.InsertFruitFromPayload()
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Erro ao criar a fruta")
		return
	}

	respondWithJSON(w, http.StatusCreated, fruit)
}

// DeleteFruit exclui uma fruta permanentemente.
func DeleteFruit(w http.ResponseWriter, r *http.Request) {
	fruitID, err := strconv.Atoi(chi.URLParam(r, "fruitID"))
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "ID de fruta inválido")
		return
	}

	err = models.Fruit{}.DeleteByID(fruitID)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Erro ao excluir a fruta")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// StartExpirationJanitor inicia um processo em background que verifica e remove
// frutas expiradas em intervalos regulares.
func StartExpirationJanitor(interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for range ticker.C {
		rowsAffected := models.Fruit{}.DeleteExpireds()

		if rowsAffected > 0 {
			log.Println(rowsAffected, "Fruta(s) expirada(s) removida(s).")
		}
	}
}
