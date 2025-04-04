package query

import (
	"github.com/specterops/bloodhound/dawgs/graph"
)

// TODO Cleanup after Tiering GA
func SearchTierZeroNodes(tieringEnabled bool) graph.Criteria {
	if tieringEnabled {
		return Kind(Node(), graph.StringKind("Tag_Tier_Zero"))
	} else {
		return StringContains(NodeProperty("system_tags"), "admin_tier_0")
	}
}

// TODO Cleanup after Tiering GA
func SearchTierZeroNodesRel(tieringEnabled bool) graph.Criteria {
	if tieringEnabled {
		return Kind(Start(), graph.StringKind("Tag_Tier_Zero"))
	} else {
		return StringContains(StartProperty("system_tags"), "admin_tier_0")
	}
}
