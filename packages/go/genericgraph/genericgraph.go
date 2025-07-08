package genericgraph

type Node struct {
	ID         string         `json:"id"`
	Kinds      []string       `json:"kinds"`
	Properties map[string]any `json:"properties"`
}

type Terminal struct {
	MatchBy string `json:"match_by"`
	Value   string `json:"value"`
}

type Edge struct {
	Start      Terminal       `json:"start"`
	End        Terminal       `json:"end"`
	Kind       string         `json:"kind"`
	Properties map[string]any `json:"properties"`
}

type Graph struct {
	Nodes []Node `json:"nodes"`
	Edges []Edge `json:"edges"`
}

type GenericObject struct {
	Graph Graph `json:"graph"`
}
