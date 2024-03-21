package pgsql

type Operator string

func (s Operator) AsExpression() Expression {
	return s
}

func (s Operator) String() string {
	return string(s)
}

func (s Operator) NodeType() string {
	return "operator"
}

const (
	OperatorEquals         Operator = "="
	OperatorPGArrayOverlap Operator = "operator (pg_catalog.&&)"
	OperatorAnd            Operator = "and"
	OperatorOr             Operator = "or"
	OperatorNot            Operator = "not"
	OperatorJSONField      Operator = "->"
	OperatorJSONTextField  Operator = "->>"
)
