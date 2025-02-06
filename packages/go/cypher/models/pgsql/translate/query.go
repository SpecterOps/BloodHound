package translate

import (
	"fmt"
	"github.com/specterops/bloodhound/cypher/models/cypher"
	"github.com/specterops/bloodhound/cypher/models/pgsql"
)

func (s *Translator) buildMultiPartSinglePartQuery(singlePartQuery *cypher.SinglePartQuery, cteChain []pgsql.CommonTableExpression) error {
	// Prepend the CTE chain to the model's
	currentPart := s.query.CurrentPart()
	currentPart.Model.CommonTableExpressions.Expressions = append(cteChain, currentPart.Model.CommonTableExpressions.Expressions...)

	return nil
}

func (s *Translator) buildSinglePartQuery(singlePartQuery *cypher.SinglePartQuery) error {
	if s.query.CurrentPart().HasMutations() {
		if err := s.translateUpdates(s.query.Scope); err != nil {
			s.SetError(err)
		}

		if err := s.buildUpdates(s.query.Scope); err != nil {
			s.SetError(err)
		}
	}

	if s.query.CurrentPart().HasDeletions() {
		if err := s.buildDeletions(s.query.Scope); err != nil {
			s.SetError(err)
		}
	}

	// If there was no return specified end the CTE chain with a bare select
	if singlePartQuery.Return == nil {
		if literalReturn, err := pgsql.AsLiteral(1); err != nil {
			s.SetError(err)
		} else {
			s.query.CurrentPart().Model.Body = pgsql.Select{
				Projection: []pgsql.SelectItem{literalReturn},
			}
		}
	} else if err := s.buildTailProjection(); err != nil {
		s.SetError(err)
	}

	return nil
}

func (s *Translator) buildMultiPartQuery(singlePartQuery *cypher.SinglePartQuery) error {
	var multipartCTEChain []pgsql.CommonTableExpression

	// In order to author the multipart query part we first have to wrap it in a
	for _, part := range s.query.Parts[:len(s.query.Parts)-1] {
		// If the part has an empty inner CTE, make sure to remove it otherwise the keyword will still render
		if len(part.Model.CommonTableExpressions.Expressions) == 0 {
			part.Model.CommonTableExpressions = nil
		}

		// Autor the part as a nested CTE
		nextCTE := pgsql.CommonTableExpression{
			Query: *part.Model,
		}

		if part.frame != nil {
			nextCTE.Alias = pgsql.TableAlias{
				Name: part.frame.Binding.Identifier,
			}
		}

		if inlineSelect, err := s.buildInlineProjection(part); err != nil {
			return err
		} else {
			nextCTE.Query.Body = inlineSelect
		}

		multipartCTEChain = append(multipartCTEChain, nextCTE)
	}

	if err := s.buildMultiPartSinglePartQuery(singlePartQuery, multipartCTEChain); err != nil {
		return err
	}

	s.translation.Statement = *s.query.CurrentPart().Model
	return nil
}

func (s *Translator) translateWith() error {
	currentPart := s.query.CurrentPart()

	if !currentPart.HasProjections() {
		currentPart.frame.Exported.Clear()
	} else {
		var (
			projectedItems  = pgsql.NewIdentifierSet()
			aggregatedItems = pgsql.NewIdentifierSet()
		)

		// If an aggregation function is being used this invokes an implicit group by of non-function projections
		for _, projectionItem := range currentPart.projections.Items {
			switch typedSelectItem := projectionItem.SelectItem.(type) {
			case pgsql.FunctionCall:
				if pgsql.IsAggregateFunction(typedSelectItem.Function) {
					aggregatedItems.Add(typedSelectItem.Function)
				}
			}
		}

		for idx, projectionItem := range currentPart.projections.Items {
			switch typedSelectItem := projectionItem.SelectItem.(type) {
			case *pgsql.BinaryExpression:
				return fmt.Errorf("unhandled case for with statement")

			case pgsql.CompoundIdentifier:
				return fmt.Errorf("unhandled case for with statement")

			case pgsql.Identifier:
				if !aggregatedItems.IsEmpty() && !aggregatedItems.Contains(typedSelectItem) {
					currentPart.projections.GroupBy = append(currentPart.projections.GroupBy, typedSelectItem)
				}

				if binding, isBound := s.query.Scope.Lookup(typedSelectItem); !isBound {
					return fmt.Errorf("unable to lookup identifer %s for with statement", typedSelectItem)
				} else {
					// Track this projected item for scope pruning
					projectedItems.Add(binding.Identifier)

					// Create a new projection that maps the identifier
					currentPart.projections.Items[idx] = &Projection{
						SelectItem: pgsql.CompoundIdentifier{
							binding.LastProjection.Binding.Identifier, typedSelectItem,
						},
						Alias: pgsql.AsOptionalIdentifier(binding.Identifier),
					}

					// Assign the frame to the binding's last projection backref
					binding.LastProjection = currentPart.frame

					// Reveal and export the identifier in the current multipart query part's frame
					currentPart.frame.Reveal(binding.Identifier)
					currentPart.frame.Export(binding.Identifier)
				}

			default:
				// If this is not an identifier then check if the alias is specified. If the alias is specified, this
				// is a pure export (left-hand side is some other expression) and a new bound identifier is being
				// introduced.
				if projectionItem.Alias.Set {
					if binding, isBound := s.query.Scope.AliasedLookup(projectionItem.Alias.Value); !isBound {
						return fmt.Errorf("unable to lookup alias %s for with statement", projectionItem.Alias.Value)
					} else {
						// Track this projected item for scope pruning
						projectedItems.Add(binding.Identifier)

						// Assign the frame to the binding's last projection backref
						binding.LastProjection = currentPart.frame

						// Reveal and export the identifier in the current multipart query part's frame
						currentPart.frame.Reveal(binding.Identifier)
						currentPart.frame.Export(binding.Identifier)

						// Rewrite this projection's alias to use the internal binding
						projectionItem.Alias.Value = binding.Identifier
					}
				}
			}

			// Prune scope to only what's being exported by the with statement
			currentPart.frame.Visible = projectedItems.Copy()
			currentPart.frame.Exported = projectedItems.Copy()
		}
	}

	return nil
}

func (s *Translator) translateMultiPartQueryPart(scope *Scope, part *cypher.MultiPartQueryPart) error {
	queryPart := s.query.CurrentPart()

	// Unwind nested frames
	return scope.UnwindToFrame(queryPart.frame)
}
