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

### 2. Deployment
Launch the stable production stack:

```bash
docker compose up -d --build
```

Reliquary will be available at `http://localhost:3000` (or whatever you set in `ORIGIN`).

---

## Running Stable & Dev Simultaneously

If you want to use Reliquary to manage your real library while also developing it, you can run two independent instances by using Docker **Project Names** (`-p`).

### 1. The Stable Instance (Full Stack)
This runs your "actual" app in the background using the default `docker-compose.yml`.
1.  Copy `.env.example` to `.env` and configure your real paths/passwords.
    ```bash
    cp .env.example .env
    ```
2.  Launch:
    ```bash
    docker compose -p reliquary up -d --build
    ```

### 2. The Development Environment (Database Only)
This provides the database for your local `air` and `pnpm dev` processes using `docker-compose.dev.yml`.
1.  Copy `.env.dev.example` to `.env` (in the root).
2. ```bash
    cp .env.dev.example .env.dev
    ```
2.  Launch the database:
    ```bash
    docker compose -p reliquary-dev -f docker-compose.dev.yml up -d --build
    ```
3.  Run your local dev tools: `air` and `pnpm dev`.

By using `-p`, Docker creates separate volumes and networks for each, so your "real" database and your "test" database never touch each other.

---

## Development Setup

Reliquary consists of a Go backend and a SvelteKit frontend.

### Prerequisites

- **Go**: 1.25+
- **Node.js**: 22+
- **pnpm**: 9+
- **PostgreSQL**: 16+

### 1. Infrastructure
Start the development database:
```bash
docker compose -f docker-compose.dev.yml up -d
```

### 2. Backend (Go)
```bash
go mod download
air
```

### 3. Frontend (SvelteKit)
```bash
cd frontend
pnpm install
pnpm db:push
pnpm dev
```

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

## License

MIT
