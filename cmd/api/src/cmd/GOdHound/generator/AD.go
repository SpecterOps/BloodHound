package generator

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"os"
	"time"

	"github.com/specterops/bloodhound/src/cmd/GOdHound/config"
)

type ADDomain struct {
	Properties        map[string]any `json:"properties"`
	domain            string         // "PHANTOM.CORP",
	name              string         // "PHANTOM.CORP",
	distinguishedname string         // "DC=PHANTOM,DC=CORP",
	domainsid         string         // "S-1-5-21-2697957641-2271029196-387917394",
	isaclprotected    bool           // false,
	description       string         // null,
	whencreated       int64          // 1653361613,
	functionallevel   string         // "2016",
	collected         bool           // true
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
	operatingsystem         string         // null,
	Sidhistory              []string       // []
	objectid                string
	//id                      string
}

type ADGroup struct {
	Properties        map[string]any `json:"properties"`
	domain            string         // "PHANTOM.CORP",
	name              string         // "ALICE-LAPTOP.PHANTOM.CORP",
	distinguishedname string         // "CN=ALICE-LAPTOP,OU=WORKSTATIONS,DC=PHANTOM,DC=CORP",
	samaccountname    string
	domainsid         string // "S-1-5-21-2697957641-2271029196-387917394",
	isaclprotected    bool
	description       string
	whencreated       int64
	admincount        bool
	objectid          string
	//id                string
}

type ADGPO struct {
	Properties        map[string]any `json:"properties"`
	domain            string         // "PHANTOM.CORP",
	name              string         // "SHARPHOUNDSERVICELOGON@PHANTOM.CORP",
	distinguishedname string         // "CN={31B0D13F-8857-4019-89A9-992FF70D869C},CN=POLICIES,CN=SYSTEM,DC=PHANTOM,DC=CORP",
	domainsid         string         // "S-1-5-21-2697957641-2271029196-387917394",
	isaclprotected    bool
	description       string //null,
	whencreated       int64  // 1664354077,
	gpcpath           string // "\\\\PHANTOM.CORP\\SYSVOL\\PHANTOM.CORP\\POLICIES\\{31B0D13F-8857-4019-89A9-992FF70D869C}"
	objectid          string
}

type ADOU struct {
	Properties        map[string]any `json:"properties"`
	domain            string         // "PHANTOM.CORP",
	name              string         // "USERS@PHANTOM.CORP",
	distinguishedname string         // "OU=USERS,OU=TIER1,DC=PHANTOM,DC=CORP",
	domainsid         string         // "S-1-5-21-2697957641-2271029196-387917394",
	isaclprotected    bool           // false,
	description       string         // null,
	whencreated       int64          // 1664356125,
	blocksinheritance bool           // false
	objectid          string
}

type ADEdge struct {
	Kind       string
	Start      string
	End        string
	Properties map[string]any `json:"properties"`
}

// https://learn.microsoft.com/en-us/windows-server/identity/ad-ds/manage/understand-security-identifiers
func GenerateSID(myrid string) string {
	var (
		// Keeps track of generated SIDs to ensure uniqueness
		generatedSIDs = make(map[string]struct{})
	)
	r := rand.New(rand.NewSource(time.Now().UnixNano()))

	// GenerateUniqueSID creates a unique, valid Windows SID that doesn't collide
	for {
		subAuth1 := r.Uint32()
		subAuth2 := r.Uint32()
		subAuth3 := r.Uint32()

		if myrid == "" {
			myrid = fmt.Sprintf("%d", uint32(1000+r.Intn(10000)))
		}

		sid := fmt.Sprintf("S-1-5-21-%d-%d-%d-%v", subAuth1, subAuth2, subAuth3, myrid)

		if _, exists := generatedSIDs[sid]; !exists {
			generatedSIDs[sid] = struct{}{}
			return sid
		}
		// else, retry
	}
}

func AddADDomain() ADDomain {
	domain := ADDomain{
		domain:            config.Domain.Name,
		name:              config.Domain.Name,
		distinguishedname: "", // "DC=PHANTOM,DC=CORP"
		domainsid:         config.Domain.SID,
		isaclprotected:    false,
		description:       "",
		whencreated:       1653366613,
		functionallevel:   "2022",
		collected:         true,
	}
	return domain
}

