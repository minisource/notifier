package email

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"net/smtp"
	"strings"
)

// EmailClient interface for sending emails
type EmailClient interface {
	SendEmail(to, subject, body string, isHTML bool) error
}

// ProviderConfig holds email provider configuration
type ProviderConfig struct {
	Provider  string `json:"provider"` // smtp, sendgrid, ses, mailgun
	Host      string `json:"host"`
	Port      int    `json:"port"`
	Username  string `json:"username"`
	Password  string `json:"password"`
	From      string `json:"from"`
	FromName  string `json:"fromName"`
	UseTLS    bool   `json:"useTls"`
	APIKey    string `json:"apiKey"`    // for SendGrid, Mailgun
	Domain    string `json:"domain"`    // for Mailgun
	Region    string `json:"region"`    // for AWS SES
	AccessID  string `json:"accessId"`  // for AWS SES
	AccessKey string `json:"accessKey"` // for AWS SES
}

// ParseProviderConfig parses JSON config string into ProviderConfig
func ParseProviderConfig(configJSON string) (*ProviderConfig, error) {
	var config ProviderConfig
	if err := json.Unmarshal([]byte(configJSON), &config); err != nil {
		return nil, fmt.Errorf("failed to parse email provider config: %w", err)
	}
	return &config, nil
}

// NewClientFromConfig creates an email client based on the provider configuration
func NewClientFromConfig(config *ProviderConfig) (EmailClient, error) {
	switch config.Provider {
	case "smtp":
		return NewSMTPClient(config)
	default:
		return nil, fmt.Errorf("unsupported email provider: %s", config.Provider)
	}
}

// SMTPClient implements EmailClient for SMTP
type SMTPClient struct {
	config *ProviderConfig
}

// NewSMTPClient creates a new SMTP email client
func NewSMTPClient(config *ProviderConfig) (*SMTPClient, error) {
	if config.Host == "" {
		return nil, fmt.Errorf("SMTP host is required")
	}
	if config.Port == 0 {
		config.Port = 587 // Default TLS port
	}
	if config.From == "" {
		return nil, fmt.Errorf("from address is required")
	}
	return &SMTPClient{config: config}, nil
}

// SendEmail sends an email via SMTP
func (c *SMTPClient) SendEmail(to, subject, body string, isHTML bool) error {
	// Prepare headers
	headers := make(map[string]string)
	if c.config.FromName != "" {
		headers["From"] = fmt.Sprintf("%s <%s>", c.config.FromName, c.config.From)
	} else {
		headers["From"] = c.config.From
	}
	headers["To"] = to
	headers["Subject"] = subject
	if isHTML {
		headers["MIME-Version"] = "1.0"
		headers["Content-Type"] = "text/html; charset=UTF-8"
	} else {
		headers["Content-Type"] = "text/plain; charset=UTF-8"
	}

	// Build message
	var msg strings.Builder
	for k, v := range headers {
		msg.WriteString(fmt.Sprintf("%s: %s\r\n", k, v))
	}
	msg.WriteString("\r\n")
	msg.WriteString(body)

	addr := fmt.Sprintf("%s:%d", c.config.Host, c.config.Port)

	// Set up authentication
	var auth smtp.Auth
	if c.config.Username != "" && c.config.Password != "" {
		auth = smtp.PlainAuth("", c.config.Username, c.config.Password, c.config.Host)
	}

	// Send with TLS if configured
	if c.config.UseTLS {
		return c.sendWithTLS(addr, auth, to, msg.String())
	}

	// Send without TLS (or STARTTLS)
	return smtp.SendMail(addr, auth, c.config.From, []string{to}, []byte(msg.String()))
}

// sendWithTLS sends email using explicit TLS connection
func (c *SMTPClient) sendWithTLS(addr string, auth smtp.Auth, to, message string) error {
	tlsConfig := &tls.Config{
		ServerName: c.config.Host,
	}

	conn, err := tls.Dial("tcp", addr, tlsConfig)
	if err != nil {
		return fmt.Errorf("failed to dial TLS: %w", err)
	}
	defer conn.Close()

	client, err := smtp.NewClient(conn, c.config.Host)
	if err != nil {
		return fmt.Errorf("failed to create SMTP client: %w", err)
	}
	defer client.Quit()

	if auth != nil {
		if err := client.Auth(auth); err != nil {
			return fmt.Errorf("SMTP auth failed: %w", err)
		}
	}

	if err := client.Mail(c.config.From); err != nil {
		return fmt.Errorf("MAIL FROM failed: %w", err)
	}

	if err := client.Rcpt(to); err != nil {
		return fmt.Errorf("RCPT TO failed: %w", err)
	}

	w, err := client.Data()
	if err != nil {
		return fmt.Errorf("DATA failed: %w", err)
	}

	if _, err := w.Write([]byte(message)); err != nil {
		return fmt.Errorf("failed to write message: %w", err)
	}

	if err := w.Close(); err != nil {
		return fmt.Errorf("failed to close data writer: %w", err)
	}

	return nil
}
