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

	kineHTTP "github.com/javiacuna/kinesio-backend/internal/kinesiologists/http"
	kineRepo "github.com/javiacuna/kinesio-backend/internal/kinesiologists/infra/gorm"
	kineUC "github.com/javiacuna/kinesio-backend/internal/kinesiologists/usecase"

	exercisePlanHTTP "github.com/javiacuna/kinesio-backend/internal/exerciseplans/http"
	exercisePlanGorm "github.com/javiacuna/kinesio-backend/internal/exerciseplans/infra/gorm"
	exercisePlanUC "github.com/javiacuna/kinesio-backend/internal/exerciseplans/usecase"

	evoHTTP "github.com/javiacuna/kinesio-backend/internal/evolutions/http"
	evoGorm "github.com/javiacuna/kinesio-backend/internal/evolutions/infra/gorm"
	evoUC "github.com/javiacuna/kinesio-backend/internal/evolutions/usecase"

	matHTTP "github.com/javiacuna/kinesio-backend/internal/materials/http"
	matGorm "github.com/javiacuna/kinesio-backend/internal/materials/infra/gorm"
	matUC "github.com/javiacuna/kinesio-backend/internal/materials/usecase"
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
	searchPatients := patientsUC.NewSearchPatientsUseCase(patientRepo)
	patientHandler := patientsHTTP.NewHandler(registerPatientUC, getPatientByIDUC, searchPatients)

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

	kRepo := kineRepo.New(db)
	listKUC := kineUC.NewListKinesiologistsUseCase(kRepo)
	kHandler := kineHTTP.NewHandler(listKUC)

	planRepo := exercisePlanGorm.NewRepository(db)
	planCreateUC := exercisePlanUC.NewCreatePlanUseCase(planRepo)
	planListUC := exercisePlanUC.NewListPlansByPatientUseCase(planRepo)
	planGetUC := exercisePlanUC.NewGetPlanByIDUseCase(planRepo)
	planHandler := exercisePlanHTTP.NewHandler(planCreateUC, planListUC, planGetUC)

	evoRepo := evoGorm.NewRepository(db)
	evoCreateUC := evoUC.NewCreateEvolutionUseCase(evoRepo)
	evoListUC := evoUC.NewListEvolutionsByPatientUseCase(evoRepo)
	evoGetUC := evoUC.NewGetEvolutionByIDUseCase(evoRepo)
	evoHandler := evoHTTP.NewHandler(evoCreateUC, evoListUC, evoGetUC)

	matRepo := matGorm.NewRepository(db)
	matCreateUC := matUC.NewCreateMaterialUseCase(matRepo)
	matListUC := matUC.NewListMaterialsUseCase(matRepo)
	matLoanUC := matUC.NewLoanMaterialUseCase(matRepo)
	matReturnUC := matUC.NewReturnMaterialUseCase(matRepo)
	matListLoansUC := matUC.NewListLoansByPatientUseCase(matRepo)
	matHandler := matHTTP.NewHandler(matCreateUC, matListUC, matLoanUC, matReturnUC, matListLoansUC)

	// API v1
	v1 := r.Group("/api/v1")

	// Auth: placeholder para Firebase (por ahora opcional)
	// Cuando se setee FIREBASE_PROJECT_ID, este middleware exigirá JWTs (Authorization: Bearer <token>).
	v1.Use(middleware.FirebaseAuthOptional(cfg.FirebaseProjectID))

	// CU01 - Registrar paciente
	v1.POST("/patients", patientHandler.RegisterPatient)
	v1.GET("/patients/:id", patientHandler.GetPatientByID)
	v1.GET("/patients", patientHandler.Search)

	v1.POST("/appointments", apptHandler.Create)
	v1.GET("/appointments", apptHandler.ListDay)
	v1.PATCH("/appointments/:id", apptHandler.Update)
	v1.GET("/appointments/:id", apptHandler.GetByID)
	v1.GET("/appointments/patient", apptHandler.ListByPatient)

	v1.GET("/kinesiologists", kHandler.List)

	v1.POST("/patients/:patient_id/exercise-plans", planHandler.CreateForPatient)
	v1.GET("/patients/:patient_id/exercise-plans", planHandler.ListByPatient)
	v1.GET("/exercise-plans/:plan_id", planHandler.GetByID)

	v1.POST("/patients/:patient_id/evolutions", evoHandler.CreateForPatient)
	v1.GET("/patients/:patient_id/evolutions", evoHandler.ListByPatient)
	v1.GET("/evolutions/:evolution_id", evoHandler.GetByID)

	v1.POST("/materials", matHandler.CreateMaterial)
	v1.GET("/materials", matHandler.ListMaterials)

	v1.POST("/material-loans", matHandler.LoanMaterial)
	v1.POST("/material-loans/:loan_id/return", matHandler.ReturnLoan)

	v1.GET("/patients/:patient_id/material-loans", matHandler.ListLoansByPatient)

	_ = db

	return r
}
