package middleware

import (
	"github.com/gofiber/fiber/v2"
	"github.com/minisource/go-common/http/middleware"
	"github.com/minisource/go-common/response"
)

// AuthContext represents the authenticated user/service context extracted from JWT or service token.
type AuthContext struct {
	UserID          string
	TenantID        string
	Roles           []string
	Permissions     []string
	Scopes          []string
	IsAuthenticated bool
	IsAdmin         bool
	IsService       bool
}

// GetAuthContext extracts authentication context from the Fiber context.
// It reads values set by go-common's AuthMiddleware or ServiceAuthMiddleware.
func GetAuthContext(c *fiber.Ctx) *AuthContext {
	ctx := &AuthContext{
		UserID:          middleware.GetUserIDFromContext(c),
		TenantID:        middleware.GetTenantID(c),
		Roles:           middleware.GetRolesFromContext(c),
		Permissions:     middleware.GetPermissionsFromContext(c),
		Scopes:          middleware.GetScopesFromContext(c),
		IsAuthenticated: false,
		IsAdmin:         false,
		IsService:       false,
	}

	// Check if authenticated via user JWT
	if ctx.UserID != "" {
		ctx.IsAuthenticated = true
	}

	// Check if authenticated via service token (service auth stores claims under "service" key)
	serviceClaims := middleware.GetServiceClaimsFromContext(c)
	if serviceClaims != nil {
		ctx.IsAuthenticated = true
		ctx.IsService = true
	}

	// Determine admin status from roles
	for _, role := range ctx.Roles {
		if role == "admin" || role == "super_admin" {
			ctx.IsAdmin = true
			break
		}
	}

	return ctx
}

// GetCurrentUserID extracts the current user ID from JWT claims.
// Returns empty string if not authenticated or not a user JWT.
func GetCurrentUserID(c *fiber.Ctx) string {
	return middleware.GetUserIDFromContext(c)
}

// IsAdmin checks if the current context has admin or super_admin role.
func IsAdmin(c *fiber.Ctx) bool {
	for _, role := range middleware.GetRolesFromContext(c) {
		if role == "admin" || role == "super_admin" {
			return true
		}
	}
	return false
}

// IsService checks if the current context is a service-to-service token.
func IsService(c *fiber.Ctx) bool {
	return middleware.GetServiceClaimsFromContext(c) != nil
}

// RequireSelfOrAdmin returns a middleware that checks:
// - If the requesting user is the same as targetUserID, or
// - If the requesting user has admin/super_admin role.
// Returns 403 Forbidden if neither condition is met.
func RequireSelfOrAdmin(targetUserID string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		currentUserID := GetCurrentUserID(c)
		if currentUserID == "" && !IsService(c) {
			return response.Forbidden(c, "Authentication required")
		}

		// Service tokens are allowed
		if IsService(c) {
			return c.Next()
		}

		// Admin can access any user's resources
		if IsAdmin(c) {
			return c.Next()
		}

		// Self-access: current user must match target
		if currentUserID == targetUserID {
			return c.Next()
		}

		return response.Forbidden(c, "You can only access your own resources")
	}
}

// RequireAdmin returns a middleware that requires admin or super_admin role.
// Returns 403 Forbidden if the user does not have admin role.
func RequireAdmin() fiber.Handler {
	return func(c *fiber.Ctx) error {
		if !IsAdmin(c) {
			return response.Forbidden(c, "Admin access required")
		}
		return c.Next()
	}
}

// RequireService returns a middleware that requires service authentication.
// Returns 403 Forbidden if not a service token.
func RequireService() fiber.Handler {
	return func(c *fiber.Ctx) error {
		if !IsService(c) {
			return response.Forbidden(c, "Service authentication required")
		}
		return c.Next()
	}
}

// RequireSelfOrAdminOrService returns a middleware that allows:
// - Self-access (current user matches target)
// - Admin access
// - Service access
func RequireSelfOrAdminOrService(targetUserID string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		if IsAdmin(c) || IsService(c) {
			return c.Next()
		}
		currentUserID := GetCurrentUserID(c)
		if currentUserID == targetUserID {
			return c.Next()
		}
		return response.Forbidden(c, "Access denied")
	}
}

// RequireSelfOrAdminFromParam returns a middleware that extracts the user ID
// from the named path parameter and checks self-or-admin access.
func RequireSelfOrAdminFromParam(paramName string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		targetUserID := c.Params(paramName)
		if targetUserID == "" {
			return response.Forbidden(c, "Access denied")
		}
		if IsAdmin(c) || IsService(c) {
			return c.Next()
		}
		currentUserID := GetCurrentUserID(c)
		if currentUserID == targetUserID {
			return c.Next()
		}
		return response.Forbidden(c, "You can only access your own resources")
	}
}
