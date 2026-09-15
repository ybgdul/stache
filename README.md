# stache

A lightweight local caching system for Go backends.

Stache runs as a tiny daemon alongside your Go application, offering fast in-memory caching without the operational overhead of running a separate Redis instance.

## Architecture

Go App (stache-go) ---> Unix Socket ---> stache-daemon (RAM Cache) ---> CLI (stache)
CLI (stache) ---> Unix Socket ---> stache-daemon

## Quick Start

### 1. Start the Daemon

```bash
go run cmd/stached/main.go --max-memory=256MB
```

### 2. Use CLI Commands

```bash 
stache status    # Verify daemon connectivity
stache stats     # View memory usage, item counts, and hit/miss metrics
stache clear     # Flush all keys from memory
stache stop      # Gracefully shut down the daemon
``` 

### 3. Use Libary Methods 

```bash 
user, err := stache.Remember(
		context.Background(),
		client,
		stache.Key("user", 123),
		5*time.Minute,
		func(ctx context.Context) (User, error) {
			// Expensive database fetch runs only on cache miss or daemon outage
			return fetchUserFromDB(123)
		},
	)

    if err != nil { panic(err) }
```