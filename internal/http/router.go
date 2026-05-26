package http

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
	"gorm.io/gorm"

	authFirebase "github.com/javiacuna/kinesio-backend/internal/auth/firebase"
	authHTTP "github.com/javiacuna/kinesio-backend/internal/auth/http"
	"github.com/javiacuna/kinesio-backend/internal/config"
	"github.com/javiacuna/kinesio-backend/internal/http/middleware"

	diagnosesHTTP "github.com/javiacuna/kinesio-backend/internal/diagnoses/http"
	diagnosesGorm "github.com/javiacuna/kinesio-backend/internal/diagnoses/infra/gorm"
	diagnosesUC "github.com/javiacuna/kinesio-backend/internal/diagnoses/usecase"

	patientsHTTP "github.com/javiacuna/kinesio-backend/internal/patients/http"
	patientsRepo "github.com/javiacuna/kinesio-backend/internal/patients/infra/gorm"
	patientsUC "github.com/javiacuna/kinesio-backend/internal/patients/usecase"

	appointmentsHTTP "github.com/javiacuna/kinesio-backend/internal/appointments/http"
	appointmentsRepo "github.com/javiacuna/kinesio-backend/internal/appointments/infra/gorm"
	appointmentsUC "github.com/javiacuna/kinesio-backend/internal/appointments/usecase"

	kineHTTP "github.com/javiacuna/kinesio-backend/internal/kinesiologists/http"
	kineRepo "github.com/javiacuna/kinesio-backend/internal/kinesiologists/infra/gorm"
	kineUC "github.com/javiacuna/kinesio-backend/internal/kinesiologists/usecase"

	kineAttachmentsHTTP "github.com/javiacuna/kinesio-backend/internal/kinesiologistattachments/http"
	kineAttachmentsGorm "github.com/javiacuna/kinesio-backend/internal/kinesiologistattachments/infra/gorm"

	exercisePlanHTTP "github.com/javiacuna/kinesio-backend/internal/exerciseplans/http"
	exercisePlanGorm "github.com/javiacuna/kinesio-backend/internal/exerciseplans/infra/gorm"
	exercisePlanUC "github.com/javiacuna/kinesio-backend/internal/exerciseplans/usecase"

	evoHTTP "github.com/javiacuna/kinesio-backend/internal/evolutions/http"
	evoGorm "github.com/javiacuna/kinesio-backend/internal/evolutions/infra/gorm"
	evoUC "github.com/javiacuna/kinesio-backend/internal/evolutions/usecase"

	financeHTTP "github.com/javiacuna/kinesio-backend/internal/finance/http"
	financeGorm "github.com/javiacuna/kinesio-backend/internal/finance/infra/gorm"

	notificationsEmail "github.com/javiacuna/kinesio-backend/internal/notifications/email"
	notificationsHTTP "github.com/javiacuna/kinesio-backend/internal/notifications/http"
	notificationsGorm "github.com/javiacuna/kinesio-backend/internal/notifications/infra/gorm"
	notificationsService "github.com/javiacuna/kinesio-backend/internal/notifications/service"

	matHTTP "github.com/javiacuna/kinesio-backend/internal/materials/http"
	matGorm "github.com/javiacuna/kinesio-backend/internal/materials/infra/gorm"
	matUC "github.com/javiacuna/kinesio-backend/internal/materials/usecase"

	attachmentsHTTP "github.com/javiacuna/kinesio-backend/internal/patientattachments/http"
	attachmentsGorm "github.com/javiacuna/kinesio-backend/internal/patientattachments/infra/gorm"

	staffHTTP "github.com/javiacuna/kinesio-backend/internal/staff/http"
	staffRepo "github.com/javiacuna/kinesio-backend/internal/staff/infra/gorm"
	staffUC "github.com/javiacuna/kinesio-backend/internal/staff/usecase"
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

	firebaseAuthClient, err := authFirebase.NewAuthClient(context.Background(), cfg.FirebaseProjectID, cfg.FirebaseCredentialsFile)
	if err != nil {
		log.Warn().Err(err).Msg("firebase admin auth client could not be initialized")
	}
	authHandler := authHTTP.NewHandler(cfg.FirebaseWebAPIKey, firebaseAuthClient)

	staffRepository := staffRepo.New(db)
	listStaffUC := staffUC.NewListStaffMembersUseCase(staffRepository)
	saveStaffUC := staffUC.NewSaveStaffMemberUseCase(staffRepository)
	staffHandler := staffHTTP.NewHandler(listStaffUC, saveStaffUC)

	// Patients wiring
	patientRepo := patientsRepo.New(db)
	registerPatientUC := patientsUC.NewRegisterPatientUseCase(patientRepo)
	updatePatientUC := patientsUC.NewUpdatePatientUseCase(patientRepo)
	deletePatientUC := patientsUC.NewDeletePatientUseCase(patientRepo)
	getPatientByIDUC := patientsUC.NewGetPatientByIDUseCase(patientRepo)
	listPatientsUC := patientsUC.NewListPatientsUseCase(patientRepo)
	searchPatients := patientsUC.NewSearchPatientsUseCase(patientRepo)
	patientHandler := patientsHTTP.NewHandler(registerPatientUC, updatePatientUC, deletePatientUC, getPatientByIDUC, listPatientsUC, searchPatients)

	kRepo := kineRepo.New(db)
	listKUC := kineUC.NewListKinesiologistsUseCase(kRepo)
	saveKUC := kineUC.NewSaveKinesiologistUseCase(kRepo)
	listSpecialtiesUC := kineUC.NewListSpecialtiesUseCase(kRepo)
	saveSpecialtyUC := kineUC.NewSaveSpecialtyUseCase(kRepo)
	listPracticesUC := kineUC.NewListPracticesUseCase(kRepo)
	savePracticeUC := kineUC.NewSavePracticeUseCase(kRepo)
	kHandler := kineHTTP.NewHandler(listKUC, saveKUC, listSpecialtiesUC, saveSpecialtyUC, listPracticesUC, savePracticeUC)

	apptRepo := appointmentsRepo.New(db)
	createApptUC := appointmentsUC.NewCreateAppointmentUseCase(apptRepo)
	listDayUC := appointmentsUC.NewListAppointmentsDayUseCase(apptRepo)
	updateApptUC := appointmentsUC.NewUpdateAppointmentUseCase(apptRepo)
	cancelApptUC := appointmentsUC.NewCancelAppointmentUseCase(apptRepo)
	createApptPackageUC := appointmentsUC.NewCreateAppointmentPackageUseCase(apptRepo)
	updateApptPackageUC := appointmentsUC.NewUpdateAppointmentPackageUseCase(apptRepo)
	notificationRepo := notificationsGorm.NewRepository(db)
	notificationMailer := notificationsEmail.NewMailer(notificationsEmail.Config{
		Enabled:   cfg.NotificationEmailEnabled,
		Host:      cfg.SMTPHost,
		Port:      cfg.SMTPPort,
		Username:  cfg.SMTPUsername,
		Password:  cfg.SMTPPassword,
		FromEmail: cfg.SMTPFromEmail,
		FromName:  cfg.SMTPFromName,
	})
	notificationService := notificationsService.New(notificationRepo, notificationMailer)
	notificationHandler := notificationsHTTP.NewHandler(notificationService)

	getApptByIDUC := appointmentsUC.NewGetAppointmentByIDUseCase(apptRepo)
	listByPatientUC := appointmentsUC.NewListAppointmentsByPatientUseCase(apptRepo)

	apptHandler := appointmentsHTTP.NewHandler(
		createApptUC,
		listDayUC,
		updateApptUC,
		cancelApptUC,
		getApptByIDUC,
		listByPatientUC,
		kRepo,
		patientRepo,
		createApptPackageUC,
		updateApptPackageUC,
		notificationService,
	)

	planRepo := exercisePlanGorm.NewRepository(db)
	planCreateUC := exercisePlanUC.NewCreatePlanUseCase(planRepo)
	planListUC := exercisePlanUC.NewListPlansByPatientUseCase(planRepo)
	planGetUC := exercisePlanUC.NewGetPlanByIDUseCase(planRepo)
	planUpdateUC := exercisePlanUC.NewUpdatePlanUseCase(planRepo)
	planHandler := exercisePlanHTTP.NewHandler(planCreateUC, planListUC, planGetUC, planUpdateUC, patientRepo, notificationService)

	evoRepo := evoGorm.NewRepository(db)
	evoCreateUC := evoUC.NewCreateEvolutionUseCase(evoRepo)
	evoListUC := evoUC.NewListEvolutionsByPatientUseCase(evoRepo)
	evoGetUC := evoUC.NewGetEvolutionByIDUseCase(evoRepo)
	evoHandler := evoHTTP.NewHandler(evoCreateUC, evoListUC, evoGetUC)

	matRepo := matGorm.NewRepository(db)
	matCreateUC := matUC.NewCreateMaterialUseCase(matRepo)
	matUpdateUC := matUC.NewUpdateMaterialUseCase(matRepo)
	matListUC := matUC.NewListMaterialsUseCase(matRepo)
	matLoanUC := matUC.NewLoanMaterialUseCase(matRepo)
	matReturnUC := matUC.NewReturnMaterialUseCase(matRepo)
	matListAllLoansUC := matUC.NewListLoansUseCase(matRepo)
	matListLoansUC := matUC.NewListLoansByPatientUseCase(matRepo)
	matHandler := matHTTP.NewHandler(
		matCreateUC,
		matUpdateUC,
		matListUC,
		matLoanUC,
		matReturnUC,
		matListAllLoansUC,
		matListLoansUC,
		matRepo,
		staffRepository,
		notificationRepo,
	)

	diagnosesRepo := diagnosesGorm.NewRepository(db)
	diagnosesSearchUC := diagnosesUC.NewSearchCIE10UseCase(diagnosesRepo)
	diagnosesListUC := diagnosesUC.NewListPatientDiagnosesUseCase(diagnosesRepo)
	diagnosesSaveUC := diagnosesUC.NewSavePatientDiagnosisUseCase(diagnosesRepo)
	diagnosesDeleteUC := diagnosesUC.NewDeletePatientDiagnosisUseCase(diagnosesRepo)
	diagnosesHandler := diagnosesHTTP.NewHandler(diagnosesSearchUC, diagnosesListUC, diagnosesSaveUC, diagnosesDeleteUC)

	attachmentRepo := attachmentsGorm.NewRepository(db)
	attachmentHandler := attachmentsHTTP.NewHandler(attachmentRepo, cfg.PatientFilesDir, patientRepo)
	kineAttachmentRepo := kineAttachmentsGorm.NewRepository(db)
	kineAttachmentHandler := kineAttachmentsHTTP.NewHandler(kineAttachmentRepo, cfg.KinesiologistFilesDir)
	financeRepo := financeGorm.NewRepository(db)
	financeHandler := financeHTTP.NewHandler(financeRepo)

	// API v1
	v1 := r.Group("/api/v1")

	v1.POST("/auth/login", authHandler.Login)
	v1.POST("/auth/password-reset", authHandler.RequestPasswordReset)

	// Cuando se setee FIREBASE_PROJECT_ID, este middleware exige y valida un ID token de Firebase.
	v1.Use(middleware.FirebaseAuthOptional(cfg.FirebaseProjectID, firebaseAuthClient))
	v1.GET("/auth/me", authHandler.Me)
	v1.POST("/auth/change-password", authHandler.ChangePassword)
	v1.GET("/notifications", middleware.RequireAuth(), notificationHandler.List)
	v1.GET("/notifications/unread-count", middleware.RequireAuth(), notificationHandler.UnreadCount)
	v1.POST("/notifications/:notification_id/read", middleware.RequireAuth(), notificationHandler.MarkRead)
	v1.POST("/notifications/read-all", middleware.RequireAuth(), notificationHandler.MarkAllRead)
	v1.GET("/admin/users", authHandler.ListUsers)
	v1.POST("/admin/users/invite", authHandler.InviteUser)
	v1.POST("/admin/users/role", authHandler.AssignRole)
	v1.GET("/staff", staffHandler.List)
	v1.POST("/staff", staffHandler.Create)
	v1.PUT("/staff/:id", staffHandler.Update)

	// CU01 - Registrar paciente
	v1.POST("/patients", patientHandler.RegisterPatient)
	v1.GET("/patients", patientHandler.Search)
	v1.POST("/patients/:patient_id/plans", planHandler.CreateForPatient)
	v1.GET("/patients/:patient_id/plans", planHandler.ListByPatient)
	v1.POST("/patients/:patient_id/evolutions", middleware.RequireRole("kinesiologo"), evoHandler.CreateForPatient)
	v1.GET("/patients/:patient_id/evolutions", middleware.RequireRole("recepcionista", "kinesiologo"), evoHandler.ListByPatient)
	v1.GET("/patients/:patient_id/material-loans", middleware.RequireRole("recepcionista", "kinesiologo"), matHandler.ListLoansByPatient)
	v1.GET("/patients/:patient_id/diagnoses", middleware.RequireRole("recepcionista", "kinesiologo"), diagnosesHandler.ListByPatient)
	v1.POST("/patients/:patient_id/diagnoses", middleware.RequireRole("recepcionista", "kinesiologo"), diagnosesHandler.CreateForPatient)
	v1.PUT("/patients/:patient_id/diagnoses/:diagnosis_id", middleware.RequireRole("recepcionista", "kinesiologo"), diagnosesHandler.Update)
	v1.PATCH("/patients/:patient_id/diagnoses/:diagnosis_id", middleware.RequireRole("recepcionista", "kinesiologo"), diagnosesHandler.Update)
	v1.DELETE("/patients/:patient_id/diagnoses/:diagnosis_id", middleware.RequireRole("recepcionista", "kinesiologo"), diagnosesHandler.Delete)
	v1.GET("/patients/:patient_id/attachments", attachmentHandler.ListByPatient)
	v1.POST("/patients/:patient_id/attachments", attachmentHandler.Upload)
	v1.PUT("/patients/:patient_id", patientHandler.UpdatePatient)
	v1.DELETE("/patients/:patient_id", patientHandler.DeletePatient)
	v1.GET("/patients/:patient_id", middleware.RequireRole("recepcionista", "kinesiologo"), patientHandler.GetPatientByID)

	v1.POST("/appointments", apptHandler.Create)
	v1.POST("/appointment-packages", middleware.RequireRole("recepcionista"), apptHandler.CreatePackage)
	v1.PUT("/appointment-packages/:package_id", middleware.RequireRole("recepcionista"), apptHandler.UpdatePackage)
	v1.PATCH("/appointment-packages/:package_id", middleware.RequireRole("recepcionista"), apptHandler.UpdatePackage)
	v1.GET("/appointments", middleware.RequireRole("recepcionista", "kinesiologo"), apptHandler.ListDay)
	v1.GET("/appointments/patient", apptHandler.ListByPatient)
	v1.PUT("/appointments/:id", apptHandler.Update)
	v1.PATCH("/appointments/:id", apptHandler.Update)
	v1.DELETE("/appointments/:id", apptHandler.Cancel)
	v1.GET("/appointments/:id", apptHandler.GetByID)

	v1.GET("/kinesiologists", middleware.RequireRole("recepcionista", "kinesiologo", "paciente"), kHandler.List)
	v1.POST("/kinesiologists", kHandler.Create)
	v1.PUT("/kinesiologists/:id", kHandler.Update)
	v1.GET("/specialties", middleware.RequireRole("recepcionista", "kinesiologo", "paciente"), kHandler.ListSpecialties)
	v1.POST("/specialties", kHandler.CreateSpecialty)
	v1.PUT("/specialties/:id", kHandler.UpdateSpecialty)
	v1.GET("/practices", middleware.RequireRole("recepcionista", "kinesiologo", "paciente"), kHandler.ListPractices)
	v1.POST("/practices", kHandler.CreatePractice)
	v1.PUT("/practices/:id", kHandler.UpdatePractice)
	v1.GET("/kinesiologists/:kinesiologist_id/attachments", middleware.RequireRole("recepcionista", "kinesiologo"), kineAttachmentHandler.ListByKinesiologist)
	v1.POST("/kinesiologists/:kinesiologist_id/attachments", middleware.RequireRole("recepcionista", "kinesiologo"), kineAttachmentHandler.Upload)

	v1.POST("/plans", planHandler.Create)
	v1.GET("/plans/:plan_id", planHandler.GetByID)
	v1.PUT("/plans/:plan_id", planHandler.Update)
	v1.PATCH("/plans/:plan_id", planHandler.Update)

	v1.GET("/evolutions/:evolution_id", middleware.RequireRole("recepcionista", "kinesiologo"), evoHandler.GetByID)

	v1.GET("/cie10", middleware.RequireRole("recepcionista", "kinesiologo"), diagnosesHandler.SearchCIE10)

	v1.POST("/materials", middleware.RequireRole("recepcionista", "kinesiologo"), matHandler.CreateMaterial)
	v1.GET("/materials", middleware.RequireRole("recepcionista", "kinesiologo"), matHandler.ListMaterials)
	v1.PUT("/materials/:material_id", middleware.RequireRole("recepcionista", "kinesiologo"), matHandler.UpdateMaterial)

	v1.GET("/material-loans", middleware.RequireRole("recepcionista", "kinesiologo"), matHandler.ListLoans)
	v1.POST("/material-loans", middleware.RequireRole("recepcionista", "kinesiologo"), matHandler.LoanMaterial)
	v1.POST("/material-loans/:loan_id/return", middleware.RequireRole("recepcionista", "kinesiologo"), matHandler.ReturnLoan)
	v1.GET("/patient-attachments/:attachment_id/download", attachmentHandler.Download)
	v1.PATCH("/patient-attachments/:attachment_id", attachmentHandler.Update)
	v1.DELETE("/patient-attachments/:attachment_id", attachmentHandler.Delete)
	v1.GET("/kinesiologist-attachments/:attachment_id/download", middleware.RequireRole("recepcionista", "kinesiologo"), kineAttachmentHandler.Download)

	v1.GET("/financiers", middleware.RequireRole("admin", "recepcionista", "kinesiologo"), financeHandler.ListFinanciers)
	v1.POST("/financiers", middleware.RequireRole("admin", "recepcionista"), financeHandler.CreateFinancier)
	v1.PUT("/financiers/:id", middleware.RequireRole("admin", "recepcionista"), financeHandler.UpdateFinancier)
	v1.GET("/financial/tariffs", middleware.RequireRole("admin", "recepcionista", "kinesiologo"), financeHandler.ListTariffs)
	v1.POST("/financial/tariffs", middleware.RequireRole("admin", "recepcionista"), financeHandler.CreateTariff)
	v1.PUT("/financial/tariffs/:id", middleware.RequireRole("admin", "recepcionista"), financeHandler.UpdateTariff)
	v1.GET("/financial/fee-rules", middleware.RequireRole("admin", "recepcionista"), financeHandler.ListFeeRules)
	v1.POST("/financial/fee-rules", middleware.RequireRole("admin", "recepcionista"), financeHandler.CreateFeeRule)
	v1.PUT("/financial/fee-rules/:id", middleware.RequireRole("admin", "recepcionista"), financeHandler.UpdateFeeRule)
	v1.GET("/financial/movements", middleware.RequireRole("admin", "recepcionista"), financeHandler.ListMovements)
	v1.PATCH("/financial/movements/:movement_id/status", middleware.RequireRole("admin", "recepcionista"), financeHandler.UpdateMovementStatuses)
	v1.POST("/financial/appointments/:appointment_id/complete", middleware.RequireRole("admin", "recepcionista", "kinesiologo"), financeHandler.CompleteAppointment)

	_ = db

	return r
}
