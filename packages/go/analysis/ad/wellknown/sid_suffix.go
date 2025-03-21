package wellknown

import (
	"fmt"
)

type SIDSuffix interface {
	fmt.Stringer
	PrependPrefix(prefix string) string
}

type sidSuffix string

var _ SIDSuffix = sidSuffix("")

func (s sidSuffix) String() string {
	return string(s)
}

func (s sidSuffix) PrependPrefix(prefix string) string {
	return fmt.Sprintf("%s%s", prefix, s.String())
}

func NewSIDSuffix(suffix string) SIDSuffix {
	return sidSuffix(suffix)
}

var (
	DomainUsersSIDSuffix        SIDSuffix = NewSIDSuffix("-513")
	AuthenticatedUsersSIDSuffix           = NewSIDSuffix("-S-1-5-11")
	EveryoneSIDSuffix                     = NewSIDSuffix("-S-1-1-0")
	DomainComputersSIDSuffix              = NewSIDSuffix("-515")
)

func DefineSID(sidPrefix string, sidSuffix SIDSuffix) string {
	return sidSuffix.PrependPrefix(sidPrefix)
}
