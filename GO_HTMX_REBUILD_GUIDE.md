# ReelScore: Go + HTMX Rebuild Guide

This document provides comprehensive specifications for rebuilding ReelScore using Go and HTMX, based on the current TypeScript/tRPC implementation.

## Table of Contents
1. [Project Overview](#project-overview)
2. [Database Schema](#database-schema)
3. [Authentication System](#authentication-system)
4. [API Endpoints](#api-endpoints)
5. [TMDB Integration](#tmdb-integration)
6. [Frontend Routes & Pages](#frontend-routes--pages)
7. [Key Features & User Flows](#key-features--user-flows)
8. [UI Components](#ui-components)
9. [Configuration](#configuration)
10. [Migration Notes](#migration-notes)

---

## Project Overview

**Current Stack:**
- Backend: Fastify (Node.js) + tRPC + Drizzle ORM
- Frontend: Next.js 14+ App Router + React
- Auth: Passport.js (OAuth2)
- Database: PostgreSQL
- Sessions: Redis
- Validation: Zod schemas

**Target Stack:**
- Backend: Go (Fiber/Echo/Chi framework)
- Frontend: Go templates + HTMX
- Auth: OAuth2 libraries (goth/oauth2)
- Database: PostgreSQL (pgx driver)
- Sessions: Redis (go-redis)
- Validation: Go struct tags + validator library

---

## Database Schema

### Tables

#### 1. User Table
```sql
CREATE TYPE "Provider" AS ENUM('GITHUB', 'GOOGLE');

CREATE TABLE "User" (
  "id" uuid PRIMARY KEY DEFAULT gen_random_uuid() NOT NULL,
  "providerId" varchar(255) NOT NULL UNIQUE,
  "provider" "Provider" NOT NULL,
  "email" varchar(255) NOT NULL,
  "name" varchar(255) NOT NULL
);
```

**Go Struct:**
```go
type Provider string

const (
    ProviderGitHub Provider = "GITHUB"
    ProviderGoogle Provider = "GOOGLE"
)

type User struct {
    ID         uuid.UUID `db:"id" json:"id"`
    ProviderID string    `db:"providerId" json:"providerId" validate:"required"`
    Provider   Provider  `db:"provider" json:"provider" validate:"required,oneof=GITHUB GOOGLE"`
    Email      string    `db:"email" json:"email" validate:"required,email"`
    Name       string    `db:"name" json:"name" validate:"required"`
}
```

#### 2. Movie Table
```sql
CREATE TABLE "Movie" (
  "id" uuid PRIMARY KEY DEFAULT gen_random_uuid() NOT NULL,
  "tmdbId" integer NOT NULL,
  "createdAt" timestamp DEFAULT now() NOT NULL,
  "updatedAt" timestamp DEFAULT now() NOT NULL,
  "title" varchar(255) NOT NULL,
  "posterPath" varchar(500),
  "releaseDate" timestamp,
  "tmdbScore" numeric(3, 1) DEFAULT '0' NOT NULL,
  "score" numeric(3, 1) DEFAULT '0' NOT NULL,
  "watched" boolean DEFAULT false NOT NULL,
  "userId" uuid NOT NULL,
  CONSTRAINT "Movie_tmdbId_userId_unique" UNIQUE("tmdbId","userId"),
  CONSTRAINT "Movie_userId_User_id_fk" FOREIGN KEY ("userId")
    REFERENCES "User"("id") ON DELETE cascade
);
```

**Go Struct:**
```go
type Movie struct {
    ID          uuid.UUID  `db:"id" json:"id"`
    TmdbID      int        `db:"tmdbId" json:"tmdbId" validate:"required"`
    CreatedAt   time.Time  `db:"createdAt" json:"createdAt"`
    UpdatedAt   time.Time  `db:"updatedAt" json:"updatedAt"`
    Title       string     `db:"title" json:"title" validate:"required"`
    PosterPath  *string    `db:"posterPath" json:"posterPath"`
    ReleaseDate *time.Time `db:"releaseDate" json:"releaseDate"`
    TmdbScore   float64    `db:"tmdbScore" json:"tmdbScore" validate:"min=0,max=10"`
    Score       float64    `db:"score" json:"score" validate:"min=0,max=10"`
    Watched     bool       `db:"watched" json:"watched"`
    UserID      uuid.UUID  `db:"userId" json:"userId" validate:"required"`
}
```

#### 3. Series Table
```sql
CREATE TABLE "Serie" (
  "id" uuid PRIMARY KEY DEFAULT gen_random_uuid() NOT NULL,
  "tmdbId" integer NOT NULL,
  "createdAt" timestamp DEFAULT now() NOT NULL,
  "updatedAt" timestamp DEFAULT now() NOT NULL,
  "title" varchar(255) NOT NULL,
  "posterPath" varchar(500),
  "firstAired" timestamp,
  "tmdbScore" numeric(3, 1) DEFAULT '0' NOT NULL,
  "score" numeric(3, 1) DEFAULT '0' NOT NULL,
  "watched" boolean DEFAULT false NOT NULL,
  "userId" uuid NOT NULL,
  CONSTRAINT "Serie_tmdbId_userId_unique" UNIQUE("tmdbId","userId"),
  CONSTRAINT "Serie_userId_User_id_fk" FOREIGN KEY ("userId")
    REFERENCES "User"("id") ON DELETE cascade
);
```

**Go Struct:**
```go
type Serie struct {
    ID         uuid.UUID  `db:"id" json:"id"`
    TmdbID     int        `db:"tmdbId" json:"tmdbId" validate:"required"`
    CreatedAt  time.Time  `db:"createdAt" json:"createdAt"`
    UpdatedAt  time.Time  `db:"updatedAt" json:"updatedAt"`
    Title      string     `db:"title" json:"title" validate:"required"`
    PosterPath *string    `db:"posterPath" json:"posterPath"`
    FirstAired *time.Time `db:"firstAired" json:"firstAired"`
    TmdbScore  float64    `db:"tmdbScore" json:"tmdbScore" validate:"min=0,max=10"`
    Score      float64    `db:"score" json:"score" validate:"min=0,max=10"`
    Watched    bool       `db:"watched" json:"watched"`
    UserID     uuid.UUID  `db:"userId" json:"userId" validate:"required"`
}
```

### Database Relationships
- Users → Movies (one-to-many, cascade delete)
- Users → Series (one-to-many, cascade delete)
- Unique constraint: (tmdbId, userId) for both Movies and Series

---

## Authentication System

### OAuth2 Flow

**Providers:** GitHub, Google

**Endpoints:**
1. `GET /auth/google/login` - Initiates Google OAuth
2. `GET /auth/google/callback` - Google callback handler
3. `GET /auth/github/login` - Initiates GitHub OAuth
4. `GET /auth/github/callback` - GitHub callback handler
5. `GET /auth/logout` - Destroys session, redirects to login

**OAuth Configuration:**

```go
// Environment variables needed
type OAuthConfig struct {
    GoogleClientID     string
    GoogleClientSecret string
    GitHubClientID     string
    GitHubClientSecret string
    CallbackHost       string // e.g., "http://localhost:4000"
}

// Scopes
// Google: ["profile", "email"]
// GitHub: ["user:email"]
```

**Callback URLs:**
- Google: `${HOST}/auth/google/callback`
- GitHub: `${HOST}/auth/github/callback`

**Flow:**
1. User clicks "Sign in with Google/GitHub"
2. Redirect to OAuth provider
3. Provider redirects back to callback URL
4. Extract profile data (id, email, name, provider)
5. Find or create user in database
6. Store user ID in Redis session
7. Set session cookie (httpOnly, secure in prod, sameSite=lax)
8. Redirect to `/movies`

**Session Management:**
- Store: Redis
- Cookie name: `session`
- TTL: 7 days (604800 seconds)
- Cookie config:
  - httpOnly: true
  - secure: true (production only)
  - sameSite: "lax"
  - maxAge: 7 days
  - path: "/"

**User Service Methods:**
```go
type UserService interface {
    FindOrCreate(providerID string, provider Provider, email, name string) (*User, error)
    Get(id uuid.UUID) (*User, error)
    GetAll() ([]*User, error)
    Update(id uuid.UUID, data map[string]interface{}) (*User, error)
    Delete(id uuid.UUID) error
}
```

**Middleware:**
```go
// Authentication middleware for protected routes
func RequireAuth(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        // Check session for user ID
        // If no user, redirect to /login
        // If user exists, add to context and call next
    })
}
```

---

## API Endpoints

All endpoints except auth use JSON for request/response bodies.

### User Router

#### GET Current User (Private)
**Endpoint:** Equivalent to `user.getCurrentUser`
```
GET /api/user/current
Authorization: Session cookie required
Response: User | null
```

**Go Handler:**
```go
func (h *UserHandler) GetCurrentUser(c *fiber.Ctx) error {
    user := c.Locals("user").(*User)
    return c.JSON(user)
}
```

---

### Movie Router

All movie endpoints require authentication.

#### 1. List Movies (Private)
**Input:**
```go
type ListMoviesRequest struct {
    Watched bool   `query:"watched" validate:"required"`
    Query   string `query:"query"`
    Page    int    `query:"page" validate:"min=1" default:"1"`
    Limit   int    `query:"limit" validate:"min=1,max=100" default:"27"`
}
```

**Output:**
```go
type PaginatedMovies struct {
    Results    []Movie `json:"results"`
    Page       int     `json:"page"`
    Count      int     `json:"count"`
    TotalPages int     `json:"totalPages"`
}
```

**SQL Query Logic:**
- Filter by userId (from session)
- Filter by watched (required)
- Optional: Filter by title (case-insensitive LIKE %query%)
- Order by: tmdbScore DESC
- Pagination: LIMIT/OFFSET

#### 2. Create Movie (Private)
**Input:**
```go
type CreateMovieRequest struct {
    TmdbID      int        `json:"tmdbId" validate:"required"`
    Title       string     `json:"title" validate:"required"`
    PosterPath  *string    `json:"posterPath"`
    ReleaseDate *time.Time `json:"releaseDate"`
    Watched     bool       `json:"watched"`
    TmdbScore   float64    `json:"tmdbScore" validate:"min=0,max=10"`
    Score       *float64   `json:"score,omitempty" validate:"omitempty,min=0,max=10"`
}
```

**Output:** Created Movie object

**Business Logic:**
- Set userId from session
- Set createdAt, updatedAt to now
- Default score to 0 if not provided
- Handle duplicate error (unique constraint on tmdbId + userId)

#### 3. Update Movie (Private)
**Input:**
```go
type UpdateMovieRequest struct {
    ID      uuid.UUID `json:"id" validate:"required"`
    Score   *float64  `json:"score,omitempty" validate:"omitempty,min=0,max=10"`
    Watched *bool     `json:"watched,omitempty"`
}
```

**Output:** Updated Movie object

**Business Logic:**
- Verify movie belongs to current user
- Update only provided fields
- Set updatedAt to now

#### 4. Delete Movie (Private)
**Input:**
```go
type DeleteMovieRequest struct {
    ID uuid.UUID `json:"id" validate:"required"`
}
```

**Output:** Deleted Movie object

**Business Logic:**
- Verify movie belongs to current user
- Delete from database

---

### Series Router

Same structure as Movie Router with Serie entity.

#### 1. List Series (Private)
**Endpoint:** `/api/series`
**Input:** Same as ListMoviesRequest
**Output:** PaginatedSeries

#### 2. Create Serie (Private)
**Endpoint:** `POST /api/series`
**Input:**
```go
type CreateSerieRequest struct {
    TmdbID     int        `json:"tmdbId" validate:"required"`
    Title      string     `json:"title" validate:"required"`
    PosterPath *string    `json:"posterPath"`
    FirstAired *time.Time `json:"firstAired"`
    Watched    bool       `json:"watched"`
    TmdbScore  float64    `json:"tmdbScore" validate:"min=0,max=10"`
    Score      *float64   `json:"score,omitempty" validate:"omitempty,min=0,max=10"`
}
```

#### 3. Update Serie (Private)
**Input:** Same structure as UpdateMovieRequest
**Output:** Updated Serie

#### 4. Delete Serie (Private)
**Input:** ID
**Output:** Deleted Serie

---

### TMDB Router

All endpoints require authentication.

#### 1. Get Movie by ID
**Endpoint:** `GET /api/tmdb/movie/:id`
**Input:** Movie ID (integer)
**Output:** TMDB Movie details

**TMDB API Call:**
```
GET https://api.themoviedb.org/3/movie/{movie_id}
Authorization: Bearer {TMDB_KEY}
```

#### 2. Get Serie by ID
**Endpoint:** `GET /api/tmdb/tv/:id`
**Input:** Serie ID (integer)
**Output:** TMDB TV details

**TMDB API Call:**
```
GET https://api.themoviedb.org/3/tv/{series_id}
Authorization: Bearer {TMDB_KEY}
```

#### 3. Search Multi (Movies + TV)
**Endpoint:** `GET /api/tmdb/search/multi`
**Input:**
```go
type SearchRequest struct {
    Query string `query:"query" validate:"required"`
    Page  int    `query:"page" validate:"min=1" default:"1"`
}
```

**TMDB API Call:**
```
GET https://api.themoviedb.org/3/search/multi
Params: query, page, language=en-US, include_adult=false
```

#### 4. Search Movies
**Endpoint:** `GET /api/tmdb/search/movie`
**Input:** Same as SearchRequest
**TMDB API:** `/3/search/movie`

#### 5. Search TV Series
**Endpoint:** `GET /api/tmdb/search/tv`
**Input:** Same as SearchRequest
**TMDB API:** `/3/search/tv`

#### 6. Popular Movies
**Endpoint:** `GET /api/tmdb/discover/movie`
**Input:**
```go
type DiscoverRequest struct {
    Page int `query:"page" validate:"min=1" default:"1"`
}
```
**TMDB API:** `/3/discover/movie`

#### 7. Popular TV Series
**Endpoint:** `GET /api/tmdb/discover/tv`
**Input:** Same as DiscoverRequest
**TMDB API:** `/3/discover/tv`

---

## TMDB Integration

### Configuration
```go
type TMDBConfig struct {
    APIKey  string // From env: TMDB_KEY
    BaseURL string // https://api.themoviedb.org
    ImageBaseURL string // https://image.tmdb.org/t/p/w500
}
```

### HTTP Client Setup
```go
// Add Authorization header to all requests
// Authorization: Bearer {TMDB_KEY}

// Default params for all requests
language := "en-US"
include_adult := false
```

### Key TMDB Endpoints Used

1. **GET /3/movie/{movie_id}** - Movie details
2. **GET /3/tv/{series_id}** - TV series details
3. **GET /3/search/movie** - Search movies
4. **GET /3/search/tv** - Search TV shows
5. **GET /3/search/multi** - Search both movies and TV
6. **GET /3/discover/movie** - Browse/discover movies (popular)
7. **GET /3/discover/tv** - Browse/discover TV series (popular)

### Response Mapping

**TMDB Movie Response:**
```json
{
  "id": 123,
  "title": "Movie Title",
  "poster_path": "/path.jpg",
  "backdrop_path": "/backdrop.jpg",
  "release_date": "2024-01-01",
  "vote_average": 7.5,
  "overview": "Description..."
}
```

**Map to Internal Structure:**
```go
type MovieData struct {
    TmdbID      int       `json:"tmdbId"`
    Title       string    `json:"title"`
    PosterPath  *string   `json:"posterPath"`
    ReleaseDate time.Time `json:"releaseDate"`
    TmdbScore   float64   `json:"tmdbScore"` // Rounded to 1 decimal
}
```

---

## Frontend Routes & Pages

### Route Structure

All routes except `/login` require authentication (middleware check).

#### Public Routes

1. **GET /login** - Login page
   - Displays GitHub and Google OAuth buttons
   - Links to: `/auth/github/login`, `/auth/google/login`

#### Protected Routes

2. **GET /movies** - Browse popular movies (default)
   - Query params: `?query=search&page=1`
   - Fetches from TMDB popular/search endpoint
   - Displays grid of movie cards
   - Each card shows: poster, title, year, TMDB score
   - Buttons: "Seen" (watched=true), "Watchlist" (watched=false)
   - Clicking card opens modal at `/movies/:id`

3. **GET /movies/:id** - Movie detail modal (parallel route)
   - Fetches movie details from TMDB
   - Displays: backdrop image, title, year, overview, TMDB score
   - 5-star rating component (1-5 scale, stored as 2-10)
   - Buttons: "Seen" (with rating), "Watchlist" (no rating)
   - On submit: Creates movie in DB, redirects back to `/movies`

4. **GET /series** - Browse popular TV series
   - Same structure as `/movies` but for TV shows
   - Uses `firstAired` instead of `releaseDate`

5. **GET /series/:id** - Serie detail modal
   - Same as movie modal but for TV series

6. **GET /search** - Search page
   - Query param: `?query=search`
   - Two sections: Movies and Series
   - Each section shows top results from TMDB search

7. **GET /library/movies/watched** - User's watched movies
   - Fetches from internal API: `/api/movies?watched=true`
   - Query params: `?query=filter&page=1`
   - Displays user's movie library
   - Shows user score if rated
   - Buttons: Delete, Update (if not watched yet)

8. **GET /library/movies/watchlist** - User's movie watchlist
   - Same as watched but `watched=false`

9. **GET /library/series/watched** - User's watched series
10. **GET /library/series/watchlist** - User's series watchlist

11. **GET /library/:media/:watched/:id** - Library item modal
    - Edit score for existing library item
    - Update watched status
    - Delete item

### HTMX Implementation

**Key Patterns:**

1. **Navigation** - Full page loads
2. **Pagination** - HTMX to swap card grid
   ```html
   <button hx-get="/movies?page=2" hx-target="#movies-grid" hx-swap="outerHTML">
     Next
   </button>
   ```

3. **Modals** - HTMX to load modal content
   ```html
   <a hx-get="/movies/123/modal" hx-target="#modal-container" hx-swap="innerHTML">
     View Details
   </a>
   ```

4. **Forms** - HTMX POST with JSON
   ```html
   <form hx-post="/api/movies" hx-swap="none" hx-on::after-request="closeModal()">
     <!-- Form fields -->
   </form>
   ```

5. **Buttons (Add/Delete)** - Single actions with feedback
   ```html
   <button
     hx-post="/api/movies"
     hx-vals='{"tmdbId": 123, "watched": true}'
     hx-swap="none"
     hx-on::after-request="showToast('Movie added')">
     Add to Watched
   </button>
   ```

---

## Key Features & User Flows

### 1. Browse Flow
1. User navigates to `/movies` or `/series`
2. Page loads with popular items from TMDB
3. User can search (updates URL with ?query=)
4. User can paginate through results
5. Click on item opens modal

### 2. Add to Library Flow
1. User clicks "Seen" or "Watchlist" on browse card
2. HTMX POST to `/api/movies` or `/api/series`
3. Server validates, creates DB record
4. Response shows success toast
5. Item added to user's library

### 3. Library Management Flow
1. Navigate to `/library/movies/watched`
2. View all watched movies
3. Click item to open edit modal
4. Update score (5-star rating)
5. Save updates to DB
6. Or delete item from library

### 4. Search Flow
1. Type in search bar (navbar)
2. Navigate to `/search?query=term`
3. View results for both movies and series
4. Click to view details in modal
5. Add to library from modal

### 5. Rating System
- **Frontend:** 5 stars (interactive)
- **Storage:** Converted to 0-10 scale
  - 1 star = 2.0
  - 2 stars = 4.0
  - 3 stars = 6.0
  - 4 stars = 8.0
  - 5 stars = 10.0
- **Display:** Show both user score and TMDB score on cards

### 6. Authentication Flow
1. Visit any protected route
2. Middleware checks session
3. If not authenticated, redirect to `/login`
4. Click OAuth provider button
5. Complete OAuth flow
6. Redirect back to `/movies`

---

## UI Components

### Layout Components

#### 1. Navbar
**Features:**
- Logo (FilmIcon + "ReelScore")
- Search input (submits to `/search?query=`)
- Dropdown: Library (Movies, Series)
- Dropdown: Browse (Movies, Series)
- Theme toggle (dark/light mode)
- Logout button

**HTML Structure:**
```html
<header class="navbar">
  <a href="/" class="logo">
    <svg><!-- Film icon --></svg>
    <span>ReelScore</span>
  </a>

  <nav>
    <form action="/search" method="get">
      <input type="text" name="query" placeholder="Search...">
    </form>

    <div class="dropdown">
      <button>Library</button>
      <ul>
        <li><a href="/library/movies/watched">Movies</a></li>
        <li><a href="/library/series/watchlist">Series</a></li>
      </ul>
    </div>

    <div class="dropdown">
      <button>Browse</button>
      <ul>
        <li><a href="/movies">Movies</a></li>
        <li><a href="/series">Series</a></li>
      </ul>
    </div>

    <button id="theme-toggle">Toggle Theme</button>
    <a href="/auth/logout">Logout</a>
  </nav>
</header>
```

#### 2. Card Component
**Props:**
- posterPath (string | null)
- title (string)
- date (Date | null) - show year
- tmdbScore (float64)
- score (float64 | null) - user's score
- children (buttons/actions)

**HTML Structure:**
```html
<div class="card">
  <div class="card-image">
    {{if .PosterPath}}
      <img src="{{.ImageURL}}" alt="{{.Title}}">
    {{else}}
      <p class="no-poster">{{.Title}}</p>
    {{end}}
  </div>

  <div class="card-overlay">
    <div class="card-header">
      <p>{{.Title}} {{if .Date}}({{.Year}}){{end}}</p>

      <div class="scores">
        {{if .Score}}
        <span class="user-score">
          <span class="label">YOU</span>
          <span>{{.Score}}</span>
          <svg><!-- Star icon --></svg>
        </span>
        {{end}}

        <span class="tmdb-score">
          <span class="label">TMDB</span>
          <span>{{.TmdbScore}}</span>
          <svg><!-- Star icon --></svg>
        </span>
      </div>
    </div>

    <div class="card-actions">
      {{.Children}}
    </div>
  </div>
</div>
```

#### 3. Rating Component
**Props:**
- rating (int 0-5)
- setRating (callback)
- readOnly (bool)

**HTML:**
```html
<div class="rating" data-rating="{{.Rating}}">
  {{range $i := seq 5}}
  <svg class="star {{if lt $i $.Rating}}filled{{end}}"
       data-value="{{$i}}"
       onclick="setRating({{$i}})">
    <!-- Star icon -->
  </svg>
  {{end}}
</div>
```

#### 4. Pagination Component
**Props:**
- currentPage (int)
- totalPages (int)

**HTML:**
```html
<div class="pagination">
  {{if gt .CurrentPage 1}}
  <button hx-get="?page={{sub .CurrentPage 1}}"
          hx-target="#content"
          hx-swap="outerHTML">
    Previous
  </button>
  {{end}}

  <span>Page {{.CurrentPage}} of {{.TotalPages}}</span>

  {{if lt .CurrentPage .TotalPages}}
  <button hx-get="?page={{add .CurrentPage 1}}"
          hx-target="#content"
          hx-swap="outerHTML">
    Next
  </button>
  {{end}}
</div>
```

#### 5. Button Component
**Variants:**
- primary (blue background)
- secondary (outline)
- ghost (transparent)
- card (small, for card actions)

**Sizes:**
- sm, md, lg, icon, card

**Props:**
- variant, size, icon, text, onClick/hx-post

#### 6. Modal Component
**Structure:**
```html
<div class="modal-backdrop" onclick="closeModal()">
  <div class="modal-content" onclick="event.stopPropagation()">
    <button class="modal-close" onclick="closeModal()">×</button>
    {{.Content}}
  </div>
</div>
```

### Page Templates

#### Login Page
```html
<main class="login-page">
  <h1>ReelScore</h1>
  <div class="auth-buttons">
    <a href="/auth/github/login" class="btn btn-github">
      <svg><!-- GitHub icon --></svg>
      Sign in with GitHub
    </a>
    <a href="/auth/google/login" class="btn btn-google">
      <svg><!-- Google icon --></svg>
      Sign in with Google
    </a>
  </div>
</main>
```

#### Browse Page (Movies/Series)
```html
{{template "navbar" .}}

<main class="browse-page">
  <div id="content">
    {{template "pagination" .Pagination}}

    <div class="cards-grid" id="movies-grid">
      {{range .Results}}
        {{template "card" .}}
      {{end}}
    </div>
  </div>
</main>

<div id="modal-container"></div>
```

#### Library Page
```html
{{template "navbar" .}}

<main class="library-page">
  <h1>{{.Title}}</h1>

  <form hx-get="{{.CurrentPath}}" hx-target="#content" hx-swap="outerHTML">
    <input type="text" name="query" placeholder="Filter..." value="{{.Query}}">
  </form>

  <div id="content">
    {{template "pagination" .Pagination}}

    <div class="cards-grid">
      {{range .Results}}
        {{template "card" .}}
      {{end}}
    </div>
  </div>
</main>
```

---

## Configuration

### Environment Variables

#### API (Go Backend)
```bash
# Server
NODE_ENV=local              # local, development, production
PORT=4000
HOST=http://localhost:4000
CLIENT_URL=http://localhost:3000
TRPC_PREFIX=/api            # Not applicable in Go, just use /api

# Database
DATABASE_URL=postgresql://psql:psql@localhost:5432/moviedb?sslmode=disable

# Redis
REDIS_HOST=localhost
REDIS_PORT=6379
REDIS_PASSWORD=
REDIS_TLS=false

# Session
SECRET_KEY=your-secret-key-min-32-chars

# OAuth
GOOGLE_CLIENT_ID=your-google-client-id
GOOGLE_CLIENT_SECRET=your-google-client-secret
GITHUB_CLIENT_ID=your-github-client-id
GITHUB_CLIENT_SECRET=your-github-client-secret

# TMDB
TMDB_KEY=your-tmdb-api-key
TMDB_URL=https://api.themoviedb.org
```

#### Frontend (Go Templates - same .env)
```bash
# Image URLs
TMDB_IMAGE_URL=https://image.tmdb.org/t/p/w500
```

### Docker Setup

**docker-compose.yml:**
```yaml
version: '3.8'
services:
  postgres:
    image: postgres:15.5
    ports:
      - 5432:5432
    environment:
      - POSTGRES_USER=psql
      - POSTGRES_PASSWORD=psql
      - POSTGRES_DB=moviedb
    volumes:
      - postgres_data:/var/lib/postgresql/data

  redis:
    image: redis:latest
    ports:
      - 6379:6379
    volumes:
      - redis_data:/var/lib/redis/data

volumes:
  postgres_data:
  redis_data:
```

### Rate Limiting
```go
// Current implementation uses Redis-backed rate limiting
// 100 requests per minute in production
// 1000 requests per minute in local/dev
// Key: User ID if authenticated, else IP address
```

### CORS Configuration
```go
allowedOrigins := []string{
    os.Getenv("CLIENT_URL"),
    "http://localhost:3000",
    "http://localhost:3001",
}

corsConfig := cors.Config{
    AllowOrigins:     allowedOrigins,
    AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "PATCH"},
    AllowCredentials: true,
    MaxAge:           86400, // 24 hours
}
```

### Security Headers (Helmet equivalent)
```go
// In production:
// - ContentSecurityPolicy
// - X-Frame-Options: DENY
// - X-Content-Type-Options: nosniff
// - Referrer-Policy: no-referrer
```

---

## Migration Notes

### Key Differences

1. **No tRPC:** Replace with standard REST API
   - JSON request/response bodies
   - Standard HTTP methods (GET, POST, PUT, DELETE)
   - Error handling via HTTP status codes

2. **No React/Next.js:** Replace with Go templates + HTMX
   - Server-side rendering
   - HTMX for dynamic updates
   - Vanilla JS for interactivity (rating, modals, theme toggle)

3. **Session Management:**
   - Use go-redis for session storage
   - gorilla/sessions or similar for cookie management
   - Middleware for authentication checks

4. **Validation:**
   - Replace Zod with go-playground/validator
   - Struct tags for validation rules
   - Custom error messages

5. **Database Layer:**
   - Replace Drizzle ORM with sqlx or GORM
   - Manual SQL queries or ORM methods
   - pgx driver for PostgreSQL

6. **Type Safety:**
   - Go structs instead of TypeScript interfaces
   - JSON struct tags for serialization
   - Validation tags for input validation

### Performance Considerations

1. **Connection Pooling:**
   - PostgreSQL: max 20 connections (configurable)
   - Redis: pipelining enabled

2. **Caching:**
   - Consider caching TMDB responses (15-minute TTL)
   - Redis for session and cache storage

3. **Pagination:**
   - Default page size: 27 items
   - Max limit: 100 items per request

### Error Handling

**Database Errors:**
```go
// PostgreSQL error codes to handle
// 23505 - Unique constraint violation (CONFLICT)
// 23503 - Foreign key violation (CONFLICT)
// 23502 - Not null violation (BAD_REQUEST)
// 22001 - String too long (PAYLOAD_TOO_LARGE)
// 22003 - Numeric out of range (BAD_REQUEST)
```

**HTTP Status Codes:**
- 200 OK - Success
- 201 Created - Resource created
- 400 Bad Request - Validation error
- 401 Unauthorized - Not authenticated
- 403 Forbidden - Authenticated but not authorized
- 404 Not Found - Resource not found
- 409 Conflict - Duplicate resource
- 413 Payload Too Large - Input too large
- 429 Too Many Requests - Rate limit exceeded
- 500 Internal Server Error - Server error

### Testing Strategy

1. **Unit Tests:**
   - Service layer (CRUD operations)
   - Validation logic
   - Helper functions

2. **Integration Tests:**
   - API endpoints
   - Database operations
   - OAuth flow (mocked)

3. **E2E Tests:**
   - Critical user flows
   - HTMX interactions
   - Form submissions

### Deployment

**Build Process:**
```bash
# Build binary
go build -o reelscore ./cmd/server

# Run migrations
./reelscore migrate up

# Start server
./reelscore serve
```

**Environment Setup:**
1. PostgreSQL database (managed service or Docker)
2. Redis instance (managed service or Docker)
3. Environment variables configured
4. OAuth apps registered (GitHub, Google)
5. TMDB API key obtained

### Additional Features to Consider

1. **Logging:**
   - Structured logging (zap/zerolog)
   - Request/response logging
   - Error tracking (Sentry integration)

2. **Metrics:**
   - Prometheus metrics
   - Request duration, count, errors
   - Database connection pool stats

3. **Health Checks:**
   - `/health` endpoint
   - Database connectivity check
   - Redis connectivity check

4. **Graceful Shutdown:**
   - Handle SIGTERM/SIGINT
   - Drain connections
   - Close database/Redis connections

---

## File Structure

Suggested Go project structure:

```
reelscore/
├── cmd/
│   └── server/
│       └── main.go
├── internal/
│   ├── config/
│   │   └── config.go
│   ├── database/
│   │   ├── db.go
│   │   └── migrations/
│   ├── handlers/
│   │   ├── auth.go
│   │   ├── movies.go
│   │   ├── series.go
│   │   ├── tmdb.go
│   │   └── users.go
│   ├── middleware/
│   │   ├── auth.go
│   │   ├── logging.go
│   │   └── ratelimit.go
│   ├── models/
│   │   ├── movie.go
│   │   ├── serie.go
│   │   └── user.go
│   ├── services/
│   │   ├── movie_service.go
│   │   ├── serie_service.go
│   │   ├── tmdb_service.go
│   │   └── user_service.go
│   ├── templates/
│   │   ├── layouts/
│   │   │   ├── base.html
│   │   │   └── navbar.html
│   │   ├── pages/
│   │   │   ├── login.html
│   │   │   ├── browse.html
│   │   │   └── library.html
│   │   └── components/
│   │       ├── card.html
│   │       ├── modal.html
│   │       └── pagination.html
│   └── static/
│       ├── css/
│       │   └── styles.css
│       └── js/
│           ├── htmx.min.js
│           └── app.js
├── go.mod
├── go.sum
├── .env.example
└── docker-compose.yml
```

---

## Summary Checklist

- [ ] Database schema created (3 tables: User, Movie, Serie)
- [ ] OAuth2 authentication (GitHub, Google)
- [ ] Session management (Redis)
- [ ] User CRUD service
- [ ] Movie CRUD service
- [ ] Serie CRUD service
- [ ] TMDB API integration (7 endpoints)
- [ ] API endpoints (REST instead of tRPC)
- [ ] Authentication middleware
- [ ] Rate limiting middleware
- [ ] CORS configuration
- [ ] Template rendering engine
- [ ] HTMX integration
- [ ] UI components (card, modal, pagination, rating)
- [ ] Route handlers (login, browse, library, search)
- [ ] Form validation
- [ ] Error handling
- [ ] Logging
- [ ] Docker setup (PostgreSQL, Redis)
- [ ] Environment configuration
- [ ] Health checks
- [ ] Graceful shutdown

---

## Additional Resources

- HTMX Documentation: https://htmx.org/docs/
- Go OAuth2 Library: https://github.com/markbates/goth
- Go Validator: https://github.com/go-playground/validator
- SQLX: https://github.com/jmoiron/sqlx
- Go-Redis: https://github.com/redis/go-redis
- Fiber Framework: https://gofiber.io/
- TMDB API Docs: https://developers.themoviedb.org/3

---

**End of Rebuild Guide**

This document provides a complete specification for rebuilding ReelScore with Go and HTMX. All API endpoints, database schemas, authentication flows, and UI components have been documented based on the current TypeScript implementation.
