package api

import (
	"fmt"

	"github.com/specterops/bloodhound/src/model"
)

type Filterable interface {
	ValidFilters() map[string][]model.FilterOperator
}

func GetFilterableColumns(f Filterable) []string {
	var columns = make([]string, 0)
	for column := range f.ValidFilters() {
		columns = append(columns, column)
	}
	return columns
}

func GetValidFilterPredicatesAsStrings(f Filterable, column string) ([]string, error) {
	if predicates, validColumn := f.ValidFilters()[column]; !validColumn {
		return []string{}, fmt.Errorf("the specified column cannot be filtered")
	} else {
		var stringPredicates = make([]string, 0)
		for _, predicate := range predicates {
			stringPredicates = append(stringPredicates, string(predicate))
		}
		return stringPredicates, nil
	}
}
