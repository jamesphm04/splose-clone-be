// Package middleware contains Gin middleware functions for authentication
// authorization, structured request logging (Zap), and panic recovery

package middleware

import (
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jamesphm04/splose-clone-be/internal/utils"
	"github.com/jamesphm04/splose-clone-be/pkg/auth"
	"go.uber.org/zap"
)

// Context keys userd to pass values between middleware and handlers
const (
	ContextKeyUserID = "userID"
	ContextKeyRole   = "role"
)

// RequestLogger logs one structured line per request: method, path, status, latency, client IP and
// the authenticated user ID (when present).
// It uses a named child logger so log lines are easy to filter
func RequestLogger(log *zap.Logger) gin.HandlerFunc {
	reqLog := log.Named("http")

	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		query := c.Request.URL.RawQuery
		method := c.Request.Method

		c.Next() // ← execute the actual handler chain. Then...

		latency := time.Since(start)
		status := c.Writer.Status()
		userID, _ := c.Get(ContextKeyUserID)

		fields := []zap.Field{
			zap.String("method", method),
			zap.String("path", path),
			zap.Int("status", status),
			zap.Duration("latency", latency),
			zap.String("ip", c.ClientIP()),
			zap.String("userAgent", c.Request.UserAgent()),
		}

		if query != "" {
			fields = append(fields, zap.String("query", query))
		}

		if uid, ok := userID.(string); ok && uid != "" {
			fields = append(fields, zap.String("userID", uid))
		}

		if errs := c.Errors.String(); errs != "" {
			fields = append(fields, zap.String("errors", errs))
		}

		switch {
		case status >= 500:
			reqLog.Error("request completed", fields...)
		case status >= 400:
			reqLog.Warn("request completed", fields...)
		default:
			reqLog.Info("request completed", fields...)
		}
	}
}

// Recovery caches panics, emits a structured error log with the panic value
// and stack trace, and returns a clean 500 to the client
func Recovery(log *zap.Logger) gin.HandlerFunc {
	recLog := log.Named("recovery")

	return func(c *gin.Context) {
		defer func() {
			if rec := recover(); rec != nil {
				recLog.Error("panic recovered",
					zap.Any("panic", rec),
					zap.String("method", c.Request.Method),
					zap.String("path", c.Request.URL.Path),
					// zap automatically captures a stack trace at Error level
					// when the logger was built with zap.AddStacktrace(zap.ErrorLevel).
				)
				c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
					"success": false,
					"error":   "internal server error",
				})
			}
		}()

		c.Next()
	}
}

// Authenticate validates the Bearer JWT in the Authorization header
// On success it stores userID and role in the Gin context
func Authenticate(jwtManager *auth.Manager) gin.HandlerFunc {
	return func(c *gin.Context) {
		header := c.GetHeader("Authorization")
		if header == "" {
			utils.Unauthorized(c, "authorization header required")
			c.Abort()
			return
		}

		// Extract access token
		parts := strings.SplitN(header, " ", 2)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "bearer") {
			utils.Unauthorized(c, "invalid authorization header format")
			c.Abort()
			return
		}

		// Parse the token
		claims, err := jwtManager.Parse(parts[1])
		if err != nil {
			switch err {
			case auth.ErrTokenExpired:
				utils.Unauthorized(c, "token has expired")
			default:
				utils.Unauthorized(c, "invalid token")
			}
			c.Abort()
			return
		}

		if claims.TokenType != auth.AccessToken {
			utils.Unauthorized(c, "invalid token type")
			c.Abort()
			return
		}

		// Authenticate
		c.Set(ContextKeyUserID, claims.UserID)
		c.Set(ContextKeyRole, claims.Role)
		c.Next()
	}
}

// RequireRole allows only users whose role is in the permitted list.
// Must be applied after Authenticate.
func RequireRole(roles ...string) gin.HandlerFunc {
	allowed := make(map[string]struct{}, len(roles))
	for _, r := range roles {
		allowed[r] = struct{}{}
	}

	return func(c *gin.Context) {
		role, _ := c.Get(ContextKeyRole)
		if _, ok := allowed[role.(string)]; !ok {
			utils.Forbidden(c)
			c.Abort()
			return
		}
		c.Next()
	}
}

// GetUserID extracts the authenticated user's ID from the Gin context.
func GetUserID(c *gin.Context) string {
	v, _ := c.Get(ContextKeyUserID)
	id, _ := v.(string)
	return id
}

// GetRole extracts the authenticated user's role from the Gin context.
func GetRole(c *gin.Context) string {
	v, _ := c.Get(ContextKeyRole)
	r, _ := v.(string)
	return r
}
