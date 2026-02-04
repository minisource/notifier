package grpc

import (
	"context"
	"encoding/json"

	"github.com/google/uuid"
	"github.com/minisource/go-common/common"
	"github.com/minisource/go-common/logging"
	pb "github.com/minisource/go-sdk/notifier/proto/notifier/v1"
	"github.com/minisource/notifier/internal/models"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// CreateNotification handles creating a new notification
func (s *Server) CreateNotification(ctx context.Context, req *pb.CreateNotificationRequest) (*pb.CreateNotificationResponse, error) {
	// Parse user ID
	var userID uuid.UUID
	var err error
	if req.UserId != "" {
		userID, err = uuid.Parse(req.UserId)
		if err != nil {
			return nil, status.Error(codes.InvalidArgument, "invalid user_id format")
		}
	} else {
		userID = uuid.New() // System notification
	}

	// Parse template ID if provided
	var templateID *uuid.UUID
	if req.TemplateId != "" {
		tid, err := uuid.Parse(req.TemplateId)
		if err != nil {
			return nil, status.Error(codes.InvalidArgument, "invalid template_id format")
		}
		templateID = &tid
	}

	// Convert metadata to JSON string
	var metadataStr string
	if len(req.Metadata) > 0 {
		metadataBytes, err := json.Marshal(req.Metadata)
		if err != nil {
			return nil, status.Error(codes.InvalidArgument, "invalid metadata format")
		}
		metadataStr = string(metadataBytes)
	} else {
		// Use empty JSON object for PostgreSQL jsonb compatibility
		metadataStr = "{}"
	}

	// Create notification model
	notification := &models.Notification{
		UserID:         userID,
		Type:           convertNotificationType(req.Type),
		Priority:       convertPriority(req.Priority),
		RecipientEmail: req.RecipientEmail,
		RecipientPhone: req.RecipientPhone,
		RecipientID:    req.RecipientId,
		Subject:        req.Subject,
		Body:           req.Body,
		Metadata:       metadataStr,
		TemplateID:     templateID,
		Status:         models.NotificationStatusPending,
	}

	if req.ScheduledAt != nil {
		t := req.ScheduledAt.AsTime()
		notification.ScheduledAt = &t
	}

	// Create notification via service
	err = s.notifSvc.CreateNotification(ctx, notification)
	if err != nil {
		return &pb.CreateNotificationResponse{
			Success: false,
			Message: err.Error(),
		}, nil
	}

	return &pb.CreateNotificationResponse{
		Success:        true,
		NotificationId: notification.ID.String(),
		Message:        "Notification created successfully",
	}, nil
}

// GetNotification retrieves a notification by ID
func (s *Server) GetNotification(ctx context.Context, req *pb.GetNotificationRequest) (*pb.GetNotificationResponse, error) {
	notifID, err := uuid.Parse(req.NotificationId)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid notification_id format")
	}

	notif, err := s.notifSvc.GetNotification(ctx, notifID)
	if err != nil {
		return nil, status.Error(codes.NotFound, err.Error())
	}

	return &pb.GetNotificationResponse{
		Notification: convertToProtoNotification(notif),
	}, nil
}

// GetUserNotifications lists notifications for a user with pagination
func (s *Server) GetUserNotifications(ctx context.Context, req *pb.GetUserNotificationsRequest) (*pb.GetUserNotificationsResponse, error) {
	userID, err := uuid.Parse(req.UserId)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid user_id format")
	}

	page := int(req.Page)
	if page < 1 {
		page = 1
	}
	pageSize := int(req.PageSize)
	if pageSize < 1 {
		pageSize = 20
	}
	offset := (page - 1) * pageSize

	notifications, total, err := s.notifSvc.GetUserNotifications(ctx, userID, pageSize, offset)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	protoNotifications := make([]*pb.Notification, len(notifications))
	for i, n := range notifications {
		protoNotifications[i] = convertToProtoNotification(n)
	}

	return &pb.GetUserNotificationsResponse{
		Notifications: protoNotifications,
		Total:         int64(total),
	}, nil
}

