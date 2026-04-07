package handler

import (
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/13axiom/weather-api/internal/models"
	"github.com/13axiom/weather-api/internal/service"
)

// AirQualityHandler handles air-quality HTTP requests.
type AirQualityHandler struct {
	svc *service.AirQualityService
}

// NewAirQualityHandler creates a new AirQualityHandler.
func NewAirQualityHandler(svc *service.AirQualityService) *AirQualityHandler {
	return &AirQualityHandler{svc: svc}
}

// GetAllCitiesAQ godoc
// @Summary     Latest air quality for all cities
// @Description Returns the most recent AQ snapshot for every tracked city
// @Tags        air-quality
// @Produce     json
// @Security    InternalKey
// @Success     200 {array}  models.AirQualitySnapshot
// @Failure     401 {object} map[string]string
// @Failure     500 {object} map[string]string
// @Router      /api/v1/air [get]
func (h *AirQualityHandler) GetAllCitiesAQ(w http.ResponseWriter, r *http.Request) {
	snapshots, err := h.svc.GetAllCitiesAirQuality()
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if snapshots == nil {
		snapshots = []models.AirQualitySnapshot{}
	}
	respond(w, http.StatusOK, snapshots)
}

// GetCityAQ godoc
// @Summary     Air quality for a specific city
// @Description Returns the current AQ snapshot and recent history for the given city
// @Tags        air-quality
// @Produce     json
// @Security    InternalKey
// @Param       city  path  string  true  "City name (e.g. Moscow)"
// @Param       limit query int     false "Number of history records" default(24)
// @Success     200 {object} models.AirQualityResponse
// @Failure     401 {object} map[string]string
// @Failure     404 {object} map[string]string
// @Router      /api/v1/air/{city} [get]
func (h *AirQualityHandler) GetCityAQ(w http.ResponseWriter, r *http.Request) {
	city := chi.URLParam(r, "city")
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if limit <= 0 {
		limit = 24
	}

	data, err := h.svc.GetAirQuality(city, limit)
	if err != nil {
		respondError(w, http.StatusNotFound, "city not found or no AQ data yet — try POST /api/v1/air/sync")
		return
	}
	respond(w, http.StatusOK, data)
}

// SyncAQ godoc
// @Summary     Trigger air quality sync
// @Description Immediately fetches fresh AQ data from OpenWeatherMap for all cities
// @Tags        air-quality
// @Produce     json
// @Security    InternalKey
// @Success     200 {array}  models.AirQualitySyncResult
// @Failure     401 {object} map[string]string
// @Router      /api/v1/air/sync [post]
func (h *AirQualityHandler) SyncAQ(w http.ResponseWriter, r *http.Request) {
	results := h.svc.SyncAll()
	respond(w, http.StatusOK, results)
}
