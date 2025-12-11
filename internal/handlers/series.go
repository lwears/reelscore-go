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

// SerieHandler handles series-related requests
type SerieHandler struct {
	serieService *services.SerieService
	logger       *log.Logger
}

// NewSerieHandler creates a new serie handler
func NewSerieHandler(serieService *services.SerieService, logger *log.Logger) *SerieHandler {
	return &SerieHandler{
		serieService: serieService,
		logger:       logger,
	}
}

// List handles GET /api/series
func (h *SerieHandler) List(w http.ResponseWriter, r *http.Request) {
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
	result, err := h.serieService.List(r.Context(), userID, models.ListSeriesInput{
		Watched: watched,
		Query:   searchQuery,
		Page:    page,
		Limit:   limit,
	})
	if err != nil {
		h.logger.Printf("Failed to list series: %v", err)
		http.Error(w, `{"error":"Failed to fetch series"}`, http.StatusInternalServerError)
		return
	}

	// Return JSON
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

// Create handles POST /api/series
func (h *SerieHandler) Create(w http.ResponseWriter, r *http.Request) {
	// Get user from context
	userID, ok := middleware.GetUserIDFromContext(r.Context())
	if !ok {
		http.Error(w, `{"error":"Unauthorized"}`, http.StatusUnauthorized)
		return
	}

	// Parse request body
	var input models.CreateSerieInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, `{"error":"Invalid request body"}`, http.StatusBadRequest)
		return
	}

	// Call service
	serie, err := h.serieService.Create(r.Context(), userID, input)
	if err != nil {
		h.logger.Printf("Failed to create serie: %v", err)
		// Check for duplicate
		if err.Error() == "duplicate key value violates unique constraint" {
			http.Error(w, `{"error":"Serie already in your library"}`, http.StatusConflict)
			return
		}
		http.Error(w, `{"error":"Failed to create serie"}`, http.StatusInternalServerError)
		return
	}

	// Return created serie with success message
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"serie":   serie,
		"message": "Series added to your library!",
	})
}

// Get handles GET /api/series/{id}
func (h *SerieHandler) Get(w http.ResponseWriter, r *http.Request) {
	// Get user from context
	userID, ok := middleware.GetUserIDFromContext(r.Context())
	if !ok {
		http.Error(w, `{"error":"Unauthorized"}`, http.StatusUnauthorized)
		return
	}

	// Get serie ID from path
	idStr := r.PathValue("id")
	serieID, err := uuid.Parse(idStr)
	if err != nil {
		http.Error(w, `{"error":"Invalid serie ID"}`, http.StatusBadRequest)
		return
	}

	// Call service
	serie, err := h.serieService.Get(r.Context(), serieID, userID)
	if err != nil {
		if err == pgx.ErrNoRows {
			http.Error(w, `{"error":"Serie not found"}`, http.StatusNotFound)
			return
		}
		h.logger.Printf("Failed to get serie: %v", err)
		http.Error(w, `{"error":"Failed to fetch serie"}`, http.StatusInternalServerError)
		return
	}

	// Return serie
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(serie)
}

// Update handles PATCH /api/series/{id}
func (h *SerieHandler) Update(w http.ResponseWriter, r *http.Request) {
	// Get user from context
	userID, ok := middleware.GetUserIDFromContext(r.Context())
	if !ok {
		http.Error(w, `{"error":"Unauthorized"}`, http.StatusUnauthorized)
		return
	}

	// Get serie ID from path
	idStr := r.PathValue("id")
	serieID, err := uuid.Parse(idStr)
	if err != nil {
		http.Error(w, `{"error":"Invalid serie ID"}`, http.StatusBadRequest)
		return
	}

	// Parse request body
	var input models.UpdateSerieInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, `{"error":"Invalid request body"}`, http.StatusBadRequest)
		return
	}
	input.ID = serieID

	// Call service
	serie, err := h.serieService.Update(r.Context(), userID, input)
	if err != nil {
		if err == pgx.ErrNoRows {
			http.Error(w, `{"error":"Serie not found"}`, http.StatusNotFound)
			return
		}
		h.logger.Printf("Failed to update serie: %v", err)
		http.Error(w, `{"error":"Failed to update serie"}`, http.StatusInternalServerError)
		return
	}

	// Return updated serie
	if input.Watched != nil && *input.Watched {
		w.Header().Set("HX-Trigger", `{"showMessage":"Series marked as watched"}`)
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(serie)
}

// Delete handles DELETE /api/series/{id}
func (h *SerieHandler) Delete(w http.ResponseWriter, r *http.Request) {
	// Get user from context
	userID, ok := middleware.GetUserIDFromContext(r.Context())
	if !ok {
		http.Error(w, `{"error":"Unauthorized"}`, http.StatusUnauthorized)
		return
	}

	// Get serie ID from path
	idStr := r.PathValue("id")
	serieID, err := uuid.Parse(idStr)
	if err != nil {
		http.Error(w, `{"error":"Invalid serie ID"}`, http.StatusBadRequest)
		return
	}

	// Call service
	err = h.serieService.Delete(r.Context(), serieID, userID)
	if err != nil {
		if err == pgx.ErrNoRows {
			http.Error(w, `{"error":"Serie not found"}`, http.StatusNotFound)
			return
		}
		h.logger.Printf("Failed to delete serie: %v", err)
		http.Error(w, `{"error":"Failed to delete serie"}`, http.StatusInternalServerError)
		return
	}

	// Return success
	w.Header().Set("HX-Trigger", `{"showMessage":"Series deleted"}`)
	w.WriteHeader(http.StatusNoContent)
}
