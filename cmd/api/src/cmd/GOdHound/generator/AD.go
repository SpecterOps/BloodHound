package generator

import (
	"fmt"
	"math/rand"
)

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
	Serviceprincipalnames   []string       // [],
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
	Sidhistory              []string       //": []
	objectid                string
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
	Serviceprincipalnames   []string       // [],
	email                   string         // null,
	operatingsystem         string         // null,
	Sidhistory              []string       // []
	objectid                string
}

type ADEdge struct {
	Kind       string
	Start      string
	End        string
	Properties map[string]any `json:"properties"`
}

// https://learn.microsoft.com/en-us/windows-server/identity/ad-ds/manage/understand-security-identifiers
func newDomainSID() (string, error) {
	// Need to add logic to make this random
	return "S-1-5-21-2697957666-2271029666-387917666", nil
}

func AddADUser(r *rand.Rand, myDomain string, x int32) ADUser {
	first := getRandomName(r, "ressources/firstname.json")
	last := getRandomName(r, "ressources/lastname.json")
	oid := fmt.Sprintf("%07d", x)
	user := ADUser{
		domain:                myDomain,
		name:                  first,
		displayname:           fmt.Sprintf(first + " " + last),
		objectid:              oid,
		enabled:               true,
		Serviceprincipalnames: []string{},
		Sidhistory:            []string{},
	}
	return user
}

func (u ADUser) ToGraphNode() GenericIngestNode {
	//return GraphNode{
	return GenericIngestNode{
		ID:    u.objectid, // Using objectid as unique identifier
		Kinds: []string{"User", "Base"},
		Properties: map[string]interface{}{
			"displayname":             u.displayname,
			"name":                    u.name,
			"objectid":                u.objectid,
			"domain":                  u.domain,
			"distinguishedname":       u.distinguishedname,
			"samaccountname":          u.samaccountname,
			"enabled":                 u.enabled,
			"admincount":              u.admincount,
			"email":                   u.email,
			"title":                   u.title,
			"lastlogon":               u.lastlogon,
			"lastlogontimestamp":      u.lastlogontimestamp,
			"pwdlastset":              u.pwdlastset,
			"serviceprincipalnames":   u.Serviceprincipalnames,
			"sidhistory":              u.Sidhistory,
			"unconstraineddelegation": u.unconstraineddelegation,
			"trustedtoauth":           u.trustedtoauth,
			"pwdneverexpires":         u.pwdneverexpires,
		},
	}
}

func AddADComputer(r *rand.Rand, myDomain string, x int32) ADComputer {
	oid := fmt.Sprintf("%07d", x)
	uName := fmt.Sprintf("COMP"+"%07d"+"."+myDomain, x) //strconv.Itoa(int(x))
	computer := ADComputer{
		domain:                myDomain,
		name:                  uName, //"COMP00001@SCOUBI.LAB",
		objectid:              oid,
		enabled:               true,
		Serviceprincipalnames: []string{},
		Sidhistory:            []string{},
	}
	return computer
}

func (c ADComputer) ToGraphNode() GenericIngestNode {
	//return GraphNode{
	return GenericIngestNode{
		ID:    c.objectid,
		Kinds: []string{"Computer", "Base"},
		Properties: map[string]interface{}{
			"name":                    c.name,
			"displayname":             c.name,
			"objectid":                c.objectid,
			"domain":                  c.domain,
			"distinguishedname":       c.distinguishedname,
			"samaccountname":          c.samaccountname,
			"haslaps":                 c.haslaps,
			"enabled":                 c.enabled,
			"description":             c.description,
			"whencreated":             c.whencreated,
			"unconstraineddelegation": c.unconstraineddelegation,
			"trustedtoauth":           c.trustedtoauth,
			"isdc":                    c.isdc,
			"lastlogon":               c.lastlogon,
			"lastlogontimestamp":      c.lastlogontimestamp,
			"pwdlastset":              c.pwdlastset,
			"serviceprincipalnames":   c.Serviceprincipalnames,
			"operatingsystem":         c.operatingsystem,
			"sidhistory":              c.Sidhistory,
		},
	}
}

func AddADEdge(i int) ADEdge {
	x := rand.Intn(i-1) + 1
	sNode := fmt.Sprintf("%07d", x)
	x = rand.Intn(i-1) + 1
	eNode := fmt.Sprintf("%07d", x)

	edge := ADEdge{
		Kind:  "Foo",
		Start: sNode,
		End:   eNode,
	}
	return edge

}

func (e ADEdge) ToGraphEdge() GenericIngestEdge {
	return GenericIngestEdge{
		Kind: e.Kind,
		Start: GenericIngestEndpoint{
			Value:   e.Start,
			MatchBy: "id",
		},
		End: GenericIngestEndpoint{
			Value:   e.End,
			MatchBy: "id",
		},
		Properties: make(map[string]any),
	}

}
