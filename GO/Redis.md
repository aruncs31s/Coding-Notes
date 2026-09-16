---
id: redis_go
aliases:
  - Redis in Go
  - go-redis v9
tags:
  - go
  - redis
  - caching
  - distributed-systems
dg-publish: true
---

# Redis in Go: Caching, Pub/Sub & Distributed Locks

**Redis** (Remote Dictionary Server) is an in-memory key-value data store used for database caching, pub/sub messaging, session management, and distributed locking. In Go, the standard library has no built-in client; the industry-standard package is **`github.com/redis/go-redis/v9`**.

```
+-------------------------------------------------------------+
|                       Go Application                        |
|                                                             |
|   Cache-Aside      Distributed Lock (SetNX)     Pub / Sub   |
+--------┬──────────────────────┬─────────────────────┬-------+
         │                      │                     │
         ▼                      ▼                     ▼
+-------------------------------------------------------------+
|                            Redis                            |
|  - Key/Value Cache (TTL)    - Mutex Keys      - Channels    |
+-------------------------------------------------------------+
```

---

## 1. Installation & Client Setup

```bash
go get github.com/redis/go-redis/v9
```

```go
package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/redis/go-redis/v9"
)

func NewRedisClient() *redis.Client {
	rdb := redis.NewClient(&redis.Options{
		Addr:         "localhost:6379",
		Password:     "", // no password set
		DB:           0,  // default DB
		PoolSize:     20, // max active connection pool
		MinIdleConns: 5,  // retain minimum idle conns
		DialTimeout:  5 * time.Second,
		ReadTimeout:  3 * time.Second,
		WriteTimeout: 3 * time.Second,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	if err := rdb.Ping(ctx).Err(); err != nil {
		log.Fatalf("Unable to connect to Redis: %v", err)
	}

	fmt.Println("Connected to Redis successfully.")
	return rdb
}
```

---

## 2. Basic Key-Value Operations & Cache-Aside Pattern

```go
import (
	"encoding/json"
	"errors"
)

type UserProfile struct {
	ID    int    `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
}

// GetUserWithCache demonstrates the Cache-Aside pattern
func GetUserWithCache(ctx context.Context, rdb *redis.Client, userID int) (*UserProfile, error) {
	cacheKey := fmt.Sprintf("user:%d", userID)

	// 1. Try reading from cache
	val, err := rdb.Get(ctx, cacheKey).Result()
	if err == nil {
		var user UserProfile
		if jsonErr := json.Unmarshal([]byte(val), &user); jsonErr == nil {
			return &user, nil // Cache Hit!
		}
	} else if !errors.Is(err, redis.Nil) {
		log.Printf("Redis error (fallback to DB): %v", err)
	}

	// 2. Cache Miss: Fetch from DB (simulated)
	user := &UserProfile{
		ID:    userID,
		Name:  "Alice",
		Email: "alice@example.com",
	}

	// 3. Serialize and write back to Redis with a 15-minute TTL
	data, _ := json.Marshal(user)
	_ = rdb.Set(ctx, cacheKey, data, 15*time.Minute).Err()

	return user, nil
}
```

---

## 3. Redis Pub / Sub

Redis Pub/Sub decouples message producers and consumers across multiple Go instances.

```go
// Publisher publishes message to a topic channel
func PublishEvent(ctx context.Context, rdb *redis.Client, channel string, message string) error {
	return rdb.Publish(ctx, channel, message).Err()
}

// SubscribeEvents listens for messages on a channel in a background goroutine
func SubscribeEvents(ctx context.Context, rdb *redis.Client, channel string) {
	pubsub := rdb.Subscribe(ctx, channel)
	defer pubsub.Close()

	ch := pubsub.Channel()

	fmt.Printf("Subscribed to channel '%s'\n", channel)
	for msg := range ch {
		fmt.Printf("Received message from [%s]: %s\n", msg.Channel, msg.Payload)
	}
}
```

---

## 4. Distributed Locking with `SetNX` & Lua Script

When multiple Go services attempt to run the same task (e.g. charging a subscription or generating an hourly report), a **Distributed Lock** ensures mutual exclusion.

### Key Requirements:
1. `SetNX` (Set if Not Exists) with a TTL so that crashes don't cause permanent deadlocks.
2. A unique random token per lock holder.
3. Safe release using a **Lua script** to verify the caller owns the lock before deleting it.

```go
import (
	"crypto/rand"
	"encoding/hex"
)

type DistributedLock struct {
	client  *redis.Client
	key     string
	token   string
	ttl     time.Duration
}

func NewLock(client *redis.Client, key string, ttl time.Duration) *DistributedLock {
	b := make([]byte, 16)
	rand.Read(b)
	token := hex.EncodeToString(b)

	return &DistributedLock{
		client: client,
		key:    "lock:" + key,
		token:  token,
		ttl:    ttl,
	}
}

// Acquire attempts to acquire the lock. Returns true if acquired.
func (l *DistributedLock) Acquire(ctx context.Context) (bool, error) {
	// SET key token NX PX ttl
	acquired, err := l.client.SetNX(ctx, l.key, l.token, l.ttl).Result()
	if err != nil {
		return false, fmt.Errorf("error acquiring lock: %w", err)
	}
	return acquired, nil
}

// Release releases the lock atomically using a Lua script.
func (l *DistributedLock) Release(ctx context.Context) error {
	// Lua script: delete only if current value matches token
	luaScript := `
		if redis.call("get", KEYS[1]) == ARGV[1] then
			return redis.call("del", KEYS[1])
		else
			return 0
		end
	`

	_, err := l.client.Eval(ctx, luaScript, []string{l.key}, l.token).Result()
	return err
}

func ExampleLockUsage(ctx context.Context, rdb *redis.Client) {
	lock := NewLock(rdb, "invoice:monthly_batch", 10*time.Second)

	ok, err := lock.Acquire(ctx)
	if err != nil || !ok {
		fmt.Println("Another instance holds the lock. Skipping execution.")
		return
	}
	defer lock.Release(ctx)

	fmt.Println("Acquired lock. Executing critical section...")
	time.Sleep(2 * time.Second) // Critical work
}
```

---

## 5. Other Useful Redis Structures in Go

- **Hashes (`HSet`, `HGet`, `HGetAll`)**: Ideal for storing objects with multiple fields without JSON serialization overhead.
- **Sets (`SAdd`, `SMembers`, `SIsMember`)**: Unique tags, tracking online users.
- **Sorted Sets (`ZAdd`, `ZRangeByScore`, `ZRevRange`)**: Leaderboards, rate limiters using sliding windows.
- **Bitmaps & HyperLogLog (`PFAdd`, `PFCount`)**: Cardinality estimation for millions of unique page views.
