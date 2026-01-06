package opengraphschema_test

import (
	"context"
	"errors"
	"testing"

	"github.com/specterops/bloodhound/cmd/api/src/database"
	"github.com/specterops/bloodhound/cmd/api/src/model"
	"github.com/specterops/bloodhound/cmd/api/src/services/opengraphschema"
	schemamocks "github.com/specterops/bloodhound/cmd/api/src/services/opengraphschema/mocks"
	"github.com/specterops/dawgs/graph"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestOpenGraphSchemaService_UpsertSchemaEnvironmentWithPrincipalKinds(t *testing.T) {
	type mocks struct {
		mockOpenGraphSchema *schemamocks.MockOpenGraphSchemaRepository
	}
	type args struct {
		environment    model.SchemaEnvironment
		principalKinds []model.SchemaEnvironmentPrincipalKind
	}
	tests := []struct {
		name       string
		mocks      mocks
		setupMocks func(t *testing.T, mock *mocks)
		args       args
		expected   error
	}{
		{
			name: "Error: Validation - Failed to retrieve environment kind",
			args: args{
				environment: model.SchemaEnvironment{
					SchemaExtensionId: int32(1),
					EnvironmentKindId: int32(1),
					SourceKindId:      int32(1),
				},
				principalKinds: []model.SchemaEnvironmentPrincipalKind{},
			},
			setupMocks: func(t *testing.T, mocks *mocks) {
				t.Helper()
				mocks.mockOpenGraphSchema.EXPECT().GetKindById(gomock.Any(), int32(1)).Return(model.Kind{}, errors.New("error"))
			},
			expected: errors.New("error validating schema environment: error retrieving environment kind: error"),
		},
		{
			name: "Error: Validation - Environment Kind doesn't exist in Kinds table",
			args: args{
				environment: model.SchemaEnvironment{
					SchemaExtensionId: int32(1),
					EnvironmentKindId: int32(1),
					SourceKindId:      int32(1),
				},
				principalKinds: []model.SchemaEnvironmentPrincipalKind{},
			},
			setupMocks: func(t *testing.T, mocks *mocks) {
				t.Helper()
				mocks.mockOpenGraphSchema.EXPECT().GetKindById(gomock.Any(), int32(1)).Return(model.Kind{}, database.ErrNotFound)
			},
			expected: errors.New("error validating schema environment: error retrieving environment kind: entity not found"),
		},
		{
			name: "Error: Validation - Failed to retrieve source kinds",
			args: args{
				environment: model.SchemaEnvironment{
					SchemaExtensionId: int32(1),
					EnvironmentKindId: int32(1),
					SourceKindId:      int32(1),
				},
				principalKinds: []model.SchemaEnvironmentPrincipalKind{},
			},
			setupMocks: func(t *testing.T, mocks *mocks) {
				t.Helper()
				mocks.mockOpenGraphSchema.EXPECT().GetKindById(gomock.Any(), int32(1)).Return(model.Kind{}, nil)
				mocks.mockOpenGraphSchema.EXPECT().GetSourceKinds(gomock.Any()).Return([]database.SourceKind{}, errors.New("error"))
			},
			expected: errors.New("error validating schema environment: error retrieving source kinds: error"),
		},
		{
			name: "Error: Validation - Source Kind doesn't exist",
			args: args{
				environment: model.SchemaEnvironment{
					SchemaExtensionId: int32(1),
					EnvironmentKindId: int32(3),
					SourceKindId:      int32(1),
				},
				principalKinds: []model.SchemaEnvironmentPrincipalKind{},
			},
			setupMocks: func(t *testing.T, mocks *mocks) {
				t.Helper()
				mocks.mockOpenGraphSchema.EXPECT().GetKindById(gomock.Any(), int32(3)).Return(model.Kind{}, nil)
				mocks.mockOpenGraphSchema.EXPECT().GetSourceKinds(gomock.Any()).Return([]database.SourceKind{
					{
						ID:   3,
						Name: graph.StringKind("kind"),
					},
				}, nil)
			},
			expected: errors.New("error validating schema environment: invalid source kind id 1"),
		},
		{
			name: "Error: Validation - Failed to retrieve principal kind",
			args: args{
				environment: model.SchemaEnvironment{
					SchemaExtensionId: int32(1),
					EnvironmentKindId: int32(3),
					SourceKindId:      int32(3),
				},
				principalKinds: []model.SchemaEnvironmentPrincipalKind{
					{
						EnvironmentId: int32(1),
						PrincipalKind: int32(99),
					},
				},
			},
			setupMocks: func(t *testing.T, mocks *mocks) {
				t.Helper()
				// Environment validation
				mocks.mockOpenGraphSchema.EXPECT().GetKindById(gomock.Any(), int32(3)).Return(model.Kind{}, nil)
				mocks.mockOpenGraphSchema.EXPECT().GetSourceKinds(gomock.Any()).Return([]database.SourceKind{
					{
						ID:   3,
						Name: graph.StringKind("kind"),
					},
				}, nil)
				// Principal kind validation
				mocks.mockOpenGraphSchema.EXPECT().GetKindById(gomock.Any(), int32(99)).Return(model.Kind{}, errors.New("error"))
			},
			expected: errors.New("error validating principal kind: error retrieving kind by id: error"),
		},
		{
			name: "Error: Validation - Principal Kind doesn't exist",
			args: args{
				environment: model.SchemaEnvironment{
					SchemaExtensionId: int32(1),
					EnvironmentKindId: int32(3),
					SourceKindId:      int32(3),
				},
				principalKinds: []model.SchemaEnvironmentPrincipalKind{
					{
						EnvironmentId: int32(1),
						PrincipalKind: int32(99),
					},
				},
			},
			setupMocks: func(t *testing.T, mocks *mocks) {
				t.Helper()
				// Environment validation
				mocks.mockOpenGraphSchema.EXPECT().GetKindById(gomock.Any(), int32(3)).Return(model.Kind{}, nil)
				mocks.mockOpenGraphSchema.EXPECT().GetSourceKinds(gomock.Any()).Return([]database.SourceKind{
					{
						ID:   3,
						Name: graph.StringKind("kind"),
					},
				}, nil)
				// Principal kind validation
				mocks.mockOpenGraphSchema.EXPECT().GetKindById(gomock.Any(), int32(99)).Return(model.Kind{}, database.ErrNotFound)
			},
			expected: errors.New("error validating principal kind: invalid principal kind id 99"),
		},
		{
			name: "Error: GetSchemaEnvironmentById fails",
			args: args{
				environment: model.SchemaEnvironment{
					SchemaExtensionId: int32(3),
					EnvironmentKindId: int32(3),
					SourceKindId:      int32(3),
				},
				principalKinds: []model.SchemaEnvironmentPrincipalKind{},
			},
			setupMocks: func(t *testing.T, mocks *mocks) {
				t.Helper()
				mocks.mockOpenGraphSchema.EXPECT().GetKindById(gomock.Any(), int32(3)).Return(model.Kind{}, nil)
				mocks.mockOpenGraphSchema.EXPECT().GetSourceKinds(gomock.Any()).Return([]database.SourceKind{
					{
						ID:   3,
						Name: graph.StringKind("kind"),
					},
				}, nil)
				mocks.mockOpenGraphSchema.EXPECT().GetSchemaEnvironmentById(gomock.Any(), int32(0)).Return(model.SchemaEnvironment{}, errors.New("error"))
			},
			expected: errors.New("error upserting schema environment: error retrieving schema environment id 0: error"),
		},
		{
			name: "Error: DeleteSchemaEnvironment fails",
			args: args{
				environment: model.SchemaEnvironment{
					SchemaExtensionId: int32(3),
					EnvironmentKindId: int32(3),
					SourceKindId:      int32(3),
				},
				principalKinds: []model.SchemaEnvironmentPrincipalKind{},
			},
			setupMocks: func(t *testing.T, mocks *mocks) {
				t.Helper()
				mocks.mockOpenGraphSchema.EXPECT().GetKindById(gomock.Any(), int32(3)).Return(model.Kind{}, nil)
				mocks.mockOpenGraphSchema.EXPECT().GetSourceKinds(gomock.Any()).Return([]database.SourceKind{
					{
						ID:   3,
						Name: graph.StringKind("kind"),
					},
				}, nil)
				mocks.mockOpenGraphSchema.EXPECT().GetSchemaEnvironmentById(gomock.Any(), int32(0)).Return(model.SchemaEnvironment{
					Serial: model.Serial{ID: 5},
				}, nil)
				mocks.mockOpenGraphSchema.EXPECT().DeleteSchemaEnvironment(gomock.Any(), int32(5)).Return(errors.New("error"))
			},
			expected: errors.New("error upserting schema environment: error deleting schema environment 5: error"),
		},
		{
			name: "Error: CreateSchemaEnvironment fails after delete",
			args: args{
				environment: model.SchemaEnvironment{
					SchemaExtensionId: int32(3),
					EnvironmentKindId: int32(3),
					SourceKindId:      int32(3),
				},
				principalKinds: []model.SchemaEnvironmentPrincipalKind{},
			},
			setupMocks: func(t *testing.T, mocks *mocks) {
				t.Helper()
				mocks.mockOpenGraphSchema.EXPECT().GetKindById(gomock.Any(), int32(3)).Return(model.Kind{}, nil)
				mocks.mockOpenGraphSchema.EXPECT().GetSourceKinds(gomock.Any()).Return([]database.SourceKind{
					{
						ID:   3,
						Name: graph.StringKind("kind"),
					},
				}, nil)
				mocks.mockOpenGraphSchema.EXPECT().GetSchemaEnvironmentById(gomock.Any(), int32(0)).Return(model.SchemaEnvironment{
					Serial: model.Serial{ID: 5},
				}, nil)
				mocks.mockOpenGraphSchema.EXPECT().DeleteSchemaEnvironment(gomock.Any(), int32(5)).Return(nil)
				mocks.mockOpenGraphSchema.EXPECT().CreateSchemaEnvironment(gomock.Any(), int32(3), int32(3), int32(3)).Return(model.SchemaEnvironment{}, errors.New("error"))
			},
			expected: errors.New("error upserting schema environment: error creating schema environment: error"),
		},
		{
			name: "Error: CreateSchemaEnvironment fails on new environment",
			args: args{
				environment: model.SchemaEnvironment{
					SchemaExtensionId: int32(3),
					EnvironmentKindId: int32(3),
					SourceKindId:      int32(3),
				},
				principalKinds: []model.SchemaEnvironmentPrincipalKind{},
			},
			setupMocks: func(t *testing.T, mocks *mocks) {
				t.Helper()
				mocks.mockOpenGraphSchema.EXPECT().GetKindById(gomock.Any(), int32(3)).Return(model.Kind{}, nil)
				mocks.mockOpenGraphSchema.EXPECT().GetSourceKinds(gomock.Any()).Return([]database.SourceKind{
					{
						ID:   3,
						Name: graph.StringKind("kind"),
					},
				}, nil)
				mocks.mockOpenGraphSchema.EXPECT().GetSchemaEnvironmentById(gomock.Any(), int32(0)).Return(model.SchemaEnvironment{}, database.ErrNotFound)
				mocks.mockOpenGraphSchema.EXPECT().CreateSchemaEnvironment(gomock.Any(), int32(3), int32(3), int32(3)).Return(model.SchemaEnvironment{}, errors.New("error"))
			},
			expected: errors.New("error upserting schema environment: error creating schema environment: error"),
		},
		{
			name: "Error: GetSchemaEnvironmentPrincipalKindsByEnvironmentId fails",
			args: args{
				environment: model.SchemaEnvironment{
					SchemaExtensionId: int32(3),
					EnvironmentKindId: int32(3),
					SourceKindId:      int32(3),
				},
				principalKinds: []model.SchemaEnvironmentPrincipalKind{
					{
						PrincipalKind: int32(3),
					},
				},
			},
			setupMocks: func(t *testing.T, mocks *mocks) {
				t.Helper()
				// Environment validation
				mocks.mockOpenGraphSchema.EXPECT().GetKindById(gomock.Any(), int32(3)).Return(model.Kind{}, nil)
				mocks.mockOpenGraphSchema.EXPECT().GetSourceKinds(gomock.Any()).Return([]database.SourceKind{
					{
						ID:   3,
						Name: graph.StringKind("kind"),
					},
				}, nil)
				// Principal kind validation
				mocks.mockOpenGraphSchema.EXPECT().GetKindById(gomock.Any(), int32(3)).Return(model.Kind{}, nil)

				// Environment upsert
				mocks.mockOpenGraphSchema.EXPECT().GetSchemaEnvironmentById(gomock.Any(), int32(0)).Return(model.SchemaEnvironment{}, database.ErrNotFound)
				mocks.mockOpenGraphSchema.EXPECT().CreateSchemaEnvironment(gomock.Any(), int32(3), int32(3), int32(3)).Return(model.SchemaEnvironment{
					Serial: model.Serial{ID: 10},
				}, nil)

				// Principal kinds upsert
				mocks.mockOpenGraphSchema.EXPECT().GetSchemaEnvironmentPrincipalKindsByEnvironmentId(gomock.Any(), int32(10)).Return(nil, errors.New("error"))
			},
			expected: errors.New("error upserting principal kinds: error retrieving existing principal kinds for environment 10: error"),
		},
		{
			name: "Error: DeleteSchemaEnvironmentPrincipalKind fails",
			args: args{
				environment: model.SchemaEnvironment{
					SchemaExtensionId: int32(3),
					EnvironmentKindId: int32(3),
					SourceKindId:      int32(3),
				},
				principalKinds: []model.SchemaEnvironmentPrincipalKind{
					{
						PrincipalKind: int32(3),
					},
				},
			},
			setupMocks: func(t *testing.T, mocks *mocks) {
				t.Helper()
				// Environment validation
				mocks.mockOpenGraphSchema.EXPECT().GetKindById(gomock.Any(), int32(3)).Return(model.Kind{}, nil)
				mocks.mockOpenGraphSchema.EXPECT().GetSourceKinds(gomock.Any()).Return([]database.SourceKind{
					{
						ID:   3,
						Name: graph.StringKind("kind"),
					},
				}, nil)
				// Principal kind validation
				mocks.mockOpenGraphSchema.EXPECT().GetKindById(gomock.Any(), int32(3)).Return(model.Kind{}, nil)

				// Environment upsert
				mocks.mockOpenGraphSchema.EXPECT().GetSchemaEnvironmentById(gomock.Any(), int32(0)).Return(model.SchemaEnvironment{}, database.ErrNotFound)
				mocks.mockOpenGraphSchema.EXPECT().CreateSchemaEnvironment(gomock.Any(), int32(3), int32(3), int32(3)).Return(model.SchemaEnvironment{
					Serial: model.Serial{ID: 10},
				}, nil)

				// Principal kinds upsert
				mocks.mockOpenGraphSchema.EXPECT().GetSchemaEnvironmentPrincipalKindsByEnvironmentId(gomock.Any(), int32(10)).Return([]model.SchemaEnvironmentPrincipalKind{
					{
						EnvironmentId: int32(10),
						PrincipalKind: int32(5),
					},
				}, nil)
				mocks.mockOpenGraphSchema.EXPECT().DeleteSchemaEnvironmentPrincipalKind(gomock.Any(), int32(10), int32(5)).Return(errors.New("error"))
			},
			expected: errors.New("error upserting principal kinds: error deleting principal kind 5 for environment 10: error"),
		},
		{
			name: "Error: CreateSchemaEnvironmentPrincipalKind fails",
			args: args{
				environment: model.SchemaEnvironment{
					SchemaExtensionId: int32(3),
					EnvironmentKindId: int32(3),
					SourceKindId:      int32(3),
				},
				principalKinds: []model.SchemaEnvironmentPrincipalKind{
					{
						PrincipalKind: int32(3),
					},
				},
			},
			setupMocks: func(t *testing.T, mocks *mocks) {
				t.Helper()
				// Environment validation
				mocks.mockOpenGraphSchema.EXPECT().GetKindById(gomock.Any(), int32(3)).Return(model.Kind{}, nil)
				mocks.mockOpenGraphSchema.EXPECT().GetSourceKinds(gomock.Any()).Return([]database.SourceKind{
					{
						ID:   3,
						Name: graph.StringKind("kind"),
					},
				}, nil)
				// Principal kind validation
				mocks.mockOpenGraphSchema.EXPECT().GetKindById(gomock.Any(), int32(3)).Return(model.Kind{}, nil)

				// Environment upsert
				mocks.mockOpenGraphSchema.EXPECT().GetSchemaEnvironmentById(gomock.Any(), int32(0)).Return(model.SchemaEnvironment{}, database.ErrNotFound)
				mocks.mockOpenGraphSchema.EXPECT().CreateSchemaEnvironment(gomock.Any(), int32(3), int32(3), int32(3)).Return(model.SchemaEnvironment{
					Serial: model.Serial{ID: 10},
				}, nil)

				// Principal kinds upsert
				mocks.mockOpenGraphSchema.EXPECT().GetSchemaEnvironmentPrincipalKindsByEnvironmentId(gomock.Any(), int32(10)).Return(nil, database.ErrNotFound)
				mocks.mockOpenGraphSchema.EXPECT().CreateSchemaEnvironmentPrincipalKind(gomock.Any(), int32(10), int32(3)).Return(model.SchemaEnvironmentPrincipalKind{}, errors.New("error"))
			},
			expected: errors.New("error upserting principal kinds: error creating principal kind 3 for environment 10: error"),
		},
		{
			name: "Success: Create new environment with principal kinds",
			args: args{
				environment: model.SchemaEnvironment{
					SchemaExtensionId: int32(3),
					EnvironmentKindId: int32(3),
					SourceKindId:      int32(3),
				},
				principalKinds: []model.SchemaEnvironmentPrincipalKind{
					{
						PrincipalKind: int32(3),
					},
					{
						PrincipalKind: int32(4),
					},
				},
			},
			setupMocks: func(t *testing.T, mocks *mocks) {
				t.Helper()
				// Environment validation
				mocks.mockOpenGraphSchema.EXPECT().GetKindById(gomock.Any(), int32(3)).Return(model.Kind{}, nil)
				mocks.mockOpenGraphSchema.EXPECT().GetSourceKinds(gomock.Any()).Return([]database.SourceKind{
					{
						ID:   3,
						Name: graph.StringKind("kind3"),
					},
					{
						ID:   4,
						Name: graph.StringKind("kind4"),
					},
				}, nil)
				// Principal kind validations
				mocks.mockOpenGraphSchema.EXPECT().GetKindById(gomock.Any(), int32(3)).Return(model.Kind{}, nil)
				mocks.mockOpenGraphSchema.EXPECT().GetKindById(gomock.Any(), int32(4)).Return(model.Kind{}, nil)

				// Environment upsert
				mocks.mockOpenGraphSchema.EXPECT().GetSchemaEnvironmentById(gomock.Any(), int32(0)).Return(model.SchemaEnvironment{}, database.ErrNotFound)
				mocks.mockOpenGraphSchema.EXPECT().CreateSchemaEnvironment(gomock.Any(), int32(3), int32(3), int32(3)).Return(model.SchemaEnvironment{
					Serial: model.Serial{ID: 10},
				}, nil)

				// Principal kinds upsert
				mocks.mockOpenGraphSchema.EXPECT().GetSchemaEnvironmentPrincipalKindsByEnvironmentId(gomock.Any(), int32(10)).Return(nil, database.ErrNotFound)
				mocks.mockOpenGraphSchema.EXPECT().CreateSchemaEnvironmentPrincipalKind(gomock.Any(), int32(10), int32(3)).Return(model.SchemaEnvironmentPrincipalKind{}, nil)
				mocks.mockOpenGraphSchema.EXPECT().CreateSchemaEnvironmentPrincipalKind(gomock.Any(), int32(10), int32(4)).Return(model.SchemaEnvironmentPrincipalKind{}, nil)
			},
			expected: nil,
		},
		{
			name: "Success: Update existing environment and replace principal kinds",
			args: args{
				environment: model.SchemaEnvironment{
					Serial:            model.Serial{ID: 5},
					SchemaExtensionId: int32(3),
					EnvironmentKindId: int32(3),
					SourceKindId:      int32(3),
				},
				principalKinds: []model.SchemaEnvironmentPrincipalKind{
					{
						PrincipalKind: int32(3),
					},
				},
			},
			setupMocks: func(t *testing.T, mocks *mocks) {
				t.Helper()
				// Environment validation
				mocks.mockOpenGraphSchema.EXPECT().GetKindById(gomock.Any(), int32(3)).Return(model.Kind{}, nil)
				mocks.mockOpenGraphSchema.EXPECT().GetSourceKinds(gomock.Any()).Return([]database.SourceKind{
					{
						ID:   3,
						Name: graph.StringKind("kind"),
					},
				}, nil)
				// Principal kind validation
				mocks.mockOpenGraphSchema.EXPECT().GetKindById(gomock.Any(), int32(3)).Return(model.Kind{}, nil)

				// Environment upsert (delete and recreate)
				mocks.mockOpenGraphSchema.EXPECT().GetSchemaEnvironmentById(gomock.Any(), int32(5)).Return(model.SchemaEnvironment{
					Serial: model.Serial{ID: 5},
				}, nil)
				mocks.mockOpenGraphSchema.EXPECT().DeleteSchemaEnvironment(gomock.Any(), int32(5)).Return(nil)
				mocks.mockOpenGraphSchema.EXPECT().CreateSchemaEnvironment(gomock.Any(), int32(3), int32(3), int32(3)).Return(model.SchemaEnvironment{
					Serial: model.Serial{ID: 10},
				}, nil)

				// Principal kinds upsert (delete old, create new)
				mocks.mockOpenGraphSchema.EXPECT().GetSchemaEnvironmentPrincipalKindsByEnvironmentId(gomock.Any(), int32(10)).Return([]model.SchemaEnvironmentPrincipalKind{
					{
						EnvironmentId: int32(10),
						PrincipalKind: int32(99),
					},
				}, nil)
				mocks.mockOpenGraphSchema.EXPECT().DeleteSchemaEnvironmentPrincipalKind(gomock.Any(), int32(10), int32(99)).Return(nil)
				mocks.mockOpenGraphSchema.EXPECT().CreateSchemaEnvironmentPrincipalKind(gomock.Any(), int32(10), int32(3)).Return(model.SchemaEnvironmentPrincipalKind{}, nil)
			},
			expected: nil,
		},
		{
			name: "Success: Create environment with no principal kinds",
			args: args{
				environment: model.SchemaEnvironment{
					SchemaExtensionId: int32(3),
					EnvironmentKindId: int32(3),
					SourceKindId:      int32(3),
				},
				principalKinds: []model.SchemaEnvironmentPrincipalKind{},
			},
			setupMocks: func(t *testing.T, mocks *mocks) {
				t.Helper()
				// Environment validation
				mocks.mockOpenGraphSchema.EXPECT().GetKindById(gomock.Any(), int32(3)).Return(model.Kind{}, nil)
				mocks.mockOpenGraphSchema.EXPECT().GetSourceKinds(gomock.Any()).Return([]database.SourceKind{
					{
						ID:   3,
						Name: graph.StringKind("kind"),
					},
				}, nil)

				// Environment upsert
				mocks.mockOpenGraphSchema.EXPECT().GetSchemaEnvironmentById(gomock.Any(), int32(0)).Return(model.SchemaEnvironment{}, database.ErrNotFound)
				mocks.mockOpenGraphSchema.EXPECT().CreateSchemaEnvironment(gomock.Any(), int32(3), int32(3), int32(3)).Return(model.SchemaEnvironment{
					Serial: model.Serial{ID: 10},
				}, nil)

				// Principal kinds upsert (no existing, no new)
				mocks.mockOpenGraphSchema.EXPECT().GetSchemaEnvironmentPrincipalKindsByEnvironmentId(gomock.Any(), int32(10)).Return(nil, database.ErrNotFound)
			},
			expected: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctrl := gomock.NewController(t)

			mocks := &mocks{
				mockOpenGraphSchema: schemamocks.NewMockOpenGraphSchemaRepository(ctrl),
			}

			tt.setupMocks(t, mocks)

			graphService := opengraphschema.NewOpenGraphSchemaService(mocks.mockOpenGraphSchema)

			err := graphService.UpsertSchemaEnvironmentWithPrincipalKinds(context.Background(), tt.args.environment, tt.args.principalKinds)
			if tt.expected != nil {
				assert.EqualError(t, tt.expected, err.Error())
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
