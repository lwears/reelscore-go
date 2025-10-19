package handlers

import (
	"log"
	"net/http"
	"strconv"

	"github.com/liamwears/reelscore/internal/middleware"
	"github.com/liamwears/reelscore/internal/models"
	"github.com/liamwears/reelscore/internal/services"
)

// PageHandler handles page rendering
type PageHandler struct {
	tmdbService  *services.TMDBService
	movieService *services.MovieService
	serieService *services.SerieService
	renderer     *Renderer
	logger       *log.Logger
}

// NewPageHandler creates a new page handler
func NewPageHandler(tmdbService *services.TMDBService, movieService *services.MovieService, serieService *services.SerieService, renderer *Renderer, logger *log.Logger) *PageHandler {
	return &PageHandler{
		tmdbService:  tmdbService,
		movieService: movieService,
		serieService: serieService,
		renderer:     renderer,
		logger:       logger,
	}
}

// BrowseMovies handles GET /movies
func (h *PageHandler) BrowseMovies(w http.ResponseWriter, r *http.Request) {
	// Get user from context
	user, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	// Parse query parameters
	query := r.URL.Query().Get("query")
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if page < 1 {
		page = 1
	}

	// Fetch movies from TMDB
	var movies interface{}
	var totalPages int

	if query != "" {
		// Search movies
		result, err := h.tmdbService.SearchMovies(r.Context(), query, page)
		if err != nil {
			h.logger.Printf("Failed to search movies: %v", err)
			http.Error(w, "Failed to search movies", http.StatusInternalServerError)
			return
		}
		movies = result.Results
		totalPages = result.TotalPages
	} else {
		// Discover popular movies
		result, err := h.tmdbService.DiscoverMovies(r.Context(), page)
		if err != nil {
			h.logger.Printf("Failed to discover movies: %v", err)
			http.Error(w, "Failed to discover movies", http.StatusInternalServerError)
			return
		}
		movies = result.Results
		totalPages = result.TotalPages
	}

	// Render template
	data := map[string]interface{}{
		"User":       user,
		"ActivePage": "movies",
		"Movies":     movies,
		"Query":      query,
		"Page":       page,
		"TotalPages": totalPages,
	}

	h.renderer.RenderPage(w, "browse-movies.html", data)
}

// BrowseSeries handles GET /series
func (h *PageHandler) BrowseSeries(w http.ResponseWriter, r *http.Request) {
	// Get user from context
	user, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	// Parse query parameters
	query := r.URL.Query().Get("query")
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if page < 1 {
		page = 1
	}

	// Fetch series from TMDB
	var series interface{}
	var totalPages int

	if query != "" {
		// Search TV series
		result, err := h.tmdbService.SearchTV(r.Context(), query, page)
		if err != nil {
			h.logger.Printf("Failed to search TV series: %v", err)
			http.Error(w, "Failed to search TV series", http.StatusInternalServerError)
			return
		}
		series = result.Results
		totalPages = result.TotalPages
	} else {
		// Discover popular TV series
		result, err := h.tmdbService.DiscoverTV(r.Context(), page)
		if err != nil {
			h.logger.Printf("Failed to discover TV series: %v", err)
			http.Error(w, "Failed to discover TV series", http.StatusInternalServerError)
			return
		}
		series = result.Results
		totalPages = result.TotalPages
	}

	// Render template
	data := map[string]interface{}{
		"User":       user,
		"ActivePage": "series",
		"Series":     series,
		"Query":      query,
		"Page":       page,
		"TotalPages": totalPages,
	}

	h.renderer.RenderPage(w, "browse-series.html", data)
}

// LibraryMovies handles GET /library/movies/watched and /library/movies/watchlist
func (h *PageHandler) LibraryMovies(w http.ResponseWriter, r *http.Request) {
	// Get user from context
	user, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}
	userID, _ := middleware.GetUserIDFromContext(r.Context())

	// Determine if this is watched or watchlist
	watched := r.PathValue("type") == "watched"

	// Parse query parameters
	query := r.URL.Query().Get("query")
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if page < 1 {
		page = 1
	}

	// Fetch movies from database
	result, err := h.movieService.List(r.Context(), userID, models.ListMoviesInput{
		Watched: watched,
		Query:   query,
		Page:    page,
		Limit:   27,
	})
	if err != nil {
		h.logger.Printf("Failed to list library movies: %v", err)
		http.Error(w, "Failed to fetch movies", http.StatusInternalServerError)
		return
	}

	// Render template
	data := map[string]interface{}{
		"User":       user,
		"ActivePage": "library-movies",
		"Movies":     result.Results,
		"Watched":    watched,
		"Query":      query,
		"Page":       page,
		"TotalPages": result.TotalPages,
	}

	h.renderer.RenderPage(w, "library-movies.html", data)
}

// LibrarySeries handles GET /library/series/watched and /library/series/watchlist
func (h *PageHandler) LibrarySeries(w http.ResponseWriter, r *http.Request) {
	// Get user from context
	user, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}
	userID, _ := middleware.GetUserIDFromContext(r.Context())

	// Determine if this is watched or watchlist
	watched := r.PathValue("type") == "watched"

	// Parse query parameters
	query := r.URL.Query().Get("query")
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if page < 1 {
		page = 1
	}

	// Fetch series from database
	result, err := h.serieService.List(r.Context(), userID, models.ListSeriesInput{
		Watched: watched,
		Query:   query,
		Page:    page,
		Limit:   27,
	})
	if err != nil {
		h.logger.Printf("Failed to list library series: %v", err)
		http.Error(w, "Failed to fetch series", http.StatusInternalServerError)
		return
	}

	// Render template
	data := map[string]interface{}{
		"User":       user,
		"ActivePage": "library-series",
		"Series":     result.Results,
		"Watched":    watched,
		"Query":      query,
		"Page":       page,
		"TotalPages": result.TotalPages,
	}

	h.renderer.RenderPage(w, "library-series.html", data)
}

// Search handles GET /search
func (h *PageHandler) Search(w http.ResponseWriter, r *http.Request) {
	// Get user from context
	user, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	// Parse query parameter
	query := r.URL.Query().Get("query")

	var movies []services.TMDBMovie
	var series []services.TMDBTV

	// Only search if query is provided
	if query != "" {
		// Search movies
		movieResult, err := h.tmdbService.SearchMovies(r.Context(), query, 1)
		if err != nil {
			h.logger.Printf("Failed to search movies: %v", err)
		} else {
			movies = movieResult.Results
			// Limit to top 10 results
			if len(movies) > 10 {
				movies = movies[:10]
			}
		}

		// Search TV series
		seriesResult, err := h.tmdbService.SearchTV(r.Context(), query, 1)
		if err != nil {
			h.logger.Printf("Failed to search TV series: %v", err)
		} else {
			series = seriesResult.Results
			// Limit to top 10 results
			if len(series) > 10 {
				series = series[:10]
			}
		}
	}

	// Render template
	data := map[string]interface{}{
		"User":       user,
		"ActivePage": "search",
		"Query":      query,
		"Movies":     movies,
		"Series":     series,
	}

	h.renderer.RenderPage(w, "search.html", data)
}
