package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/carsondecker/JobFinder/internal/auth"
	"github.com/carsondecker/JobFinder/internal/database"
	"github.com/google/uuid"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

type config struct {
	db        *database.Queries
	jwtSecret string
	platform  string
}

type User struct {
	ID        uuid.UUID `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Username  string    `json:"username"`
}

func main() {
	godotenv.Load()
	dbUrl := os.Getenv("DB_URL")
	db, err := sql.Open("postgres", dbUrl)
	if err != nil {
		fmt.Printf("error: %v\n", err)
		os.Exit(1)
	}

	secret := os.Getenv("JWT_SECRET")
	platform := os.Getenv("PLATFORM")

	cfg := &config{
		db:        database.New(db),
		jwtSecret: secret,
		platform:  platform,
	}

	mux := http.NewServeMux()
	server := http.Server{
		Addr:    ":8080",
		Handler: mux,
	}

	mux.HandleFunc("POST /admin/reset", cfg.resetHandler)
	mux.HandleFunc("POST /api/auth/register", cfg.registerHandler)
	mux.HandleFunc("POST /api/auth/login", cfg.loginHandler)

	server.ListenAndServe()
}

func (cfg *config) resetHandler(w http.ResponseWriter, r *http.Request) {
	if cfg.platform != "dev" {
		w.WriteHeader(403)
		return
	}

	err := cfg.db.ResetUsers(r.Context())
	if err != nil {
		respondWithError(w, 500, "could not reset database")
	}

	w.WriteHeader(200)
}

func (cfg *config) registerHandler(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}

	params := parameters{}

	decoder := json.NewDecoder(r.Body)
	defer r.Body.Close()
	if err := decoder.Decode(&params); err != nil {
		respondWithError(w, 400, "could not decode request")
		return
	}

	passwordHash, err := auth.HashPassword(params.Password)
	if err != nil {
		respondWithError(w, 500, "failed to hash password")
	}

	user, err := cfg.db.CreateUser(r.Context(), database.CreateUserParams{
		Username:     params.Username,
		PasswordHash: passwordHash,
	})
	if err != nil {
		respondWithError(w, 400, "could not register user")
		return
	}

	respondWithJSON(w, 201, User(user))
}

func (cfg *config) loginHandler(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}

	params := parameters{}

	decoder := json.NewDecoder(r.Body)
	defer r.Body.Close()
	if err := decoder.Decode(&params); err != nil {
		respondWithError(w, 400, "could not decode request")
		return
	}

	passwordHash, err := auth.HashPassword(params.Password)
	if err != nil {
		respondWithError(w, 500, "could not hash password")
		return
	}

	hashedPassword, err := cfg.db.GetPasswordHashByUsername(r.Context(), params.Username)
	if err != nil {
		respondWithError(w, 500, "could not fetch user's password")
		return
	}

	if err := auth.CheckPasswordHash(passwordHash, hashedPassword); err != nil {
		w.WriteHeader(401)
		return
	}

	user, err := cfg.db.GetUserByUsername(r.Context(), params.Username)
	if err != nil {
		respondWithError(w, 500, "could not fetch user")
	}

	token, err := auth.MakeJWT(user.ID, cfg.jwtSecret, time.Hour)
	if err != nil {
		respondWithError(w, 500, "could not create jwt")
		return
	}

	type UserWithToken struct {
		ID        uuid.UUID `json:"id"`
		CreatedAt time.Time `json:"created_at"`
		UpdatedAt time.Time `json:"updated_at"`
		Username  string    `json:"username"`
		Token     string    `json:"token"`
	}

	respondWithJSON(w, 200, UserWithToken{
		ID:        user.ID,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
		Username:  user.Username,
		Token:     token,
	})
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
