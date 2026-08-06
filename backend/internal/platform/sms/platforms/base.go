package providers

import "context"

type SmsClient interface {
	SendMessage(param map[string]string, targetPhoneNumber ...string) error
}

// HealthCheckable is implemented by clients that can verify connectivity or
// credentials against the real provider API.
type HealthCheckable interface {
	Check(ctx context.Context) error
}

// RequestDescriber is implemented by clients that can describe the exact
// outbound HTTP request they will make for a send. The provider-attempt
// logger uses it to record the REAL request (URL + form body) instead of the
// internal params map, so operators can verify exactly what the provider
// receives. The returned URL/body must be sanitized (no secrets).
type RequestDescriber interface {
	DescribeRequest(param map[string]string, targetPhoneNumber ...string) (method, url, body string)
}