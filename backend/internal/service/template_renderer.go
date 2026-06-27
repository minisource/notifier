package service

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/minisource/notifier/internal/models"
)

var (
	// placeholderPattern matches {{variableName}} in template text
	placeholderPattern = regexp.MustCompile(`\{\{([^}]+)\}\}`)
)

// RenderRequest represents the input for rendering a template
type RenderRequest struct {
	Template  *models.NotificationTemplate
	Variables map[string]string
	Locale    string // Target locale: "en", "fa"
}

// RenderResult represents the rendered output
type RenderResult struct {
	Subject      string            `json:"subject"`
	Body         string            `json:"body"`
	UsedVariables []string         `json:"usedVariables"`
	MissingVariables []string      `json:"missingVariables,omitempty"`
}

// RenderTemplate renders a notification template with the given variables.
// Missing variables are substituted with the placeholder key wrapped in brackets
// (e.g., {{code}} becomes "[code]") to avoid displaying raw placeholders.
func RenderTemplate(req *RenderRequest) (*RenderResult, error) {
	if req.Template == nil {
		return nil, fmt.Errorf("template is required")
	}

	if req.Variables == nil {
		req.Variables = make(map[string]string)
	}

	result := &RenderResult{}
	usedVars := make(map[string]bool)

	// Render subject
	if req.Template.Subject != "" {
		result.Subject = replacePlaceholders(req.Template.Subject, req.Variables, usedVars)
	}

	// Render body
	if req.Template.Body != "" {
		result.Body = replacePlaceholders(req.Template.Body, req.Variables, usedVars)
	}

	// Collect used and missing variables
	for v := range usedVars {
		result.UsedVariables = append(result.UsedVariables, v)
	}

	// Check for expected variables that are missing
	expectedVars := req.Template.ParseVariables()
	for _, v := range expectedVars {
		if _, ok := req.Variables[v]; !ok {
			result.MissingVariables = append(result.MissingVariables, v)
		}
	}

	return result, nil
}

// replacePlaceholders replaces all {{key}} placeholders with their values.
// Variables used in the template are tracked in usedVars.
// Missing variables are replaced with "[key]" as a safe fallback.
func replacePlaceholders(text string, variables map[string]string, usedVars map[string]bool) string {
	return placeholderPattern.ReplaceAllStringFunc(text, func(match string) string {
		// Extract the key from {{key}}, trimming whitespace
		key := strings.TrimSpace(match[2 : len(match)-2])

		if key == "" {
			return match
		}

		if value, ok := variables[key]; ok {
			usedVars[key] = true
			return value
		}

		// Missing variable: use safe fallback
		usedVars[key] = false
		return "[" + key + "]"
	})
}

// ValidateTemplateVariables checks that all referenced variables in the template
// have corresponding entries in the variables map. Returns missing variables.
func ValidateTemplateVariables(template *models.NotificationTemplate, variables map[string]string) []string {
	expected := template.ParseVariables()
	if len(expected) == 0 {
		return nil
	}

	var missing []string
	for _, v := range expected {
		if _, ok := variables[v]; !ok {
			missing = append(missing, v)
		}
	}
	return missing
}
