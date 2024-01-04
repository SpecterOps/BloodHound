// Copyright 2023 Specter Ops, Inc.
//
// Licensed under the Apache License, Version 2.0
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//
// SPDX-License-Identifier: Apache-2.0

package integration

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/specterops/bloodhound/dawgs/cardinality"
	_ "github.com/specterops/bloodhound/dawgs/drivers/neo4j"
	"github.com/specterops/bloodhound/dawgs/graph"
	"github.com/specterops/bloodhound/dawgs/ops"
	"github.com/specterops/bloodhound/graphschema/ad"
	"github.com/specterops/bloodhound/graphschema/azure"
	"github.com/specterops/bloodhound/graphschema/common"
	"github.com/specterops/bloodhound/log"
	"github.com/specterops/bloodhound/src/test"
	"github.com/specterops/bloodhound/src/test/must"
	"github.com/stretchr/testify/require"
)

var DefaultRelProperties = graph.AsProperties(graph.PropertyMap{
	common.LastSeen: time.Now().Format(time.RFC3339),
})

func NewGraphTestContext(testCtrl test.Controller) *GraphTestContext {
	testCtx := &GraphTestContext{
		testCtrl:     testCtrl,
		nodesCreated: cardinality.NewBitmap32(),
		GraphDB:      OpenNeo4jGraphDB(testCtrl),
	}

	testCtrl.Cleanup(testCtx.Cleanup)
	return testCtx
}

type GraphTestContext struct {
	testCtrl     test.Controller
	tx           graph.Transaction
	nodesCreated cardinality.Duplex[uint32]
	Harness      HarnessDetails
	GraphDB      graph.Database
}

func (s *GraphTestContext) NodeObjectID(node *graph.Node) string {
	objectID, err := node.Properties.Get(common.ObjectID.String()).String()
	require.Nilf(s.testCtrl, err, "Expected node %d to have a valid %s property: %v", node.ID, common.ObjectID.String(), err)

	return objectID
}

func (s *GraphTestContext) FindNode(criteria graph.Criteria) *graph.Node {
	var node *graph.Node

	require.Nil(s.testCtrl, s.GraphDB.ReadTransaction(context.Background(), func(tx graph.Transaction) error {
		fetchedNode, err := tx.Nodes().Filter(criteria).First()
		node = fetchedNode
		return err
	}))

	return node
}

func (s *GraphTestContext) UpdateNode(node *graph.Node) {
	require.Nil(s.testCtrl, s.tx.UpdateNode(node))
}

func (s *GraphTestContext) Cleanup() {
	if err := s.GraphDB.BatchOperation(context.Background(), func(batch graph.Batch) error {
		return s.nodesCreated.Each(func(nodeID uint32) (bool, error) {
			if err := batch.DeleteNode(graph.ID(nodeID)); err != nil {
				return false, err
			}

			return true, nil
		})
	}); err != nil {
		s.testCtrl.Errorf("Failed to clear DB after tests: %v", err)
	}
}

func (s *GraphTestContext) EmptyDatabaseTest(dbDelegate func(harness HarnessDetails, db graph.Database) error) {
	log.ConfigureDefaults()

	require.Nil(s.testCtrl, dbDelegate(s.Harness, s.GraphDB))
}

func (s *GraphTestContext) DatabaseTest(dbDelegate func(harness HarnessDetails, db graph.Database) error) {
	log.ConfigureDefaults()

	require.Nil(s.testCtrl, s.GraphDB.WriteTransaction(context.Background(), func(tx graph.Transaction) error {
		s.tx = tx

		defer func() {
			s.tx = nil
		}()

		if err := tx.Nodes().Delete(); err != nil {
			return err
		}

		s.setupActiveDirectory()
		s.setupAzure()
		return nil
	}))

	require.Nil(s.testCtrl, dbDelegate(s.Harness, s.GraphDB))
}

func (s *GraphTestContext) DatabaseTestWithSetup(setup func(harness *HarnessDetails), dbDelegate func(harness HarnessDetails, db graph.Database) error) {
	log.Configure(&log.Configuration{
		Level: log.LevelDebug,
	})

	require.Nil(s.testCtrl, s.GraphDB.WriteTransaction(context.Background(), func(tx graph.Transaction) error {
		s.tx = tx

		defer func() {
			s.tx = nil
		}()

		if err := tx.Nodes().Delete(); err != nil {
			return err
		}

		setup(&s.Harness)
		return nil
	}))

	require.Nil(s.testCtrl, dbDelegate(s.Harness, s.GraphDB))
}

