package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/github"
	"golang.org/x/oauth2/google"

	"github.com/liamwears/reelscore/internal/database"
	"github.com/liamwears/reelscore/internal/middleware"
	"github.com/liamwears/reelscore/internal/models"
	"github.com/liamwears/reelscore/internal/services"
)

// AuthHandler handles authentication requests
type AuthHandler struct {
	userService    *services.UserService
	sessionStore   *database.SessionStore
	authMiddleware *middleware.AuthMiddleware
	googleConfig   *oauth2.Config
	githubConfig   *oauth2.Config
	renderer       *Renderer
	logger         *log.Logger
}

// AuthConfig holds authentication configuration
type AuthConfig struct {
	GoogleClientID     string
	GoogleClientSecret string
	GitHubClientID     string
	GitHubClientSecret string
	CallbackHost       string
}

// NewAuthHandler creates a new auth handler
func NewAuthHandler(
	userService *services.UserService,
	sessionStore *database.SessionStore,
	authMiddleware *middleware.AuthMiddleware,
	renderer *Renderer,
	cfg AuthConfig,
	logger *log.Logger,
) *AuthHandler {
	ghConfig := &oauth2.Config{
		ClientID:     cfg.GitHubClientID,
		ClientSecret: cfg.GitHubClientSecret,
		RedirectURL:  fmt.Sprintf("%s/auth/github/callback", cfg.CallbackHost),
		Scopes:       []string{"user:email"},
		Endpoint:     github.Endpoint,
	}

	googleConfig := &oauth2.Config{
		ClientID:     cfg.GoogleClientID,
		ClientSecret: cfg.GoogleClientSecret,
		RedirectURL:  fmt.Sprintf("%s/auth/google/callback", cfg.CallbackHost),
		Scopes:       []string{"profile", "email"},
		Endpoint:     google.Endpoint,
	}

	// Log the constructed callback URL for debugging
	logger.Printf("Google OAuth Callback URL: %s", googleConfig.RedirectURL)

	return &AuthHandler{
		userService:    userService,
		sessionStore:   sessionStore,
		authMiddleware: authMiddleware,
		renderer:       renderer,
		logger:         logger,
		googleConfig:   googleConfig,
		githubConfig:   ghConfig,
	}
}

// Login displays the login page
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	h.renderer.RenderPage(w, "login.html", nil)
}

