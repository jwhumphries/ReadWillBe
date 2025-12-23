package main

import (
	"sync"
	"time"

	"readwillbe/types"
)

type cachedUser struct {
	user      types.User
	expiresAt time.Time
}

type UserCache struct {
	cache sync.Map
	ttl   time.Duration
}

func NewUserCache(ttl time.Duration) *UserCache {
	return &UserCache{
		ttl: ttl,
	}
}

func (c *UserCache) Get(id uint) (types.User, bool) {
	val, ok := c.cache.Load(id)
	if !ok {
		return types.User{}, false
	}

	cached := val.(cachedUser)
	if time.Now().After(cached.expiresAt) {
		c.cache.Delete(id)
		return types.User{}, false
	}

	return cached.user, true
}

func (c *UserCache) Set(user types.User) {
	c.cache.Store(user.ID, cachedUser{
		user:      user,
		expiresAt: time.Now().Add(c.ttl),
	})
}

func (c *UserCache) Invalidate(id uint) {
	c.cache.Delete(id)
}
