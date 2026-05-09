package usecase

import (
	"context"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/javiacuna/kinesio-backend/internal/appointments/domain"
	"github.com/javiacuna/kinesio-backend/internal/appointments/ports"
)

type appointmentPackageUpdater interface {
	ports.Repository
	GetPackageByID(ctx context.Context, id uuid.UUID) (domain.AppointmentPackage, bool, error)
	ListByPackage(ctx context.Context, packageID uuid.UUID) ([]domain.Appointment, error)
	UpdatePackage(ctx context.Context, pkg domain.AppointmentPackage) (domain.AppointmentPackage, error)
	HasOverlapIgnoringAppointments(
		ctx context.Context,
		kinesiologistID uuid.UUID,
		startAt time.Time,
		endAt time.Time,
		excludeIDs []uuid.UUID,
	) (bool, error)
}

type UpdateAppointmentPackageInput struct {
	StartDate     *string
	StartTime     *string
	DurationMin   *int
	WorkDays      []int
	Notes         *string
	WorkStartTime string
	WorkEndTime   string
	AllowedDays   []int
}

type UpdateAppointmentPackageUseCase struct {
	repo appointmentPackageUpdater
}

func NewUpdateAppointmentPackageUseCase(repo appointmentPackageUpdater) *UpdateAppointmentPackageUseCase {
	return &UpdateAppointmentPackageUseCase{repo: repo}
}

func (uc *UpdateAppointmentPackageUseCase) GetPackage(
	ctx context.Context,
	id string,
) (domain.AppointmentPackage, bool, map[string]string, error) {
	packageID, err := uuid.Parse(strings.TrimSpace(id))
	if err != nil {
		return domain.AppointmentPackage{}, false, map[string]string{"id": "UUID inválido"}, domain.ErrValidation
	}
	pkg, found, err := uc.repo.GetPackageByID(ctx, packageID)
	return pkg, found, nil, err
}

func (uc *UpdateAppointmentPackageUseCase) Execute(
	ctx context.Context,
	id string,
	in UpdateAppointmentPackageInput,
) (domain.AppointmentPackage, []domain.Appointment, map[string]string, error) {
	errs := map[string]string{}

	packageID, err := uuid.Parse(strings.TrimSpace(id))
	if err != nil {
		return domain.AppointmentPackage{}, nil, map[string]string{"id": "UUID inválido"}, domain.ErrValidation
	}

	pkg, found, err := uc.repo.GetPackageByID(ctx, packageID)
	if err != nil {
		return domain.AppointmentPackage{}, nil, nil, err
	}
	if !found {
		return domain.AppointmentPackage{}, nil, nil, domain.ErrNotFound
	}

	newStartTime := pkg.StartTime
	if in.StartTime != nil {
		newStartTime = strings.TrimSpace(*in.StartTime)
		if _, err := time.Parse("15:04", newStartTime); err != nil {
			errs["start_time"] = "Formato inválido (usar HH:MM)"
		}
	}

	newDuration := pkg.DurationMin
	if in.DurationMin != nil {
		newDuration = *in.DurationMin
		if newDuration <= 0 {
			errs["duration_min"] = "Debe ser mayor a 0"
		}
	}

	if len(errs) > 0 {
		return domain.AppointmentPackage{}, nil, errs, domain.ErrValidation
	}

	loc := clinicLocation()
	newStartDate := pkg.StartDate.In(loc)
	if in.StartDate != nil {
		parsedDate, err := time.ParseInLocation("2006-01-02", strings.TrimSpace(*in.StartDate), loc)
		if err != nil {
			return domain.AppointmentPackage{}, nil, map[string]string{
				"start_date": "Formato inválido (usar YYYY-MM-DD)",
			}, domain.ErrValidation
		}
		newStartDate = parsedDate
	}

	packageDays := effectivePackageDays(in.WorkDays, pkg.WorkDays, in.AllowedDays)
	if len(in.WorkDays) > 0 && !daysAllowed(in.WorkDays, in.AllowedDays) {
		return domain.AppointmentPackage{}, nil, map[string]string{
			"work_days": "Los días elegidos deben estar dentro de los días laborales del kinesiólogo",
		}, domain.ErrValidation
	}

	active, err := uc.repo.IsPatientActive(ctx, pkg.PatientID)
	if err != nil {
		return domain.AppointmentPackage{}, nil, nil, err
	}
	if !active {
		return domain.AppointmentPackage{}, nil, nil, domain.ErrPatientInactive
	}

	appointments, err := uc.repo.ListByPackage(ctx, packageID)
	if err != nil {
		return domain.AppointmentPackage{}, nil, nil, err
	}

	timeChanged := in.StartTime != nil || in.DurationMin != nil
	scheduleChanged := timeChanged || in.StartDate != nil || len(in.WorkDays) > 0
	editableAppointments := make([]domain.Appointment, 0, len(appointments))
	for _, appointment := range appointments {
		if appointment.Status != domain.StatusScheduled || !appointment.StartAt.After(time.Now().UTC()) {
			continue
		}
		editableAppointments = append(editableAppointments, appointment)
	}

	var slots []PackageSlot
	if scheduleChanged && len(editableAppointments) > 0 {
		var slotErrs map[string]string
		slots, slotErrs, err = BuildPackageSlots(
			newStartDate.Format("2006-01-02"),
			newStartTime,
			newDuration,
			len(editableAppointments),
			false,
			packageDays,
		)
		if err != nil {
			return domain.AppointmentPackage{}, nil, slotErrs, err
		}
	}

	excludeIDs := make([]uuid.UUID, 0, len(editableAppointments))
	for _, appointment := range editableAppointments {
		excludeIDs = append(excludeIDs, appointment.ID)
	}

	updated := make([]domain.Appointment, 0, len(editableAppointments))
	for index, appointment := range editableAppointments {

		next := appointment
		if scheduleChanged {
			newStart := slots[index].StartAt
			newEnd := slots[index].EndAt
			if !newStart.After(time.Now().UTC()) {
				return domain.AppointmentPackage{}, nil, map[string]string{
					"start_time": "No se pueden programar sesiones en horarios pasados",
				}, domain.ErrValidation
			}
			if !insideWorkingHours(newStart, newEnd, in.WorkStartTime, in.WorkEndTime, in.AllowedDays) {
				return domain.AppointmentPackage{}, nil, map[string]string{
					"start_time": "Fuera del horario laboral o días laborales del kinesiólogo",
				}, domain.ErrValidation
			}

			overlap, err := uc.repo.HasOverlapIgnoringAppointments(ctx, appointment.KinesiologistID, newStart, newEnd, excludeIDs)
			if err != nil {
				return domain.AppointmentPackage{}, nil, nil, err
			}
			if overlap {
				return domain.AppointmentPackage{}, nil, map[string]string{
					"session": "La sesión se superpone con otro turno",
				}, domain.ErrOverlap
			}
			next.StartAt = newStart
			next.EndAt = newEnd
		}
		if in.Notes != nil {
			next.Notes = trimPtr(in.Notes)
		}

		stored, err := uc.repo.Update(ctx, next)
		if err != nil {
			return domain.AppointmentPackage{}, nil, nil, err
		}
		updated = append(updated, stored)
	}

	pkg.StartTime = newStartTime
	pkg.DurationMin = newDuration
	pkg.StartDate = newStartDate
	pkg.WorkDays = packageDays
	if in.Notes != nil {
		pkg.Notes = trimPtr(in.Notes)
	}
	updatedPackage, err := uc.repo.UpdatePackage(ctx, pkg)
	if err != nil {
		return domain.AppointmentPackage{}, nil, nil, err
	}

	return updatedPackage, updated, nil, nil
}

