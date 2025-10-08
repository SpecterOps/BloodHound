package v2_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	v2 "github.com/specterops/bloodhound/cmd/api/src/api/v2"
	"github.com/specterops/bloodhound/cmd/api/src/services/dogtags"
)

// Mock DogtagsService for testing
type mockDogtagsService struct {
	flags map[dogtags.FlagKey]interface{}
}

func (m *mockDogtagsService) GetAllFlags(ctx context.Context) map[dogtags.FlagKey]interface{} {
	return m.flags
}


func TestGetDogtags(t *testing.T) {
	// Setup mock provider
	mockProvider := &mockDogtagsService{
		flags: map[dogtags.FlagKey]interface{}{
			dogtags.BypassCypherQueryLimits: true,
			dogtags.CypherMutability:        false,
			dogtags.ZoneAllocation:          int64(100),
			dogtags.LabelAllocation:         int64(50),
		},
	}

	// Create resources with service using mock provider
	resources := v2.Resources{
		DogtagsService: mockProvider,
	}

	// Create test request
	req := httptest.NewRequest("GET", "/api/v2/dogtags", nil)
	w := httptest.NewRecorder()

	// Call the handler
	resources.GetDogtags(w, req)

	// Check response
	resp := w.Result()
	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}

	// Parse response body - WriteBasicResponse wraps the flags
	var rawResponse map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&rawResponse); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	// Extract the actual flags data
	flagsData, ok := rawResponse["data"].(map[string]interface{})
	if !ok {
		t.Fatalf("Expected 'data' field with flags, got: %+v", rawResponse)
	}

	// Verify response data - JSON unmarshaling converts numbers to float64
	expectedFlags := map[string]interface{}{
		"bypass_cypher_query_limits": true,
		"cypher_mutability":          false,
		"zone_allocation":            float64(100),
		"label_allocation":           float64(50),
	}

	for key, expectedValue := range expectedFlags {
		if actualValue, exists := flagsData[key]; !exists {
			t.Errorf("Expected flag %s to exist in response", key)
		} else if actualValue != expectedValue {
			t.Errorf("Expected flag %s to be %v, got %v", key, expectedValue, actualValue)
		}
	}
}

func TestGetDogtagsServiceNotAvailable(t *testing.T) {
	// Create resources without dogtags service (nil)
	resources := v2.Resources{
		DogtagsService: nil,
	}

	// Create test request
	req := httptest.NewRequest("GET", "/api/v2/dogtags", nil)
	w := httptest.NewRecorder()

	// Call the handler
	resources.GetDogtags(w, req)

	// Check response
	resp := w.Result()
	if resp.StatusCode != http.StatusServiceUnavailable {
		t.Errorf("Expected status 503, got %d", resp.StatusCode)
	}
}