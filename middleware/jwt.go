package middleware

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/Alugara-Inovasi-Utama/emerti/config"
	"github.com/Alugara-Inovasi-Utama/emerti/internal/base"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

var jwtSecret = []byte(getEnvOrDefault("JWT_SECRET", "your-secret-key-change-this-in-production"))

func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

type Claims struct {
	UserID      int    `json:"user_id"`
	Username    string `json:"username"`
	Email       string `json:"email"`
	Role        string `json:"role"`
	CompanyID   *uint  `json:"company_id,omitempty"`
	CompanyName string `json:"company_name,omitempty"`
	jwt.RegisteredClaims
}

// GenerateToken membuat JWT token untuk user
func GenerateToken(userID int, username, email, role string, companyID *uint, companyName string) (string, error) {
	claims := Claims{
		UserID:      userID,
		Username:    username,
		Email:       email,
		Role:        role,
		CompanyID:   companyID,
		CompanyName: companyName,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour * 24 * 7)), // 7 days
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(jwtSecret)
}

// sessionTokenKey menghasilkan Redis key unik per token menggunakan hash
func sessionTokenKey(userID int, token string) string {
	hash := sha256.Sum256([]byte(token))
	shortHash := hex.EncodeToString(hash[:8])
	return fmt.Sprintf("%s%d:%s", getEnvOrDefault("REDIS_PREFIX", "emerti:session:"), userID, shortHash)
}

// StoreSession menyimpan session token ke Redis (tidak menghapus token lama)
func StoreSession(ctx context.Context, userID int, token string) error {
	key := sessionTokenKey(userID, token)
	expiration := 24 * 7 * time.Hour // 7 days, same as token expiration
	return config.SetSession(ctx, key, token, expiration)
}

// InvalidateSession menghapus session spesifik dari Redis (untuk logout)
func InvalidateSession(ctx context.Context, userID int, token string) error {
	key := sessionTokenKey(userID, token)
	return config.DeleteSession(ctx, key)
}

// JWTMiddleware verifikasi JWT token
func JWTMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			base.ErrorResponse(c, http.StatusUnauthorized, "Authorization header required", nil)
			c.Abort()
			return
		}

		// Format: Bearer <token>
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			base.ErrorResponse(c, http.StatusUnauthorized, "Invalid authorization format", nil)
			c.Abort()
			return
		}

		tokenString := parts[1]
		claims := &Claims{}

		token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
			return jwtSecret, nil
		})

		if err != nil || !token.Valid {
			base.ErrorResponse(c, http.StatusUnauthorized, "Invalid or expired token", err)
			c.Abort()
			return
		}

		// Validate session exists in Redis
		ctx := context.Background()
		key := sessionTokenKey(claims.UserID, tokenString)
		exists, err := config.SessionExists(ctx, key)

		if err != nil {
			log.Printf("[AUTH] Redis error for user=%d key=%s: %v", claims.UserID, key, err)
			base.ErrorResponse(c, http.StatusServiceUnavailable, "Session service temporarily unavailable", err)
			c.Abort()
			return
		}

		if !exists {
			// Log detail untuk debugging session hilang
			tokenShort := tokenString
			if len(tokenShort) > 20 {
				tokenShort = tokenShort[:20] + "..."
			}
			log.Printf("[AUTH] Session NOT FOUND: user_id=%d key=%s token=%s path=%s",
				claims.UserID, key, tokenShort, c.Request.URL.Path)
			base.ErrorResponse(c, http.StatusUnauthorized, "Session expired or invalid", nil)
			c.Abort()
			return
		}

		// Set user info ke context untuk dipakai di handler
		c.Set("user_id", claims.UserID)
		c.Set("username", claims.Username)
		c.Set("email", claims.Email)
		c.Set("role", claims.Role)
		if claims.CompanyID != nil {
			c.Set("company_id", *claims.CompanyID)
		}
		if claims.CompanyName != "" {
			c.Set("company_name", claims.CompanyName)
		}

		c.Next()
	}
}
