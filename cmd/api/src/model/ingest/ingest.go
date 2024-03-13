package ingest

import (
	"encoding/json"
	"github.com/specterops/bloodhound/dawgs/graph"
	"github.com/specterops/bloodhound/errors"
	"github.com/specterops/bloodhound/graphschema/ad"
)

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
)
