package main

import (
	"database/sql"
	"embed"
	"io"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/go-chi/chi"
	"github.com/go-chi/cors"
	"github.com/joho/godotenv"

	"github.com/bootdotdev/learn-cicd-starter/internal/database"

	_ "github.com/tursodatabase/libsql-client-go/libsql"
)

type apiConfig struct {
	DB *database.Queries
}

//go:embed static/*
var staticFiles embed.FS

func main() {
	// .env ist optional – bei Fehler nur warnen und weitermachen
	if err := godotenv.Load(".env"); err != nil {
		log.Printf("warning: assuming default configuration. .env unreadable: %v", err)
	}

	port := os.Getenv("PORT")
	if port == "" {
		log.Fatal("PORT environment variable is not set")
	}

	apiCfg := apiConfig{}

	// Optional: DB verbinden, wenn DATABASE_URL gesetzt ist
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		log.Println("DATABASE_URL environment variable is not set")
		log.Println("Running without CRUD endpoints")
	} else {
		db, err := sql.Open("libsql", dbURL)
		if err != nil {
			log.Fatal(err)
		}
		apiCfg.DB = database.New(db)
		log.Println("Connected to database!")
	}

	// Router + CORS
	router := chi.NewRouter()
	router.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"https://*", "http://*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"*"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: false,
		MaxAge:           300,
	}))

	// Root: index.html aus embed FS ausliefern
	router.Get("/", func(w http.ResponseWriter, r *http.Request) {
		f, err := staticFiles.Open("static/index.html")
		if err != nil {
			http.Error(w, "index.html not found", http.StatusInternalServerError)
			return
		}
		defer f.Close()

		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		if _, err := io.Copy(w, f); err != nil {
			log.Printf("write response failed: %v", err)
		}
	})

	// --- v1 API Routen ---
	v1 := chi.NewRouter()

	// Readiness (freie Funktion: (w, r))
	v1.Get("/ready", handlerReadiness)

	// Users:
	// - Create: (w, r) -> direkt
	// - Get:    (w, r, user) -> über Auth-Middleware
	v1.Post("/users", apiCfg.handlerUsersCreate)
	v1.Get("/users", apiCfg.middlewareAuth(apiCfg.handlerUsersGet))

	// Notes: beide (w, r, user) -> über Auth-Middleware
	v1.Group(func(r chi.Router) {
		r.Get("/notes", apiCfg.middlewareAuth(apiCfg.handlerNotesGet))
		r.Post("/notes", apiCfg.middlewareAuth(apiCfg.handlerNotesCreate))
	})

	// Unterpfad mounten
	router.Mount("/v1", v1)

	// HTTP-Server mit Timeouts (Slowloris-/Ressourcenschutz)
	srv := &http.Server{
		Addr:              ":" + port,
		Handler:           router,
		ReadHeaderTimeout: 5 * time.Second,  // wichtig gegen Slowloris
		ReadTimeout:       15 * time.Second, // optional
		WriteTimeout:      15 * time.Second, // optional
		IdleTimeout:       60 * time.Second, // optional
	}

	log.Printf("listening on :%s", port)
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatal(err)
	}
}
