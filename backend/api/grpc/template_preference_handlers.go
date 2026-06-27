package grpc

import (
	"context"
	"encoding/json"

	"github.com/google/uuid"
	"github.com/minisource/go-common/logging"
	pb "github.com/minisource/go-sdk/notifier/proto/notifier/v1"
	"github.com/minisource/notifier/internal/models"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// Template Service Handlers

// CreateTemplate creates a new notification template
func (s *Server) CreateTemplate(ctx context.Context, req *pb.CreateTemplateRequest) (*pb.CreateTemplateResponse, error) {
	s.logger.Debug(logging.General, logging.Api, "CreateTemplate gRPC request", map[logging.ExtraKey]interface{}{
		"name": req.Name,
		"type": req.Type.String(),
	})

	// Convert variables to JSON string
	variablesJSON, _ := json.Marshal(req.Variables)

	template := &models.NotificationTemplate{
		Name:             req.Name,
		Type:             models.NotificationType(req.Type.String()),
		Subject:          req.Subject,
		Body:             req.Body,
		Description:      req.Description,
		Variables:        string(variablesJSON),
		Provider:         req.Provider,
		ProviderTemplate: req.ProviderTemplate,
		IsActive:         true,
	}

	created, err := s.templateSvc.CreateTemplate(ctx, template)
	if err != nil {
		s.logger.Error(logging.General, logging.Api, "Failed to create template", map[logging.ExtraKey]interface{}{
			"error": err.Error(),
		})
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &pb.CreateTemplateResponse{
		TemplateId: created.ID.String(),
		Success:    true,
		Message:    "Template created successfully",
	}, nil
}

// GetTemplate retrieves a template by ID
func (s *Server) GetTemplate(ctx context.Context, req *pb.GetTemplateRequest) (*pb.GetTemplateResponse, error) {
	s.logger.Debug(logging.General, logging.Api, "GetTemplate gRPC request", map[logging.ExtraKey]interface{}{
		"templateId": req.TemplateId,
	})

	id, err := uuid.Parse(req.TemplateId)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid template ID")
	}

	template, err := s.templateSvc.GetTemplate(ctx, id)
	if err != nil {
		s.logger.Error(logging.General, logging.Api, "Failed to get template", map[logging.ExtraKey]interface{}{
			"error":      err.Error(),
			"templateId": req.TemplateId,
		})
		return nil, status.Error(codes.NotFound, err.Error())
	}

	// Parse variables JSON
	var variables []string
	if template.Variables != "" {
		json.Unmarshal([]byte(template.Variables), &variables)
	}

	pbTemplate := &pb.Template{
		Id:               template.ID.String(),
		Name:             template.Name,
		Type:             pb.NotificationType(pb.NotificationType_value[string(template.Type)]),
		Subject:          template.Subject,
		Body:             template.Body,
		Description:      template.Description,
		Variables:        variables,
		Provider:         template.Provider,
		ProviderTemplate: template.ProviderTemplate,
		IsActive:         template.IsActive,
		CreatedAt:        timestamppb.New(template.CreatedAt),
		UpdatedAt:        timestamppb.New(template.UpdatedAt),
	}

	return &pb.GetTemplateResponse{
		Template: pbTemplate,
	}, nil
}

// UpdateTemplate updates an existing template
func (s *Server) UpdateTemplate(ctx context.Context, req *pb.UpdateTemplateRequest) (*pb.UpdateTemplateResponse, error) {
	s.logger.Debug(logging.General, logging.Api, "UpdateTemplate gRPC request", map[logging.ExtraKey]interface{}{
		"templateId": req.TemplateId,
	})

	id, err := uuid.Parse(req.TemplateId)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid template ID")
	}

	// Convert variables to JSON string
	variablesJSON, _ := json.Marshal(req.Variables)

	updates := &models.NotificationTemplate{
		Name:             req.Name,
		Type:             models.NotificationType(req.Type.String()),
		Subject:          req.Subject,
		Body:             req.Body,
		Description:      req.Description,
		Variables:        string(variablesJSON),
		Provider:         req.Provider,
		ProviderTemplate: req.ProviderTemplate,
		IsActive:         true,
	}

	_, err = s.templateSvc.UpdateTemplate(ctx, id, updates)
	if err != nil {
		s.logger.Error(logging.General, logging.Api, "Failed to update template", map[logging.ExtraKey]interface{}{
			"error":      err.Error(),
			"templateId": req.TemplateId,
		})
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &pb.UpdateTemplateResponse{
		Success: true,
		Message: "Template updated successfully",
	}, nil
}

// DeleteTemplate deletes a template
func (s *Server) DeleteTemplate(ctx context.Context, req *pb.DeleteTemplateRequest) (*pb.DeleteTemplateResponse, error) {
	s.logger.Debug(logging.General, logging.Api, "DeleteTemplate gRPC request", map[logging.ExtraKey]interface{}{
		"templateId": req.TemplateId,
	})

	id, err := uuid.Parse(req.TemplateId)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid template ID")
	}

	err = s.templateSvc.DeleteTemplate(ctx, id)
	if err != nil {
		s.logger.Error(logging.General, logging.Api, "Failed to delete template", map[logging.ExtraKey]interface{}{
			"error":      err.Error(),
			"templateId": req.TemplateId,
		})
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &pb.DeleteTemplateResponse{
		Success: true,
		Message: "Template deleted successfully",
	}, nil
}

// GetAllTemplates lists all templates
func (s *Server) GetAllTemplates(ctx context.Context, req *pb.GetAllTemplatesRequest) (*pb.GetAllTemplatesResponse, error) {
	s.logger.Debug(logging.General, logging.Api, "GetAllTemplates gRPC request", map[logging.ExtraKey]interface{}{
		"page":     req.Page,
		"pageSize": req.PageSize,
	})

	page := int(req.Page)
	pageSize := int(req.PageSize)
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 20
	}

	templates, total, err := s.templateSvc.ListTemplates(ctx, page, pageSize)
	if err != nil {
		s.logger.Error(logging.General, logging.Api, "Failed to list templates", map[logging.ExtraKey]interface{}{
			"error": err.Error(),
		})
		return nil, status.Error(codes.Internal, err.Error())
	}

	pbTemplates := make([]*pb.Template, len(templates))
	for i, t := range templates {
		var variables []string
		if t.Variables != "" {
			json.Unmarshal([]byte(t.Variables), &variables)
		}

		pbTemplates[i] = &pb.Template{
			Id:               t.ID.String(),
			Name:             t.Name,
			Type:             pb.NotificationType(pb.NotificationType_value[string(t.Type)]),
			Subject:          t.Subject,
			Body:             t.Body,
			Description:      t.Description,
			Variables:        variables,
			Provider:         t.Provider,
			ProviderTemplate: t.ProviderTemplate,
			IsActive:         t.IsActive,
			CreatedAt:        timestamppb.New(t.CreatedAt),
			UpdatedAt:        timestamppb.New(t.UpdatedAt),
		}
	}

	return &pb.GetAllTemplatesResponse{
		Templates: pbTemplates,
		Total:     total,
		Page:      int32(page),
		PageSize:  int32(pageSize),
	}, nil
}

