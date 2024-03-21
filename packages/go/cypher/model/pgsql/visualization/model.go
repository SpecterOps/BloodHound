package visualization

type Graph struct {
	Title         string         `json:"-"`
	Style         Style          `json:"style"`
	Nodes         []Node         `json:"nodes"`
	Relationships []Relationship `json:"relationships"`
}

type Node struct {
	Caption    string         `json:"caption"`
	ID         string         `json:"id"`
	Labels     []string       `json:"labels"`
	Properties map[string]any `json:"properties"`
	Position   Position       `json:"position"`
	Style      Style          `json:"style"`
}

type Relationship struct {
	ID         string         `json:"id"`
	FromID     string         `json:"fromId"`
	ToID       string         `json:"toId"`
	Type       string         `json:"type"`
	Properties map[string]any `json:"properties"`
	Style      Style          `json:"style"`
}

type Style struct{}

type Position struct {
	X int `json:"x"`
	Y int `json:"y"`
}
