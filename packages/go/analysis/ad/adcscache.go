package ad

import (
	"context"
	"github.com/specterops/bloodhound/dawgs/cardinality"
	"github.com/specterops/bloodhound/dawgs/graph"
	"github.com/specterops/bloodhound/graphschema/ad"
	"github.com/specterops/bloodhound/log"
)

type ADCSCache struct {
	AuthStoreForChainValid  map[graph.ID]cardinality.Duplex[uint32]
	RootCAForChainValid     map[graph.ID]cardinality.Duplex[uint32]
	CertTemplateControllers map[graph.ID][]*graph.Node
	EnterpriseCAEnrollers   map[graph.ID][]*graph.Node
	PublishedTemplateCache  map[graph.ID][]*graph.Node
}

func NewADCSCache() ADCSCache {
	return ADCSCache{
		AuthStoreForChainValid:  make(map[graph.ID]cardinality.Duplex[uint32]),
		RootCAForChainValid:     make(map[graph.ID]cardinality.Duplex[uint32]),
		CertTemplateControllers: make(map[graph.ID][]*graph.Node),
		EnterpriseCAEnrollers:   make(map[graph.ID][]*graph.Node),
		PublishedTemplateCache:  make(map[graph.ID][]*graph.Node),
	}
}

func (s ADCSCache) BuildCache(ctx context.Context, db graph.Database, enterpriseCAs, certTemplates []*graph.Node) {
	db.ReadTransaction(ctx, func(tx graph.Transaction) error {
		for _, ct := range certTemplates {
			if firstDegreePrincipals, err := fetchFirstDegreeNodes(tx, ct, ad.Enroll, ad.GenericAll, ad.AllExtendedRights); err != nil {
				log.Errorf("error fetching enrollers for cert template %d: %w", ct.ID, err)
			} else {
				s.CertTemplateControllers[ct.ID] = firstDegreePrincipals.Slice()
			}
		}

		for _, eca := range enterpriseCAs {
			if firstDegreeEnrollers, err := fetchFirstDegreeNodes(tx, eca, ad.Enroll); err != nil {
				log.Errorf("error fetching enrollers for enterprise ca %d: %w", eca.ID, err)
			} else {
				s.EnterpriseCAEnrollers[eca.ID] = firstDegreeEnrollers.Slice()
			}

			if publishedTemplates, err := FetchCertTemplatesPublishedToCA(tx, eca); err != nil {
				log.Errorf("error fetching published cert templates for enterprise ca %d: %w", eca.ID, err)
			} else {
				s.PublishedTemplateCache[eca.ID] = publishedTemplates.Slice()
			}
		}

		if domains, err := FetchCollectedDomains(tx); err != nil {
			log.Errorf("error fetching collected domains for esc cache: %w", err)
		} else {
			for _, domain := range domains {
				if rootCaForNodes, err := FetchEnterpriseCAsRootCAForPathToDomain(tx, domain); err != nil {
					log.Errorf("error getting cas via rootcafor for domain %d: %w", domain.ID, err)
				} else if authStoreForNodes, err := FetchEnterpriseCAsTrustedForNTAuthToDomain(tx, domain); err != nil {
					log.Errorf("error getting cas via authstorefor for domain %d: %w", domain.ID, err)
				} else {
					s.AuthStoreForChainValid[domain.ID] = nodeSetToBitmap(authStoreForNodes)
					s.RootCAForChainValid[domain.ID] = nodeSetToBitmap(rootCaForNodes)
				}
			}
		}

		return nil
	})

	log.Infof("Finished building adcs cache")
}

func nodeSetToBitmap(set graph.NodeSet) cardinality.Duplex[uint32] {
	bitmap := cardinality.NewBitmap32()
	for _, s := range set {
		bitmap.Add(s.ID.Uint32())
	}
	return bitmap
}

func (s ADCSCache) DoesCAChainProperlyToDomain(enterpriseCA, domain *graph.Node) bool {
	var domainID = domain.ID
	var caID = enterpriseCA.ID.Uint32()

	if _, ok := s.RootCAForChainValid[domainID]; !ok {
		return false
	} else if _, ok := s.AuthStoreForChainValid[domainID]; !ok {
		return false
	} else {
		return s.RootCAForChainValid[domainID].Contains(caID) && s.AuthStoreForChainValid[domainID].Contains(caID)
	}
}
