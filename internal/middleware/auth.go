package middleware

import (
	"context"
	"net/http"

	"github.com/google/uuid"
	"github.com/liamwears/reelscore/internal/database"
	"github.com/liamwears/reelscore/internal/models"
	"github.com/liamwears/reelscore/internal/services"
)

// ContextKey is a custom type for context keys to avoid collisions
type ContextKey string

const (
	// UserContextKey is the key for storing user in context
	UserContextKey ContextKey = "user"
	// UserIDContextKey is the key for storing user ID in context
	UserIDContextKey ContextKey = "userID"
)

// AuthMiddleware handles authentication for protected routes
type AuthMiddleware struct {
	sessionStore *database.SessionStore
	userService  *services.UserService
	cookieName   string
	isProduction bool
}

// NewAuthMiddleware creates a new authentication middleware
func NewAuthMiddleware(sessionStore *database.SessionStore, userService *services.UserService, cookieName string, isProduction bool) *AuthMiddleware {
	if cookieName == "" {
		cookieName = "session"
	}
	return &AuthMiddleware{
		sessionStore: sessionStore,
		userService:  userService,
		cookieName:   cookieName,
		isProduction: isProduction,
	}
}

// RequireAuth ensures the user is authenticated
func (m *AuthMiddleware) RequireAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Get session cookie
		cookie, err := r.Cookie(m.cookieName)
		if err != nil {
			// No session cookie, redirect to login
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}

		// Get user ID from session
		userID, err := m.sessionStore.Get(r.Context(), cookie.Value)
		if err != nil {
			// Invalid or expired session, redirect to login
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}

		// Get user from database
		user, err := m.userService.Get(r.Context(), userID)
		if err != nil {
			// User not found, clear session and redirect
			m.sessionStore.Delete(r.Context(), cookie.Value)
			http.SetCookie(w, &http.Cookie{
				Name:   m.cookieName,
				Value:  "",
				Path:   "/",
				MaxAge: -1,
			})
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}

		// Add user to context
		ctx := context.WithValue(r.Context(), UserContextKey, user)
		ctx = context.WithValue(ctx, UserIDContextKey, userID)

		// Call next handler
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// OptionalAuth checks for authentication but doesn't require it
func (m *AuthMiddleware) OptionalAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Get session cookie
		cookie, err := r.Cookie(m.cookieName)
		if err != nil {
			// No session, continue without user
			next.ServeHTTP(w, r)
			return
		}

		// Get user ID from session
		userID, err := m.sessionStore.Get(r.Context(), cookie.Value)
		if err != nil {
			// Invalid session, continue without user
			next.ServeHTTP(w, r)
			return
		}

		// Get user from database
		user, err := m.userService.Get(r.Context(), userID)
		if err != nil {
			// User not found, continue without user
			next.ServeHTTP(w, r)
			return
		}

		// Add user to context
		ctx := context.WithValue(r.Context(), UserContextKey, user)
		ctx = context.WithValue(ctx, UserIDContextKey, userID)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// RequireAuthAPI ensures the user is authenticated for API requests
func (m *AuthMiddleware) RequireAuthAPI(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Get session cookie
		cookie, err := r.Cookie(m.cookieName)
		if err != nil {
			http.Error(w, `{"error":"Unauthorized"}`, http.StatusUnauthorized)
			return
		}

		// Get user ID from session
		userID, err := m.sessionStore.Get(r.Context(), cookie.Value)
		if err != nil {
			http.Error(w, `{"error":"Unauthorized"}`, http.StatusUnauthorized)
			return
		}

		// Get user from database
		user, err := m.userService.Get(r.Context(), userID)
		if err != nil {
			http.Error(w, `{"error":"Unauthorized"}`, http.StatusUnauthorized)
			return
		}

		// Add user to context
		ctx := context.WithValue(r.Context(), UserContextKey, user)
		ctx = context.WithValue(ctx, UserIDContextKey, userID)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// GetUserFromContext retrieves the user from request context
func GetUserFromContext(ctx context.Context) (*models.User, bool) {
	user, ok := ctx.Value(UserContextKey).(*models.User)
	return user, ok
}

// GetUserIDFromContext retrieves the user ID from request context
func GetUserIDFromContext(ctx context.Context) (uuid.UUID, bool) {
	userID, ok := ctx.Value(UserIDContextKey).(uuid.UUID)
	return userID, ok
}

// SetSessionCookie sets a session cookie
func (m *AuthMiddleware) SetSessionCookie(w http.ResponseWriter, sessionID string) {
	cookie := &http.Cookie{
		Name:     m.cookieName,
		Value:    sessionID,
		Path:     "/",
		MaxAge:   7 * 24 * 60 * 60, // 7 days
		HttpOnly: true,
		Secure:   m.isProduction,
		SameSite: http.SameSiteLaxMode,
	}
	http.SetCookie(w, cookie)
}

// ClearSessionCookie clears the session cookie
func (m *AuthMiddleware) ClearSessionCookie(w http.ResponseWriter) {
	cookie := &http.Cookie{
		Name:     m.cookieName,
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   m.isProduction,
		SameSite: http.SameSiteLaxMode,
	}
	http.SetCookie(w, cookie)
}
