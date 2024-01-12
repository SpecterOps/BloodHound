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
	User1           *graph.Node
	User2           *graph.Node
	User3           *graph.Node
	User4           *graph.Node
	User5           *graph.Node
	LocalGroup      *graph.Node
	GroupLocalGroup *graph.Node
}

func (s *SearchHarness) Setup(graphTestContext *GraphTestContext) {
	s.User1 = graphTestContext.NewActiveDirectoryUser("USER NUMBER ONE", graphTestContext.Harness.RootADHarness.ActiveDirectoryDomainSID)
	s.User2 = graphTestContext.NewActiveDirectoryUser("USER NUMBER TWO", graphTestContext.Harness.RootADHarness.ActiveDirectoryDomainSID)
	s.User3 = graphTestContext.NewActiveDirectoryUser("USER NUMBER THREE", graphTestContext.Harness.RootADHarness.ActiveDirectoryDomainSID)
	s.User4 = graphTestContext.NewActiveDirectoryUser("USER NUMBER FOUR", graphTestContext.Harness.RootADHarness.ActiveDirectoryDomainSID)
	s.User5 = graphTestContext.NewActiveDirectoryUser("USER NUMBER FIVE", graphTestContext.Harness.RootADHarness.ActiveDirectoryDomainSID)

	s.LocalGroup = graphTestContext.NewActiveDirectoryLocalGroup("REMOTE DESKTOP USERS", graphTestContext.Harness.RootADHarness.ActiveDirectoryDomainSID)

	s.GroupLocalGroup = graphTestContext.NewActiveDirectoryLocalGroup("ACCOUNT OPERATORS", graphTestContext.Harness.RootADHarness.ActiveDirectoryDomainSID)
	s.GroupLocalGroup.AddKinds(ad.Group)
	graphTestContext.UpdateNode(s.GroupLocalGroup)
}

type ADCSESC1Harness struct {
	AuthStore1    *graph.Node
	RootCA1       *graph.Node
	EnterpriseCA1 *graph.Node
	CertTemplate1 *graph.Node
	Domain1       *graph.Node
	Group11       *graph.Node
	Group12       *graph.Node
	Group13       *graph.Node
	User11        *graph.Node
	User12        *graph.Node
	User13        *graph.Node
	User14        *graph.Node
	User15        *graph.Node

	Domain2        *graph.Node
	RootCA2        *graph.Node
	AuthStore2     *graph.Node
	CertTemplate2  *graph.Node
	EnterpriseCA21 *graph.Node
	EnterpriseCA22 *graph.Node
	Group21        *graph.Node
	Group22        *graph.Node

	Domain3        *graph.Node
	RootCA3        *graph.Node
	AuthStore3     *graph.Node
	EnterpriseCA31 *graph.Node
	EnterpriseCA32 *graph.Node
	CertTemplate3  *graph.Node
	Group31        *graph.Node
	Group32        *graph.Node

	Domain4        *graph.Node
	AuthStore4     *graph.Node
	RootCA4        *graph.Node
	Group41        *graph.Node
	Group42        *graph.Node
	Group43        *graph.Node
	Group44        *graph.Node
	Group45        *graph.Node
	Group46        *graph.Node
	EnterpriseCA4  *graph.Node
	CertTemplate41 *graph.Node
	CertTemplate42 *graph.Node
	CertTemplate43 *graph.Node
	CertTemplate44 *graph.Node
	CertTemplate45 *graph.Node
	CertTemplate46 *graph.Node
}