func (d ADDomain) ToGraphNode() GenericIngestNode {
	return GenericIngestNode{
		ID:    "0000000",
		Kinds: []string{"Domain", "Base"},
		Properties: map[string]interface{}{
			"domain":            d.domain,
			"name":              d.name,
			"distinguishedname": d.distinguishedname,
			"domainsid":         d.domainsid,
			"isaclprotected":    d.isaclprotected,
			"description":       d.description,
			"whencreated":       1653361613,
			"functionallevel":   d.functionallevel,
			"collected":         d.collected,
		},
	}
}

func AddADUser(r *rand.Rand, x int32, prefix string) ADUser {

	//oid := fmt.Sprintf("%07d", x)
	rid := ""
	sid := GenerateSID(rid)
	uName := ""
	displayName := ""

	if prefix == "" {
		first := getRandomName(r, "resources/firstname.json")
		last := getRandomName(r, "resources/lastname.json")
		uName = fmt.Sprintf(fmt.Sprintf("%.1s", first) + last)
		displayName = fmt.Sprintf(first + " " + last)
	} else {
		uName = fmt.Sprintf(prefix + fmt.Sprintf("%07d", x))
		displayName = uName
	}

	user := ADUser{
		domain:                config.Domain.Name,
		name:                  uName,
		displayname:           displayName,
		domainsid:             config.Domain.SID,
		objectid:              sid,
		enabled:               true,
		email:                 fmt.Sprintf(uName + "@" + config.Domain.Name),
		Serviceprincipalnames: []string{},
		Sidhistory:            []string{},
	}
	return user
}

