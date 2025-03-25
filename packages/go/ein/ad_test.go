// Copyright 2024 Specter Ops, Inc.
//
// Licensed under the Apache License, Version 2.0
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//
// SPDX-License-Identifier: Apache-2.0

package ein_test

import (
	"testing"

	"github.com/specterops/bloodhound/ein"
	"github.com/specterops/bloodhound/graphschema/ad"
	"github.com/stretchr/testify/assert"
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

func TestConvertComputerToNode(t *testing.T) {
	computer := ein.Computer{
		IngestBase: ein.IngestBase{
			Properties: map[string]any{
				"isdc": true,
			},
		},
		NTLMRegistryData: ein.NTLMRegistryDataAPIResult{
			APIResult: ein.APIResult{
				Collected: true,
			},
			Result: ein.NTLMRegistryInfo{
				RestrictSendingNtlmTraffic: 1,
			},
		},
		IsWebClientRunning: ein.BoolAPIResult{
			APIResult: ein.APIResult{
				Collected: true,
			},
			Result: true,
		},
		SmbInfo: ein.SMBSigningAPIResult{
			APIResult: ein.APIResult{
				Collected: true,
			},
			SigningEnabled: true,
		},
	}

	result := ein.ConvertComputerToNode(computer)
	assert.Equal(t, true, result.PropertyMap[ad.IsDC.String()])
	assert.Equal(t, true, result.PropertyMap[ad.WebClientRunning.String()])
	assert.Equal(t, true, result.PropertyMap[ad.RestrictOutboundNTLM.String()])
	assert.Equal(t, true, result.PropertyMap[ad.SMBSigning.String()])
}