func (s *ADCSESC1Harness) Setup(graphTestContext *GraphTestContext) {
	emptyEkus := make([]string, 0)
	sid := RandomDomainSID()
	s.Domain1 = graphTestContext.NewActiveDirectoryDomain("domain 1", sid, false, true)
	s.AuthStore1 = graphTestContext.NewActiveDirectoryNTAuthStore("ntauthstore 1", sid)
	s.EnterpriseCA1 = graphTestContext.NewActiveDirectoryEnterpriseCA("eca 1", sid)
	s.RootCA1 = graphTestContext.NewActiveDirectoryRootCA("rca 1", sid)
	s.CertTemplate1 = graphTestContext.NewActiveDirectoryCertTemplate("certtemplate 1", sid, false, true, true, false, 1, 0, emptyEkus, emptyEkus)
	s.Group11 = graphTestContext.NewActiveDirectoryGroup("group1-1", sid)
	s.Group12 = graphTestContext.NewActiveDirectoryGroup("group1-2", sid)
	s.Group13 = graphTestContext.NewActiveDirectoryGroup("group1-3", sid)
	s.User11 = graphTestContext.NewActiveDirectoryUser("user1-1", sid)
	s.User12 = graphTestContext.NewActiveDirectoryUser("user1-2", sid)
	s.User13 = graphTestContext.NewActiveDirectoryUser("user1-3", sid)
	s.User14 = graphTestContext.NewActiveDirectoryUser("user1-4", sid)
	s.User15 = graphTestContext.NewActiveDirectoryUser("user1-5", sid)

	graphTestContext.NewRelationship(s.AuthStore1, s.Domain1, ad.NTAuthStoreFor)
	graphTestContext.NewRelationship(s.RootCA1, s.Domain1, ad.RootCAFor)
	graphTestContext.NewRelationship(s.EnterpriseCA1, s.AuthStore1, ad.TrustedForNTAuth)
	graphTestContext.NewRelationship(s.EnterpriseCA1, s.RootCA1, ad.EnterpriseCAFor)
	graphTestContext.NewRelationship(s.CertTemplate1, s.EnterpriseCA1, ad.PublishedTo)
	graphTestContext.NewRelationship(s.Group11, s.CertTemplate1, ad.Enroll)
	graphTestContext.NewRelationship(s.Group12, s.EnterpriseCA1, ad.Enroll)
	graphTestContext.NewRelationship(s.Group13, s.EnterpriseCA1, ad.Enroll)
	graphTestContext.NewRelationship(s.Group13, s.CertTemplate1, ad.Enroll)

	graphTestContext.NewRelationship(s.User15, s.Group11, ad.MemberOf)
	graphTestContext.NewRelationship(s.User13, s.Group11, ad.MemberOf)
	graphTestContext.NewRelationship(s.User11, s.Group11, ad.MemberOf)
	graphTestContext.NewRelationship(s.User13, s.Group12, ad.MemberOf)
	graphTestContext.NewRelationship(s.User14, s.Group12, ad.MemberOf)
	graphTestContext.NewRelationship(s.User12, s.Group13, ad.MemberOf)
	graphTestContext.NewRelationship(s.User11, s.Group13, ad.MemberOf)

	sid = RandomDomainSID()
	s.Domain2 = graphTestContext.NewActiveDirectoryDomain("domain 2", sid, false, true)
	s.RootCA2 = graphTestContext.NewActiveDirectoryRootCA("rca2", sid)
	s.AuthStore2 = graphTestContext.NewActiveDirectoryNTAuthStore("authstore2", sid)
	s.EnterpriseCA21 = graphTestContext.NewActiveDirectoryEnterpriseCA("eca2-1", sid)
	s.EnterpriseCA22 = graphTestContext.NewActiveDirectoryEnterpriseCA("eca2-2", sid)
	s.Group21 = graphTestContext.NewActiveDirectoryGroup("group2-1", sid)
	s.Group22 = graphTestContext.NewActiveDirectoryGroup("group2-2", sid)
	s.CertTemplate2 = graphTestContext.NewActiveDirectoryCertTemplate("certtemplate 2", sid, false, true, true, false, 1, 0, emptyEkus, emptyEkus)

	graphTestContext.NewRelationship(s.RootCA2, s.Domain2, ad.RootCAFor)
	graphTestContext.NewRelationship(s.AuthStore2, s.Domain2, ad.NTAuthStoreFor)
	graphTestContext.NewRelationship(s.EnterpriseCA21, s.AuthStore2, ad.TrustedForNTAuth)
	graphTestContext.NewRelationship(s.EnterpriseCA21, s.RootCA2, ad.EnterpriseCAFor)
	graphTestContext.NewRelationship(s.EnterpriseCA22, s.RootCA2, ad.IssuedSignedBy)
	graphTestContext.NewRelationship(s.Group21, s.EnterpriseCA22, ad.Enroll)
	graphTestContext.NewRelationship(s.Group21, s.CertTemplate2, ad.Enroll)
	graphTestContext.NewRelationship(s.CertTemplate2, s.EnterpriseCA22, ad.PublishedTo)
	graphTestContext.NewRelationship(s.CertTemplate2, s.EnterpriseCA21, ad.PublishedTo)
	graphTestContext.NewRelationship(s.Group22, s.CertTemplate2, ad.Enroll)
	graphTestContext.NewRelationship(s.Group22, s.EnterpriseCA21, ad.Enroll)

	sid = RandomDomainSID()
	s.Domain3 = graphTestContext.NewActiveDirectoryDomain("domain 3", sid, false, true)
	s.RootCA3 = graphTestContext.NewActiveDirectoryRootCA("rca3", sid)
	s.AuthStore3 = graphTestContext.NewActiveDirectoryNTAuthStore("authstore3", sid)
	s.EnterpriseCA31 = graphTestContext.NewActiveDirectoryEnterpriseCA("eca3-1", sid)
	s.EnterpriseCA32 = graphTestContext.NewActiveDirectoryEnterpriseCA("eca3-2", sid)
	s.Group31 = graphTestContext.NewActiveDirectoryGroup("group3-1", sid)
	s.Group32 = graphTestContext.NewActiveDirectoryGroup("group3-2", sid)
	s.CertTemplate3 = graphTestContext.NewActiveDirectoryCertTemplate("certtemplate 3", sid, false, true, true, false, 1, 0, emptyEkus, emptyEkus)

	graphTestContext.NewRelationship(s.RootCA3, s.Domain3, ad.RootCAFor)
	graphTestContext.NewRelationship(s.AuthStore3, s.Domain3, ad.NTAuthStoreFor)
	graphTestContext.NewRelationship(s.EnterpriseCA31, s.AuthStore3, ad.TrustedForNTAuth)
	graphTestContext.NewRelationship(s.EnterpriseCA31, s.RootCA3, ad.EnterpriseCAFor)
	graphTestContext.NewRelationship(s.EnterpriseCA32, s.AuthStore3, ad.TrustedForNTAuth)
	graphTestContext.NewRelationship(s.CertTemplate3, s.EnterpriseCA31, ad.PublishedTo)
	graphTestContext.NewRelationship(s.CertTemplate3, s.EnterpriseCA32, ad.PublishedTo)
	graphTestContext.NewRelationship(s.Group31, s.EnterpriseCA32, ad.Enroll)
	graphTestContext.NewRelationship(s.Group31, s.CertTemplate3, ad.Enroll)
	graphTestContext.NewRelationship(s.Group32, s.CertTemplate3, ad.Enroll)
	graphTestContext.NewRelationship(s.Group32, s.EnterpriseCA31, ad.Enroll)

	sid = RandomDomainSID()
	s.Domain4 = graphTestContext.NewActiveDirectoryDomain("domain 4", sid, false, true)
	s.AuthStore4 = graphTestContext.NewActiveDirectoryNTAuthStore("authstore 4", sid)
	s.RootCA4 = graphTestContext.NewActiveDirectoryRootCA("rca4", sid)
	s.EnterpriseCA4 = graphTestContext.NewActiveDirectoryEnterpriseCA("eca4", sid)
	s.Group41 = graphTestContext.NewActiveDirectoryGroup("group4-1", sid)
	s.Group42 = graphTestContext.NewActiveDirectoryGroup("group4-2", sid)
	s.Group43 = graphTestContext.NewActiveDirectoryGroup("group4-3", sid)
	s.Group44 = graphTestContext.NewActiveDirectoryGroup("group4-4", sid)
	s.Group45 = graphTestContext.NewActiveDirectoryGroup("group4-5", sid)
	s.Group46 = graphTestContext.NewActiveDirectoryGroup("group4-6", sid)
	s.CertTemplate41 = graphTestContext.NewActiveDirectoryCertTemplate("certtemplate 4-1", sid, false, true, true, false, 2, 1, emptyEkus, emptyEkus)
	s.CertTemplate42 = graphTestContext.NewActiveDirectoryCertTemplate("certtemplate 4-2", sid, false, true, true, false, 2, 0, emptyEkus, emptyEkus)
	s.CertTemplate43 = graphTestContext.NewActiveDirectoryCertTemplate("certtemplate 4-3", sid, false, true, true, false, 1, 0, emptyEkus, emptyEkus)
	s.CertTemplate44 = graphTestContext.NewActiveDirectoryCertTemplate("certtemplate 4-4", sid, true, true, true, false, 1, 0, emptyEkus, emptyEkus)
	s.CertTemplate45 = graphTestContext.NewActiveDirectoryCertTemplate("certtemplate 4-5", sid, false, false, true, false, 1, 0, emptyEkus, emptyEkus)
	s.CertTemplate46 = graphTestContext.NewActiveDirectoryCertTemplate("certtemplate 4-6", sid, false, true, false, true, 1, 0, emptyEkus, emptyEkus)

	graphTestContext.NewRelationship(s.AuthStore4, s.Domain4, ad.NTAuthStoreFor)
	graphTestContext.NewRelationship(s.RootCA4, s.Domain4, ad.RootCAFor)
	graphTestContext.NewRelationship(s.EnterpriseCA4, s.AuthStore4, ad.TrustedForNTAuth)
	graphTestContext.NewRelationship(s.EnterpriseCA4, s.RootCA4, ad.EnterpriseCAFor)
	graphTestContext.NewRelationship(s.Group41, s.EnterpriseCA4, ad.Enroll)
	graphTestContext.NewRelationship(s.Group41, s.CertTemplate41, ad.Enroll)
	graphTestContext.NewRelationship(s.Group42, s.EnterpriseCA4, ad.Enroll)
	graphTestContext.NewRelationship(s.Group42, s.CertTemplate42, ad.Enroll)
	graphTestContext.NewRelationship(s.Group43, s.EnterpriseCA4, ad.Enroll)
	graphTestContext.NewRelationship(s.Group43, s.CertTemplate43, ad.Enroll)
	graphTestContext.NewRelationship(s.Group44, s.EnterpriseCA4, ad.Enroll)
	graphTestContext.NewRelationship(s.Group44, s.CertTemplate44, ad.Enroll)
	graphTestContext.NewRelationship(s.Group45, s.EnterpriseCA4, ad.Enroll)
	graphTestContext.NewRelationship(s.Group45, s.CertTemplate45, ad.Enroll)
	graphTestContext.NewRelationship(s.Group46, s.EnterpriseCA4, ad.Enroll)
	graphTestContext.NewRelationship(s.Group46, s.CertTemplate46, ad.Enroll)
	graphTestContext.NewRelationship(s.CertTemplate41, s.EnterpriseCA4, ad.PublishedTo)
	graphTestContext.NewRelationship(s.CertTemplate42, s.EnterpriseCA4, ad.PublishedTo)
	graphTestContext.NewRelationship(s.CertTemplate43, s.EnterpriseCA4, ad.PublishedTo)
	graphTestContext.NewRelationship(s.CertTemplate44, s.EnterpriseCA4, ad.PublishedTo)
	graphTestContext.NewRelationship(s.CertTemplate45, s.EnterpriseCA4, ad.PublishedTo)
	graphTestContext.NewRelationship(s.CertTemplate46, s.EnterpriseCA4, ad.PublishedTo)
}

