package translate

import (
	"fmt"
	"github.com/specterops/bloodhound/cypher/model/pgsql"
	"strconv"
)

type IdentifierGenerator struct {
	patternBindingID int
	nodeBindingID    int
	edgeBindingID    int
}

func (s *IdentifierGenerator) PatternBinding() pgsql.Identifier {
	bindingID := s.patternBindingID
	s.patternBindingID += 1

	return pgsql.Identifier("p" + strconv.Itoa(bindingID))
}

func (s *IdentifierGenerator) NodeBinding() pgsql.Identifier {
	bindingID := s.nodeBindingID
	s.nodeBindingID += 1

	return pgsql.Identifier("n" + strconv.Itoa(bindingID))
}

func (s *IdentifierGenerator) EdgeBinding() pgsql.Identifier {
	bindingID := s.edgeBindingID
	s.edgeBindingID += 1

	return pgsql.Identifier("e" + strconv.Itoa(bindingID))
}

func NewIdentifierGenerator() *IdentifierGenerator {
	return &IdentifierGenerator{
		patternBindingID: 1,
		nodeBindingID:    1,
		edgeBindingID:    1,
	}
}

type Constraint struct {
	Dependencies pgsql.IdentifierSet
	Expression   pgsql.Expression
}

type ConstraintTracker struct {
	Constraints []*Constraint
}

func NewConstraintTracker() *ConstraintTracker {
	return &ConstraintTracker{}
}

func (s *ConstraintTracker) ConsumeAll() pgsql.Expression {
	var topLevelExpression pgsql.Expression

	for _, constraint := range s.Constraints {
		topLevelExpression = OptionalAnd(topLevelExpression, constraint.Expression)
	}

	s.Constraints = nil
	return topLevelExpression
}

func (s *ConstraintTracker) Consume(available pgsql.IdentifierSet) (pgsql.IdentifierSet, pgsql.Expression) {
	var (
		matched    = pgsql.IdentifierSet{}
		constraint pgsql.Expression
	)

	for idx := 0; idx < len(s.Constraints); {
		nextConstraint := s.Constraints[idx]

		if available.Satisfies(nextConstraint.Dependencies) {
			// Remove this constraint
			s.Constraints = append(s.Constraints[:idx], s.Constraints[idx+1:]...)

			// Append the constraint as a conjoined expression
			constraint = OptionalAnd(constraint, nextConstraint.Expression)

			// Track which identifiers were satisfied
			matched.Merge(nextConstraint.Dependencies)
		} else {
			// If this constraint isn't satisfied by the available namespace move to the next constraint
			idx += 1
		}
	}

	return matched, constraint
}

func (s *ConstraintTracker) Constrain(dependencies pgsql.IdentifierSet, constraintExpression pgsql.Expression) {
	for _, constraint := range s.Constraints {
		if constraint.Dependencies.Matches(dependencies) {
			constraint.Expression = OptionalAnd(constraint.Expression, constraintExpression)
			return
		}
	}

	s.Constraints = append(s.Constraints, &Constraint{
		Dependencies: dependencies,
		Expression:   constraintExpression,
	})
}

type TrackedIdentifier struct {
	Identifier   pgsql.Identifier
	Alias        pgsql.Identifier
	Dependencies []*TrackedIdentifier
	DataType     pgsql.DataType
}

func (s *TrackedIdentifier) BuildFromClauses() ([]pgsql.FromClause, error) {
	var fromClauses []pgsql.FromClause
	switch s.DataType {
	case pgsql.NodeComposite, pgsql.EdgeComposite:
		fromClauses = append(fromClauses, pgsql.FromClause{
			Relation: pgsql.TableReference{
				Name: pgsql.CompoundIdentifier{s.Identifier},
			},
		})

	case pgsql.PathComposite:
		for _, dependency := range s.Dependencies {
			fromClauses = append(fromClauses, pgsql.FromClause{
				Relation: pgsql.TableReference{
					Name: pgsql.CompoundIdentifier{dependency.Identifier},
				},
			})
		}
	}

	return fromClauses, nil
}

