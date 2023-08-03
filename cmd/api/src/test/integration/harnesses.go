// Copyright 2023 Specter Ops, Inc.
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

package integration

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/gofrs/uuid"
	adAnalysis "github.com/specterops/bloodhound/analysis/ad"
	"github.com/specterops/bloodhound/dawgs/graph"
	"github.com/specterops/bloodhound/graphschema/ad"
	"github.com/specterops/bloodhound/graphschema/azure"
	"github.com/specterops/bloodhound/graphschema/common"
	"github.com/specterops/bloodhound/src/test"
)

func RandomObjectID(t test.Controller) string {
	newUUID, err := uuid.NewV4()

	if err != nil {
		t.Fatalf("Failed to generate a new UUID: %v", err)
	}

	return newUUID.String()
}

func RandomDomainSID() string {
	var (
		authority     = rand.Int31()
		subAuthority1 = rand.Int31()
		subAuthority2 = rand.Int31()
	)

	return fmt.Sprintf("S-1-5-21-%d-%d-%d", authority, subAuthority1, subAuthority2)
}

const (
	HarnessUserName             = "user"
	HarnessUserDescription      = "A user"
	HarnessUserLicenses         = "licenses"
	HarnessUserMFAEnabled       = false
	HarnessAppName              = "application"
	HarnessServicePrincipalName = "service_principal"
)

type GraphTestHarness interface {
	Setup(testContext *GraphTestContext)
}

type CompletenessHarness struct {
	UserA        *graph.Node
	UserB        *graph.Node
	UserC        *graph.Node
	UserD        *graph.Node
	UserInactive *graph.Node
	ComputerA    *graph.Node
	ComputerB    *graph.Node
	ComputerC    *graph.Node
	ComputerD    *graph.Node
	Group        *graph.Node
	DomainSid    string
}

func (s *CompletenessHarness) Setup(testCtx *GraphTestContext) {
	s.DomainSid = RandomDomainSID()
	s.UserA = testCtx.NewActiveDirectoryUser("CUserA", s.DomainSid)
	s.UserB = testCtx.NewActiveDirectoryUser("CUserB", s.DomainSid)
	s.UserC = testCtx.NewActiveDirectoryUser("CUserC", s.DomainSid)
	s.UserD = testCtx.NewActiveDirectoryUser("CUserD", s.DomainSid)
	s.Group = testCtx.NewActiveDirectoryGroup("CGroup", s.DomainSid)
	s.UserInactive = testCtx.NewActiveDirectoryUser("CUserInactive", s.DomainSid)
	s.ComputerA = testCtx.NewActiveDirectoryComputer("CComputerA", s.DomainSid)
	s.ComputerB = testCtx.NewActiveDirectoryComputer("CComputerB", s.DomainSid)
	s.ComputerC = testCtx.NewActiveDirectoryComputer("CComputerC", s.DomainSid)
	s.ComputerD = testCtx.NewActiveDirectoryComputer("CComputerD", s.DomainSid)

	testCtx.NewRelationship(s.ComputerA, s.UserA, ad.HasSession)
	testCtx.NewRelationship(s.ComputerA, s.UserB, ad.HasSession)
	testCtx.NewRelationship(s.ComputerB, s.UserB, ad.HasSession)
	testCtx.NewRelationship(s.UserA, s.Group, ad.MemberOf)
	testCtx.NewRelationship(s.UserB, s.Group, ad.MemberOf)
	testCtx.NewRelationship(s.UserC, s.Group, ad.MemberOf)
	testCtx.NewRelationship(s.UserD, s.Group, ad.MemberOf)
	testCtx.NewRelationship(s.UserInactive, s.Group, ad.MemberOf)
	testCtx.NewRelationship(s.ComputerA, s.Group, ad.MemberOf)
	testCtx.NewRelationship(s.ComputerB, s.Group, ad.MemberOf)
	testCtx.NewRelationship(s.ComputerC, s.Group, ad.MemberOf)
	testCtx.NewRelationship(s.ComputerD, s.Group, ad.MemberOf)
	testCtx.NewRelationship(s.Group, s.ComputerC, ad.AdminTo)
	testCtx.NewRelationship(s.UserD, s.ComputerC, ad.AdminTo)
	s.UserA.Properties.Set(ad.LastLogonTimestamp.String(), time.Now().UTC())
	testCtx.UpdateNode(s.UserA)
	s.UserB.Properties.Set(ad.LastLogonTimestamp.String(), time.Now().UTC())
	testCtx.UpdateNode(s.UserB)
	s.UserC.Properties.Set(ad.LastLogonTimestamp.String(), time.Now().UTC())
	testCtx.UpdateNode(s.UserC)
	s.UserD.Properties.Set(ad.LastLogonTimestamp.String(), time.Now().UTC())
	testCtx.UpdateNode(s.UserD)
	s.UserInactive.Properties.Set(ad.LastLogonTimestamp.String(), time.Now().UTC().Add(-time.Hour*3000))
	testCtx.UpdateNode(s.UserInactive)
	s.ComputerC.Properties.Set(common.PasswordLastSet.String(), time.Now().UTC())
	s.ComputerC.Properties.Set(common.OperatingSystem.String(), "WINDOWS")
	testCtx.UpdateNode(s.ComputerC)
	s.ComputerD.Properties.Set(common.PasswordLastSet.String(), time.Now().UTC())
	s.ComputerD.Properties.Set(common.OperatingSystem.String(), "WINDOWS")
	testCtx.UpdateNode(s.ComputerD)
}

type TrustDCSyncHarness struct {
	DomainA *graph.Node
	DomainB *graph.Node
	DomainC *graph.Node
	DomainD *graph.Node
	GPOA    *graph.Node
	GPOB    *graph.Node
	OU      *graph.Node
	GroupA  *graph.Node
	GroupB  *graph.Node
	UserA   *graph.Node
	UserB   *graph.Node
	UserC   *graph.Node
}

func (s *TrustDCSyncHarness) Setup(testCtx *GraphTestContext) {
	s.DomainA = testCtx.NewActiveDirectoryDomain("DomainA", RandomDomainSID(), false, true)
	s.DomainB = testCtx.NewActiveDirectoryDomain("DomainB", RandomDomainSID(), false, true)
	s.DomainC = testCtx.NewActiveDirectoryDomain("DomainC", RandomDomainSID(), false, false)
	s.DomainD = testCtx.NewActiveDirectoryDomain("DomainD", RandomDomainSID(), false, false)
	s.GPOA = testCtx.NewActiveDirectoryGPO("GPOA", testCtx.Harness.RootADHarness.ActiveDirectoryDomainSID)
	s.GPOB = testCtx.NewActiveDirectoryGPO("GPOB", testCtx.Harness.RootADHarness.ActiveDirectoryDomainSID)
	s.OU = testCtx.NewActiveDirectoryOU("OU", testCtx.Harness.RootADHarness.ActiveDirectoryDomainSID, false)
	s.GroupA = testCtx.NewActiveDirectoryGroup("GroupA", testCtx.Harness.RootADHarness.ActiveDirectoryDomainSID)
	s.GroupB = testCtx.NewActiveDirectoryGroup("GroupB", testCtx.Harness.RootADHarness.ActiveDirectoryDomainSID)
	s.UserA = testCtx.NewActiveDirectoryUser("UserA", testCtx.Harness.RootADHarness.ActiveDirectoryDomainSID)
	s.UserB = testCtx.NewActiveDirectoryUser("UserB", testCtx.Harness.RootADHarness.ActiveDirectoryDomainSID)
	s.UserC = testCtx.NewActiveDirectoryUser("UserC", testCtx.Harness.RootADHarness.ActiveDirectoryDomainSID)

	testCtx.NewRelationship(s.DomainA, s.DomainB, ad.TrustedBy)
	testCtx.NewRelationship(s.DomainB, s.DomainA, ad.TrustedBy)
	testCtx.NewRelationship(s.DomainA, s.DomainC, ad.TrustedBy)

	testCtx.NewRelationship(s.DomainB, s.DomainD, ad.TrustedBy)
	testCtx.NewRelationship(s.DomainD, s.DomainB, ad.TrustedBy)

	testCtx.NewRelationship(s.GPOA, s.DomainA, ad.GPLink, graph.AsProperties(graph.PropertyMap{
		ad.Enforced: false,
	}))

	testCtx.NewRelationship(s.GPOB, s.OU, ad.GPLink, graph.AsProperties(graph.PropertyMap{
		ad.Enforced: false,
	}))

	testCtx.NewRelationship(s.DomainA, s.OU, ad.Contains)

	testCtx.NewRelationship(s.GroupA, s.DomainA, ad.GetChanges)
	testCtx.NewRelationship(s.UserA, s.DomainA, ad.GetChanges)
	testCtx.NewRelationship(s.UserA, s.DomainA, ad.GetChangesAll)
	testCtx.NewRelationship(s.GroupB, s.DomainA, ad.GetChangesAll)
	testCtx.NewRelationship(s.UserB, s.DomainA, ad.DCSync)
	testCtx.NewRelationship(s.UserA, s.DomainA, ad.DCSync)

	testCtx.NewRelationship(s.UserB, s.GroupA, ad.MemberOf)
	testCtx.NewRelationship(s.UserB, s.GroupB, ad.MemberOf)
	testCtx.NewRelationship(s.UserC, s.GroupB, ad.MemberOf)
}

