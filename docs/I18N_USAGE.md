# I18n and Enhanced Error Handling - Usage Examples

This document provides practical examples of using i18n and enhanced error handling in the notifier service.

## Setup

1. **Environment Configuration**

Create or update `.env` file:
```env
APP_ENV=development
# or
APP_ENV=production
```

2. **Import Required Packages**

```go
import (
    "github.com/gofiber/fiber/v2"
    "github.com/minisource/go-common/common"
    "github.com/minisource/go-common/http/helper"
    "github.com/minisource/go-common/i18n"
    "github.com/minisource/go-common/logging"
    "github.com/minisource/go-common/service_errors"
)
```

## Example 1: Simple I18n Response

```go
func (h *Handler) CreateNotification(c *fiber.Ctx) error {
    // ... create logic ...
    
    // Success response with translated message
    return c.Status(fiber.StatusCreated).JSON(
        helper.GenerateI18nResponse(
            c,                                      // Context for language detection
            map[string]interface{}{"id": id},       // Result data
            true,                                   // Success flag
            0,                                      // Result code
            "notifications.notification_created",   // Translation key
        ),
    )
}
```

**English Response (`?lang=en` or `Accept-Language: en`)**:
```json
{
  "result": {"id": "123"},
  "success": true,
  "resultCode": 0,
  "message": "Notification created successfully"
}
```

**Persian Response (`?lang=fa` or `Accept-Language: fa`)**:
```json
{
  "result": {"id": "123"},
  "success": true,
  "resultCode": 0,
  "message": "اعلان با موفقیت ایجاد شد"
}
```

## Example 2: Error Response with Environment-Aware Details

```go
func (h *Handler) GetNotification(c *fiber.Ctx) error {
    id, err := uuid.Parse(c.Params("id"))
    if err != nil {
        // Create service error
        serviceErr := service_errors.NewServiceError(
            "validation_error",
            "Invalid notification ID",
            err.Error(),
        ).WithDetails(map[string]interface{}{
            "provided_id": c.Params("id"),
        })
        
        // Return environment-aware error response
        return c.Status(fiber.StatusBadRequest).JSON(
            helper.GenerateBaseResponseWithServiceError(
                c,                          // Context
                nil,                        // Result
                false,                      // Success
                helper.ValidationError,     // Result code
                serviceErr,                 // Service error
                common.IsDevelopment(),     // Show details in dev mode
            ),
        )
    }
    
    // ... rest of handler ...
}
```

**Development Mode Response**:
```json
{
  "result": null,
  "success": false,
  "resultCode": 40001,
  "message": "Validation failed",
  "error": {
    "message": "Invalid notification ID",
    "code": "validation_error",
    "technical_message": "invalid UUID format",
    "error": "invalid UUID length: 3",
    "details": {
      "provided_id": "abc"
    }
  }
}
```

**Production Mode Response**:
```json
{
  "result": null,
  "success": false,
  "resultCode": 40001,
  "message": "Validation failed"
}
```

## Example 3: Debug Logging in Development

```go
func (h *Handler) ProcessNotification(c *fiber.Ctx) error {
    // Create debug context
    debugCtx := &logging.DebugContext{
        RequestID: c.Get("X-Request-ID"),
        UserID:    c.Get("X-User-ID"),
        Method:    c.Method(),
        Path:      c.Path(),
    }
    
    // Only log in development
    if common.IsDevelopment() {
        h.logger.Debug(
            logging.Api,
            logging.Request,
            "Processing notification request",
            debugCtx.WithExtra("notification_type", req.Type).ToMap(),
        )
    }
    
    // Process notification
    err := h.service.Process(c.Context(), req)
    
    if err != nil {
        // Log error with details in development
        if common.IsDevelopment() {
            h.logger.Error(
                logging.Api,
                logging.Internal,
                "Failed to process notification",
                debugCtx.WithExtra("error", err.Error()).ToMap(),
            )
        }
        
        return c.Status(500).JSON(
            helper.GenerateI18nResponse(
                c,
                nil,
                false,
                helper.InternalError,
                "errors.internal_error",
            ),
        )
    }
    
    return c.JSON(helper.GenerateI18nResponse(
        c,
        nil,
        true,
        0,
        "notifications.notification_sent",
    ))
}
```

## Example 4: Custom Service Errors

Create custom errors in `internal/service/errors.go`:

```go
package service

import "github.com/minisource/go-common/service_errors"

const (
    NotificationNotFound = "notification_not_found"
    InvalidRecipient     = "recipient_required"
)

func NewNotificationNotFoundError(id string) *service_errors.ServiceError {
    return service_errors.NewServiceError(
        NotificationNotFound,
        "Notification not found",
        fmt.Sprintf("No notification found with ID: %s", id),
    ).WithDetails(map[string]interface{}{
        "notification_id": id,
    })
}

func NewInvalidRecipientError() *service_errors.ServiceError {
    return service_errors.NewServiceError(
        InvalidRecipient,
        "At least one recipient is required",
        "RecipientEmail, RecipientPhone, or RecipientID must be provided",
    )
}
```

Usage in handler:

