package v2_test

import (
	"fmt"
	uuid2 "github.com/gofrs/uuid"
	"github.com/gorilla/mux"
	"github.com/specterops/bloodhound/headers"
	"github.com/specterops/bloodhound/mediatypes"
	"github.com/specterops/bloodhound/src/api"
	v2 "github.com/specterops/bloodhound/src/api/v2"
	"github.com/specterops/bloodhound/src/auth"
	"github.com/specterops/bloodhound/src/ctx"
	"github.com/specterops/bloodhound/src/database/mocks"
	"github.com/specterops/bloodhound/src/model"
	"github.com/specterops/bloodhound/src/test/must"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
)

func TestResources_ListSavedQueries_SortingError(t *testing.T) {
	var (
		mockCtrl  = gomock.NewController(t)
		resources = v2.Resources{}
	)
	defer mockCtrl.Finish()

	endpoint := "/api/v2/saved-queries"

	if req, err := http.NewRequest("GET", endpoint, nil); err != nil {
		t.Fatal(err)
	} else {
		q := url.Values{}
		q.Add("sort_by", "invalidColumn")

		req.Header.Set(headers.ContentType.String(), mediatypes.ApplicationJson.String())
		req.URL.RawQuery = q.Encode()

		router := mux.NewRouter()
		router.HandleFunc(endpoint, resources.ListSavedQueries).Methods("GET")

		response := httptest.NewRecorder()
		router.ServeHTTP(response, req)
		require.Equal(t, http.StatusBadRequest, response.Code)

		require.Contains(t, response.Body.String(), api.ErrorResponseDetailsNotSortable)
	}
}

func TestResources_ListSavedQueries_InvalidFilterColumn(t *testing.T) {
	var (
		mockCtrl  = gomock.NewController(t)
		resources = v2.Resources{}
	)
	defer mockCtrl.Finish()

	endpoint := "/api/v2/saved-queries"

	if req, err := http.NewRequest("GET", endpoint, nil); err != nil {
		t.Fatal(err)
	} else {
		q := url.Values{}
		q.Add("foo", "gt:0")

		req.Header.Set(headers.ContentType.String(), mediatypes.ApplicationJson.String())
		req.URL.RawQuery = q.Encode()

		router := mux.NewRouter()
		router.HandleFunc(endpoint, resources.ListSavedQueries).Methods("GET")

		response := httptest.NewRecorder()
		router.ServeHTTP(response, req)
		require.Equal(t, http.StatusBadRequest, response.Code)

		require.Contains(t, response.Body.String(), "column cannot be filtered")
	}
}

func TestResources_ListSavedQueries_InvalidFilterPredicate(t *testing.T) {
	var (
		mockCtrl  = gomock.NewController(t)
		resources = v2.Resources{}
	)
	defer mockCtrl.Finish()

	endpoint := "/api/v2/saved-queries"

	if req, err := http.NewRequest("GET", endpoint, nil); err != nil {
		t.Fatal(err)
	} else {
		q := url.Values{}
		q.Add("name", "gt:0")

		req.Header.Set(headers.ContentType.String(), mediatypes.ApplicationJson.String())
		req.URL.RawQuery = q.Encode()

		router := mux.NewRouter()
		router.HandleFunc(endpoint, resources.ListSavedQueries).Methods("GET")

		response := httptest.NewRecorder()
		router.ServeHTTP(response, req)
		require.Equal(t, http.StatusBadRequest, response.Code)
		require.Contains(t, response.Body.String(), api.ErrorResponseDetailsFilterPredicateNotSupported)
	}
}