func (s *TrackedIdentifier) BuildCompositeValue() (pgsql.CompositeValue, error) {
	switch s.DataType {
	case pgsql.NodeComposite:
		return pgsql.CompositeValue{
			Values: []pgsql.Expression{
				pgsql.CompoundIdentifier{s.Identifier, pgsql.ColumnID},
				pgsql.CompoundIdentifier{s.Identifier, pgsql.ColumnKindIDs},
				pgsql.CompoundIdentifier{s.Identifier, pgsql.ColumnProperties},
			},
			DataType: s.DataType,
		}, nil

	case pgsql.EdgeComposite:
		return pgsql.CompositeValue{
			Values: []pgsql.Expression{
				pgsql.CompoundIdentifier{s.Identifier, pgsql.ColumnID},
				pgsql.CompoundIdentifier{s.Identifier, pgsql.ColumnStartID},
				pgsql.CompoundIdentifier{s.Identifier, pgsql.ColumnEndID},
				pgsql.CompoundIdentifier{s.Identifier, pgsql.ColumnKindID},
				pgsql.CompoundIdentifier{s.Identifier, pgsql.ColumnProperties},
			},
			DataType: s.DataType,
		}, nil

	case pgsql.PathComposite:
		var (
			nodeReferences = pgsql.ArrayLiteral{}
			edgeReferences = pgsql.ArrayLiteral{}
		)

		for _, dependency := range s.Dependencies {
			if compositeValue, err := dependency.BuildCompositeValue(); err != nil {
				return compositeValue, err
			} else {
				switch dependency.DataType {
				case pgsql.NodeComposite:
					nodeReferences.Values = append(nodeReferences.Values, compositeValue)

				case pgsql.EdgeComposite:
					edgeReferences.Values = append(edgeReferences.Values, compositeValue)

				default:
					return pgsql.CompositeValue{}, fmt.Errorf("unsupported nested composite type for pathcomposite: %s", s.DataType)
				}
			}
		}

		return pgsql.CompositeValue{
			Values: []pgsql.Expression{
				nodeReferences,
				edgeReferences,
			},
			DataType: s.DataType,
		}, nil

	default:
		return pgsql.CompositeValue{}, fmt.Errorf("unsupported composite type: %s", s.DataType)
	}
}

func (s *TrackedIdentifier) BuildProjection() (pgsql.Projection, error) {
	if compositeValue, err := s.BuildCompositeValue(); err != nil {
		return nil, err
	} else {
		return s.AliasProjection(compositeValue), nil
	}
}

func (s *TrackedIdentifier) AliasExpression(expression pgsql.Expression) pgsql.Expression {
	if s.Alias != "" {
		return pgsql.AliasedExpression{
			Expression: expression,
			Alias:      s.Alias,
		}
	}

	return expression
}

func (s *TrackedIdentifier) AliasProjection(projection pgsql.Projection) pgsql.Projection {
	if s.Alias != "" {
		return pgsql.AliasedExpression{
			Expression: projection,
			Alias:      s.Alias,
		}
	}

	return projection
}

type IdentifierTracker struct {
	aliases            map[string]pgsql.Identifier
	trackedIdentifiers map[pgsql.Identifier]*TrackedIdentifier
}

func NewIdentifierTracker() *IdentifierTracker {
	return &IdentifierTracker{
		aliases:            map[string]pgsql.Identifier{},
		trackedIdentifiers: map[pgsql.Identifier]*TrackedIdentifier{},
	}
}

func (s *IdentifierTracker) DependsOn(identifier pgsql.Identifier, dependencies ...pgsql.Identifier) error {
	if trackedIdentifier, isTracked := s.trackedIdentifiers[identifier]; !isTracked {
		return fmt.Errorf("unknown identifier: %s", identifier)
	} else {
		for _, dependency := range dependencies {
			if trackedDependency, isTracked := s.trackedIdentifiers[dependency]; !isTracked {
				return fmt.Errorf("unknown dependent identifier: %s", dependency)
			} else {
				trackedIdentifier.Dependencies = append(trackedIdentifier.Dependencies, trackedDependency)
			}
		}
	}

	return nil
}

func (s *IdentifierTracker) Track(identifier pgsql.Identifier, dataType pgsql.DataType) *TrackedIdentifier {
	newTrackedIdentifier := &TrackedIdentifier{
		Identifier: identifier,
		DataType:   dataType,
	}

	s.aliases[identifier.String()] = identifier
	s.trackedIdentifiers[identifier] = newTrackedIdentifier

	return newTrackedIdentifier
}

func (s *IdentifierTracker) Alias(oldIdentifier string, identifier pgsql.Identifier, dataType pgsql.DataType) {
	s.aliases[oldIdentifier] = identifier

	newTrackedIdentifier := s.Track(identifier, dataType)
	newTrackedIdentifier.Alias = pgsql.Identifier(oldIdentifier)
}

func (s *IdentifierTracker) TrackString(identifier string, dataType pgsql.DataType) {
	s.Track(pgsql.Identifier(identifier), dataType)
}

func (s *IdentifierTracker) Lookup(identifier pgsql.Identifier) (*TrackedIdentifier, bool) {
	trackedIdentifier, found := s.trackedIdentifiers[identifier]
	return trackedIdentifier, found
}

func (s *IdentifierTracker) LookupAlias(oldIdentifier string) (pgsql.Identifier, bool) {
	alias, found := s.aliases[oldIdentifier]
	return alias, found
}
