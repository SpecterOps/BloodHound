package database

import (
	"fmt"
	"github.com/gofrs/uuid"
	"github.com/specterops/bloodhound/src/model"
)

func (s *BloodhoundDB) ListSavedQueries(userID uuid.UUID, order string, filter model.SQLFilter, skip, limit int) (model.SavedQueries, int, error) {
	var (
		queries model.SavedQueries
		count   int64
	)

	cursor := s.Scope(Paginate(skip, limit)).Where("user_id = ?", userID).Order(order)

	if filter.SQLString != "" {
		cursor = cursor.Where(filter.SQLString, filter.Params)
	}

	if order != "" {
		cursor = cursor.Order(order)
	}

	if countResult := cursor.Count(&count); countResult.Error != nil {
		return queries, 0, fmt.Errorf("error fetching count: %v", countResult.Error.Error())
	}

	result := cursor.Find(&queries)
	return queries, int(count), CheckError(result)
}
