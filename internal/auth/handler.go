package auth

import (
	"net/http"
	"strings"

	"github.com/Alugara-Inovasi-Utama/emerti/internal/base"
	"github.com/Alugara-Inovasi-Utama/emerti/utils"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	service Service
}

func NewHandler(service Service) *Handler {
	return &Handler{service: service}
}

// Login handles user login
// @Summary      User login
// @Description  Login dengan username dan password untuk mendapatkan JWT token
// @Tags         Authentication
// @Accept       json
// @Produce      json
// @Param        request body LoginRequest true "Login credentials"
// @Success      200 {object} map[string]interface{} "success: true, data: {token, user}"
// @Failure      400 {object} map[string]string "error: Invalid request"
// @Failure      401 {object} map[string]string "error: invalid credentials"
// @Router       /auth/login [post]
func (h *Handler) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		base.ErrorResponse(c, http.StatusBadRequest, "Invalid request", err)
		return
	}

	ipAddress := utils.GetClientIP(c)
	userAgent := utils.GetUserAgent(c)

	response, err := h.service.Login(req.Username, req.Password, ipAddress, userAgent)
	if err != nil {
		base.ErrorResponse(c, http.StatusUnauthorized, "Login failed", err)
		return
	}

	base.SuccessResponse(c, http.StatusOK, "Login successful", response)
}

// Logout handles user logout
// @Summary      User logout
// @Description  Logout dan invalidate session JWT token
// @Tags         Authentication
// @Produce      json
// @Security     Bearer
// @Success      200 {object} map[string]interface{} "success: true, message: Logout successful"
// @Failure      401 {object} map[string]string "error: Unauthorized"
// @Router       /auth/logout [post]
func (h *Handler) Logout(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		base.ErrorResponse(c, http.StatusUnauthorized, "Unauthorized", nil)
		return
	}

	// Extract token from Authorization header for specific session invalidation
	authHeader := c.GetHeader("Authorization")
	token := strings.TrimPrefix(authHeader, "Bearer ")

	username, _ := c.Get("username")
	ipAddress := utils.GetClientIP(c)
	userAgent := utils.GetUserAgent(c)

	usernameStr, _ := username.(string)

	err := h.service.Logout(userID.(int), token, usernameStr, ipAddress, userAgent)
	if err != nil {
		base.ErrorResponse(c, http.StatusInternalServerError, "Logout failed", err)
		return
	}

	base.SuccessResponse(c, http.StatusOK, "Logout successful", nil)
}