type EnrollOnBehalfOfHarnessTwo struct {
	Domain2        *graph.Node
	AuthStore2     *graph.Node
	RootCA2        *graph.Node
	EnterpriseCA2  *graph.Node
	CertTemplate21 *graph.Node
	CertTemplate22 *graph.Node
	CertTemplate23 *graph.Node
	CertTemplate24 *graph.Node
	CertTemplate25 *graph.Node
}

func (s *EnrollOnBehalfOfHarnessTwo) Setup(gt *GraphTestContext) {
	certRequestAgentEKU := make([]string, 0)
	certRequestAgentEKU = append(certRequestAgentEKU, adAnalysis.EkuCertRequestAgent)
	emptyAppPolicies := make([]string, 0)
	sid := RandomDomainSID()
	s.Domain2 = gt.NewActiveDirectoryDomain("domain2", sid, false, true)
	s.AuthStore2 = gt.NewActiveDirectoryNTAuthStore("authstore2", sid)
	s.RootCA2 = gt.NewActiveDirectoryRootCA("rca2", sid)
	s.EnterpriseCA2 = gt.NewActiveDirectoryEnterpriseCA("eca2", sid)
	s.CertTemplate21 = gt.NewActiveDirectoryCertTemplate("certtemplate2-1", sid, false, false, false, false, 1, 0, certRequestAgentEKU, emptyAppPolicies)
	s.CertTemplate22 = gt.NewActiveDirectoryCertTemplate("certtemplate2-2", sid, false, false, false, false, 1, 0, []string{adAnalysis.EkuCertRequestAgent, adAnalysis.EkuAnyPurpose}, emptyAppPolicies)
	s.CertTemplate23 = gt.NewActiveDirectoryCertTemplate("certtemplate2-3", sid, false, false, false, false, 2, 1, certRequestAgentEKU, []string{adAnalysis.EkuCertRequestAgent})
	s.CertTemplate24 = gt.NewActiveDirectoryCertTemplate("certtemplate2-4", sid, false, false, false, false, 2, 1, []string{}, emptyAppPolicies)
	s.CertTemplate25 = gt.NewActiveDirectoryCertTemplate("certtemplate2-5", sid, false, false, false, true, 1, 1, []string{}, emptyAppPolicies)

	gt.NewRelationship(s.AuthStore2, s.Domain2, ad.NTAuthStoreFor)
	gt.NewRelationship(s.RootCA2, s.Domain2, ad.RootCAFor)
	gt.NewRelationship(s.EnterpriseCA2, s.AuthStore2, ad.TrustedForNTAuth)
	gt.NewRelationship(s.EnterpriseCA2, s.RootCA2, ad.EnterpriseCAFor)
	gt.NewRelationship(s.CertTemplate21, s.EnterpriseCA2, ad.PublishedTo)
	gt.NewRelationship(s.CertTemplate22, s.EnterpriseCA2, ad.PublishedTo)
	gt.NewRelationship(s.CertTemplate23, s.EnterpriseCA2, ad.PublishedTo)
	gt.NewRelationship(s.CertTemplate24, s.EnterpriseCA2, ad.PublishedTo)
	gt.NewRelationship(s.CertTemplate25, s.EnterpriseCA2, ad.PublishedTo)
}

