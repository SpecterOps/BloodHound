package dogtags

// Typed flag keys - each type can only be used with its matching getter
type BoolDogTag string
type StringDogTag string
type IntDogTag string
type FloatDogTag string

type BoolDogTagSpec struct {
	Description string `json:"description,omitempty"`
	Default     bool   `json:"default"`
}

type StringDogTagSpec struct {
	Description string `json:"description,omitempty"`
	Default     string `json:"default"`
}

type IntDogTagSpec struct {
	Description string `json:"description,omitempty"`
	Default     int64  `json:"default"`
}

type FloatDogTagSpec struct {
	Description string  `json:"description,omitempty"`
	Default     float64 `json:"default"`
}

const (
	PZ_MULTI_TIER_ANALYSIS BoolDogTag = "privilege_zones.multi_tier_analysis"

	PZ_TIER_LIMIT  IntDogTag = "privilege_zones.tier_limit"
	PZ_LABEL_LIMIT IntDogTag = "privilege_zones.label_limit"
)

var AllBoolDogTags = map[BoolDogTag]BoolDogTagSpec{
	PZ_MULTI_TIER_ANALYSIS: {Description: "PZ Multi Tier Analysis", Default: false},
}

var AllIntDogTags = map[IntDogTag]IntDogTagSpec{
	PZ_TIER_LIMIT:  {Description: "PZ Tier Limit", Default: 1},
	PZ_LABEL_LIMIT: {Description: "PZ Label Limit", Default: 10},
}

var AllStringDogTags = map[StringDogTag]StringDogTagSpec{
	// none yet
}

var AllFloatDogTags = map[FloatDogTag]FloatDogTagSpec{
	// none yet
}
