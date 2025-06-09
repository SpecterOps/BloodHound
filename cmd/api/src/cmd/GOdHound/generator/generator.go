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

type ADUser struct {
	Properties              map[string]any `json:"properties"`
	domain                  string         // "PHANTOM.CORP"
	name                    string         // "STEPHEN@PHANTOM.CORP",
	distinguishedname       string         // "CN=DC01.PHANTOM.CORP,CN=USERS,DC=PHANTOM,DC=CORP",
	domainsid               string         // "S-1-5-21-2697957641-2271029196-387917394",
	samaccountname          string         // "stephen",
	isaclprotected          bool           //": true,
	description             string         // null,
	whencreated             int64          // 1699347828,
	sensitive               bool           // false,
	dontreqpreauth          bool           // false,
	passwordnotreqd         bool           // false,
	unconstraineddelegation bool           // false,
	pwdneverexpires         bool           // true,
	enabled                 bool           // true,
	trustedtoauth           bool           // false,
	lastlogon               int64          // 1708508097,
	lastlogontimestamp      int64          // 1708507965,
	pwdlastset              int64          // 1699376628,
	serviceprincipalnames   []string       // [],
	hasspn                  bool           // false,
	displayname             string         // "stephen",
	email                   string         // null,
	title                   string         // null,
	homedirectory           string         // null,
	userpassword            string         // null,
	unixpassword            string         // null,
	unicodepassword         string         // null,
	sfupassword             string         // null,
	logonscript             string         // null,
	admincount              bool           // false,
	sidhistory              []string       //": []
}
type ADComputer struct {
	Properties              map[string]any `json:"properties"`
	domain                  string         // "PHANTOM.CORP",
	name                    string         // "ALICE-LAPTOP.PHANTOM.CORP",
	distinguishedname       string         // "CN=ALICE-LAPTOP,OU=WORKSTATIONS,DC=PHANTOM,DC=CORP",
	domainsid               string         // "S-1-5-21-2697957641-2271029196-387917394",
	samaccountname          string         // "ALICE-LAPTOP$",
	haslaps                 string         // false,
	isaclprotected          bool           // false,
	description             string         // null,
	whencreated             int64          // 1681261713,
	enabled                 bool           //true,
	unconstraineddelegation bool           // false,
	trustedtoauth           bool           // false,
	isdc                    bool           // false,
	lastlogon               int64          // 0,
	lastlogontimestamp      int64          // -1,
	pwdlastset              int64          // 1681286913,
	serviceprincipalnames   []string       // [],
	email                   string         // null,
	operatingsystem         string         // null,
	sidhistory              []string       // []
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

func AddADUser() ADUser {
	user := ADUser{
		domain:  "SCOUBI.LAB",
		name:    "MEGATEST@SCOUBI.LAB",
		enabled: true,
	}
	return user
}

func AddADComputer() ADComputer {
	computer := ADComputer{
		domain:  "SCOUBI.LAB",
		name:    "COMP00001@SCOUBI.LAB",
		enabled: true,
	}
	return computer
}
