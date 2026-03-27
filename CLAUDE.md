# SkyTrack — Real-Time Flight Tracker GraphQL API

## Project Overview
A Go GraphQL API for real-time flight tracking, built as a learning project to master GraphQL concepts incrementally. Uses **gqlgen** (schema-first, code-generated) with real data from the OpenSky Network API.

## Tech Stack
- **Language:** Go 1.22+
- **GraphQL:** github.com/99designs/gqlgen
- **Database:** PostgreSQL (reference data: airports, airlines, routes)
- **Cache/PubSub:** Redis (subscriptions, caching OpenSky responses)
- **HTTP Router:** chi (compatible with gqlgen out of the box)
- **DataLoader:** github.com/graph-gophers/dataloader/v7
- **Migrations:** github.com/golang-migrate/migrate/v4
- **Testing:** standard library + github.com/stretchr/testify

## External Data Sources
- **OpenSky Network API** (https://opensky-network.org/api) — live ADS-B aircraft positions. Free, no key needed for anonymous (10 req/min). Docs: https://openskynetwork.github.io/opensky-api/rest.html
- **OurAirports CSV** (https://ourairports.com/data/) — static airport/runway/frequency reference data, seed into Postgres

## Project Structure
```
skytrack/
├── CLAUDE.md
├── cmd/
│   └── server/
│       └── main.go            # Entrypoint
├── graph/
│   ├── schema/
│   │   ├── schema.graphqls     # Root Query/Mutation/Subscription
│   │   ├── airport.graphqls
│   │   ├── flight.graphqls
│   │   └── airline.graphqls
│   ├── model/
│   │   └── models_gen.go       # gqlgen generated (DO NOT EDIT)
│   ├── generated.go            # gqlgen generated (DO NOT EDIT)
│   ├── resolver.go             # Root resolver struct + dependencies
│   ├── schema.resolvers.go     # Resolver implementations
│   └── dataloader/
│       └── loaders.go          # DataLoader definitions
├── internal/
│   ├── database/
│   │   ├── postgres.go         # Connection pool setup
│   │   └── migrations/         # SQL migration files
│   ├── opensky/
│   │   └── client.go           # OpenSky API client
│   ├── repository/             # Data access layer (Postgres queries)
│   ├── service/                # Business logic layer
│   └── auth/                   # JWT middleware + directives
├── gqlgen.yml                  # gqlgen configuration
├── go.mod
├── go.sum
├── .env.example
├── docker-compose.yml          # Postgres + Redis
└── Makefile
```

## Learning Progression (build in this order)

### Phase 1: Schema & Basic Queries ✅ START HERE
- Define SDL schema for Airport, Airline types
- Run `go run github.com/99designs/gqlgen generate`
- Implement resolvers that return hardcoded data first
- Then wire up Postgres repository layer
- **GraphQL concepts:** schema definition, types, queries, resolvers, field resolvers

### Phase 2: Mutations & Input Types
- Add CreateAirport, UpdateAirport mutations
- Custom scalars: Coordinates (lat/lng), ICAOCode, IATACode
- Input validation in resolvers
- **GraphQL concepts:** mutations, input types, custom scalars, error handling

### Phase 3: Relationships & DataLoaders
- Flight → Airline, Flight → DepartureAirport, Flight → ArrivalAirport
- Demonstrate N+1 problem first (log SQL queries)
- Then add DataLoaders to batch-resolve
- **GraphQL concepts:** nested resolvers, N+1 problem, DataLoader pattern

### Phase 4: Pagination & Filtering
- Relay-style cursor pagination for flights list
- Filter flights by: airline, departure airport, status, altitude range
- **GraphQL concepts:** connections, edges, pageInfo, cursor encoding, arguments

### Phase 5: Subscriptions (the showcase feature)
- WebSocket subscription for live flight positions from OpenSky
- Redis PubSub to fan out updates
- `subscription { trackFlight(icao24: "...") { latitude longitude altitude velocity heading } }`
- **GraphQL concepts:** subscriptions, WebSocket transport, server-sent events

### Phase 6: Auth & Directives
- JWT middleware on chi router
- Schema directive `@auth(requires: ROLE)` for protected mutations
- **GraphQL concepts:** context propagation, custom directives, middleware

### Phase 7: Federation (stretch goal)
- Split into flights-service and airports-service
- Apollo Federation with gqlgen federation plugin
- **GraphQL concepts:** federated schemas, entities, @key directive

## Commands
```bash
make generate    # Run gqlgen codegen
make run         # Start server (default :8080)
make migrate-up  # Run DB migrations
make seed        # Seed airport/airline data from CSVs
make test        # Run tests
make lint        # golangci-lint
```

## Coding Conventions
- Standard Go project layout (cmd/, internal/, graph/)
- Repository pattern for data access — no SQL in resolvers
- Resolvers are thin: validate input → call service → return
- Errors: use gqlgen's graphql.Error for user-facing, fmt.Errorf for internal
- Context: propagate request-scoped values (user, request ID) via context.Context
- Naming: Go conventions (PascalCase exports, camelCase private), GraphQL conventions in SDL (camelCase fields, PascalCase types)
- All new features get at least one resolver test using a test helper that sets up an in-memory resolver

## Environment
```
DATABASE_URL=postgres://skytrack:skytrack@localhost:5432/skytrack?sslmode=disable
REDIS_URL=redis://localhost:6379
OPENSKY_USERNAME=         # optional, increases rate limit
OPENSKY_PASSWORD=         # optional
PORT=8080
JWT_SECRET=dev-secret
```

## gqlgen Workflow Reminder
1. Edit .graphqls schema files
2. Run `make generate`
3. Implement the new resolver stubs in schema.resolvers.go
4. Never edit generated.go or models_gen.go manually
