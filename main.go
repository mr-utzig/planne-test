package main

import (
	"log"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/mr-utzig/planne-test/database"
	"github.com/mr-utzig/planne-test/handlers"
)

func main() {
	// Inicializa o banco de dados SQLite
	if err := database.InitDB(); err != nil {
		log.Fatalf("Falha ao inicializar o banco de dados: %v", err)
	}
	defer database.DB.Close()

	// Inicia a rotina em background para remover frutas expiradas
	// a cada 1 segundo.
	go handlers.StartExpirationJanitor(1 * time.Second)

	// Configura o roteador Chi
	r := chi.NewRouter()

	// Middlewares para logging e recuperação de panics
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	// Define as rotas da API
	r.Route("/v1", func(r chi.Router) {
		r.Route("/buckets", func(r chi.Router) {
			r.Get("/", handlers.ListBuckets)
			r.Post("/", handlers.CreateBucket)
			r.Delete("/{bucketID}", handlers.DeleteBucket)

			r.Post("/{bucketID}/fruits", handlers.DepositFruit)
			r.Delete("/{bucketID}/fruits/{fruitID}", handlers.RemoveFruitFromBucket)
		})

		r.Route("/fruits", func(r chi.Router) {
			r.Post("/", handlers.CreateFruit)
			r.Delete("/{fruitID}", handlers.DeleteFruit)
		})
	})

	log.Println("Servidor iniciado na porta :8080")
	if err := http.ListenAndServe(":8080", r); err != nil {
		log.Fatalf("Erro ao iniciar o servidor: %v", err)
	}
}
