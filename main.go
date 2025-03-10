package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	"github.com/carsondecker/JobFinder/internal/database"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

type config struct {
	db *database.Queries
}

func main() {
	godotenv.Load()
	dbUrl := os.Getenv("DB_URL")
	db, err := sql.Open("postgres", dbUrl)
	if err != nil {
		fmt.Printf("error: %v\n", err)
		os.Exit(1)
	}

	cfg := &config{
		db: database.New(db),
	}

	mux := http.NewServeMux()
	server := http.Server{
		Addr:    ":8080",
		Handler: mux,
	}

	mux.HandleFunc("GET /", cfg.testHandler)

	server.ListenAndServe()
}

func (cfg *config) testHandler(w http.ResponseWriter, r *http.Request) {
	type test struct {
		Test string `json:"test"`
	}

	respondWithJSON(w, 200, test{Test: "this is a test."})
}

func respondWithError(w http.ResponseWriter, code int, msg string) {
	type resError struct {
		Error string `json:"error"`
	}
	respondWithJSON(w, code, resError{Error: msg})
}

func respondWithJSON(w http.ResponseWriter, code int, body interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	data, err := json.Marshal(body)
	if err != nil {
		return
	}
	w.Write(data)
}
