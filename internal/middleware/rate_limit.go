package middleware

import (
	"net/http"
	"sync"
	"golang.org/x/time/rate" // You'll need to go get this
)

// IPRateLimiter manages a map of limiters based on IP addresses.
type IPRateLimiter struct {
	ips map[string]*rate.Limiter
	mu  sync.Mutex
	r   rate.Limit // Rate: how many tokens per second
	b   int        // Burst: max tokens in the bucket
}

func NewIPRateLimiter(r rate.Limit, b int) *IPRateLimiter {
	return &IPRateLimiter{
		ips: make(map[string]*rate.Limiter),
		r:   r,
		b:   b,
	}
}

// GetLimiter returns the rate limiter for the provided IP address.
func (i *IPRateLimiter) GetLimiter(ip string) *rate.Limiter {
	i.mu.Lock()
	defer i.mu.Unlock()

	limiter, exists := i.ips[ip]
	if !exists {
		// If this IP has never visited, create a new bucket for them
		limiter = rate.NewLimiter(i.r, i.b)
		i.ips[ip] = limiter
	}

	return limiter
}

// RateLimitMiddleware is the actual function that wraps our routes
func (i *IPRateLimiter) RateLimitMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// 1. Get the IP address
		ip := r.RemoteAddr

		// 2. Get the specific bucket for this IP
		limiter := i.GetLimiter(ip)

		// 3. Try to take a token
		if !limiter.Allow() {
			// No tokens left! Return 429
			http.Error(w, "Too Many Requests - Slow down, Faisal!", http.StatusTooManyRequests)
			return
		}

		// 4. Token granted, proceed to the next handler
		next.ServeHTTP(w, r)
	})
}