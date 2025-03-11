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

type Job struct {
	ID          uuid.UUID `json:"id"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	City        string    `json:"city"`
	UserID      uuid.UUID `json:"user_id"`
}

type Application struct {
	ID        uuid.UUID `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	CoverNote string    `json:"cover_note"`
	JobID     uuid.UUID `json:"job_id"`
	UserID    uuid.UUID `json:"user_id"`
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
	mux.HandleFunc("POST /api/jobs", cfg.createJobHandler)
	mux.HandleFunc("GET /api/jobs", cfg.getJobsHandler)
	mux.HandleFunc("GET /api/jobs/{jobID}", cfg.getJobByIDHandler)
	mux.HandleFunc("DELETE /api/jobs/{jobID}", cfg.deleteJobHandler)
	mux.HandleFunc("POST /api/jobs/{jobID}/apply", cfg.applyHandler)
	mux.HandleFunc("GET /api/application", cfg.getApplicationByUserHandler)
	mux.HandleFunc("GET /api/jobs/{jobID}/applications", cfg.getApplicationsByJobHandler)

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

	hashedPassword, err := cfg.db.GetPasswordHashByUsername(r.Context(), params.Username)
	if err != nil {
		respondWithError(w, 500, "could not fetch user's password")
		return
	}

	if err := auth.CheckPasswordHash(params.Password, hashedPassword); err != nil {
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

func (cfg *config) createJobHandler(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		Title       string `json:"title"`
		Description string `json:"description"`
		City        string `json:"city"`
	}

	params := parameters{}

	decoder := json.NewDecoder(r.Body)
	defer r.Body.Close()
	if err := decoder.Decode(&params); err != nil {
		respondWithError(w, 500, "Failed to decode json")
		return
	}

	token, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, 500, err.Error())
		return
	}

	id, err := auth.ValidateJWT(token, cfg.jwtSecret)
	if err != nil {
		respondWithError(w, 401, "invalid token")
		return
	}

	job, err := cfg.db.CreateJob(r.Context(), database.CreateJobParams{
		Title:       params.Title,
		Description: params.Description,
		City:        params.City,
		UserID:      id,
	})
	if err != nil {
		respondWithError(w, 500, "could not create job")
		return
	}

	respondWithJSON(w, 201, Job(job))
}

func (cfg *config) getJobsHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Query().Get("title") != "" {
		cfg.getJobsByTitleHandler(w, r)
		return
	}

	jobs, err := cfg.db.GetJobs(r.Context())
	if err != nil {
		respondWithError(w, 500, "could not get jobs")
		return
	}

	convertedJobs := make([]Job, len(jobs))
	for i, job := range jobs {
		convertedJobs[i] = Job(job)
	}

	respondWithJSON(w, 200, convertedJobs)
}

func (cfg *config) getJobByIDHandler(w http.ResponseWriter, r *http.Request) {
	jobID, err := uuid.Parse(r.PathValue("jobID"))
	if err != nil {
		respondWithError(w, 400, "could not get job id from url")
		return
	}

	job, err := cfg.db.GetJobByID(r.Context(), jobID)
	if err != nil {
		respondWithError(w, 400, "could not get job")
		return
	}

	respondWithJSON(w, 200, Job(job))
}

func (cfg *config) getJobsByTitleHandler(w http.ResponseWriter, r *http.Request) {
	title := r.URL.Query().Get("title")

	jobs, err := cfg.db.GetJobsByTitle(r.Context(), title)
	if err != nil {
		respondWithError(w, 500, "could not get jobs")
		return
	}

	convertedJobs := make([]Job, len(jobs))
	for i, job := range jobs {
		convertedJobs[i] = Job(job)
	}

	respondWithJSON(w, 200, convertedJobs)
}

func (cfg *config) deleteJobHandler(w http.ResponseWriter, r *http.Request) {
	jobID, err := uuid.Parse(r.PathValue("jobID"))
	if err != nil {
		respondWithError(w, 400, "could not parse job id")
		return
	}

	token, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, 500, err.Error())
		return
	}

	id, err := auth.ValidateJWT(token, cfg.jwtSecret)
	if err != nil {
		respondWithError(w, 401, "invalid token")
		return
	}

	job, err := cfg.db.GetJobByID(r.Context(), jobID)
	if err != nil {
		respondWithError(w, 400, "could not get job")
		return
	}

	if id != job.UserID {
		w.WriteHeader(401)
		return
	}

	if err = cfg.db.DeleteJob(r.Context(), jobID); err != nil {
		respondWithError(w, 500, "could not delete job")
		return
	}

	w.WriteHeader(200)
}

func (cfg *config) applyHandler(w http.ResponseWriter, r *http.Request) {
	jobID, err := uuid.Parse(r.PathValue("jobID"))
	if err != nil {
		respondWithError(w, 400, "could not parse job id")
		return
	}

	type parameters struct {
		CoverNote string `json:"cover_note"`
	}

	params := parameters{}

	decoder := json.NewDecoder(r.Body)
	defer r.Body.Close()
	if err := decoder.Decode(&params); err != nil {
		respondWithError(w, 500, "Failed to decode json")
		return
	}

	token, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, 500, err.Error())
		return
	}

	userID, err := auth.ValidateJWT(token, cfg.jwtSecret)
	if err != nil {
		respondWithError(w, 401, "invalid token")
		return
	}

	job, err := cfg.db.GetJobByID(r.Context(), jobID)
	if err != nil {
		respondWithError(w, 400, "could not get job")
		return
	}

	if userID == job.UserID {
		respondWithError(w, 400, "cannot apply to own job")
		return
	}

	application, err := cfg.db.CreateApplication(r.Context(), database.CreateApplicationParams{
		CoverNote: params.CoverNote,
		JobID:     job.ID,
		UserID:    userID,
	})

	respondWithJSON(w, 201, Application(application))
}

func (cfg *config) getApplicationByUserHandler(w http.ResponseWriter, r *http.Request) {
	token, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, 500, err.Error())
		return
	}

	id, err := auth.ValidateJWT(token, cfg.jwtSecret)
	if err != nil {
		respondWithError(w, 401, "invalid token")
		return
	}

	applications, err := cfg.db.GetApplicationsByUserID(r.Context(), id)
	if err != nil {
		respondWithError(w, 500, "could not get applications")
		return
	}

	convertedApps := make([]Application, len(applications))
	for i, app := range applications {
		convertedApps[i] = Application(app)
	}

	respondWithJSON(w, 200, convertedApps)
}

func (cfg *config) getApplicationsByJobHandler(w http.ResponseWriter, r *http.Request) {
	jobID, err := uuid.Parse(r.PathValue("jobID"))
	if err != nil {
		respondWithError(w, 400, "could not parse job id")
		return
	}

	token, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, 500, err.Error())
		return
	}

	userID, err := auth.ValidateJWT(token, cfg.jwtSecret)
	if err != nil {
		respondWithError(w, 401, "invalid token")
		return
	}

	job, err := cfg.db.GetJobByID(r.Context(), jobID)
	if err != nil {
		respondWithError(w, 400, "could not get job")
		return
	}

	if userID != job.UserID {
		w.WriteHeader(401)
		return
	}

	applications, err := cfg.db.GetApplicationsByJobID(r.Context(), jobID)
	if err != nil {
		respondWithError(w, 500, "could not get applications")
		return
	}

	convertedApps := make([]Application, len(applications))
	for i, app := range applications {
		convertedApps[i] = Application(app)
	}

	respondWithJSON(w, 200, convertedApps)
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
