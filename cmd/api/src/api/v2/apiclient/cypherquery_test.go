package apiclient_test

import (
	"testing"

	v2 "github.com/specterops/bloodhound/src/api/v2"
	"github.com/specterops/bloodhound/src/api/v2/apiclient"
	"github.com/specterops/bloodhound/src/model"
	"go.uber.org/mock/gomock"
)


// TODO: BED-5640 - This is going to require refactoring to abstract the HTTP
func TestManagementResource_CypherQuery(t *testing.T) {
	t.Parallel()

	type mock struct {
		// mockDatabase *dbMocks.MockDatabase
	}
	type args struct {
		request v2.CypherQueryPayload
	}
	type want struct {
		res model.UnifiedGraph
		err error
	}
	type testData struct {
		name             string
		args             args
		emulateWithMocks func(t *testing.T, mock *mock)
		want             want
	}

	tt := []testData{
		// {
		// 	name: "Success",
		// 	args: args{
		// 		v2.CypherQueryPayload{},
		// 	},
		// 	emulateWithMocks: func(t *testing.T, mock *mock) {},
		// 	want: want{
		// 		res: model.UnifiedGraph{},
		// 		err: nil,
		// 	},
		// },
	}
	for _, testCase := range tt {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			_ = gomock.NewController(t)

			mocks := &mock{

			}

			testCase.emulateWithMocks(t, mocks)

			client := apiclient.Client{

			}
			client.CypherQuery(testCase.args.request)
		})
	}
}