type EnrollOnBehalfOfHarnessOne struct {
	Domain1        *graph.Node
	AuthStore1     *graph.Node
	RootCA1        *graph.Node
	EnterpriseCA1  *graph.Node
	CertTemplate11 *graph.Node
	CertTemplate12 *graph.Node
	CertTemplate13 *graph.Node
}

func (s *EnrollOnBehalfOfHarnessOne) Setup(gt *GraphTestContext) {
	sid := RandomDomainSID()
	anyPurposeEkus := make([]string, 0)
	anyPurposeEkus = append(anyPurposeEkus, adAnalysis.EkuAnyPurpose)
	emptyAppPolicies := make([]string, 0)
	s.Domain1 = gt.NewActiveDirectoryDomain("domain1", sid, false, true)
	s.AuthStore1 = gt.NewActiveDirectoryNTAuthStore("authstore1", sid)
	s.RootCA1 = gt.NewActiveDirectoryRootCA("rca1", sid)
	s.EnterpriseCA1 = gt.NewActiveDirectoryEnterpriseCA("eca1", sid)
	s.CertTemplate11 = gt.NewActiveDirectoryCertTemplate("certtemplate1-1", sid, false, false, false, false, 2, 0, anyPurposeEkus, emptyAppPolicies)
	s.CertTemplate12 = gt.NewActiveDirectoryCertTemplate("certtemplate1-2", sid, false, false, false, false, 1, 0, anyPurposeEkus, emptyAppPolicies)
	s.CertTemplate13 = gt.NewActiveDirectoryCertTemplate("certtemplate1-3", sid, false, false, false, false, 2, 0, anyPurposeEkus, emptyAppPolicies)

	gt.NewRelationship(s.AuthStore1, s.Domain1, ad.NTAuthStoreFor)
	gt.NewRelationship(s.RootCA1, s.Domain1, ad.RootCAFor)
	gt.NewRelationship(s.EnterpriseCA1, s.AuthStore1, ad.TrustedForNTAuth)
	gt.NewRelationship(s.EnterpriseCA1, s.RootCA1, ad.EnterpriseCAFor)
	gt.NewRelationship(s.CertTemplate11, s.EnterpriseCA1, ad.PublishedTo)
	gt.NewRelationship(s.CertTemplate12, s.EnterpriseCA1, ad.PublishedTo)
	gt.NewRelationship(s.CertTemplate13, s.EnterpriseCA1, ad.PublishedTo)
}

