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

package ingest

import (
	"encoding/json"
	"errors"

	"github.com/specterops/bloodhound/packages/go/graphschema/ad"
	"github.com/specterops/bloodhound/packages/go/mediatypes"
	"github.com/specterops/dawgs/graph"
)

var AllowedZipFileUploadTypes = []string{
	mediatypes.ApplicationZip.String(),
	"application/x-zip-compressed", // Not currently available in mediatypes
	"application/zip-compressed",   // Not currently available in mediatypes
}

var AllowedFileUploadTypes = append([]string{mediatypes.ApplicationJson.String()}, AllowedZipFileUploadTypes...)

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
	case DataTypeIssuancePolicy:
		return ad.IssuancePolicy, true
	}

	return nil, false
}

type DataType string

const (
	DataTypeSession        DataType = "sessions"
	DataTypeUser           DataType = "users"
	DataTypeGroup          DataType = "groups"
	DataTypeComputer       DataType = "computers"
	DataTypeGPO            DataType = "gpos"
	DataTypeOU             DataType = "ous"
	DataTypeDomain         DataType = "domains"
	DataTypeRemoved        DataType = "deleted"
	DataTypeContainer      DataType = "containers"
	DataTypeLocalGroups    DataType = "localgroups"
	DataTypeAIACA          DataType = "aiacas"
	DataTypeRootCA         DataType = "rootcas"
	DataTypeEnterpriseCA   DataType = "enterprisecas"
	DataTypeNTAuthStore    DataType = "ntauthstores"
	DataTypeCertTemplate   DataType = "certtemplates"
	DataTypeAzure          DataType = "azure"
	DataTypeIssuancePolicy DataType = "issuancepolicies"
	DataTypeOpenGraph      DataType = "opengraph"
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
		DataTypeIssuancePolicy,
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
	CollectionMethodGroup CollectionMethod = 1 << iota
	CollectionMethodLocalAdmin
	CollectionMethodGPOLocalGroup
	CollectionMethodSession
	CollectionMethodLoggedOn
	CollectionMethodTrusts
	CollectionMethodACL
	CollectionMethodContainer
	CollectionMethodRDP
	CollectionMethodObjectProps
	CollectionMethodSessionLoop
	CollectionMethodLoggedOnLoop
	CollectionMethodDCOM
	CollectionMethodSPNTargets
	CollectionMethodPSRemote
	CollectionMethodUserRights
	CollectionMethodCARegistry
	CollectionMethodDCRegistry
	CollectionMethodCertServices
	CollectionMethodBackup
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
		CollectionMethodBackup,
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

const (
	DelimOpenBracket        = json.Delim('{')
	DelimCloseBracket       = json.Delim('}')
	DelimOpenSquareBracket  = json.Delim('[')
	DelimCloseSquareBracket = json.Delim(']')
)

var (
	ErrMetaTagNotFound     = errors.New("no valid meta tag found")
	ErrDataTagNotFound     = errors.New("no data tag found")
	ErrNoTagFound          = errors.New("no valid meta tag or data tag found")
	ErrInvalidDataTag      = errors.New("invalid data tag found")
	ErrJSONDecoderInternal = errors.New("json decoder internal error")
	ErrInvalidZipFile      = errors.New("failed to find zip file header")
	ErrMixedIngestFormat   = errors.New("request must use either the classic format (meta/data) or the generic format (graph), not both")

	ErrOpenGraphMetaTagValidation = errors.New("metadata tag is invalid")
)
