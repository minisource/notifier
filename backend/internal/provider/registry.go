package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"

	"github.com/minisource/go-common/logging"
	"github.com/minisource/notifier/internal/platform/email"
	"github.com/minisource/notifier/internal/repository"
)

// ProviderRegistry manages available providers and provides selection logic
type ProviderRegistry struct {
	mu               sync.RWMutex
	providers        map[string]Provider // key: "channel:name" (e.g., "sms:kavenegar")
	defaultProviders map[Channel]string  // channel → default provider name
	logger           logging.Logger
}

// NewProviderRegistry creates an empty provider registry
func NewProviderRegistry(logger logging.Logger) *ProviderRegistry {
	return &ProviderRegistry{
		providers:        make(map[string]Provider),
		defaultProviders: make(map[Channel]string),
		logger:           logger,
	}
}

// Register adds a provider to the registry
func (r *ProviderRegistry) Register(p Provider) {
	r.mu.Lock()
	defer r.mu.Unlock()
	key := providerKey(p.Channel(), p.Name())
	r.providers[key] = p
	r.logger.Debug(logging.General, logging.Insert, "Registered provider", map[logging.ExtraKey]interface{}{
		logging.ExtraKey("channel"): p.Channel(),
		logging.ExtraKey("name"):    p.Name(),
	})
}

// SetDefault sets the default provider for a channel
func (r *ProviderRegistry) SetDefault(channel Channel, name string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.defaultProviders[channel] = name
}

// GetProvider retrieves a specific provider by channel and name
func (r *ProviderRegistry) GetProvider(channel Channel, name string) (Provider, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	p, ok := r.providers[providerKey(channel, name)]
	return p, ok
}

// GetDefaultProvider returns the default provider for a channel
func (r *ProviderRegistry) GetDefaultProvider(channel Channel) (Provider, error) {
	r.mu.RLock()
	defaultName, hasDefault := r.defaultProviders[channel]
	r.mu.RUnlock()

	if hasDefault {
		if p, ok := r.GetProvider(channel, defaultName); ok {
			return p, nil
		}
	}

	// Fallback: try to find any provider for this channel
	r.mu.RLock()
	defer r.mu.RUnlock()
	for key, p := range r.providers {
		if p.Channel() == channel {
			r.logger.Debug(logging.General, logging.Select, "Using fallback provider for channel", map[logging.ExtraKey]interface{}{
				logging.ExtraKey("channel"):  channel,
				logging.ExtraKey("provider"): p.Name(),
				logging.ExtraKey("key"):      key,
			})
			return p, nil
		}
	}

	return nil, &ProviderNotFoundError{Channel: channel}
}

// HasProvider checks if any provider is registered for a channel
func (r *ProviderRegistry) HasProvider(channel Channel) bool {
	r.mu.RLock()
	defer r.mu.RUnlock()
	for _, p := range r.providers {
		if p.Channel() == channel {
			return true
		}
	}
	return false
}

// SendViaDefault sends a message through the default provider for the channel
func (r *ProviderRegistry) SendViaDefault(ctx context.Context, msg *Message) (*SendResult, error) {
	channel := ChannelFromMessage(msg)
	if channel == "" {
		return nil, fmt.Errorf("cannot determine channel from message: no channel in metadata")
	}
	provider, err := r.GetDefaultProvider(channel)
	if err != nil {
		return nil, err
	}
	return provider.Send(ctx, msg)
}

// ChannelFromMessage determines the channel from a message's metadata or content
func ChannelFromMessage(msg *Message) Channel {
	if msg.Metadata != nil {
		if ch, ok := msg.Metadata["channel"]; ok {
			if parsed, err := ParseChannel(ch); err == nil {
				return parsed
			}
		}
	}
	return ""
}

// BuildDefaultRegistry creates a registry with mock providers for all channels
func BuildDefaultRegistry(logger logging.Logger, notifRepo repository.NotificationRepository) *ProviderRegistry {
	registry := NewProviderRegistry(logger)

	// Register mock providers for all channels
	for _, ch := range AllChannels() {
		mock := NewMockProvider(ch)
		registry.Register(mock)
		registry.SetDefault(ch, mock.Name())
	}

	// Register in-app provider
	inapp := NewInAppProvider(notifRepo, logger)
	registry.Register(inapp)
	registry.SetDefault(ChannelInApp, inapp.Name())

	return registry
}

