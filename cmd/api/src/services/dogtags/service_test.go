// service_test.go
package dogtags

import (
	"testing"
)

func TestService(t *testing.T) {
	svc := NewDefaultService()

	// Test typed getters return defaults
	tierLimit := svc.GetFlagAsInt(PZ_TIER_LIMIT)
	if tierLimit != 1 {
		t.Errorf("expected 1, got %d", tierLimit)
	}

	labelLimit := svc.GetFlagAsInt(PZ_LABEL_LIMIT)
	if labelLimit != 10 {
		t.Errorf("expected 10, got %d", labelLimit)
	}

	multiTier := svc.GetFlagAsBool(PZ_MULTI_TIER_ANALYSIS)
	if multiTier != false {
		t.Errorf("expected false, got %v", multiTier)
	}

	// Test GetAllDogTags
	all := svc.GetAllDogTags()
	t.Logf("All dogtags: %+v", all)
}
