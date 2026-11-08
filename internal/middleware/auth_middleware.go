package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"stockwise-ai/internal/model"
	"stockwise-ai/internal/service"
)

const (
	CtxUserID   = "ctx_user_id"
	CtxUsername = "ctx_username"
	CtxEmail    = "ctx_email"
	CtxRole     = "ctx_role"
)

func JWTAuthMiddleware(authService service.AuthService) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			// Check cookie as fallback for web browser access
			if cookieToken, err := c.Cookie("access_token"); err == nil && cookieToken != "" {
				authHeader = "Bearer " + cookieToken
			}
		}

		if authHeader == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, model.ErrorResponse("Authorization header or cookie is required"))
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, model.ErrorResponse("Invalid authorization header format. Expected 'Bearer <token>'"))
			return
		}

		tokenString := parts[1]
		claims, err := authService.ValidateToken(tokenString)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, model.ErrorResponse("Unauthorized: invalid or expired token"))
			return
		}

		if claims.TokenType != "access" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, model.ErrorResponse("Unauthorized: provided token is not an access token"))
			return
		}

		// Store user identity in Gin context
		c.Set(CtxUserID, claims.UserID)
		c.Set(CtxUsername, claims.Username)
		c.Set(CtxEmail, claims.Email)
		c.Set(CtxRole, claims.Role)

		c.Next()
	}
}
