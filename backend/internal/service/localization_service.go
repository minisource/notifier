package service

import (
	"context"
	"fmt"

	"github.com/minisource/go-common/i18n"
	"github.com/minisource/go-common/logging"
	"github.com/minisource/notifier/internal/models"
	"github.com/minisource/notifier/internal/repository"
)

// LocalizationService handles locale-aware template selection.
// It selects the best template for a given locale (with fallback to 'en').
type LocalizationService struct {
	templateRepo repository.NotificationTemplateRepository
	logger       logging.Logger
}

// NewLocalizationService creates a new localization service
func NewLocalizationService(
	templateRepo repository.NotificationTemplateRepository,
	logger logging.Logger,
) *LocalizationService {
	return &LocalizationService{
		templateRepo: templateRepo,
		logger:       logger,
	}
}

// GetLocalizedTemplate retrieves the best template for a key and locale,
// falling back to 'en' if the requested locale is not available.
func (s *LocalizationService) GetLocalizedTemplate(ctx context.Context, key string, locale string) (*models.NotificationTemplate, error) {
	if locale == "" {
		locale = "en"
	}

	template, err := s.templateRepo.GetByKeyAndLocale(ctx, key, locale)
	if err != nil {
		s.logger.Debug(logging.General, logging.Select, "Template not found for locale, falling back to 'en'", map[logging.ExtraKey]interface{}{
			"key":    key,
			"locale": locale,
		})
		template, err = s.templateRepo.GetByKeyAndLocale(ctx, key, "en")
		if err != nil {
			return nil, fmt.Errorf("template not found for key=%s in any locale", key)
		}
		return template, nil
	}
	return template, nil
}

// TranslateKey translates a single i18n key using the context's locale.
func (s *LocalizationService) TranslateKey(ctx context.Context, key string) string {
	return i18n.T(ctx, key)
}

// GetSupportedLocales returns the list of supported locales
func (s *LocalizationService) GetSupportedLocales() []string {
	return []string{"en", "fa", "ar"}
}

// IsLocaleSupported checks if a locale is in the supported list
func (s *LocalizationService) IsLocaleSupported(locale string) bool {
	for _, l := range s.GetSupportedLocales() {
		if l == locale {
			return true
		}
	}
	return false
}
