package handler

import (
	"net/http"

	"github.com/13axiom/weather-api/internal/service"
)

// SyncHandler handles manual sync trigger requests.
type SyncHandler struct {
	svc *service.SyncService
}

// NewSyncHandler creates a new SyncHandler.
func NewSyncHandler(svc *service.SyncService) *SyncHandler {
	return &SyncHandler{svc: svc}
}

// TriggerSync godoc
// @Summary     Manually trigger data sync
// @Description Immediately syncs weather data for all configured cities
// @Tags        sync
// @Produce     json
// @Success     200 {array}  models.SyncResult
// @Router      /api/v1/sync [post]
func (h *SyncHandler) TriggerSync(w http.ResponseWriter, r *http.Request) {
	results := h.svc.SyncAll()
	respond(w, http.StatusOK, results)
}
