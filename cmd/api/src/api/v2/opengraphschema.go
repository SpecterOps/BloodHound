package v2

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/specterops/bloodhound/cmd/api/src/api"
	"github.com/specterops/bloodhound/cmd/api/src/model"
)

func (s Resources) OpenGraphSchemaIngest(response http.ResponseWriter, request *http.Request) {
	var (
		ctx         = request.Context()
		err         error
		graphSchema model.GraphSchema
	)

	err = json.NewDecoder(request.Body).Decode(&graphSchema)
	if err != nil {
		// return 400
		api.WriteErrorResponse(ctx, api.BuildErrorResponse(http.StatusBadRequest, fmt.Sprintf("unable to parse opengraph schema: %v", err), request), response)
		return
	}

	/*
		1. payload hits endpoint and assume valid auth and method
		2. determine schema extension format based on content-type header
		3. either parse file or parse json schema into schema_extension api model
		    - can be json file, form data or zip with json file
		    - cant decode = 400
		4. validate schema extension api model
		   - ensure it has a non-empty extension
		   - ensure node/edge kind and property slices arent empty
		5. Pass extension schema to service layer
	*/

	err = s.OpenGraphSchemaService.UpsertGraphSchemaExtension(ctx, graphSchema)
	if err != nil {
		api.WriteErrorResponse(ctx, api.BuildErrorResponse(http.StatusInternalServerError, fmt.Sprintf("unable to update graph schema: %v", err), request), response)
		return
	}
	response.WriteHeader(http.StatusCreated)
}
