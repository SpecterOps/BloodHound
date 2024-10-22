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
