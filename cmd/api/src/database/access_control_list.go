package database

import (
	"context"

	"github.com/gofrs/uuid"
	"github.com/specterops/bloodhound/cmd/api/src/model"
	"gorm.io/gorm"
)

const (
	EnvironmentAccessControlTable = "environment_access_control"
)

type EnvironmentAccessControlData interface {
	GetEnvironmentAccessListForUser(ctx context.Context, userUuid uuid.UUID) (EnvironmentAccessList, error)
	UpdateEnvironmentListForUser(ctx context.Context, environments []string, userUuid uuid.UUID) error
}

// EnvironmentAccess defines the model for a row in the environment_access_control table
type EnvironmentAccess struct {
	UserID      string `json:"user_id"`
	Environment string `json:"environment"`
	model.BigSerial
}

// EnvironmentAccessList is a slice of EnvironmentAccess that provides additional helper methods
type EnvironmentAccessList []EnvironmentAccess

// CheckUserAccess returns true if the environment provided exists in the user's TAC list and false if not
func (s EnvironmentAccessList) CheckUserAccess(environment string) bool {
	for _, accessControl := range s {
		if accessControl.Environment == environment {
			return true
		}
	}

	return false
}

// GetEnvironmentAccessListForUser given a user's id, this will return all access control list rows for the user
func (s *BloodhoundDB) GetEnvironmentAccessListForUser(ctx context.Context, userUuid uuid.UUID) (EnvironmentAccessList, error) {
	var accessControlList []EnvironmentAccess

	result := s.db.WithContext(ctx).Table(EnvironmentAccessControlTable).Select("environment").Where("user_id = ?", userUuid.String()).Scan(&accessControlList)
	return accessControlList, CheckError(result)
}

// UpdateEnvironmentListForUser will remove all entries in the access control list for a user and add a new entry for each environment provided
func (s *BloodhoundDB) UpdateEnvironmentListForUser(ctx context.Context, environments []string, userUuid uuid.UUID) error {
	return s.db.Transaction(func(tx *gorm.DB) error {
		if err := CheckError(tx.WithContext(ctx).Table(EnvironmentAccessControlTable).Where("user_id = ?", userUuid).Delete(&EnvironmentAccess{})); err != nil {
			return err
		} else {
			for _, environment := range environments {
				newAccessControl := EnvironmentAccess{
					UserID:      userUuid.String(),
					Environment: environment,
				}

				result := tx.WithContext(ctx).Table(EnvironmentAccessControlTable).Create(&newAccessControl)
				if err := CheckError(result); err != nil {
					return err
				}
			}
		}
		return nil
	})
}

// func (s *BloodhoundDB) CheckUserAccessToEnvironment(ctx context.Context, environment ...string) bool {
// 	return s.db.WithContext(ctx).Table(EnvironmentAccessControlTable).Select("COUNT(*) > 0").Where("environment = ?")
// }
