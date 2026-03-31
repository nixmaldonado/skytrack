package dataloader

import (
	"context"
	"net/http"

	"github.com/graph-gophers/dataloader/v7"

	"github.com/nixmaldonado/skytrack/graph/model"
	"github.com/nixmaldonado/skytrack/internal/store"
)

type contextKey string

const loadersKey contextKey = "dataloaders"

type Loaders struct {
	AirlineLoader *dataloader.Loader[string, *model.Airline]
	AirportLoader *dataloader.Loader[string, *model.Airport]
}

func NewLoaders(store *store.Store) *Loaders {
	return &Loaders{
		AirlineLoader: dataloader.NewBatchedLoader(func(ctx context.Context, ids []string) []*dataloader.Result[*model.Airline] {
			airlines := store.FindAirlinesByIDs(ids)

			result := make([]*dataloader.Result[*model.Airline], len(ids))
			for i, id := range ids {
				result[i] = &dataloader.Result[*model.Airline]{Data: airlines[id]}
			}

			return result
		}),

		AirportLoader: dataloader.NewBatchedLoader(func(ctx context.Context, ids []string) []*dataloader.Result[*model.Airport] {
			airports := store.FindAirportsByIDs(ids)

			result := make([]*dataloader.Result[*model.Airport], len(ids))
			for i, id := range ids {
				result[i] = &dataloader.Result[*model.Airport]{Data: airports[id]}
			}

			return result
		}),
	}
}

func FromContext(ctx context.Context) *Loaders {
	return ctx.Value(loadersKey).(*Loaders)
}

func Middleware(store *store.Store, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		loaders := NewLoaders(store)
		ctx := context.WithValue(r.Context(), loadersKey, loaders)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
