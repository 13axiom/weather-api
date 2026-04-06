package handler

import (
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/13axiom/weather-api/internal/service"
)

// WeatherHandler handles weather-related HTTP requests.
type WeatherHandler struct {
	svc *service.WeatherService
}

// NewWeatherHandler creates a new WeatherHandler.
func NewWeatherHandler(svc *service.WeatherService) *WeatherHandler {
	return &WeatherHandler{svc: svc}
}

// GetCities godoc
// @Summary     List all tracked cities
// @Description Returns all cities that are being monitored for weather data
// @Tags        weather
// @Produce     json
// @Success     200 {array}  models.City
// @Failure     500 {object} map[string]string
// @Router      /api/v1/cities [get]
func (h *WeatherHandler) GetCities(w http.ResponseWriter, r *http.Request) {
	cities, err := h.svc.GetCities()
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respond(w, http.StatusOK, cities)
}

// GetWeather godoc
// @Summary     Get weather for a city
// @Description Returns the current snapshot and historical data for the given city
// @Tags        weather
// @Produce     json
// @Param       city  path  string  true  "City name (e.g. Moscow)"
// @Param       limit query int     false "Number of history records to return" default(24)
// @Success     200 {object} models.WeatherResponse
// @Failure     404 {object} map[string]string
// @Router      /api/v1/weather/{city} [get]
func (h *WeatherHandler) GetWeather(w http.ResponseWriter, r *http.Request) {
	city := chi.URLParam(r, "city")
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if limit <= 0 {
		limit = 24
	}

	data, err := h.svc.GetWeather(city, limit)
	if err != nil {
		respondError(w, http.StatusNotFound, "city not found or no data yet — try POST /api/v1/sync first")
		return
	}
	respond(w, http.StatusOK, data)
}