// GoogleLogin initiates Google OAuth flow
func (h *AuthHandler) GoogleLogin(w http.ResponseWriter, r *http.Request) {
	// Generate state token for CSRF protection
	state, err := h.sessionStore.GenerateSessionID()
	if err != nil {
		h.logger.Printf("Failed to generate state token: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Store state in session temporarily (5 minutes)
	// In production, you might want to use a separate cache for state tokens
	ctx := context.WithValue(r.Context(), "oauth_state", state)

	// Redirect to Google
	url := h.googleConfig.AuthCodeURL(state, oauth2.AccessTypeOffline)
	http.Redirect(w, r.WithContext(ctx), url, http.StatusTemporaryRedirect)
}

// GoogleCallback handles Google OAuth callback
func (h *AuthHandler) GoogleCallback(w http.ResponseWriter, r *http.Request) {
	// Verify state (simplified - in production use proper state validation)
	code := r.URL.Query().Get("code")
	if code == "" {
		http.Error(w, "No code provided", http.StatusBadRequest)
		return
	}

	// Exchange code for token
	token, err := h.googleConfig.Exchange(r.Context(), code)
	if err != nil {
		h.logger.Printf("Failed to exchange code: %v", err)
		http.Error(w, "Failed to exchange code", http.StatusInternalServerError)
		return
	}

	// Get user info from Google
	client := h.googleConfig.Client(r.Context(), token)
	resp, err := client.Get("https://www.googleapis.com/oauth2/v2/userinfo")
	if err != nil {
		h.logger.Printf("Failed to get user info: %v", err)
		http.Error(w, "Failed to get user info", http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	var userInfo struct {
		ID    string `json:"id"`
		Email string `json:"email"`
		Name  string `json:"name"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&userInfo); err != nil {
		h.logger.Printf("Failed to decode user info: %v", err)
		http.Error(w, "Failed to decode user info", http.StatusInternalServerError)
		return
	}

	// Find or create user
	user, err := h.userService.FindOrCreate(
		r.Context(),
		userInfo.ID,
		models.ProviderGoogle,
		userInfo.Email,
		userInfo.Name,
	)
	if err != nil {
		h.logger.Printf("Failed to find or create user: %v", err)
		http.Error(w, "Failed to create user", http.StatusInternalServerError)
		return
	}

	// Create session
	sessionID, err := h.sessionStore.GenerateSessionID()
	if err != nil {
		h.logger.Printf("Failed to generate session ID: %v", err)
		http.Error(w, "Failed to create session", http.StatusInternalServerError)
		return
	}

	if err := h.sessionStore.Set(r.Context(), sessionID, user.ID); err != nil {
		h.logger.Printf("Failed to store session: %v", err)
		http.Error(w, "Failed to store session", http.StatusInternalServerError)
		return
	}

	// Set cookie
	h.authMiddleware.SetSessionCookie(w, sessionID)

	// Redirect to movies page
	http.Redirect(w, r, "/movies", http.StatusSeeOther)
}

// GitHubLogin initiates GitHub OAuth flow
func (h *AuthHandler) GitHubLogin(w http.ResponseWriter, r *http.Request) {
	// Generate state token for CSRF protection
	state, err := h.sessionStore.GenerateSessionID()
	if err != nil {
		h.logger.Printf("Failed to generate state token: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Redirect to GitHub
	url := h.githubConfig.AuthCodeURL(state)
	http.Redirect(w, r, url, http.StatusTemporaryRedirect)
}

// GitHubCallback handles GitHub OAuth callback
func (h *AuthHandler) GitHubCallback(w http.ResponseWriter, r *http.Request) {
	// Verify state (simplified - in production use proper state validation)
	code := r.URL.Query().Get("code")
	if code == "" {
		http.Error(w, "No code provided", http.StatusBadRequest)
		return
	}

	// Exchange code for token
	token, err := h.githubConfig.Exchange(r.Context(), code)
	if err != nil {
		h.logger.Printf("Failed to exchange code: %v", err)
		http.Error(w, "Failed to exchange code", http.StatusInternalServerError)
		return
	}

	// Get user info from GitHub
	client := h.githubConfig.Client(r.Context(), token)
	resp, err := client.Get("https://api.github.com/user")
	if err != nil {
		h.logger.Printf("Failed to get user info: %v", err)
		http.Error(w, "Failed to get user info", http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	var userInfo struct {
		ID    int    `json:"id"`
		Email string `json:"email"`
		Name  string `json:"name"`
		Login string `json:"login"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&userInfo); err != nil {
		h.logger.Printf("Failed to decode user info: %v", err)
		http.Error(w, "Failed to decode user info", http.StatusInternalServerError)
		return
	}

	// GitHub might not return email in main user object, need to fetch separately if null
	if userInfo.Email == "" {
		emailResp, err := client.Get("https://api.github.com/user/emails")
		if err == nil {
			defer emailResp.Body.Close()
			var emails []struct {
				Email   string `json:"email"`
				Primary bool   `json:"primary"`
			}
			if err := json.NewDecoder(emailResp.Body).Decode(&emails); err == nil {
				for _, email := range emails {
					if email.Primary {
						userInfo.Email = email.Email
						break
					}
				}
			}
		}
	}

	// Use login if name is empty
	if userInfo.Name == "" {
		userInfo.Name = userInfo.Login
	}

	// Find or create user
	user, err := h.userService.FindOrCreate(
		r.Context(),
		fmt.Sprintf("%d", userInfo.ID),
		models.ProviderGitHub,
		userInfo.Email,
		userInfo.Name,
	)
	if err != nil {
		h.logger.Printf("Failed to find or create user: %v", err)
		http.Error(w, "Failed to create user", http.StatusInternalServerError)
		return
	}

	// Create session
	sessionID, err := h.sessionStore.GenerateSessionID()
	if err != nil {
		h.logger.Printf("Failed to generate session ID: %v", err)
		http.Error(w, "Failed to create session", http.StatusInternalServerError)
		return
	}

	if err := h.sessionStore.Set(r.Context(), sessionID, user.ID); err != nil {
		h.logger.Printf("Failed to store session: %v", err)
		http.Error(w, "Failed to store session", http.StatusInternalServerError)
		return
	}

	// Set cookie
	h.authMiddleware.SetSessionCookie(w, sessionID)

	// Redirect to movies page
	http.Redirect(w, r, "/movies", http.StatusSeeOther)
}

// Logout handles user logout
func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	// Get session cookie
	cookie, err := r.Cookie("session")
	if err == nil {
		// Delete session from Redis
		h.sessionStore.Delete(r.Context(), cookie.Value)
	}

	// Clear cookie
	h.authMiddleware.ClearSessionCookie(w)

	// Redirect to login
	http.Redirect(w, r, "/login", http.StatusSeeOther)
}