func (s *GraphTestContext) BatchTest(batchDelegate func(harness HarnessDetails, batch graph.Batch), assertionDelegate func(details HarnessDetails, tx graph.Transaction)) {
	log.ConfigureDefaults()

	require.Nil(s.testCtrl, s.GraphDB.WriteTransaction(context.Background(), func(tx graph.Transaction) error {
		s.tx = tx

		defer func() {
			s.tx = nil
		}()

		if err := tx.Nodes().Delete(); err != nil {
			return err
		}

		s.setupActiveDirectory()
		s.setupAzure()
		return nil
	}))

	require.Nil(s.testCtrl, s.GraphDB.BatchOperation(context.Background(), func(batch graph.Batch) error {
		batchDelegate(s.Harness, batch)
		return nil
	}))

	require.Nil(s.testCtrl, s.GraphDB.WriteTransaction(context.Background(), func(tx graph.Transaction) error {
		assertionDelegate(s.Harness, tx)
		return nil
	}))
}

func (s *GraphTestContext) TransactionalTest(txDelegate func(harness HarnessDetails, tx graph.Transaction)) {
	log.ConfigureDefaults()

	require.Nil(s.testCtrl, s.GraphDB.WriteTransaction(context.Background(), func(tx graph.Transaction) error {
		s.tx = tx

		defer func() {
			s.tx = nil
		}()

		if err := tx.Nodes().Delete(); err != nil {
			return err
		}

		s.setupActiveDirectory()
		s.setupAzure()
		return nil
	}))

	require.Nil(s.testCtrl, s.GraphDB.WriteTransaction(context.Background(), func(tx graph.Transaction) error {
		s.tx = tx

		defer func() {
			s.tx = nil
		}()

		txDelegate(s.Harness, tx)
		return nil
	}))
}

func (s *GraphTestContext) ReadTransactionTest(setup func(harness *HarnessDetails), txDelegate func(harness HarnessDetails, tx graph.Transaction)) {
	log.ConfigureDefaults()

	require.Nil(s.testCtrl, s.GraphDB.WriteTransaction(context.Background(), func(tx graph.Transaction) error {
		s.tx = tx

		defer func() {
			s.tx = nil
		}()

		if err := tx.Nodes().Delete(); err != nil {
			return err
		}

		setup(&s.Harness)
		return nil
	}))

	require.Nil(s.testCtrl, s.GraphDB.ReadTransaction(context.Background(), func(tx graph.Transaction) error {
		s.tx = tx

		defer func() {
			s.tx = nil
		}()

		txDelegate(s.Harness, tx)
		return nil
	}))
}

func (s *GraphTestContext) WriteTransactionTest(setup func(harness *HarnessDetails), txDelegate func(harness HarnessDetails, tx graph.Transaction)) {
	log.ConfigureDefaults()

	require.Nil(s.testCtrl, s.GraphDB.WriteTransaction(context.Background(), func(tx graph.Transaction) error {
		s.tx = tx

		defer func() {
			s.tx = nil
		}()

		if err := tx.Nodes().Delete(); err != nil {
			return err
		}

		setup(&s.Harness)
		return nil
	}))

	require.Nil(s.testCtrl, s.GraphDB.WriteTransaction(context.Background(), func(tx graph.Transaction) error {
		s.tx = tx

		defer func() {
			s.tx = nil
		}()

		txDelegate(s.Harness, tx)
		return nil
	}))
}

func (s *GraphTestContext) DeleteNode(tx graph.Transaction, target *graph.Node) {
	err := ops.DeleteNodes(tx, target.ID)
	require.Nilf(s.testCtrl, err, "Error deleting node: %v", err)

	s.nodesCreated.Remove(target.ID.Uint32())
}

func (s *GraphTestContext) NewNode(properties *graph.Properties, kinds ...graph.Kind) *graph.Node {
	newNode, err := s.tx.CreateNode(properties, kinds...)
	require.Nilf(s.testCtrl, err, "Error creating node: %v", err)

	s.nodesCreated.Add(newNode.ID.Uint32())
	return newNode
}

func (s *GraphTestContext) NewAzureApplication(name, objectID, tenantID string) *graph.Node {
	return s.NewNode(graph.AsProperties(graph.PropertyMap{
		common.Name:     name,
		common.ObjectID: objectID,
		azure.TenantID:  tenantID,
	}), azure.Entity, azure.App)
}