func TestResources_ListSavedQueries_InvalidSkip(t *testing.T) {
	var (
		mockCtrl  = gomock.NewController(t)
		resources = v2.Resources{}
	)
	defer mockCtrl.Finish()

	endpoint := "/api/v2/saved-queries"

	if req, err := http.NewRequest("GET", endpoint, nil); err != nil {
		t.Fatal(err)
	} else {
		q := url.Values{}
		q.Add("skip", "-1")

		req.Header.Set(headers.ContentType.String(), mediatypes.ApplicationJson.String())
		req.URL.RawQuery = q.Encode()

		router := mux.NewRouter()
		router.HandleFunc(endpoint, resources.ListSavedQueries).Methods("GET")

		response := httptest.NewRecorder()
		router.ServeHTTP(response, req)
		require.Equal(t, http.StatusBadRequest, response.Code)
		require.Contains(t, response.Body.String(), "invalid skip")
	}
}

func TestResources_ListSavedQueries_InvalidLimit(t *testing.T) {
	var (
		mockCtrl  = gomock.NewController(t)
		resources = v2.Resources{}
	)
	defer mockCtrl.Finish()

	endpoint := "/api/v2/saved-queries"

	if req, err := http.NewRequest("GET", endpoint, nil); err != nil {
		t.Fatal(err)
	} else {
		q := url.Values{}
		q.Add("limit", "-1")

		req.Header.Set(headers.ContentType.String(), mediatypes.ApplicationJson.String())
		req.URL.RawQuery = q.Encode()

		router := mux.NewRouter()
		router.HandleFunc(endpoint, resources.ListSavedQueries).Methods("GET")

		response := httptest.NewRecorder()
		router.ServeHTTP(response, req)
		require.Equal(t, http.StatusBadRequest, response.Code)
		require.Contains(t, response.Body.String(), "invalid limit")
	}
}

func TestResources_ListSavedQueries_DBError(t *testing.T) {
	var (
		mockCtrl  = gomock.NewController(t)
		mockDB    = mocks.NewMockDatabase(mockCtrl)
		resources = v2.Resources{DB: mockDB}
	)
	defer mockCtrl.Finish()

	endpoint := "/api/v2/saved-queries"
	userId, err := uuid2.NewV4()
	require.Nil(t, err)

	bhCtx := ctx.Context{
		RequestID: "",
		AuthCtx: auth.Context{
			Session: model.UserSession{
				User:   model.User{},
				UserID: userId,
			},
		},
		Host: nil,
	}
	goContext := bhCtx.ConstructGoContext()

	mockDB.EXPECT().ListSavedQueries(userId, "", model.SQLFilter{}, 0, 10000).Return(model.SavedQueries{}, 0, fmt.Errorf("foo"))

	if req, err := http.NewRequest("GET", endpoint, nil); err != nil {
		t.Fatal(err)
	} else {
		req = req.WithContext(goContext)
		req.Header.Set(headers.ContentType.String(), mediatypes.ApplicationJson.String())

		router := mux.NewRouter()
		router.HandleFunc(endpoint, resources.ListSavedQueries).Methods("GET")

		response := httptest.NewRecorder()
		router.ServeHTTP(response, req)
		require.Equal(t, http.StatusInternalServerError, response.Code)
		require.Contains(t, response.Body.String(), api.ErrorResponseDetailsInternalServerError)
	}
}

func TestResources_ListSavedQueries(t *testing.T) {
	var (
		mockCtrl  = gomock.NewController(t)
		mockDB    = mocks.NewMockDatabase(mockCtrl)
		resources = v2.Resources{DB: mockDB}
	)
	defer mockCtrl.Finish()

	endpoint := "/api/v2/saved-queries"
	userId, err := uuid2.NewV4()
	require.Nil(t, err)

	bhCtx := ctx.Context{
		RequestID: "",
		AuthCtx: auth.Context{
			Session: model.UserSession{
				User:   model.User{},
				UserID: userId,
			},
		},
		Host: nil,
	}
	goContext := bhCtx.ConstructGoContext()

	mockDB.EXPECT().ListSavedQueries(userId, gomock.Any(), gomock.Any(), 1, 10).Return(model.SavedQueries{
		{
			UserID: userId.String(),
			Name:   "myQuery",
			Query:  "Match(n) return n;",
		},
	}, 1, nil)

	if req, err := http.NewRequest("GET", endpoint, nil); err != nil {
		t.Fatal(err)
	} else {
		req = req.WithContext(goContext)
		q := url.Values{}
		q.Add("sort_by", "-name")
		q.Add("name", "eq:myQuery")
		q.Add("skip", "1")
		q.Add("limit", "10")

		req.Header.Set(headers.ContentType.String(), mediatypes.ApplicationJson.String())
		req.URL.RawQuery = q.Encode()

		router := mux.NewRouter()
		router.HandleFunc(endpoint, resources.ListSavedQueries).Methods("GET")

		response := httptest.NewRecorder()
		router.ServeHTTP(response, req)
		require.Equal(t, http.StatusOK, response.Code)
	}
}