type ForeignDomainHarness struct {
	LocalGPO         *graph.Node
	LocalDomain      *graph.Node
	LocalOUA         *graph.Node
	LocalOUB         *graph.Node
	LocalGroup       *graph.Node
	LocalComputer    *graph.Node
	ForeignUserA     *graph.Node
	ForeignUserB     *graph.Node
	ForeignGroup     *graph.Node
	LocalDomainSID   string
	ForeignDomainSID string
}

func (s *ForeignDomainHarness) Setup(testCtx *GraphTestContext) {
	s.LocalDomainSID = RandomDomainSID()
	s.ForeignDomainSID = RandomDomainSID()

	s.LocalGPO = testCtx.NewActiveDirectoryGPO("LocalGPO", s.LocalDomainSID)
	s.LocalDomain = testCtx.NewActiveDirectoryDomain("LocalDomain", s.LocalDomainSID, false, true)
	s.LocalOUA = testCtx.NewActiveDirectoryOU("LocalOU A", s.LocalDomainSID, false)
	s.LocalOUB = testCtx.NewActiveDirectoryOU("LocalOU B", s.LocalDomainSID, false)
	s.LocalGroup = testCtx.NewActiveDirectoryGroup("LocalGroup", s.LocalDomainSID)
	s.LocalComputer = testCtx.NewActiveDirectoryComputer("LocalComputer", s.LocalDomainSID)

	s.ForeignUserA = testCtx.NewActiveDirectoryUser("ForeignUser A", s.ForeignDomainSID)
	s.ForeignUserB = testCtx.NewActiveDirectoryUser("ForeignUser B", s.ForeignDomainSID)
	s.ForeignGroup = testCtx.NewActiveDirectoryGroup("ForeignGroup", s.ForeignDomainSID)

	testCtx.NewRelationship(s.LocalGPO, s.LocalDomain, ad.GPLink)
	testCtx.NewRelationship(s.LocalDomain, s.LocalOUA, ad.Contains)
	testCtx.NewRelationship(s.LocalOUA, s.LocalGroup, ad.Contains)
	testCtx.NewRelationship(s.LocalGroup, s.LocalComputer, ad.AdminTo)
	testCtx.NewRelationship(s.ForeignUserA, s.LocalGPO, ad.GenericAll)
	testCtx.NewRelationship(s.ForeignGroup, s.LocalGroup, ad.MemberOf)
	testCtx.NewRelationship(s.ForeignUserA, s.ForeignGroup, ad.MemberOf)
	testCtx.NewRelationship(s.ForeignUserB, s.LocalGroup, ad.MemberOf)
	testCtx.NewRelationship(s.ForeignUserB, s.LocalComputer, ad.AdminTo)
	testCtx.NewRelationship(s.ForeignGroup, s.LocalGPO, ad.GenericAll)
	testCtx.NewRelationship(s.LocalGPO, s.LocalOUB, ad.GPLink)
}

type MembershipHarness struct {
	UserA     *graph.Node
	ComputerA *graph.Node
	ComputerB *graph.Node
	GroupA    *graph.Node
	GroupB    *graph.Node
	GroupC    *graph.Node
}

func (s *MembershipHarness) Setup(testCtx *GraphTestContext) {
	s.UserA = testCtx.NewActiveDirectoryUser("User", testCtx.Harness.RootADHarness.ActiveDirectoryDomainSID)
	s.ComputerA = testCtx.NewActiveDirectoryComputer("Computer 1", testCtx.Harness.RootADHarness.ActiveDirectoryDomainSID)
	s.ComputerB = testCtx.NewActiveDirectoryComputer("Computer 2", testCtx.Harness.RootADHarness.ActiveDirectoryDomainSID)
	s.GroupA = testCtx.NewActiveDirectoryGroup("Group A", testCtx.Harness.RootADHarness.ActiveDirectoryDomainSID)
	s.GroupB = testCtx.NewActiveDirectoryGroup("Group B", testCtx.Harness.RootADHarness.ActiveDirectoryDomainSID)
	s.GroupC = testCtx.NewActiveDirectoryGroup("Group C", testCtx.Harness.RootADHarness.ActiveDirectoryDomainSID)

	testCtx.NewRelationship(s.ComputerB, s.GroupC, ad.MemberOf)
	testCtx.NewRelationship(s.UserA, s.GroupB, ad.MemberOf)
	testCtx.NewRelationship(s.ComputerA, s.GroupB, ad.MemberOf)
	testCtx.NewRelationship(s.GroupB, s.GroupA, ad.MemberOf)
	testCtx.NewRelationship(s.GroupA, s.GroupC, ad.MemberOf)
}

type OUContainedHarness struct {
	Domain *graph.Node
	OUA    *graph.Node
	OUB    *graph.Node
	OUC    *graph.Node
	OUD    *graph.Node
	OUE    *graph.Node
	UserA  *graph.Node
	UserB  *graph.Node
	UserC  *graph.Node
}

func (s *OUContainedHarness) Setup(testCtx *GraphTestContext) {
	s.Domain = testCtx.NewActiveDirectoryDomain("Domain", RandomObjectID(testCtx.testCtrl), false, true)
	s.OUA = testCtx.NewActiveDirectoryOU("OUA", testCtx.Harness.RootADHarness.ActiveDirectoryDomainSID, false)
	s.OUB = testCtx.NewActiveDirectoryOU("OUB", testCtx.Harness.RootADHarness.ActiveDirectoryDomainSID, false)
	s.OUC = testCtx.NewActiveDirectoryOU("OUC", testCtx.Harness.RootADHarness.ActiveDirectoryDomainSID, false)
	s.OUD = testCtx.NewActiveDirectoryOU("OUD", testCtx.Harness.RootADHarness.ActiveDirectoryDomainSID, false)
	s.OUE = testCtx.NewActiveDirectoryOU("OUE", testCtx.Harness.RootADHarness.ActiveDirectoryDomainSID, false)
	s.UserA = testCtx.NewActiveDirectoryUser("UserA", testCtx.Harness.RootADHarness.ActiveDirectoryDomainSID)
	s.UserB = testCtx.NewActiveDirectoryUser("UserB", testCtx.Harness.RootADHarness.ActiveDirectoryDomainSID)
	s.UserC = testCtx.NewActiveDirectoryUser("UserC", testCtx.Harness.RootADHarness.ActiveDirectoryDomainSID)

	testCtx.NewRelationship(s.Domain, s.OUA, ad.Contains)
	testCtx.NewRelationship(s.Domain, s.OUB, ad.Contains)
	testCtx.NewRelationship(s.OUA, s.UserA, ad.Contains)
	testCtx.NewRelationship(s.OUA, s.OUC, ad.Contains)
	testCtx.NewRelationship(s.OUC, s.UserB, ad.Contains)
	testCtx.NewRelationship(s.OUB, s.OUD, ad.Contains)
	testCtx.NewRelationship(s.OUD, s.OUE, ad.Contains)
	testCtx.NewRelationship(s.OUE, s.UserC, ad.Contains)
}

type LocalGroupHarness struct {
	ComputerA *graph.Node
	ComputerB *graph.Node
	ComputerC *graph.Node
	UserA     *graph.Node
	UserB     *graph.Node
	UserC     *graph.Node
	UserD     *graph.Node
	GroupA    *graph.Node
	GroupB    *graph.Node
}

func (s *LocalGroupHarness) Setup(testCtx *GraphTestContext) {
	s.ComputerA = testCtx.NewActiveDirectoryComputer("ComputerA", testCtx.Harness.RootADHarness.ActiveDirectoryDomainSID)
	s.ComputerB = testCtx.NewActiveDirectoryComputer("ComputerB", testCtx.Harness.RootADHarness.ActiveDirectoryDomainSID)
	s.ComputerC = testCtx.NewActiveDirectoryComputer("ComputerC", testCtx.Harness.RootADHarness.ActiveDirectoryDomainSID)
	s.UserA = testCtx.NewActiveDirectoryUser("UserA", testCtx.Harness.RootADHarness.ActiveDirectoryDomainSID)
	s.UserB = testCtx.NewActiveDirectoryUser("UserB", testCtx.Harness.RootADHarness.ActiveDirectoryDomainSID)
	s.UserC = testCtx.NewActiveDirectoryUser("UserC", testCtx.Harness.RootADHarness.ActiveDirectoryDomainSID)
	s.UserD = testCtx.NewActiveDirectoryUser("UserD", testCtx.Harness.RootADHarness.ActiveDirectoryDomainSID)
	s.GroupA = testCtx.NewActiveDirectoryGroup("GroupA", testCtx.Harness.RootADHarness.ActiveDirectoryDomainSID)
	s.GroupB = testCtx.NewActiveDirectoryGroup("GroupB", testCtx.Harness.RootADHarness.ActiveDirectoryDomainSID)

	testCtx.NewRelationship(s.UserB, s.GroupA, ad.MemberOf)
	testCtx.NewRelationship(s.ComputerA, s.GroupB, ad.MemberOf)
	testCtx.NewRelationship(s.GroupA, s.ComputerA, ad.AdminTo)
	testCtx.NewRelationship(s.UserA, s.ComputerA, ad.AdminTo)
	testCtx.NewRelationship(s.ComputerA, s.ComputerB, ad.AdminTo)
	testCtx.NewRelationship(s.GroupB, s.ComputerC, ad.AdminTo)
	testCtx.NewRelationship(s.UserC, s.ComputerA, ad.SQLAdmin)
	testCtx.NewRelationship(s.UserD, s.ComputerA, ad.AllowedToDelegate)
}