type ADCSGoldenCertHarness struct {
	NTAuthStore1  *graph.Node
	RootCA1       *graph.Node
	EnterpriseCA1 *graph.Node
	Computer1     *graph.Node
	Domain1       *graph.Node

	Domain2        *graph.Node
	RootCA2        *graph.Node
	NTAuthStore2   *graph.Node
	EnterpriseCA21 *graph.Node
	EnterpriseCA22 *graph.Node
	EnterpriseCA23 *graph.Node
	Computer21     *graph.Node
	Computer22     *graph.Node
	Computer23     *graph.Node

	NTAuthStore3  *graph.Node
	RootCA3       *graph.Node
	EnterpriseCA3 *graph.Node
	Computer3     *graph.Node
	Domain3       *graph.Node
}

func (s *ADCSGoldenCertHarness) Setup(graphTestContext *GraphTestContext) {
	//Positive test cases for GoldenCert edge
	sid := RandomDomainSID()
	s.Domain1 = graphTestContext.NewActiveDirectoryDomain("domain 1", sid, false, true)
	s.RootCA1 = graphTestContext.NewActiveDirectoryRootCA("rca 1", sid)
	s.NTAuthStore1 = graphTestContext.NewActiveDirectoryNTAuthStore("ntauthstore 1", sid)
	s.EnterpriseCA1 = graphTestContext.NewActiveDirectoryEnterpriseCA("eca 1", sid)
	s.Computer1 = graphTestContext.NewActiveDirectoryComputer("computer 1", sid)

	graphTestContext.NewRelationship(s.NTAuthStore1, s.Domain1, ad.NTAuthStoreFor)
	graphTestContext.NewRelationship(s.RootCA1, s.Domain1, ad.RootCAFor)
	graphTestContext.NewRelationship(s.EnterpriseCA1, s.NTAuthStore1, ad.TrustedForNTAuth)
	graphTestContext.NewRelationship(s.EnterpriseCA1, s.RootCA1, ad.EnterpriseCAFor)
	graphTestContext.NewRelationship(s.Computer1, s.EnterpriseCA1, ad.HostsCAService)

	sid = RandomDomainSID()
	s.Domain3 = graphTestContext.NewActiveDirectoryDomain("domain 3", sid, false, true)
	s.RootCA3 = graphTestContext.NewActiveDirectoryRootCA("rca 3", sid)
	s.NTAuthStore3 = graphTestContext.NewActiveDirectoryNTAuthStore("ntauthstore 3", sid)
	s.EnterpriseCA3 = graphTestContext.NewActiveDirectoryEnterpriseCA("eca 3", sid)
	s.Computer3 = graphTestContext.NewActiveDirectoryComputer("computer 3", sid)

	graphTestContext.NewRelationship(s.NTAuthStore3, s.Domain3, ad.NTAuthStoreFor)
	graphTestContext.NewRelationship(s.RootCA3, s.Domain3, ad.RootCAFor)
	graphTestContext.NewRelationship(s.EnterpriseCA3, s.NTAuthStore3, ad.TrustedForNTAuth)
	graphTestContext.NewRelationship(s.EnterpriseCA3, s.RootCA3, ad.IssuedSignedBy)
	graphTestContext.NewRelationship(s.Computer3, s.EnterpriseCA3, ad.HostsCAService)

	//Negative test cases for GoldenCert edge
	sid = RandomDomainSID()
	s.Domain2 = graphTestContext.NewActiveDirectoryDomain("domain 2", sid, false, true)
	s.RootCA2 = graphTestContext.NewActiveDirectoryRootCA("rca2", sid)
	s.NTAuthStore2 = graphTestContext.NewActiveDirectoryNTAuthStore("authstore2", sid)
	s.EnterpriseCA21 = graphTestContext.NewActiveDirectoryEnterpriseCA("eca 3", sid)
	s.EnterpriseCA22 = graphTestContext.NewActiveDirectoryEnterpriseCA("eca 4", sid)
	s.EnterpriseCA23 = graphTestContext.NewActiveDirectoryEnterpriseCA("eca 5", sid)
	s.Computer21 = graphTestContext.NewActiveDirectoryComputer("computer 3", sid)
	s.Computer22 = graphTestContext.NewActiveDirectoryComputer("computer 4", sid)
	s.Computer23 = graphTestContext.NewActiveDirectoryComputer("computer 5", sid)

	graphTestContext.NewRelationship(s.RootCA2, s.Domain2, ad.RootCAFor)
	graphTestContext.NewRelationship(s.NTAuthStore2, s.Domain2, ad.NTAuthStoreFor)
	graphTestContext.NewRelationship(s.EnterpriseCA23, s.NTAuthStore2, ad.TrustedForNTAuth)
	graphTestContext.NewRelationship(s.EnterpriseCA21, s.RootCA2, ad.EnterpriseCAFor)
	graphTestContext.NewRelationship(s.EnterpriseCA22, s.RootCA2, ad.IssuedSignedBy)
	graphTestContext.NewRelationship(s.Computer21, s.EnterpriseCA21, ad.HostsCAService)
	graphTestContext.NewRelationship(s.Computer22, s.EnterpriseCA22, ad.HostsCAService)
	graphTestContext.NewRelationship(s.Computer23, s.EnterpriseCA23, ad.HostsCAService)

}