```go
notification, err := h.service.GetByID(c.Context(), id)
if err != nil {
    if errors.Is(err, gorm.ErrRecordNotFound) {
        serviceErr := service.NewNotificationNotFoundError(id.String())
        return c.Status(404).JSON(
            helper.GenerateBaseResponseWithServiceError(
                c,
                nil,
                false,
                helper.NotFoundError,
                serviceErr,
                common.IsDevelopment(),
            ),
        )
    }
    // ... handle other errors ...
}
```

## Example 5: Validation with I18n

```go
func (h *Handler) ValidateAndCreate(c *fiber.Ctx) error {
    req := new(CreateNotificationRequest)
    if err := c.BodyParser(req); err != nil {
        return c.Status(400).JSON(
            helper.GenerateBaseResponseWithValidationError(
                nil,
                false,
                helper.ValidationError,
                err,
            ),
        )
    }
    
    // Custom validation with i18n
    if req.RecipientEmail == "" && req.RecipientPhone == "" {
        return c.Status(400).JSON(
            helper.GenerateI18nResponse(
                c,
                nil,
                false,
                helper.ValidationError,
                "notifications.recipient_required",
            ),
        )
    }
    
    if req.Priority < 1 || req.Priority > 4 {
        return c.Status(400).JSON(
            helper.GenerateI18nResponse(
                c,
                nil,
                false,
                helper.ValidationError,
                "notifications.invalid_priority",
            ),
        )
    }
    
    // ... create logic ...
}
```

## Example 6: Testing with Different Languages

```bash
# Test with English (default)
curl http://localhost:8080/api/v1/notifications/invalid-id

# Test with Persian
curl -H "Accept-Language: fa" http://localhost:8080/api/v1/notifications/invalid-id

# Or use query parameter
curl http://localhost:8080/api/v1/notifications/invalid-id?lang=fa
```

## Example 7: Environment-Specific Behavior

```go
func (h *Handler) SomeHandler(c *fiber.Ctx) error {
    // Different behavior based on environment
    switch common.GetEnvironment() {
    case common.EnvDevelopment:
        // Development-specific logic
        h.logger.Debug(logging.Api, logging.Internal, "Dev mode enabled", nil)
        
    case common.EnvStaging:
        // Staging-specific logic
        
    case common.EnvProduction:
        // Production-specific logic
        // Disable verbose logging, enable monitoring, etc.
    }
    
    // Conditional detailed errors
    if common.ShouldShowDetailedErrors() {
        // Add extra debug information
    }
    
    return nil
}
```

## Example 8: Adding New Translations

1. Add to `common_go/i18n/locales/en.json`:
```json
{
  "notifications": {
    "scheduled_successfully": "Notification scheduled for {{.Time}}"
  }
}
```

2. Add to `common_go/i18n/locales/fa.json`:
```json
{
  "notifications": {
    "scheduled_successfully": "اعلان برای {{.Time}} زمان‌بندی شد"
  }
}
```

3. Use in code:
```go
return c.JSON(
    helper.GenerateI18nResponse(
        c,
        result,
        true,
        0,
        "notifications.scheduled_successfully",
        map[string]interface{}{
            "Time": scheduledTime.Format("2006-01-02 15:04"),
        },
    ),
)
```

## Testing Checklist

- [ ] Test endpoints with `?lang=en`
- [ ] Test endpoints with `?lang=fa`
- [ ] Test with `Accept-Language` header
- [ ] Test in development mode (`APP_ENV=development`)
- [ ] Test in production mode (`APP_ENV=production`)
- [ ] Verify error details are hidden in production
- [ ] Verify debug logs only appear in development
- [ ] Test all error scenarios
- [ ] Verify all translations are present

## Common Patterns

### Pattern 1: Standard CRUD Response

```go
// Create
return c.Status(201).JSON(
    helper.GenerateI18nResponse(c, result, true, 0, "notifications.notification_created"))

// Update  
return c.Status(200).JSON(
    helper.GenerateI18nResponse(c, result, true, 0, "notifications.template_updated"))

// Delete
return c.Status(200).JSON(
    helper.GenerateI18nResponse(c, nil, true, 0, "notifications.template_deleted"))
```

### Pattern 2: Error Handling

```go
if err != nil {
    if common.IsDevelopment() {
        h.logger.Error(logging.Api, logging.Internal, "Operation failed", 
            map[logging.ExtraKey]interface{}{"error": err.Error()})
    }
    
    return c.Status(500).JSON(
        helper.GenerateI18nResponse(c, nil, false, helper.InternalError, "errors.internal_error"))
}
```

### Pattern 3: Validation Errors

```go
if !isValid {
    return c.Status(400).JSON(
        helper.GenerateI18nResponse(c, nil, false, helper.ValidationError, "validation.invalid_format"))
}
```

## Available Translation Keys

### Errors
- `errors.unexpected_error`
- `errors.record_not_found`
- `errors.permission_denied`
- `errors.validation_error`
- `errors.internal_error`
- And more...

### Notifications
- `notifications.notification_created`
- `notifications.notification_sent`
- `notifications.notification_failed`
- `notifications.template_created`
- And more...

### Validation
- `validation.required`
- `validation.invalid_email`
- `validation.invalid_phone`
- `validation.invalid_uuid`

### Common
- `common.success`
- `common.failed`
- `common.save`
- `common.delete`

See full list in:
- `common_go/i18n/locales/en.json`
- `common_go/i18n/locales/fa.json`
