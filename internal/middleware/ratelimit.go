package middleware

import (
	"math"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

type bucket struct {
	tokens float64
	last   time.Time
}

// RateLimiter é um token bucket em memória por IP.
type RateLimiter struct {
	mu        sync.Mutex
	buckets   map[string]*bucket
	capacity  float64
	refill    float64 // tokens por segundo
	lastSweep time.Time
}

func NewRateLimiter(perMinute int) *RateLimiter {
	return &RateLimiter{
		buckets:   make(map[string]*bucket),
		capacity:  float64(perMinute),
		refill:    float64(perMinute) / 60.0,
		lastSweep: time.Now(),
	}
}

func (rl *RateLimiter) allow(key string) (bool, time.Duration) {
	now := time.Now()

	rl.mu.Lock()
	defer rl.mu.Unlock()

	// Um bucket parado por 1 min já está cheio, então pode ser descartado.
	if now.Sub(rl.lastSweep) > time.Minute {
		for k, b := range rl.buckets {
			if now.Sub(b.last) > time.Minute {
				delete(rl.buckets, k)
			}
		}
		rl.lastSweep = now
	}

	b, ok := rl.buckets[key]
	if !ok {
		b = &bucket{tokens: rl.capacity, last: now}
		rl.buckets[key] = b
	} else {
		b.tokens = math.Min(rl.capacity, b.tokens+now.Sub(b.last).Seconds()*rl.refill)
		b.last = now
	}

	if b.tokens >= 1 {
		b.tokens--
		return true, 0
	}
	return false, time.Duration((1 - b.tokens) / rl.refill * float64(time.Second))
}

func (rl *RateLimiter) Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		ok, wait := rl.allow(c.ClientIP())
		if !ok {
			c.Header("Retry-After", strconv.Itoa(int(math.Ceil(wait.Seconds()))))
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{"erro": "muitas requisições, tente novamente em instantes"})
			return
		}
		c.Next()
	}
}