// MarkAsRead marks a notification as read
func (s *Server) MarkAsRead(ctx context.Context, req *pb.MarkAsReadRequest) (*pb.MarkAsReadResponse, error) {
	notifID, err := uuid.Parse(req.NotificationId)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid notification_id format")
	}

	err = s.notifSvc.MarkAsRead(ctx, notifID)
	if err != nil {
		return &pb.MarkAsReadResponse{
			Success: false,
			Message: err.Error(),
		}, nil
	}

	return &pb.MarkAsReadResponse{
		Success: true,
		Message: "Notification marked as read",
	}, nil
}

// CreateBatchNotifications sends batch notifications
func (s *Server) CreateBatchNotifications(ctx context.Context, req *pb.CreateBatchNotificationsRequest) (*pb.CreateBatchNotificationsResponse, error) {
	var notifications []*models.Notification

	for _, n := range req.Notifications {
		var userID uuid.UUID
		var err error
		if n.UserId != "" {
			userID, err = uuid.Parse(n.UserId)
			if err != nil {
				continue // Skip invalid user IDs
			}
		} else {
			userID = uuid.New()
		}

		notification := &models.Notification{
			UserID:         userID,
			Type:           convertNotificationType(n.Type),
			Priority:       convertPriority(n.Priority),
			RecipientEmail: n.RecipientEmail,
			RecipientPhone: n.RecipientPhone,
			RecipientID:    n.RecipientId,
			Subject:        n.Subject,
			Body:           n.Body,
			Status:         models.NotificationStatusPending,
		}
		notifications = append(notifications, notification)
	}

	successIDs, errors := s.notifSvc.CreateBatchNotifications(ctx, notifications)

	// Convert success IDs to strings
	var successIDStrings []string
	for _, id := range successIDs {
		if id != uuid.Nil {
			successIDStrings = append(successIDStrings, id.String())
		}
	}

	// Convert errors to strings
	var errorMessages []string
	for _, err := range errors {
		if err != nil {
			errorMessages = append(errorMessages, err.Error())
		}
	}

	return &pb.CreateBatchNotificationsResponse{
		SuccessIds:    successIDStrings,
		ErrorMessages: errorMessages,
		SuccessCount:  int32(len(successIDStrings)),
		FailedCount:   int32(len(errorMessages)),
	}, nil
}

// StreamNotifications streams notifications (not fully implemented)
func (s *Server) StreamNotifications(req *pb.StreamNotificationsRequest, stream pb.NotificationService_StreamNotificationsServer) error {
	return status.Error(codes.Unimplemented, "streaming not yet implemented")
}

// GetUnreadNotifications lists unread notifications for a user
func (s *Server) GetUnreadNotifications(ctx context.Context, req *pb.GetUnreadNotificationsRequest) (*pb.GetUnreadNotificationsResponse, error) {
	userID, err := uuid.Parse(req.UserId)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid user_id format")
	}

	page := int(req.Page)
	if page < 1 {
		page = 1
	}
	pageSize := int(req.PageSize)
	if pageSize < 1 {
		pageSize = 20
	}
	offset := (page - 1) * pageSize

	notifications, total, err := s.notifSvc.GetUnreadNotifications(ctx, userID, pageSize, offset)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	protoNotifications := make([]*pb.Notification, len(notifications))
	for i, n := range notifications {
		protoNotifications[i] = convertToProtoNotification(n)
	}

	return &pb.GetUnreadNotificationsResponse{
		Notifications: protoNotifications,
		Total:         int64(total),
	}, nil
}

