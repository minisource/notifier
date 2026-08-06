package middleware

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/minisource/go-common/http/middleware"
	"github.com/minisource/go-common/response"
	gosdkAuth "github.com/minisource/go-sdk/auth"
)

// ---------------------------------------------------------------------------
// Validation modes
// ---------------------------------------------------------------------------

const (
	ValidationModeJWKS         = "jwks"
	ValidationModeIntrospection = "introspection"
	ValidationModeHS256        = "hs256"
)

// ---------------------------------------------------------------------------
// RealAuthConfig
// ---------------------------------------------------------------------------

// RealAuthConfig holds configuration for the real auth middleware.
type RealAuthConfig struct {
	ValidationMode string

	// JWKS
	JWKSURL              string
	JWKSCacheTTL         time.Duration
	JWKSRefreshOnKidMiss bool

	// Introspection
	IntrospectionURL    string
	IntrospectionTimeout time.Duration

	// HS256 (dev fallback)
	HS256Secret string

	// Validation rules
	Issuer   string
	Audience string

	// Behaviour
	SkipPaths   []string
	RequireAuth bool

	// App environment (used for validation warnings)
	AppEnv string
}

// RealAuthMiddleware creates an auth middleware that validates JWT tokens via
// JWKS, introspection, or HS256 — delegating the actual validation to the
// shared go-sdk/auth package.
func RealAuthMiddleware(cfg RealAuthConfig) fiber.Handler {
	if err := validateAuthConfig(cfg); err != nil {
		log.Fatalf("[AUTH CONFIG ERROR] %v", err)
	}
	logAuthStartup(cfg)

	validator := createValidator(cfg)

	return func(c *fiber.Ctx) error {
		path := c.Path()
		for _, skip := range cfg.SkipPaths {
			if strings.HasPrefix(path, skip) || path == skip {
				return c.Next()
			}
		}

		authHeader := c.Get("Authorization")
		if authHeader == "" {
			if !cfg.RequireAuth {
				return c.Next()
			}
			return response.New().Status(fiber.StatusUnauthorized).Error("AUTH_REQUIRED", "Authentication required").Send(c)
		}

		if !strings.HasPrefix(authHeader, "Bearer ") {
			return response.New().Status(fiber.StatusUnauthorized).Error("AUTH_REQUIRED", "Invalid authorization header format. Expected: Bearer <token>").Send(c)
		}

		token := authHeader[7:]
		claims, err := validator.ValidateToken(token)
		if err != nil {
			return authErrorResponse(c, err)
		}

		// Store in Fiber locals (compatible with go-common helpers)
		c.Locals("userId", claims.UserID)
		c.Locals("email", claims.Email)
		c.Locals("roles", claims.Roles)
		c.Locals("permissions", claims.Permissions)
		c.Locals("tenantId", claims.TenantID)
		c.Locals("sessionId", claims.SessionID)
		c.Locals("tokenType", claims.TokenType)
		c.Locals("isSuperAdmin", claims.IsSuperAdmin)

		// Also store the full auth context for advanced use
		c.Locals("authContext", &AuthContext{
			UserID:          claims.UserID,
			Email:           claims.Email,
			Roles:           claims.Roles,
			Permissions:     claims.Permissions,
			TenantID:        claims.TenantID,
			IsSuperAdmin:    claims.IsSuperAdmin,
			SessionID:       claims.SessionID,
			TokenType:       claims.TokenType,
			Issuer:          claims.Issuer,
			Audience:        claims.Audience,
			IsAuthenticated: true,
			IsAdmin:         middleware.IsAdmin(c),
		})

		return c.Next()
	}
}

// ---------------------------------------------------------------------------
// Config validation
// ---------------------------------------------------------------------------

