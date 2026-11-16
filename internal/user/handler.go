package user

import (
	"net/http"
	"strconv"

	"github.com/Alugara-Inovasi-Utama/emerti/internal/base"
	"github.com/Alugara-Inovasi-Utama/emerti/model"
	"github.com/gin-gonic/gin"
)

type Handler struct {
	service Service
}

func NewHandler(service Service) *Handler {
	return &Handler{service: service}
}

func extractUserScope(c *gin.Context) (string, *uint) {
	role := c.GetString("role")
	var compID *uint
	if cidVal, exists := c.Get("company_id"); exists {
		switch v := cidVal.(type) {
		case uint:
			if v > 0 {
				compID = &v
			}
		case int:
			if v > 0 {
				u := uint(v)
				compID = &u
			}
		}
	}
	return role, compID
}

func extractCallerUserID(c *gin.Context) uint {
	if uidVal, exists := c.Get("user_id"); exists {
		switch v := uidVal.(type) {
		case int:
			if v > 0 {
				return uint(v)
			}
		case uint:
			return v
		case float64:
			if v > 0 {
				return uint(v)
			}
		}
	}
	return 0
}

// Create handles user creation (admin or company_admin)
// @Summary      Create new user
// @Description  Admin atau Company Admin dapat membuat user baru
// @Tags         User Management
// @Accept       json
// @Produce      json
// @Security     Bearer
// @Param        request body CreateUserRequest true "User data"
// @Success      201 {object} map[string]interface{} "success: true, data: user"
// @Failure      400 {object} map[string]string "error: Invalid request"
// @Failure      403 {object} map[string]string "error: Access denied"
// @Router       /user [post]
func (h *Handler) Create(c *gin.Context) {
	var req CreateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		base.ErrorResponse(c, http.StatusBadRequest, "Invalid request", err)
		return
	}

	callerRole, callerCompID := extractUserScope(c)

	// Validate role creation authority: cannot create a role with higher rank, and non-superadmin cannot create superadmin
	if !model.CanAssignRole(callerRole, req.Role) {
		base.ErrorResponse(c, http.StatusForbidden, "Cannot create user with role '"+req.Role+"' (insufficient role privileges)", nil)
		return
	}

	// If company_admin: restrict to own company and can only create operator or viewer
	if callerRole == model.RoleCompanyAdmin {
		if callerCompID == nil {
			base.ErrorResponse(c, http.StatusForbidden, "Company admin must belong to a company", nil)
			return
		}
		req.CompanyID = callerCompID

		if req.Role != model.RoleOperator && req.Role != model.RoleViewer && req.Role != model.RoleGuest {
			base.ErrorResponse(c, http.StatusForbidden, "Company admin can only create operator or viewer accounts", nil)
			return
		}
	}

	response, err := h.service.Create(req.Username, req.Password, req.Role, req.CompanyID)
	if err != nil {
		base.ErrorResponse(c, http.StatusBadRequest, "Failed to create user", err)
		return
	}

	base.SuccessResponse(c, http.StatusCreated, "User created successfully", response)
}

// GetAll returns all users (scoped to company if company_admin)
// @Summary      Get all users
// @Description  Mendapatkan list semua users
// @Tags         User Management
// @Produce      json
// @Security     Bearer
// @Success      200 {object} map[string]interface{} "success: true, data: users array"
// @Failure      500 {object} map[string]string "error: Failed to fetch users"
// @Router       /user [get]
func (h *Handler) GetAll(c *gin.Context) {
	callerRole, callerCompID := extractUserScope(c)

	var targetCompID *uint
	if model.IsCompanyScoped(callerRole) {
		targetCompID = callerCompID
	} else if cQuery := c.Query("company_id"); cQuery != "" {
		if cid, err := strconv.ParseUint(cQuery, 10, 32); err == nil && cid > 0 {
			uCid := uint(cid)
			targetCompID = &uCid
		}
	}

	users, err := h.service.GetAll(targetCompID)
	if err != nil {
		base.ErrorResponse(c, http.StatusInternalServerError, "Failed to fetch users", err)
		return
	}

	base.SuccessResponse(c, http.StatusOK, "Users fetched successfully", users)
}

// GetByID returns user by ID
// @Summary      Get user by ID
// @Description  Mendapatkan detail user berdasarkan ID
// @Tags         User Management
// @Produce      json
// @Security     Bearer
// @Param        id path int true "User ID"
// @Success      200 {object} map[string]interface{} "success: true, data: user"
// @Failure      400 {object} map[string]string "error: Invalid ID"
// @Failure      404 {object} map[string]string "error: User not found"
// @Router       /user/{id} [get]
func (h *Handler) GetByID(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		base.ErrorResponse(c, http.StatusBadRequest, "Invalid ID", err)
		return
	}

	callerRole, callerCompID := extractUserScope(c)

	user, err := h.service.GetByID(uint(id))
	if err != nil {
		base.ErrorResponse(c, http.StatusNotFound, "User not found", err)
		return
	}

	if model.IsCompanyScoped(callerRole) {
		if user.CompanyID == nil || callerCompID == nil || *user.CompanyID != *callerCompID {
			base.ErrorResponse(c, http.StatusForbidden, "Access denied to view user from another company", nil)
			return
		}
	}

	base.SuccessResponse(c, http.StatusOK, "User fetched successfully", user)
}

