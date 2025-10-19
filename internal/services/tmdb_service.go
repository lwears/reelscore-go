package services

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// TMDBService handles interactions with The Movie Database API
type TMDBService struct {
	client       *http.Client
	apiKey       string
	baseURL      string
	imageBaseURL string
}

// TMDBConfig holds TMDB service configuration
type TMDBConfig struct {
	APIKey       string
	BaseURL      string
	ImageBaseURL string
}

// NewTMDBService creates a new TMDB service
func NewTMDBService(cfg TMDBConfig) *TMDBService {
	return &TMDBService{
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
		apiKey:       cfg.APIKey,
		baseURL:      cfg.BaseURL,
		imageBaseURL: cfg.ImageBaseURL,
	}
}

// TMDBMovie represents a movie from TMDB API
type TMDBMovie struct {
	ID           int     `json:"id"`
	Title        string  `json:"title"`
	PosterPath   *string `json:"poster_path"`
	BackdropPath *string `json:"backdrop_path"`
	ReleaseDate  string  `json:"release_date"`
	VoteAverage  float64 `json:"vote_average"`
	Overview     string  `json:"overview"`
	MediaType    string  `json:"media_type,omitempty"`
}

// TMDBTV represents a TV series from TMDB API
type TMDBTV struct {
	ID           int     `json:"id"`
	Name         string  `json:"name"`
	PosterPath   *string `json:"poster_path"`
	BackdropPath *string `json:"backdrop_path"`
	FirstAirDate string  `json:"first_air_date"`
	VoteAverage  float64 `json:"vote_average"`
	Overview     string  `json:"overview"`
	MediaType    string  `json:"media_type,omitempty"`
}

// TMDBSearchResponse represents a search response from TMDB
type TMDBSearchResponse struct {
	Page         int           `json:"page"`
	Results      []interface{} `json:"results"`
	TotalPages   int           `json:"total_pages"`
	TotalResults int           `json:"total_results"`
}

// TMDBMovieResponse represents a movie search response
type TMDBMovieResponse struct {
	Page         int         `json:"page"`
	Results      []TMDBMovie `json:"results"`
	TotalPages   int         `json:"total_pages"`
	TotalResults int         `json:"total_results"`
}

// TMDBTVResponse represents a TV search response
type TMDBTVResponse struct {
	Page         int      `json:"page"`
	Results      []TMDBTV `json:"results"`
	TotalPages   int      `json:"total_pages"`
	TotalResults int      `json:"total_results"`
}

// doRequest performs an HTTP request to TMDB API
func (s *TMDBService) doRequest(ctx context.Context, endpoint string, params map[string]string) ([]byte, error) {
	url := fmt.Sprintf("%s%s", s.baseURL, endpoint)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Add authorization header
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", s.apiKey))
	req.Header.Set("Content-Type", "application/json")

	// Add query parameters
	q := req.URL.Query()
	q.Add("language", "en-US")
	q.Add("include_adult", "false")
	for key, value := range params {
		q.Add(key, value)
	}
	req.URL.RawQuery = q.Encode()

	resp, err := s.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("TMDB API error: status %d, body: %s", resp.StatusCode, string(body))
	}

	return body, nil
}

// GetMovie retrieves a movie by ID
func (s *TMDBService) GetMovie(ctx context.Context, movieID int) (*TMDBMovie, error) {
	endpoint := fmt.Sprintf("/movie/%d", movieID)
	body, err := s.doRequest(ctx, endpoint, nil)
	if err != nil {
		return nil, err
	}

	var movie TMDBMovie
	if err := json.Unmarshal(body, &movie); err != nil {
		return nil, fmt.Errorf("failed to unmarshal movie: %w", err)
	}

	return &movie, nil
}

// GetTV retrieves a TV series by ID
func (s *TMDBService) GetTV(ctx context.Context, tvID int) (*TMDBTV, error) {
	endpoint := fmt.Sprintf("/tv/%d", tvID)
	body, err := s.doRequest(ctx, endpoint, nil)
	if err != nil {
		return nil, err
	}

	var tv TMDBTV
	if err := json.Unmarshal(body, &tv); err != nil {
		return nil, fmt.Errorf("failed to unmarshal TV series: %w", err)
	}

	return &tv, nil
}

// SearchMulti searches both movies and TV series
func (s *TMDBService) SearchMulti(ctx context.Context, query string, page int) ([]byte, error) {
	if page < 1 {
		page = 1
	}

	params := map[string]string{
		"query": query,
		"page":  fmt.Sprintf("%d", page),
	}

	return s.doRequest(ctx, "/search/multi", params)
}

// SearchMovies searches for movies
func (s *TMDBService) SearchMovies(ctx context.Context, query string, page int) (*TMDBMovieResponse, error) {
	if page < 1 {
		page = 1
	}

	params := map[string]string{
		"query": query,
		"page":  fmt.Sprintf("%d", page),
	}

	body, err := s.doRequest(ctx, "/search/movie", params)
	if err != nil {
		return nil, err
	}

	var response TMDBMovieResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("failed to unmarshal search results: %w", err)
	}

	return &response, nil
}

// SearchTV searches for TV series
func (s *TMDBService) SearchTV(ctx context.Context, query string, page int) (*TMDBTVResponse, error) {
	if page < 1 {
		page = 1
	}

	params := map[string]string{
		"query": query,
		"page":  fmt.Sprintf("%d", page),
	}

	body, err := s.doRequest(ctx, "/search/tv", params)
	if err != nil {
		return nil, err
	}

	var response TMDBTVResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("failed to unmarshal search results: %w", err)
	}

	return &response, nil
}

// DiscoverMovies gets popular/discover movies
func (s *TMDBService) DiscoverMovies(ctx context.Context, page int) (*TMDBMovieResponse, error) {
	if page < 1 {
		page = 1
	}

	params := map[string]string{
		"page":    fmt.Sprintf("%d", page),
		"sort_by": "popularity.desc",
	}

	body, err := s.doRequest(ctx, "/discover/movie", params)
	if err != nil {
		return nil, err
	}

	var response TMDBMovieResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("failed to unmarshal discover results: %w", err)
	}

	return &response, nil
}

// DiscoverTV gets popular/discover TV series
func (s *TMDBService) DiscoverTV(ctx context.Context, page int) (*TMDBTVResponse, error) {
	if page < 1 {
		page = 1
	}

	params := map[string]string{
		"page":    fmt.Sprintf("%d", page),
		"sort_by": "popularity.desc",
	}

	body, err := s.doRequest(ctx, "/discover/tv", params)
	if err != nil {
		return nil, err
	}

	var response TMDBTVResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("failed to unmarshal discover results: %w", err)
	}

	return &response, nil
}

// GetImageURL returns the full URL for an image path
func (s *TMDBService) GetImageURL(path string) string {
	if path == "" {
		return ""
	}
	return s.imageBaseURL + path
}
