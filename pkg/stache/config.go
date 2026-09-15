package stache

import (
	"context"
	"sync"
	"time"
)

var ( 
	defaultClient *Client
	defaultOnce sync.Once
)

func Default() *Client { 
	defaultOnce.Do(
		func() { 
			if defaultClient == nil {
				defaultClient = NewClient(Config{})
			}
		})
	return defaultClient
}

func SetDefault(c *Client) { 
	defaultClient = c
}

func EasyRemember[T any](
	ctx context.Context,
	key string, 
	ttl time.Duration,
	fn func(ctx context.Context) (T, error),
) (T, error) {
	return Remember[T](ctx, Default(),key, ttl, fn)
}