func TestResources_CreateSavedQuery_InvalidBody(t *testing.T) {
	var (
		mockCtrl  = gomock.NewController(t)
		mockDB    = mocks.NewMockDatabase(mockCtrl)
		resources = v2.Resources{DB: mockDB}
	)
	defer mockCtrl.Finish()

	endpoint := "/api/v2/saved-queries"
	userId, err := uuid2.NewV4()
	require.Nil(t, err)

	bhCtx := ctx.Context{
		RequestID: "",
		AuthCtx: auth.Context{
			Session: model.UserSession{
				User:   model.User{},
				UserID: userId,
			},
		},
		Host: nil,
	}
	goContext := bhCtx.ConstructGoContext()

	payload := "foobar"

	req, err := http.NewRequest("POST", endpoint, must.MarshalJSONReader(payload))
	require.Nil(t, err)

	req = req.WithContext(goContext)
	req.Header.Set(headers.ContentType.String(), mediatypes.ApplicationJson.String())

	router := mux.NewRouter()
	router.HandleFunc(endpoint, resources.CreateSavedQuery).Methods("POST")

	response := httptest.NewRecorder()
	router.ServeHTTP(response, req)
	require.Equal(t, http.StatusBadRequest, response.Code)
}

func TestResources_CreateSavedQuery_EmptyBody(t *testing.T) {
	var (
		mockCtrl  = gomock.NewController(t)
		mockDB    = mocks.NewMockDatabase(mockCtrl)
		resources = v2.Resources{DB: mockDB}
	)
	defer mockCtrl.Finish()

	endpoint := "/api/v2/saved-queries"
	userId, err := uuid2.NewV4()
	require.Nil(t, err)

	bhCtx := ctx.Context{
		RequestID: "",
		AuthCtx: auth.Context{
			Session: model.UserSession{
				User:   model.User{},
				UserID: userId,
			},
		},
		Host: nil,
	}
	goContext := bhCtx.ConstructGoContext()

	payload := v2.CreateSavedQueryRequest{}

	req, err := http.NewRequest("POST", endpoint, must.MarshalJSONReader(payload))
	require.Nil(t, err)

	req = req.WithContext(goContext)
	req.Header.Set(headers.ContentType.String(), mediatypes.ApplicationJson.String())

	router := mux.NewRouter()
	router.HandleFunc(endpoint, resources.CreateSavedQuery).Methods("POST")

	response := httptest.NewRecorder()
	router.ServeHTTP(response, req)
	require.Equal(t, http.StatusBadRequest, response.Code)
	require.Contains(t, response.Body.String(), "field is empty")
}

