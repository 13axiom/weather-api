// Package middleware provides HTTP middleware for the Weather API.
//
// # Internal API Key
//
// Air-quality endpoints (and any future protected routes) require an internal
// API key sent by the frontend.  This prevents random internet users from
// hammering our backend and eating our OpenWeatherMap quota.
//
// How it works:
//   1. Generate a random secret: `openssl rand -hex 32`
//   2. Set it in weather-api/.env:       INTERNAL_API_KEY=<secret>
//   3. Set it in weather-ui/.env.local:  NEXT_PUBLIC_INTERNAL_KEY=<secret>
//      (Next.js will embed this at build time — it is safe because our
//       backend is the gatekeeper, not the key itself.)
//   4. The frontend sends: X-Internal-Key: <secret> on every AQ request.
//   5. This middleware compares it with the server-side value.
//
// Security note: the INTERNAL_API_KEY is NOT an external-provider key.
// The OpenWeatherMap key (OWM_API_KEY) is NEVER sent to the browser.
package middleware

import (
	"net/http"
)

// RequireInternalKey returns middleware that checks the X-Internal-Key header.
// If the key is missing or wrong, it responds with 401 Unauthorized.
// If INTERNAL_API_KEY is empty (e.g. local dev without .env), the check is skipped.
func RequireInternalKey(expectedKey string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Skip check in dev mode when no key is configured
			if expectedKey == "" {
				next.ServeHTTP(w, r)
				return
			}

			got := r.Header.Get("X-Internal-Key")
			if got != expectedKey {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusUnauthorized)
				w.Write([]byte(`{"error":"invalid or missing X-Internal-Key"}`))
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}
