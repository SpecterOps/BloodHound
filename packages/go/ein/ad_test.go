package ein_test

import (
	"github.com/specterops/bloodhound/ein"
	"github.com/specterops/bloodhound/graphschema/ad"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestConvertObjectToNode_DomainInvalidProperties(t *testing.T) {
	baseItem := ein.IngestBase{
		ObjectIdentifier: "ABC123",
		Properties: map[string]any{
			"machineaccountquota":                    "1",
			"minpwdlength":                           "1",
			"pwdproperties":                          "1",
			"pwdhistorylength":                       "1",
			"lockoutthreshold":                       "1",
			"expirepasswordsonsmartcardonlyaccounts": "false",
		},
		Aces:           nil,
		IsDeleted:      false,
		IsACLProtected: false,
		ContainedBy:    ein.TypedPrincipal{},
	}

	result := ein.ConvertObjectToNode(baseItem, ad.Domain)
	props := result.PropertyMap
	assert.Contains(t, props, "machineaccountquota")
	assert.Contains(t, props, "minpwdlength")
	assert.Contains(t, props, "pwdproperties")
	assert.Contains(t, props, "pwdhistorylength")
	assert.Contains(t, props, "lockoutthreshold")
	assert.Contains(t, props, "expirepasswordsonsmartcardonlyaccounts")
	assert.Equal(t, 1, props["machineaccountquota"])
	assert.Equal(t, 1, props["minpwdlength"])
	assert.Equal(t, 1, props["pwdproperties"])
	assert.Equal(t, 1, props["pwdhistorylength"])
	assert.Equal(t, 1, props["lockoutthreshold"])
	assert.Equal(t, false, props["expirepasswordsonsmartcardonlyaccounts"])

	baseItem = ein.IngestBase{
		ObjectIdentifier: "ABC123",
		Properties: map[string]any{
			"machineaccountquota":                    1,
			"minpwdlength":                           1,
			"pwdproperties":                          1,
			"pwdhistorylength":                       1,
			"lockoutthreshold":                       1,
			"expirepasswordsonsmartcardonlyaccounts": false,
		},
		Aces:           nil,
		IsDeleted:      false,
		IsACLProtected: false,
		ContainedBy:    ein.TypedPrincipal{},
	}

	result = ein.ConvertObjectToNode(baseItem, ad.Domain)
	props = result.PropertyMap
	assert.Contains(t, props, "machineaccountquota")
	assert.Contains(t, props, "minpwdlength")
	assert.Contains(t, props, "pwdproperties")
	assert.Contains(t, props, "pwdhistorylength")
	assert.Contains(t, props, "lockoutthreshold")
	assert.Contains(t, props, "expirepasswordsonsmartcardonlyaccounts")
	assert.Equal(t, 1, props["machineaccountquota"])
	assert.Equal(t, 1, props["minpwdlength"])
	assert.Equal(t, 1, props["pwdproperties"])
	assert.Equal(t, 1, props["pwdhistorylength"])
	assert.Equal(t, 1, props["lockoutthreshold"])
	assert.Equal(t, false, props["expirepasswordsonsmartcardonlyaccounts"])
}

func TestParseDomainTrusts_TrustAttributesFix(t *testing.T) {
	domainObject := ein.Domain{
		IngestBase:   ein.IngestBase{},
		ChildObjects: nil,
		Trusts:       make([]ein.Trust, 0),
		Links:        nil,
	}

	domainObject.Trusts = append(domainObject.Trusts, ein.Trust{
		TargetDomainSid:      "abc123",
		IsTransitive:         false,
		TrustDirection:       ein.TrustDirectionInbound,
		TrustType:            "abc",
		SidFilteringEnabled:  false,
		TargetDomainName:     "abc456",
		TGTDelegationEnabled: false,
		TrustAttributes:      "12345",
	})

	result := ein.ParseDomainTrusts(domainObject)
	assert.Len(t, result.TrustRelationships, 1)

	rel := result.TrustRelationships[0]
	assert.Contains(t, rel.RelProps, "trustattributes")
	assert.Equal(t, rel.RelProps["trustattributes"], 12345)

	domainObject.Trusts = make([]ein.Trust, 0)
	domainObject.Trusts = append(domainObject.Trusts, ein.Trust{
		TargetDomainSid:      "abc123",
		IsTransitive:         false,
		TrustDirection:       ein.TrustDirectionInbound,
		TrustType:            "abc",
		SidFilteringEnabled:  false,
		TargetDomainName:     "abc456",
		TGTDelegationEnabled: false,
		TrustAttributes:      12345,
	})

	result = ein.ParseDomainTrusts(domainObject)
	assert.Len(t, result.TrustRelationships, 1)

	rel = result.TrustRelationships[0]
	assert.Contains(t, rel.RelProps, "trustattributes")
	assert.Equal(t, rel.RelProps["trustattributes"], 12345)
}
