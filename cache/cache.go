package cache

import (
	"context"
	"errors"
	"time"
)

var (
	// ErrItemExpired is returned in Cache.Get when the item found in the cache
	// has expired.
	ErrItemExpired error = errors.New("item has expired")
	// ErrKeyNotFound is returned in Cache.Get and Cache.Delete when the
	// provided key could not be found in cache.
	ErrKeyNotFound error = errors.New("key not found in cache")
)

type Cache interface {
	Get(ctx context.Context, key string) (any, time.Time, error)
	Set(ctx context.Context, key string, value any, d time.Duration) error
	SetNX(ctx context.Context, key string, value any, d time.Duration) (bool, error)
	Delete(ctx context.Context, key string) error
	String() string
}

type Item struct {
	Value      any
	Expiration int64
}

// Expired returns true if the item has expired.
func (i *Item) Expired() bool {
	if i.Expiration == 0 {
		return false
	}

	return time.Now().UnixNano() > i.Expiration
}
