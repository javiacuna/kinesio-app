package gorm

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/javiacuna/kinesio-backend/internal/patientattachments/domain"
)

type Repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) Create(ctx context.Context, attachment domain.Attachment) (domain.Attachment, error) {
	model := toModel(attachment)
	if err := r.db.WithContext(ctx).Create(&model).Error; err != nil {
		return domain.Attachment{}, err
	}
	return toDomain(model), nil
}

func (r *Repository) ListByPatient(ctx context.Context, patientID uuid.UUID) ([]domain.Attachment, error) {
	return r.listByPatient(ctx, patientID, false)
}

func (r *Repository) ListVisibleByPatient(ctx context.Context, patientID uuid.UUID) ([]domain.Attachment, error) {
	return r.listByPatient(ctx, patientID, true)
}

func (r *Repository) listByPatient(ctx context.Context, patientID uuid.UUID, visibleOnly bool) ([]domain.Attachment, error) {
	var models []AttachmentModel
	tx := r.db.WithContext(ctx).Where("patient_id = ?", patientID.String())
	if visibleOnly {
		tx = tx.Where("patient_visible = true")
	}
	if err := tx.Order("created_at desc").Find(&models).Error; err != nil {
		return nil, err
	}

	out := make([]domain.Attachment, 0, len(models))
	for _, model := range models {
		out = append(out, toDomain(model))
	}
	return out, nil
}

func (r *Repository) GetByID(ctx context.Context, id uuid.UUID) (domain.Attachment, bool, error) {
	var model AttachmentModel
	err := r.db.WithContext(ctx).First(&model, "id = ?", id.String()).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return domain.Attachment{}, false, nil
		}
		return domain.Attachment{}, false, err
	}
	return toDomain(model), true, nil
}

func (r *Repository) Update(ctx context.Context, attachment domain.Attachment) (domain.Attachment, error) {
	updates := map[string]any{
		"file_name":        attachment.FileName,
		"category":         attachment.Category,
		"patient_visible":  attachment.PatientVisible,
		"notes":            attachment.Notes,
		"updated_by_email": attachment.UpdatedByEmail,
		"updated_by_role":  attachment.UpdatedByRole,
		"updated_at":       attachment.UpdatedAt,
	}

	tx := r.db.WithContext(ctx).
		Model(&AttachmentModel{}).
		Where("id = ?", attachment.ID.String()).
		Updates(updates)
	if tx.Error != nil {
		return domain.Attachment{}, tx.Error
	}
	if tx.RowsAffected == 0 {
		return domain.Attachment{}, gorm.ErrRecordNotFound
	}

	out, found, err := r.GetByID(ctx, attachment.ID)
	if err != nil {
		return domain.Attachment{}, err
	}
	if !found {
		return domain.Attachment{}, gorm.ErrRecordNotFound
	}
	return out, nil
}

func (r *Repository) Delete(ctx context.Context, id uuid.UUID) error {
	tx := r.db.WithContext(ctx).Delete(&AttachmentModel{}, "id = ?", id.String())
	if tx.Error != nil {
		return tx.Error
	}
	if tx.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

func toModel(attachment domain.Attachment) AttachmentModel {
	return AttachmentModel{
		ID:              attachment.ID.String(),
		PatientID:       attachment.PatientID.String(),
		FileName:        attachment.FileName,
		ContentType:     attachment.ContentType,
		SizeBytes:       attachment.SizeBytes,
		StoragePath:     attachment.StoragePath,
		Kind:            attachment.Kind,
		Category:        attachment.Category,
		PatientVisible:  attachment.PatientVisible,
		Notes:           attachment.Notes,
		UploadedByEmail: attachment.UploadedByEmail,
		UploadedByRole:  attachment.UploadedByRole,
		CreatedAt:       attachment.CreatedAt,
		UpdatedByEmail:  attachment.UpdatedByEmail,
		UpdatedByRole:   attachment.UpdatedByRole,
		UpdatedAt:       attachment.UpdatedAt,
	}
}

func toDomain(model AttachmentModel) domain.Attachment {
	return domain.Attachment{
		ID:              uuid.MustParse(model.ID),
		PatientID:       uuid.MustParse(model.PatientID),
		FileName:        model.FileName,
		ContentType:     model.ContentType,
		SizeBytes:       model.SizeBytes,
		StoragePath:     model.StoragePath,
		Kind:            model.Kind,
		Category:        model.Category,
		PatientVisible:  model.PatientVisible,
		Notes:           model.Notes,
		UploadedByEmail: model.UploadedByEmail,
		UploadedByRole:  model.UploadedByRole,
		CreatedAt:       model.CreatedAt,
		UpdatedByEmail:  model.UpdatedByEmail,
		UpdatedByRole:   model.UpdatedByRole,
		UpdatedAt:       model.UpdatedAt,
	}
}
