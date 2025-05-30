package api

import (
	"context"
	"fmt"
	"strconv"

	"github.com/specterops/bloodhound/src/database"
	"github.com/specterops/bloodhound/src/model"
)

func ParseTierIdWithTierZeroFallback(ctx context.Context, db database.Database, maybeTierIdParam []string) (int, error) {
	if len(maybeTierIdParam) != 0 {
		if tierIdParam, err := strconv.Atoi(maybeTierIdParam[0]); err != nil {
			return 0, err
		} else if _, err = db.GetAssetGroupTag(ctx, tierIdParam); err != nil {
			return 0, err
		} else {
			return tierIdParam, nil
		}
	}
	if agt, err := db.GetAssetGroupTags(ctx, model.SQLFilter{SQLString: "type = ? AND position = ?", Params: []interface{}{model.AssetGroupTagTypeTier, model.AssetGroupTierZeroPosition}}); err != nil {
		return 0, err
	} else if len(agt) == 0 {
		return 0, fmt.Errorf("no asset group tag found for tier zero")
	} else {
		return agt[0].ID, nil
	}
}
