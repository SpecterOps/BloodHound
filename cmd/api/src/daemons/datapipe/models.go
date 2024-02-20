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

package datapipe

import (
	"encoding/json"

	"github.com/bloodhoundad/azurehound/v2/enums"
	"github.com/specterops/bloodhound/dawgs/graph"
	"github.com/specterops/bloodhound/ein"
	"github.com/specterops/bloodhound/graphschema/ad"
)

type DataWrapper struct {
	Metadata Metadata        `json:"meta"`
	Payload  json.RawMessage `json:"data"`
}

type Metadata struct {
	Type    DataType         `json:"type"`
	Methods CollectionMethod `json:"methods"`
	Version int              `json:"version"`
}

func (s Metadata) MatchKind() (graph.Kind, bool) {
	switch s.Type {
	case DataTypeComputer:
		return ad.Computer, true

	case DataTypeUser:
		return ad.User, true

	case DataTypeGroup:
		return ad.Group, true

	case DataTypeGPO:
		return ad.GPO, true

	case DataTypeDomain:
		return ad.Domain, true

	case DataTypeOU:
		return ad.OU, true

	case DataTypeContainer:
		return ad.Container, true

	case DataTypeAIACA:
		return ad.AIACA, true

	case DataTypeRootCA:
		return ad.RootCA, true

	case DataTypeEnterpriseCA:
		return ad.EnterpriseCA, true

	case DataTypeNTAuthStore:
		return ad.NTAuthStore, true

	case DataTypeCertTemplate:
		return ad.CertTemplate, true
	}

	return nil, false
}

type DataType string

const (
	DataTypeSession      DataType = "sessions"
	DataTypeUser         DataType = "users"
	DataTypeGroup        DataType = "groups"
	DataTypeComputer     DataType = "computers"
	DataTypeGPO          DataType = "gpos"
	DataTypeOU           DataType = "ous"
	DataTypeDomain       DataType = "domains"
	DataTypeRemoved      DataType = "deleted"
	DataTypeContainer    DataType = "containers"
	DataTypeLocalGroups  DataType = "localgroups"
	DataTypeAIACA        DataType = "aiacas"
	DataTypeRootCA       DataType = "rootcas"
	DataTypeEnterpriseCA DataType = "enterprisecas"
	DataTypeNTAuthStore  DataType = "ntauthstores"
	DataTypeCertTemplate DataType = "certtemplates"
	DataTypeAzure        DataType = "azure"
)

func AllIngestDataTypes() []DataType {
	return []DataType{
		DataTypeSession,
		DataTypeUser,
		DataTypeGroup,
		DataTypeComputer,
		DataTypeGPO,
		DataTypeOU,
		DataTypeDomain,
		DataTypeRemoved,
		DataTypeContainer,
		DataTypeLocalGroups,
		DataTypeAIACA,
		DataTypeRootCA,
		DataTypeEnterpriseCA,
		DataTypeNTAuthStore,
		DataTypeCertTemplate,
		DataTypeAzure,
	}
}

func (s DataType) IsValid() bool {
	for _, method := range AllIngestDataTypes() {
		if s == method {
			return true
		}
	}

	return false
}

type CollectionMethod uint64

const (
	CollectionMethodGroup         CollectionMethod = 1
	CollectionMethodLocalAdmin    CollectionMethod = 1 << 1
	CollectionMethodGPOLocalGroup CollectionMethod = 1 << 2
	CollectionMethodSession       CollectionMethod = 1 << 3
	CollectionMethodLoggedOn      CollectionMethod = 1 << 4
	CollectionMethodTrusts        CollectionMethod = 1 << 5
	CollectionMethodACL           CollectionMethod = 1 << 6
	CollectionMethodContainer     CollectionMethod = 1 << 7
	CollectionMethodRDP           CollectionMethod = 1 << 8
	CollectionMethodObjectProps   CollectionMethod = 1 << 9
	CollectionMethodSessionLoop   CollectionMethod = 1 << 10
	CollectionMethodLoggedOnLoop  CollectionMethod = 1 << 11
	CollectionMethodDCOM          CollectionMethod = 1 << 12
	CollectionMethodSPNTargets    CollectionMethod = 1 << 13
	CollectionMethodPSRemote      CollectionMethod = 1 << 14
	CollectionMethodUserRights    CollectionMethod = 1 << 15
	CollectionMethodCARegistry    CollectionMethod = 1 << 16
	CollectionMethodDCRegistry    CollectionMethod = 1 << 17
	CollectionMethodCertServices  CollectionMethod = 1 << 18
)

func AllCollectionMethods() []CollectionMethod {
	return []CollectionMethod{
		CollectionMethodGroup,
		CollectionMethodLocalAdmin,
		CollectionMethodGPOLocalGroup,
		CollectionMethodSession,
		CollectionMethodLoggedOn,
		CollectionMethodTrusts,
		CollectionMethodACL,
		CollectionMethodContainer,
		CollectionMethodRDP,
		CollectionMethodObjectProps,
		CollectionMethodSessionLoop,
		CollectionMethodLoggedOnLoop,
		CollectionMethodDCOM,
		CollectionMethodSPNTargets,
		CollectionMethodPSRemote,
		CollectionMethodUserRights,
		CollectionMethodCARegistry,
		CollectionMethodDCRegistry,
		CollectionMethodCertServices,
	}
}

func (s CollectionMethod) IsValid() bool {
	for _, method := range AllCollectionMethods() {
		if s.Has(method) {
			return true
		}
	}

	return false
}

func (s CollectionMethod) Has(flag CollectionMethod) bool {
	return s.And(flag) != 0
}

func (s CollectionMethod) And(flag CollectionMethod) CollectionMethod {
	return s & flag
}

func (s CollectionMethod) Or(flag CollectionMethod) CollectionMethod {
	return s | flag
}

type ConvertedData struct {
	NodeProps []ein.IngestibleNode
	RelProps  []ein.IngestibleRelationship
}

func (s *ConvertedData) Clear() {
	s.NodeProps = s.NodeProps[:0]
	s.RelProps = s.RelProps[:0]
}

type ConvertedGroupData struct {
	NodeProps              []ein.IngestibleNode
	RelProps               []ein.IngestibleRelationship
	DistinguishedNameProps []ein.IngestibleRelationship
}

func (s *ConvertedGroupData) Clear() {
	s.NodeProps = s.NodeProps[:0]
	s.RelProps = s.RelProps[:0]
	s.DistinguishedNameProps = s.DistinguishedNameProps[:0]
}

type ConvertedSessionData struct {
	SessionProps []ein.IngestibleSession
}

func (s *ConvertedSessionData) Clear() {
	s.SessionProps = s.SessionProps[:0]
}

type AzureBase struct {
	Kind enums.Kind      `json:"kind"`
	Data json.RawMessage `json:"data"`
}

type ConvertedAzureData struct {
	NodeProps   []ein.IngestibleNode
	RelProps    []ein.IngestibleRelationship
	OnPremNodes []ein.IngestibleNode
}

func (s *ConvertedAzureData) Clear() {
	s.NodeProps = s.NodeProps[:0]
	s.RelProps = s.RelProps[:0]
	s.OnPremNodes = s.OnPremNodes[:0]
}
