package model

import (
	"fmt"
	"github.com/specterops/bloodhound/analysis"
	"github.com/specterops/bloodhound/dawgs/graph"
	"github.com/specterops/bloodhound/graphschema/ad"
	"github.com/specterops/bloodhound/graphschema/common"
	"time"
)

type ComboNodeElement struct {
	AdminCount        bool      `json:"admin_count"` // WHAT IS THIS
	Category          string    `json:"category"`
	Description       string    `json:"description"`
	DistinguishedName string    `json:"distinguished_name"`
	Domain            string    `json:"domain"`
	DomainSID         string    `json:"domain_sid"`
	HighValue         bool      `json:"high_value"`
	LastSeen          time.Time `json:"last_seen"`
	Level             int       `json:"level"`
	Name              string    `json:"name"`
	Neo4jImportID     string    `json:"neo4j_import_id"`
	Kind              string    `json:"kind"`
	ObjectID          string    `json:"object_id"`
	SystemTags        string    `json:"system_tags"`
}

func (s *ComboNodeElement) FromNode(node *graph.Node) error {
	var err error
	s.Neo4jImportID, s.Kind = node.ID.String(), analysis.GetNodeKindDisplayLabel(node)

	if s.Description, err = node.Properties.Get(common.Description.String()).String(); err != nil {
		return fmt.Errorf("error type-negotiating Description: %v", err)
	} else if s.DistinguishedName, err = node.Properties.Get(ad.DistinguishedName.String()).String(); err != nil {
		return fmt.Errorf("error type-negotiating DistinguishedName: %v", err)
	} else if s.Domain, err = node.Properties.Get(ad.Domain.String()).String(); err != nil {
		return fmt.Errorf("error type-negotiating Domain: %v", err)
	} else if s.DomainSID, err = node.Properties.Get(ad.DomainSID.String()).String(); err != nil {
		return fmt.Errorf("error type-negotiating DomainSID: %v", err)
	} else if s.HighValue, err = node.Properties.Get(ad.HighValue.String()).Bool(); err != nil {
		return fmt.Errorf("error type-negotiating HighValue: %v", err)
	} else if s.LastSeen, err = node.Properties.Get(common.LastSeen.String()).Time(); err != nil {
		return fmt.Errorf("error type-negotiating LastSeen: %v", err)
	} else if s.ObjectID, err = node.Properties.Get(common.ObjectID.String()).String(); err != nil {
		return fmt.Errorf("error type-negotiating ObjectID: %v", err)
	} else if s.SystemTags, err = node.Properties.Get(common.SystemTags.String()).String(); err != nil {
		return fmt.Errorf("error type-negotiating SystemTags: %v", err)
	} else if s.Category, err = node.Properties.Get("category").String(); err != nil {
		return fmt.Errorf("error type-negotiating Category: %v", err)
	} else if s.Level, err = node.Properties.Get("level").Int(); err != nil {
		return fmt.Errorf("error type-negotiating Level: %v", err)
	} else {
		return nil
	}
}

type ComboNode []ComboNodeElement
