// Package cache provides in-memory caches used by the readwillbe server.
package cache

import (
	"sync"
	"time"

	"readwillbe/internal/model"
)

type cachedUser struct {
	user      model.User
	expiresAt time.Time
}

// UserCache is a thread-safe TTL cache for [model.User] values keyed by user ID.
type UserCache struct {
	cache sync.Map
	ttl   time.Duration
}

// NewUserCache creates a new user cache with the given TTL and cleanup interval.
// It starts a background goroutine to clean up expired entries.
func NewUserCache(ttl time.Duration, cleanupInterval time.Duration) *UserCache {
	uc := &UserCache{
		ttl: ttl,
	}

	go func() {
		t := time.NewTicker(cleanupInterval)
		defer t.Stop()
		for range t.C {
			uc.cleanup()
		}
	}()

	return uc
}

func (c *UserCache) cleanup() {
	now := time.Now()
	c.cache.Range(func(key, value interface{}) bool {
		cached, ok := value.(cachedUser)
		if !ok || now.After(cached.expiresAt) {
			c.cache.Delete(key)
		}
		return true
	})
}

// Get returns the cached user for id if it exists and has not expired.
func (c *UserCache) Get(id uint) (model.User, bool) {
	val, ok := c.cache.Load(id)
	if !ok {
		return model.User{}, false
	}

	cached, ok := val.(cachedUser)
	if !ok {
		c.cache.Delete(id)
		return model.User{}, false
	}
	if time.Now().After(cached.expiresAt) {
		c.cache.Delete(id)
		return model.User{}, false
	}

	return cached.user, true
}

// Set stores user in the cache with the configured TTL.
func (c *UserCache) Set(user model.User) {
	c.cache.Store(user.ID, cachedUser{
		user:      user,
		expiresAt: time.Now().Add(c.ttl),
	})
}

// Invalidate removes the cache entry for id, if any.
func (c *UserCache) Invalidate(id uint) {
	c.cache.Delete(id)
}
