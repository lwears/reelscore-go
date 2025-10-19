package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/liamwears/reelscore/internal/middleware"
	"github.com/liamwears/reelscore/internal/models"
	"github.com/liamwears/reelscore/internal/services"
)

// MovieHandler handles movie-related requests
type MovieHandler struct {
	movieService *services.MovieService
	logger       *log.Logger
}

// NewMovieHandler creates a new movie handler
func NewMovieHandler(movieService *services.MovieService, logger *log.Logger) *MovieHandler {
	return &MovieHandler{
		movieService: movieService,
		logger:       logger,
	}
}

// List handles GET /api/movies
func (h *MovieHandler) List(w http.ResponseWriter, r *http.Request) {
	// Get user from context
	userID, ok := middleware.GetUserIDFromContext(r.Context())
	if !ok {
		http.Error(w, `{"error":"Unauthorized"}`, http.StatusUnauthorized)
		return
	}

	// Parse query parameters
	query := r.URL.Query()

	watched := query.Get("watched") == "true"
	searchQuery := query.Get("query")

	page, _ := strconv.Atoi(query.Get("page"))
	if page < 1 {
		page = 1
	}

	limit, _ := strconv.Atoi(query.Get("limit"))
	if limit < 1 || limit > 100 {
		limit = 27
	}

	// Call service
	result, err := h.movieService.List(r.Context(), userID, models.ListMoviesInput{
		Watched: watched,
		Query:   searchQuery,
		Page:    page,
		Limit:   limit,
	})
	if err != nil {
		h.logger.Printf("Failed to list movies: %v", err)
		http.Error(w, `{"error":"Failed to fetch movies"}`, http.StatusInternalServerError)
		return
	}

	// Return JSON
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

// Create handles POST /api/movies
func (h *MovieHandler) Create(w http.ResponseWriter, r *http.Request) {
	// Get user from context
	userID, ok := middleware.GetUserIDFromContext(r.Context())
	if !ok {
		http.Error(w, `{"error":"Unauthorized"}`, http.StatusUnauthorized)
		return
	}

	// Parse request body
	var input models.CreateMovieInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		h.logger.Printf("Failed to decode request body: %v", err)
		http.Error(w, `{"error":"Invalid request body"}`, http.StatusBadRequest)
		return
	}
	h.logger.Printf("Received movie input: %+v", input)

	// Call service
	movie, err := h.movieService.Create(r.Context(), userID, input)
	if err != nil {
		h.logger.Printf("Failed to create movie: %v", err)
		// Check for duplicate
		if err.Error() == "duplicate key value violates unique constraint" {
			http.Error(w, `{"error":"Movie already in your library"}`, http.StatusConflict)
			return
		}
		http.Error(w, `{"error":"Failed to create movie"}`, http.StatusInternalServerError)
		return
	}

	// Return created movie with success message
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"movie":   movie,
		"message": "Movie added to your library!",
	})
}

// Get handles GET /api/movies/{id}
func (h *MovieHandler) Get(w http.ResponseWriter, r *http.Request) {
	// Get user from context
	userID, ok := middleware.GetUserIDFromContext(r.Context())
	if !ok {
		http.Error(w, `{"error":"Unauthorized"}`, http.StatusUnauthorized)
		return
	}

	// Get movie ID from path
	idStr := r.PathValue("id")
	movieID, err := uuid.Parse(idStr)
	if err != nil {
		http.Error(w, `{"error":"Invalid movie ID"}`, http.StatusBadRequest)
		return
	}

	// Call service
	movie, err := h.movieService.Get(r.Context(), movieID, userID)
	if err != nil {
		if err == pgx.ErrNoRows {
			http.Error(w, `{"error":"Movie not found"}`, http.StatusNotFound)
			return
		}
		h.logger.Printf("Failed to get movie: %v", err)
		http.Error(w, `{"error":"Failed to fetch movie"}`, http.StatusInternalServerError)
		return
	}

	// Return movie
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(movie)
}

// Update handles PATCH /api/movies/{id}
func (h *MovieHandler) Update(w http.ResponseWriter, r *http.Request) {
	// Get user from context
	userID, ok := middleware.GetUserIDFromContext(r.Context())
	if !ok {
		http.Error(w, `{"error":"Unauthorized"}`, http.StatusUnauthorized)
		return
	}

	// Get movie ID from path
	idStr := r.PathValue("id")
	movieID, err := uuid.Parse(idStr)
	if err != nil {
		http.Error(w, `{"error":"Invalid movie ID"}`, http.StatusBadRequest)
		return
	}

	// Parse request body
	var input models.UpdateMovieInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, `{"error":"Invalid request body"}`, http.StatusBadRequest)
		return
	}
	input.ID = movieID

	// Call service
	movie, err := h.movieService.Update(r.Context(), userID, input)
	if err != nil {
		if err == pgx.ErrNoRows {
			http.Error(w, `{"error":"Movie not found"}`, http.StatusNotFound)
			return
		}
		h.logger.Printf("Failed to update movie: %v", err)
		http.Error(w, `{"error":"Failed to update movie"}`, http.StatusInternalServerError)
		return
	}

	// Return updated movie
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(movie)
}

// Delete handles DELETE /api/movies/{id}
func (h *MovieHandler) Delete(w http.ResponseWriter, r *http.Request) {
	// Get user from context
	userID, ok := middleware.GetUserIDFromContext(r.Context())
	if !ok {
		http.Error(w, `{"error":"Unauthorized"}`, http.StatusUnauthorized)
		return
	}

	// Get movie ID from path
	idStr := r.PathValue("id")
	movieID, err := uuid.Parse(idStr)
	if err != nil {
		http.Error(w, `{"error":"Invalid movie ID"}`, http.StatusBadRequest)
		return
	}

	// Call service
	err = h.movieService.Delete(r.Context(), movieID, userID)
	if err != nil {
		if err == pgx.ErrNoRows {
			http.Error(w, `{"error":"Movie not found"}`, http.StatusNotFound)
			return
		}
		h.logger.Printf("Failed to delete movie: %v", err)
		http.Error(w, `{"error":"Failed to delete movie"}`, http.StatusInternalServerError)
		return
	}

	// Return success
	w.WriteHeader(http.StatusNoContent)
}
