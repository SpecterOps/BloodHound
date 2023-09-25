package model

import "fmt"

type SavedQuery struct {
	UserID string `json:"user_id" gorm:"index:,unique,composite:compositeIndex"`
	Name   string `json:"name" gorm:"index:,unique,composite:compositeIndex"`
	Query  string `json:"query"`

	BigSerial
}

type SavedQueries []SavedQuery

func (s SavedQueries) IsSortable(column string) bool {
	switch column {
	case "user_id",
		"name",
		"query",
		"id",
		"created_at",
		"updated_at",
		"deleted_at":
		return true
	default:
		return false
	}
}

func (s SavedQueries) ValidFilters() map[string][]FilterOperator {
	return map[string][]FilterOperator{
		"user_id": {Equals, NotEquals},
		"name":    {Equals, NotEquals},
		"query":   {Equals, NotEquals},
	}
}

func (s SavedQueries) GetFilterableColumns() []string {
	var columns = make([]string, 0)
	for column := range s.ValidFilters() {
		columns = append(columns, column)
	}
	return columns
}

func (s SavedQueries) GetValidFilterPredicatesAsStrings(column string) ([]string, error) {
	if predicates, validColumn := s.ValidFilters()[column]; !validColumn {
		return []string{}, fmt.Errorf(ErrorResponseDetailsColumnNotFilterable)
	} else {
		var stringPredicates = make([]string, 0)
		for _, predicate := range predicates {
			stringPredicates = append(stringPredicates, string(predicate))
		}
		return stringPredicates, nil
	}
}

func (s SavedQueries) IsString(column string) bool {
	switch column {
	case "name",
		"query":
		return true
	default:
		return false
	}
}
