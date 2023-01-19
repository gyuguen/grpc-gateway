package rpc

import (
	"context"
	"net/http"
	"time"

	"golang.org/x/time/rate"
)

// It initially issues 3 requests and creates 2 tokens every 1 second.

var limiter = rate.NewLimiter(Per(1, time.Second), 3)

func Per(eventCount int, duration time.Duration) rate.Limit {
	return rate.Every(duration / time.Duration(eventCount))
}

type grpcLimiter struct {
	limiter *rate.Limiter
}

func (s *grpcLimiter) Limit() bool {
	if err := s.limiter.Wait(context.Background()); err != nil {
		return true
	}

	return false
}

func limitREST(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := limiter.Wait(context.Background()); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		next.ServeHTTP(w, r)

	})
}