func TestResources_CreateSavedQuery_DuplicateName(t *testing.T) {
	var (
		mockCtrl  = gomock.NewController(t)
		mockDB    = mocks.NewMockDatabase(mockCtrl)
		resources = v2.Resources{DB: mockDB}
	)
	defer mockCtrl.Finish()

	endpoint := "/api/v2/saved-queries"
	userId, err := uuid2.NewV4()
	require.Nil(t, err)

	bhCtx := ctx.Context{
		RequestID: "",
		AuthCtx: auth.Context{
			Session: model.UserSession{
				User:   model.User{},
				UserID: userId,
			},
		},
		Host: nil,
	}
	goContext := bhCtx.ConstructGoContext()

	mockDB.EXPECT().CreateSavedQuery(gomock.Any(), gomock.Any(), gomock.Any()).Return(model.SavedQuery{}, fmt.Errorf("duplicate key value violates unique constraint \"idx_saved_queries_composite_index\""))

	payload := v2.CreateSavedQueryRequest{
		Query: "Match(n) return n",
		Name:  "myQuery",
	}

	req, err := http.NewRequest("POST", endpoint, must.MarshalJSONReader(payload))
	require.Nil(t, err)

	req = req.WithContext(goContext)
	req.Header.Set(headers.ContentType.String(), mediatypes.ApplicationJson.String())

	router := mux.NewRouter()
	router.HandleFunc(endpoint, resources.CreateSavedQuery).Methods("POST")

	response := httptest.NewRecorder()
	router.ServeHTTP(response, req)
	require.Equal(t, http.StatusBadRequest, response.Code)
	require.Contains(t, response.Body.String(), "duplicate")
}

func TestResources_CreateSavedQuery_CreateFailure(t *testing.T) {
	var (
		mockCtrl  = gomock.NewController(t)
		mockDB    = mocks.NewMockDatabase(mockCtrl)
		resources = v2.Resources{DB: mockDB}
	)
	defer mockCtrl.Finish()

	endpoint := "/api/v2/saved-queries"
	userId, err := uuid2.NewV4()
	require.Nil(t, err)

	bhCtx := ctx.Context{
		RequestID: "",
		AuthCtx: auth.Context{
			Session: model.UserSession{
				User:   model.User{},
				UserID: userId,
			},
		},
		Host: nil,
	}
	goContext := bhCtx.ConstructGoContext()

	payload := v2.CreateSavedQueryRequest{
		Query: "Match(n) return n",
		Name:  "myCustomQuery1",
	}

	mockDB.EXPECT().CreateSavedQuery(userId, payload.Name, payload.Query).Return(model.SavedQuery{}, fmt.Errorf("foo"))

	req, err := http.NewRequest("POST", endpoint, must.MarshalJSONReader(payload))
	require.Nil(t, err)

	req = req.WithContext(goContext)
	req.Header.Set(headers.ContentType.String(), mediatypes.ApplicationJson.String())

	router := mux.NewRouter()
	router.HandleFunc(endpoint, resources.CreateSavedQuery).Methods("POST")

	response := httptest.NewRecorder()
	router.ServeHTTP(response, req)
	require.Equal(t, http.StatusInternalServerError, response.Code)
}

func TestResources_CreateSavedQuery(t *testing.T) {
	var (
		mockCtrl  = gomock.NewController(t)
		mockDB    = mocks.NewMockDatabase(mockCtrl)
		resources = v2.Resources{DB: mockDB}
	)
	defer mockCtrl.Finish()

	endpoint := "/api/v2/saved-queries"
	userId, err := uuid2.NewV4()
	require.Nil(t, err)

	bhCtx := ctx.Context{
		RequestID: "",
		AuthCtx: auth.Context{
			Session: model.UserSession{
				User:   model.User{},
				UserID: userId,
			},
		},
		Host: nil,
	}
	goContext := bhCtx.ConstructGoContext()

	payload := v2.CreateSavedQueryRequest{
		Query: "Match(n) return n",
		Name:  "myCustomQuery1",
	}

	mockDB.EXPECT().CreateSavedQuery(userId, payload.Name, payload.Query).Return(model.SavedQuery{
		UserID: userId.String(),
		Name:   payload.Name,
		Query:  payload.Query,
	}, nil)

	req, err := http.NewRequest("POST", endpoint, must.MarshalJSONReader(payload))
	require.Nil(t, err)

	req = req.WithContext(goContext)
	req.Header.Set(headers.ContentType.String(), mediatypes.ApplicationJson.String())

	router := mux.NewRouter()
	router.HandleFunc(endpoint, resources.CreateSavedQuery).Methods("POST")

	response := httptest.NewRecorder()
	router.ServeHTTP(response, req)
	require.Equal(t, http.StatusCreated, response.Code)
}
