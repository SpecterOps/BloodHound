//go:build integration

package database_test

import (
	"testing"

	"github.com/specterops/bloodhound/cmd/api/src/database"
	"github.com/specterops/bloodhound/cmd/api/src/model"
	"github.com/stretchr/testify/require"
)

func TestBloodhoundDB_CreateAndGetExtensionSchemaNodeKind(t *testing.T) {

	testSuite := setupIntegrationTestSuite(t)

	defer teardownIntegrationTestSuite(t, &testSuite)

	extension, err := testSuite.BHDatabase.CreateGraphSchemaExtension(testSuite.Context, "test_extension", "test_extension", "1.0.0")
	require.NoError(t, err)
	var (
		nodeKind1 = model.SchemaNodeKind{
			Name:              "Test_Kind_1",
			SchemaExtensionId: extension.ID,
			DisplayName:       "Test_Kind_1",
			Description:       "A test kind",
			IsDisplayKind:     false,
			Icon:              "test_icon",
			IconColor:         "blue",
		}
		nodeKind2 = model.SchemaNodeKind{
			Name:              "Test_Kind_2",
			SchemaExtensionId: extension.ID,
			DisplayName:       "Test_Kind_2",
			Description:       "A test kind",
			IsDisplayKind:     false,
			Icon:              "test_icon",
			IconColor:         "blue",
		}

		want = model.SchemaNodeKind{
			Name:              "Test_Kind_1",
			SchemaExtensionId: extension.ID,
			DisplayName:       "Test_Kind_1",
			Description:       "A test kind",
			IsDisplayKind:     false,
			Icon:              "test_icon",
			IconColor:         "blue",
			Serial: model.Serial{
				ID: 1,
			},
		}
		want2 = model.SchemaNodeKind{
			Serial: model.Serial{
				ID: 2,
			},
			Name:              "Test_Kind_2",
			SchemaExtensionId: extension.ID,
			DisplayName:       "Test_Kind_2",
			Description:       "A test kind",
			IsDisplayKind:     false,
			Icon:              "test_icon",
			IconColor:         "blue",
		}
	)

	// Successfully create one model.SchemaNodeKind
	gotNodeKind1, err := testSuite.BHDatabase.CreateSchemaNodeKind(testSuite.Context, nodeKind1.Name, nodeKind1.SchemaExtensionId, nodeKind1.DisplayName, nodeKind1.Description, nodeKind1.IsDisplayKind, nodeKind1.Icon, nodeKind1.IconColor)
	require.NoError(t, err)
	compareSchemaNodeKind(t, gotNodeKind1, want)
	// Successfully create a second model.SchemaNodeKind
	gotNodeKind2, err := testSuite.BHDatabase.CreateSchemaNodeKind(testSuite.Context, nodeKind2.Name, nodeKind2.SchemaExtensionId, nodeKind2.DisplayName, nodeKind2.Description, nodeKind2.IsDisplayKind, nodeKind2.Icon, nodeKind2.IconColor)
	require.NoError(t, err)
	compareSchemaNodeKind(t, gotNodeKind2, want2)
	// Successfully get the first model.SchemaNodeKind
	gotNodeKind1, err = testSuite.BHDatabase.GetSchemaNodeKindByID(testSuite.Context, 1)
	require.NoError(t, err)
	compareSchemaNodeKind(t, gotNodeKind1, want)
	// fail - return error for record that does not exist
	gotNodeKind1, err = testSuite.BHDatabase.GetSchemaNodeKindByID(testSuite.Context, 21321)
	require.EqualError(t, err, "entity not found")
	// fail - return error indicating non unique name
	_, err = testSuite.BHDatabase.CreateSchemaNodeKind(testSuite.Context, nodeKind2.Name, nodeKind2.SchemaExtensionId, nodeKind2.DisplayName, nodeKind2.Description, nodeKind2.IsDisplayKind, nodeKind2.Icon, nodeKind2.IconColor)
	require.ErrorIs(t, err, database.ErrDuplicateSchemaNodeKindName)
}

func compareSchemaNodeKind(t *testing.T, got, want model.SchemaNodeKind) {
	t.Helper()
	require.Equalf(t, want.ID, got.ID, "CreateSchemaNodeKind(%v) - id mismatch", got.ID)
	require.Equalf(t, want.Name, got.Name, "CreateSchemaNodeKind(%v) - name mismatch", got.Name)
	require.Equalf(t, want.SchemaExtensionId, got.SchemaExtensionId, "CreateSchemaNodeKind(%v) - extension_id mismatch", got.SchemaExtensionId)
	require.Equalf(t, want.DisplayName, got.DisplayName, "CreateSchemaNodeKind(%v) - display_name mismatch", got.DisplayName)
	require.Equalf(t, want.Description, got.Description, "CreateSchemaNodeKind(%v) - description mismatch", got.Description)
	require.Equalf(t, want.IsDisplayKind, got.IsDisplayKind, "CreateSchemaNodeKind(%v) - is_display_kind mismatch", got.IsDisplayKind)
	require.Equalf(t, want.Icon, got.Icon, "CreateSchemaNodeKind(%v) - icon mismatch", got.Icon)
	require.Equalf(t, want.IconColor, got.IconColor, "CreateSchemaNodeKind(%v) - icon_color mismatch", got.IconColor)
}
