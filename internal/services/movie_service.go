package services

import (
	"context"
	"fmt"
	"math"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/liamwears/reelscore/internal/models"
)

// MovieService handles movie-related business logic
type MovieService struct {
	db *pgxpool.Pool
}

// NewMovieService creates a new MovieService
func NewMovieService(db *pgxpool.Pool) *MovieService {
	return &MovieService{db: db}
}

// List retrieves movies for a user with pagination and filtering
func (s *MovieService) List(ctx context.Context, userID uuid.UUID, input models.ListMoviesInput) (*models.PaginatedMovies, error) {
	// Set defaults
	if input.Page < 1 {
		input.Page = 1
	}
	if input.Limit < 1 || input.Limit > 100 {
		input.Limit = 27
	}

	offset := (input.Page - 1) * input.Limit

	// Build query
	baseQuery := `
		FROM "Movie"
		WHERE "userId" = $1 AND watched = $2
	`
	args := []interface{}{userID, input.Watched}
	argCount := 2

	// Add search filter if provided
	if input.Query != "" {
		argCount++
		baseQuery += fmt.Sprintf(" AND title ILIKE $%d", argCount)
		args = append(args, "%"+input.Query+"%")
	}

	// Count total
	var total int
	countQuery := "SELECT COUNT(*) " + baseQuery
	err := s.db.QueryRow(ctx, countQuery, args...).Scan(&total)
	if err != nil {
		return nil, fmt.Errorf("failed to count movies: %w", err)
	}

	// Get movies
	query := `
		SELECT id, "tmdbId", "createdAt", "updatedAt", title, "posterPath",
		       "releaseDate", "tmdbScore", score, watched, "userId"
	` + baseQuery + `
		ORDER BY "tmdbScore" DESC
		LIMIT $` + fmt.Sprintf("%d", argCount+1) + ` OFFSET $` + fmt.Sprintf("%d", argCount+2)

	args = append(args, input.Limit, offset)

	rows, err := s.db.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query movies: %w", err)
	}
	defer rows.Close()

	var movies []models.Movie
	for rows.Next() {
		var movie models.Movie
		err := rows.Scan(
			&movie.ID,
			&movie.TmdbID,
			&movie.CreatedAt,
			&movie.UpdatedAt,
			&movie.Title,
			&movie.PosterPath,
			&movie.ReleaseDate,
			&movie.TmdbScore,
			&movie.Score,
			&movie.Watched,
			&movie.UserID,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan movie: %w", err)
		}
		movies = append(movies, movie)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating movies: %w", err)
	}

	totalPages := int(math.Ceil(float64(total) / float64(input.Limit)))

	return &models.PaginatedMovies{
		Results:    movies,
		Page:       input.Page,
		Count:      total,
		TotalPages: totalPages,
	}, nil
}

// Create creates a new movie
func (s *MovieService) Create(ctx context.Context, userID uuid.UUID, input models.CreateMovieInput) (*models.Movie, error) {
	score := 0.0
	if input.Score != nil {
		score = *input.Score
	}

	// Parse release date from string to time.Time
	var releaseDate *time.Time
	if input.ReleaseDate != nil && *input.ReleaseDate != "" {
		parsedDate, err := time.Parse("2006-01-02", *input.ReleaseDate)
		if err == nil {
			releaseDate = &parsedDate
		}
	}

	query := `
		INSERT INTO "Movie" ("tmdbId", title, "posterPath", "releaseDate", "tmdbScore", score, watched, "userId")
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id, "tmdbId", "createdAt", "updatedAt", title, "posterPath",
		          "releaseDate", "tmdbScore", score, watched, "userId"
	`

	var movie models.Movie
	err := s.db.QueryRow(ctx, query,
		input.TmdbID,
		input.Title,
		input.PosterPath,
		releaseDate,
		input.TmdbScore,
		score,
		input.Watched,
		userID,
	).Scan(
		&movie.ID,
		&movie.TmdbID,
		&movie.CreatedAt,
		&movie.UpdatedAt,
		&movie.Title,
		&movie.PosterPath,
		&movie.ReleaseDate,
		&movie.TmdbScore,
		&movie.Score,
		&movie.Watched,
		&movie.UserID,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to create movie: %w", err)
	}

	return &movie, nil
}

// Get retrieves a movie by ID
func (s *MovieService) Get(ctx context.Context, id uuid.UUID, userID uuid.UUID) (*models.Movie, error) {
	query := `
		SELECT id, "tmdbId", "createdAt", "updatedAt", title, "posterPath",
		       "releaseDate", "tmdbScore", score, watched, "userId"
		FROM "Movie"
		WHERE id = $1 AND "userId" = $2
	`

	var movie models.Movie
	err := s.db.QueryRow(ctx, query, id, userID).Scan(
		&movie.ID,
		&movie.TmdbID,
		&movie.CreatedAt,
		&movie.UpdatedAt,
		&movie.Title,
		&movie.PosterPath,
		&movie.ReleaseDate,
		&movie.TmdbScore,
		&movie.Score,
		&movie.Watched,
		&movie.UserID,
	)

	if err != nil {
		return nil, err
	}

	return &movie, nil
}

// Update updates a movie
func (s *MovieService) Update(ctx context.Context, userID uuid.UUID, input models.UpdateMovieInput) (*models.Movie, error) {
	// Build dynamic update query
	query := `UPDATE "Movie" SET "updatedAt" = NOW()`
	args := []interface{}{}
	argCount := 0

	if input.Score != nil {
		argCount++
		query += fmt.Sprintf(`, score = $%d`, argCount)
		args = append(args, *input.Score)
	}

	if input.Watched != nil {
		argCount++
		query += fmt.Sprintf(`, watched = $%d`, argCount)
		args = append(args, *input.Watched)
	}

	argCount++
	query += fmt.Sprintf(` WHERE id = $%d`, argCount)
	args = append(args, input.ID)

	argCount++
	query += fmt.Sprintf(` AND "userId" = $%d`, argCount)
	args = append(args, userID)

	query += `
		RETURNING id, "tmdbId", "createdAt", "updatedAt", title, "posterPath",
		          "releaseDate", "tmdbScore", score, watched, "userId"
	`

	var movie models.Movie
	err := s.db.QueryRow(ctx, query, args...).Scan(
		&movie.ID,
		&movie.TmdbID,
		&movie.CreatedAt,
		&movie.UpdatedAt,
		&movie.Title,
		&movie.PosterPath,
		&movie.ReleaseDate,
		&movie.TmdbScore,
		&movie.Score,
		&movie.Watched,
		&movie.UserID,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to update movie: %w", err)
	}

	return &movie, nil
}

// Delete deletes a movie
func (s *MovieService) Delete(ctx context.Context, id uuid.UUID, userID uuid.UUID) error {
	query := `DELETE FROM "Movie" WHERE id = $1 AND "userId" = $2`

	result, err := s.db.Exec(ctx, query, id, userID)
	if err != nil {
		return fmt.Errorf("failed to delete movie: %w", err)
	}

	if result.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}

	return nil
}
