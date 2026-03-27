# SkyTrack ✈️

A real-time flight tracking GraphQL API built with Go, designed to demonstrate modern GraphQL patterns including subscriptions, DataLoaders, cursor pagination, and schema directives.

## Features

- **Live Flight Tracking** — Real-time aircraft positions via OpenSky Network ADS-B data
- **GraphQL Subscriptions** — WebSocket-based live updates for tracked flights
- **Cursor Pagination** — Relay-style connections for browsing airports, airlines, and flights
- **DataLoaders** — Efficient batched data fetching, solving the N+1 problem
- **Auth Directives** — JWT authentication with schema-level `@auth` directives
- **Schema-First** — Built with [gqlgen](https://github.com/99designs/gqlgen), Go's leading GraphQL library

## Tech Stack

| Component | Technology |
|-----------|-----------|
| Language | Go 1.22+ |
| GraphQL | gqlgen |
| Database | PostgreSQL 16 |
| Cache / PubSub | Redis 7 |
| Live Data | OpenSky Network API |
| Reference Data | OurAirports.com |

## Quick Start

```bash
# Clone and setup
git clone https://github.com/yourusername/skytrack.git
cd skytrack
cp .env.example .env

# Start dependencies
make docker-up

# Run migrations and seed data
make migrate-up
make seed

# Generate GraphQL code and start
make generate
make run
```

Open the GraphQL Playground at `http://localhost:8080/` and try:

```graphql
query {
  airports(first: 5) {
    edges {
      node {
        icao
        name
        city
        country
      }
    }
    pageInfo {
      hasNextPage
      endCursor
    }
  }
}
```

## Architecture

```
cmd/server/         → HTTP server entrypoint
graph/schema/       → GraphQL SDL schema files (source of truth)
graph/              → Generated code + resolver implementations
internal/database/  → Postgres connection + migrations
internal/opensky/   → OpenSky Network API client
internal/repository/→ Data access layer
internal/service/   → Business logic
internal/auth/      → JWT + directive authorization
```

## License

MIT
