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

func (c *UserCache) Set(user model.User) {
	c.cache.Store(user.ID, cachedUser{
		user:      user,
		expiresAt: time.Now().Add(c.ttl),
	})
}

func (c *UserCache) Invalidate(id uint) {
	c.cache.Delete(id)
}
