package wellknown

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNodeNamePrefix_String(t *testing.T) {
	tests := []struct {
		name   string
		prefix nodeNamePrefix
		want   string
	}{
		{
			name:   "empty prefix",
			prefix: nodeNamePrefix(""),
			want:   "",
		},
		{
			name:   "non-empty prefix",
			prefix: nodeNamePrefix("TEST"),
			want:   "TEST",
		},
		{
			name:   "with special characters",
			prefix: nodeNamePrefix("TEST-123"),
			want:   "TEST-123",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.prefix.String()
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestNodeNamePrefix_AppendSuffix(t *testing.T) {
	tests := []struct {
		name   string
		prefix nodeNamePrefix
		suffix string
		want   string
	}{
		{
			name:   "empty prefix with empty suffix",
			prefix: nodeNamePrefix(""),
			suffix: "",
			want:   "@",
		},
		{
			name:   "non-empty prefix with empty suffix",
			prefix: nodeNamePrefix("TEST"),
			suffix: "",
			want:   "TEST@",
		},
		{
			name:   "empty prefix with non-empty suffix",
			prefix: nodeNamePrefix(""),
			suffix: "example.com",
			want:   "@example.com",
		},
		{
			name:   "non-empty prefix with non-empty suffix",
			prefix: nodeNamePrefix("TEST"),
			suffix: "example.com",
			want:   "TEST@example.com",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.prefix.AppendSuffix(tt.suffix)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestNewNodeNamePrefix(t *testing.T) {
	tests := []struct {
		name   string
		prefix string
		want   string
	}{
		{
			name:   "empty prefix",
			prefix: "",
			want:   "",
		},
		{
			name:   "non-empty prefix",
			prefix: "TEST",
			want:   "TEST",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NewNodeNamePrefix(tt.prefix)
			assert.Equal(t, tt.want, got.String())
		})
	}
}

func TestDefineNodeName(t *testing.T) {
	tests := []struct {
		name   string
		prefix NodeNamePrefix
		suffix string
		want   string
	}{
		{
			name:   "with predefined prefix",
			prefix: DomainUsersNodeNamePrefix,
			suffix: "example.com",
			want:   "DOMAIN USERS@example.com",
		},
		{
			name:   "with custom prefix",
			prefix: NewNodeNamePrefix("CUSTOM"),
			suffix: "example.com",
			want:   "CUSTOM@example.com",
		},
		{
			name:   "with empty suffix",
			prefix: NewNodeNamePrefix("TEST"),
			suffix: "",
			want:   "TEST@",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := DefineNodeName(tt.prefix, tt.suffix)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestPredefinedNodeNamePrefixes(t *testing.T) {
	tests := []struct {
		name   string
		prefix NodeNamePrefix
		want   string
	}{
		{
			name:   "DomainUsersNodeNamePrefix",
			prefix: DomainUsersNodeNamePrefix,
			want:   "DOMAIN USERS",
		},
		{
			name:   "AuthenticatedUsersNodeNamePrefix",
			prefix: AuthenticatedUsersNodeNamePrefix,
			want:   "AUTHENTICATED USERS",
		},
		{
			name:   "EveryoneNodeNamePrefix",
			prefix: EveryoneNodeNamePrefix,
			want:   "EVERYONE",
		},
		{
			name:   "DomainComputerNodeNamePrefix",
			prefix: DomainComputerNodeNamePrefix,
			want:   "DOMAIN COMPUTERS",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.prefix.String()
			assert.Equal(t, tt.want, got)
		})
	}
}
