package database

import (
	"context"
	"fmt"

	"github.com/specterops/bloodhound/src/database/types/null"
	"github.com/specterops/bloodhound/src/model"
)

const (
	assetGroupHistoryTable = "asset_group_history"
)

// AssetGroupHistoryData defines the methods required to interact with the asset_group_history table
type AssetGroupHistoryData interface {
	CreateAssetGroupHistoryRecord(ctx context.Context, actor, target string, action model.AssetGroupHistoryAction, assetGroupLabelId int, environmentId, note string) error
	GetAssetGroupHistoryRecords(ctx context.Context) ([]model.AssetGroupHistory, error)
}

func (s *BloodhoundDB) CreateAssetGroupHistoryRecord(ctx context.Context, actor, target string, action model.AssetGroupHistoryAction, assetGroupLabelId int, environmentId, note string) error {
	return CheckError(s.db.WithContext(ctx).Exec(fmt.Sprintf("INSERT INTO %s (actor, target, action, asset_group_label_id, environment_id, note) VALUES (?, ?, ?, ?, ?, ?)", assetGroupHistoryTable),
		actor, target, action, assetGroupLabelId, null.StringFrom(environmentId), null.StringFrom(note)))
}

func (s *BloodhoundDB) GetAssetGroupHistoryRecords(ctx context.Context) ([]model.AssetGroupHistory, error) {
	var result []model.AssetGroupHistory
	return result, CheckError(s.db.WithContext(ctx).Raw(fmt.Sprintf("SELECT id, actor, target, action, asset_group_label_id, environment_id, note, created_at FROM %s", assetGroupHistoryTable)).Find(&result))
}
