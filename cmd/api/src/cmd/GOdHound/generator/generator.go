package generator

type GenericIngestNode struct {
	ID         string         `json:"id"`
	Kinds      []string       `json:"kinds"`
	Properties map[string]any `json:"properties"`
}

type GenericIngestEndpoint struct {
	MatchBy string `json:"match_by"`
	Kind    string `json:"kind"`
	Value   string `json:"value"`
}

type GenericIngestEdge struct {
	Kind       string                `json:"kind"`
	Start      GenericIngestEndpoint `json:"start"`
	End        GenericIngestEndpoint `json:"end"`
	Properties map[string]any        `json:"properties"`
}

type GenericIngestGraph struct {
	Nodes []GenericIngestNode `json:"nodes"`
	Edges []GenericIngestEdge `json:"edges"`
}

type GenericIngestPayload struct {
	Graph GenericIngestGraph `json:"graph"`
}

// https://learn.microsoft.com/en-us/windows-server/identity/ad-ds/manage/understand-security-identifiers
func newDomainSID() (string, error) {
	return "", nil
}

func MakeDomain() (GenericIngestPayload, error) {
	domain := GenericIngestNode{
		ID:    "1234",
		Kinds: []string{"Domain", "Base"},
		Properties: map[string]any{
			// This MUST match the ID value of the GenericIngestNode otherwise it will fail validation
			"objectid": "1234",
		},
	}

	return GenericIngestPayload{
		Graph: GenericIngestGraph{
			Nodes: []GenericIngestNode{domain},
		},
	}, nil
}
