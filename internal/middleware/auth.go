package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"

	"registro-despesas/internal/models"
)

// Chaves do contexto Gin preenchidas por Auth.
const (
	CtxUserID    = "userID"
	CtxUserEmail = "userEmail"
	CtxUserNivel = "userNivel"
)

// Auth valida o JWT e injeta userID, userEmail e userNivel no contexto.
func Auth(secret string) gin.HandlerFunc {
	key := []byte(secret)

	return func(c *gin.Context) {
		raw := extractToken(c)
		if raw == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"erro": "token não informado"})
			return
		}

		claims := &models.JWTClaims{}
		token, err := jwt.ParseWithClaims(raw, claims,
			func(*jwt.Token) (any, error) { return key, nil },
			jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}),
			jwt.WithExpirationRequired(),
		)
		if err != nil || !token.Valid || claims.UserID == 0 ||
			(claims.Nivel != models.NivelAdmin && claims.Nivel != models.NivelRead) {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"erro": "token inválido ou expirado"})
			return
		}

		c.Set(CtxUserID, claims.UserID)
		c.Set(CtxUserEmail, claims.Email)
		c.Set(CtxUserNivel, claims.Nivel)
		c.Next()
	}
}

// RequireAdmin deve vir depois de Auth; barra nível READ com 403 (rotas de escrita e GET /api/usuarios).
func RequireAdmin() gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.GetString(CtxUserNivel) != models.NivelAdmin {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"erro": "acesso restrito a administradores"})
			return
		}
		c.Next()
	}
}

// extractToken segue a ordem: X-Auth-Token, Authorization Bearer, query ?token=.
func extractToken(c *gin.Context) string {
	if t := strings.TrimSpace(c.GetHeader("X-Auth-Token")); t != "" {
		return t
	}

	const prefix = "bearer "
	if h := c.GetHeader("Authorization"); len(h) > len(prefix) && strings.EqualFold(h[:len(prefix)], prefix) {
		if t := strings.TrimSpace(h[len(prefix):]); t != "" {
			return t
		}
	}

	return strings.TrimSpace(c.Query("token"))
}
