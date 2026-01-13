package database_test

// import (
// 	"context"
// 	"testing"

// 	"github.com/specterops/bloodhound/cmd/api/src/database"
// 	"github.com/stretchr/testify/assert"
// 	"github.com/stretchr/testify/require"
// )

// func TestBloodhoundDB_UpsertGraphSchemaExtension(t *testing.T) {
// 	type args struct {
// 		environments []database.EnvironmentInput
// 		findings     []database.FindingInput
// 	}
// 	tests := []struct {
// 		name          string
// 		setupData     func(t *testing.T, db *database.BloodhoundDB) int32 // Returns extensionId
// 		args          args
// 		assert        func(t *testing.T, db *database.BloodhoundDB, extensionId int32)
// 		expectedError string
// 	}{
// 		{
// 			name: "Success: Create new environments and findings with remediations",
// 			setupData: func(t *testing.T, db *database.BloodhoundDB) int32 {
// 				t.Helper()
// 				ext, err := db.CreateGraphSchemaExtension(context.Background(), "TestExt", "Test", "v1.0.0")
// 				require.NoError(t, err)

// 				_, err = db.CreateSchemaEnvironment(context.Background(), ext.ID, int32(1), int32(1))
// 				require.NoError(t, err)

// 				return ext.ID
// 			},
// 			args: args{
// 				environments: []database.EnvironmentInput{
// 					{
// 						EnvironmentKindName: "Tag_Tier_Zero",
// 						SourceKindName:      "Base",
// 						PrincipalKinds:  []string{"Tag_Owned", "Tag_Tier_Zero"},
// 					},
// 				},
// 				findings: []database.FindingInput{
// 					{
// 						Name:                 "T0WriteOwner",
// 						DisplayName:          "Write Owner",
// 						RelationshipKindName: "Tag_Tier_Zero",
// 						EnvironmentKindName:  "Tag_Tier_Zero",
// 						RemediationInput: database.RemediationInput{
// 							ShortDescription: "User has write owner permission",
// 							LongDescription:  "This permission allows modifying object owner",
// 							ShortRemediation: "Remove write owner permissions",
// 							LongRemediation:  "Review and remove unnecessary permissions",
// 						},
// 					},
// 					{
// 						Name:                 "T0DCSync",
// 						DisplayName:          "DCSync Attack",
// 						RelationshipKindName: "Tag_Tier_Zero",
// 						EnvironmentKindName:  "Tag_Tier_Zero",
// 						RemediationInput: database.RemediationInput{
// 							ShortDescription: "Principal can perform DCSync",
// 							LongDescription:  "Can extract password hashes",
// 							ShortRemediation: "Revoke replication permissions",
// 							LongRemediation:  "Remove DS-Replication-Get-Changes permissions",
// 						},
// 					},
// 				},
// 			},
// 			assert: func(t *testing.T, db *database.BloodhoundDB, extensionId int32) {
// 				t.Helper()

// 				// Verify findings were created
// 				finding1, err := db.GetSchemaRelationshipFindingByName(context.Background(), "T0WriteOwner")
// 				require.NoError(t, err)
// 				assert.Equal(t, extensionId, finding1.SchemaExtensionId)
// 				assert.Equal(t, "Write Owner", finding1.DisplayName)

// 				finding2, err := db.GetSchemaRelationshipFindingByName(context.Background(), "T0DCSync")
// 				require.NoError(t, err)
// 				assert.Equal(t, extensionId, finding2.SchemaExtensionId)
// 				assert.Equal(t, "DCSync Attack", finding2.DisplayName)

// 				// Verify remediations were created
// 				remediation1, err := db.GetRemediationByFindingId(context.Background(), finding1.ID)
// 				require.NoError(t, err)
// 				assert.Equal(t, "User has write owner permission", remediation1.ShortDescription)
// 				assert.Equal(t, "Remove write owner permissions", remediation1.ShortRemediation)

// 				remediation2, err := db.GetRemediationByFindingId(context.Background(), finding2.ID)
// 				require.NoError(t, err)
// 				assert.Equal(t, "Principal can perform DCSync", remediation2.ShortDescription)
// 				assert.Equal(t, "Revoke replication permissions", remediation2.ShortRemediation)
// 			},
// 		},
// 		// {
// 		// 	name: "Success: Update existing findings and remediations",
// 		// 	setupData: func(t *testing.T, db *database.BloodhoundDB) int32 {
// 		// 		t.Helper()
// 		// 		ext, err := db.CreateGraphSchemaExtension(context.Background(), "TestExt2", "Test2", "v1.0.0")
// 		// 		require.NoError(t, err)

