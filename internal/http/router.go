package http

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"github.com/javiacuna/kinesio-backend/internal/config"
	"github.com/javiacuna/kinesio-backend/internal/http/middleware"

	patientsHTTP "github.com/javiacuna/kinesio-backend/internal/patients/http"
	patientsRepo "github.com/javiacuna/kinesio-backend/internal/patients/infra/gorm"
	patientsUC "github.com/javiacuna/kinesio-backend/internal/patients/usecase"
)

type RouterDeps struct {
	Cfg config.Config
	DB  *gorm.DB
}

func NewRouter(cfg config.Config, db *gorm.DB) http.Handler {
	r := gin.New()

	// middlewares globales
	r.Use(gin.Recovery())
	r.Use(middleware.RequestID())
	r.Use(middleware.Logger())

	// health
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})
	r.GET("/version", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"app": cfg.AppName,
			"env": cfg.Env,
		})
	})

	// Patients wiring
	patientRepo := patientsRepo.New(db)
	registerPatientUC := patientsUC.NewRegisterPatientUseCase(patientRepo)
	patientHandler := patientsHTTP.NewHandler(registerPatientUC)

	// API v1
	v1 := r.Group("/api/v1")
	// CU01 - Registrar paciente
	v1.POST("/patients", patientHandler.RegisterPatient)

	// Auth: placeholder para Firebase (por ahora opcional)
	// Cuando setees FIREBASE_PROJECT_ID, este middleware exigirá JWTs (Authorization: Bearer <token>).
	v1.Use(middleware.FirebaseAuthOptional(cfg.FirebaseProjectID))

	// Aquí iremos agregando recursos endpoint por endpoint:
	// v1.POST("/patients", ...)
	// v1.GET("/patients/:id", ...)
	_ = db

	return r
}
