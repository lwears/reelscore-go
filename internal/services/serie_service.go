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

// SerieService handles serie-related business logic
type SerieService struct {
	db *pgxpool.Pool
}

// NewSerieService creates a new SerieService
func NewSerieService(db *pgxpool.Pool) *SerieService {
	return &SerieService{db: db}
}

// List retrieves series for a user with pagination and filtering
func (s *SerieService) List(ctx context.Context, userID uuid.UUID, input models.ListSeriesInput) (*models.PaginatedSeries, error) {
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
		FROM "Serie"
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
		return nil, fmt.Errorf("failed to count series: %w", err)
	}

	// Get series
	query := `
		SELECT id, "tmdbId", "createdAt", "updatedAt", title, "posterPath",
		       "firstAired", "tmdbScore", score, watched, "userId"
	` + baseQuery + `
		ORDER BY "tmdbScore" DESC
		LIMIT $` + fmt.Sprintf("%d", argCount+1) + ` OFFSET $` + fmt.Sprintf("%d", argCount+2)

	args = append(args, input.Limit, offset)

	rows, err := s.db.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query series: %w", err)
	}
	defer rows.Close()

	var series []models.Serie
	for rows.Next() {
		var serie models.Serie
		err := rows.Scan(
			&serie.ID,
			&serie.TmdbID,
			&serie.CreatedAt,
			&serie.UpdatedAt,
			&serie.Title,
			&serie.PosterPath,
			&serie.FirstAired,
			&serie.TmdbScore,
			&serie.Score,
			&serie.Watched,
			&serie.UserID,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan serie: %w", err)
		}
		series = append(series, serie)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating series: %w", err)
	}

	totalPages := int(math.Ceil(float64(total) / float64(input.Limit)))

	return &models.PaginatedSeries{
		Results:    series,
		Page:       input.Page,
		Count:      total,
		TotalPages: totalPages,
	}, nil
}

// Create creates a new serie
func (s *SerieService) Create(ctx context.Context, userID uuid.UUID, input models.CreateSerieInput) (*models.Serie, error) {
	score := 0.0
	if input.Score != nil {
		score = *input.Score
	}

	// Parse firstAired date from string to time.Time
	var firstAired *time.Time
	if input.FirstAired != nil && *input.FirstAired != "" {
		parsedDate, err := time.Parse("2006-01-02", *input.FirstAired)
		if err == nil {
			firstAired = &parsedDate
		}
	}

	query := `
		INSERT INTO "Serie" ("tmdbId", title, "posterPath", "firstAired", "tmdbScore", score, watched, "userId")
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id, "tmdbId", "createdAt", "updatedAt", title, "posterPath",
		          "firstAired", "tmdbScore", score, watched, "userId"
	`

	var serie models.Serie
	err := s.db.QueryRow(ctx, query,
		input.TmdbID,
		input.Title,
		input.PosterPath,
		firstAired,
		input.TmdbScore,
		score,
		input.Watched,
		userID,
	).Scan(
		&serie.ID,
		&serie.TmdbID,
		&serie.CreatedAt,
		&serie.UpdatedAt,
		&serie.Title,
		&serie.PosterPath,
		&serie.FirstAired,
		&serie.TmdbScore,
		&serie.Score,
		&serie.Watched,
		&serie.UserID,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to create serie: %w", err)
	}

	return &serie, nil
}

// Get retrieves a serie by ID
func (s *SerieService) Get(ctx context.Context, id uuid.UUID, userID uuid.UUID) (*models.Serie, error) {
	query := `
		SELECT id, "tmdbId", "createdAt", "updatedAt", title, "posterPath",
		       "firstAired", "tmdbScore", score, watched, "userId"
		FROM "Serie"
		WHERE id = $1 AND "userId" = $2
	`

	var serie models.Serie
	err := s.db.QueryRow(ctx, query, id, userID).Scan(
		&serie.ID,
		&serie.TmdbID,
		&serie.CreatedAt,
		&serie.UpdatedAt,
		&serie.Title,
		&serie.PosterPath,
		&serie.FirstAired,
		&serie.TmdbScore,
		&serie.Score,
		&serie.Watched,
		&serie.UserID,
	)

	if err != nil {
		return nil, err
	}

	return &serie, nil
}

// Update updates a serie
func (s *SerieService) Update(ctx context.Context, userID uuid.UUID, input models.UpdateSerieInput) (*models.Serie, error) {
	// Build dynamic update query
	query := `UPDATE "Serie" SET "updatedAt" = NOW()`
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
		          "firstAired", "tmdbScore", score, watched, "userId"
	`

	var serie models.Serie
	err := s.db.QueryRow(ctx, query, args...).Scan(
		&serie.ID,
		&serie.TmdbID,
		&serie.CreatedAt,
		&serie.UpdatedAt,
		&serie.Title,
		&serie.PosterPath,
		&serie.FirstAired,
		&serie.TmdbScore,
		&serie.Score,
		&serie.Watched,
		&serie.UserID,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to update serie: %w", err)
	}

	return &serie, nil
}

// Delete deletes a serie
func (s *SerieService) Delete(ctx context.Context, id uuid.UUID, userID uuid.UUID) error {
	query := `DELETE FROM "Serie" WHERE id = $1 AND "userId" = $2`

	result, err := s.db.Exec(ctx, query, id, userID)
	if err != nil {
		return fmt.Errorf("failed to delete serie: %w", err)
	}

	if result.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}

	return nil
}