func rescheduleOnSameLocalDate(startAt time.Time, startTime string, durationMin int) (time.Time, time.Time, error) {
	loc := clinicLocation()
	local := startAt.In(loc)
	clock, err := time.Parse("15:04", strings.TrimSpace(startTime))
	if err != nil {
		return time.Time{}, time.Time{}, err
	}

	nextStart := time.Date(
		local.Year(),
		local.Month(),
		local.Day(),
		clock.Hour(),
		clock.Minute(),
		0,
		0,
		loc,
	)
	nextEnd := nextStart.Add(time.Duration(durationMin) * time.Minute)
	return nextStart.UTC(), nextEnd.UTC(), nil
}

func insideWorkingHours(startAt, endAt time.Time, workStartTime, workEndTime string, workDays []int) bool {
	loc := clinicLocation()
	localStart := startAt.In(loc)
	localEnd := endAt.In(loc)
	if localStart.Year() != localEnd.Year() || localStart.YearDay() != localEnd.YearDay() {
		return false
	}
	if len(workDays) > 0 && !workDayAllowed(localStart.Weekday(), workDays) {
		return false
	}

	startMin := localStart.Hour()*60 + localStart.Minute()
	endMin := localEnd.Hour()*60 + localEnd.Minute()
	workStartMin, ok := parseHHMMMinutes(workStartTime)
	if !ok {
		workStartMin = 8 * 60
	}
	workEndMin, ok := parseHHMMMinutes(workEndTime)
	if !ok {
		workEndMin = 20 * 60
	}

	return startMin >= workStartMin && endMin <= workEndMin
}

func workDayAllowed(weekday time.Weekday, workDays []int) bool {
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

func parseHHMMMinutes(value string) (int, bool) {
	parsed, err := time.Parse("15:04", strings.TrimSpace(value))
	if err != nil {
		return 0, false
	}
	return parsed.Hour()*60 + parsed.Minute(), true
}

func effectivePackageDays(inputDays []int, currentDays []int, allowedDays []int) []int {
	if len(inputDays) > 0 {
		return normalizeDays(inputDays)
	}
	if len(currentDays) > 0 {
		return normalizeDays(currentDays)
	}
	if len(allowedDays) > 0 {
		return normalizeDays(allowedDays)
	}
	return []int{1, 2, 3, 4, 5}
}

func normalizeDays(days []int) []int {
	seen := map[int]struct{}{}
	out := make([]int, 0, len(days))
	for _, day := range days {
		if day < 1 || day > 7 {
			continue
		}
		if _, ok := seen[day]; ok {
			continue
		}
		seen[day] = struct{}{}
		out = append(out, day)
	}
	return out
}

func daysAllowed(days []int, allowedDays []int) bool {
	if len(allowedDays) == 0 {
		return true
	}
	allowed := map[int]struct{}{}
	for _, day := range allowedDays {
		allowed[day] = struct{}{}
	}
	for _, day := range days {
		if _, ok := allowed[day]; !ok {
			return false
		}
	}
	return true
}
