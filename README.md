# ReelScore

A movie and TV series tracking application built with Go and HTMX.

## Features

- Track movies and TV series you've watched or want to watch
- Rate content with a 5-star system
- Browse popular movies and series from TMDB
- Search across movies and TV shows
- OAuth authentication (GitHub, Google)

## Tech Stack

- **Backend**: Go (standard library)
- **Frontend**: Go templates + HTMX
- **Database**: PostgreSQL
- **Cache/Sessions**: Redis
- **Authentication**: OAuth2 (GitHub, Google)
- **External API**: TMDB (The Movie Database)

## Prerequisites

- Go 1.23 or higher
- Docker and Docker Compose (for PostgreSQL and Redis)
- TMDB API key ([Get one here](https://www.themoviedb.org/settings/api))
- OAuth credentials:
  - GitHub OAuth App ([Create one](https://github.com/settings/developers))
  - Google OAuth Client ([Create one](https://console.cloud.google.com/apis/credentials))

## Getting Started

### 1. Clone the repository

```bash
git clone https://github.com/liamwears/reelscore.git
cd reelscore
```

### 2. Set up environment variables

```bash
cp .env.example .env
# Edit .env with your actual credentials
```

### 3. Start database and Redis

```bash
docker-compose up -d
```

### 4. Install dependencies

```bash
go mod download
```

### 5. Run database migrations

```bash
make migrate
```

### 6. Start the server

```bash
go run cmd/server/main.go
```

The server will start on `http://localhost:4000` (or the port specified in your `.env` file).

## Development

### Project Structure

```
reelscore/
├── cmd/
│   └── server/          # Application entry point
├── internal/
│   ├── config/          # Configuration management
│   ├── database/        # Database connection and migrations
│   ├── handlers/        # HTTP request handlers
│   ├── middleware/      # HTTP middleware (auth, logging, etc.)
│   ├── models/          # Data models
│   ├── services/        # Business logic
│   ├── templates/       # HTML templates
│   └── static/          # CSS, JS, images
├── .env.example         # Example environment variables
├── docker-compose.yml   # Docker services (Postgres, Redis)
└── go.mod              # Go module dependencies
```

### Running Tests

```bash
go test ./...
```

### Building for Production

```bash
go build -o reelscore cmd/server/main.go
./reelscore
```

## OAuth Setup

### GitHub OAuth App

1. Go to [GitHub Developer Settings](https://github.com/settings/developers)
2. Click "New OAuth App"
3. Fill in:
   - Application name: ReelScore
   - Homepage URL: `http://localhost:4000`
   - Authorization callback URL: `http://localhost:4000/auth/github/callback`
4. Copy the Client ID and Client Secret to your `.env` file

### Google OAuth Client

1. Go to [Google Cloud Console](https://console.cloud.google.com/apis/credentials)
2. Create a new OAuth 2.0 Client ID
3. Configure:
   - Application type: Web application
   - Authorized redirect URIs: `http://localhost:4000/auth/google/callback`
4. Copy the Client ID and Client Secret to your `.env` file

## License

MIT

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.
