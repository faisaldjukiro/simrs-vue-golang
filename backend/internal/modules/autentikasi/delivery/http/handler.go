package autentikasihttp

import (
	"errors"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"simrs-backend/internal/modules/autentikasi"
	"simrs-backend/internal/shared/httpresponse"
)

type Handler struct {
	layanan *autentikasi.Layanan
}

func NewHandler(layanan *autentikasi.Layanan) *Handler {
	return &Handler{layanan: layanan}
}

func (h *Handler) Register(router *gin.RouterGroup) {
	router.POST("/login", h.login)

	protected := router.Group("")
	protected.Use(h.Middleware())
	protected.GET("/me", h.me)
	protected.POST("/logout", h.logout)
}

type loginRequest struct {
	Username string `json:"username" binding:"required,max=100"`
	Password string `json:"password" binding:"required,max=255"`
}

func (h *Handler) login(c *gin.Context) {
	var request loginRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		httpresponse.Error(c, http.StatusUnprocessableEntity, "VALIDATION_ERROR", "Username dan password wajib diisi")
		return
	}
	c.Set("audit_username", strings.TrimSpace(request.Username))

	result, err := h.layanan.Login(c.Request.Context(), request.Username, request.Password)
	if errors.Is(err, autentikasi.ErrKredensialTidakValid) {
		httpresponse.Error(c, http.StatusUnauthorized, "INVALID_CREDENTIALS", "Username atau password salah")
		return
	}
	if err != nil {
		httpresponse.Error(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Terjadi kesalahan pada server")
		return
	}
	c.Set(autentikasi.ContextKeyPengguna, result.User)
	httpresponse.Success(c, http.StatusOK, result)
}

func (h *Handler) me(c *gin.Context) {
	user, exists := c.Get(autentikasi.ContextKeyPengguna)
	if !exists {
		httpresponse.Error(c, http.StatusUnauthorized, "UNAUTHENTICATED", "Token tidak valid")
		return
	}
	httpresponse.Success(c, http.StatusOK, user)
}

func (h *Handler) logout(c *gin.Context) {
	token := c.GetString(autentikasi.ContextKeyToken)
	if err := h.layanan.Logout(c.Request.Context(), token); err != nil {
		httpresponse.Error(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Terjadi kesalahan pada server")
		return
	}
	httpresponse.Success(c, http.StatusOK, gin.H{"message": "Logout berhasil"})
}

func (h *Handler) Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		header := strings.TrimSpace(c.GetHeader("Authorization"))
		parts := strings.Fields(header)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
			httpresponse.Error(c, http.StatusUnauthorized, "UNAUTHENTICATED", "Bearer token wajib diisi")
			return
		}

		user, err := h.layanan.Authenticate(c.Request.Context(), parts[1])
		if errors.Is(err, autentikasi.ErrTokenTidakValid) {
			httpresponse.Error(c, http.StatusUnauthorized, "UNAUTHENTICATED", "Token tidak valid atau kedaluwarsa")
			return
		}
		if err != nil {
			httpresponse.Error(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Terjadi kesalahan pada server")
			return
		}

		c.Set(autentikasi.ContextKeyPengguna, user)
		c.Set(autentikasi.ContextKeyToken, parts[1])
		c.Next()
	}
}
