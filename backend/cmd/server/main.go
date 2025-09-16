package main

import (
	"log"
	"net/http"
	"os"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/cors"

	ih "github.com/levyxx/pokemon-silhouette-quiz/backend/internal/api"
	"github.com/levyxx/pokemon-silhouette-quiz/backend/internal/poke"
	"github.com/levyxx/pokemon-silhouette-quiz/backend/internal/quiz"
)

func main() {
	r := chi.NewRouter()
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: false,
		MaxAge:           300,
	}))

	// dependencies
	client := poke.NewClient(30 * time.Minute)
	store := quiz.NewStore()
	h := ih.NewHandlers(client, store)

	h.Register(r)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	log.Printf("server listening on :%s", port)
	log.Fatal(http.ListenAndServe(":"+port, r))
}
