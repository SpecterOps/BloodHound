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
	"github.com/specterops/bloodhound/packages/go/analysis"
	"github.com/specterops/bloodhound/packages/go/graphschema/ad"
	"github.com/specterops/dawgs/graph"
)

const (
	TrustDirectionDisabled          = "Disabled"
	TrustDirectionOutbound          = "Outbound"
	TrustDirectionInbound           = "Inbound"
	TrustDirectionBidirectional     = "Bidirectional"
	IgnoredName                     = "IGNOREME"
	UserRightRemoteInteractiveLogon = "SeRemoteInteractiveLogonRight"
	UserRightBackup 				= "SeBackupPrivilege"
	UserRightRestore 				= "SeRestorePrivilege"
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
	PrincipalSID    string
	PrincipalType   string
	RightName       string
	IsInherited     bool
	InheritanceHash string
}

type SPNTarget struct {
	ComputerSID string
	Port        int
	Service     string
}

func (s ACE) Kind() graph.Kind {
	return parseADKind(s.PrincipalType)
}

func (s ACE) GetCachedValue() WriteOwnerLimitedPrincipal {
	return WriteOwnerLimitedPrincipal{
		SourceData: IngestibleEndpoint{
			Value: s.PrincipalSID,
			Kind:  s.Kind(),
		},
		IsInherited:     s.IsInherited,
		InheritanceHash: s.InheritanceHash,
	}
}

type IngestBase struct {
	ObjectIdentifier string
	Properties       map[string]any
	Aces             []ACE
	IsDeleted        bool
	IsACLProtected   bool
	ContainedBy      TypedPrincipal
}

type Oid struct {
	Name  string
	Value string
}

type CertificateExtension struct {
	Oid      Oid
	Critical bool
}

type Certificate struct {
	Thumbprint                string
	Name                      string
	Chain                     []string
	HasBasicConstraints       bool
	BasicConstraintPathLength int
	EnhancedKeyUsageOids      []Oid
	CertificateExtensions     []CertificateExtension
}

type EnrollmentAgentRestriction struct {
	AccessType   string
	Agent        TypedPrincipal
	AllTemplates bool
	Targets      []TypedPrincipal
	Template     TypedPrincipal
}

type EnrollmentAgentRestrictions struct {
	APIResult
	Restrictions []EnrollmentAgentRestriction
}

type CASecurity struct {
	APIResult
	Data []ACE
}

type IsUserSpecifiesSanEnabled struct {
	APIResult
	Value bool
}

type RoleSeparationEnabled struct {
	APIResult
	Value bool
}

type CARegistryData struct {
	CASecurity                  CASecurity
	EnrollmentAgentRestrictions EnrollmentAgentRestrictions
	IsUserSpecifiesSanEnabled   IsUserSpecifiesSanEnabled
	RoleSeparationEnabled       RoleSeparationEnabled
}

type DCRegistryData struct {
	CertificateMappingMethods            CertificateMappingMethods
	StrongCertificateBindingEnforcement  StrongCertificateBindingEnforcement
	VulnerableNetlogonSecurityDescriptor VulnerableNetlogonSecurityDescriptor
}

type CertificateMappingMethods struct {
	APIResult
	Value int
}

type StrongCertificateBindingEnforcement struct {
	APIResult
	Value int
}

type VulnerableNetlogonSecurityDescriptor struct {
	APIResult
	Value string
}

type GPO IngestBase

type AIACA IngestBase

type IssuancePolicy struct {
	IngestBase
	GroupLink TypedPrincipal
}

type RootCA struct {
	IngestBase
	DomainSID string
}

type EnterpriseCA struct {
	IngestBase
	CARegistryData          CARegistryData
	EnabledCertTemplates    []TypedPrincipal
	HostingComputer         string
	DomainSID               string
	HttpEnrollmentEndpoints []CAEnrollmentAPIResult
}

type CAEnrollmentAPIResult struct {
	APIResult
	Result CAEnrollmentEndpoint
}

