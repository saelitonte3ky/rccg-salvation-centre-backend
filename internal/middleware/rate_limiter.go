// internal/middleware/rate_limiter.go
package middleware

import (
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

type visitor struct {
	lastSeen time.Time
	count    int
}

type rateLimiter struct {
	visitors map[string]*visitor
	mu       sync.RWMutex
	rate     int
	window   time.Duration
}

func newRateLimiter(rate int, window time.Duration) *rateLimiter {
	rl := &rateLimiter{
		visitors: make(map[string]*visitor),
		rate:     rate,
		window:   window,
	}

	go rl.cleanupLoop()

	return rl
}

func (rl *rateLimiter) cleanupLoop() {
	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		now := time.Now()
		rl.mu.Lock()
		for ip, v := range rl.visitors {
			if now.Sub(v.lastSeen) > rl.window {
				delete(rl.visitors, ip)
			}
		}
		rl.mu.Unlock()
	}
}

func (rl *rateLimiter) allow(ip string) bool {
	now := time.Now()

	rl.mu.Lock()
	defer rl.mu.Unlock()

	v, exists := rl.visitors[ip]

	if !exists {
		rl.visitors[ip] = &visitor{lastSeen: now, count: 1}
		return true
	}

	if now.Sub(v.lastSeen) > rl.window {
		v.count = 1
		v.lastSeen = now
		return true
	}

	if v.count >= rl.rate {
		return false
	}

	v.count++
	v.lastSeen = now
	return true
}

func SimpleRateLimiter(rate int, window time.Duration) gin.HandlerFunc {
	limiter := newRateLimiter(rate, window)

	return func(c *gin.Context) {
		ip := c.ClientIP()

		if !limiter.allow(ip) {
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
				"error":               "Too many requests",
				"message":             "Rate limit exceeded. Please try again later.",
				"retry_after_seconds": int(window.Seconds()),
			})
			return
		}

		c.Next()
	}
}

func RateLimiter() gin.HandlerFunc {
	return SimpleRateLimiter(100, time.Minute)
}

func AuthRateLimiter() gin.HandlerFunc {
	return SimpleRateLimiter(10, 15*time.Minute)
}

func StrictRateLimiter() gin.HandlerFunc {
	return SimpleRateLimiter(5, time.Hour)
}

func CustomRateLimiter(rate int, window time.Duration) gin.HandlerFunc {
	return SimpleRateLimiter(rate, window)
}
