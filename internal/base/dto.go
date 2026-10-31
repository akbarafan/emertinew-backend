package base

import (
	"time"

	"github.com/gin-gonic/gin"
)

var WIB = func() *time.Location {
	loc, err := time.LoadLocation("Asia/Jakarta")
	if err != nil {
		panic("failed to load Asia/Jakarta timezone: " + err.Error())
	}
	return loc
}()

// ResponseStd adalah standard structure untuk semua response
type ResponseStd struct {
	Status     string      `json:"status"`
	StatusCode int         `json:"status_code"`
	Message    string      `json:"message"`
	Pagination *Pagination `json:"pagination,omitempty"`
	Data       interface{} `json:"data,omitempty"`
	Error      string      `json:"error,omitempty"`
}

// Pagination untuk metadata pagination
type Pagination struct {
	CurrentPage int   `json:"current_page"`
	TotalPages  int   `json:"total_pages"`
	TotalItems  int64 `json:"total_items"`
	Limit       int   `json:"limit"`
}

// PaginationRequest untuk query parameter pagination
type PaginationRequest struct {
	Page  int `form:"page"`
	Limit int `form:"limit"`
}

// GetPaginationParams mengambil pagination params dari query
// Return page, limit, offset
// Jika page/limit tidak ada, return 0 (artinya no pagination)
func GetPaginationParams(c *gin.Context) (page, limit, offset int) {
	var req PaginationRequest

	if err := c.ShouldBindQuery(&req); err != nil {
		return 0, 0, 0
	}

	// Kalau page atau limit kosong, return 0 (no pagination)
	if req.Page <= 0 || req.Limit <= 0 {
		return 0, 0, 0
	}

	page = req.Page
	limit = req.Limit
	offset = (page - 1) * limit

	return page, limit, offset
}

// SuccessResponse helper untuk success response
func SuccessResponse(c *gin.Context, statusCode int, message string, data interface{}) {
	c.JSON(statusCode, ResponseStd{
		Status:     "success",
		StatusCode: statusCode,
		Message:    message,
		Data:       data,
	})
}

// ErrorResponse helper untuk error response
func ErrorResponse(c *gin.Context, statusCode int, message string, err error) {
	errMsg := ""
	if err != nil {
		errMsg = err.Error()
	}

	c.JSON(statusCode, ResponseStd{
		Status:     "error",
		StatusCode: statusCode,
		Message:    message,
		Error:      errMsg,
	})
}

// PaginatedResponse helper untuk response dengan pagination
func PaginatedResponse(c *gin.Context, statusCode int, message string, data interface{}, page, limit int, totalItems int64) {
	pagination := CalculatePagination(page, limit, totalItems)

	c.JSON(statusCode, ResponseStd{
		Status:     "success",
		StatusCode: statusCode,
		Message:    message,
		Pagination: &pagination,
		Data:       data,
	})
}

// CalculatePagination menghitung pagination metadata
func CalculatePagination(page, limit int, totalItems int64) Pagination {
	totalPages := int(totalItems) / limit
	if int(totalItems)%limit > 0 {
		totalPages++
	}

	return Pagination{
		CurrentPage: page,
		TotalPages:  totalPages,
		TotalItems:  totalItems,
		Limit:       limit,
	}
}
