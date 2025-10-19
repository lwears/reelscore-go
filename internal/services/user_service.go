package services

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/liamwears/reelscore/internal/models"
)

// UserService handles user-related business logic
type UserService struct {
	db *pgxpool.Pool
}

// NewUserService creates a new UserService
func NewUserService(db *pgxpool.Pool) *UserService {
	return &UserService{db: db}
}

// FindOrCreate finds a user by provider ID or creates a new one
func (s *UserService) FindOrCreate(ctx context.Context, providerID string, provider models.Provider, email, name string) (*models.User, error) {
	// Try to find existing user
	user, err := s.FindByProviderID(ctx, providerID)
	if err == nil {
		return user, nil
	}

	// If not found, create new user
	if err == pgx.ErrNoRows {
		return s.Create(ctx, providerID, provider, email, name)
	}

	return nil, fmt.Errorf("failed to find user: %w", err)
}

// FindByProviderID finds a user by their provider ID
func (s *UserService) FindByProviderID(ctx context.Context, providerID string) (*models.User, error) {
	query := `
		SELECT id, "providerId", provider, email, name, "createdAt", "updatedAt"
		FROM "User"
		WHERE "providerId" = $1
	`

	var user models.User
	err := s.db.QueryRow(ctx, query, providerID).Scan(
		&user.ID,
		&user.ProviderID,
		&user.Provider,
		&user.Email,
		&user.Name,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	return &user, nil
}

// Create creates a new user
func (s *UserService) Create(ctx context.Context, providerID string, provider models.Provider, email, name string) (*models.User, error) {
	if !provider.IsValid() {
		return nil, fmt.Errorf("invalid provider: %s", provider)
	}

	query := `
		INSERT INTO "User" ("providerId", provider, email, name)
		VALUES ($1, $2, $3, $4)
		RETURNING id, "providerId", provider, email, name, "createdAt", "updatedAt"
	`

	var user models.User
	err := s.db.QueryRow(ctx, query, providerID, provider, email, name).Scan(
		&user.ID,
		&user.ProviderID,
		&user.Provider,
		&user.Email,
		&user.Name,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	return &user, nil
}

// Get retrieves a user by ID
func (s *UserService) Get(ctx context.Context, id uuid.UUID) (*models.User, error) {
	query := `
		SELECT id, "providerId", provider, email, name, "createdAt", "updatedAt"
		FROM "User"
		WHERE id = $1
	`

	var user models.User
	err := s.db.QueryRow(ctx, query, id).Scan(
		&user.ID,
		&user.ProviderID,
		&user.Provider,
		&user.Email,
		&user.Name,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	return &user, nil
}

// GetAll retrieves all users (mainly for admin purposes)
func (s *UserService) GetAll(ctx context.Context) ([]*models.User, error) {
	query := `
		SELECT id, "providerId", provider, email, name, "createdAt", "updatedAt"
		FROM "User"
		ORDER BY "createdAt" DESC
	`

	rows, err := s.db.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query users: %w", err)
	}
	defer rows.Close()

	var users []*models.User
	for rows.Next() {
		var user models.User
		err := rows.Scan(
			&user.ID,
			&user.ProviderID,
			&user.Provider,
			&user.Email,
			&user.Name,
			&user.CreatedAt,
			&user.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan user: %w", err)
		}
		users = append(users, &user)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating users: %w", err)
	}

	return users, nil
}

// Update updates a user's information
func (s *UserService) Update(ctx context.Context, id uuid.UUID, email, name string) (*models.User, error) {
	query := `
		UPDATE "User"
		SET email = $2, name = $3, "updatedAt" = NOW()
		WHERE id = $1
		RETURNING id, "providerId", provider, email, name, "createdAt", "updatedAt"
	`

	var user models.User
	err := s.db.QueryRow(ctx, query, id, email, name).Scan(
		&user.ID,
		&user.ProviderID,
		&user.Provider,
		&user.Email,
		&user.Name,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to update user: %w", err)
	}

	return &user, nil
}

// Delete deletes a user by ID
func (s *UserService) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM "User" WHERE id = $1`

	result, err := s.db.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete user: %w", err)
	}

	if result.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}

	return nil
}
