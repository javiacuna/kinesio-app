package http

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"time"

	"firebase.google.com/go/v4/auth"
	"github.com/gin-gonic/gin"
	"google.golang.org/api/iterator"

	"github.com/javiacuna/kinesio-backend/internal/http/middleware"
)

type Handler struct {
	apiKey      string
	client      *http.Client
	adminClient *auth.Client
}

func NewHandler(apiKey string, adminClient ...*auth.Client) *Handler {
	var firebaseAdmin *auth.Client
	if len(adminClient) > 0 {
		firebaseAdmin = adminClient[0]
	}

	return &Handler{
		apiKey:      strings.TrimSpace(apiKey),
		client:      &http.Client{Timeout: 10 * time.Second},
		adminClient: firebaseAdmin,
	}
}

type loginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type firebaseLoginRequest struct {
	Email             string `json:"email"`
	Password          string `json:"password"`
	ReturnSecureToken bool   `json:"returnSecureToken"`
}

type firebaseLoginResponse struct {
	IDToken      string `json:"idToken"`
	RefreshToken string `json:"refreshToken"`
	ExpiresIn    string `json:"expiresIn"`
	LocalID      string `json:"localId"`
	Email        string `json:"email"`
}

type loginResponse struct {
	IDToken      string `json:"id_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    string `json:"expires_in"`
	UID          string `json:"uid"`
	Email        string `json:"email"`
}

type assignRoleRequest struct {
	Email string `json:"email"`
	UID   string `json:"uid"`
	Role  string `json:"role"`
}

type adminUserResponse struct {
	UID      string `json:"uid"`
	Email    string `json:"email"`
	Role     string `json:"role"`
	Disabled bool   `json:"disabled"`
}

var validRoles = map[string]struct{}{
	"admin":         {},
	"recepcionista": {},
	"kinesiologo":   {},
	"paciente":      {},
}

func (h *Handler) Login(c *gin.Context) {
	if h.apiKey == "" {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "firebase_api_key_not_configured"})
		return
	}

	var req loginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_json"})
		return
	}

	email := strings.TrimSpace(req.Email)
	password := strings.TrimSpace(req.Password)
	if email == "" || password == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "validation_error",
			"details": map[string]string{
				"email":    requiredIfEmpty(email),
				"password": requiredIfEmpty(password),
			},
		})
		return
	}

	payload, err := json.Marshal(firebaseLoginRequest{
		Email:             email,
		Password:          password,
		ReturnSecureToken: true,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal_error"})
		return
	}

	endpoint := "https://identitytoolkit.googleapis.com/v1/accounts:signInWithPassword?key=" + h.apiKey
	httpReq, err := http.NewRequestWithContext(c.Request.Context(), http.MethodPost, endpoint, bytes.NewReader(payload))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal_error"})
		return
	}
	httpReq.Header.Set("Content-Type", "application/json")

	res, err := h.client.Do(httpReq)
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": "firebase_unavailable"})
		return
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": "firebase_unavailable"})
		return
	}

	if res.StatusCode == http.StatusBadRequest || res.StatusCode == http.StatusUnauthorized {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid_credentials"})
		return
	}
	if res.StatusCode < 200 || res.StatusCode >= 300 {
		c.JSON(http.StatusBadGateway, gin.H{"error": "firebase_error"})
		return
	}

	var fbResp firebaseLoginResponse
	if err := json.Unmarshal(body, &fbResp); err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": "firebase_error"})
		return
	}

	c.JSON(http.StatusOK, loginResponse{
		IDToken:      fbResp.IDToken,
		RefreshToken: fbResp.RefreshToken,
		ExpiresIn:    fbResp.ExpiresIn,
		UID:          fbResp.LocalID,
		Email:        fbResp.Email,
	})
}

func (h *Handler) AssignRole(c *gin.Context) {
	if !middleware.HasRole(c, "admin") {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	if h.adminClient == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "firebase_auth_not_configured"})
		return
	}

	var req assignRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_json"})
		return
	}

	role := strings.ToLower(strings.TrimSpace(req.Role))
	if _, ok := validRoles[role]; !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_role"})
		return
	}

	uid := strings.TrimSpace(req.UID)
	email := strings.TrimSpace(req.Email)
	var user *auth.UserRecord
	var err error

	if uid != "" {
		user, err = h.adminClient.GetUser(c.Request.Context(), uid)
	} else if email != "" {
		user, err = h.adminClient.GetUserByEmail(c.Request.Context(), email)
	} else {
		c.JSON(http.StatusBadRequest, gin.H{"error": "email_or_uid_required"})
		return
	}
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "firebase_user_not_found"})
		return
	}

	claims := map[string]any{}
	for key, value := range user.CustomClaims {
		claims[key] = value
	}
	claims["role"] = role

	if err := h.adminClient.SetCustomUserClaims(c.Request.Context(), user.UID, claims); err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": "firebase_error"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"uid":   user.UID,
		"email": user.Email,
		"role":  role,
	})
}

func (h *Handler) ListUsers(c *gin.Context) {
	if !middleware.HasRole(c, "admin") {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	if h.adminClient == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "firebase_auth_not_configured"})
		return
	}

	iter := h.adminClient.Users(c.Request.Context(), "")
	out := []adminUserResponse{}
	for {
		user, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			c.JSON(http.StatusBadGateway, gin.H{"error": "firebase_error"})
			return
		}

		out = append(out, adminUserResponse{
			UID:      user.UID,
			Email:    user.Email,
			Role:     extractRole(user.CustomClaims),
			Disabled: user.Disabled,
		})
	}

	c.JSON(http.StatusOK, out)
}

func (h *Handler) Me(c *gin.Context) {
	user, ok := middleware.CurrentUser(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"uid":   user.UID,
		"email": user.Email,
		"role":  user.Role,
	})
}

func requiredIfEmpty(value string) string {
	if value == "" {
		return "required"
	}
	return ""
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