type CAEnrollmentEndpoint struct {
	Url                    string
	ADCSWebEnrollmentHTTP  bool
	ADCSWebEnrollmentHTTPS bool
	ADCSWebEnrollmentEPA   bool
}

type NTAuthStore struct {
	IngestBase
	DomainSID string
}

type CertTemplate IngestBase

type Session struct {
	ComputerSID string
	UserSID     string
	LogonType   int
}

type Group struct {
	IngestBase
	Members       []TypedPrincipal
	HasSIDHistory []TypedPrincipal
}

type User struct {
	IngestBase
	AllowedToDelegate       []TypedPrincipal
	SPNTargets              []SPNTarget
	PrimaryGroupSID         string
	HasSIDHistory           []TypedPrincipal
	DomainSID               string
	UnconstrainedDelegation bool
}

type Container struct {
	IngestBase
	ChildObjects      []TypedPrincipal
	InheritanceHashes []string
}

type Trust struct {
	TargetDomainSid      string
	IsTransitive         bool
	TrustDirection       string
	TrustType            string
	SidFilteringEnabled  bool
	TargetDomainName     string
	TGTDelegationEnabled bool
	TrustAttributes      any
}

type GPLink struct {
	Guid       string
	IsEnforced bool
}

type Domain struct {
	IngestBase
	ChildObjects      []TypedPrincipal
	Trusts            []Trust
	Links             []GPLink
	GPOChanges        GPOChanges
	InheritanceHashes []string
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

type BoolAPIResult struct {
	APIResult
	Result bool
}

type SMBSigningAPIResult struct {
	APIResult
	Result SMBSigningResult
}

type SMBSigningResult struct {
	SigningEnabled bool
}

type NTLMRegistryDataAPIResult struct {
	APIResult
	Result NTLMRegistryInfo
}

type NTLMRegistryInfo struct {
	RestrictSendingNtlmTraffic   *uint
	RequireSecuritySignature     *uint
	EnableSecuritySignature      *uint
	RestrictReceivingNTLMTraffic *uint
	NtlmMinServerSec             *uint
	NtlmMinClientSec             *uint
	LmCompatibilityLevel         *uint
	UseMachineId                 *uint
	ClientAllowedNTLMServers     *[]string
}

type Computer struct {
	IngestBase
	PrimaryGroupSID         string
	AllowedToDelegate       []TypedPrincipal
	AllowedToAct            []TypedPrincipal
	DumpSMSAPassword        []TypedPrincipal
	Sessions                SessionAPIResult
	PrivilegedSessions      SessionAPIResult
	RegistrySessions        SessionAPIResult
	LocalGroups             []LocalGroupAPIResult
	UserRights              []UserRightsAssignmentAPIResult
	DCRegistryData          DCRegistryData
	Status                  ComputerStatus
	HasSIDHistory           []TypedPrincipal
	IsDC                    bool
	DomainSID               string
	UnconstrainedDelegation bool
	SmbInfo                 SMBSigningAPIResult
	IsWebClientRunning      BoolAPIResult
	NTLMRegistryData        NTLMRegistryDataAPIResult
}

type GPOChanges struct {
	LocalAdmins        []TypedPrincipal
	RemoteDesktopUsers []TypedPrincipal
	DcomUsers          []TypedPrincipal
	PSRemoteUsers      []TypedPrincipal
	AffectedComputers  []TypedPrincipal
}

type OU struct {
	IngestBase
	ChildObjects      []TypedPrincipal
	Links             []GPLink
	GPOChanges        GPOChanges
	InheritanceHashes []string
}

type GenericMetadata struct {
	SourceKind string `json:"source_kind"`
}

type GenericNode struct {
	ID         string
	Kinds      []string
	Properties map[string]any
}

type GenericEdge struct {
	Start      EdgeEndpoint
	End        EdgeEndpoint
	Kind       string
	Properties map[string]any
}

type EdgeEndpoint struct {
	Value   string
	Kind    string
	MatchBy string `json:"match_by"`
}