type AssetGroupComboNodeHarness struct {
	GroupA *graph.Node
	GroupB *graph.Node
}

func (s *AssetGroupComboNodeHarness) Setup(testCtx *GraphTestContext) {
	s.GroupA = testCtx.NewActiveDirectoryGroup("GroupA", RandomObjectID(testCtx.testCtrl))
	s.GroupB = testCtx.NewActiveDirectoryGroup("GroupB", RandomObjectID(testCtx.testCtrl))
	s.GroupB.Properties.Set(common.SystemTags.String(), ad.AdminTierZero)
	testCtx.UpdateNode(s.GroupB)

	testCtx.NewRelationship(s.GroupA, s.GroupB, ad.MemberOf)
}

type InboundControlHarness struct {
	ControlledUser  *graph.Node
	ControlledGroup *graph.Node
	GroupA          *graph.Node
	GroupB          *graph.Node
	GroupC          *graph.Node
	GroupD          *graph.Node
	UserA           *graph.Node
	UserB           *graph.Node
	UserC           *graph.Node
	UserD           *graph.Node
	UserE           *graph.Node
	UserF           *graph.Node
	UserG           *graph.Node
	UserH           *graph.Node
}

func (s *InboundControlHarness) Setup(testCtx *GraphTestContext) {
	s.ControlledUser = testCtx.NewActiveDirectoryUser("ControlledUser", testCtx.Harness.RootADHarness.ActiveDirectoryDomainSID)
	s.ControlledGroup = testCtx.NewActiveDirectoryUser("ControlledGroup", testCtx.Harness.RootADHarness.ActiveDirectoryDomainSID)
	s.GroupA = testCtx.NewActiveDirectoryGroup("GroupA", testCtx.Harness.RootADHarness.ActiveDirectoryDomainSID)
	s.GroupB = testCtx.NewActiveDirectoryGroup("GroupB", testCtx.Harness.RootADHarness.ActiveDirectoryDomainSID)
	s.GroupC = testCtx.NewActiveDirectoryGroup("GroupC", testCtx.Harness.RootADHarness.ActiveDirectoryDomainSID)
	s.GroupD = testCtx.NewActiveDirectoryGroup("GroupD", testCtx.Harness.RootADHarness.ActiveDirectoryDomainSID)
	s.UserA = testCtx.NewActiveDirectoryUser("UserA", testCtx.Harness.RootADHarness.ActiveDirectoryDomainSID)
	s.UserB = testCtx.NewActiveDirectoryUser("UserB", testCtx.Harness.RootADHarness.ActiveDirectoryDomainSID)
	s.UserC = testCtx.NewActiveDirectoryUser("UserC", testCtx.Harness.RootADHarness.ActiveDirectoryDomainSID)
	s.UserD = testCtx.NewActiveDirectoryUser("UserD", testCtx.Harness.RootADHarness.ActiveDirectoryDomainSID)
	s.UserE = testCtx.NewActiveDirectoryUser("UserE", testCtx.Harness.RootADHarness.ActiveDirectoryDomainSID)
	s.UserF = testCtx.NewActiveDirectoryUser("UserF", testCtx.Harness.RootADHarness.ActiveDirectoryDomainSID)
	s.UserG = testCtx.NewActiveDirectoryUser("UserG", testCtx.Harness.RootADHarness.ActiveDirectoryDomainSID)
	s.UserH = testCtx.NewActiveDirectoryUser("UserH", testCtx.Harness.RootADHarness.ActiveDirectoryDomainSID)

	testCtx.NewRelationship(s.GroupA, s.GroupB, ad.MemberOf)
	testCtx.NewRelationship(s.UserA, s.GroupB, ad.MemberOf)
	testCtx.NewRelationship(s.UserG, s.GroupC, ad.MemberOf)
	testCtx.NewRelationship(s.UserH, s.GroupD, ad.MemberOf)

	testCtx.NewRelationship(s.UserB, s.GroupB, ad.GenericAll)
	testCtx.NewRelationship(s.UserC, s.UserD, ad.GenericAll)
	testCtx.NewRelationship(s.UserD, s.ControlledUser, ad.GenericAll)
	testCtx.NewRelationship(s.GroupB, s.ControlledUser, ad.GenericAll)

	testCtx.NewRelationship(s.GroupC, s.ControlledGroup, ad.GenericWrite)
	testCtx.NewRelationship(s.GroupD, s.ControlledGroup, ad.GenericWrite)
	testCtx.NewRelationship(s.UserE, s.ControlledGroup, ad.GenericWrite)
	testCtx.NewRelationship(s.UserF, s.ControlledGroup, ad.GenericWrite)

	testCtx.NewRelationship(s.GroupC, s.ControlledGroup, ad.WriteDACL)
	testCtx.NewRelationship(s.GroupD, s.ControlledGroup, ad.WriteDACL)

	testCtx.NewRelationship(s.GroupC, s.ControlledGroup, ad.WriteOwner)
	testCtx.NewRelationship(s.GroupD, s.ControlledGroup, ad.WriteOwner)

	testCtx.NewRelationship(s.GroupD, s.ControlledGroup, ad.Owns)
}

type OutboundControlHarness struct {
	Controller  *graph.Node
	UserA       *graph.Node
	UserB       *graph.Node
	UserC       *graph.Node
	GroupA      *graph.Node
	GroupB      *graph.Node
	GroupC      *graph.Node
	ComputerA   *graph.Node
	ComputerB   *graph.Node
	ComputerC   *graph.Node
	ControllerB *graph.Node
	Computer1   *graph.Node
	Group1      *graph.Node
	Group2      *graph.Node
}

func (s *OutboundControlHarness) Setup(testCtx *GraphTestContext) {
	s.Controller = testCtx.NewActiveDirectoryUser("Controller", testCtx.Harness.RootADHarness.ActiveDirectoryDomainSID)
	s.UserA = testCtx.NewActiveDirectoryUser("UserA", testCtx.Harness.RootADHarness.ActiveDirectoryDomainSID)
	s.UserB = testCtx.NewActiveDirectoryUser("UserB", testCtx.Harness.RootADHarness.ActiveDirectoryDomainSID)
	s.UserC = testCtx.NewActiveDirectoryUser("UserC", testCtx.Harness.RootADHarness.ActiveDirectoryDomainSID)
	s.GroupA = testCtx.NewActiveDirectoryGroup("GroupA", testCtx.Harness.RootADHarness.ActiveDirectoryDomainSID)
	s.GroupB = testCtx.NewActiveDirectoryGroup("GroupB", testCtx.Harness.RootADHarness.ActiveDirectoryDomainSID)
	s.GroupC = testCtx.NewActiveDirectoryGroup("GroupC", testCtx.Harness.RootADHarness.ActiveDirectoryDomainSID)
	s.ComputerA = testCtx.NewActiveDirectoryComputer("ComputerA", testCtx.Harness.RootADHarness.ActiveDirectoryDomainSID)
	s.ComputerB = testCtx.NewActiveDirectoryComputer("ComputerB", testCtx.Harness.RootADHarness.ActiveDirectoryDomainSID)
	s.ComputerC = testCtx.NewActiveDirectoryComputer("ComputerC", testCtx.Harness.RootADHarness.ActiveDirectoryDomainSID)

	testCtx.NewRelationship(s.UserA, s.GroupA, ad.MemberOf)
	testCtx.NewRelationship(s.UserB, s.GroupB, ad.MemberOf)
	testCtx.NewRelationship(s.GroupA, s.GroupC, ad.MemberOf)
	testCtx.NewRelationship(s.Controller, s.GroupA, ad.MemberOf)

	testCtx.NewRelationship(s.Controller, s.UserC, ad.GenericAll)
	testCtx.NewRelationship(s.Controller, s.GroupB, ad.GenericAll)
	testCtx.NewRelationship(s.GroupA, s.ComputerA, ad.GenericAll)
	testCtx.NewRelationship(s.UserC, s.ComputerB, ad.GenericAll)
	testCtx.NewRelationship(s.GroupC, s.ComputerC, ad.GenericAll)
	testCtx.NewRelationship(s.UserC, s.GroupB, ad.MemberOf)

	s.ControllerB = testCtx.NewActiveDirectoryGroup("ControllerB", testCtx.Harness.RootADHarness.ActiveDirectoryDomainSID)
	s.Computer1 = testCtx.NewActiveDirectoryComputer("Computer1", testCtx.Harness.RootADHarness.ActiveDirectoryDomainSID)
	s.Group1 = testCtx.NewActiveDirectoryComputer("Group1", testCtx.Harness.RootADHarness.ActiveDirectoryDomainSID)
	s.Group2 = testCtx.NewActiveDirectoryComputer("Group2", testCtx.Harness.RootADHarness.ActiveDirectoryDomainSID)

	testCtx.NewRelationship(s.ControllerB, s.Computer1, ad.WriteAccountRestrictions)
	testCtx.NewRelationship(s.Computer1, s.Group1, ad.MemberOf)
	testCtx.NewRelationship(s.Computer1, s.Group2, ad.MemberOf)
}

