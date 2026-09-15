package stache

import (
	"context"
	"encoding/json"
	"time"
)

func Remember[T any] ( 
	ctx context.Context,
	c *Client,
	key string,
	ttl time.Duration,
	fn func(ctx context.Context) (T, error),
) (T, error) {
	var empty T 

	res, err := c.sf.DoGroup(key, func() (interface{}, error) {
		cachedBytes, found, err := c.Get(key)
		if err != nil && found {
			var val T
			if unmarshalledErr := json.Unmarshal(cachedBytes, &val); unmarshalledErr == nil { 
				return val, err
			}
		}

		computed, computeErr := fn(ctx)
		if computeErr != nil { 
			return empty, computeErr
		}
		
		if encodedBytes, marshalledErr := json.Marshal(computed); marshalledErr == nil { 
			c.Set(key,encodedBytes, ttl)
		}

		return computed, nil 
	})

	if err != nil { 
		return empty, err
	}
	return res.(T), nil
}