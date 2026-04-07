// @title           Weather Dashboard API
// @version         1.1
// @description     Погодный дашборд + качество воздуха. Go backend.
// @description     Weather: Open-Meteo (бесплатно, без ключа).
// @description     Air Quality: OpenWeatherMap Air Pollution API (бесплатно, нужен OWM_API_KEY).
// @host            localhost:8080
// @BasePath        /
// @schemes         http
//
// @securityDefinitions.apikey InternalKey
// @in header
// @name X-Internal-Key
// @description Internal API key shared between weather-ui and weather-api. Set via INTERNAL_API_KEY env var.
package main

import (
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/joho/godotenv"
	httpSwagger "github.com/swaggo/http-swagger"

	_ "github.com/13axiom/weather-api/docs"
	"github.com/13axiom/weather-api/internal/client"
	"github.com/13axiom/weather-api/internal/config"
	"github.com/13axiom/weather-api/internal/database"
	"github.com/13axiom/weather-api/internal/handler"
	mw "github.com/13axiom/weather-api/internal/middleware"
	"github.com/13axiom/weather-api/internal/service"
)

func main() {
	// Load .env (ignored in production — vars come from environment)
	_ = godotenv.Load()

	cfg := config.Load()

	// ── Database ────────────────────────────────────────────────────────────
	db, err := database.New(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}
	defer db.Close()

	if err := db.Migrate(); err != nil {
		log.Fatalf("migration failed: %v", err)
	}

	// ── Clients ─────────────────────────────────────────────────────────────
	meteoClient := client.New() // Open-Meteo weather (no key)

	// ── Weather services ────────────────────────────────────────────────────
	weatherSvc := service.NewWeatherService(db)
	syncSvc    := service.NewSyncService(db, meteoClient, cfg.SyncInterval(), cfg.DefaultCities)
	syncSvc.Start()

	// ── Air Quality services (only if OWM key is configured) ────────────────
	var aqHandler *handler.AirQualityHandler
	if cfg.OWMEnabled() {
		owmClient := client.NewOWMClient(cfg.OWMAPIKey)
		aqSvc := service.NewAirQualityService(
			db, owmClient, meteoClient, cfg.SyncInterval(), cfg.DefaultCities,
		)
		aqSvc.Start()
		aqHandler = handler.NewAirQualityHandler(aqSvc)
		log.Printf("🌫  Air Quality sync enabled (OWM key configured)")
	} else {
		log.Printf("⚠️  OWM_API_KEY not set — air quality endpoints will return 503")
	}

	// ── Handlers ────────────────────────────────────────────────────────────
	weatherH := handler.NewWeatherHandler(weatherSvc)
	syncH    := handler.NewSyncHandler(syncSvc)

	// ── Router ──────────────────────────────────────────────────────────────
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins: cfg.AllowedOrigins,
		AllowedMethods: []string{"GET", "POST", "OPTIONS"},
		AllowedHeaders: []string{"Content-Type", "Accept", "X-Internal-Key"},
	}))

	// Health check (no auth — for uptime monitors)
	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"status":"ok"}`))
	})

	// Swagger UI
	r.Get("/swagger/*", httpSwagger.WrapHandler)

	// ── API v1 ──────────────────────────────────────────────────────────────
	r.Route("/api/v1", func(r chi.Router) {
		// Weather (public — no internal key needed for backward compat)
		r.Get("/cities",          weatherH.GetCities)
		r.Get("/weather/{city}", weatherH.GetWeather)
		r.Post("/sync",          syncH.TriggerSync)

		// Air Quality — protected by internal key
		r.Group(func(r chi.Router) {
			r.Use(mw.RequireInternalKey(cfg.InternalAPIKey))

			if aqHandler != nil {
				r.Get("/air",           aqHandler.GetAllCitiesAQ)
				r.Get("/air/{city}",    aqHandler.GetCityAQ)
				r.Post("/air/sync",     aqHandler.SyncAQ)
			} else {
				// OWM key not configured — return informative 503
				r.HandleFunc("/air", func(w http.ResponseWriter, r *http.Request) {
					w.Header().Set("Content-Type", "application/json")
					w.WriteHeader(http.StatusServiceUnavailable)
					w.Write([]byte(`{"error":"OWM_API_KEY not configured on server"}`))
				})
				r.HandleFunc("/air/*", func(w http.ResponseWriter, r *http.Request) {
					w.Header().Set("Content-Type", "application/json")
					w.WriteHeader(http.StatusServiceUnavailable)
					w.Write([]byte(`{"error":"OWM_API_KEY not configured on server"}`))
				})
			}
		})
	})

	log.Printf("🌤  Weather API started on :%s  (sync every %v)", cfg.Port, cfg.SyncInterval())
	log.Printf("📖  Swagger UI: http://localhost:%s/swagger/index.html", cfg.Port)
	log.Fatal(http.ListenAndServe(":"+cfg.Port, r))
}
