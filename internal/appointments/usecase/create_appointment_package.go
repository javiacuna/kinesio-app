package usecase

import (
	"context"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/javiacuna/kinesio-backend/internal/appointments/domain"
	"github.com/javiacuna/kinesio-backend/internal/appointments/ports"
)

type appointmentPackageCreator interface {
	ports.Repository
	CreatePackageWithAppointments(
		ctx context.Context,
		pkg domain.AppointmentPackage,
		appointments []domain.Appointment,
	) (domain.AppointmentPackage, []domain.Appointment, error)
}

type CreateAppointmentPackageInput struct {
	PatientID       string
	KinesiologistID string
	PracticeID      *string
	FinancierID     *string
	StartDate       string
	StartTime       string
	DurationMin     int
	SessionsCount   int
	WeekdaysOnly    bool
	WorkDays        []int
	Notes           *string
}

type PackageSlot struct {
	SessionNumber int
	StartAt       time.Time
	EndAt         time.Time
}

type CreateAppointmentPackageUseCase struct {
	repo appointmentPackageCreator
}

func NewCreateAppointmentPackageUseCase(repo appointmentPackageCreator) *CreateAppointmentPackageUseCase {
	return &CreateAppointmentPackageUseCase{repo: repo}
}

func (uc *CreateAppointmentPackageUseCase) Execute(
	ctx context.Context,
	in CreateAppointmentPackageInput,
) (domain.AppointmentPackage, []domain.Appointment, map[string]string, error) {
	errs := map[string]string{}

	pid, err := uuid.Parse(strings.TrimSpace(in.PatientID))
	if err != nil {
		errs["patient_id"] = "UUID inválido"
	}
	kid, err := uuid.Parse(strings.TrimSpace(in.KinesiologistID))
	if err != nil {
		errs["kinesiologist_id"] = "UUID inválido"
	}
	var practiceID *uuid.UUID
	if in.PracticeID != nil && strings.TrimSpace(*in.PracticeID) != "" {
		id, err := uuid.Parse(strings.TrimSpace(*in.PracticeID))
		if err != nil {
			errs["practice_id"] = "UUID inválido"
		} else {
			practiceID = &id
		}
	}
	var financierID *uuid.UUID
	if in.FinancierID != nil && strings.TrimSpace(*in.FinancierID) != "" {
		id, err := uuid.Parse(strings.TrimSpace(*in.FinancierID))
		if err != nil {
			errs["financier_id"] = "UUID inválido"
		} else {
			financierID = &id
		}
	}

	slots, slotErrs, _ := BuildPackageSlots(in.StartDate, in.StartTime, in.DurationMin, in.SessionsCount, in.WeekdaysOnly, in.WorkDays)
	for key, value := range slotErrs {
		errs[key] = value
	}
	if len(errs) > 0 {
		return domain.AppointmentPackage{}, nil, errs, domain.ErrValidation
	}

	active, err := uc.repo.IsPatientActive(ctx, pid)
	if err != nil {
		return domain.AppointmentPackage{}, nil, nil, err
	}
	if !active {
		return domain.AppointmentPackage{}, nil, nil, domain.ErrPatientInactive
	}

	for _, slot := range slots {
		overlap, err := uc.repo.HasOverlap(ctx, kid, slot.StartAt, slot.EndAt, nil)
		if err != nil {
			return domain.AppointmentPackage{}, nil, nil, err
		}
		if overlap {
			return domain.AppointmentPackage{}, nil, map[string]string{
				"session": "La sesión " + strconv.Itoa(slot.SessionNumber) + " se superpone con otro turno",
			}, domain.ErrOverlap
		}
	}

	packageID := uuid.New()
	appointments := make([]domain.Appointment, 0, len(slots))
	for _, slot := range slots {
		sessionNumber := slot.SessionNumber
		appointments = append(appointments, domain.Appointment{
			ID:                   uuid.New(),
			PatientID:            pid,
			KinesiologistID:      kid,
			PracticeID:           practiceID,
			FinancierID:          financierID,
			PackageID:            &packageID,
			PackageSessionNumber: &sessionNumber,
			StartAt:              slot.StartAt,
			EndAt:                slot.EndAt,
			Status:               domain.StatusScheduled,
			Notes:                trimPtr(in.Notes),
		})
	}

	loc := clinicLocation()
	startDate, _ := time.ParseInLocation("2006-01-02", strings.TrimSpace(in.StartDate), loc)
	pkg := domain.AppointmentPackage{
		ID:              packageID,
		PatientID:       pid,
		KinesiologistID: kid,
		PracticeID:      practiceID,
		FinancierID:     financierID,
		SessionsCount:   in.SessionsCount,
		DurationMin:     in.DurationMin,
		StartDate:       startDate,
		StartTime:       strings.TrimSpace(in.StartTime),
		WeekdaysOnly:    in.WeekdaysOnly,
		WorkDays:        in.WorkDays,
		Notes:           trimPtr(in.Notes),
	}

	createdPackage, createdAppointments, err := uc.repo.CreatePackageWithAppointments(ctx, pkg, appointments)
	if err != nil {
		return domain.AppointmentPackage{}, nil, nil, err
	}
	return createdPackage, createdAppointments, nil, nil
}

