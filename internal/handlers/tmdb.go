package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"

	"github.com/liamwears/reelscore/internal/services"
)

// TMDBHandler handles TMDB API requests
type TMDBHandler struct {
	tmdbService *services.TMDBService
	logger      *log.Logger
}

// NewTMDBHandler creates a new TMDB handler
func NewTMDBHandler(tmdbService *services.TMDBService, logger *log.Logger) *TMDBHandler {
	return &TMDBHandler{
		tmdbService: tmdbService,
		logger:      logger,
	}
}

// GetMovie handles GET /api/tmdb/movie/{id}
func (h *TMDBHandler) GetMovie(w http.ResponseWriter, r *http.Request) {
	// Get movie ID from path
	idStr := r.PathValue("id")
	movieID, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, `{"error":"Invalid movie ID"}`, http.StatusBadRequest)
		return
	}

	// Call TMDB service
	movie, err := h.tmdbService.GetMovie(r.Context(), movieID)
	if err != nil {
		h.logger.Printf("Failed to fetch movie from TMDB: %v", err)
		http.Error(w, `{"error":"Failed to fetch movie"}`, http.StatusInternalServerError)
		return
	}

	// Return movie
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(movie)
}

// GetTV handles GET /api/tmdb/tv/{id}
func (h *TMDBHandler) GetTV(w http.ResponseWriter, r *http.Request) {
	// Get TV ID from path
	idStr := r.PathValue("id")
	tvID, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, `{"error":"Invalid TV ID"}`, http.StatusBadRequest)
		return
	}

	// Call TMDB service
	tv, err := h.tmdbService.GetTV(r.Context(), tvID)
	if err != nil {
		h.logger.Printf("Failed to fetch TV from TMDB: %v", err)
		http.Error(w, `{"error":"Failed to fetch TV series"}`, http.StatusInternalServerError)
		return
	}

	// Return TV series
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(tv)
}

// SearchMulti handles GET /api/tmdb/search/multi
func (h *TMDBHandler) SearchMulti(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query().Get("query")
	if query == "" {
		http.Error(w, `{"error":"Query parameter is required"}`, http.StatusBadRequest)
		return
	}

	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if page < 1 {
		page = 1
	}

	// Call TMDB service
	result, err := h.tmdbService.SearchMulti(r.Context(), query, page)
	if err != nil {
		h.logger.Printf("Failed to search TMDB: %v", err)
		http.Error(w, `{"error":"Failed to search"}`, http.StatusInternalServerError)
		return
	}

	// Return raw JSON from TMDB
	w.Header().Set("Content-Type", "application/json")
	w.Write(result)
}

// SearchMovies handles GET /api/tmdb/search/movie
func (h *TMDBHandler) SearchMovies(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query().Get("query")
	if query == "" {
		http.Error(w, `{"error":"Query parameter is required"}`, http.StatusBadRequest)
		return
	}

	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if page < 1 {
		page = 1
	}

	// Call TMDB service
	result, err := h.tmdbService.SearchMovies(r.Context(), query, page)
	if err != nil {
		h.logger.Printf("Failed to search movies: %v", err)
		http.Error(w, `{"error":"Failed to search movies"}`, http.StatusInternalServerError)
		return
	}

	// Return results
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

// SearchTV handles GET /api/tmdb/search/tv
func (h *TMDBHandler) SearchTV(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query().Get("query")
	if query == "" {
		http.Error(w, `{"error":"Query parameter is required"}`, http.StatusBadRequest)
		return
	}

	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if page < 1 {
		page = 1
	}

	// Call TMDB service
	result, err := h.tmdbService.SearchTV(r.Context(), query, page)
	if err != nil {
		h.logger.Printf("Failed to search TV: %v", err)
		http.Error(w, `{"error":"Failed to search TV series"}`, http.StatusInternalServerError)
		return
	}

	// Return results
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

// DiscoverMovies handles GET /api/tmdb/discover/movie
func (h *TMDBHandler) DiscoverMovies(w http.ResponseWriter, r *http.Request) {
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if page < 1 {
		page = 1
	}

	// Call TMDB service
	result, err := h.tmdbService.DiscoverMovies(r.Context(), page)
	if err != nil {
		h.logger.Printf("Failed to discover movies: %v", err)
		http.Error(w, `{"error":"Failed to discover movies"}`, http.StatusInternalServerError)
		return
	}

	// Return results
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

// DiscoverTV handles GET /api/tmdb/discover/tv
func (h *TMDBHandler) DiscoverTV(w http.ResponseWriter, r *http.Request) {
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if page < 1 {
		page = 1
	}

	// Call TMDB service
	result, err := h.tmdbService.DiscoverTV(r.Context(), page)
	if err != nil {
		h.logger.Printf("Failed to discover TV: %v", err)
		http.Error(w, `{"error":"Failed to discover TV series"}`, http.StatusInternalServerError)
		return
	}

	// Return results
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}