// Update handles user update
// @Summary      Update user
// @Description  Admin atau Company Admin dapat update user
// @Tags         User Management
// @Accept       json
// @Produce      json
// @Security     Bearer
// @Param        id path int true "User ID"
// @Param        request body UpdateUserRequest true "User data"
// @Success      200 {object} map[string]interface{} "success: true, data: user"
// @Failure      400 {object} map[string]string "error: Invalid request"
// @Failure      403 {object} map[string]string "error: Admin access required"
// @Failure      404 {object} map[string]string "error: User not found"
// @Router       /user/{id} [patch]
func (h *Handler) Update(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		base.ErrorResponse(c, http.StatusBadRequest, "Invalid ID", err)
		return
	}

	callerRole, callerCompID := extractUserScope(c)

	existing, err := h.service.GetByID(uint(id))
	if err != nil {
		base.ErrorResponse(c, http.StatusNotFound, "User not found", err)
		return
	}

	// 1. Cannot modify user with higher role (e.g. admin cannot modify superadmin)
	if !model.CanUpdateUser(callerRole, existing.Role) {
		base.ErrorResponse(c, http.StatusForbidden, "Cannot modify user with role '"+existing.Role+"' (insufficient role privileges)", nil)
		return
	}

	// 2. If company_admin: verify ownership
	if model.IsCompanyScoped(callerRole) {
		if existing.CompanyID == nil || callerCompID == nil || *existing.CompanyID != *callerCompID {
			base.ErrorResponse(c, http.StatusForbidden, "Access denied to update user from another company", nil)
			return
		}
	}

	var req UpdateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		base.ErrorResponse(c, http.StatusBadRequest, "Invalid request", err)
		return
	}

	// 3. Cannot escalate role to one higher than caller's authority, or to superadmin if not superadmin
	if req.Role != "" && !model.CanAssignRole(callerRole, req.Role) {
		base.ErrorResponse(c, http.StatusForbidden, "Cannot assign role '"+req.Role+"' (insufficient role privileges)", nil)
		return
	}

	// 4. Company admin cannot escalate role or change company
	if callerRole == model.RoleCompanyAdmin {
		req.CompanyID = callerCompID
		if req.Role != "" && req.Role != model.RoleOperator && req.Role != model.RoleViewer && req.Role != model.RoleGuest {
			base.ErrorResponse(c, http.StatusForbidden, "Company admin can only set role to operator or viewer", nil)
			return
		}
	}

	response, err := h.service.Update(uint(id), req.Username, req.Password, req.Role, req.CompanyID)
	if err != nil {
		base.ErrorResponse(c, http.StatusBadRequest, "Failed to update user", err)
		return
	}

	base.SuccessResponse(c, http.StatusOK, "User updated successfully", response)
}

// Delete handles user deletion
// @Summary      Delete user
// @Description  Admin atau Company Admin dapat delete user
// @Tags         User Management
// @Produce      json
// @Security     Bearer
// @Param        id path int true "User ID"
// @Success      200 {object} map[string]interface{} "success: true"
// @Failure      400 {object} map[string]string "error: Invalid ID"
// @Failure      403 {object} map[string]string "error: Admin access required"
// @Failure      404 {object} map[string]string "error: User not found"
// @Router       /user/{id} [delete]
func (h *Handler) Delete(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		base.ErrorResponse(c, http.StatusBadRequest, "Invalid ID", err)
		return
	}

	callerID := extractCallerUserID(c)
	if callerID > 0 && callerID == uint(id) {
		base.ErrorResponse(c, http.StatusBadRequest, "Cannot delete your own account", nil)
		return
	}

	callerRole, callerCompID := extractUserScope(c)

	existing, err := h.service.GetByID(uint(id))
	if err != nil {
		base.ErrorResponse(c, http.StatusNotFound, "User not found", err)
		return
	}

	// 1. Cannot delete user with higher role (e.g. admin cannot delete superadmin)
	if !model.CanDeleteUser(callerRole, existing.Role) {
		base.ErrorResponse(c, http.StatusForbidden, "Cannot delete user with role '"+existing.Role+"' (insufficient role privileges)", nil)
		return
	}

	// 2. If company_admin: verify ownership
	if model.IsCompanyScoped(callerRole) {
		if existing.CompanyID == nil || callerCompID == nil || *existing.CompanyID != *callerCompID {
			base.ErrorResponse(c, http.StatusForbidden, "Access denied to delete user from another company", nil)
			return
		}
	}

	err = h.service.Delete(uint(id))
	if err != nil {
		base.ErrorResponse(c, http.StatusBadRequest, "Failed to delete user", err)
		return
	}

	base.SuccessResponse(c, http.StatusOK, "User deleted successfully", nil)
}

// GetCurrent returns current logged in user info from JWT token
// @Summary      Get current user
// @Description  Mendapatkan informasi user yang sedang login dari token
// @Tags         User Management
// @Produce      json
// @Security     Bearer
// @Success      200 {object} map[string]interface{} "success: true, data: user"
// @Failure      401 {object} map[string]string "error: Unauthorized"
// @Router       /user/current [get]
func (h *Handler) GetCurrent(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		base.ErrorResponse(c, http.StatusUnauthorized, "Unauthorized", nil)
		return
	}

	user, err := h.service.GetByID(uint(userID.(int)))
	if err != nil {
		base.ErrorResponse(c, http.StatusNotFound, "User not found", err)
		return
	}

	base.SuccessResponse(c, http.StatusOK, "Current user fetched successfully", user)
}
