package wellknown

import "fmt"

type NodeNamePrefix interface {
	fmt.Stringer
	AppendSuffix(suffix string) string
}

type nodeNamePrefix string

var _ NodeNamePrefix = nodeNamePrefix("")

func (n nodeNamePrefix) String() string {
	return string(n)
}

func (n nodeNamePrefix) AppendSuffix(suffix string) string {
	return fmt.Sprintf("%s@%s", n.String(), suffix)
}

func NewNodeNamePrefix(prefix string) NodeNamePrefix {
	return nodeNamePrefix(prefix)
}

var (
	DomainUsersNodeNamePrefix        NodeNamePrefix = NewNodeNamePrefix("DOMAIN USERS")
	AuthenticatedUsersNodeNamePrefix                = NewNodeNamePrefix("AUTHENTICATED USERS")
	EveryoneNodeNamePrefix                          = NewNodeNamePrefix("EVERYONE")
	DomainComputerNodeNamePrefix                    = NewNodeNamePrefix("DOMAIN COMPUTERS")
)

func DefineNodeName(prefix NodeNamePrefix, suffix string) string {
	return prefix.AppendSuffix(suffix)
}
