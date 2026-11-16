package reports

import (
	"encoding/json"
	"net/http"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
)

// SessionCheckFn is a function that returns true when a Bearer token is valid.
// Injected at startup to avoid circular imports with the auth package.
type SessionCheckFn func(token string) bool

type Handler struct {
	repo         *Repository
	sessionCheck SessionCheckFn
}

func NewHandler(repo *Repository) *Handler {
	return &Handler{repo: repo}
}

// WithSessionCheck attaches the session validator so the reports middleware
// can accept regular user sessions in addition to the service token.
func (h *Handler) WithSessionCheck(fn SessionCheckFn) *Handler {
	h.sessionCheck = fn
	return h
}

// Middleware accepts either:
//   - The SERVICE_TOKEN env var — static key for M2M integrations (ArcanePulse), OR
//   - A valid user session token from /auth/login.
//
// Register reports routes under this middleware, not under the global api group.
func (h *Handler) Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		raw := c.GetHeader("Authorization")
		token := strings.TrimPrefix(raw, "Bearer ")
		if token == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Autenticação necessária"})
			c.Abort()
			return
		}

		// 1. Static service token (never expires, set via SERVICE_TOKEN env var).
		if svcToken := os.Getenv("SERVICE_TOKEN"); svcToken != "" && token == svcToken {
			c.Next()
			return
		}

		// 2. Regular user session.
		if h.sessionCheck != nil && h.sessionCheck(token) {
			c.Next()
			return
		}

		c.JSON(http.StatusUnauthorized, gin.H{"error": "Service token inválido ou sessão expirada"})
		c.Abort()
	}
}

// Get handles GET /reports/collection
func (h *Handler) Get(c *gin.Context) {
	report, err := h.repo.Build()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, report)
}

// Schema handles GET /reports/collection/schema
func (h *Handler) Schema(c *gin.Context) {
	var raw json.RawMessage
	if err := json.Unmarshal([]byte(JSONSchema), &raw); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "invalid schema"})
		return
	}
	c.JSON(http.StatusOK, raw)
}
