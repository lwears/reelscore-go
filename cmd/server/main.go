package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/liamwears/reelscore/internal/config"
	"github.com/liamwears/reelscore/internal/database"
	"github.com/liamwears/reelscore/internal/handlers"
	"github.com/liamwears/reelscore/internal/middleware"
	"github.com/liamwears/reelscore/internal/services"
)

func main() {
	// Check for migrate command
	if len(os.Args) > 1 && os.Args[1] == "migrate" {
		runMigrations()
		return
	}

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Initialize logger
	logger := log.New(os.Stdout, "[reelscore] ", log.LstdFlags|log.Lshortfile)
	logger.Printf("Starting ReelScore server in %s mode", cfg.Server.Env)

	// Initialize database connection
	db, err := database.New(database.Config{
		URL: cfg.Database.URL,
	})
	if err != nil {
		logger.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// Initialize Redis connection
	redisClient, err := database.NewRedisClient(database.RedisConfig{
		Addr:     cfg.RedisAddr(),
		Password: cfg.Redis.Password,
		DB:       0,
		TLS:      cfg.Redis.TLS,
	})
	if err != nil {
		logger.Fatalf("Failed to connect to Redis: %v", err)
	}
	defer redisClient.Close()

	// Initialize session store
	sessionStore := database.NewSessionStore(redisClient, 7*24*time.Hour)

	// Initialize services
	userService := services.NewUserService(db.Pool)
	movieService := services.NewMovieService(db.Pool)
	serieService := services.NewSerieService(db.Pool)
	tmdbService := services.NewTMDBService(services.TMDBConfig{
		APIKey:       cfg.TMDB.APIKey,
		BaseURL:      "https://api.themoviedb.org/3",
		ImageBaseURL: "https://image.tmdb.org/t/p/w500",
	})

	// Initialize middleware
	authMiddleware := middleware.NewAuthMiddleware(sessionStore, userService, "session", cfg.IsProduction())

	// Initialize rate limiter (100 req/min in production, unlimited in local/dev)
	maxRequests := 1000 // High limit for local/dev
	if cfg.IsProduction() {
		maxRequests = 100
	}
	rateLimiter := middleware.NewRateLimiter(redisClient.Client, maxRequests, time.Minute, cfg.IsProduction())

	// Initialize renderer
	renderer, err := handlers.NewRenderer(logger)
	if err != nil {
		logger.Fatalf("Failed to initialize renderer: %v", err)
	}

	// Initialize handlers
	authHandler := handlers.NewAuthHandler(
		userService,
		sessionStore,
		authMiddleware,
		renderer,
		handlers.AuthConfig{
			GoogleClientID:     cfg.OAuth.GoogleClientID,
			GoogleClientSecret: cfg.OAuth.GoogleClientSecret,
			GitHubClientID:     cfg.OAuth.GitHubClientID,
			GitHubClientSecret: cfg.OAuth.GitHubClientSecret,
			CallbackHost:       cfg.OAuth.CallbackHost,
		},
		logger,
	)
	movieHandler := handlers.NewMovieHandler(movieService, logger)
	serieHandler := handlers.NewSerieHandler(serieService, logger)
	tmdbHandler := handlers.NewTMDBHandler(tmdbService, logger)
	pageHandler := handlers.NewPageHandler(tmdbService, movieService, serieService, renderer, logger)

	// Set up HTTP router with logging
	mux := http.NewServeMux()

	// Auth routes (public)
	mux.HandleFunc("/login", authHandler.Login)
	mux.HandleFunc("/auth/google/login", authHandler.GoogleLogin)
	mux.HandleFunc("/auth/google/callback", authHandler.GoogleCallback)
	mux.HandleFunc("/auth/github/login", authHandler.GitHubLogin)
	mux.HandleFunc("/auth/github/callback", authHandler.GitHubCallback)
	mux.HandleFunc("/auth/logout", authHandler.Logout)

	// Page routes (protected)
	mux.Handle("/movies", authMiddleware.RequireAuth(http.HandlerFunc(pageHandler.BrowseMovies)))
	mux.Handle("/series", authMiddleware.RequireAuth(http.HandlerFunc(pageHandler.BrowseSeries)))
	mux.Handle("/search", authMiddleware.RequireAuth(http.HandlerFunc(pageHandler.Search)))
	mux.Handle("/library/movies/{type}", authMiddleware.RequireAuth(http.HandlerFunc(pageHandler.LibraryMovies)))
	mux.Handle("/library/series/{type}", authMiddleware.RequireAuth(http.HandlerFunc(pageHandler.LibrarySeries)))

	// Movie API routes (protected with auth and rate limiting)
	mux.Handle("GET /api/movies", rateLimiter.Limit(authMiddleware.RequireAuthAPI(http.HandlerFunc(movieHandler.List))))
	mux.Handle("POST /api/movies", rateLimiter.Limit(authMiddleware.RequireAuthAPI(http.HandlerFunc(movieHandler.Create))))
	mux.Handle("GET /api/movies/{id}", rateLimiter.Limit(authMiddleware.RequireAuthAPI(http.HandlerFunc(movieHandler.Get))))
	mux.Handle("PATCH /api/movies/{id}", rateLimiter.Limit(authMiddleware.RequireAuthAPI(http.HandlerFunc(movieHandler.Update))))
	mux.Handle("DELETE /api/movies/{id}", rateLimiter.Limit(authMiddleware.RequireAuthAPI(http.HandlerFunc(movieHandler.Delete))))

	// Serie API routes (protected with auth and rate limiting)
	mux.Handle("GET /api/series", rateLimiter.Limit(authMiddleware.RequireAuthAPI(http.HandlerFunc(serieHandler.List))))
	mux.Handle("POST /api/series", rateLimiter.Limit(authMiddleware.RequireAuthAPI(http.HandlerFunc(serieHandler.Create))))
	mux.Handle("GET /api/series/{id}", rateLimiter.Limit(authMiddleware.RequireAuthAPI(http.HandlerFunc(serieHandler.Get))))
	mux.Handle("PATCH /api/series/{id}", rateLimiter.Limit(authMiddleware.RequireAuthAPI(http.HandlerFunc(serieHandler.Update))))
	mux.Handle("DELETE /api/series/{id}", rateLimiter.Limit(authMiddleware.RequireAuthAPI(http.HandlerFunc(serieHandler.Delete))))

	// TMDB API routes (protected with auth and rate limiting)
	mux.Handle("GET /api/tmdb/movie/{id}", rateLimiter.Limit(authMiddleware.RequireAuthAPI(http.HandlerFunc(tmdbHandler.GetMovie))))
	mux.Handle("GET /api/tmdb/tv/{id}", rateLimiter.Limit(authMiddleware.RequireAuthAPI(http.HandlerFunc(tmdbHandler.GetTV))))
	mux.Handle("GET /api/tmdb/search/multi", rateLimiter.Limit(authMiddleware.RequireAuthAPI(http.HandlerFunc(tmdbHandler.SearchMulti))))
	mux.Handle("GET /api/tmdb/search/movie", rateLimiter.Limit(authMiddleware.RequireAuthAPI(http.HandlerFunc(tmdbHandler.SearchMovies))))
	mux.Handle("GET /api/tmdb/search/tv", rateLimiter.Limit(authMiddleware.RequireAuthAPI(http.HandlerFunc(tmdbHandler.SearchTV))))
	mux.Handle("GET /api/tmdb/discover/movie", rateLimiter.Limit(authMiddleware.RequireAuthAPI(http.HandlerFunc(tmdbHandler.DiscoverMovies))))
	mux.Handle("GET /api/tmdb/discover/tv", rateLimiter.Limit(authMiddleware.RequireAuthAPI(http.HandlerFunc(tmdbHandler.DiscoverTV))))

	// Serve static files
	fs := http.FileServer(http.Dir("internal/static"))
	mux.Handle("/static/", http.StripPrefix("/static/", fs))

	// Wrap with logging middleware
	handler := middleware.Logger(logger)(mux)

	// Health check endpoint
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		// Check database health
		dbErr := db.Health(r.Context())
		redisErr := redisClient.Health(r.Context())

		if dbErr != nil || redisErr != nil {
			w.WriteHeader(http.StatusServiceUnavailable)
			dbStatus := "up"
			if dbErr != nil {
				dbStatus = "down"
			}
			redisStatus := "up"
			if redisErr != nil {
				redisStatus = "down"
			}
			fmt.Fprintf(w, `{"status":"unhealthy","database":"%s","redis":"%s"}`, dbStatus, redisStatus)
			return
		}

		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, `{"status":"ok","database":"up","redis":"up"}`)
	})

	// Placeholder for other routes
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		fmt.Fprintf(w, "ReelScore API - Coming soon!")
	})

	// Create HTTP server
	addr := fmt.Sprintf(":%s", cfg.Server.Port)
	srv := &http.Server{
		Addr:         addr,
		Handler:      handler,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start server in a goroutine
	go func() {
		logger.Printf("Server listening on %s", addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatalf("Server failed to start: %v", err)
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Println("Shutting down server...")

	// Graceful shutdown with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		logger.Fatalf("Server forced to shutdown: %v", err)
	}

	// Close connections
	db.Close()
	redisClient.Close()

	logger.Println("Server exited")
}

// runMigrations runs database migrations
func runMigrations() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	db, err := database.New(database.Config{
		URL: cfg.Database.URL,
	})
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	migrator := database.NewMigrator(db.Pool)

	ctx := context.Background()
	if err := migrator.Up(ctx); err != nil {
		log.Fatalf("Failed to run migrations: %v", err)
	}

	log.Println("Migrations completed successfully")
}
