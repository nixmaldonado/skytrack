#!/bin/bash
set -e

echo "🛫 Bootstrapping SkyTrack..."

# Init Go module
go mod init github.com/nixmaldonado/skytrack
go get github.com/99designs/gqlgen@latest
go get github.com/vektah/gqlparser/v2@latest

# Create directory structure
mkdir -p cmd/server
mkdir -p graph/schema graph/model graph/dataloader
mkdir -p internal/database/migrations
mkdir -p internal/opensky
mkdir -p internal/repository
mkdir -p internal/service
mkdir -p internal/auth

# Create gqlgen config
cat > gqlgen.yml << 'EOF'
schema:
  - graph/schema/*.graphqls

exec:
  filename: graph/generated.go
  package: graph

model:
  filename: graph/model/models_gen.go
  package: model

resolver:
  layout: follow-schema
  dir: graph
  package: graph
  filename_template: "{name}.resolvers.go"

autobind:
  - github.com/nixmaldonado/skytrack/graph/model

omit_slice_element_pointers: true
EOF

# Create initial schema — Phase 1
cat > graph/schema/schema.graphqls << 'EOF'
type Query {
  airports(first: Int = 25, after: String): AirportConnection!
  airport(icao: String!): Airport
  airportByIata(iata: String!): Airport
}

type Airport {
  id: ID!
  icao: String!
  iata: String
  name: String!
  city: String
  country: String!
  latitude: Float!
  longitude: Float!
  elevation: Int
  type: AirportType!
}

enum AirportType {
  LARGE
  MEDIUM
  SMALL
  HELIPORT
  SEAPLANE_BASE
  CLOSED
}

type AirportConnection {
  edges: [AirportEdge!]!
  pageInfo: PageInfo!
  totalCount: Int!
}

type AirportEdge {
  cursor: String!
  node: Airport!
}

type PageInfo {
  hasNextPage: Boolean!
  hasPreviousPage: Boolean!
  startCursor: String
  endCursor: String
}
EOF

# Create docker-compose
cat > docker-compose.yml << 'EOF'
services:
  postgres:
    image: postgres:16-alpine
    environment:
      POSTGRES_USER: skytrack
      POSTGRES_PASSWORD: skytrack
      POSTGRES_DB: skytrack
    ports:
      - "5432:5432"
    volumes:
      - pgdata:/var/lib/postgresql/data

  redis:
    image: redis:7-alpine
    ports:
      - "6379:6379"

volumes:
  pgdata:
EOF

# Create .env.example
cat > .env.example << 'EOF'
DATABASE_URL=postgres://skytrack:skytrack@localhost:5432/skytrack?sslmode=disable
REDIS_URL=redis://localhost:6379
OPENSKY_USERNAME=
OPENSKY_PASSWORD=
PORT=8080
JWT_SECRET=dev-secret-change-me
EOF

# Create Makefile
cat > Makefile << 'EOF'
.PHONY: generate run test lint migrate-up migrate-down seed

generate:
  go run github.com/99designs/gqlgen generate

run:
  go run cmd/server/main.go

test:
  go test ./... -v -race

lint:
  golangci-lint run ./...

migrate-up:
  go run -tags migrate cmd/migrate/main.go up

migrate-down:
  go run -tags migrate cmd/migrate/main.go down

seed:
  go run cmd/seed/main.go

docker-up:
  docker compose up -d

docker-down:
  docker compose down
EOF

# Create .gitignore
cat > .gitignore << 'EOF'
.env
*.exe
vendor/
tmp/
EOF

echo ""
echo "✅ SkyTrack scaffolded!"
echo ""
echo "Next steps:"
echo "  1. Update go.mod module path with your GitHub username"
echo "  2. Update gqlgen.yml autobind path to match"
echo "  3. Run 'make generate' to generate resolvers"
echo "  4. Implement resolver stubs in graph/schema.resolvers.go"
echo "  5. Ask Claude Code: 'Let's implement Phase 1 — wire up the airport query with hardcoded data first'"
EOF