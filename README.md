# Reliquary 📚

Reliquary is a modern, self-hosted book management application designed as an open-source alternative to Calibre and Booklore. It provides a clean, responsive web interface for organizing personal book collections, tracking reading progress, and managing metadata.

![Reliquary Screenshot](https://raw.githubusercontent.com/j-m-m-y/reliquary/main/screenshot.png) *(Placeholder: Update with real screenshot)*

## Features

- **Multi-Library Support**: Organize your books into logical libraries and custom shelves.
- **Reading Progress**: Track exactly where you are in every book with reading sessions.
- **Metadata Management**: Automatic extraction from EPUB and PDF files.
- **Bookdrop Workflow**: Scan directories for new books and import them with a review step.
- **Modern Tech Stack**: Built with Svelte 5 (Runes), Go, and PostgreSQL.
- **Self-Hosted**: Full control over your data and files.

## Installation (Docker)

The easiest way to run Reliquary is using Docker Compose.

### 1. Prepare Environment
Copy `.env.example` to `.env` and fill in your values.

```bash
cp .env.example .env
```

**Key Variables:**
- `POSTGRES_PASSWORD`: A secure password for the database.
- `BETTER_AUTH_SECRET`: A random 32+ character secret (Generate with `openssl rand -hex 32`).
- `ORIGIN`: The public URL of your instance (e.g., `http://localhost:3000`).
- `BOOKS_PATH`: Path to your book library on the host machine.

### 2. Deployment
Reliquary includes a PostgreSQL database in its Compose stack. If you set `DATABASE_URL` manually in your `.env`, it will prioritize that over the internal service.

```bash
docker compose up -d
```

Reliquary will be available at `http://localhost:3000` (or whatever you set in `ORIGIN`).

---

## Development Setup

Reliquary consists of a Go backend and a SvelteKit frontend.

### Prerequisites

- **Go**: 1.25+
- **Node.js**: 22+
- **pnpm**: 9+
- **PostgreSQL**: 16+

### 1. Infrastructure
Copy `.env.dev.example` to `.env` in the root and start the database.

```bash
cp .env.dev.example .env
docker compose up -d postgres
```

### 2. Backend (Go)
The backend handles file operations and metadata extraction.
```bash
# Install dependencies
go mod download

# Run with hot-reload (requires Air)
air
```
*The Go API runs on `http://localhost:5321` by default.*

### 3. Frontend (SvelteKit)
The frontend handles the UI and authentication.
```bash
cd frontend
pnpm install

# Sync database schema
pnpm db:push

# Start development server
pnpm dev
```
*The dev server runs on `http://localhost:5173`.*

---

## Key Commands

| Command | Location | Description |
|---|---|---|
| `air` | Root | Start Go backend with hot-reload |
| `pnpm dev` | `frontend/` | Start SvelteKit dev server |
| `pnpm db:push` | `frontend/` | Sync schema to database |
| `pnpm db:studio` | `frontend/` | Open Drizzle Studio |
| `pnpm lint` | `frontend/` | Run ESLint and Prettier |
| `pnpm check` | `frontend/` | Run Svelte-check |

## Project Structure

- `cmd/api/`: Go application entry point.
- `internal/`: Backend logic (handlers, models, database access).
- `frontend/`: SvelteKit application.
  - `src/lib/api/`: Reactive API client classes (Svelte 5 runes).
  - `src/lib/server/db/`: Drizzle schema and database configuration.
- `books/`: Default directory for book storage (configurable).

## License

MIT