type SessionHarness struct {
	User      *graph.Node
	ComputerA *graph.Node
	ComputerB *graph.Node
	GroupA    *graph.Node
	GroupB    *graph.Node
	GroupC    *graph.Node
}

func (s *SessionHarness) Setup(testCtx *GraphTestContext) {
	s.ComputerA = testCtx.NewActiveDirectoryComputer("ComputerA", testCtx.Harness.RootADHarness.ActiveDirectoryDomainSID)
	s.ComputerB = testCtx.NewActiveDirectoryComputer("ComputerB", testCtx.Harness.RootADHarness.ActiveDirectoryDomainSID)
	s.User = testCtx.NewActiveDirectoryUser("User", testCtx.Harness.RootADHarness.ActiveDirectoryDomainSID)
	s.GroupA = testCtx.NewActiveDirectoryGroup("GroupA", testCtx.Harness.RootADHarness.ActiveDirectoryDomainSID)
	s.GroupB = testCtx.NewActiveDirectoryGroup("GroupB", testCtx.Harness.RootADHarness.ActiveDirectoryDomainSID)
	s.GroupC = testCtx.NewActiveDirectoryGroup("GroupC", testCtx.Harness.RootADHarness.ActiveDirectoryDomainSID)

	testCtx.NewRelationship(s.ComputerA, s.User, ad.HasSession)
	testCtx.NewRelationship(s.ComputerB, s.User, ad.HasSession)
	testCtx.NewRelationship(s.User, s.GroupA, ad.MemberOf)
	testCtx.NewRelationship(s.User, s.GroupB, ad.MemberOf)
	testCtx.NewRelationship(s.GroupB, s.GroupC, ad.MemberOf)

}

type RDPHarness2 struct {
	Computer            *graph.Node
	RDPLocalGroup       *graph.Node
	UserA               *graph.Node
	UserB               *graph.Node
	UserC               *graph.Node
	RDPDomainUsersGroup *graph.Node
}

func (s *RDPHarness2) Setup(testCtx *GraphTestContext) {
	s.Computer = testCtx.NewActiveDirectoryComputer("WIN11", testCtx.Harness.RootADHarness.ActiveDirectoryDomainSID)
	s.RDPDomainUsersGroup = testCtx.NewActiveDirectoryGroup("RDP Domain Users 2", testCtx.Harness.RootADHarness.ActiveDirectoryDomainSID)
	s.RDPLocalGroup = testCtx.NewActiveDirectoryLocalGroup("Remote Desktop Users 2", testCtx.Harness.RootADHarness.ActiveDirectoryDomainSID)

	rdpLocalGroupObjectID, _ := s.RDPLocalGroup.Properties.Get(common.ObjectID.String()).String()

	s.RDPLocalGroup.Properties.Set(
		common.ObjectID.String(),
		rdpLocalGroupObjectID+adAnalysis.RDPGroupSuffix,
	)
	testCtx.UpdateNode(s.RDPLocalGroup)

	s.UserA = testCtx.NewActiveDirectoryUser("UserA", testCtx.Harness.RootADHarness.ActiveDirectoryDomainSID)
	s.UserB = testCtx.NewActiveDirectoryUser("UserB", testCtx.Harness.RootADHarness.ActiveDirectoryDomainSID)
	s.UserC = testCtx.NewActiveDirectoryUser("UserC", testCtx.Harness.RootADHarness.ActiveDirectoryDomainSID)

	testCtx.NewRelationship(s.RDPLocalGroup, s.Computer, ad.RemoteInteractiveLogonPrivilege)
	testCtx.NewRelationship(s.RDPLocalGroup, s.Computer, ad.LocalToComputer)
	testCtx.NewRelationship(s.RDPDomainUsersGroup, s.RDPLocalGroup, ad.MemberOfLocalGroup)
	testCtx.NewRelationship(s.UserA, s.RDPDomainUsersGroup, ad.MemberOf)
	testCtx.NewRelationship(s.UserB, s.RDPDomainUsersGroup, ad.MemberOf)
	testCtx.NewRelationship(s.UserC, s.RDPDomainUsersGroup, ad.MemberOf)
}

type RDPHarness struct {
	IrshadUser          *graph.Node
	EliUser             *graph.Node
	DillonUser          *graph.Node
	UliUser             *graph.Node
	AlyxUser            *graph.Node
	AndyUser            *graph.Node
	RohanUser           *graph.Node
	JohnUser            *graph.Node
	LocalGroupA         *graph.Node
	DomainGroupA        *graph.Node
	DomainGroupB        *graph.Node
	DomainGroupC        *graph.Node
	DomainGroupD        *graph.Node
	DomainGroupE        *graph.Node
	RDPLocalGroup       *graph.Node
	Computer            *graph.Node
	RDPDomainUsersGroup *graph.Node
}

func (s *RDPHarness) Setup(testCtx *GraphTestContext) {
	s.Computer = testCtx.NewActiveDirectoryComputer("WIN10", testCtx.Harness.RootADHarness.ActiveDirectoryDomainSID)
	s.RDPDomainUsersGroup = testCtx.NewActiveDirectoryGroup("RDP Domain Users", testCtx.Harness.RootADHarness.ActiveDirectoryDomainSID)
	s.RDPLocalGroup = testCtx.NewActiveDirectoryLocalGroup("Remote Desktop Users", testCtx.Harness.RootADHarness.ActiveDirectoryDomainSID)

	rdpLocalGroupObjectID, _ := s.RDPLocalGroup.Properties.Get(common.ObjectID.String()).String()

	s.RDPLocalGroup.Properties.Set(
		common.ObjectID.String(),
		rdpLocalGroupObjectID+adAnalysis.RDPGroupSuffix,
	)
	testCtx.UpdateNode(s.RDPLocalGroup)

	// Users
	s.IrshadUser = testCtx.NewActiveDirectoryUser("Irshad", testCtx.Harness.RootADHarness.ActiveDirectoryDomainSID)
	s.EliUser = testCtx.NewActiveDirectoryUser("Eli", testCtx.Harness.RootADHarness.ActiveDirectoryDomainSID)
	s.DillonUser = testCtx.NewActiveDirectoryUser("Dillon", testCtx.Harness.RootADHarness.ActiveDirectoryDomainSID)
	s.UliUser = testCtx.NewActiveDirectoryUser("Uli", testCtx.Harness.RootADHarness.ActiveDirectoryDomainSID)
	s.AlyxUser = testCtx.NewActiveDirectoryUser("Alyx", testCtx.Harness.RootADHarness.ActiveDirectoryDomainSID)
	s.AndyUser = testCtx.NewActiveDirectoryUser("Andy", testCtx.Harness.RootADHarness.ActiveDirectoryDomainSID)
	s.RohanUser = testCtx.NewActiveDirectoryUser("Rohan", testCtx.Harness.RootADHarness.ActiveDirectoryDomainSID)
	s.JohnUser = testCtx.NewActiveDirectoryUser("John", testCtx.Harness.RootADHarness.ActiveDirectoryDomainSID)

	// Groups
	s.LocalGroupA = testCtx.NewActiveDirectoryGroup("Local Group A", testCtx.Harness.RootADHarness.ActiveDirectoryDomainSID)
	s.DomainGroupA = testCtx.NewActiveDirectoryGroup("Domain Group A", testCtx.Harness.RootADHarness.ActiveDirectoryDomainSID)
	s.DomainGroupB = testCtx.NewActiveDirectoryGroup("Domain Group B", testCtx.Harness.RootADHarness.ActiveDirectoryDomainSID)
	s.DomainGroupC = testCtx.NewActiveDirectoryGroup("Domain Group C", testCtx.Harness.RootADHarness.ActiveDirectoryDomainSID)
	s.DomainGroupD = testCtx.NewActiveDirectoryGroup("Domain Group D", testCtx.Harness.RootADHarness.ActiveDirectoryDomainSID)
	s.DomainGroupE = testCtx.NewActiveDirectoryGroup("Domain Group E", testCtx.Harness.RootADHarness.ActiveDirectoryDomainSID)

	testCtx.NewRelationship(s.EliUser, s.RDPLocalGroup, ad.MemberOfLocalGroup, DefaultRelProperties)
	testCtx.NewRelationship(s.EliUser, s.Computer, ad.RemoteInteractiveLogonPrivilege, DefaultRelProperties)

	testCtx.NewRelationship(s.DomainGroupA, s.RDPLocalGroup, ad.MemberOfLocalGroup, DefaultRelProperties)
	testCtx.NewRelationship(s.DomainGroupA, s.Computer, ad.RemoteInteractiveLogonPrivilege, DefaultRelProperties)

	testCtx.NewRelationship(s.JohnUser, s.DomainGroupA, ad.MemberOf, DefaultRelProperties)
	testCtx.NewRelationship(s.RDPDomainUsersGroup, s.DomainGroupA, ad.MemberOf, DefaultRelProperties)
	testCtx.NewRelationship(s.JohnUser, s.RDPDomainUsersGroup, ad.MemberOf, DefaultRelProperties)
	testCtx.NewRelationship(s.RohanUser, s.RDPDomainUsersGroup, ad.MemberOf, DefaultRelProperties)

	testCtx.NewRelationship(s.DomainGroupB, s.Computer, ad.RemoteInteractiveLogonPrivilege, DefaultRelProperties)
	testCtx.NewRelationship(s.RohanUser, s.DomainGroupB, ad.MemberOf, DefaultRelProperties)

	testCtx.NewRelationship(s.AndyUser, s.DomainGroupB, ad.MemberOf, DefaultRelProperties)
	testCtx.NewRelationship(s.DomainGroupB, s.RDPLocalGroup, ad.MemberOfLocalGroup, DefaultRelProperties)
	testCtx.NewRelationship(s.DomainGroupB, s.Computer, ad.RemoteInteractiveLogonPrivilege, DefaultRelProperties)

	testCtx.NewRelationship(s.IrshadUser, s.RDPLocalGroup, ad.MemberOfLocalGroup, DefaultRelProperties)
	testCtx.NewRelationship(s.IrshadUser, s.DomainGroupD, ad.MemberOf, DefaultRelProperties)

	testCtx.NewRelationship(s.DomainGroupD, s.DomainGroupE, ad.MemberOf, DefaultRelProperties)

	testCtx.NewRelationship(s.DomainGroupE, s.Computer, ad.RemoteInteractiveLogonPrivilege, DefaultRelProperties)

	testCtx.NewRelationship(s.DomainGroupC, s.RDPLocalGroup, ad.MemberOfLocalGroup, DefaultRelProperties)
	testCtx.NewRelationship(s.DillonUser, s.DomainGroupC, ad.MemberOf, DefaultRelProperties)
	testCtx.NewRelationship(s.DillonUser, s.Computer, ad.RemoteInteractiveLogonPrivilege, DefaultRelProperties)

	testCtx.NewRelationship(s.UliUser, s.RDPLocalGroup, ad.MemberOfLocalGroup, DefaultRelProperties)
	testCtx.NewRelationship(s.UliUser, s.LocalGroupA, ad.MemberOfLocalGroup, DefaultRelProperties)
	testCtx.NewRelationship(s.LocalGroupA, s.Computer, ad.LocalToComputer, DefaultRelProperties)
	testCtx.NewRelationship(s.LocalGroupA, s.Computer, ad.RemoteInteractiveLogonPrivilege, DefaultRelProperties)

	testCtx.NewRelationship(s.AlyxUser, s.Computer, ad.RemoteInteractiveLogonPrivilege, DefaultRelProperties)

	testCtx.NewRelationship(s.RDPLocalGroup, s.Computer, ad.LocalToComputer, DefaultRelProperties)
}

