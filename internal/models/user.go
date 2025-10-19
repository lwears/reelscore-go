package models

import (
	"time"

	"github.com/google/uuid"
)

// Provider represents the OAuth provider type
type Provider string

const (
	ProviderGitHub Provider = "GITHUB"
	ProviderGoogle Provider = "GOOGLE"
)

// User represents a user in the system
type User struct {
	ID         uuid.UUID `db:"id" json:"id"`
	ProviderID string    `db:"providerId" json:"providerId"`
	Provider   Provider  `db:"provider" json:"provider"`
	Email      string    `db:"email" json:"email"`
	Name       string    `db:"name" json:"name"`
	CreatedAt  time.Time `db:"createdAt" json:"createdAt"`
	UpdatedAt  time.Time `db:"updatedAt" json:"updatedAt"`
}

// String returns the string representation of Provider
func (p Provider) String() string {
	return string(p)
}

// IsValid checks if the provider is valid
func (p Provider) IsValid() bool {
	return p == ProviderGitHub || p == ProviderGoogle
}