// SendSMS sends an SMS notification
// For template-based providers (Kavenegar), pass template key via req.Template
// and token values via req.Tokens map (e.g., {"code": "123456"})
// The notifier service will look up the template mapping and apply it
func (s *Server) SendSMS(ctx context.Context, req *pb.SendSMSRequest) (*pb.SendSMSResponse, error) {
	// ============ ENTRY POINT LOGGING ============
	s.logger.Info(logging.General, logging.Api, "========== gRPC SendSMS HANDLER CALLED ==========", map[logging.ExtraKey]interface{}{
		"entry": "START",
	})

	// Build metadata JSON with template and data
	// This will be used by the SMS handler to look up template and construct message
	metadata := make(map[string]interface{})

	// Add template key if specified
	if req.Template != "" {
		metadata["template"] = req.Template
	}

	// Build data map from tokens
	// The "data" key contains the caller's data dictionary
	// The handler will map these to provider-specific tokens
	data := make(map[string]interface{})
	tokens := req.GetTokens() // Use getter method for proper protobuf deserialization
	s.logger.Info(logging.General, logging.Api, "gRPC SendSMS received request", map[logging.ExtraKey]interface{}{
		"to":             req.GetTo(),
		"template":       req.GetTemplate(),
		"body":           req.GetBody(),
		"tokens":         tokens,
		"tokens_nil":     tokens == nil,
		"tokens_len":     len(tokens),
		"req_Tokens":     req.Tokens,
		"req_Tokens_nil": req.Tokens == nil,
		"req_Tokens_len": len(req.Tokens),
	})

	for k, v := range tokens {
		data[k] = v
	}
	if len(tokens) > 0 {
		s.logger.Info(logging.General, logging.Api, "gRPC SendSMS tokens converted to data", map[logging.ExtraKey]interface{}{
			"tokensCount": len(tokens),
			"data":        data,
		})
	}

	// If no tokens provided but body is set, use body as "code"
	body := req.GetBody()
	if len(data) == 0 && body != "" {
		data["code"] = body
		s.logger.Info(logging.General, logging.Api, "gRPC SendSMS using body as code", map[logging.ExtraKey]interface{}{
			"body": body,
		})
	}

	// CRITICAL: Nest data under "data" key for handler parsing
	if len(data) > 0 {
		metadata["data"] = data
		s.logger.Info(logging.General, logging.Api, "gRPC SendSMS metadata built", map[logging.ExtraKey]interface{}{
			"metadata": metadata,
		})
	} else {
		s.logger.Warn(logging.General, logging.Api, "gRPC SendSMS no data provided", map[logging.ExtraKey]interface{}{
			"tokens": tokens,
			"body":   body,
		})
	}

	// Convert metadata to JSON
	var metadataJSON string
	if len(metadata) > 0 {
		if jsonBytes, err := json.Marshal(metadata); err == nil {
			metadataJSON = string(jsonBytes)
		}
	}

	// Determine phone number (support both fields)
	phone := req.PhoneNumber
	if phone == "" {
		phone = req.To
	}

	// Normalize phone number to E.164 format (+989123456789)
	normalizedPhone := common.NormalizeIranPhone(phone)

	// Create a notification for SMS
	notification := &models.Notification{
		UserID:         uuid.New(), // System notification
		Type:           models.NotificationTypeSMS,
		Priority:       models.NotificationPriorityNormal,
		RecipientPhone: normalizedPhone,
		Body:           body,
		Metadata:       metadataJSON,
		Status:         models.NotificationStatusPending,
	}

	err := s.notifSvc.CreateNotification(ctx, notification)
	if err != nil {
		return &pb.SendSMSResponse{
			Success: false,
			Message: err.Error(),
		}, nil
	}

	return &pb.SendSMSResponse{
		Success:   true,
		MessageId: notification.ID.String(),
		Message:   "SMS notification created",
	}, nil
}

// SendEmail sends an email notification
func (s *Server) SendEmail(ctx context.Context, req *pb.SendEmailRequest) (*pb.SendEmailResponse, error) {
	// Create a notification for Email
	notification := &models.Notification{
		UserID:         uuid.New(), // System notification
		Type:           models.NotificationTypeEmail,
		Priority:       models.NotificationPriorityNormal,
		RecipientEmail: req.To,
		Subject:        req.Subject,
		Body:           req.Body,
		Status:         models.NotificationStatusPending,
	}

	err := s.notifSvc.CreateNotification(ctx, notification)
	if err != nil {
		return &pb.SendEmailResponse{
			Success: false,
			Message: err.Error(),
		}, nil
	}

	return &pb.SendEmailResponse{
		Success:   true,
		MessageId: notification.ID.String(),
		Message:   "Email notification created",
	}, nil
}

// Helper functions

