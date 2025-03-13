package model

import (
	"time"

	"github.com/specterops/bloodhound/src/database/types/null"
)

type AssetGroupHistoryAction string

const (
	AssetGroupHistoryActorSystem = "SYSTEM"
)

const (
	AssetGroupHistoryActionCreateSelector AssetGroupHistoryAction = "CreateSelector"
	AssetGroupHistoryActionUpdateSelector AssetGroupHistoryAction = "UpdateSelector"
	AssetGroupHistoryActionDeleteSelector AssetGroupHistoryAction = "DeleteSelector"
)

// AssetGroupHistory is the record of CRUD changes associated with v2 of the asset groups feature
type AssetGroupHistory struct {
	ID                int64                   `json:"id" gorm:"primaryKey"`
	CreatedAt         time.Time               `json:"created_at"`
	Actor             string                  `json:"actor"`
	Action            AssetGroupHistoryAction `json:"action"`
	Target            string                  `json:"target"`
	AssetGroupLabelId int                     `json:"assetGroupLabelId"`
	EnvironmentId     null.String             `json:"environmentId"`
	Note              null.String             `json:"note"`
}