type WeakCertBindingAndUPNCertMappingHarness struct {
	EnterpriseCA1 *graph.Node
	EnterpriseCA2 *graph.Node
	Computer1     *graph.Node
	Computer2     *graph.Node
	Computer3     *graph.Node
	Computer4     *graph.Node
	Computer5     *graph.Node
	Domain1       *graph.Node
	Domain2       *graph.Node
	Domain3       *graph.Node
}

func (s *WeakCertBindingAndUPNCertMappingHarness) Setup(graphTestContext *GraphTestContext) {
	domainSid1 := "S-1-5-21-2697957641-2271029196-387917394"
	domainSid2 := "S-1-5-21-2697957641-2271029196-387917395"
	domainSid3 := "S-1-5-21-2697957641-2271029196-387917396"

	// Set up ECA nodes
	s.EnterpriseCA1 = graphTestContext.NewActiveDirectoryEnterpriseCAWithThumbprint("EnterpriseCA1", domainSid1, "a")
	s.EnterpriseCA2 = graphTestContext.NewActiveDirectoryEnterpriseCAWithThumbprint("EnterpriseCA2", domainSid3, "b")

	// Set up Domain nodes
	s.Domain1 = graphTestContext.NewActiveDirectoryDomain("Domain1", domainSid1, false, true)
	s.Domain2 = graphTestContext.NewActiveDirectoryDomain("Domain2", domainSid2, false, true)
	s.Domain3 = graphTestContext.NewActiveDirectoryDomain("Domain3", domainSid3, false, true)

	// Set up Computer nodes
	s.Computer1 = graphTestContext.NewActiveDirectoryComputer("Computer1", domainSid1)
	s.Computer1.Properties.Set(ad.CertificateMappingMethodsRaw.String(), []string{"4"})
	s.Computer1.Properties.Set(ad.StrongCertificateBindingEnforcementRaw.String(), []string{"1"})
	graphTestContext.UpdateNode(s.Computer1)

	s.Computer2 = graphTestContext.NewActiveDirectoryComputer("Computer2", domainSid2)
	s.Computer2.Properties.Set(ad.CertificateMappingMethodsRaw.String(), []string{"11"})
	s.Computer2.Properties.Set(ad.StrongCertificateBindingEnforcementRaw.String(), []string{"0"})
	graphTestContext.UpdateNode(s.Computer2)

	s.Computer3 = graphTestContext.NewActiveDirectoryComputer("Computer3", domainSid2)
	s.Computer3.Properties.Set(ad.CertificateMappingMethodsRaw.String(), []string{"31"})
	s.Computer3.Properties.Set(ad.StrongCertificateBindingEnforcementRaw.String(), []string{"2"})
	graphTestContext.UpdateNode(s.Computer3)

	s.Computer4 = graphTestContext.NewActiveDirectoryComputer("Computer4", domainSid2)
	s.Computer4.Properties.Set(ad.CertificateMappingMethodsRaw.String(), nil)
	s.Computer4.Properties.Set(ad.StrongCertificateBindingEnforcementRaw.String(), nil)
	graphTestContext.UpdateNode(s.Computer4)

	s.Computer5 = graphTestContext.NewActiveDirectoryComputer("Computer5", domainSid3)
	s.Computer5.Properties.Set(ad.CertificateMappingMethodsRaw.String(), []string{"15"})
	s.Computer5.Properties.Set(ad.StrongCertificateBindingEnforcementRaw.String(), []string{"2"})
	graphTestContext.UpdateNode(s.Computer5)

	// Set up edges from ECA nodes
	graphTestContext.NewRelationship(s.EnterpriseCA1, s.Computer1, ad.CanAbuseUPNCertMapping)
	graphTestContext.NewRelationship(s.EnterpriseCA1, s.Computer1, ad.CanAbuseWeakCertBinding)
	graphTestContext.NewRelationship(s.EnterpriseCA1, s.Computer2, ad.CanAbuseWeakCertBinding)
	graphTestContext.NewRelationship(s.EnterpriseCA1, s.Computer3, ad.CanAbuseUPNCertMapping)

	graphTestContext.NewRelationship(s.EnterpriseCA2, s.Computer5, ad.CanAbuseUPNCertMapping)

	// Set up edges from Computer nodes
	graphTestContext.NewRelationship(s.Computer1, s.Domain1, ad.DCFor)
	graphTestContext.NewRelationship(s.Computer2, s.Domain2, ad.DCFor)
	graphTestContext.NewRelationship(s.Computer3, s.Domain2, ad.DCFor)
	graphTestContext.NewRelationship(s.Computer4, s.Domain2, ad.DCFor)
	graphTestContext.NewRelationship(s.Computer5, s.Domain3, ad.DCFor)

	// Set up edges from Domain nodes
	graphTestContext.NewRelationship(s.Domain1, s.Domain2, ad.TrustedBy, graph.AsProperties(graph.PropertyMap{ad.TrustType: "ParentChild"}))
	graphTestContext.NewRelationship(s.Domain2, s.Domain3, ad.TrustedBy, graph.AsProperties(graph.PropertyMap{ad.TrustType: "External"}))
}