type GPOEnforcementHarness struct {
	GPOEnforced         *graph.Node
	GPOUnenforced       *graph.Node
	Domain              *graph.Node
	OrganizationalUnitA *graph.Node
	OrganizationalUnitB *graph.Node
	OrganizationalUnitC *graph.Node
	OrganizationalUnitD *graph.Node
	UserA               *graph.Node
	UserB               *graph.Node
	UserC               *graph.Node
	UserD               *graph.Node
}

func (s *GPOEnforcementHarness) Setup(testCtx *GraphTestContext) {
	s.GPOEnforced = testCtx.NewActiveDirectoryGPO("Enforced GPO", testCtx.Harness.RootADHarness.ActiveDirectoryDomainSID)
	s.GPOUnenforced = testCtx.NewActiveDirectoryGPO("Unenforced GPO", testCtx.Harness.RootADHarness.ActiveDirectoryDomainSID)
	s.Domain = testCtx.NewActiveDirectoryDomain("TESTLAB.2", RandomDomainSID(), false, true)
	s.OrganizationalUnitA = testCtx.NewActiveDirectoryOU("OU A", testCtx.Harness.RootADHarness.ActiveDirectoryDomainSID, true)
	s.OrganizationalUnitB = testCtx.NewActiveDirectoryOU("OU B", testCtx.Harness.RootADHarness.ActiveDirectoryDomainSID, false)
	s.OrganizationalUnitC = testCtx.NewActiveDirectoryOU("OU C", testCtx.Harness.RootADHarness.ActiveDirectoryDomainSID, false)
	s.OrganizationalUnitD = testCtx.NewActiveDirectoryOU("OU D", testCtx.Harness.RootADHarness.ActiveDirectoryDomainSID, true)
	s.UserA = testCtx.NewActiveDirectoryUser("GPO Test User A", testCtx.Harness.RootADHarness.ActiveDirectoryDomainSID)
	s.UserB = testCtx.NewActiveDirectoryUser("GPO Test User B", testCtx.Harness.RootADHarness.ActiveDirectoryDomainSID)
	s.UserC = testCtx.NewActiveDirectoryUser("GPO Test User C", testCtx.Harness.RootADHarness.ActiveDirectoryDomainSID)
	s.UserD = testCtx.NewActiveDirectoryUser("GPO Test User D", testCtx.Harness.RootADHarness.ActiveDirectoryDomainSID)

	testCtx.NewRelationship(s.GPOUnenforced, s.Domain, ad.GPLink, DefaultRelProperties, graph.AsProperties(graph.PropertyMap{
		ad.Enforced: false,
	}))

	testCtx.NewRelationship(s.GPOEnforced, s.Domain, ad.GPLink, DefaultRelProperties, graph.AsProperties(graph.PropertyMap{
		ad.Enforced: true,
	}))

	testCtx.NewRelationship(s.Domain, s.OrganizationalUnitA, ad.Contains, DefaultRelProperties)
	testCtx.NewRelationship(s.Domain, s.OrganizationalUnitB, ad.Contains, DefaultRelProperties)
	testCtx.NewRelationship(s.OrganizationalUnitB, s.OrganizationalUnitD, ad.Contains, DefaultRelProperties)
	testCtx.NewRelationship(s.OrganizationalUnitA, s.OrganizationalUnitC, ad.Contains, DefaultRelProperties)
	testCtx.NewRelationship(s.OrganizationalUnitA, s.UserA, ad.Contains, DefaultRelProperties)
	testCtx.NewRelationship(s.OrganizationalUnitC, s.UserC, ad.Contains, DefaultRelProperties)
	testCtx.NewRelationship(s.OrganizationalUnitD, s.UserD, ad.Contains, DefaultRelProperties)
	testCtx.NewRelationship(s.OrganizationalUnitB, s.UserB, ad.Contains, DefaultRelProperties)
}

type AZBaseHarness struct {
	Tenant                *graph.Node
	User                  *graph.Node
	Application           *graph.Node
	ServicePrincipal      *graph.Node
	Nodes                 graph.NodeKindSet
	UserFirstDegreeGroups graph.NodeSet
	NumPaths              int
}

func (s *AZBaseHarness) Setup(testCtx *GraphTestContext) {
	const (
		numVMs    = 5
		numGroups = 5
		numRoles  = 5
	)
	tenantID := RandomObjectID(testCtx.testCtrl)

	s.Nodes = graph.NewNodeKindSet()
	s.Tenant = testCtx.NewAzureTenant(tenantID)
	s.User = testCtx.NewAzureUser(HarnessUserName, HarnessUserName, HarnessUserDescription, RandomObjectID(testCtx.testCtrl), HarnessUserLicenses, tenantID, HarnessUserMFAEnabled)
	s.Application = testCtx.NewAzureApplication(HarnessAppName, RandomObjectID(testCtx.testCtrl), tenantID)
	s.ServicePrincipal = testCtx.NewAzureServicePrincipal(HarnessServicePrincipalName, RandomObjectID(testCtx.testCtrl), tenantID)
	s.Nodes.Add(s.Tenant, s.User, s.Application, s.ServicePrincipal)
	s.UserFirstDegreeGroups = graph.NewNodeSet()
	s.NumPaths = 1287

	// Tie the user to the tenant and vice-versa
	// Note: This will cause a full re-traversal of paths outbound from the user object
	testCtx.NewRelationship(s.Tenant, s.User, azure.Contains)
	testCtx.NewRelationship(s.User, s.Tenant, azure.PrivilegedRoleAdmin)

	// Create some MemberOf relationships for the new user
	for nestingDepth := numGroups; nestingDepth > 0; nestingDepth-- {
		newGroups := s.CreateAzureNestedGroupChain(testCtx, tenantID, nestingDepth)
		s.Nodes.Add(newGroups.Slice()...)

		for _, newGroup := range newGroups {
			// Tie the groups to the tenant
			testCtx.NewRelationship(s.Tenant, newGroup, azure.Contains)
		}
	}

	// Create some VMs that the user has access to
	for vmIdx := 0; vmIdx < numVMs; vmIdx++ {
		newVM := testCtx.NewAzureVM(fmt.Sprintf("vm %d", vmIdx), RandomObjectID(testCtx.testCtrl), tenantID)
		s.Nodes.Add(newVM)

		// Tie the vm to the tenant
		testCtx.NewRelationship(s.Tenant, newVM, azure.Contains)

		// User has contributor rights to the new VM
		testCtx.NewRelationship(s.User, newVM, azure.Contributor)
	}

	// Create some role assignments for the user
	for roleIdx := 0; roleIdx < numRoles; roleIdx++ {
		var (
			objectID       = RandomObjectID(testCtx.testCtrl)
			roleTemplateID = RandomObjectID(testCtx.testCtrl)
			newRole        = testCtx.NewAzureRole(fmt.Sprintf("AZRole_%s", objectID), objectID, roleTemplateID, tenantID)
		)
		s.Nodes.Add(newRole)

		testCtx.NewRelationship(s.User, newRole, azure.HasRole)

		// Each role has contributor on all VMs, creating more attack paths
		for _, vm := range s.Nodes.Get(azure.VM) {
			testCtx.NewRelationship(newRole, vm, azure.Contributor)
		}

		// Roles may be granted by groups
		for _, group := range s.Nodes.Get(azure.Group) {
			testCtx.NewRelationship(group, newRole, azure.HasRole)
		}
	}

	// Tie the application and service principal to the user
	testCtx.NewRelationship(s.User, s.Application, azure.Owner)
	testCtx.NewRelationship(s.User, s.ServicePrincipal, azure.Owner)

	// Tie the service principal to the application
	testCtx.NewRelationship(s.Application, s.ServicePrincipal, azure.RunsAs)
}

