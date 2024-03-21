package pgsql

const (
	WildcardIdentifier Identifier = "*"
)

func AsWildcardIdentifier(identifier Identifier) CompoundIdentifier {
	return CompoundIdentifier{identifier, WildcardIdentifier}
}