type IssuedSignedByHarness struct {
	RootCA1       *graph.Node
	RootCA2       *graph.Node
	EnterpriseCA1 *graph.Node
	EnterpriseCA2 *graph.Node
	EnterpriseCA3 *graph.Node
}

func (s *IssuedSignedByHarness) Setup(graphTestContext *GraphTestContext) {
	sid := RandomDomainSID()
	s.RootCA1 = graphTestContext.NewActiveDirectoryRootCAWithThumbprint("rca1", sid, "a")
	s.RootCA2 = graphTestContext.NewActiveDirectoryRootCAWithThumbprint("rca2", sid, "b")
	s.EnterpriseCA1 = graphTestContext.NewActiveDirectoryEnterpriseCAWithThumbprint("eca1", sid, "c")
	s.EnterpriseCA2 = graphTestContext.NewActiveDirectoryEnterpriseCAWithThumbprint("eca2", sid, "d")
	s.EnterpriseCA3 = graphTestContext.NewActiveDirectoryEnterpriseCAWithThumbprint("eca2", sid, "e")

	s.RootCA1.Properties.Set(ad.CertChain.String(), []string{"a"})
	s.RootCA2.Properties.Set(ad.CertChain.String(), []string{"b", "a"})
	s.EnterpriseCA1.Properties.Set(ad.CertChain.String(), []string{"c", "b", "a"})
	s.EnterpriseCA2.Properties.Set(ad.CertChain.String(), []string{"d", "c", "b", "a"})
	s.EnterpriseCA3.Properties.Set(ad.CertChain.String(), []string{"e"})

	graphTestContext.UpdateNode(s.RootCA1)
	graphTestContext.UpdateNode(s.RootCA2)
	graphTestContext.UpdateNode(s.EnterpriseCA1)
	graphTestContext.UpdateNode(s.EnterpriseCA2)
	graphTestContext.UpdateNode(s.EnterpriseCA3)
}

