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
	"time"

	"github.com/specterops/bloodhound/packages/go/ein"
	"github.com/specterops/bloodhound/packages/go/graphschema/ad"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConvertDomainToNode_DomainInvalidProperties(t *testing.T) {
	baseItem := ein.Domain{
		IngestBase: ein.IngestBase{
			ObjectIdentifier: "ABC123",
			Properties: map[string]any{
				ad.MachineAccountQuota.String():                    "1",
				ad.MinPwdLength.String():                           "1",
				ad.PwdProperties.String():                          "1",
				ad.PwdHistoryLength.String():                       "1",
				ad.LockoutThreshold.String():                       "1",
				ad.ExpirePasswordsOnSmartCardOnlyAccounts.String(): "false",
			},
			Aces:           nil,
			IsDeleted:      false,
			IsACLProtected: false,
			ContainedBy:    ein.TypedPrincipal{},
		},
	}

	result := ein.ConvertDomainToNode(baseItem, time.Now().UTC())
	props := result.PropertyMap
	assert.Contains(t, props, ad.MachineAccountQuota.String())
	assert.Contains(t, props, ad.MinPwdLength.String())
	assert.Contains(t, props, ad.PwdProperties.String())
	assert.Contains(t, props, ad.PwdHistoryLength.String())
	assert.Contains(t, props, ad.LockoutThreshold.String())
	assert.Contains(t, props, ad.ExpirePasswordsOnSmartCardOnlyAccounts.String())
	assert.Equal(t, 1, props[ad.MachineAccountQuota.String()])
	assert.Equal(t, 1, props[ad.MinPwdLength.String()])
	assert.Equal(t, 1, props[ad.PwdProperties.String()])
	assert.Equal(t, 1, props[ad.PwdHistoryLength.String()])
	assert.Equal(t, 1, props[ad.LockoutThreshold.String()])
	assert.Equal(t, false, props[ad.ExpirePasswordsOnSmartCardOnlyAccounts.String()])

	baseItem = ein.Domain{
		IngestBase: ein.IngestBase{
			ObjectIdentifier: "ABC123",
			Properties: map[string]any{
				ad.MachineAccountQuota.String():                    1,
				ad.MinPwdLength.String():                           1,
				ad.PwdProperties.String():                          1,
				ad.PwdHistoryLength.String():                       1,
				ad.LockoutThreshold.String():                       1,
				ad.ExpirePasswordsOnSmartCardOnlyAccounts.String(): false,
			},
			Aces:           nil,
			IsDeleted:      false,
			IsACLProtected: false,
			ContainedBy:    ein.TypedPrincipal{},
		},
	}

	result = ein.ConvertDomainToNode(baseItem, time.Now().UTC())
	props = result.PropertyMap
	assert.Contains(t, props, ad.MachineAccountQuota.String())
	assert.Contains(t, props, ad.MinPwdLength.String())
	assert.Contains(t, props, ad.PwdProperties.String())
	assert.Contains(t, props, ad.PwdHistoryLength.String())
	assert.Contains(t, props, ad.LockoutThreshold.String())
	assert.Contains(t, props, ad.ExpirePasswordsOnSmartCardOnlyAccounts.String())
	assert.Equal(t, 1, props[ad.MachineAccountQuota.String()])
	assert.Equal(t, 1, props[ad.MinPwdLength.String()])
	assert.Equal(t, 1, props[ad.PwdProperties.String()])
	assert.Equal(t, 1, props[ad.PwdHistoryLength.String()])
	assert.Equal(t, 1, props[ad.LockoutThreshold.String()])
	assert.Equal(t, false, props[ad.ExpirePasswordsOnSmartCardOnlyAccounts.String()])
}

func TestConvertDomainToNode_InheritanceHashes(t *testing.T) {
	testHash := "abc123"
	domainObject := ein.Domain{
		IngestBase:        ein.IngestBase{},
		InheritanceHashes: []string{testHash},
	}

	result := ein.ConvertDomainToNode(domainObject, time.Now().UTC())
	assert.Contains(t, result.PropertyMap[ad.InheritanceHashes.String()], testHash)
}

func TestConvertOUToNode_InheritanceHashes(t *testing.T) {
	testHash := "abc123"
	ouObject := ein.OU{
		IngestBase:        ein.IngestBase{},
		InheritanceHashes: []string{testHash},
	}

	result := ein.ConvertOUToNode(ouObject, time.Now().UTC())
	assert.Contains(t, result.PropertyMap[ad.InheritanceHashes.String()], testHash)
}