func (s *AZBaseHarness) CreateAzureNestedGroupChain(testCtx *GraphTestContext, tenantID string, chainDepth int) graph.NodeSet {
	var (
		previousGroup *graph.Node
		groupNodes    = graph.NewNodeSet()
	)

	for groupIdx := 0; groupIdx < chainDepth; groupIdx++ {
		var (
			objectID = RandomObjectID(testCtx.testCtrl)
			newGroup = testCtx.NewAzureGroup(fmt.Sprintf("AZGroup_%s", objectID), objectID, tenantID)
		)

		if previousGroup == nil {
			testCtx.NewRelationship(s.User, newGroup, azure.MemberOf)
			s.UserFirstDegreeGroups.Add(newGroup)
		} else {
			testCtx.NewRelationship(previousGroup, newGroup, azure.MemberOf)
		}

		groupNodes.Add(newGroup)
		previousGroup = newGroup
	}

	return groupNodes
}

type AZGroupMembershipHarness struct {
	Tenant *graph.Node
	UserA  *graph.Node
	UserB  *graph.Node
	UserC  *graph.Node
	Group  *graph.Node
}

func (s *AZGroupMembershipHarness) Setup(testCtx *GraphTestContext) {
	tenantID := RandomObjectID(testCtx.testCtrl)
	s.UserA = testCtx.NewAzureUser("UserA", "UserA", "", RandomObjectID(testCtx.testCtrl), "", tenantID, false)
	s.UserB = testCtx.NewAzureUser("UserB", "UserB", "", RandomObjectID(testCtx.testCtrl), "", tenantID, false)
	s.UserC = testCtx.NewAzureUser("UserC", "UserC", "", RandomObjectID(testCtx.testCtrl), "", tenantID, false)
	s.Group = testCtx.NewAzureGroup("Group", RandomObjectID(testCtx.testCtrl), tenantID)

	testCtx.NewRelationship(s.UserA, s.Group, azure.MemberOf)
	testCtx.NewRelationship(s.UserB, s.Group, azure.MemberOf)
	testCtx.NewRelationship(s.UserC, s.Group, azure.MemberOf)
}

type AZEntityPanelHarness struct {
	Application      *graph.Node
	Device           *graph.Node
	Group            *graph.Node
	ManagementGroup  *graph.Node
	ResourceGroup    *graph.Node
	KeyVault         *graph.Node
	Role             *graph.Node
	ServicePrincipal *graph.Node
	Subscription     *graph.Node
	Tenant           *graph.Node
	User             *graph.Node
	VM               *graph.Node
}

func (s *AZEntityPanelHarness) Setup(testCtx *GraphTestContext) {
	tenantID := RandomObjectID(testCtx.testCtrl)
	s.Application = testCtx.NewAzureApplication("App", RandomObjectID(testCtx.testCtrl), tenantID)
	s.Device = testCtx.NewAzureDevice("Device", RandomObjectID(testCtx.testCtrl), RandomObjectID(testCtx.testCtrl), tenantID)
	s.Group = testCtx.NewAzureGroup("Group", RandomObjectID(testCtx.testCtrl), tenantID)
	s.ManagementGroup = testCtx.NewAzureResourceGroup("Mgmt Group", RandomObjectID(testCtx.testCtrl), tenantID)
	s.ResourceGroup = testCtx.NewAzureResourceGroup("Resource Group", RandomObjectID(testCtx.testCtrl), tenantID)
	s.KeyVault = testCtx.NewAzureKeyVault("Key Vault", RandomObjectID(testCtx.testCtrl), tenantID)
	s.Role = testCtx.NewAzureRole("Role", RandomObjectID(testCtx.testCtrl), RandomObjectID(testCtx.testCtrl), tenantID)
	s.ServicePrincipal = testCtx.NewAzureServicePrincipal("Service Principal", RandomObjectID(testCtx.testCtrl), tenantID)
	s.Subscription = testCtx.NewAzureSubscription("Sub", RandomObjectID(testCtx.testCtrl), tenantID)
	s.Tenant = testCtx.NewAzureTenant(tenantID)
	s.User = testCtx.NewAzureUser("User", "UserPrincipal", "Test User", RandomObjectID(testCtx.testCtrl), "Licenses", tenantID, false)
	s.VM = testCtx.NewAzureVM("VM", RandomObjectID(testCtx.testCtrl), tenantID)

	// Application
	testCtx.NewRelationship(s.User, s.Application, azure.Owner)

	// Device
	testCtx.NewRelationship(s.User, s.Device, azure.Owns)
	testCtx.NewRelationship(s.User, s.Device, azure.ExecuteCommand)

	// Groups
	testCtx.NewRelationship(s.User, s.Group, azure.Owns)
	testCtx.NewRelationship(s.User, s.ResourceGroup, azure.Owns)
	testCtx.NewRelationship(s.User, s.ManagementGroup, azure.Owner)

	// Key Vault
	testCtx.NewRelationship(s.User, s.KeyVault, azure.Owns)

	// Role
	testCtx.NewRelationship(s.Group, s.Role, azure.HasRole)
	testCtx.NewRelationship(s.User, s.Role, azure.HasRole)

	// Service Principal
	testCtx.NewRelationship(s.User, s.ServicePrincipal, azure.Owner)
	testCtx.NewRelationship(s.Application, s.ServicePrincipal, azure.RunsAs)

	// Subscription
	testCtx.NewRelationship(s.User, s.Subscription, azure.Owns)

	// Tenant
	testCtx.NewRelationship(s.User, s.Tenant, azure.PrivilegedRoleAdmin)

	// User
	testCtx.NewRelationship(s.Tenant, s.User, azure.Contains)

	// VM
	testCtx.NewRelationship(s.Tenant, s.VM, azure.Contains)
	testCtx.NewRelationship(s.User, s.VM, azure.Contributor)
	testCtx.NewRelationship(s.Role, s.VM, azure.Contributor)
}

type AZMGApplicationReadWriteAllHarness struct {
	Application       *graph.Node
	ServicePrincipal  *graph.Node
	ServicePrincipalB *graph.Node
	MicrosoftGraph    *graph.Node
	Tenant            *graph.Node
}

func (s *AZMGApplicationReadWriteAllHarness) Setup(testCtx *GraphTestContext) {
	tenantID := RandomObjectID(testCtx.testCtrl)
	s.Tenant = testCtx.NewAzureTenant(tenantID)
	s.MicrosoftGraph = testCtx.NewAzureServicePrincipal("Microsoft Graph", RandomObjectID(testCtx.testCtrl), tenantID)

	s.Application = testCtx.NewAzureApplication("App", RandomObjectID(testCtx.testCtrl), tenantID)
	s.ServicePrincipal = testCtx.NewAzureServicePrincipal("Service Principal", RandomObjectID(testCtx.testCtrl), tenantID)
	s.ServicePrincipalB = testCtx.NewAzureServicePrincipal("Service Principal B", RandomObjectID(testCtx.testCtrl), tenantID)

	testCtx.NewRelationship(s.Tenant, s.MicrosoftGraph, azure.Contains)
	testCtx.NewRelationship(s.Tenant, s.Application, azure.Contains)
	testCtx.NewRelationship(s.Tenant, s.ServicePrincipal, azure.Contains)
	testCtx.NewRelationship(s.Tenant, s.ServicePrincipalB, azure.Contains)

	testCtx.NewRelationship(s.ServicePrincipal, s.MicrosoftGraph, azure.ApplicationReadWriteAll)

	testCtx.NewRelationship(s.ServicePrincipal, s.MicrosoftGraph, azure.AZMGAddSecret)
	testCtx.NewRelationship(s.ServicePrincipal, s.MicrosoftGraph, azure.AZMGAddOwner)

	testCtx.NewRelationship(s.ServicePrincipal, s.Application, azure.AZMGAddSecret)
	testCtx.NewRelationship(s.ServicePrincipal, s.Application, azure.AZMGAddOwner)

	testCtx.NewRelationship(s.ServicePrincipal, s.ServicePrincipalB, azure.AZMGAddSecret)
	testCtx.NewRelationship(s.ServicePrincipal, s.ServicePrincipalB, azure.AZMGAddOwner)
}

type AZMGAppRoleManagementReadWriteAllHarness struct {
	ServicePrincipal *graph.Node
	MicrosoftGraph   *graph.Node
	Tenant           *graph.Node
}

