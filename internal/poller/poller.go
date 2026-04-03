package poller

import (
	"context"
	"errors"
	"log"
	"sync"
	"time"

	"github.com/nixmaldonado/skytrack/graph/model"
	"github.com/nixmaldonado/skytrack/internal/opensky"
	"github.com/nixmaldonado/skytrack/internal/pubsub"
)

// Poller periodically fetches aircraft positions from OpenSky and publishes
// them to Redis. It uses reference counting so multiple subscribers watching
// the same aircraft only trigger one poll.
type Poller struct {
	opensky  *opensky.Client
	pubsub   *pubsub.RedisClient
	interval time.Duration

	mu      sync.Mutex
	tracked map[string]int // icao24 -> subscriber count
}

func NewPoller(osClient *opensky.Client, ps *pubsub.RedisClient, interval time.Duration) *Poller {
	return &Poller{
		opensky:  osClient,
		pubsub:   ps,
		interval: interval,
		tracked:  make(map[string]int),
	}
}

// Track increments the subscriber count for an aircraft.
func (p *Poller) Track(icao24 string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.tracked[icao24]++
	log.Printf("poller: tracking %s (subscribers: %d)", icao24, p.tracked[icao24])
}

// Untrack decrements the subscriber count. When it reaches 0, the aircraft
// is removed and the poller stops fetching it.
func (p *Poller) Untrack(icao24 string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.tracked[icao24]--
	if p.tracked[icao24] <= 0 {
		delete(p.tracked, icao24)
		log.Printf("poller: stopped tracking %s (no subscribers)", icao24)
	} else {
		log.Printf("poller: untracked %s (subscribers: %d)", icao24, p.tracked[icao24])
	}
}

// Start launches the background polling loop. It runs until ctx is cancelled.
func (p *Poller) Start(ctx context.Context) {
	go func() {
		ticker := time.NewTicker(p.interval)
		defer ticker.Stop()

		log.Printf("poller: started (interval: %s)", p.interval)

		for {
			select {
			case <-ctx.Done():
				log.Println("poller: stopped")
				return
			case <-ticker.C:
				p.pollAll(ctx)
			}
		}
	}()
}

// pollAll fetches the state of every tracked aircraft and publishes updates.
func (p *Poller) pollAll(ctx context.Context) {
	p.mu.Lock()
	// Copy the tracked set so we don't hold the lock during HTTP calls.
	aircraft := make([]string, 0, len(p.tracked))
	for icao24 := range p.tracked {
		aircraft = append(aircraft, icao24)
	}
	p.mu.Unlock()

	if len(aircraft) == 0 {
		return
	}

	for _, icao24 := range aircraft {
		if ctx.Err() != nil {
			return
		}

		sv, err := p.opensky.GetAircraftState(ctx, icao24)
		if err != nil {
			if errors.Is(err, opensky.ErrRateLimited) {
				log.Printf("poller: rate limited by OpenSky, skipping remaining aircraft this tick")
				return
			}
			log.Printf("poller: error fetching %s: %v", icao24, err)
			continue
		}

		if sv == nil {
			// Aircraft not broadcasting — skip silently.
			continue
		}

		pos := stateVectorToPosition(sv)
		if err := p.pubsub.Publish(ctx, icao24, pos); err != nil {
			log.Printf("poller: error publishing %s: %v", icao24, err)
		}
	}
}

func stateVectorToPosition(sv *opensky.StateVector) *model.FlightPosition {
	pos := &model.FlightPosition{
		Icao24:   sv.Icao24,
		OnGround: sv.OnGround,
	}

	if sv.Callsign != "" {
		pos.Callsign = &sv.Callsign
	}
	pos.Latitude = sv.Latitude
	pos.Longitude = sv.Longitude
	pos.Altitude = sv.BaroAltitude
	pos.Velocity = sv.Velocity
	pos.Heading = sv.TrueTrack
	pos.VerticalRate = sv.VerticalRate

	if sv.TimePosition != nil {
		ts := int(*sv.TimePosition)
		pos.Timestamp = ts
	} else {
		pos.Timestamp = int(time.Now().Unix())
	}

	return pos
}