// Preference Service Handlers

// GetUserPreferences retrieves user notification preferences
func (s *Server) GetUserPreferences(ctx context.Context, req *pb.GetUserPreferencesRequest) (*pb.GetUserPreferencesResponse, error) {
	s.logger.Debug(logging.General, logging.Api, "GetUserPreferences gRPC request", map[logging.ExtraKey]interface{}{
		"userId": req.UserId,
	})

	userID, err := uuid.Parse(req.UserId)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid user ID")
	}

	preferences, err := s.preferenceSvc.GetUserPreferences(ctx, userID)
	if err != nil {
		s.logger.Error(logging.General, logging.Api, "Failed to get preferences", map[logging.ExtraKey]interface{}{
			"error":  err.Error(),
			"userId": req.UserId,
		})
		return nil, status.Error(codes.Internal, err.Error())
	}

	pbPreferences := make([]*pb.Preference, len(preferences))
	for i, p := range preferences {
		var categorySettings map[string]bool
		if p.CategorySettings != "" {
			json.Unmarshal([]byte(p.CategorySettings), &categorySettings)
		}

		pbPreferences[i] = &pb.Preference{
			Id:               p.ID.String(),
			UserId:           p.UserID.String(),
			Type:             pb.NotificationType(pb.NotificationType_value[string(p.Type)]),
			IsEnabled:        p.IsEnabled,
			AllowInstant:     p.AllowInstant,
			AllowDigest:      p.AllowDigest,
			DigestFrequency:  p.DigestFrequency,
			CategorySettings: categorySettings,
		}
	}

	return &pb.GetUserPreferencesResponse{
		Preferences: pbPreferences,
	}, nil
}

// UpdatePreference updates user notification preferences
func (s *Server) UpdatePreference(ctx context.Context, req *pb.UpdatePreferenceRequest) (*pb.UpdatePreferenceResponse, error) {
	s.logger.Debug(logging.General, logging.Api, "UpdatePreference gRPC request", map[logging.ExtraKey]interface{}{
		"userId": req.UserId,
		"type":   req.Type.String(),
	})

	userID, err := uuid.Parse(req.UserId)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid user ID")
	}

	// Convert category settings to JSON
	categorySettingsJSON, _ := json.Marshal(req.CategorySettings)

	preference := &models.NotificationPreference{
		UserID:           userID,
		Type:             models.NotificationType(req.Type.String()),
		IsEnabled:        req.IsEnabled,
		AllowInstant:     req.AllowInstant,
		AllowDigest:      req.AllowDigest,
		DigestFrequency:  req.DigestFrequency,
		CategorySettings: string(categorySettingsJSON),
	}

	err = s.preferenceSvc.UpdatePreference(ctx, preference)
	if err != nil {
		s.logger.Error(logging.General, logging.Api, "Failed to update preference", map[logging.ExtraKey]interface{}{
			"error":  err.Error(),
			"userId": req.UserId,
		})
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &pb.UpdatePreferenceResponse{
		Success: true,
		Message: "Preference updated successfully",
	}, nil
}
