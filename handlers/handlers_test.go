package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/mr-utzig/planne-test/database"
	"github.com/mr-utzig/planne-test/models"
)

var r *chi.Mux

// TestMain é executado antes de todos os testes neste pacote.
// É usado para configurar o ambiente de teste (banco de dados em memória)
// e limpar após a execução.
func TestMain(m *testing.M) {
	// Configura o banco de dados em memória para os testes
	database.DB, _ = database.InitDBTest()
	defer database.DB.Close()

	// Configura o roteador com as mesmas rotas da aplicação principal
	r = chi.NewRouter()
	r.Route("/buckets", func(r chi.Router) {
		r.Post("/", CreateBucket)
		r.Get("/", ListBuckets)
		r.Delete("/{bucketID}", DeleteBucket)
		r.Post("/{bucketID}/fruits", DepositFruit)
		r.Delete("/{bucketID}/fruits/{fruitID}", RemoveFruitFromBucket)
	})
	r.Route("/fruits", func(r chi.Router) {
		r.Post("/", CreateFruit)
		r.Delete("/{fruitID}", DeleteFruit)
	})

	// Executa os testes
	exitCode := m.Run()

	// Limpa o banco de dados após os testes
	clearTables()

	os.Exit(exitCode)
}

// clearTables limpa todas as tabelas para garantir que os testes sejam independentes.
func clearTables() {
	database.DB.Exec("DELETE FROM fruits")
	database.DB.Exec("DELETE FROM buckets")
	database.DB.Exec("UPDATE sqlite_sequence SET seq = 0 WHERE name = 'fruits'")
	database.DB.Exec("UPDATE sqlite_sequence SET seq = 0 WHERE name = 'buckets'")
}

// executeRequest é uma função auxiliar para executar requisições HTTP contra o nosso roteador de teste.
func executeRequest(req *http.Request) *httptest.ResponseRecorder {
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)
	return rr
}

// checkResponseCode é uma função auxiliar para verificar o status code da resposta.
func checkResponseCode(t *testing.T, expected, actual int) {
	if expected != actual {
		t.Errorf("Expected response code %d. Got %d\n", expected, actual)
	}
}

// TestCreateBucket verifica a criação de um novo balde.
func TestCreateBucket(t *testing.T) {
	clearTables()

	payload := []byte(`{"capacity": 10}`)
	req, _ := http.NewRequest("POST", "/buckets", bytes.NewBuffer(payload))
	response := executeRequest(req)

	checkResponseCode(t, http.StatusCreated, response.Code)

	var bucket models.Bucket
	json.Unmarshal(response.Body.Bytes(), &bucket)

	if bucket.Capacity != 10 {
		t.Errorf("Expected capacity to be 10. Got %v", bucket.Capacity)
	}
	if bucket.ID != 1 {
		t.Errorf("Expected bucket ID to be 1. Got %v", bucket.ID)
	}
}

// TestCreateFruit verifica a criação de uma nova fruta.
func TestCreateFruit(t *testing.T) {
	clearTables()

	payload := []byte(`{"name": "Test Fruit", "price": 9.99, "expires_in_seconds": 60}`)
	req, _ := http.NewRequest("POST", "/fruits", bytes.NewBuffer(payload))
	response := executeRequest(req)

	checkResponseCode(t, http.StatusCreated, response.Code)

	var fruit models.Fruit
	json.Unmarshal(response.Body.Bytes(), &fruit)

	if fruit.Name != "Test Fruit" {
		t.Errorf("Expected fruit name to be 'Test Fruit'. Got '%s'", fruit.Name)
	}
	if fruit.Price != 9.99 {
		t.Errorf("Expected fruit price to be 9.99. Got %v", fruit.Price)
	}
}

