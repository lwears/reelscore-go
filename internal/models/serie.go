package models

import (
	"time"

	"github.com/google/uuid"
)

// Serie represents a TV series in the user's library
type Serie struct {
	ID         uuid.UUID  `db:"id" json:"id"`
	TmdbID     int        `db:"tmdbId" json:"tmdbId"`
	CreatedAt  time.Time  `db:"createdAt" json:"createdAt"`
	UpdatedAt  time.Time  `db:"updatedAt" json:"updatedAt"`
	Title      string     `db:"title" json:"title"`
	PosterPath *string    `db:"posterPath" json:"posterPath"`
	FirstAired *time.Time `db:"firstAired" json:"firstAired"`
	TmdbScore  float64    `db:"tmdbScore" json:"tmdbScore"`
	Score      float64    `db:"score" json:"score"`
	Watched    bool       `db:"watched" json:"watched"`
	UserID     uuid.UUID  `db:"userId" json:"userId"`
}

// CreateSerieInput represents the input for creating a serie
type CreateSerieInput struct {
	TmdbID     int      `json:"tmdbId" validate:"required"`
	Title      string   `json:"title" validate:"required"`
	PosterPath *string  `json:"posterPath"`
	FirstAired *string  `json:"firstAired"`
	Watched    bool     `json:"watched"`
	TmdbScore  float64  `json:"tmdbScore" validate:"min=0,max=10"`
	Score      *float64 `json:"score,omitempty" validate:"omitempty,min=0,max=10"`
}

// UpdateSerieInput represents the input for updating a serie
type UpdateSerieInput struct {
	ID      uuid.UUID `json:"id" validate:"required"`
	Score   *float64  `json:"score,omitempty" validate:"omitempty,min=0,max=10"`
	Watched *bool     `json:"watched,omitempty"`
}

// ListSeriesInput represents the input for listing series
type ListSeriesInput struct {
	Watched bool   `query:"watched"`
	Query   string `query:"query"`
	Page    int    `query:"page" validate:"min=1"`
	Limit   int    `query:"limit" validate:"min=1,max=100"`
}

// PaginatedSeries represents a paginated list of series
type PaginatedSeries struct {
	Results    []Serie `json:"results"`
	Page       int     `json:"page"`
	Count      int     `json:"count"`
	TotalPages int     `json:"totalPages"`
}
