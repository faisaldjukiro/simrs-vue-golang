package monitoring_bed

import (
	"database/sql"
	"net/http"

	"github.com/gin-gonic/gin"

	"simrs-backend/internal/shared/httpresponse"
)

type Handler struct {
	repo *Repositori
}

func NewHandler(db *sql.DB) *Handler {
	return &Handler{repo: NewRepositori(db)}
}

func (h *Handler) Route(r *gin.RouterGroup) {
	r.GET("", h.MonitoringBed)
}

func (h *Handler) MonitoringBed(c *gin.Context) {
	hasil, err := h.repo.Daftar(c.Request.Context(), Filter{
		KataKunci: c.Query("q"),
		Status:    c.Query("status"),
	})
	if err != nil {
		httpresponse.Error(c, http.StatusInternalServerError, "MONITORING_BED_UNAVAILABLE", "Data monitoring bed tidak dapat dibaca")
		return
	}
	httpresponse.Success(c, http.StatusOK, hasil)
}
