package csgen

import (
	"fmt"
	"strconv"
	"strings"
)

type OutputBuilder struct {
	MaterializeParameters bool
	StripLiterals         bool
	parameters            map[string]any
	builder               *strings.Builder
}

func NewOutputBuilder() *OutputBuilder {
	return &OutputBuilder{
		builder: &strings.Builder{},
	}
}

func (s *OutputBuilder) HasOutput() bool {
	return s.builder.Len() > 0
}

func (s *OutputBuilder) Write(values ...any) {
	for _, value := range values {
		switch typedValue := value.(type) {
		case string:
			s.builder.WriteString(typedValue)

		case fmt.Stringer:
			s.builder.WriteString(typedValue.String())

		default:
			panic(fmt.Sprintf("invalid write parameter type: %T", value))
		}
	}
}

func (s *OutputBuilder) Build() string {
	return s.builder.String()
}

func formatValue(builder *OutputBuilder, value any) error {
	switch typedValue := value.(type) {
	case uint:
		builder.Write(strconv.FormatUint(uint64(typedValue), 10))

	case uint8:
		builder.Write(strconv.FormatUint(uint64(typedValue), 10))

	case uint16:
		builder.Write(strconv.FormatUint(uint64(typedValue), 10))

	case uint32:
		builder.Write(strconv.FormatUint(uint64(typedValue), 10))

	case uint64:
		builder.Write(strconv.FormatUint(typedValue, 10))

	case int:
		builder.Write(strconv.FormatInt(int64(typedValue), 10))

	case int8:
		builder.Write(strconv.FormatInt(int64(typedValue), 10))

	case int16:
		builder.Write(strconv.FormatInt(int64(typedValue), 10))

	case int32:
		builder.Write(strconv.FormatInt(int64(typedValue), 10))

	case int64:
		builder.Write(strconv.FormatInt(typedValue, 10))

	case string:
		builder.Write("\"", typedValue, "\"")

	case bool:
		builder.Write(strconv.FormatBool(typedValue))

	case float32:
		builder.Write(strconv.FormatFloat(float64(typedValue), 'f', -1, 64))

	case float64:
		builder.Write(strconv.FormatFloat(typedValue, 'f', -1, 64))

	default:
		return fmt.Errorf("unsupported literal type: %T", value)
	}

	return nil
}

func formatLiteral(builder *OutputBuilder, literal Literal) error {
	if literal.Null {
		builder.Write("null")
		return nil
	}

	return formatValue(builder, literal.Value)
}
