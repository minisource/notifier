package service

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/minisource/go-common/logging"
	"github.com/minisource/notifier/internal/models"
	"github.com/minisource/notifier/internal/repository"
)

// TemplateService handles notification template operations
type TemplateService struct {
	templateRepo repository.NotificationTemplateRepository
	logger       logging.Logger
}

// NewTemplateService creates a new template service
func NewTemplateService(templateRepo repository.NotificationTemplateRepository, logger logging.Logger) *TemplateService {
	return &TemplateService{
		templateRepo: templateRepo,
		logger:       logger,
	}
}

// CreateTemplate creates a new notification template
func (s *TemplateService) CreateTemplate(ctx context.Context, template *models.NotificationTemplate) (*models.NotificationTemplate, error) {
	// Validate template
	if err := s.validateTemplate(template); err != nil {
		s.logger.Error(logging.Validation, logging.Insert, "Template validation failed", map[logging.ExtraKey]interface{}{
			"error": err.Error(),
			"name":  template.Name,
		})
		return nil, err
	}

	// Generate UUID if not set
	if template.ID == uuid.Nil {
		template.ID = uuid.New()
	}

	// Create template in repository
	if err := s.templateRepo.Create(ctx, template); err != nil {
		s.logger.Error(logging.Postgres, logging.Insert, "Failed to create template", map[logging.ExtraKey]interface{}{
			"error": err.Error(),
			"name":  template.Name,
		})
		return nil, fmt.Errorf("failed to create template: %w", err)
	}

	s.logger.Info(logging.General, logging.Insert, "Template created successfully", map[logging.ExtraKey]interface{}{
		"templateId": template.ID,
		"name":       template.Name,
	})

	return template, nil
}

// GetTemplate retrieves a template by ID
func (s *TemplateService) GetTemplate(ctx context.Context, id uuid.UUID) (*models.NotificationTemplate, error) {
	template, err := s.templateRepo.GetByID(ctx, id)
	if err != nil {
		s.logger.Error(logging.Postgres, logging.Select, "Failed to get template", map[logging.ExtraKey]interface{}{
			"error":      err.Error(),
			"templateId": id,
		})
		return nil, fmt.Errorf("failed to get template: %w", err)
	}

	if template == nil {
		return nil, fmt.Errorf("template not found")
	}

	return template, nil
}

// GetTemplateByName retrieves a template by name and type
func (s *TemplateService) GetTemplateByName(ctx context.Context, name string, notifType models.NotificationType) (*models.NotificationTemplate, error) {
	template, err := s.templateRepo.GetByName(ctx, name, notifType)
	if err != nil {
		s.logger.Error(logging.Postgres, logging.Select, "Failed to get template by name", map[logging.ExtraKey]interface{}{
			"error": err.Error(),
			"name":  name,
			"type":  notifType,
		})
		return nil, fmt.Errorf("failed to get template: %w", err)
	}

	if template == nil {
		return nil, fmt.Errorf("template not found")
	}

	return template, nil
}

// ListTemplates retrieves all templates with pagination
func (s *TemplateService) ListTemplates(ctx context.Context, page, pageSize int) ([]*models.NotificationTemplate, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	offset := (page - 1) * pageSize
	templates, total, err := s.templateRepo.GetAll(ctx, pageSize, offset)
	if err != nil {
		s.logger.Error(logging.Postgres, logging.Select, "Failed to list templates", map[logging.ExtraKey]interface{}{
			"error": err.Error(),
		})
		return nil, 0, fmt.Errorf("failed to list templates: %w", err)
	}

	return templates, total, nil
}

// UpdateTemplate updates an existing template
func (s *TemplateService) UpdateTemplate(ctx context.Context, id uuid.UUID, updates *models.NotificationTemplate) (*models.NotificationTemplate, error) {
	// Check if template exists
	existing, err := s.GetTemplate(ctx, id)
	if err != nil {
		return nil, err
	}

	// Validate updates
	if err := s.validateTemplate(updates); err != nil {
		s.logger.Error(logging.Validation, logging.Update, "Template validation failed", map[logging.ExtraKey]interface{}{
			"error":      err.Error(),
			"templateId": id,
		})
		return nil, err
	}

	// Update fields
	existing.Name = updates.Name
	existing.Subject = updates.Subject
	existing.Body = updates.Body
	existing.Description = updates.Description
	existing.Variables = updates.Variables
	existing.IsActive = updates.IsActive

	// Update in repository
	if err := s.templateRepo.Update(ctx, existing); err != nil {
		s.logger.Error(logging.Postgres, logging.Update, "Failed to update template", map[logging.ExtraKey]interface{}{
			"error":      err.Error(),
			"templateId": id,
		})
		return nil, fmt.Errorf("failed to update template: %w", err)
	}

	s.logger.Info(logging.General, logging.Update, "Template updated successfully", map[logging.ExtraKey]interface{}{
		"templateId": id,
		"name":       existing.Name,
	})

	return existing, nil
}

// DeleteTemplate soft deletes a template
func (s *TemplateService) DeleteTemplate(ctx context.Context, id uuid.UUID) error {
	// Check if template exists
	if _, err := s.GetTemplate(ctx, id); err != nil {
		return err
	}

	// Delete from repository
	if err := s.templateRepo.Delete(ctx, id); err != nil {
		s.logger.Error(logging.Postgres, logging.Delete, "Failed to delete template", map[logging.ExtraKey]interface{}{
			"error":      err.Error(),
			"templateId": id,
		})
		return fmt.Errorf("failed to delete template: %w", err)
	}

	s.logger.Info(logging.General, logging.Delete, "Template deleted successfully", map[logging.ExtraKey]interface{}{
		"templateId": id,
	})

	return nil
}

// validateTemplate validates template fields
func (s *TemplateService) validateTemplate(template *models.NotificationTemplate) error {
	if template == nil {
		return fmt.Errorf("template cannot be nil")
	}

	if template.Name == "" {
		return fmt.Errorf("template name is required")
	}

	if template.Subject == "" && template.Body == "" {
		return fmt.Errorf("template must have either subject or body")
	}

	// Validate template syntax (basic check for {{variables}})
	// TODO: Implement more sophisticated template validation

	return nil
}