func (s *AZMGAppRoleManagementReadWriteAllHarness) Setup(testCtx *GraphTestContext) {
	tenantID := RandomObjectID(testCtx.testCtrl)
	s.Tenant = testCtx.NewAzureTenant(tenantID)
	s.MicrosoftGraph = testCtx.NewAzureServicePrincipal("Microsoft Graph", RandomObjectID(testCtx.testCtrl), tenantID)

	s.ServicePrincipal = testCtx.NewAzureServicePrincipal("Service Principal", RandomObjectID(testCtx.testCtrl), tenantID)

	testCtx.NewRelationship(s.Tenant, s.MicrosoftGraph, azure.Contains)
	testCtx.NewRelationship(s.Tenant, s.ServicePrincipal, azure.Contains)

	testCtx.NewRelationship(s.ServicePrincipal, s.MicrosoftGraph, azure.AppRoleAssignmentReadWriteAll)

	testCtx.NewRelationship(s.ServicePrincipal, s.Tenant, azure.AZMGGrantAppRoles)
}

type AZMGDirectoryReadWriteAllHarness struct {
	Group            *graph.Node
	ServicePrincipal *graph.Node
	MicrosoftGraph   *graph.Node
	Tenant           *graph.Node
}

func (s *AZMGDirectoryReadWriteAllHarness) Setup(testCtx *GraphTestContext) {
	tenantID := RandomObjectID(testCtx.testCtrl)
	s.Tenant = testCtx.NewAzureTenant(tenantID)
	s.MicrosoftGraph = testCtx.NewAzureServicePrincipal("Microsoft Graph", RandomObjectID(testCtx.testCtrl), tenantID)

	s.Group = testCtx.NewAzureGroup("Group", RandomObjectID(testCtx.testCtrl), tenantID)
	s.ServicePrincipal = testCtx.NewAzureServicePrincipal("Service Principal", RandomObjectID(testCtx.testCtrl), tenantID)

	testCtx.NewRelationship(s.Tenant, s.MicrosoftGraph, azure.Contains)
	testCtx.NewRelationship(s.Tenant, s.Group, azure.Contains)
	testCtx.NewRelationship(s.Tenant, s.ServicePrincipal, azure.Contains)

	testCtx.NewRelationship(s.ServicePrincipal, s.MicrosoftGraph, azure.DirectoryReadWriteAll)
	testCtx.NewRelationship(s.ServicePrincipal, s.Group, azure.AZMGAddMember)
	testCtx.NewRelationship(s.ServicePrincipal, s.Group, azure.AZMGAddOwner)
}

type AZMGGroupReadWriteAllHarness struct {
	Group            *graph.Node
	ServicePrincipal *graph.Node
	MicrosoftGraph   *graph.Node
	Tenant           *graph.Node
}

func (s *AZMGGroupReadWriteAllHarness) Setup(testCtx *GraphTestContext) {
	tenantID := RandomObjectID(testCtx.testCtrl)
	s.Tenant = testCtx.NewAzureTenant(tenantID)
	s.MicrosoftGraph = testCtx.NewAzureServicePrincipal("Microsoft Graph", RandomObjectID(testCtx.testCtrl), tenantID)

	s.Group = testCtx.NewAzureGroup("Group", RandomObjectID(testCtx.testCtrl), tenantID)
	s.ServicePrincipal = testCtx.NewAzureServicePrincipal("Service Principal", RandomObjectID(testCtx.testCtrl), tenantID)

	testCtx.NewRelationship(s.Tenant, s.MicrosoftGraph, azure.Contains)
	testCtx.NewRelationship(s.Tenant, s.Group, azure.Contains)
	testCtx.NewRelationship(s.Tenant, s.ServicePrincipal, azure.Contains)

	testCtx.NewRelationship(s.ServicePrincipal, s.MicrosoftGraph, azure.GroupReadWriteAll)
	testCtx.NewRelationship(s.ServicePrincipal, s.Group, azure.AZMGAddMember)
	testCtx.NewRelationship(s.ServicePrincipal, s.Group, azure.AZMGAddOwner)
}

type AZMGGroupMemberReadWriteAllHarness struct {
	Group            *graph.Node
	ServicePrincipal *graph.Node
	MicrosoftGraph   *graph.Node
	Tenant           *graph.Node
}

func (s *AZMGGroupMemberReadWriteAllHarness) Setup(testCtx *GraphTestContext) {
	tenantID := RandomObjectID(testCtx.testCtrl)
	s.Tenant = testCtx.NewAzureTenant(tenantID)
	s.MicrosoftGraph = testCtx.NewAzureServicePrincipal("Microsoft Graph", RandomObjectID(testCtx.testCtrl), tenantID)

	s.Group = testCtx.NewAzureGroup("Group", RandomObjectID(testCtx.testCtrl), tenantID)
	s.ServicePrincipal = testCtx.NewAzureServicePrincipal("Service Principal", RandomObjectID(testCtx.testCtrl), tenantID)

	testCtx.NewRelationship(s.Tenant, s.MicrosoftGraph, azure.Contains)
	testCtx.NewRelationship(s.Tenant, s.Group, azure.Contains)
	testCtx.NewRelationship(s.Tenant, s.ServicePrincipal, azure.Contains)

	testCtx.NewRelationship(s.ServicePrincipal, s.MicrosoftGraph, azure.GroupMemberReadWriteAll)
	testCtx.NewRelationship(s.ServicePrincipal, s.Group, azure.AZMGAddMember)
	testCtx.NewRelationship(s.ServicePrincipal, s.Group, azure.AZMGAddOwner)
}

type AZMGRoleManagementReadWriteDirectoryHarness struct {
	Application       *graph.Node
	Group             *graph.Node
	Role              *graph.Node
	ServicePrincipal  *graph.Node
	ServicePrincipalB *graph.Node
	MicrosoftGraph    *graph.Node
	Tenant            *graph.Node
}

func (s *AZMGRoleManagementReadWriteDirectoryHarness) Setup(testCtx *GraphTestContext) {
	tenantID := RandomObjectID(testCtx.testCtrl)
	s.Tenant = testCtx.NewAzureTenant(tenantID)
	s.MicrosoftGraph = testCtx.NewAzureServicePrincipal("Microsoft Graph", RandomObjectID(testCtx.testCtrl), tenantID)

	s.Application = testCtx.NewAzureApplication("App", RandomObjectID(testCtx.testCtrl), tenantID)
	s.Group = testCtx.NewAzureGroup("Group", RandomObjectID(testCtx.testCtrl), tenantID)
	s.Role = testCtx.NewAzureRole("Role", RandomObjectID(testCtx.testCtrl), RandomObjectID(testCtx.testCtrl), tenantID)
	s.ServicePrincipal = testCtx.NewAzureServicePrincipal("Service Principal", RandomObjectID(testCtx.testCtrl), tenantID)
	s.ServicePrincipalB = testCtx.NewAzureServicePrincipal("Service Principal B", RandomObjectID(testCtx.testCtrl), tenantID)

	testCtx.NewRelationship(s.Tenant, s.MicrosoftGraph, azure.Contains)
	testCtx.NewRelationship(s.Tenant, s.Application, azure.Contains)
	testCtx.NewRelationship(s.Tenant, s.Role, azure.Contains)
	testCtx.NewRelationship(s.Tenant, s.Group, azure.Contains)
	testCtx.NewRelationship(s.Tenant, s.ServicePrincipal, azure.Contains)
	testCtx.NewRelationship(s.Tenant, s.ServicePrincipalB, azure.Contains)

	testCtx.NewRelationship(s.ServicePrincipal, s.MicrosoftGraph, azure.RoleManagementReadWriteDirectory)

	testCtx.NewRelationship(s.ServicePrincipal, s.MicrosoftGraph, azure.AZMGAddSecret)
	testCtx.NewRelationship(s.ServicePrincipal, s.MicrosoftGraph, azure.AZMGAddOwner)

	testCtx.NewRelationship(s.ServicePrincipal, s.ServicePrincipalB, azure.AZMGAddSecret)
	testCtx.NewRelationship(s.ServicePrincipal, s.ServicePrincipalB, azure.AZMGAddOwner)

	testCtx.NewRelationship(s.ServicePrincipal, s.Application, azure.AZMGAddSecret)
	testCtx.NewRelationship(s.ServicePrincipal, s.Application, azure.AZMGAddOwner)

	testCtx.NewRelationship(s.ServicePrincipal, s.Group, azure.AZMGAddSecret)
	testCtx.NewRelationship(s.ServicePrincipal, s.Group, azure.AZMGAddOwner)

	testCtx.NewRelationship(s.ServicePrincipal, s.Role, azure.AZMGGrantRole)
}

type AZMGServicePrincipalEndpointReadWriteAllHarness struct {
	ServicePrincipal  *graph.Node
	ServicePrincipalB *graph.Node
	MicrosoftGraph    *graph.Node
	Tenant            *graph.Node
}

