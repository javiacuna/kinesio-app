package middleware

import (
	"net/http"
	"strings"

	"firebase.google.com/go/v4/auth"
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
)

const AuthUserKey = "auth_user"

const DemoReceptionistToken = "Bearer demo-recepcionista-token"

// demoAuthEnabled controla si el bypass DemoReceptionistToken es aceptado.
// Arranca en true para no romper los tests unitarios de los handlers (que
// llaman RequireAuth/RequireRole/HasRole directamente, sin pasar por
// config.MustLoad ni SetDemoAuthEnabled). En un proceso real, NewRouter lo
// fija explícitamente según config.Config.AllowDemoAuth (por defecto false).
var demoAuthEnabled = true

// SetDemoAuthEnabled habilita o deshabilita el bypass de autenticación de
// demo en tiempo de ejecución. Debe llamarse una sola vez al armar el router
// de la aplicación, nunca desde un handler.
func SetDemoAuthEnabled(enabled bool) {
	demoAuthEnabled = enabled
}

type AuthUser struct {
	UID    string
	Email  string
	Role   string
	Claims map[string]any
}

// FirebaseAuthOptional:
// - Si projectID == "" => NO valida (modo local/placeholder).
// - Si projectID != "" => exige Authorization: Bearer <jwt> y deja los claims mínimos en contexto.
func FirebaseAuthOptional(projectID string, authClient *auth.Client) gin.HandlerFunc {
	return func(c *gin.Context) {
		if projectID == "" {
			c.Next()
			return
		}

		idToken, ok := bearerToken(c.GetHeader("Authorization"))
		if !ok {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "missing bearer token"})
			return
		}

		if authClient == nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "firebase_auth_not_configured"})
			return
		}

		token, err := authClient.VerifyIDToken(c.Request.Context(), idToken)
		if err != nil {
			log.Warn().Err(err).Msg("firebase token validation failed")
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid_token"})
			return
		}

		c.Set(AuthUserKey, AuthUser{
			UID:    token.UID,
			Email:  claimString(token.Claims, "email"),
			Role:   extractRole(token.Claims),
			Claims: token.Claims,
		})
		c.Next()
	}
}

func RequireAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		if _, ok := CurrentUser(c); ok {
			c.Next()
			return
		}
		if isDemoReceptionist(c.GetHeader("Authorization")) {
			c.Set(AuthUserKey, AuthUser{UID: "demo", Email: "demo@local", Role: "recepcionista"})
			c.Next()
			return
		}
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
	}
}

func RequireRole(roles ...string) gin.HandlerFunc {
	allowed := map[string]struct{}{}
	for _, role := range roles {
		role = strings.ToLower(strings.TrimSpace(role))
		if role != "" {
			allowed[role] = struct{}{}
		}
	}

	return func(c *gin.Context) {
		user, ok := CurrentUser(c)
		if !ok && isDemoReceptionist(c.GetHeader("Authorization")) {
			user = AuthUser{UID: "demo", Email: "demo@local", Role: "recepcionista"}
			c.Set(AuthUserKey, user)
			ok = true
		}
		if !ok {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			return
		}

		role := strings.ToLower(strings.TrimSpace(user.Role))
		if _, ok := allowed[role]; ok {
			c.Next()
			return
		}
		if role == "admin" {
			c.Next()
			return
		}

		c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "forbidden"})
	}
}

func CurrentUser(c *gin.Context) (AuthUser, bool) {
	raw, ok := c.Get(AuthUserKey)
	if !ok {
		return AuthUser{}, false
	}
	user, ok := raw.(AuthUser)
	return user, ok
}

func HasRole(c *gin.Context, roles ...string) bool {
	user, ok := CurrentUser(c)
	if ok {
		return roleAllowed(user.Role, roles...)
	}
	if isDemoReceptionist(c.GetHeader("Authorization")) {
		return roleAllowed("recepcionista", roles...)
	}
	return false
}

func bearerToken(header string) (string, bool) {
	header = strings.TrimSpace(header)
	if header == "" || !strings.HasPrefix(strings.ToLower(header), "bearer ") {
		return "", false
	}

	token := strings.TrimSpace(header[len("Bearer "):])
	return token, token != ""
}

func isDemoReceptionist(header string) bool {
	if !demoAuthEnabled {
		return false
	}
	return strings.EqualFold(strings.TrimSpace(header), DemoReceptionistToken)
}

func roleAllowed(role string, roles ...string) bool {
	role = strings.ToLower(strings.TrimSpace(role))
	if role == "" {
		return false
	}
	if role == "admin" {
		return true
	}
	for _, allowed := range roles {
		if role == strings.ToLower(strings.TrimSpace(allowed)) {
			return true
		}
	}
	return false
}

func extractRole(claims map[string]any) string {
	if role := claimString(claims, "role"); role != "" {
		return role
	}
	if role := claimString(claims, "roles"); role != "" {
		return role
	}
	return ""
}

func claimString(claims map[string]any, key string) string {
	value, ok := claims[key]
	if !ok {
		return ""
	}
	switch typed := value.(type) {
	case string:
		return strings.TrimSpace(typed)
	case []string:
		if len(typed) > 0 {
			return strings.TrimSpace(typed[0])
		}
	case []any:
		if len(typed) > 0 {
			if s, ok := typed[0].(string); ok {
				return strings.TrimSpace(s)
			}
		}
	}
	return ""
}
