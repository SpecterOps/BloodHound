package pgsql

import "errors"

var (
	ErrNonArrayDataType = errors.New("data type is not an array type")
)

const (
	TableNode Identifier = "node"
	TableEdge Identifier = "edge"

	ColumnID         Identifier = "id"
	ColumnProperties Identifier = "properties"
	ColumnKindIDs    Identifier = "kind_ids"
	ColumnKindID     Identifier = "kind_id"
	ColumnGraphID    Identifier = "graph_id"
	ColumnStartID    Identifier = "start_id"
	ColumnEndID      Identifier = "end_id"
)

type DataType string

func (s DataType) NodeType() string {
	return "sql_data_type"
}

const (
	UnsetDataType            DataType = ""
	UnknownDataType          DataType = "UNKNOWN"
	Reference                DataType = "REFERENCE"
	Null                     DataType = "NULL"
	NodeComposite            DataType = "nodecomposite"
	NodeCompositeArray       DataType = "nodecomposite[]"
	EdgeComposite            DataType = "edgecomposite"
	EdgeCompositeArray       DataType = "edgecomposite[]"
	PathComposite            DataType = "pathcomposite"
	Int2                     DataType = "int2"
	Int2Array                DataType = "int2[]"
	Int4                     DataType = "int4"
	Int4Array                DataType = "int4[]"
	Int8                     DataType = "int8"
	Int8Array                DataType = "int8[]"
	Float4                   DataType = "float4"
	Float4Array              DataType = "float4[]"
	Float8                   DataType = "float8"
	Float8Array              DataType = "float8[]"
	Boolean                  DataType = "bool"
	Text                     DataType = "text"
	TextArray                DataType = "text[]"
	JSONB                    DataType = "jsonb"
	Date                     DataType = "date"
	TimeWithTimeZone         DataType = "time with time zone"
	TimeWithoutTimeZone      DataType = "time without time zone"
	Interval                 DataType = "interval"
	TimestampWithTimeZone    DataType = "timestamp with time zone"
	TimestampWithoutTimeZone DataType = "timestamp without time zone"
)

func (s DataType) IsArrayType() bool {
	switch s {
	case Int2Array, Int4Array, Int8Array, Float4Array, Float8Array, TextArray:
		return true
	}

	return false
}

func (s DataType) ArrayBaseType() (DataType, error) {
	switch s {
	case Int2Array:
		return Int2, nil
	case Int4Array:
		return Int4, nil
	case Int8Array:
		return Int8, nil
	case Float4Array:
		return Float4, nil
	case Float8Array:
		return Float8, nil
	case TextArray:
		return Text, nil
	default:
		return UnknownDataType, ErrNonArrayDataType
	}
}

func (s DataType) String() string {
	return string(s)
}

var CompositeTypes = []DataType{NodeComposite, NodeCompositeArray, EdgeComposite, EdgeCompositeArray, PathComposite}
