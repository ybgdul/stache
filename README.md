# stache

A small local cache for Go backends. Runs as a daemon next to your app, keeps things in RAM, and talks over a Unix socket. No Redis, no network hops, no extra infra to babysit.

## How it fits together

```
Go App (stache-go) ---> Unix Socket ---> stache-daemon (RAM Cache)
CLI (stache)       ---> Unix Socket ---> stache-daemon
```

- **`stache-daemon`** — the background process. Holds items in RAM, evicts with LRU, expires keys on TTL.
- **`stache-go`** — the client library. Handles fallbacks and single-flight coalescing. No hard dependencies.
- **`stache`** — CLI for admin tasks.

## Getting started

### 1. Run the daemon

```bash
go run cmd/stached/main.go --max-memory=256MB --max-item-size=4MB --socket=/tmp/stache.sock
```

All flags are optional. Defaults are fine for local dev.

### 2. CLI

The CLI talks to the daemon over the same socket.

```bash
stache status    # check the daemon is up
stache stats     # memory, item counts, hit/miss
stache clear     # drop everything
stache stop      # shut it down gracefully
```

### 3. Set up the client

Create one `stache.Client` at startup and pass it around.

```go
import (
	"time"
	"stache/pkg/stache"
)

client := stache.NewClient(stache.Config{
	SocketPath: "/tmp/stache.sock",
	Namespace:  "user-service",
	Timeout:    500 * time.Millisecond,
})
```

| Field | Type | Notes |
| --- | --- | --- |
| `SocketPath` | `string` | Path to the daemon socket. Defaults to `/tmp/stache.sock`. |
| `Namespace` | `string` | Optional prefix added to every key, e.g. `user-service:key`. Keeps services from stepping on each other. |
| `Timeout` | `time.Duration` | Dial + read timeout. On timeout, calls fall back. |

### 4. Usage

#### `stache.Remember`

Wrap a slow call. On a hit, the function never runs. On a miss, or if the daemon is down, it runs normally.

```go
user, err := stache.Remember(
	ctx,
	client,
	stache.Key("user", 123), // "user:123"
	5*time.Minute,
	func(ctx context.Context) (User, error) {
		return db.GetUser(ctx, 123)
	},
)
if err != nil {
	log.Fatalf("Failed to retrieve user: %v", err)
}
```

#### `client.Delete`

Invalidate a key after you change the source of truth.

```go
func (s *UserService) UpdateUser(ctx context.Context, user User) error {
	if err := s.db.UpdateUser(ctx, user); err != nil {
		return err
	}

	// true if evicted or absent, false if the daemon was unreachable
	s.client.Delete(stache.Key("user", user.ID))

	return nil
}
```

