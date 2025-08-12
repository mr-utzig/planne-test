package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"sort"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/mr-utzig/planne-test/models"
)

// CreateBucket cria um novo balde.
func CreateBucket(w http.ResponseWriter, r *http.Request) {
	var bucket models.Bucket
	if err := json.NewDecoder(r.Body).Decode(&bucket); err != nil {
		respondWithError(w, http.StatusBadRequest, "Payload inválido")
		return
	}

	if bucket.Capacity <= 0 {
		respondWithError(w, http.StatusBadRequest, "A capacidade deve ser maior que zero")
		return
	}

	if err := bucket.Insert(); err != nil {
		respondWithError(w, http.StatusInternalServerError, "Erro ao criar o balde")
		return
	}

	respondWithJSON(w, http.StatusCreated, bucket)
}

// DeleteBucket exclui um balde, se ele estiver vazio.
func DeleteBucket(w http.ResponseWriter, r *http.Request) {
	bucketID, err := strconv.Atoi(chi.URLParam(r, "bucketID"))
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "ID de balde inválido")
		return
	}

	fruitsInBucket, err := models.Fruit{}.GetFruitsInBucket(bucketID)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Erro ao verificar o balde")
		return
	}

	if len(fruitsInBucket) > 0 {
		respondWithError(w, http.StatusBadRequest, "Não é possível excluir um balde que não está vazio")
		return
	}

	err = models.Bucket{}.DeleteByID(bucketID)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Erro ao excluir o balde")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// ListBuckets lista todos os baldes com detalhes, ordenados por ocupação.
func ListBuckets(w http.ResponseWriter, r *http.Request) {
	buckets, err := models.Bucket{}.GetAll()
	if err != nil {
		if err == sql.ErrNoRows {
			respondWithError(w, http.StatusNotFound, "Nenhum balde encontrado")
			return
		}

		respondWithError(w, http.StatusInternalServerError, "Erro ao buscar baldes")
		return
	}

	var allBucketsDetails []models.BucketDetails
	for _, bucket := range buckets {
		fruitsInBucket, err := models.Fruit{}.GetFruitsInBucket(bucket.ID)
		if err != nil && err != sql.ErrNoRows {
			respondWithError(w, http.StatusInternalServerError, "Erro ao buscar frutas do balde")
			return
		}

		bucketDetails := models.BucketDetails{
			ID:       bucket.ID,
			Capacity: bucket.Capacity,
			Fruits:   fruitsInBucket,
		}

		bucketDetails.CalcTotalValue()
		bucketDetails.CalcOccupancyPercentage()

		allBucketsDetails = append(allBucketsDetails, bucketDetails)
	}

	// Ordena os baldes pela ocupação em ordem decrescente
	sort.Slice(allBucketsDetails, func(i, j int) bool {
		return allBucketsDetails[i].Occupancy > allBucketsDetails[j].Occupancy
	})

	respondWithJSON(w, http.StatusOK, allBucketsDetails)
}

// DepositFruit deposita uma fruta em um balde.
func DepositFruit(w http.ResponseWriter, r *http.Request) {
	bucketID, err := strconv.Atoi(chi.URLParam(r, "bucketID"))
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "ID de balde inválido")
		return
	}

	var payload struct {
		FruitID int `json:"fruit_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		respondWithError(w, http.StatusBadRequest, "Payload inválido")
		return
	}

	// Verifica a capacidade do balde
	bucket := models.Bucket{}
	if err := bucket.GetByID(bucketID); err != nil {
		if err == sql.ErrNoRows {
			respondWithError(w, http.StatusNotFound, "Balde não encontrado")
			return
		}

		respondWithError(w, http.StatusInternalServerError, "Erro ao verificar capacidade do balde")
		return
	}

	fruitsInBucket, err := models.Fruit{}.GetFruitsInBucket(bucket.ID)
	if err != nil && err != sql.ErrNoRows {
		respondWithError(w, http.StatusInternalServerError, "Erro ao buscar frutas do balde")
		return
	}

	if len(fruitsInBucket) >= bucket.Capacity {
		respondWithError(w, http.StatusBadRequest, "Capacidade máxima do balde atingida")
		return
	}

	// Verifica se a fruta existe e não está em outro balde
	fruit := models.Fruit{}
	if err := fruit.GetByID(payload.FruitID); err != nil {
		if err == sql.ErrNoRows {
			respondWithError(w, http.StatusNotFound, "Fruta não encontrada")
			return
		}

		respondWithError(w, http.StatusInternalServerError, "Erro ao verificar a fruta")
		return
	}

	if fruit.BucketID.Valid {
		respondWithError(w, http.StatusBadRequest, "A fruta já está em outro balde")
		return
	}

	// Deposita a fruta
	_, err = fruit.AddToBucket(bucket.ID)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Erro ao depositar a fruta")
		return
	}

	respondWithJSON(w, http.StatusOK, map[string]string{"message": "Fruta depositada com sucesso"})
}

// RemoveFruitFromBucket remove uma fruta de um balde.
func RemoveFruitFromBucket(w http.ResponseWriter, r *http.Request) {
	bucketID, err := strconv.Atoi(chi.URLParam(r, "bucketID"))
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "ID de balde inválido")
		return
	}
	fruitID, err := strconv.Atoi(chi.URLParam(r, "fruitID"))
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "ID de fruta inválido")
		return
	}

	rowsAffected, err := models.Fruit{}.RemoveFromBucket(fruitID, bucketID)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Erro ao remover a fruta do balde")
		return
	}

	if rowsAffected == 0 {
		respondWithError(w, http.StatusNotFound, "Fruta não encontrada neste balde")
		return
	}

	respondWithJSON(w, http.StatusOK, map[string]string{"message": "Fruta removida com sucesso"})
}
