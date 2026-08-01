package middleware

import "net/http"

// CORS wraps a handler with permissive cross-origin headers so the frontend
// (served separately, not from this Go binary) can call the API.
// Origin is wide open ("*") — fine while there's a single trusted client,
// worth tightening to a specific origin before this goes properly public.
func CORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}
