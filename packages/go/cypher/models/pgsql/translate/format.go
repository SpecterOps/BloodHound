package translate

import (
	"bytes"
	"github.com/specterops/bloodhound/cypher/models/cypher"
	cypherFormat "github.com/specterops/bloodhound/cypher/models/cypher/format"
	"github.com/specterops/bloodhound/cypher/models/pgsql"
	"github.com/specterops/bloodhound/cypher/models/pgsql/format"
)

func Translated(translation Result) (string, error) {
	return format.Statement(translation.Statement, format.NewOutputBuilder())
}

func FromCypher(regularQuery *cypher.RegularQuery, kindMapper pgsql.KindMapper, stripLiterals bool) (format.Formatted, error) {
	var (
		output  = &bytes.Buffer{}
		emitter = cypherFormat.NewCypherEmitter(stripLiterals)
	)

	output.WriteString("-- ")

	if err := emitter.Write(regularQuery, output); err != nil {
		return format.Formatted{}, err
	}

	output.WriteString("\n")

	if translation, err := Translate(regularQuery, kindMapper); err != nil {
		return format.Formatted{}, err
	} else if sqlQuery, err := format.Statement(translation.Statement, format.NewOutputBuilder()); err != nil {
		return format.Formatted{}, err
	} else {
		output.WriteString(sqlQuery)

		return format.Formatted{
			Statement:  output.String(),
			Parameters: translation.Parameters,
		}, nil
	}
}