func validateAuthConfig(cfg RealAuthConfig) error {
	switch cfg.ValidationMode {
	case ValidationModeJWKS:
		if cfg.JWKSURL == "" {
			return fmt.Errorf("AUTH_VALIDATION_MODE=jwks requires AUTH_JWKS_URL to be set")
		}
	case ValidationModeIntrospection:
		if cfg.IntrospectionURL == "" {
			return fmt.Errorf("AUTH_VALIDATION_MODE=introspection requires AUTH_INTROSPECTION_URL to be set")
		}
	case ValidationModeHS256:
		if cfg.HS256Secret == "" {
			return fmt.Errorf("AUTH_VALIDATION_MODE=hs256 requires AUTH_JWT_SECRET to be set")
		}
		if cfg.AppEnv == "production" {
			log.Println("[WARNING] HS256 in production is not recommended.")
		}
	default:
		return fmt.Errorf("invalid AUTH_VALIDATION_MODE: %s", cfg.ValidationMode)
	}

	if cfg.Issuer == "" {
		return fmt.Errorf("AUTH_ISSUER must be configured")
	}
	return nil
}

func logAuthStartup(cfg RealAuthConfig) {
	log.Printf("Auth validation mode: %s", cfg.ValidationMode)
	log.Printf("Auth issuer: %s", cfg.Issuer)
	log.Printf("Auth audience: %s", cfg.Audience)
	if cfg.ValidationMode == ValidationModeJWKS {
		log.Printf("JWKS URL: %s", cfg.JWKSURL)
	}
	if cfg.ValidationMode == ValidationModeIntrospection {
		log.Printf("Introspection URL: %s", cfg.IntrospectionURL)
	}
}

// ---------------------------------------------------------------------------
// Validator factory
// ---------------------------------------------------------------------------

func createValidator(cfg RealAuthConfig) gosdkAuth.TokenValidator {
	switch cfg.ValidationMode {
	case ValidationModeJWKS:
		return gosdkAuth.NewJWKSValidator(gosdkAuth.JWKSValidatorConfig{
			JWKSURL:       cfg.JWKSURL,
			CacheTTL:      cfg.JWKSCacheTTL,
			RefreshOnMiss: cfg.JWKSRefreshOnKidMiss,
			HTTPTimeout:   5 * time.Second,
			Issuer:        cfg.Issuer,
			Audience:      cfg.Audience,
		})
	case ValidationModeIntrospection:
		return gosdkAuth.NewIntrospectionClient(
			cfg.IntrospectionURL,
			cfg.IntrospectionTimeout,
			cfg.Issuer,
			cfg.Audience,
		)
	case ValidationModeHS256:
		return gosdkAuth.NewHS256Validator(cfg.HS256Secret, cfg.Issuer, cfg.Audience)
	default:
		return gosdkAuth.NewHS256Validator(cfg.HS256Secret, cfg.Issuer, cfg.Audience)
	}
}

// ---------------------------------------------------------------------------
// Error responses
// ---------------------------------------------------------------------------

func authErrorResponse(c *fiber.Ctx, err error) error {
	msg := err.Error()
	switch msg {
	case "AUTH_REQUIRED":
		return response.New().Status(fiber.StatusUnauthorized).Error("AUTH_REQUIRED", "Authentication required").Send(c)
	case "INVALID_TOKEN":
		return response.New().Status(fiber.StatusUnauthorized).Error("INVALID_TOKEN", "Invalid or malformed token").Send(c)
	case "TOKEN_EXPIRED":
		return response.New().Status(fiber.StatusUnauthorized).Error("TOKEN_EXPIRED", "Access token has expired").Send(c)
	case "INVALID_ISSUER":
		return response.New().Status(fiber.StatusUnauthorized).Error("INVALID_ISSUER", "Token issuer is not valid").Send(c)
	case "INVALID_AUDIENCE":
		return response.New().Status(fiber.StatusUnauthorized).Error("INVALID_AUDIENCE", "Token audience is not valid").Send(c)
	case "AUTH_SERVICE_UNAVAILABLE":
		return response.New().Status(fiber.StatusServiceUnavailable).Error("AUTH_SERVICE_UNAVAILABLE", "Authentication service is temporarily unavailable").Send(c)
	default:
		return response.New().Status(fiber.StatusUnauthorized).Error("INVALID_TOKEN", "Invalid token").Send(c)
	}
}
