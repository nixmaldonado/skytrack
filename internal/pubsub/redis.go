package pubsub

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/nixmaldonado/skytrack/graph/model"
	"github.com/redis/go-redis/v9"
)

// RedisClient wraps a Redis client for flight position pub/sub.
type RedisClient struct {
	client *redis.Client
}

// NewRedisClient creates a Redis connection and verifies it with a ping.
func NewRedisClient(addr string) (*RedisClient, error) {
	client := redis.NewClient(&redis.Options{Addr: addr})

	if err := client.Ping(context.Background()).Err(); err != nil {
		return nil, fmt.Errorf("redis: unable to connect to %s: %w", addr, err)
	}

	return &RedisClient{client: client}, nil
}

// channelName returns the Redis channel for a given aircraft.
func channelName(icao24 string) string {
	return "flight:" + icao24
}

// Publish sends a FlightPosition update to the Redis channel for the aircraft.
func (r *RedisClient) Publish(ctx context.Context, icao24 string, pos *model.FlightPosition) error {
	data, err := json.Marshal(pos)
	if err != nil {
		return fmt.Errorf("pubsub: marshal position: %w", err)
	}
	return r.client.Publish(ctx, channelName(icao24), data).Err()
}

// Subscribe returns a channel that receives FlightPosition updates for the given
// aircraft, and a cleanup function to close the subscription. The channel is closed
// when ctx is cancelled or cleanup is called.
func (r *RedisClient) Subscribe(ctx context.Context, icao24 string) (<-chan *model.FlightPosition, func()) {
	sub := r.client.Subscribe(ctx, channelName(icao24))
	ch := make(chan *model.FlightPosition)

	go func() {
		defer close(ch)
		redisCh := sub.Channel()

		for {
			select {
			case <-ctx.Done():
				return
			case msg, ok := <-redisCh:
				if !ok {
					return
				}
				var pos model.FlightPosition
				if err := json.Unmarshal([]byte(msg.Payload), &pos); err != nil {
					log.Printf("pubsub: unmarshal error: %v", err)
					continue
				}
				select {
				case ch <- &pos:
				case <-ctx.Done():
					return
				}
			}
		}
	}()

	cleanup := func() {
		sub.Close()
	}

	return ch, cleanup
}

// Close shuts down the Redis connection.
func (r *RedisClient) Close() error {
	return r.client.Close()
}
