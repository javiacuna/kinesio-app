package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
)

// FirebaseAuthOptional:
// - Si projectID == "" => NO valida (modo local/placeholder).
// - Si projectID != "" => exige Authorization: Bearer <jwt> y deja los claims mínimos en contexto.
// Nota: para producción, conviene validar JWT contra JWKS de Google y verificar aud/iss.
// Acá dejamos el "esqueleto" para que no te frene el armado endpoint por endpoint.
func FirebaseAuthOptional(projectID string) gin.HandlerFunc {
	return func(c *gin.Context) {
		if projectID == "" {
			c.Next()
			return
		}

		auth := c.GetHeader("Authorization")
		if auth == "" || !strings.HasPrefix(auth, "Bearer ") {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "missing bearer token"})
			return
		}

		token := strings.TrimPrefix(auth, "Bearer ")
		if token == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "missing bearer token"})
			return
		}

		// TODO: validar token real con Firebase Admin SDK o JWKS (según enfoque).
		// Por ahora solo dejamos trazabilidad.
		log.Warn().Msg("firebase auth validation is not implemented yet (skeleton only)")
		c.Set("auth_subject", "TODO")
		c.Next()
	}
}