type TrustedForNTAuthHarness struct {
	EnterpriseCA1 *graph.Node
	EnterpriseCA2 *graph.Node
	EnterpriseCA3 *graph.Node

	NTAuthStore *graph.Node

	Domain *graph.Node
}

func (s *TrustedForNTAuthHarness) Setup(graphTestContext *GraphTestContext) {
	sid := RandomDomainSID()

	s.Domain = graphTestContext.NewActiveDirectoryDomain("domain", sid, false, true)

	s.NTAuthStore = graphTestContext.NewActiveDirectoryNTAuthStore("ntauthstore", sid)

	s.EnterpriseCA1 = graphTestContext.NewActiveDirectoryEnterpriseCAWithThumbprint("eca 1", sid, "a")

	s.EnterpriseCA2 = graphTestContext.NewActiveDirectoryEnterpriseCAWithThumbprint("eca 2", sid, "b")

	s.EnterpriseCA3 = graphTestContext.NewActiveDirectoryEnterpriseCA("eca 3", sid)
}

type ShortcutHarness struct {
	Group1 *graph.Node
	Group2 *graph.Node
	Group3 *graph.Node
	Group4 *graph.Node
	User1  *graph.Node
}

func (s *ShortcutHarness) Setup(graphTestContext *GraphTestContext) {
	sid := RandomDomainSID()
	s.Group1 = graphTestContext.NewActiveDirectoryGroup("GROUP ONE", sid)
	s.Group2 = graphTestContext.NewActiveDirectoryGroup("GROUP TWO", sid)
	s.Group3 = graphTestContext.NewActiveDirectoryGroup("GROUP THREE", sid)
	s.Group4 = graphTestContext.NewActiveDirectoryGroup("GROUP FOUR", sid)
	s.User1 = graphTestContext.NewActiveDirectoryUser("USER ONE", sid)

	graphTestContext.NewRelationship(s.Group4, s.Group1, ad.MemberOf)
	graphTestContext.NewRelationship(s.Group3, s.Group2, ad.MemberOf)
	graphTestContext.NewRelationship(s.Group3, s.Group1, ad.MemberOf)
	graphTestContext.NewRelationship(s.User1, s.Group4, ad.MemberOf)
	graphTestContext.NewRelationship(s.User1, s.Group3, ad.MemberOf)
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
	ShortcutHarness                                 ShortcutHarness
	ADCSESC1Harness                                 ADCSESC1Harness
	EnrollOnBehalfOfHarnessOne                      EnrollOnBehalfOfHarnessOne
	EnrollOnBehalfOfHarnessTwo                      EnrollOnBehalfOfHarnessTwo
	ADCSGoldenCertHarness                           ADCSGoldenCertHarness
	IssuedSignedByHarness                           IssuedSignedByHarness
	WeakCertBindingAndUPNCertMappingHarness         WeakCertBindingAndUPNCertMappingHarness
	TrustedForNTAuthHarness                         TrustedForNTAuthHarness
	NumCollectedActiveDirectoryDomains              int
	AZInboundControlHarness                         AZInboundControlHarness
}