func (s *GraphTestContext) NewAzureServicePrincipal(name, objectID, tenantID string) *graph.Node {
	return s.NewNode(graph.AsProperties(graph.PropertyMap{
		common.Name:     name,
		common.ObjectID: objectID,
		azure.TenantID:  tenantID,
	}), azure.Entity, azure.ServicePrincipal)
}

func (s *GraphTestContext) NewAzureUser(name, principalName, description, objectID, licenses, tenantID string, mfaEnabled bool) *graph.Node {
	return s.NewNode(graph.AsProperties(graph.PropertyMap{
		common.Name:             name,
		azure.UserPrincipalName: principalName,
		common.Description:      description,
		common.ObjectID:         objectID,
		azure.Licenses:          licenses,
		azure.MFAEnabled:        mfaEnabled,
		azure.TenantID:          tenantID,
	}), azure.Entity, azure.User)
}

func (s *GraphTestContext) NewAzureGroup(name, objectID, tenantID string) *graph.Node {
	return s.NewNode(graph.AsProperties(graph.PropertyMap{
		common.Name:              name,
		common.ObjectID:          objectID,
		azure.TenantID:           tenantID,
		azure.IsAssignableToRole: true,
	}), azure.Entity, azure.Group)
}

func (s *GraphTestContext) NewAzureVM(name, objectID, tenantID string) *graph.Node {
	return s.NewNode(graph.AsProperties(graph.PropertyMap{
		common.Name:     name,
		common.ObjectID: objectID,
		azure.TenantID:  tenantID,
	}), azure.Entity, azure.VM)
}

func (s *GraphTestContext) NewAzureRole(name, objectID, roleTemplateID, tenantID string) *graph.Node {
	return s.NewNode(graph.AsProperties(graph.PropertyMap{
		common.Name:          name,
		common.ObjectID:      objectID,
		azure.RoleTemplateID: roleTemplateID,
		azure.TenantID:       tenantID,
	}), azure.Entity, azure.Role)
}

func (s *GraphTestContext) NewAzureDevice(name, objectID, deviceID, tenantID string) *graph.Node {
	return s.NewNode(graph.AsProperties(graph.PropertyMap{
		common.Name:     name,
		common.ObjectID: objectID,
		azure.DeviceID:  deviceID,
		azure.TenantID:  tenantID,
	}), azure.Entity, azure.Device)
}

func (s *GraphTestContext) NewAzureResourceGroup(name, objectID, tenantID string) *graph.Node {
	return s.NewNode(graph.AsProperties(graph.PropertyMap{
		common.Name:     name,
		common.ObjectID: objectID,
		azure.TenantID:  tenantID,
	}), azure.Entity, azure.ResourceGroup)
}

func (s *GraphTestContext) NewAzureManagementGroup(name, objectID, tenantID string) *graph.Node {
	return s.NewNode(graph.AsProperties(graph.PropertyMap{
		common.Name:     name,
		common.ObjectID: objectID,
		azure.TenantID:  tenantID,
	}), azure.Entity, azure.ManagementGroup)
}

func (s *GraphTestContext) NewAzureKeyVault(name, objectID, tenantID string) *graph.Node {
	return s.NewNode(graph.AsProperties(graph.PropertyMap{
		common.Name:     name,
		common.ObjectID: objectID,
		azure.TenantID:  tenantID,
	}), azure.Entity, azure.KeyVault)
}

func (s *GraphTestContext) NewAzureSubscription(name, objectID, tenantID string) *graph.Node {
	return s.NewNode(graph.AsProperties(graph.PropertyMap{
		common.Name:     name,
		common.ObjectID: objectID,
		azure.TenantID:  tenantID,
	}), azure.Entity, azure.Subscription)
}

func (s *GraphTestContext) NewRelationship(startNode, endNode *graph.Node, kind graph.Kind, properties ...*graph.Properties) *graph.Relationship {
	var nodeProperties *graph.Properties

	if len(properties) > 0 {
		nodeProperties = properties[0]

		if len(properties) > 1 {
			for _, additionalProperties := range properties[1:] {
				nodeProperties.SetAll(additionalProperties.Map)
			}
		}
	} else {
		nodeProperties = graph.NewProperties()
	}

	newRelationship, err := s.tx.CreateRelationship(startNode, endNode, kind, nodeProperties)

	require.Nil(s.testCtrl, err, fmt.Sprintf("error: %v", err))
	return newRelationship
}

