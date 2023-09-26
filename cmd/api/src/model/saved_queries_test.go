package model_test

import (
	"github.com/specterops/bloodhound/src/model"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm/utils"
	"testing"
)

func TestSavedQueries_IsSortable(t *testing.T) {
	savedQueries := model.SavedQueries{}
	for _, column := range []string{"user_id", "name", "query", "id", "created_at", "updated_at", "deleted_at"} {
		require.True(t, savedQueries.IsSortable(column))
	}

	require.False(t, savedQueries.IsSortable("foobar"))
}

func TestSavedQueries_ValidFilters(t *testing.T) {
	savedQueries := model.SavedQueries{}
	validFilters := savedQueries.ValidFilters()
	require.Equal(t, 3, len(validFilters))

	for _, column := range []string{"user_id", "name", "query"} {
		operators, ok := validFilters[column]
		require.True(t, ok)
		require.Equal(t, 2, len(operators))
	}
}

func TestSavedQueries_GetValidFilterPredicatesAsStrings(t *testing.T) {
	savedQueries := model.SavedQueries{}
	for _, column := range []string{"user_id", "name", "query"} {
		predicates, err := savedQueries.GetValidFilterPredicatesAsStrings(column)
		require.Nil(t, err)
		require.Equal(t, 2, len(predicates))
		require.True(t, utils.Contains(predicates, string(model.Equals)))
		require.True(t, utils.Contains(predicates, string(model.NotEquals)))
	}
}

func TestSavedQueries_IsString(t *testing.T) {
	savedQueries := model.SavedQueries{}
	for _, column := range []string{"name", "query"} {
		require.True(t, savedQueries.IsString(column))
	}
}
