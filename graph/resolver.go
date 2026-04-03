package graph

// This file will not be regenerated automatically.
//
// It serves as dependency injection for your app, add any dependencies you require
// here.

import (
	"github.com/nixmaldonado/skytrack/internal/poller"
	"github.com/nixmaldonado/skytrack/internal/pubsub"
	"github.com/nixmaldonado/skytrack/internal/store"
)

type Resolver struct {
	Store  *store.Store
	Poller *poller.Poller
	PubSub *pubsub.RedisClient
}