func (s *GraphTestContext) CreateAzureRelatedRoles(root *graph.Node, tenantID string, numRoles int) graph.NodeSet {
	roles := graph.NewNodeSet()

	for roleIdx := 0; roleIdx < numRoles; roleIdx++ {
		var (
			objectID       = must.NewUUIDv4().String()
			roleTemplateID = must.NewUUIDv4().String()
			newRole        = s.NewAzureRole(fmt.Sprintf("AZRole_%s", objectID), objectID, roleTemplateID, tenantID)
		)

		s.NewRelationship(root, newRole, azure.HasRole)
		roles.Add(newRole)
	}

	return roles
}

func (s *GraphTestContext) NewAzureTenant(tenantID string) *graph.Node {
	return s.NewNode(graph.AsProperties(graph.PropertyMap{
		common.Name:     "New Tenant",
		common.ObjectID: tenantID,
		azure.License:   "license",
	}), azure.Entity, azure.Tenant)
}

func (s *GraphTestContext) NewActiveDirectoryDomain(name, objectID string, blocksInheritance, collected bool) *graph.Node {
	if collected {
		s.Harness.NumCollectedActiveDirectoryDomains++
	}

	return s.NewNode(graph.AsProperties(graph.PropertyMap{
		common.Name:          name,
		common.ObjectID:      objectID,
		ad.DomainSID:         objectID,
		common.Collected:     collected,
		ad.BlocksInheritance: blocksInheritance,
	}), ad.Entity, ad.Domain)
}

func (s *GraphTestContext) NewActiveDirectoryComputer(name, domainSID string) *graph.Node {
	return s.NewNode(graph.AsProperties(graph.PropertyMap{
		common.Name:     name,
		common.ObjectID: must.NewUUIDv4().String(),
		ad.DomainSID:    domainSID,
	}), ad.Entity, ad.Computer)
}

func (s *GraphTestContext) NewActiveDirectoryUser(name, domainSID string, isTierZero ...bool) *graph.Node {
	return s.NewNode(graph.AsProperties(graph.PropertyMap{
		common.Name:     name,
		common.ObjectID: strings.ToUpper(must.NewUUIDv4().String()),
		ad.DomainSID:    domainSID,
	}), ad.Entity, ad.User)
}

func (s *GraphTestContext) NewActiveDirectoryGroup(name, domainSID string) *graph.Node {
	return s.NewNode(graph.AsProperties(graph.PropertyMap{
		common.Name:     name,
		common.ObjectID: must.NewUUIDv4().String(),
		ad.DomainSID:    domainSID,
	}), ad.Entity, ad.Group)
}

func (s *GraphTestContext) NewActiveDirectoryLocalGroup(name, domainSID string) *graph.Node {
	return s.NewNode(graph.AsProperties(graph.PropertyMap{
		common.Name:     name,
		common.ObjectID: must.NewUUIDv4().String(),
		ad.DomainSID:    domainSID,
	}), ad.Entity, ad.LocalGroup)
}

func (s *GraphTestContext) NewActiveDirectoryOU(name, domainSID string, blocksInheritance bool) *graph.Node {
	return s.NewNode(graph.AsProperties(graph.PropertyMap{
		common.Name:          name,
		common.ObjectID:      must.NewUUIDv4().String(),
		ad.DomainSID:         domainSID,
		ad.BlocksInheritance: blocksInheritance,
	}), ad.Entity, ad.OU)
}

func (s *GraphTestContext) NewActiveDirectoryGPO(name, domainSID string) *graph.Node {
	return s.NewNode(graph.AsProperties(graph.PropertyMap{
		common.Name:     name,
		common.ObjectID: must.NewUUIDv4().String(),
		ad.DomainSID:    domainSID,
	}), ad.Entity, ad.GPO)
}

func (s *GraphTestContext) NewActiveDirectoryNTAuthStore(name, domainSID string) *graph.Node {
	return s.NewNode(graph.AsProperties(graph.PropertyMap{
		common.Name:        name,
		common.ObjectID:    must.NewUUIDv4().String(),
		ad.DomainSID:       domainSID,
		ad.CertThumbprints: []string{"a", "b", "c"},
	}), ad.Entity, ad.NTAuthStore)
}

func (s *GraphTestContext) NewActiveDirectoryEnterpriseCA(name, domainSID string) *graph.Node {
	return s.NewNode(graph.AsProperties(graph.PropertyMap{
		common.Name:     name,
		common.ObjectID: must.NewUUIDv4().String(),
		ad.DomainSID:    domainSID,
	}), ad.Entity, ad.EnterpriseCA)
}

