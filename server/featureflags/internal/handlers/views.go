package handlers

import (
	"database/sql"
	"encoding/json"
	"time"

	"github.com/specterops/bloodhound/server/featureflags/internal/services"
)

type FeatureFlagView struct {
	ID            int32        `json:"id"`
	CreatedAt     time.Time    `json:"created_at"`
	UpdatedAt     time.Time    `json:"updated_at"`
	DeletedAt     sql.NullTime `json:"deleted_at"`
	Key           string       `json:"key"`
	Name          string       `json:"name"`
	Description   string       `json:"description"`
	Enabled       bool         `json:"enabled"`
	UserUpdatable bool         `json:"user_updatable"`
}

func BuildFeatureFlagView(tf services.FeatureFlag) FeatureFlagView {
	return FeatureFlagView{
		ID:            tf.ID,
		CreatedAt:     tf.CreatedAt,
		UpdatedAt:     tf.UpdatedAt,
		Key:           tf.Key,
		Name:          tf.Name,
		Description:   tf.Description,
		Enabled:       tf.Enabled,
		UserUpdatable: tf.UserUpdatable,
	}
}

func (s FeatureFlagView) JSONView() ([]byte, error) { return json.Marshal(s) }

type FeatureFlagsView []FeatureFlagView

func BuildFeatureFlagsView(flags []services.FeatureFlag) FeatureFlagsView {
	views := make(FeatureFlagsView, 0, len(flags))
	for _, flag := range flags {
		views = append(views, BuildFeatureFlagView(flag))
	}
	return views
}

func (s FeatureFlagsView) JSONView() ([]byte, error) { return json.Marshal(s) }
