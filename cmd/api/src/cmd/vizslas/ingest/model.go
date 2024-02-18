package ingest

import (
	"encoding/json"
	"errors"
	"fmt"
)

type Extension struct {
	Name string          `json:"name"`
	Data json.RawMessage `json:"data"`
}

type Metadata struct {
	Version    int         `json:"version"`
	Extensions []Extension `json:"extensions"`
}

type Entity struct {
	Kind       string         `json:"kind"`
	IDKeys     []string       `json:"id_keys"`
	Properties map[string]any `json:"properties"`
}

type Node struct {
	Entity
	ExtendedKinds []string `json:"extended_kinds"`
}

type Edge struct {
	Entity
	Start Node `json:"start"`
	End   Node `json:"end"`
}

type Entities struct {
	Nodes []Node `json:"nodes"`
	Edges []Edge `json:"edges"`
}

func (s *Entities) AddNode(node Node) {
	s.Nodes = append(s.Nodes, node)
}

func (s *Entities) AddEdge(edge Edge) {
	s.Edges = append(s.Edges, edge)
}

type Payload struct {
	Metadata Metadata `json:"metadata"`
	Deleted  Entities `json:"deleted"`
	Visited  Entities `json:"visited"`
}

var ErrEntityIDNotFound = errors.New("entity ID not found")

func UnwrapEntity[E Node | Edge | Entity](rawEntity E) (Entity, error) {
	var entity Entity

	// Find the entity details based on the type of the entity container that was passed in
	switch typedEntity := any(rawEntity).(type) {
	case Node:
		entity = typedEntity.Entity
	case Edge:
		entity = typedEntity.Entity
	case Entity:
		entity = typedEntity
	default:
		return entity, fmt.Errorf("unsupported entity type %T", rawEntity)
	}

	return entity, nil
}

func CheckEntity[E Node | Edge | Entity](rawEntity E) error {
	var entity Entity

	switch typedEntity := any(rawEntity).(type) {
	case Node:
		entity = typedEntity.Entity

	case Edge:
		entity = typedEntity.Entity

		// If this is an edge then also validate the start and end node entities
		if err := CheckEntity(typedEntity.Start); err != nil {
			return err
		}

		if err := CheckEntity(typedEntity.End); err != nil {
			return err
		}

	case Entity:
		entity = typedEntity

	default:
		return fmt.Errorf("unsupported entity type %T", rawEntity)
	}

	// Entities must have ID keys or they are not valid
	if len(entity.IDKeys) == 0 {
		return ErrEntityIDNotFound
	}

	// Validate existence of the ID properties
	for _, idKey := range entity.IDKeys {
		if _, hasProperty := entity.Properties[idKey]; !hasProperty {
			return fmt.Errorf("entity is missing property: %s", idKey)
		}
	}

	return nil
}