// ConfigureFromDatabase loads provider configs from the settings table
func (r *ProviderRegistry) ConfigureFromDatabase(ctx context.Context, settingRepo repository.SettingRepository) {
	r.loadSMSProvider(ctx, settingRepo)
	r.loadEmailProvider(ctx, settingRepo)
	r.loadPushProvider(ctx, settingRepo)
	r.loadBaleSafirProvider(ctx, settingRepo)
}

func (r *ProviderRegistry) loadSMSProvider(ctx context.Context, settingRepo repository.SettingRepository) {
	setting, err := settingRepo.GetByKey(ctx, "sms.providers")
	if err != nil {
		r.logger.Warn(logging.General, logging.Select, "No SMS provider config in database, using mock", nil)
		return
	}

	smsProvider, err := NewSMSProviderFromConfig(setting.Value)
	if err != nil {
		r.logger.Warn(logging.General, logging.Insert, "Failed to create SMS provider from config", map[logging.ExtraKey]interface{}{
			logging.ExtraKey("error"): err.Error(),
		})
		return
	}

	r.Register(smsProvider)
	r.SetDefault(ChannelSMS, smsProvider.Name())
	r.logger.Info(logging.General, logging.Insert, "SMS provider configured", map[logging.ExtraKey]interface{}{
		logging.ExtraKey("provider"): smsProvider.Name(),
	})
}

func (r *ProviderRegistry) loadEmailProvider(ctx context.Context, settingRepo repository.SettingRepository) {
	setting, err := settingRepo.GetByKey(ctx, "email.providers")
	if err != nil {
		r.logger.Warn(logging.General, logging.Select, "No email provider config in database, using mock", nil)
		return
	}

	var emailCfg email.ProviderConfig
	if err := json.Unmarshal([]byte(setting.Value), &emailCfg); err != nil {
		r.logger.Warn(logging.General, logging.Insert, "Failed to parse email provider config", map[logging.ExtraKey]interface{}{
			logging.ExtraKey("error"): err.Error(),
		})
		return
	}

	emailProvider, err := NewSMTPProvider(&emailCfg)
	if err != nil {
		r.logger.Warn(logging.General, logging.Insert, "Failed to create email provider from config", map[logging.ExtraKey]interface{}{
			logging.ExtraKey("error"): err.Error(),
		})
		return
	}

	r.Register(emailProvider)
	r.SetDefault(ChannelEmail, emailProvider.Name())
	r.logger.Info(logging.General, logging.Insert, "Email provider configured", map[logging.ExtraKey]interface{}{
		logging.ExtraKey("provider"): emailProvider.Name(),
	})
}

func (r *ProviderRegistry) loadPushProvider(ctx context.Context, settingRepo repository.SettingRepository) {
	setting, err := settingRepo.GetByKey(ctx, "push.providers")
	if err != nil {
		r.logger.Warn(logging.General, logging.Select, "No push provider config in database, using mock", nil)
		return
	}

	var pushCfg struct {
		Provider string `json:"provider"`
	}
	if err := json.Unmarshal([]byte(setting.Value), &pushCfg); err != nil {
		r.logger.Warn(logging.General, logging.Insert, "Failed to parse push provider config", map[logging.ExtraKey]interface{}{
			logging.ExtraKey("error"): err.Error(),
		})
		return
	}

	r.logger.Info(logging.General, logging.Insert, "Push provider found in config (delegated to existing push client)", map[logging.ExtraKey]interface{}{
		logging.ExtraKey("provider"): pushCfg.Provider,
	})
	// Push notifications are currently delegated to the existing push handler adapter
}

const settingKeyBaleSafir = "sms.providers.bale_safir"

func (r *ProviderRegistry) loadBaleSafirProvider(ctx context.Context, settingRepo repository.SettingRepository) {
	setting, err := settingRepo.GetByKey(ctx, settingKeyBaleSafir)
	if err != nil {
		r.logger.Warn(logging.General, logging.Select, "No Bale Safir provider config in database, skipped", nil)
		return
	}

	baleSafir, err := NewBaleSafirProvider("bale_safir", setting.Value)
	if err != nil {
		r.logger.Warn(logging.General, logging.Insert, "Failed to create Bale Safir provider from config", map[logging.ExtraKey]interface{}{
			logging.ExtraKey("error"): err.Error(),
		})
		return
	}

	r.Register(baleSafir)
	r.SetDefault(ChannelSMS, baleSafir.Name())
	r.logger.Info(logging.General, logging.Insert, "Bale Safir SMS provider configured as default", nil)
}

func providerKey(channel Channel, name string) string {
	return string(channel) + ":" + name
}
