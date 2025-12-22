package dogtags

import "errors"

// ErrNotFound is returned when a dogtag key doesn't exist in the provider
var ProviderNotImplemented = errors.New("no-op provider not implemented")

// Provider is the interface for dogtag backends.
// Providers are simple key-value stores - they don't know about defaults.
// The service layer handles defaults when a key isn't found.
type Provider interface {
	GetFlagAsBool(key string) (bool, error)
	GetFlagAsString(key string) (string, error)
	GetFlagAsInt(key string) (int64, error)
}

// NoopProvider returns ErrNotFound for all keys.
// The service will use defaults for everything.
type NoopProvider struct{}

func NewNoopProvider() *NoopProvider {
	return &NoopProvider{}
}

func (p *NoopProvider) GetFlagAsBool(key string) (bool, error) {
	return false, ProviderNotImplemented
}

func (p *NoopProvider) GetFlagAsString(key string) (string, error) {
	return "", ProviderNotImplemented
}

func (p *NoopProvider) GetFlagAsInt(key string) (int64, error) {
	return 0, ProviderNotImplemented
}
