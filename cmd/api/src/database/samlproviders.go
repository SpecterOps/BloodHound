package database

import (
	"context"

	"github.com/specterops/bloodhound/src/database/types/null"
	"github.com/specterops/bloodhound/src/model"
	"gorm.io/gorm"
)

const (
	samlProvidersTableName = "saml_providers"
)

// SAMLProviderData defines the interface required to interact with the oidc_providers table
type SAMLProviderData interface {
	CreateSAMLIdentityProvider(ctx context.Context, samlProvider model.SAMLProvider) (model.SAMLProvider, error)
	GetAllSAMLProviders(ctx context.Context) (model.SAMLProviders, error)
	GetSAMLProvider(ctx context.Context, id int32) (model.SAMLProvider, error)
	GetSAMLProviderUsers(ctx context.Context, id int32) (model.Users, error)
	UpdateSAMLIdentityProvider(ctx context.Context, samlProvider model.SAMLProvider) error
}

// CreateSAMLIdentityProvider creates a new saml_providers row using the data in the input struct
// This also creates the corresponding sso_provider entry
// INSERT INTO saml_identity_providers (...) VALUES (...)
func (s *BloodhoundDB) CreateSAMLIdentityProvider(ctx context.Context, samlProvider model.SAMLProvider) (model.SAMLProvider, error) {
	// Set the current version for root_uri_version
	samlProvider.RootURIVersion = model.SAMLRootURIVersion2

	auditEntry := model.AuditEntry{
		Action: model.AuditLogActionCreateSAMLIdentityProvider,
		Model:  &samlProvider, // Pointer is required to ensure success log contains updated fields after transaction
	}

	err := s.AuditableTransaction(ctx, auditEntry, func(tx *gorm.DB) error {
		bhdb := NewBloodhoundDB(tx, s.idResolver)

		// Create the associated SSO provider
		if ssoProvider, err := bhdb.CreateSSOProvider(ctx, samlProvider.Name, model.SessionAuthProviderSAML); err != nil {
			return err
		} else {
			samlProvider.SSOProviderID = null.Int32From(ssoProvider.ID)
			return CheckError(tx.WithContext(ctx).Create(&samlProvider))
		}
	})

	return samlProvider, err
}

// GetAllSAMLProviders returns all SAML providers
// SELECT * FROM saml_providers
func (s *BloodhoundDB) GetAllSAMLProviders(ctx context.Context) (model.SAMLProviders, error) {
	var (
		samlProviders model.SAMLProviders
		result        = s.db.WithContext(ctx).Find(&samlProviders)
	)

	return samlProviders, CheckError(result)
}

// GetSAMLProvider returns a SAML provider corresponding to the ID provided
// SELECT * FOM saml_providers WHERE id = ..
func (s *BloodhoundDB) GetSAMLProvider(ctx context.Context, id int32) (model.SAMLProvider, error) {
	var (
		samlProvider model.SAMLProvider
		result       = s.db.WithContext(ctx).First(&samlProvider, id)
	)

	return samlProvider, CheckError(result)
}

// GetSAMLProviderUsers returns all users that are bound to the SAML provider ID provided
// SELECT * FROM users WHERE saml_provider_id = ..
func (s *BloodhoundDB) GetSAMLProviderUsers(ctx context.Context, id int32) (model.Users, error) {
	var users model.Users
	return users, CheckError(s.preload(model.UserAssociations()).WithContext(ctx).Where("sso_provider_id = ?", id).Find(&users))
}

// CreateSAMLProvider updates a saml_providers row using the data in the input struct
// UPDATE saml_identity_providers SET (...) VALUES (...) WHERE id = ...
func (s *BloodhoundDB) UpdateSAMLIdentityProvider(ctx context.Context, provider model.SAMLProvider) error {
	auditEntry := model.AuditEntry{
		Action: model.AuditLogActionUpdateSAMLIdentityProvider,
		Model:  &provider, // Pointer is required to ensure success log contains updated fields after transaction
	}

	return s.AuditableTransaction(ctx, auditEntry, func(tx *gorm.DB) error {
		return CheckError(tx.WithContext(ctx).Save(&provider))
	})
}
