package dogtags

import "errors"

// ErrNotFound is returned when a dogtag key doesn't exist in the provider
var ProviderNotImplemented = errors.New("No-op provider not implemented")

// Provider is the interface for dogtag backends.
// Providers are simple key-value stores - they don't know about defaults.
// The service layer handles defaults when a key isn't found.
type Provider interface {
	GetBool(key string) (bool, error)
	GetString(key string) (string, error)
	GetInt(key string) (int64, error)
	GetFloat(key string) (float64, error)
}

// NoopProvider returns ErrNotFound for all keys.
// The service will use defaults for everything.
type NoopProvider struct{}

func NewNoopProvider() *NoopProvider {
	return &NoopProvider{}
}

func (p *NoopProvider) GetBool(key string) (bool, error) {
	return false, ProviderNotImplemented
}

func (p *NoopProvider) GetString(key string) (string, error) {
	return "", ProviderNotImplemented
}

func (p *NoopProvider) GetInt(key string) (int64, error) {
	return 0, ProviderNotImplemented
}

func (p *NoopProvider) GetFloat(key string) (float64, error) {
	return 0, ProviderNotImplemented
}
