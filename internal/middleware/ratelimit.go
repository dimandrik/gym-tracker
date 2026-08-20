package middleware

import (
	"net/http"
	"sync"
	"time"
)

type rateLimiter struct {
	mu       sync.Mutex
	attempts map[string][]time.Time
	limit    int
	window   time.Duration
}

// лимитер в памяти процесса: при нескольких инстансах сервера за балансировщиком
// каждый инстанс считает свой лимит отдельно, единого лимита на пользователя пока нет
func NewRateLimiter(limit int, window time.Duration) *rateLimiter {
	rl := &rateLimiter{
		attempts: make(map[string][]time.Time),
		limit:    limit,
		window:   window,
	}
	rl.startCleanup()
	return rl
}

func (rl *rateLimiter) Allow(key string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	cutoff := now.Add(-rl.window)

	var recent []time.Time
	for _, t := range rl.attempts[key] {
		if t.After(cutoff) {
			recent = append(recent, t)
		}
	}

	if len(recent) >= rl.limit {
		rl.attempts[key] = recent
		return false
	}

	recent = append(recent, now)
	rl.attempts[key] = recent
	return true
}

func RateLimitMiddleware(rl *rateLimiter) func(http.HandlerFunc) http.HandlerFunc {
	return func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			// RemoteAddr — это адрес прямого соединения; если сервис стоит за reverse proxy,
			// сюда попадёт адрес прокси, и лимит станет общим на всех, а не per-client
			ip := r.RemoteAddr

			if !rl.Allow(ip) {
				http.Error(w, "too many requests, try again later", http.StatusTooManyRequests)
				return
			}

			next(w, r)
		}
	}
}

// без периодической очистки attempts бесконечно растёт за счёт ключей (IP),
// которые больше никогда не появятся снова — особенно при переборе/сканировании
func (rl *rateLimiter) startCleanup() {
	ticker := time.NewTicker(rl.window)
	go func() {
		for range ticker.C {
			rl.cleanup()
		}
	}()
}

// удаляет ключи, у которых не осталось попыток внутри текущего окна
func (rl *rateLimiter) cleanup() {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	cutoff := time.Now().Add(-rl.window)

	for key, times := range rl.attempts {
		var recent []time.Time
		for _, t := range times {
			if t.After(cutoff) {
				recent = append(recent, t)
			}
		}

		if len(recent) == 0 {
			delete(rl.attempts, key)
		} else {
			rl.attempts[key] = recent
		}
	}
}
