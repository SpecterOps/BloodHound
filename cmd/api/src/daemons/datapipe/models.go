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
	}

	return nil, false
}

type DataType string

const (
	DataTypeSession     DataType = "sessions"
	DataTypeUser        DataType = "users"
	DataTypeGroup       DataType = "groups"
	DataTypeComputer    DataType = "computers"
	DataTypeGPO         DataType = "gpos"
	DataTypeOU          DataType = "ous"
	DataTypeDomain      DataType = "domains"
	DataTypeRemoved     DataType = "deleted"
	DataTypeContainer   DataType = "containers"
	DataTypeLocalGroups DataType = "localgroups"
	DataTypeAzure       DataType = "azure"
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
		DataTypeAzure,
	}
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

type ConvertedGroupData struct {
	NodeProps              []ein.IngestibleNode
	RelProps               []ein.IngestibleRelationship
	DistinguishedNameProps []ein.IngestibleRelationship
}

type ConvertedSessionData struct {
	SessionProps []ein.IngestibleSession
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

func CreateConvertedAzureData(count int) ConvertedAzureData {
	converted := ConvertedAzureData{}
	converted.NodeProps = make([]ein.IngestibleNode, count)
	converted.RelProps = make([]ein.IngestibleRelationship, 0)

	return converted
}
