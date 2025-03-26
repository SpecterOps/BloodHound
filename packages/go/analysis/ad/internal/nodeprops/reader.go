package nodeprops

import (
	"fmt"
	"strings"

	"github.com/specterops/bloodhound/dawgs/graph"
	"github.com/specterops/bloodhound/graphschema/ad"
	"github.com/specterops/bloodhound/graphschema/common"
)

// ReadDomainIDandNameAsString extracts the domain SID and domain name from the given node. This function is
// intentionally placed in the internal package to avoid import cycle issues that would arise if it were implemented as
// an unexported function within the ad package and subsequently referenced by ad_test.go as of commit '87afb00a'. Using
// the internal package pattern here preserves proper package boundaries while still allowing ad_test.go to access the
// functionality needed for data generation and testing.
func ReadDomainIDandNameAsString(nodeToRead *graph.Node) (string, string, error) {
	if nodeToRead == nil {
		return "", "", fmt.Errorf("given nodeToRead is nil")
	}

	domainSID := nodeToRead.Properties.Get(ad.DomainSID.String())
	if domainSID.IsNil() {
		return "", "", fmt.Errorf("read domain SID property value is nil")
	}

	domainSIDStr, err := domainSID.String()
	if err != nil {
		return "", "", fmt.Errorf("failed to convert domainSID to string: %s", err)
	}

	if len(strings.TrimSpace(domainSIDStr)) == 0 {
		return "", "", fmt.Errorf("read domain SID is empty or blank")
	}

	domainName := nodeToRead.Properties.Get(common.Name.String())
	if domainName.IsNil() {
		return "", "", fmt.Errorf("read domain name property value is nil")
	}

	domainNameStr, err := domainName.String()
	if err != nil {
		return "", "", fmt.Errorf("failed to convert domain name to string: %s", err)
	}

	if len(strings.TrimSpace(domainNameStr)) == 0 {
		return "", "", fmt.Errorf("read domain name is empty or blank")
	}

	return domainSIDStr, domainNameStr, nil
}