func (s *GraphTestContext) NewActiveDirectoryEnterpriseCAWithThumbprint(name, domainSID, certThumbprint string) *graph.Node {
	return s.NewNode(graph.AsProperties(graph.PropertyMap{
		common.Name:       name,
		common.ObjectID:   must.NewUUIDv4().String(),
		ad.DomainSID:      domainSID,
		ad.CertThumbprint: certThumbprint,
	}), ad.Entity, ad.EnterpriseCA)
}

func (s *GraphTestContext) NewActiveDirectoryRootCA(name, domainSID string) *graph.Node {
	return s.NewNode(graph.AsProperties(graph.PropertyMap{
		common.Name:     name,
		common.ObjectID: must.NewUUIDv4().String(),
		ad.DomainSID:    domainSID,
	}), ad.Entity, ad.RootCA)
}

func (s *GraphTestContext) NewActiveDirectoryRootCAWithThumbprint(name, domainSID string, certThumbprint string) *graph.Node {
	return s.NewNode(graph.AsProperties(graph.PropertyMap{
		common.Name:       name,
		common.ObjectID:   must.NewUUIDv4().String(),
		ad.DomainSID:      domainSID,
		ad.CertThumbprint: certThumbprint,
	}), ad.Entity, ad.RootCA)
}

func (s *GraphTestContext) NewActiveDirectoryCertTemplate(name, domainSID string, requiresManagerApproval, authenticationEnabled, enrolleeSupplieSubject, subjectAltRequireUpn bool, schemaVersion, authorizedSignatures int, ekus, applicationPolicies []string) *graph.Node {
	return s.NewNode(graph.AsProperties(graph.PropertyMap{
		common.Name:                name,
		common.ObjectID:            must.NewUUIDv4().String(),
		ad.DomainSID:               domainSID,
		ad.RequiresManagerApproval: requiresManagerApproval,
		ad.AuthenticationEnabled:   authenticationEnabled,
		ad.EnrolleeSuppliesSubject: enrolleeSupplieSubject,
		ad.SchemaVersion:           float64(schemaVersion),
		ad.AuthorizedSignatures:    float64(authorizedSignatures),
		ad.EKUs:                    ekus,
		ad.ApplicationPolicies:     applicationPolicies,
		ad.SubjectAltRequireUPN:    subjectAltRequireUpn,
	}), ad.Entity, ad.CertTemplate)
}

func (s *GraphTestContext) setupAzure() {
	s.Harness.AZBaseHarness.Setup(s)
	s.Harness.AZGroupMembership.Setup(s)
	s.Harness.AZEntityPanelHarness.Setup(s)
	s.Harness.AZMGApplicationReadWriteAllHarness.Setup(s)
	s.Harness.AZMGAppRoleManagementReadWriteAllHarness.Setup(s)
	s.Harness.AZMGDirectoryReadWriteAllHarness.Setup(s)
	s.Harness.AZMGGroupReadWriteAllHarness.Setup(s)
	s.Harness.AZMGGroupMemberReadWriteAllHarness.Setup(s)
	s.Harness.AZMGRoleManagementReadWriteDirectoryHarness.Setup(s)
	s.Harness.AZMGServicePrincipalEndpointReadWriteAllHarness.Setup(s)
	s.Harness.AZInboundControlHarness.Setup(s)
}

func (s *GraphTestContext) setupActiveDirectory() {
	// startServer a host of Tier Zero tagged assets
	s.Harness.RootADHarness.Setup(s)

	// startServer GPO Enforcement harness
	s.Harness.GPOEnforcement.Setup(s)

	// startServer CanRDP harness
	s.Harness.RDP.Setup(s)
	s.Harness.RDPB.Setup(s)

	//startServer Session Harness
	s.Harness.Session.Setup(s)

	//startServer localgroup harness
	s.Harness.LocalGroupSQL.Setup(s)

	//startServer control harnesses
	s.Harness.OutboundControl.Setup(s)
	s.Harness.InboundControl.Setup(s)

	s.Harness.OUHarness.Setup(s)
	s.Harness.MembershipHarness.Setup(s)
	s.Harness.ForeignHarness.Setup(s)
	s.Harness.TrustDCSync.Setup(s)
	s.Harness.Completeness.Setup(s)
	s.Harness.ShortcutHarness.Setup(s)

	s.Harness.AssetGroupComboNodeHarness.Setup(s)
}
