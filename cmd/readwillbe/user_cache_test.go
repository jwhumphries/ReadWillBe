package main

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
	"readwillbe/types"
)

func TestUserCache_Cleanup(t *testing.T) {
	ttl := 100 * time.Millisecond
	cache := NewUserCache(ttl, 50*time.Millisecond)

	user := types.User{
		Model: gorm.Model{ID: 1},
		Email: "test@example.com",
	}

	cache.Set(user)

	// Immediately check it's there
	cached, found := cache.Get(user.ID)
	assert.True(t, found)
	assert.Equal(t, user.Email, cached.Email)

	// Wait for TTL + cleanup
	time.Sleep(200 * time.Millisecond)

	// Verify cleanup via cache inspection
	count := 0
	cache.cache.Range(func(key, value interface{}) bool {
		count++
		return true
	})
	assert.Equal(t, 0, count, "Cache should be empty after cleanup")
}

func TestUserCache_Get_Expired(t *testing.T) {
	ttl := 50 * time.Millisecond
	// Longer cleanup so it doesn't interfere with this specific test logic
	// although Get() checks expiry anyway.
	cache := NewUserCache(ttl, 1*time.Second)

	user := types.User{Model: gorm.Model{ID: 1}}
	cache.Set(user)

	time.Sleep(100 * time.Millisecond)

	_, found := cache.Get(1)
	assert.False(t, found, "Should return false for expired user")
}

func TestUserCache_Invalidate(t *testing.T) {
	cache := NewUserCache(1*time.Hour, 1*time.Hour)
	user := types.User{Model: gorm.Model{ID: 1}}
	cache.Set(user)

	_, found := cache.Get(1)
	assert.True(t, found)

	cache.Invalidate(1)

	_, found = cache.Get(1)
	assert.False(t, found)
}
