package api

import (
	"errors"
	"net/url"

	"github.com/specterops/bloodhound/dawgs/query"
	"github.com/specterops/bloodhound/src/model"
)

const (
	ErrResponseDetailsBadQueryParameterFilters    = "there are errors in the query parameter filters specified"
	ErrResponseDetailsFilterPredicateNotSupported = "the specified filter predicate is not supported for this column"
	ErrResponseDetailsColumnNotFilterable         = "the specified column cannot be filtered"
	ErrResponseDetailsColumnNotSortable           = "the specified column cannot be sorted"
)

type Sortable interface {
	IsSortable(column string) bool
}

func GetOrderCriteria(s Sortable, params url.Values) (model.OrderCriteria, error) {
	var (
		sortByColumns = params["sort_by"]
		orderCriteria model.OrderCriteria
	)

	for _, column := range sortByColumns {
		criterion := model.OrderCriterion{}

		if string(column[0]) == "-" {
			column = column[1:]
			criterion.Order = query.Descending()
		} else {
			criterion.Order = query.Ascending()
		}
		criterion.Property = column

		if !s.IsSortable(column) {
			return model.OrderCriteria{}, errors.New(ErrResponseDetailsColumnNotSortable)
		}

		orderCriteria = append(orderCriteria, criterion)
	}
	return orderCriteria, nil
}