func convertNotificationType(t pb.NotificationType) models.NotificationType {
	switch t {
	case pb.NotificationType_NOTIFICATION_TYPE_SMS:
		return models.NotificationTypeSMS
	case pb.NotificationType_NOTIFICATION_TYPE_EMAIL:
		return models.NotificationTypeEmail
	case pb.NotificationType_NOTIFICATION_TYPE_PUSH:
		return models.NotificationTypePush
	case pb.NotificationType_NOTIFICATION_TYPE_IN_APP:
		return models.NotificationTypeInApp
	default:
		return models.NotificationTypeEmail
	}
}

func convertPriority(p pb.NotificationPriority) models.NotificationPriority {
	switch p {
	case pb.NotificationPriority_NOTIFICATION_PRIORITY_LOW:
		return models.NotificationPriorityLow
	case pb.NotificationPriority_NOTIFICATION_PRIORITY_NORMAL:
		return models.NotificationPriorityNormal
	case pb.NotificationPriority_NOTIFICATION_PRIORITY_HIGH:
		return models.NotificationPriorityHigh
	case pb.NotificationPriority_NOTIFICATION_PRIORITY_URGENT:
		return models.NotificationPriorityUrgent
	default:
		return models.NotificationPriorityNormal
	}
}

func convertToProtoNotification(n *models.Notification) *pb.Notification {
	notif := &pb.Notification{
		Id:             n.ID.String(),
		UserId:         n.UserID.String(),
		Type:           convertToProtoType(n.Type),
		Priority:       convertToProtoPriority(n.Priority),
		RecipientEmail: n.RecipientEmail,
		RecipientPhone: n.RecipientPhone,
		Subject:        n.Subject,
		Body:           n.Body,
		Status:         convertToProtoStatus(n.Status),
		CreatedAt:      timestamppb.New(n.CreatedAt),
		UpdatedAt:      timestamppb.New(n.UpdatedAt),
	}

	if n.ReadAt != nil {
		notif.ReadAt = timestamppb.New(*n.ReadAt)
	}
	if n.SentAt != nil {
		notif.SentAt = timestamppb.New(*n.SentAt)
	}

	return notif
}

func convertToProtoType(t models.NotificationType) pb.NotificationType {
	switch t {
	case models.NotificationTypeSMS:
		return pb.NotificationType_NOTIFICATION_TYPE_SMS
	case models.NotificationTypeEmail:
		return pb.NotificationType_NOTIFICATION_TYPE_EMAIL
	case models.NotificationTypePush:
		return pb.NotificationType_NOTIFICATION_TYPE_PUSH
	case models.NotificationTypeInApp:
		return pb.NotificationType_NOTIFICATION_TYPE_IN_APP
	default:
		return pb.NotificationType_NOTIFICATION_TYPE_UNSPECIFIED
	}
}

func convertToProtoPriority(p models.NotificationPriority) pb.NotificationPriority {
	switch p {
	case models.NotificationPriorityLow:
		return pb.NotificationPriority_NOTIFICATION_PRIORITY_LOW
	case models.NotificationPriorityNormal:
		return pb.NotificationPriority_NOTIFICATION_PRIORITY_NORMAL
	case models.NotificationPriorityHigh:
		return pb.NotificationPriority_NOTIFICATION_PRIORITY_HIGH
	case models.NotificationPriorityUrgent:
		return pb.NotificationPriority_NOTIFICATION_PRIORITY_URGENT
	default:
		return pb.NotificationPriority_NOTIFICATION_PRIORITY_UNSPECIFIED
	}
}

func convertToProtoStatus(s models.NotificationStatus) pb.NotificationStatus {
	switch s {
	case models.NotificationStatusPending:
		return pb.NotificationStatus_NOTIFICATION_STATUS_PENDING
	case models.NotificationStatusSending:
		return pb.NotificationStatus_NOTIFICATION_STATUS_SENDING
	case models.NotificationStatusSent:
		return pb.NotificationStatus_NOTIFICATION_STATUS_SENT
	case models.NotificationStatusFailed:
		return pb.NotificationStatus_NOTIFICATION_STATUS_FAILED
	case models.NotificationStatusCanceled:
		return pb.NotificationStatus_NOTIFICATION_STATUS_CANCELED
	default:
		return pb.NotificationStatus_NOTIFICATION_STATUS_UNSPECIFIED
	}
}
