# Reliquary

A self-hosted book management application — a modern, open-source alternative to Calibre and Booklore.

Reliquary lets you organise your personal library into libraries and shelves, track reading progress, and manage metadata — all from a clean, responsive web interface you host yourself.

## Tech Stack

| Layer | Technology |
|---|---|
| Frontend | SvelteKit, TypeScript, Tailwind CSS v4 |
| UI Components | shadcn-svelte, bits-ui, Lucide icons |
| Backend | Go, Chi router |
| Database | SQLite via Drizzle ORM + LibSQL |
| Auth | Better Auth (email/password) |

## Features

- **Authentication** — email/password sign up, sign in, and sign out
- **Protected routes** — all pages require authentication; handled centrally in SvelteKit hooks
- **Sidebar navigation** — collapsible sidebar with Libraries and Shelves sections, dark/light mode
- **Dark mode** — system-aware dark mode via `mode-watcher`
- **Self-hostable** — SQLite database, no external services required

## Getting Started

### Prerequisites

- Node.js 20+ and [pnpm](https://pnpm.io)
- Go 1.22+

### Frontend

```bash
cd frontend
pnpm install
```

Copy the example environment file and fill in the values:

```bash
cp .env.example .env
```

```env
DATABASE_URL=file:../reliquary.db   # path to the SQLite database
ORIGIN=http://localhost:5173        # base URL of your instance
BETTER_AUTH_SECRET=                 # random 32+ character secret
```

Run database migrations:

```bash
pnpm db:push
```

Start the dev server:

```bash
pnpm dev
```

### Backend

```bash
go run ./cmd/api
```

The API runs on port **5321** by default.

For hot reload during development, install [Air](https://github.com/air-verse/air) and run:

```bash
air
```

## Project Structure

```
reliquary/
├── cmd/
│   └── api/
│       └── main.go          # Go API entry point
├── frontend/
│   ├── src/
│   │   ├── routes/          # SvelteKit pages and server actions
│   │   │   ├── login/
│   │   │   ├── register/
│   │   │   └── +page.svelte # Dashboard
│   │   ├── lib/
│   │   │   ├── components/  # UI components (sidebar, nav, shadcn)
│   │   │   └── server/
│   │   │       ├── auth.ts  # Better Auth configuration
│   │   │       └── db/      # Drizzle schema and client
│   │   └── hooks.server.ts  # Auth middleware / route protection
│   └── drizzle.config.ts
├── go.mod
└── reliquary.db
```

## Status

Reliquary is in early development. The foundation — authentication, navigation, and database schema — is in place. Book management features (importing, metadata editing, reading progress) are actively being built.
