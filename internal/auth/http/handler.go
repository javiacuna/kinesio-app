package http

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/javiacuna/kinesio-backend/internal/http/middleware"
)

type Handler struct {
	apiKey string
	client *http.Client
}

func NewHandler(apiKey string) *Handler {
	return &Handler{
		apiKey: strings.TrimSpace(apiKey),
		client: &http.Client{Timeout: 10 * time.Second},
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