// TestDepositFruitInBucket verifica se uma fruta pode ser depositada em um balde.
func TestDepositFruitInBucket(t *testing.T) {
	clearTables()
	// 1. Cria um balde
	database.DB.Exec("INSERT INTO buckets (id, capacity) VALUES (1, 5)")
	// 2. Cria uma fruta
	database.DB.Exec("INSERT INTO fruits (id, name, price, expiration_time) VALUES (1, 'Apple', 1.0, ?)", time.Now().Add(1*time.Hour).Unix())

	payload := []byte(`{"fruit_id": 1}`)
	req, _ := http.NewRequest("POST", "/buckets/1/fruits", bytes.NewBuffer(payload))
	response := executeRequest(req)

	checkResponseCode(t, http.StatusOK, response.Code)

	// Verifica no DB
	var bucketID int
	err := database.DB.QueryRow("SELECT bucket_id FROM fruits WHERE id = 1").Scan(&bucketID)
	if err != nil || bucketID != 1 {
		t.Errorf("Expected fruit to be in bucket 1. Got error: %v or wrong bucketID: %d", err, bucketID)
	}
}

// TestDepositFruitInFullBucket verifica a restrição de capacidade do balde.
func TestDepositFruitInFullBucket(t *testing.T) {
	clearTables()
	// 1. Cria um balde com capacidade 1
	database.DB.Exec("INSERT INTO buckets (id, capacity) VALUES (1, 1)")
	// 2. Cria duas frutas
	database.DB.Exec("INSERT INTO fruits (id, name, price, expiration_time, bucket_id) VALUES (1, 'Apple', 1.0, ?, 1)", time.Now().Add(1*time.Hour).Unix())
	database.DB.Exec("INSERT INTO fruits (id, name, price, expiration_time) VALUES (2, 'Orange', 1.2, ?)", time.Now().Add(1*time.Hour).Unix())

	// 3. Tenta depositar a segunda fruta
	payload := []byte(`{"fruit_id": 2}`)
	req, _ := http.NewRequest("POST", "/buckets/1/fruits", bytes.NewBuffer(payload))
	response := executeRequest(req)

	checkResponseCode(t, http.StatusBadRequest, response.Code)
}

// TestRemoveFruitFromBucket verifica se uma fruta pode ser removida de um balde.
func TestRemoveFruitFromBucket(t *testing.T) {
	clearTables()
	// 1. Cria um balde e uma fruta já dentro dele
	database.DB.Exec("INSERT INTO buckets (id, capacity) VALUES (1, 5)")
	database.DB.Exec("INSERT INTO fruits (id, name, price, expiration_time, bucket_id) VALUES (1, 'Apple', 1.0, ?, 1)", time.Now().Add(1*time.Hour).Unix())

	req, _ := http.NewRequest("DELETE", "/buckets/1/fruits/1", nil)
	response := executeRequest(req)

	checkResponseCode(t, http.StatusOK, response.Code)

	// Verifica no DB
	var bucketID *int
	err := database.DB.QueryRow("SELECT bucket_id FROM fruits WHERE id = 1").Scan(&bucketID)
	if err != nil {
		t.Errorf("Expected bucket_id to be NULL, but got error: %v", err)
	}
	if bucketID != nil {
		t.Errorf("Expected bucket_id to be NULL, but got %v", *bucketID)
	}
}

// TestDeleteNonEmptyBucket verifica que um balde com frutas não pode ser excluído.
func TestDeleteNonEmptyBucket(t *testing.T) {
	clearTables()
	// 1. Cria um balde e uma fruta dentro dele
	database.DB.Exec("INSERT INTO buckets (id, capacity) VALUES (1, 5)")
	database.DB.Exec("INSERT INTO fruits (id, name, price, expiration_time, bucket_id) VALUES (1, 'Apple', 1.0, ?, 1)", time.Now().Add(1*time.Hour).Unix())

	req, _ := http.NewRequest("DELETE", "/buckets/1", nil)
	response := executeRequest(req)

	checkResponseCode(t, http.StatusBadRequest, response.Code)
}

// TestDeleteEmptyBucket verifica que um balde vazio pode ser excluído.
func TestDeleteEmptyBucket(t *testing.T) {
	clearTables()
	// 1. Cria um balde vazio
	database.DB.Exec("INSERT INTO buckets (id, capacity) VALUES (1, 5)")

	req, _ := http.NewRequest("DELETE", "/buckets/1", nil)
	response := executeRequest(req)

	checkResponseCode(t, http.StatusNoContent, response.Code)
}
