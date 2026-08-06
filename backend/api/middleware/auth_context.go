package middleware

import (
	"github.com/gofiber/fiber/v2"
	commiddleware "github.com/minisource/go-common/http/middleware"
	"github.com/minisource/go-common/response"
)

// AuthContext represents the authenticated user/service context.
type AuthContext struct {
	UserID          string
	Email           string
	Roles           []string
	Permissions     []string
	TenantID        string
	IsSuperAdmin    bool
	SessionID       string
	TokenType       string
	Issuer          string
	Audience        []string
	IsAuthenticated bool
	IsAdmin         bool
	IsService       bool
}

// GetAuthContext extracts the full auth context from the request.
// It first checks the new authContext local (set by RealAuthMiddleware),
// then falls back to go-common's Fiber locals for backward compatibility.
func GetAuthContext(c *fiber.Ctx) *AuthContext {
	if ctx, ok := c.Locals("authContext").(*AuthContext); ok {
		return ctx
	}

	ctx := &AuthContext{
		UserID:          commiddleware.GetUserIDFromContext(c),
		TenantID:        commiddleware.GetTenantID(c),
		Roles:           commiddleware.GetRolesFromContext(c),
		Permissions:     commiddleware.GetPermissionsFromContext(c),
		Email:           commiddleware.GetEmailFromContext(c),
		SessionID:       commiddleware.GetSessionIDFromContext(c),
		IsAuthenticated: commiddleware.GetUserIDFromContext(c) != "",
		IsService:       commiddleware.GetServiceClaimsFromContext(c) != nil,
	}

	ctx.IsAdmin = isAdminRole(ctx.Roles, ctx.Permissions, ctx.IsSuperAdmin)
	return ctx
}

// GetCurrentUserID extracts the current user ID.
func GetCurrentUserID(c *fiber.Ctx) string {
	if ctx, ok := c.Locals("authContext").(*AuthContext); ok {
		return ctx.UserID
	}
	return commiddleware.GetUserIDFromContext(c)
}

// IsAdmin checks admin status using go-common's helper.
func IsAdmin(c *fiber.Ctx) bool {
	return commiddleware.IsAdmin(c)
}

// IsService checks service token status.
func IsService(c *fiber.Ctx) bool {
	return commiddleware.GetServiceClaimsFromContext(c) != nil
}

// HasPermission checks a specific permission using go-common's helper.
func HasPermission(c *fiber.Ctx, permission string) bool {
	return commiddleware.HasPermission(c, permission)
}

// HasAnyPermission checks any of the given permissions.
func HasAnyPermission(c *fiber.Ctx, permissions ...string) bool {
	return commiddleware.HasAnyPermission(c, permissions...)
}

// HasRole checks a specific role using go-common's helper.
func HasRole(c *fiber.Ctx, role string) bool {
	return commiddleware.HasRole(c, role)
}

// HasAnyRole checks any of the given roles.
func HasAnyRole(c *fiber.Ctx, roles ...string) bool {
	return commiddleware.HasAnyRole(c, roles...)
}

// ---------------------------------------------------------------------------
// Middleware creators (wrapper over go-common with consistent error format)
// ---------------------------------------------------------------------------

// RequireAuthenticated requires any valid authentication.
func RequireAuthenticated() fiber.Handler {
	return commiddleware.RequireAuthenticated()
}

// RequireAdmin requires admin privileges.
func RequireAdmin() fiber.Handler {
	return commiddleware.RequireAdmin()
}

// RequireRole requires a specific role.
func RequireRole(role string) fiber.Handler {
	return commiddleware.RequireRole(role)
}

// RequireAnyRole requires any of the given roles.
func RequireAnyRole(roles ...string) fiber.Handler {
	return commiddleware.RequireAnyRole(roles...)
}

// RequirePermission requires a specific permission.
func RequirePermission(permission string) fiber.Handler {
	return commiddleware.RequirePermission(permission)
}

// RequireAnyPermission requires any of the given permissions.
func RequireAnyPermission(permissions ...string) fiber.Handler {
	return commiddleware.RequireAnyPermission(permissions...)
}

// RequireAdminOrPermission requires admin OR a specific permission.
func RequireAdminOrPermission(permission string) fiber.Handler {
	return commiddleware.RequireAdminOrPermission(permission)
}

// Legacy helpers (backward compatible)

// RequireSelfOrAdmin checks self-or-admin access.
func RequireSelfOrAdmin(targetUserID string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		currentUserID := GetCurrentUserID(c)
		if currentUserID == "" && !IsService(c) {
			return response.Forbidden(c, "Authentication required")
		}
		if IsService(c) || IsAdmin(c) {
			return c.Next()
		}
		if currentUserID == targetUserID {
			return c.Next()
		}
		return response.Forbidden(c, "You can only access your own resources")
	}
}

// RequireSelfOrAdminFromParam checks self-or-admin using a path param.
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

// RequireService requires service authentication.
func RequireService() fiber.Handler {
	return func(c *fiber.Ctx) error {
		if !IsService(c) {
			return response.Forbidden(c, "Service authentication required")
		}
		return c.Next()
	}
}

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

func isAdminRole(roles []string, permissions []string, isSuperAdmin bool) bool {
	if isSuperAdmin {
		return true
	}
	for _, role := range roles {
		if role == "admin" || role == "super_admin" || role == "system_admin" {
			return true
		}
	}
	for _, perm := range permissions {
		if perm == "notifier:admin" {
			return true
		}
	}
	return false
}
