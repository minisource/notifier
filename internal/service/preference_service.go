package service

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/minisource/go-common/logging"
	"github.com/minisource/notifier/internal/models"
	"github.com/minisource/notifier/internal/repository"
)

// PreferenceService handles notification preference operations
type PreferenceService struct {
	preferenceRepo repository.NotificationPreferenceRepository
	logger         logging.Logger
}

// NewPreferenceService creates a new preference service
func NewPreferenceService(preferenceRepo repository.NotificationPreferenceRepository, logger logging.Logger) *PreferenceService {
	return &PreferenceService{
		preferenceRepo: preferenceRepo,
		logger:         logger,
	}
}

// GetUserPreferences retrieves all notification preferences for a user
func (s *PreferenceService) GetUserPreferences(ctx context.Context, userID uuid.UUID) ([]*models.NotificationPreference, error) {
	preferences, err := s.preferenceRepo.GetByUserID(ctx, userID)
	if err != nil {
		s.logger.Error(logging.Postgres, logging.Select, "Failed to get user preferences", map[logging.ExtraKey]interface{}{
			"error":  err.Error(),
			"userId": userID,
		})
		return nil, fmt.Errorf("failed to get user preferences: %w", err)
	}

	// If no preferences found, return defaults
	if len(preferences) == 0 {
		s.logger.Debug(logging.Postgres, logging.Select, "No preferences found, returning defaults", map[logging.ExtraKey]interface{}{
			"userId": userID,
		})
		return s.getDefaultPreferences(userID), nil
	}

	return preferences, nil
}

// UpdatePreference updates or creates a user's notification preference
func (s *PreferenceService) UpdatePreference(ctx context.Context, preference *models.NotificationPreference) error {
	// Validate preference
	if err := s.validatePreference(preference); err != nil {
		s.logger.Error(logging.Validation, logging.Update, "Preference validation failed", map[logging.ExtraKey]interface{}{
			"error":  err.Error(),
			"userId": preference.UserID,
			"type":   preference.Type,
		})
		return err
	}

	// Check if preference exists
	existing, err := s.preferenceRepo.GetByUserIDAndType(ctx, preference.UserID, preference.Type)
	if err != nil {
		s.logger.Error(logging.Postgres, logging.Select, "Failed to check existing preference", map[logging.ExtraKey]interface{}{
			"error":  err.Error(),
			"userId": preference.UserID,
			"type":   preference.Type,
		})
		return fmt.Errorf("failed to check existing preference: %w", err)
	}

	if existing != nil {
		// Update existing preference
		existing.IsEnabled = preference.IsEnabled
		existing.AllowInstant = preference.AllowInstant
		existing.AllowDigest = preference.AllowDigest
		existing.DigestFrequency = preference.DigestFrequency
		existing.QuietHours = preference.QuietHours
		existing.CategorySettings = preference.CategorySettings

		if err := s.preferenceRepo.Update(ctx, existing); err != nil {
			s.logger.Error(logging.Postgres, logging.Update, "Failed to update preference", map[logging.ExtraKey]interface{}{
				"error":  err.Error(),
				"userId": preference.UserID,
				"type":   preference.Type,
			})
			return fmt.Errorf("failed to update preference: %w", err)
		}

		s.logger.Info(logging.General, logging.Update, "Preference updated successfully", map[logging.ExtraKey]interface{}{
			"userId": preference.UserID,
			"type":   preference.Type,
		})
	} else {
		// Create new preference
		if preference.ID == uuid.Nil {
			preference.ID = uuid.New()
		}

		if err := s.preferenceRepo.Create(ctx, preference); err != nil {
			s.logger.Error(logging.Postgres, logging.Insert, "Failed to create preference", map[logging.ExtraKey]interface{}{
				"error":  err.Error(),
				"userId": preference.UserID,
				"type":   preference.Type,
			})
			return fmt.Errorf("failed to create preference: %w", err)
		}

		s.logger.Info(logging.General, logging.Insert, "Preference created successfully", map[logging.ExtraKey]interface{}{
			"userId": preference.UserID,
			"type":   preference.Type,
		})
	}

	return nil
}

// BatchUpdatePreferences updates multiple preferences for a user
func (s *PreferenceService) BatchUpdatePreferences(ctx context.Context, userID uuid.UUID, preferences []*models.NotificationPreference) error {
	for _, pref := range preferences {
		pref.UserID = userID
		if err := s.UpdatePreference(ctx, pref); err != nil {
			return err
		}
	}

	return nil
}

// ResetToDefaults resets user preferences to system defaults
func (s *PreferenceService) ResetToDefaults(ctx context.Context, userID uuid.UUID) error {
	// Delete all existing preferences
	if err := s.preferenceRepo.DeleteAllByUserID(ctx, userID); err != nil {
		s.logger.Error(logging.Postgres, logging.Delete, "Failed to delete user preferences", map[logging.ExtraKey]interface{}{
			"error":  err.Error(),
			"userId": userID,
		})
		return fmt.Errorf("failed to reset preferences: %w", err)
	}

	s.logger.Info(logging.General, logging.Delete, "User preferences reset to defaults", map[logging.ExtraKey]interface{}{
		"userId": userID,
	})

	return nil
}

// validatePreference validates preference fields
func (s *PreferenceService) validatePreference(preference *models.NotificationPreference) error {
	if preference == nil {
		return fmt.Errorf("preference cannot be nil")
	}

	if preference.UserID == uuid.Nil {
		return fmt.Errorf("user ID is required")
	}

	if preference.Type == "" {
		return fmt.Errorf("notification type is required")
	}

	// Warn if all disabled, but don't error
	if !preference.IsEnabled {
		s.logger.Warn(logging.Validation, logging.Update, "Notification disabled for user", map[logging.ExtraKey]interface{}{
			"userId": preference.UserID,
			"type":   preference.Type,
		})
	}

	return nil
}

// getDefaultPreferences returns default notification preferences
func (s *PreferenceService) getDefaultPreferences(userID uuid.UUID) []*models.NotificationPreference {
	types := []models.NotificationType{
		models.NotificationTypeEmail,
		models.NotificationTypeSMS,
		models.NotificationTypePush,
		models.NotificationTypeInApp,
	}
	defaults := make([]*models.NotificationPreference, len(types))

	for i, notifType := range types {
		defaults[i] = &models.NotificationPreference{
			ID:              uuid.New(),
			UserID:          userID,
			Type:            notifType,
			IsEnabled:       true,
			AllowInstant:    true,
			AllowDigest:     notifType != models.NotificationTypeSMS, // SMS only instant
			DigestFrequency: "daily",
		}
	}

	return defaults
}
