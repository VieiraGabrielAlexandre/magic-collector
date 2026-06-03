package auth

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	service *Service
}

func NewHandler(svc *Service) *Handler {
	return &Handler{service: svc}
}

func (h *Handler) Login(c *gin.Context) {
	var input LoginInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Username e password são obrigatórios"})
		return
	}

	sess, user, err := h.service.Login(input.Username, input.Password)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"token":              sess.Token,
		"user":               user,
		"session_created_at": sess.CreatedAt,
	})
}

func (h *Handler) Logout(c *gin.Context) {
	token := extractToken(c)
	if token != "" {
		_ = h.service.Logout(token)
	}
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

func (h *Handler) Me(c *gin.Context) {
	token := extractToken(c)
	if token == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Token não fornecido"})
		return
	}

	sess, user, err := h.service.GetSession(token)
	if err != nil || sess == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Sessão inválida ou expirada"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"user":               user,
		"session_created_at": sess.CreatedAt,
	})
}

// Middleware valida o token Bearer e aborta com 401 se inválido.
func (h *Handler) Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		token := extractToken(c)
		if token == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Autenticação necessária"})
			c.Abort()
			return
		}

		sess, user, err := h.service.GetSession(token)
		if err != nil || sess == nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Sessão inválida ou expirada"})
			c.Abort()
			return
		}

		c.Set("auth_user", user)
		c.Next()
	}
}

func extractToken(c *gin.Context) string {
	h := c.GetHeader("Authorization")
	if strings.HasPrefix(h, "Bearer ") {
		return strings.TrimPrefix(h, "Bearer ")
	}
	return ""
}
