package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/handler/extension"
	"github.com/99designs/gqlgen/graphql/handler/transport"
	"github.com/99designs/gqlgen/graphql/playground"
	"github.com/gorilla/websocket"

	"github.com/nixmaldonado/skytrack/graph"
	"github.com/nixmaldonado/skytrack/graph/dataloader"
	"github.com/nixmaldonado/skytrack/internal/opensky"
	"github.com/nixmaldonado/skytrack/internal/poller"
	"github.com/nixmaldonado/skytrack/internal/pubsub"
	"github.com/nixmaldonado/skytrack/internal/store"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	redisAddr := os.Getenv("REDIS_URL")
	if redisAddr == "" {
		redisAddr = "localhost:6379"
	}

	// Initialize dependencies
	store := store.NewStore()

	redisPubSub, err := pubsub.NewRedisClient(redisAddr)
	if err != nil {
		log.Fatalf("Failed to connect to Redis: %v\nRun: docker compose up -d redis", err)
	}
	defer redisPubSub.Close()

	openskyClient := opensky.NewClient()

	flightPoller := poller.NewPoller(openskyClient, redisPubSub, 10*time.Second)

	// Build GraphQL handler
	srv := handler.New(graph.NewExecutableSchema(graph.Config{
		Resolvers: &graph.Resolver{
			Store:  store,
			Poller: flightPoller,
			PubSub: redisPubSub,
		},
	}))

	// WebSocket transport must be registered first for subscription support.
	// It checks for the Upgrade: websocket header and handles the graphql-transport-ws protocol.
	srv.AddTransport(transport.Websocket{
		KeepAlivePingInterval: 10 * time.Second,
		Upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				return true // allow all origins in development
			},
		},
	})
	srv.AddTransport(transport.Options{})
	srv.AddTransport(transport.GET{})
	srv.AddTransport(transport.POST{})
	srv.AddTransport(transport.MultipartForm{})
	srv.Use(extension.Introspection{})

	http.Handle("/", playground.Handler("SkyTrack", "/query"))
	http.Handle("/query", dataloader.Middleware(store, srv))

	// Graceful shutdown: cancel context on SIGINT/SIGTERM to stop the poller.
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	flightPoller.Start(ctx)

	httpServer := &http.Server{Addr: ":" + port}

	go func() {
		log.Printf("SkyTrack GraphQL playground at http://localhost:%s/", port)
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("HTTP server error: %v", err)
		}
	}()

	// Wait for shutdown signal
	<-ctx.Done()
	log.Println("Shutting down...")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := httpServer.Shutdown(shutdownCtx); err != nil {
		log.Printf("HTTP shutdown error: %v", err)
	}
}