func (u ADUser) ToGraphNode() GenericIngestNode {
	return GenericIngestNode{
		ID:    u.objectid, // Using objectid as unique identifier
		Kinds: []string{"User", "Base"},
		Properties: map[string]interface{}{
			"displayname":             u.displayname,
			"name":                    u.name,
			"objectid":                u.objectid,
			"domain":                  u.domain,
			"domainsid":               u.domainsid,
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

func AddADComputer(r *rand.Rand, x int32, prefix string) ADComputer {
	//oid := fmt.Sprintf("%07d", x)
	rid := ""
	sid := GenerateSID(rid)
	cName := fmt.Sprintf(prefix+"%07d"+"."+config.Domain.Name, x)

	computer := ADComputer{
		domain:                config.Domain.Name,
		name:                  cName,
		objectid:              sid,
		domainsid:             config.Domain.SID,
		enabled:               true,
		Serviceprincipalnames: []string{},
		Sidhistory:            []string{},
	}
	return computer
}

func (c ADComputer) ToGraphNode() GenericIngestNode {
	return GenericIngestNode{
		ID:    c.objectid,
		Kinds: []string{"Computer", "Base"},
		Properties: map[string]interface{}{
			"name":                    c.name,
			"displayname":             c.name,
			"objectid":                c.objectid,
			"domain":                  c.domain,
			"domainsid":               c.domainsid,
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

func AddADGoup(x int32, name string, rid string) ADGroup {
	//id := fmt.Sprintf("%07d", x)
	sid := GenerateSID(rid)
	gName := ""
	samAccountName := ""
	isAdmin := false
	if name == "" {
		gName = fmt.Sprintf("GROUP"+"%07d"+"@"+config.Domain.Name, x)
		samAccountName = fmt.Sprintf("GROUP"+"%07d", x)
	} else {
		gName = fmt.Sprintf(name + "@" + config.Domain.Name)
		samAccountName = name
		isAdmin = true
	}

	group := ADGroup{
		domain:            config.Domain.Name,
		name:              gName,
		distinguishedname: "", // "CN=PRINT OPERATORS,CN=BUILTIN,DC=PHANTOM,DC=CORP",
		samaccountname:    samAccountName,
		isaclprotected:    true,
		objectid:          sid,
		domainsid:         config.Domain.SID,
		admincount:        isAdmin,
		//id:                sid,
	}
	return group
}

func (g ADGroup) ToGraphNode() GenericIngestNode {
	return GenericIngestNode{
		ID:    g.objectid,
		Kinds: []string{"Group", "Base"},
		Properties: map[string]interface{}{
			"name":              g.name,
			"displayname":       g.name,
			"objectid":          g.objectid,
			"domain":            g.domain,
			"domainsid":         g.domainsid,
			"distinguishedname": g.distinguishedname,
			"samaccountname":    g.samaccountname,
			"isaclprotected":    g.isaclprotected,
			"description":       g.description,
			"whencreated":       g.whencreated,
			"admincount":        g.admincount,
		},
	}
}

func AddADGPO(r *rand.Rand, x int32) ADGPO {
	//oid := fmt.Sprintf("%07d", x)
	rid := ""
	sid := GenerateSID(rid)
	GPO := ADGPO{
		domain:            config.Domain.Name,
		name:              fmt.Sprintf("GP0"+"%07d"+"@"+config.Domain.Name, x), // "USERS@PHANTOM.CORP",
		distinguishedname: "",                                                  // "OU=USERS,OU=TIER1,DC=PHANTOM,DC=CORP",
		domainsid:         config.Domain.SID,
		isaclprotected:    false, // false,
		description:       "",
		whencreated:       0, // 1664356125,
		gpcpath:           "",
		objectid:          sid,
	}
	return GPO
}

func (g ADGPO) ToGraphNode() GenericIngestNode {
	return GenericIngestNode{
		ID:    g.objectid,
		Kinds: []string{"GPO", "Base"},
		Properties: map[string]interface{}{
			"domain":            g.objectid,
			"name":              g.name,
			"distinguishedname": g.distinguishedname,
			"domainsid":         g.domainsid,
			"isaclprotected":    g.isaclprotected,
			"description":       g.description,
			"whencreated":       g.whencreated,
			"gpcpath":           g.gpcpath,
		},
	}
}

func AddADOU(r *rand.Rand, x int32) ADOU {
	//oid := fmt.Sprintf("%07d", x)
	rid := ""
	sid := GenerateSID(rid)
	OU := ADOU{
		domain:            config.Domain.Name,
		name:              fmt.Sprintf("GP0"+"%07d"+"@"+config.Domain.Name, x), // "USERS@PHANTOM.CORP",
		distinguishedname: "",                                                  // "OU=USERS,OU=TIER1,DC=PHANTOM,DC=CORP",
		domainsid:         config.Domain.SID,
		isaclprotected:    false, // false,
		description:       "",
		whencreated:       0,     // 1664356125,
		blocksinheritance: false, // false
		objectid:          sid,
	}
	return OU
}

func (o ADOU) ToGraphNode() GenericIngestNode {
	return GenericIngestNode{
		ID:    o.objectid,
		Kinds: []string{"OU", "Base"},
		Properties: map[string]interface{}{
			"domain":            o.objectid,
			"name":              o.name,
			"distinguishedname": o.distinguishedname,
			"domainsid":         o.domainsid,
			"isaclprotected":    o.isaclprotected,
			"description":       o.description,
			"whencreated":       o.whencreated,
			"blocksinheritance": o.blocksinheritance,
		},
	}
}

/*
func AddADEdge(r *rand.Rand, i int) ADEdge {
	x := rand.Intn(i-1) + 1

	sNode := fmt.Sprintf("%07d", x)
	x = rand.Intn(i-1) + 1
	eNode := fmt.Sprintf("%07d", x)

	// Open the Edge file
	file, err := os.Open("./resources/AD/edges.json")
	if err != nil {
		fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	var edges []string
	if err := json.NewDecoder(file).Decode(&edges); err != nil {
		fmt.Errorf("failed to decode JSON: %w", err)
	}

	// Check if file is empty
	if len(edges) == 0 {
		fmt.Errorf("no edges found in file")
	}

	edgeType := edges[r.Intn(len(edges))]

	edge := ADEdge{
		Kind:  edgeType,
		Start: sNode,
		End:   eNode,
	}
	return edge

}
*/

func AddADEdge(r *rand.Rand, eType string, sNode string, eNode string) ADEdge {

	if eType == "" {
		// Open the Edge file
		file, err := os.Open("./resources/AD/edges.json")
		if err != nil {
			fmt.Errorf("failed to open file: %w", err)
		}
		defer file.Close()

		var edges []string
		if err := json.NewDecoder(file).Decode(&edges); err != nil {
			fmt.Errorf("failed to decode JSON: %w", err)
		}

		// Check if file is empty
		if len(edges) == 0 {
			fmt.Errorf("no edges found in file")
		}

		eType = edges[r.Intn(len(edges))]
	}

	edge := ADEdge{
		Kind:  eType,
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

/*
func AddMandatoryEdges(eType string, sNode string, eNode string) ADEdge {
	edge := ADEdge{
		Kind:  eType,
		Start: sNode,
		End:   eNode,
	}
	return edge
}
*/
