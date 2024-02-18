package cve

import "encoding/json"

type Entry struct {
	Containers  Containers `json:"containers"`
	Metadata    Metadata   `json:"cveMetadata"`
	DataType    string     `json:"dataType"`
	DataVersion string     `json:"dataVersion"`
}

type NumberingAuthority struct {
	Affected         []Affected                      `json:"affected"`
	Credits          []Credit                        `json:"credits"`
	DatePublic       string                          `json:"datePublic"`
	Descriptions     []NumberingAuthorityDescription `json:"descriptions"`
	Exploits         []NumberingAuthorityDescription `json:"exploits"`
	Metrics          []Metric                        `json:"metrics"`
	ProblemTypes     []ProblemType                   `json:"problemTypes"`
	ProviderMetadata ProviderMetadata                `json:"providerMetadata"`
	References       []Reference                     `json:"references"`
	Solutions        []NumberingAuthorityDescription `json:"solutions"`
	Source           Source                          `json:"source"`
	Timeline         []TimelineEntry                 `json:"timeline"`
	Title            string                          `json:"title"`
	Workarounds      []NumberingAuthorityDescription `json:"workarounds"`
	XGenerator       json.RawMessage                 `json:"x_generator"`
}

type Metadata struct {
	AssignerOrgID     string `json:"assignerOrgId"`
	AssignerShortName string `json:"assignerShortName"`
	ID                string `json:"cveId"`
	DatePublished     string `json:"datePublished"`
	DateReserved      string `json:"dateReserved"`
	DateUpdated       string `json:"dateUpdated"`
	State             string `json:"state"`
}

type Change struct {
	At     string `json:"at"`
	Status string `json:"status"`
}

type ScoringSystemV31 struct {
	AttackComplexity      string  `json:"attackComplexity"`
	AttackVector          string  `json:"attackVector"`
	AvailabilityImpact    string  `json:"availabilityImpact"`
	BaseScore             float64 `json:"baseScore"`
	BaseSeverity          string  `json:"baseSeverity"`
	ConfidentialityImpact string  `json:"confidentialityImpact"`
	IntegrityImpact       string  `json:"integrityImpact"`
	PrivilegesRequired    string  `json:"privilegesRequired"`
	Scope                 string  `json:"scope"`
	UserInteraction       string  `json:"userInteraction"`
	VectorString          string  `json:"vectorString"`
	Version               string  `json:"version"`
}

type SupportingMedia struct {
	Base64 bool   `json:"base64"`
	Type   string `json:"type"`
	Value  string `json:"value"`
}

type Version struct {
	Changes     []Change `json:"changes"`
	LessThan    string   `json:"lessThan"`
	Status      string   `json:"status"`
	Version     string   `json:"version"`
	VersionType string   `json:"versionType"`
}

type Containers struct {
	NumberingAuthority NumberingAuthority `json:"cna"`
}

type Metric struct {
	Scoring   ScoringSystemV31 `json:"cvssV3_1"`
	Format    string           `json:"format"`
	Scenarios []Scenario       `json:"scenarios"`
}

type ProblemTypeDescription struct {
	CWEID       string `json:"cweId"`
	Description string `json:"description"`
	Lang        string `json:"lang"`
	Type        string `json:"type"`
}

type ProviderMetadata struct {
	DateUpdated string `json:"dateUpdated"`
	OrgID       string `json:"orgId"`
	ShortName   string `json:"shortName"`
}

type Affected struct {
	DefaultStatus string    `json:"defaultStatus"`
	Platforms     []string  `json:"platforms"`
	Product       string    `json:"product"`
	Vendor        string    `json:"vendor"`
	Versions      []Version `json:"versions"`
}

type Source struct {
	Defect    json.RawMessage `json:"defect"`
	Discovery string          `json:"discovery"`
}

type ProblemType struct {
	Descriptions []ProblemTypeDescription `json:"descriptions"`
}

type XGenerator struct {
	Engine string `json:"engine"`
}

type NumberingAuthorityDescription struct {
	Lang            string            `json:"lang"`
	SupportingMedia []SupportingMedia `json:"supportingMedia"`
	Value           string            `json:"value"`
}

type TimelineEntry struct {
	Lang  string `json:"lang"`
	Time  string `json:"time"`
	Value string `json:"value"`
}

type Credit struct {
	Lang  string `json:"lang"`
	Type  string `json:"type"`
	Value string `json:"value"`
}

type Scenario struct {
	Lang  string `json:"lang"`
	Value string `json:"value"`
}

type Reference struct {
	URL string `json:"url"`
}
