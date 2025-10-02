//go:build !noop && !yaml

package providers

import (
	"context"
	"crypto"
	"crypto/md5"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"fmt"
	"log/slog"
	"os"
	"sync"

	"gopkg.in/yaml.v3"
)

const replicatedPublicKey = `-----BEGIN PUBLIC KEY-----
MIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEA1mTL6f4ziBazQWg3xNu0
m7OufoX35gS+B+eoDJ77un34nfJlgzPpddKMMJKomxJ3cEmqQM0LwYaVW6P7ZUpW
y8vE4im7Gyz9KvjNnERJPadLd86q6xnPkBrsyxxizd7E/3/mamBCtV44xWss7Ems
I4qdHzseyhTDdr8o/ybpKbQa2JHChtmIqM+RIsNiJMc7duztzepUwBhDPrTBJQqx
3APk98D2shzDlruhyx35X57BEnq9z/HfwH36SGI/UaTXi2ag5xQM8lgH6uzgPtpp
loCJVoUrhdFbMl2xxo/dCpQrYcMT6E+58kC9huhK78f8KXRt7disgSjvQt9t0yo1
qwIDAQAB
-----END PUBLIC KEY-----`

// ReplicatedLicense represents the structure of a Replicated license YAML file
type ReplicatedLicense struct {
	APIVersion string `yaml:"apiVersion"`
	Kind       string `yaml:"kind"`
	Metadata   struct {
		Name string `yaml:"name"`
	} `yaml:"metadata"`
	Spec struct {
		LicenseID    string `yaml:"licenseID"`
		AppSlug      string `yaml:"appSlug"`
		ChannelName  string `yaml:"channelName"`
		Entitlements map[string]struct {
			Title       string      `yaml:"title"`
			Description string      `yaml:"description,omitempty"`
			Value       interface{} `yaml:"value"`
			ValueType   string      `yaml:"valueType"`
			Signature   struct {
				V1 string `yaml:"v1"`
			} `yaml:"signature"`
		} `yaml:"entitlements"`
	} `yaml:"spec"`
}

type ReplicatedProvider struct {
	flags    map[string]interface{}
	mu       sync.RWMutex
	filePath string
}

func NewProvider(filePath string) (*ReplicatedProvider, error) {
	if filePath == "" {
		filePath = "license.yaml"
	}

	p := &ReplicatedProvider{
		flags:    make(map[string]interface{}),
		filePath: filePath,
	}

	if err := p.loadFlags(); err != nil {
		return nil, fmt.Errorf("failed to load replicated license: %w", err)
	}

	return p, nil
}

func (p *ReplicatedProvider) GetBoolFlag(ctx context.Context, key string, defaultValue bool) bool {
	p.mu.RLock()
	defer p.mu.RUnlock()

	if value, exists := p.flags[key]; exists {
		if boolVal, ok := value.(bool); ok {
			return boolVal
		}
	}
	return defaultValue
}

func (p *ReplicatedProvider) GetStringFlag(ctx context.Context, key string, defaultValue string) string {
	p.mu.RLock()
	defer p.mu.RUnlock()

	if value, exists := p.flags[key]; exists {
		if strVal, ok := value.(string); ok {
			return strVal
		}
	}
	return defaultValue
}

func (p *ReplicatedProvider) GetIntFlag(ctx context.Context, key string, defaultValue int64) int64 {
	p.mu.RLock()
	defer p.mu.RUnlock()

	if value, exists := p.flags[key]; exists {
		switch val := value.(type) {
		case int:
			return int64(val)
		case int64:
			return val
		case float64:
			return int64(val)
		}
	}
	return defaultValue
}

func (p *ReplicatedProvider) GetFloatFlag(ctx context.Context, key string, defaultValue float64) float64 {
	p.mu.RLock()
	defer p.mu.RUnlock()

	if value, exists := p.flags[key]; exists {
		if floatVal, ok := value.(float64); ok {
			return floatVal
		}
	}
	return defaultValue
}

