package nodeprops

import (
	"fmt"
	"strings"

	"github.com/specterops/bloodhound/dawgs/graph"
	"github.com/specterops/bloodhound/graphschema/ad"
	"github.com/specterops/bloodhound/graphschema/common"
)

func ReadDomainIDandNameAsString(nodeToRead *graph.Node) (string, string, error) {
	var domainSIDStr, domainNameStr string
	var err error

	errMessageFunc := func(errMsg string) error {
		return fmt.Errorf("failed to read domain SID and name: %s", errMsg)
	}

	if nodeToRead == nil {
		return "", "", errMessageFunc("given nodeToRead is nil")
	}

	if domainSID := nodeToRead.Properties.Get(ad.DomainSID.String()); domainSID.IsNil() {
		return "", "", errMessageFunc("read domain SID property value is nil")
	} else {
		if domainSIDStr, err = domainSID.String(); err != nil {
			return "", "", errMessageFunc(err.Error())
		} else {
			if len(strings.TrimSpace(domainSIDStr)) == 0 {
				return "", "", errMessageFunc("read domain SID is empty or blank")
			}
		}
	}

	if domainName := nodeToRead.Properties.Get(common.Name.String()); domainName.IsNil() {
		return "", "", errMessageFunc("read domain name property value is nil")
	} else {
		if domainNameStr, err = domainName.String(); err != nil {
			return "", "", errMessageFunc(err.Error())
		} else {
			if len(strings.TrimSpace(domainNameStr)) == 0 {
				return "", "", errMessageFunc("read domain name is empty or blank")
			}
		}
	}

	return domainSIDStr, domainNameStr, nil
}
