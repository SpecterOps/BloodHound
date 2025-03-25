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

	returnFunc := func(errMsg string) (string, string, error) {
		return domainSIDStr, domainNameStr, fmt.Errorf("failed to read domain SID and name: %s", errMsg)
	}

	if nodeToRead == nil {
		return returnFunc("given nodeToRead is nil")
	}

	if domainSID := nodeToRead.Properties.Get(ad.DomainSID.String()); domainSID.IsNil() {
		return returnFunc("read domain SID property value is nil")
	} else {
		if domainSIDStr, err = domainSID.String(); err != nil {
			return returnFunc(err.Error())
		} else {
			if len(strings.TrimSpace(domainSIDStr)) == 0 {
				return returnFunc("read domain SID is empty or blank")
			}
		}
	}

	if domainName := nodeToRead.Properties.Get(common.Name.String()); domainName.IsNil() {
		return returnFunc("read domain name property value is nil")
	} else {
		if domainNameStr, err = domainName.String(); err != nil {
			return returnFunc(err.Error())
		} else {
			if len(strings.TrimSpace(domainNameStr)) == 0 {
				return returnFunc("read domain name is empty or blank")
			}
		}
	}

	return domainSIDStr, domainNameStr, nil
}
