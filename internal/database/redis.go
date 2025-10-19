package database

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"log"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

// RedisClient wraps the redis client
type RedisClient struct {
	*redis.Client
}

// RedisConfig holds Redis configuration
type RedisConfig struct {
	Addr     string
	Password string
	DB       int
	TLS      bool
}

// NewRedisClient creates a new Redis client
func NewRedisClient(cfg RedisConfig) (*RedisClient, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     cfg.Addr,
		Password: cfg.Password,
		DB:       cfg.DB,
	})

	// Test connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("unable to ping Redis: %w", err)
	}

	log.Println("Successfully connected to Redis")

	return &RedisClient{Client: client}, nil
}

// Close closes the Redis connection
func (r *RedisClient) Close() error {
	if r.Client != nil {
		log.Println("Closing Redis connection")
		return r.Client.Close()
	}
	return nil
}

// Health checks the Redis connection health
func (r *RedisClient) Health(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()

	return r.Ping(ctx).Err()
}

// SessionStore handles session storage in Redis
type SessionStore struct {
	client *RedisClient
	ttl    time.Duration
}

// NewSessionStore creates a new session store
func NewSessionStore(client *RedisClient, ttl time.Duration) *SessionStore {
	if ttl == 0 {
		ttl = 7 * 24 * time.Hour // default 7 days
	}
	return &SessionStore{
		client: client,
		ttl:    ttl,
	}
}

// GenerateSessionID generates a cryptographically secure session ID
func (s *SessionStore) GenerateSessionID() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("failed to generate session ID: %w", err)
	}
	return base64.URLEncoding.EncodeToString(b), nil
}

// Set stores a user ID in a session
func (s *SessionStore) Set(ctx context.Context, sessionID string, userID uuid.UUID) error {
	key := fmt.Sprintf("session:%s", sessionID)
	return s.client.Set(ctx, key, userID.String(), s.ttl).Err()
}

// Get retrieves a user ID from a session
func (s *SessionStore) Get(ctx context.Context, sessionID string) (uuid.UUID, error) {
	key := fmt.Sprintf("session:%s", sessionID)

	val, err := s.client.Get(ctx, key).Result()
	if err == redis.Nil {
		return uuid.Nil, fmt.Errorf("session not found")
	}
	if err != nil {
		return uuid.Nil, fmt.Errorf("failed to get session: %w", err)
	}

	userID, err := uuid.Parse(val)
	if err != nil {
		return uuid.Nil, fmt.Errorf("invalid user ID in session: %w", err)
	}

	// Refresh TTL on access
	s.client.Expire(ctx, key, s.ttl)

	return userID, nil
}

// Delete removes a session
func (s *SessionStore) Delete(ctx context.Context, sessionID string) error {
	key := fmt.Sprintf("session:%s", sessionID)
	return s.client.Del(ctx, key).Err()
}

// Exists checks if a session exists
func (s *SessionStore) Exists(ctx context.Context, sessionID string) (bool, error) {
	key := fmt.Sprintf("session:%s", sessionID)

	result, err := s.client.Exists(ctx, key).Result()
	if err != nil {
		return false, fmt.Errorf("failed to check session existence: %w", err)
	}

	return result > 0, nil
}
