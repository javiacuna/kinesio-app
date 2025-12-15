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

	appointmentsHTTP "github.com/javiacuna/kinesio-backend/internal/appointments/http"
	appointmentsRepo "github.com/javiacuna/kinesio-backend/internal/appointments/infra/gorm"
	appointmentsUC "github.com/javiacuna/kinesio-backend/internal/appointments/usecase"
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
	getPatientByIDUC := patientsUC.NewGetPatientByIDUseCase(patientRepo)
	patientHandler := patientsHTTP.NewHandler(registerPatientUC, getPatientByIDUC)

	apptRepo := appointmentsRepo.New(db)
	createApptUC := appointmentsUC.NewCreateAppointmentUseCase(apptRepo)
	listDayUC := appointmentsUC.NewListAppointmentsDayUseCase(apptRepo)
	updateApptUC := appointmentsUC.NewUpdateAppointmentUseCase(apptRepo)

	getApptByIDUC := appointmentsUC.NewGetAppointmentByIDUseCase(apptRepo)
	listByPatientUC := appointmentsUC.NewListAppointmentsByPatientUseCase(apptRepo)

	apptHandler := appointmentsHTTP.NewHandler(
		createApptUC,
		listDayUC,
		updateApptUC,
		getApptByIDUC,
		listByPatientUC,
	)

	// API v1
	v1 := r.Group("/api/v1")

	// Auth: placeholder para Firebase (por ahora opcional)
	// Cuando setees FIREBASE_PROJECT_ID, este middleware exigirá JWTs (Authorization: Bearer <token>).
	v1.Use(middleware.FirebaseAuthOptional(cfg.FirebaseProjectID))

	// CU01 - Registrar paciente
	v1.POST("/patients", patientHandler.RegisterPatient)
	v1.GET("/patients/:id", patientHandler.GetPatientByID)

	v1.POST("/appointments", apptHandler.Create)
	v1.GET("/appointments", apptHandler.ListDay)
	v1.PATCH("/appointments/:id", apptHandler.Update)
	v1.GET("/appointments/:id", apptHandler.GetByID)
	v1.GET("/appointments/patient", apptHandler.ListByPatient)

	_ = db

	return r
}