func TestParseDomainTrusts_TrustAttributes(t *testing.T) {
	domainObject := ein.Domain{
		IngestBase:   ein.IngestBase{},
		ChildObjects: nil,
		Trusts:       make([]ein.Trust, 0),
		Links:        nil,
	}

	t.Run("TrustAttributes Parse String Success", func(t *testing.T) {
		domainObject.Trusts = []ein.Trust{
			{
				TargetDomainSid:      "abc123",
				IsTransitive:         false,
				TrustDirection:       ein.TrustDirectionOutbound,
				TrustType:            "abc",
				SidFilteringEnabled:  true,
				TargetDomainName:     "abc456",
				TGTDelegationEnabled: false,
				TrustAttributes:      "12345",
			},
		}

		result := ein.ParseDomainTrusts(domainObject)
		require.Len(t, result.TrustRelationships, 1)

		rel := result.TrustRelationships[0]
		assert.Contains(t, rel.RelProps, ad.TrustAttributesOutbound.String())
		assert.Equal(t, rel.RelProps[ad.TrustAttributesOutbound.String()], 12345)
	})

	t.Run("TrustAttributes Parse Int Success", func(t *testing.T) {
		domainObject.Trusts = []ein.Trust{
			{
				TargetDomainSid:      "abc123",
				IsTransitive:         false,
				TrustDirection:       ein.TrustDirectionOutbound,
				TrustType:            "abc",
				SidFilteringEnabled:  true,
				TargetDomainName:     "abc456",
				TGTDelegationEnabled: false,
				TrustAttributes:      12345,
			},
		}

		result := ein.ParseDomainTrusts(domainObject)
		require.Len(t, result.TrustRelationships, 1)

		rel := result.TrustRelationships[0]
		assert.Contains(t, rel.RelProps, ad.TrustAttributesOutbound.String())
		assert.Equal(t, rel.RelProps[ad.TrustAttributesOutbound.String()], 12345)
	})

	t.Run("TrustAttributes Parse Float64 Success", func(t *testing.T) {
		domainObject.Trusts = []ein.Trust{
			{
				TargetDomainSid:      "abc123",
				IsTransitive:         false,
				TrustDirection:       ein.TrustDirectionInbound,
				TrustType:            "abc",
				SidFilteringEnabled:  true,
				TargetDomainName:     "abc456",
				TGTDelegationEnabled: false,
				TrustAttributes:      float64(12345),
			},
		}

		result := ein.ParseDomainTrusts(domainObject)
		require.Len(t, result.TrustRelationships, 1)

		rel := result.TrustRelationships[0]
		assert.Contains(t, rel.RelProps, ad.TrustAttributesInbound.String())
		assert.Equal(t, 12345, rel.RelProps[ad.TrustAttributesInbound.String()])
	})

	t.Run("TrustAttributes Parse Float32 Success", func(t *testing.T) {
		domainObject.Trusts = []ein.Trust{
			{
				TargetDomainSid:      "abc123",
				IsTransitive:         false,
				TrustDirection:       ein.TrustDirectionInbound,
				TrustType:            "abc",
				SidFilteringEnabled:  true,
				TargetDomainName:     "abc456",
				TGTDelegationEnabled: false,
				TrustAttributes:      float32(12345),
			},
		}

		result := ein.ParseDomainTrusts(domainObject)
		require.Len(t, result.TrustRelationships, 1)

		rel := result.TrustRelationships[0]
		assert.Contains(t, rel.RelProps, ad.TrustAttributesInbound.String())
		assert.Equal(t, 12345, rel.RelProps[ad.TrustAttributesInbound.String()])
	})

	t.Run("TrustAttributes Parse Unknown Type Failure", func(t *testing.T) {
		domainObject.Trusts = []ein.Trust{
			{
				TargetDomainSid:      "abc123",
				IsTransitive:         false,
				TrustDirection:       ein.TrustDirectionInbound,
				TrustType:            "abc",
				SidFilteringEnabled:  true,
				TargetDomainName:     "abc456",
				TGTDelegationEnabled: false,
				TrustAttributes:      uint32(12345),
			},
		}

		result := ein.ParseDomainTrusts(domainObject)
		require.Len(t, result.TrustRelationships, 1)

		rel := result.TrustRelationships[0]
		assert.Contains(t, rel.RelProps, ad.TrustAttributesInbound.String())
		assert.Equal(t, nil, rel.RelProps[ad.TrustAttributesInbound.String()])
	})
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
				RestrictSendingNtlmTraffic: 2,
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
			Result: ein.SMBSigningResult{
				SigningEnabled: true,
			},
		},
	}

	result := ein.ConvertComputerToNode(computer, time.Now())
	assert.Equal(t, true, result.PropertyMap[ad.IsDC.String()])
	assert.Equal(t, true, result.PropertyMap[ad.WebClientRunning.String()])
	assert.Equal(t, true, result.PropertyMap[ad.RestrictOutboundNTLM.String()])
	assert.Equal(t, true, result.PropertyMap[ad.SMBSigning.String()])
}
