package models

import (
	"time"

	"github.com/google/uuid"
)

// Movie represents a movie in the user's library
type Movie struct {
	ID          uuid.UUID  `db:"id" json:"id"`
	TmdbID      int        `db:"tmdbId" json:"tmdbId"`
	CreatedAt   time.Time  `db:"createdAt" json:"createdAt"`
	UpdatedAt   time.Time  `db:"updatedAt" json:"updatedAt"`
	Title       string     `db:"title" json:"title"`
	PosterPath  *string    `db:"posterPath" json:"posterPath"`
	ReleaseDate *time.Time `db:"releaseDate" json:"releaseDate"`
	TmdbScore   float64    `db:"tmdbScore" json:"tmdbScore"`
	Score       float64    `db:"score" json:"score"`
	Watched     bool       `db:"watched" json:"watched"`
	UserID      uuid.UUID  `db:"userId" json:"userId"`
}

// CreateMovieInput represents the input for creating a movie
type CreateMovieInput struct {
	TmdbID      int      `json:"tmdbId" validate:"required"`
	Title       string   `json:"title" validate:"required"`
	PosterPath  *string  `json:"posterPath"`
	ReleaseDate *string  `json:"releaseDate"`
	Watched     bool     `json:"watched"`
	TmdbScore   float64  `json:"tmdbScore" validate:"min=0,max=10"`
	Score       *float64 `json:"score,omitempty" validate:"omitempty,min=0,max=10"`
}

// UpdateMovieInput represents the input for updating a movie
type UpdateMovieInput struct {
	ID      uuid.UUID `json:"id" validate:"required"`
	Score   *float64  `json:"score,omitempty" validate:"omitempty,min=0,max=10"`
	Watched *bool     `json:"watched,omitempty"`
}

// ListMoviesInput represents the input for listing movies
type ListMoviesInput struct {
	Watched bool   `query:"watched"`
	Query   string `query:"query"`
	Page    int    `query:"page" validate:"min=1"`
	Limit   int    `query:"limit" validate:"min=1,max=100"`
}

// PaginatedMovies represents a paginated list of movies
type PaginatedMovies struct {
	Results    []Movie `json:"results"`
	Page       int     `json:"page"`
	Count      int     `json:"count"`
	TotalPages int     `json:"totalPages"`
}