func (p *ReplicatedProvider) loadFlags() error {
	slog.Info("================================================================================")
	slog.Info("||                                                                            ||")
	slog.Info("||                    REPLICATED LICENSE VALIDATION                          ||")
	slog.Info("||                                                                            ||")
	slog.Info("================================================================================")
	slog.Info(fmt.Sprintf("License file path: %s", p.filePath))

	// Read license file
	slog.Info("Step 1/4: Reading license file from disk...")
	data, err := os.ReadFile(p.filePath)
	if err != nil {
		slog.Error("================================================================================")
		slog.Error("||  FATAL: LICENSE FILE NOT FOUND OR UNREADABLE                            ||")
		slog.Error("================================================================================")
		panic(fmt.Sprintf("failed to read license file %s: %v", p.filePath, err))
	}
	slog.Info("✓ License file read successfully")

	// Parse YAML
	slog.Info("Step 2/4: Parsing license YAML structure...")
	var license ReplicatedLicense
	if err := yaml.Unmarshal(data, &license); err != nil {
		slog.Error("================================================================================")
		slog.Error("||  FATAL: LICENSE FILE MALFORMED OR CORRUPTED                             ||")
		slog.Error("================================================================================")
		panic(fmt.Sprintf("failed to parse license YAML file %s: %v", p.filePath, err))
	}
	slog.Info("✓ License YAML parsed successfully")

	// Validate license structure
	slog.Info("Step 3/4: Validating license metadata...")
	if license.Kind != "License" {
		slog.Error("================================================================================")
		slog.Error("||  FATAL: INVALID LICENSE FILE FORMAT                                     ||")
		slog.Error("================================================================================")
		panic(fmt.Sprintf("invalid license file: expected kind 'License', got '%s'", license.Kind))
	}
	slog.Info(fmt.Sprintf("  License ID: %s", license.Spec.LicenseID))
	slog.Info(fmt.Sprintf("  App Slug: %s", license.Spec.AppSlug))
	slog.Info(fmt.Sprintf("  Channel: %s", license.Spec.ChannelName))
	slog.Info(fmt.Sprintf("  Entitlements found: %d", len(license.Spec.Entitlements)))
	slog.Info("✓ License metadata valid")

	// Parse public key
	slog.Info("Step 4/4: Verifying cryptographic signatures...")
	pubKey, err := parsePublicKey(replicatedPublicKey)
	if err != nil {
		slog.Error("================================================================================")
		slog.Error("||  FATAL: INTERNAL ERROR - INVALID PUBLIC KEY                             ||")
		slog.Error("================================================================================")
		panic(fmt.Sprintf("failed to parse replicated public key: %v", err))
	}

	// Extract and verify entitlements
	flags := make(map[string]interface{})
	for key, entitlement := range license.Spec.Entitlements {
		slog.Info(fmt.Sprintf("  Verifying entitlement: %s", key))

		// Check that signature exists (basic tamper detection)
		if entitlement.Signature.V1 == "" {
			slog.Error("================================================================================")
			slog.Error(fmt.Sprintf("||  FATAL: MISSING SIGNATURE FOR '%s'", key))
			slog.Error("||  This indicates license tampering or corruption.                        ||")
			slog.Error("================================================================================")
			panic(fmt.Sprintf("license entitlement '%s' is missing signature - possible tampering detected", key))
		}

		// Verify signature
		if err := verifySignature(entitlement.Value, entitlement.Signature.V1, pubKey); err != nil {
			slog.Error("================================================================================")
			slog.Error(fmt.Sprintf("||  FATAL: SIGNATURE VERIFICATION FAILED FOR '%s'", key))
			slog.Error("||  This indicates license tampering or manual modification.               ||")
			slog.Error("================================================================================")
			panic(fmt.Sprintf("license entitlement '%s' signature verification failed: %v", key, err))
		}

		slog.Info(fmt.Sprintf("    ✓ %s = %v (signature valid)", key, entitlement.Value))
		flags[key] = entitlement.Value
	}

	p.mu.Lock()
	p.flags = flags
	p.mu.Unlock()

	slog.Info("================================================================================")
	slog.Info("||                                                                            ||")
	slog.Info("||             ✓ LICENSE VALIDATION SUCCESSFUL                               ||")
	slog.Info("||                                                                            ||")
	slog.Info("================================================================================")
	slog.Info("Loaded Entitlements:")
	for key, value := range flags {
		slog.Info(fmt.Sprintf("  • %s = %v", key, value))
	}
	slog.Info("================================================================================")

	return nil
}

func parsePublicKey(pemStr string) (*rsa.PublicKey, error) {
	block, _ := pem.Decode([]byte(pemStr))
	if block == nil {
		return nil, fmt.Errorf("failed to decode PEM block")
	}

	pub, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("failed to parse public key: %w", err)
	}

	rsaPub, ok := pub.(*rsa.PublicKey)
	if !ok {
		return nil, fmt.Errorf("not an RSA public key")
	}

	return rsaPub, nil
}

func verifySignature(value interface{}, signatureB64 string, pubKey *rsa.PublicKey) error {
	// Decode base64 signature
	sig, err := base64.StdEncoding.DecodeString(signatureB64)
	if err != nil {
		return fmt.Errorf("failed to decode signature: %w", err)
	}

	// Convert value to string and compute MD5 hash (per Replicated docs)
	valueStr := fmt.Sprintf("%v", value)
	hash := md5.Sum([]byte(valueStr))

	// Verify RSA-PSS signature
	err = rsa.VerifyPSS(pubKey, crypto.MD5, hash[:], sig, nil)
	if err != nil {
		return fmt.Errorf("signature verification failed: %w", err)
	}

	return nil
}
