package config

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"
)

type Config struct {
	AppName  string
	Env      string
	HTTPPort string

	DBHost     string
	DBPort     string
	DBName     string
	DBUser     string
	DBPassword string
	DBSSLMode  string

	FirebaseProjectID       string
	FirebaseWebAPIKey       string
	FirebaseCredentialsFile string

	// PublicBaseURL es la URL publica de esta app (ej. https://kinesio-app.onrender.com).
	// Si esta seteada, el link de "olvide mi contrasena" lleva a nuestra propia pantalla
	// de reseteo en vez de a la generica de Firebase. Requiere que el dominio este en
	// Firebase Console > Authentication > Settings > Authorized domains.
	PublicBaseURL string

	PatientFilesDir       string
	KinesiologistFilesDir string

	NotificationEmailEnabled bool
	SMTPHost                 string
	SMTPPort                 string
	SMTPUsername             string
	SMTPPassword             string
	SMTPFromEmail            string
	SMTPFromName             string

	SupportEmail    string
	SupportPhone    string
	SupportFilesDir string

	VideoCallProvider  string
	DailyAPIKey        string
	WherebyAPIKey      string
	WherebyRoomPrefix  string
	VideoCallExpiryMin int

	// StaticDir, si se setea, hace que el backend sirva el frontend ya compilado
	// (frontend/dist) desde ese directorio, ademas de la API. Vacio = desactivado
	// (modo desarrollo normal, con el frontend corriendo aparte via Vite).
	StaticDir string

	// AllowDemoAuth habilita el bypass "Bearer demo-recepcionista-token" usado para
	// probar endpoints de recepcionista sin credenciales reales de Firebase (util en
	// local/demo). Por defecto esta APAGADO: hay que setear ALLOW_DEMO_AUTH=true a
	// proposito para habilitarlo, incluso en ENV=local.
	AllowDemoAuth bool
}

func MustLoad() Config {
	cfg := Config{
		AppName:                  getenv("APP_NAME", "kinesio-app"),
		Env:                      getenv("ENV", "local"),
		HTTPPort:                 getenv("HTTP_PORT", "8080"),
		DBHost:                   getenv("DB_HOST", "localhost"),
		DBPort:                   getenv("DB_PORT", "5432"),
		DBName:                   getenv("DB_NAME", "kinesio"),
		DBUser:                   getenv("DB_USER", "kinesio"),
		DBPassword:               getenv("DB_PASSWORD", "kinesio"),
		DBSSLMode:                getenv("DB_SSLMODE", "disable"),
		FirebaseProjectID:        getenv("FIREBASE_PROJECT_ID", ""),
		FirebaseWebAPIKey:        getenv("FIREBASE_WEB_API_KEY", ""),
		FirebaseCredentialsFile:  getenv("GOOGLE_APPLICATION_CREDENTIALS", ""),
		PublicBaseURL:            strings.TrimRight(getenv("PUBLIC_BASE_URL", ""), "/"),
		PatientFilesDir:          getenv("PATIENT_FILES_DIR", "uploads/patient-attachments"),
		KinesiologistFilesDir:    getenv("KINESIOLOGIST_FILES_DIR", "uploads/kinesiologist-attachments"),
		NotificationEmailEnabled: getenvBool("NOTIFICATIONS_EMAIL_ENABLED", false),
		SMTPHost:                 getenv("SMTP_HOST", ""),
		SMTPPort:                 getenv("SMTP_PORT", "587"),
		SMTPUsername:             getenv("SMTP_USERNAME", ""),
		SMTPPassword:             getenv("SMTP_PASSWORD", ""),
		SMTPFromEmail:            getenv("SMTP_FROM_EMAIL", ""),
		SMTPFromName:             getenv("SMTP_FROM_NAME", "Kinesio App"),
		SupportEmail:             getenv("SUPPORT_EMAIL", ""),
		SupportPhone:             getenv("SUPPORT_PHONE", ""),
		SupportFilesDir:          getenv("SUPPORT_FILES_DIR", "uploads/support-attachments"),
		VideoCallProvider:        getenv("VIDEO_CALL_PROVIDER", ""),
		DailyAPIKey:              getenv("DAILY_API_KEY", ""),
		WherebyAPIKey:            getenv("WHEREBY_API_KEY", ""),
		WherebyRoomPrefix:        getenv("WHEREBY_ROOM_PREFIX", "kinesio"),
		VideoCallExpiryMin:       getenvInt("VIDEO_CALL_EXPIRY_MIN", 60),
		StaticDir:                getenv("STATIC_DIR", ""),
		AllowDemoAuth:            getenvBool("ALLOW_DEMO_AUTH", false),
	}

	// Validaciones mínimas
	if cfg.HTTPPort == "" {
		panic("HTTP_PORT is required")
	}
	if cfg.DBHost == "" || cfg.DBPort == "" || cfg.DBName == "" || cfg.DBUser == "" {
		panic("DB_* configuration is required")
	}
	return cfg
}

func getenvInt(k string, def int) int {
	value := os.Getenv(k)
	if value == "" {
		return def
	}
	var out int
	if _, err := fmt.Sscanf(value, "%d", &out); err != nil || out <= 0 {
		return def
	}
	return out
}

func (c Config) PostgresDSN() string {
	// Ej: host=localhost user=foo password=bar dbname=baz port=5432 sslmode=disable TimeZone=UTC
	return fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=%s TimeZone=UTC",
		c.DBHost, c.DBUser, c.DBPassword, c.DBName, c.DBPort, c.DBSSLMode,
	)
}

func ShutdownContext() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), 10*time.Second)
}

func getenv(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}

func getenvBool(k string, def bool) bool {
	value := os.Getenv(k)
	if value == "" {
		return def
	}
	switch value {
	case "1", "true", "TRUE", "True", "yes", "YES", "Yes", "on", "ON", "On":
		return true
	default:
		return false
	}
}
