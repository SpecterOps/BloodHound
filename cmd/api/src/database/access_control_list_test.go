//go:build integration
// +build integration

package database_test

import (
	"context"
	"testing"

	"github.com/specterops/bloodhound/cmd/api/src/database/types/null"
	"github.com/specterops/bloodhound/cmd/api/src/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBloodhoundDB_AccessControlList(t *testing.T) {
	var (
		ctx = context.Background()
		db  = setupIntegrationTestSuite(t)
	)
	defer teardownIntegrationTestSuite(t, &db)

	newUser, err := db.BHDatabase.CreateUser(context.Background(), model.User{
		FirstName:     null.StringFrom("First"),
		LastName:      null.StringFrom("Last"),
		EmailAddress:  null.StringFrom(userPrincipal),
		PrincipalName: userPrincipal,
	})
	require.NoError(t, err)

	userUuid := newUser.ID

	t.Run("UpdateEnvironmentListForUser", func(t *testing.T) {
		err = db.BHDatabase.UpdateEnvironmentListForUser(ctx, []string{"1234", "1234"}, userUuid)
		require.NoError(t, err)
	})

	t.Run("GetEnvironmentAccessListForUser", func(t *testing.T) {
		result, err := db.BHDatabase.GetEnvironmentAccessListForUser(ctx, userUuid)
		require.NoError(t, err)
		assert.Len(t, result, 2)
	})

	t.Run("Deleting User Removes ACL", func(t *testing.T) {
		err := db.BHDatabase.DeleteUser(ctx, newUser)
		require.NoError(t, err)

		result, err := db.BHDatabase.GetEnvironmentAccessListForUser(ctx, userUuid)
		require.NoError(t, err)
		assert.Len(t, result, 0)
	})
}
