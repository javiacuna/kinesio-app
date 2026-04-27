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
	var models []AttachmentModel
	if err := r.db.WithContext(ctx).
		Order("created_at desc").
		Find(&models, "patient_id = ?", patientID.String()).
		Error; err != nil {
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

func toModel(attachment domain.Attachment) AttachmentModel {
	return AttachmentModel{
		ID:              attachment.ID.String(),
		PatientID:       attachment.PatientID.String(),
		FileName:        attachment.FileName,
		ContentType:     attachment.ContentType,
		SizeBytes:       attachment.SizeBytes,
		StoragePath:     attachment.StoragePath,
		Kind:            attachment.Kind,
		Notes:           attachment.Notes,
		UploadedByEmail: attachment.UploadedByEmail,
		UploadedByRole:  attachment.UploadedByRole,
		CreatedAt:       attachment.CreatedAt,
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
		Notes:           model.Notes,
		UploadedByEmail: model.UploadedByEmail,
		UploadedByRole:  model.UploadedByRole,
		CreatedAt:       model.CreatedAt,
	}
}
