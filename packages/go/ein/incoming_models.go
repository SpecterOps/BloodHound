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

package ein

import (
	"github.com/specterops/bloodhound/analysis"
	"github.com/specterops/bloodhound/dawgs/graph"
	"github.com/specterops/bloodhound/graphschema/ad"
)

const (
	TrustDirectionDisabled          = "Disabled"
	TrustDirectionOutbound          = "Outbound"
	TrustDirectionInbound           = "Inbound"
	TrustDirectionBidirectional     = "Bidirectional"
	IgnoredName                     = "IGNOREME"
	UserRightRemoteInteractiveLogon = "SeRemoteInteractiveLogonRight"
)

func parseADKind(rawKindStr string) graph.Kind {
	if kind, err := analysis.ParseKind(rawKindStr); err != nil {
		// TODO: Figure out a logging strategy for this since the context is wrapped in a very tight loop. It is
		//       possible that careless logging here may flood application logs and potentially DOS centralized
		//       logging infrastructure.
		return ad.Entity
	} else {
		return kind
	}
}

type TypedPrincipal struct {
	ObjectIdentifier string
	ObjectType       string
}

func (s TypedPrincipal) Kind() graph.Kind {
	return parseADKind(s.ObjectType)
}

type ACE struct {
	PrincipalSID  string
	PrincipalType string
	RightName     string
	IsInherited   bool
}

type SPNTarget struct {
	ComputerSID string
	Port        int
	Service     string
}

func (s ACE) Kind() graph.Kind {
	return parseADKind(s.PrincipalType)
}

type IngestBase struct {
	ObjectIdentifier string
	Properties       map[string]any
	Aces             []ACE
	IsDeleted        bool
	IsACLProtected   bool
	ContainedBy      TypedPrincipal
}

type GPO IngestBase

type Session struct {
	ComputerSID string
	UserSID     string
	LogonType   int
}

type Group struct {
	IngestBase
	Members []TypedPrincipal
}

type User struct {
	IngestBase
	AllowedToDelegate []TypedPrincipal
	SPNTargets        []SPNTarget
	PrimaryGroupSID   string
	HasSIDHistory     []TypedPrincipal
}

type Container struct {
	IngestBase
	ChildObjects []TypedPrincipal
}

type Trust struct {
	TargetDomainSid     string
	IsTransitive        bool
	TrustDirection      string
	TrustType           string
	SidFilteringEnabled bool
	TargetDomainName    string
}

type GPLink struct {
	Guid       string
	IsEnforced bool
}

type Domain struct {
	IngestBase
	ChildObjects []TypedPrincipal
	Trusts       []Trust
	Links        []GPLink
}

type SessionAPIResult struct {
	APIResult
	Results []Session
}

type ComputerStatus struct {
	Connectable bool
	Error       string
}

type APIResult struct {
	Collected     bool
	FailureReason string
}

type NamedPrincipal struct {
	ObjectIdentifier string
	PrincipalName    string
}

type LocalGroupAPIResult struct {
	APIResult
	Results          []TypedPrincipal
	LocalNames       []NamedPrincipal
	Name             string
	ObjectIdentifier string
}

type UserRightsAssignmentAPIResult struct {
	APIResult
	Results    []TypedPrincipal
	LocalNames []NamedPrincipal
	Privilege  string
}

type Computer struct {
	IngestBase
	PrimaryGroupSID   string
	AllowedToDelegate []TypedPrincipal
	AllowedToAct      []TypedPrincipal
	DumpSMSAPassword  []TypedPrincipal
	Sessions          SessionAPIResult
	LocalGroups       []LocalGroupAPIResult
	UserRights        []UserRightsAssignmentAPIResult
	Status            ComputerStatus
	HasSIDHistory     []TypedPrincipal
}

type OU struct {
	IngestBase
	ChildObjects []TypedPrincipal
	Links        []GPLink
}