func BuildPackageSlots(
	startDateValue string,
	startTimeValue string,
	durationMin int,
	sessionsCount int,
	weekdaysOnly bool,
	workDays []int,
) ([]PackageSlot, map[string]string, error) {
	errs := map[string]string{}
	loc := clinicLocation()

	startDate, err := time.ParseInLocation("2006-01-02", strings.TrimSpace(startDateValue), loc)
	if err != nil {
		errs["start_date"] = "Formato inválido (usar YYYY-MM-DD)"
	}
	startClock, err := time.Parse("15:04", strings.TrimSpace(startTimeValue))
	if err != nil {
		errs["start_time"] = "Formato inválido (usar HH:MM)"
	}
	if durationMin <= 0 {
		errs["duration_min"] = "Debe ser mayor a 0"
	}
	if sessionsCount <= 0 || sessionsCount > 80 {
		errs["sessions_count"] = "Debe estar entre 1 y 80"
	}
	if len(errs) > 0 {
		return nil, errs, domain.ErrValidation
	}

	slots := make([]PackageSlot, 0, sessionsCount)
	cursor := startDate
	for len(slots) < sessionsCount {
		if !packageDayAllowed(cursor.Weekday(), weekdaysOnly, workDays) {
			cursor = cursor.AddDate(0, 0, 1)
			continue
		}

		startLocal := time.Date(
			cursor.Year(),
			cursor.Month(),
			cursor.Day(),
			startClock.Hour(),
			startClock.Minute(),
			0,
			0,
			loc,
		)
		endLocal := startLocal.Add(time.Duration(durationMin) * time.Minute)
		if !startLocal.After(time.Now().In(loc)) {
			return nil, map[string]string{
				"start_date": "No se pueden crear paquetes con sesiones en horarios pasados",
			}, domain.ErrValidation
		}

		slots = append(slots, PackageSlot{
			SessionNumber: len(slots) + 1,
			StartAt:       startLocal.UTC(),
			EndAt:         endLocal.UTC(),
		})
		cursor = cursor.AddDate(0, 0, 1)
	}

	return slots, nil, nil
}

func packageDayAllowed(weekday time.Weekday, weekdaysOnly bool, workDays []int) bool {
	if len(workDays) > 0 {
		day := int(weekday)
		if day == 0 {
			day = 7
		}
		for _, allowed := range workDays {
			if allowed == day {
				return true
			}
		}
		return false
	}
	if weekdaysOnly {
		return weekday != time.Saturday && weekday != time.Sunday
	}
	return true
}

func clinicLocation() *time.Location {
	loc, err := time.LoadLocation("America/Argentina/Cordoba")
	if err == nil {
		return loc
	}
	return time.FixedZone("ART", -3*60*60)
}