func (s *AZMGServicePrincipalEndpointReadWriteAllHarness) Setup(testCtx *GraphTestContext) {
	tenantID := RandomObjectID(testCtx.testCtrl)
	s.Tenant = testCtx.NewAzureTenant(tenantID)
	s.MicrosoftGraph = testCtx.NewAzureServicePrincipal("Microsoft Graph", RandomObjectID(testCtx.testCtrl), tenantID)

	s.ServicePrincipal = testCtx.NewAzureServicePrincipal("Service Principal", RandomObjectID(testCtx.testCtrl), tenantID)
	s.ServicePrincipalB = testCtx.NewAzureServicePrincipal("Service Principal B", RandomObjectID(testCtx.testCtrl), tenantID)

	testCtx.NewRelationship(s.Tenant, s.MicrosoftGraph, azure.Contains)
	testCtx.NewRelationship(s.Tenant, s.ServicePrincipal, azure.Contains)
	testCtx.NewRelationship(s.Tenant, s.ServicePrincipalB, azure.Contains)

	testCtx.NewRelationship(s.ServicePrincipal, s.MicrosoftGraph, azure.ServicePrincipalEndpointReadWriteAll)

	testCtx.NewRelationship(s.ServicePrincipal, s.MicrosoftGraph, azure.AZMGAddOwner)

	testCtx.NewRelationship(s.ServicePrincipal, s.ServicePrincipalB, azure.AZMGAddOwner)
}

type AZInboundControlHarness struct {
	ControlledAZUser    *graph.Node
	AZAppA              *graph.Node
	AZGroupA            *graph.Node
	AZGroupB            *graph.Node
	AZUserA             *graph.Node
	AZUserB             *graph.Node
	AZServicePrincipalA *graph.Node
	AZServicePrincipalB *graph.Node
}

func (s *AZInboundControlHarness) Setup(testCtx *GraphTestContext) {
	tenantID := RandomObjectID(testCtx.testCtrl)
	s.ControlledAZUser = testCtx.NewAzureUser("Controlled AZUser", "Controlled AZUser", "", RandomObjectID(testCtx.testCtrl), HarnessUserLicenses, tenantID, HarnessUserMFAEnabled)
	s.AZAppA = testCtx.NewAzureApplication("AZAppA", RandomObjectID(testCtx.testCtrl), tenantID)
	s.AZGroupA = testCtx.NewAzureGroup("AZGroupA", RandomObjectID(testCtx.testCtrl), tenantID)
	s.AZGroupB = testCtx.NewAzureGroup("AZGroupB", RandomObjectID(testCtx.testCtrl), tenantID)
	s.AZUserA = testCtx.NewAzureUser("AZUserA", "AZUserA", "", RandomObjectID(testCtx.testCtrl), HarnessUserLicenses, tenantID, HarnessUserMFAEnabled)
	s.AZUserB = testCtx.NewAzureUser("AZUserB", "AZUserB", "", RandomObjectID(testCtx.testCtrl), HarnessUserLicenses, tenantID, HarnessUserMFAEnabled)
	s.AZServicePrincipalA = testCtx.NewAzureServicePrincipal("AZServicePrincipalA", RandomObjectID(testCtx.testCtrl), tenantID)
	s.AZServicePrincipalB = testCtx.NewAzureServicePrincipal("AZServicePrincipalB", RandomObjectID(testCtx.testCtrl), tenantID)

	testCtx.NewRelationship(s.AZUserA, s.AZGroupA, azure.MemberOf)
	testCtx.NewRelationship(s.AZServicePrincipalB, s.AZGroupB, azure.MemberOf)

	testCtx.NewRelationship(s.AZAppA, s.AZServicePrincipalA, azure.RunsAs)

	testCtx.NewRelationship(s.AZGroupA, s.ControlledAZUser, azure.ResetPassword)
	testCtx.NewRelationship(s.AZGroupB, s.ControlledAZUser, azure.ResetPassword)
	testCtx.NewRelationship(s.AZUserB, s.ControlledAZUser, azure.ResetPassword)
	testCtx.NewRelationship(s.AZServicePrincipalA, s.ControlledAZUser, azure.ResetPassword)
}

type SearchHarness struct {
	User1      *graph.Node
	User2      *graph.Node
	User3      *graph.Node
	User4      *graph.Node
	User5      *graph.Node
	LocalGroup *graph.Node
}

func (s *SearchHarness) Setup(graphTestContext *GraphTestContext) {
	s.User1 = graphTestContext.NewActiveDirectoryUser("USER NUMBER ONE", graphTestContext.Harness.RootADHarness.ActiveDirectoryDomainSID)
	s.User2 = graphTestContext.NewActiveDirectoryUser("USER NUMBER TWO", graphTestContext.Harness.RootADHarness.ActiveDirectoryDomainSID)
	s.User3 = graphTestContext.NewActiveDirectoryUser("USER NUMBER THREE", graphTestContext.Harness.RootADHarness.ActiveDirectoryDomainSID)
	s.User4 = graphTestContext.NewActiveDirectoryUser("USER NUMBER FOUR", graphTestContext.Harness.RootADHarness.ActiveDirectoryDomainSID)
	s.User5 = graphTestContext.NewActiveDirectoryUser("USER NUMBER FIVE", graphTestContext.Harness.RootADHarness.ActiveDirectoryDomainSID)
	s.LocalGroup = graphTestContext.NewActiveDirectoryLocalGroup("REMOTE DESKTOP USERS", graphTestContext.Harness.RootADHarness.ActiveDirectoryDomainSID)
}

type RootADHarness struct {
	TierZero                                graph.NodeSet
	ActiveDirectoryDomainSID                string
	ActiveDirectoryDomain                   *graph.Node
	ActiveDirectoryRDPDomainGroup           *graph.Node
	ActiveDirectoryDomainUsersGroup         *graph.Node
	ActiveDirectoryUser                     *graph.Node
	ActiveDirectoryOU                       *graph.Node
	ActiveDirectoryGPO                      *graph.Node
	ActiveDirectoryDCSyncMetaRelationship   *graph.Relationship
	ActiveDirectoryDCSyncAtomicRelationship *graph.Relationship
	NumCollectedDomains                     int
}

func (s *RootADHarness) Setup(graphTestContext *GraphTestContext) {
	s.ActiveDirectoryDomainSID = RandomDomainSID()
	s.ActiveDirectoryDomain = graphTestContext.NewActiveDirectoryDomain("TESTLAB.LOCAL", s.ActiveDirectoryDomainSID, false, true)
	s.ActiveDirectoryUser = graphTestContext.NewActiveDirectoryUser("AD User", s.ActiveDirectoryDomainSID)
	s.ActiveDirectoryOU = graphTestContext.NewActiveDirectoryOU("OU", s.ActiveDirectoryDomainSID, false)
	s.ActiveDirectoryGPO = graphTestContext.NewActiveDirectoryGPO("GPO Policy", s.ActiveDirectoryDomainSID)
	s.ActiveDirectoryDomainUsersGroup = graphTestContext.NewActiveDirectoryGroup("Domain Users", s.ActiveDirectoryDomainSID)

	// Allow the user to do DCSync to the domain
	s.ActiveDirectoryDCSyncAtomicRelationship = graphTestContext.NewRelationship(s.ActiveDirectoryUser, s.ActiveDirectoryDomain, ad.DCSync, DefaultRelProperties)

	// Contain the OU and user
	graphTestContext.NewRelationship(s.ActiveDirectoryDomain, s.ActiveDirectoryOU, ad.Contains)
	graphTestContext.NewRelationship(s.ActiveDirectoryOU, s.ActiveDirectoryUser, ad.Contains)
	graphTestContext.NewRelationship(s.ActiveDirectoryUser, s.ActiveDirectoryDomainUsersGroup, ad.MemberOf)

	// Apply a GPO
	graphTestContext.NewRelationship(s.ActiveDirectoryGPO, s.ActiveDirectoryOU, ad.GPLink)
}

type HarnessDetails struct {
	RDP                                             RDPHarness
	RDPB                                            RDPHarness2
	GPOEnforcement                                  GPOEnforcementHarness
	Session                                         SessionHarness
	LocalGroupSQL                                   LocalGroupHarness
	OutboundControl                                 OutboundControlHarness
	InboundControl                                  InboundControlHarness
	AssetGroupComboNodeHarness                      AssetGroupComboNodeHarness
	OUHarness                                       OUContainedHarness
	MembershipHarness                               MembershipHarness
	ForeignHarness                                  ForeignDomainHarness
	TrustDCSync                                     TrustDCSyncHarness
	Completeness                                    CompletenessHarness
	AZBaseHarness                                   AZBaseHarness
	AZGroupMembership                               AZGroupMembershipHarness
	AZEntityPanelHarness                            AZEntityPanelHarness
	AZMGApplicationReadWriteAllHarness              AZMGApplicationReadWriteAllHarness
	AZMGAppRoleManagementReadWriteAllHarness        AZMGAppRoleManagementReadWriteAllHarness
	AZMGDirectoryReadWriteAllHarness                AZMGDirectoryReadWriteAllHarness
	AZMGGroupReadWriteAllHarness                    AZMGGroupReadWriteAllHarness
	AZMGGroupMemberReadWriteAllHarness              AZMGGroupMemberReadWriteAllHarness
	AZMGRoleManagementReadWriteDirectoryHarness     AZMGRoleManagementReadWriteDirectoryHarness
	AZMGServicePrincipalEndpointReadWriteAllHarness AZMGServicePrincipalEndpointReadWriteAllHarness
	RootADHarness                                   RootADHarness
	SearchHarness                                   SearchHarness
	NumCollectedActiveDirectoryDomains              int
	AZInboundControlHarness                         AZInboundControlHarness
}