// 		// 		env, err := db.CreateSchemaEnvironment(context.Background(), ext.ID, 1, 1)
// 		// 		require.NoError(t, err)

// 		// 		// Create initial finding with remediation
// 		// 		finding, err := db.CreateSchemaRelationshipFinding(context.Background(), ext.ID, 1, env.ID, "ExistingFinding", "Old Display Name")
// 		// 		require.NoError(t, err)

// 		// 		_, err = db.CreateRemediation(context.Background(), finding.ID, "old short", "old long", "old short rem", "old long rem")
// 		// 		require.NoError(t, err)

// 		// 		return ext.ID
// 		// 	},
// 		// 	args: args{
// 		// 		environments: []database.EnvironmentInput{
// 		// 			{
// 		// 				EnvironmentKind: "Domain",
// 		// 				SourceKind:      "Base",
// 		// 				PrincipalKinds:  []string{"User"},
// 		// 			},
// 		// 		},
// 		// 		findings: []database.FindingInput{
// 		// 			{
// 		// 				Name:                 "ExistingFinding",
// 		// 				DisplayName:          "Updated Display Name",
// 		// 				RelationshipKindName: "WriteOwner",
// 		// 				EnvironmentKindName:  "Domain",
// 		// 				RemediationInput: database.RemediationInput{
// 		// 					ShortDescription: "updated short",
// 		// 					LongDescription:  "updated long",
// 		// 					ShortRemediation: "updated short rem",
// 		// 					LongRemediation:  "updated long rem",
// 		// 				},
// 		// 			},
// 		// 		},
// 		// 	},
// 		// 	assert: func(t *testing.T, db *database.BloodhoundDB, extensionId int32) {
// 		// 		t.Helper()

// 		// 		// Verify finding was updated (deleted and recreated)
// 		// 		finding, err := db.GetSchemaRelationshipFindingByName(context.Background(), "ExistingFinding")
// 		// 		require.NoError(t, err)
// 		// 		assert.Equal(t, extensionId, finding.SchemaExtensionId)
// 		// 		assert.Equal(t, "Updated Display Name", finding.DisplayName)

// 		// 		// Verify remediation was updated
// 		// 		remediation, err := db.GetRemediationByFindingId(context.Background(), finding.ID)
// 		// 		require.NoError(t, err)
// 		// 		assert.Equal(t, "updated short", remediation.ShortDescription)
// 		// 		assert.Equal(t, "updated long", remediation.LongDescription)
// 		// 		assert.Equal(t, "updated short rem", remediation.ShortRemediation)
// 		// 		assert.Equal(t, "updated long rem", remediation.LongRemediation)
// 		// 	},
// 		// },
// 		// {
// 		// 	name: "Success: Empty environments and findings",
// 		// 	setupData: func(t *testing.T, db *database.BloodhoundDB) int32 {
// 		// 		t.Helper()
// 		// 		ext, err := db.CreateGraphSchemaExtension(context.Background(), "TestExt3", "Test3", "v1.0.0")
// 		// 		require.NoError(t, err)

// 		// 		return ext.ID
// 		// 	},
// 		// 	args: args{
// 		// 		environments: []database.EnvironmentInput{},
// 		// 		findings:     []database.FindingInput{},
// 		// 	},
// 		// 	assert: func(t *testing.T, db *database.BloodhoundDB, extensionId int32) {
// 		// 		t.Helper()
// 		// 		// Nothing to assert - just verify no error occurred
// 		// 	},
// 		// },
// 	}
// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			testSuite := setupIntegrationTestSuite(t)
// 			defer teardownIntegrationTestSuite(t, &testSuite)

// 			extensionId := tt.setupData(t, testSuite.BHDatabase)

// 			err := testSuite.BHDatabase.UpsertGraphSchemaExtension(
// 				context.Background(),
// 				extensionId,
// 				tt.args.environments,
// 				tt.args.findings,
// 			)

// 			if tt.expectedError != "" {
// 				require.Error(t, err)
// 				assert.Contains(t, err.Error(), tt.expectedError)
// 			} else {
// 				require.NoError(t, err)
// 			}

// 			if tt.assert != nil {
// 				tt.assert(t, testSuite.BHDatabase, extensionId)
// 			}
// 		})
// 	}
// }
