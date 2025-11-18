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
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/specterops/bloodhound/cmd/api/src/test"
	"github.com/specterops/bloodhound/cmd/api/src/test/must"
	"github.com/specterops/bloodhound/packages/go/graphschema/ad"
	"github.com/specterops/bloodhound/packages/go/graphschema/azure"
	"github.com/specterops/bloodhound/packages/go/graphschema/common"
	_ "github.com/specterops/dawgs/drivers/neo4j"
	"github.com/specterops/dawgs/graph"
)

var DefaultRelProperties = graph.AsProperties(graph.PropertyMap{
	common.LastSeen: time.Now().Format(time.RFC3339),
})

// NewGraphTestContext creates a new GraphTestContext
//
// Deprecated: this suite of integration utils is deprecated and should be avoided
// Integration tests should be updated to reflect the latest standards.
// See commit https://github.com/SpecterOps/BloodHound/commit/a6cc43013fd769b97cc52cbc60b2314494054c9a#diff-e6bcb50873ade3cf33cef4e3e0ff566fb8ac1367b4ade36f4511bc2172a760e1
// for implementation guidance. Additional detailed information can be found in Confluence.
func NewGraphTestContext(t *testing.T, schema graph.Schema) *GraphTestContext {
	testCtx := test.NewContext(t)

	return &GraphTestContext{
		testCtx: testCtx,
		Graph:   NewGraphContext(t, testCtx, schema),
	}
}

type GraphTestContext struct {
	testCtx test.Context
	Harness HarnessDetails
	Graph   *GraphContext
}

// TODO: This is a responsibility violation
func (s *GraphTestContext) Context() test.Context {
	return s.testCtx
}

func (s *GraphTestContext) NodeObjectID(node *graph.Node) string {
	objectID, err := node.Properties.Get(common.ObjectID.String()).String()

	test.RequireNilErrf(s.testCtx, err, "expected node %d to have a valid %s property: %v", node.ID, common.ObjectID.String(), err)

	return objectID
}

func (s *GraphTestContext) FindNode(criteria graph.Criteria) *graph.Node {
	var (
		node *graph.Node
		err  error
	)

	s.Graph.ReadTransaction(s.testCtx, func(tx graph.Transaction) error {
		node, err = tx.Nodes().Filter(criteria).First()
		return err
	})

	return node
}

func (s *GraphTestContext) UpdateNode(node *graph.Node) {
	s.Graph.WriteTransaction(s.testCtx, func(tx graph.Transaction) error {
		return tx.UpdateNode(node)
	})
}

func (s *GraphTestContext) InitializeHarness(harness GraphTestHarness) {
	s.Graph.WriteTransaction(s.testCtx, func(tx graph.Transaction) error {
		harness.Setup(s)
		return nil
	})
}

func (s *GraphTestContext) DatabaseTest(dbDelegate func(harness HarnessDetails, db graph.Database)) {
	dbDelegate(s.Harness, s.Graph.Database)
}

func (s *GraphTestContext) SetupHarness(setup func(harness *HarnessDetails) error) {
	s.Graph.WriteTransaction(s.testCtx, func(tx graph.Transaction) error {
		return setup(&s.Harness)
	})
}

func (s *GraphTestContext) DatabaseTestWithSetup(setup func(harness *HarnessDetails) error, dbDelegate func(harness HarnessDetails, db graph.Database)) {
	s.Graph.WriteTransaction(s.testCtx, func(tx graph.Transaction) error {
		return setup(&s.Harness)
	})

	dbDelegate(s.Harness, s.Graph.Database)
}

func (s *GraphTestContext) BatchTest(batchDelegate func(harness HarnessDetails, batch graph.Batch), assertionDelegate func(details HarnessDetails, tx graph.Transaction)) {
	s.SetupAzureAndActiveDirectory()

	s.Graph.BatchOperation(s.testCtx, func(batch graph.Batch) error {
		batchDelegate(s.Harness, batch)
		return nil
	})

	s.Graph.ReadTransaction(s.testCtx, func(tx graph.Transaction) error {
		assertionDelegate(s.Harness, tx)
		return nil
	})
}

func (s *GraphTestContext) TransactionalTest(txDelegate func(harness HarnessDetails, tx graph.Transaction)) {
	s.SetupAzureAndActiveDirectory()

	s.Graph.WriteTransaction(s.testCtx, func(tx graph.Transaction) error {
		txDelegate(s.Harness, tx)
		return nil
	})
}

func (s *GraphTestContext) ReadTransactionTestWithSetup(setup func(harness *HarnessDetails) error, txDelegate func(harness HarnessDetails, tx graph.Transaction)) {
	s.Graph.WriteTransaction(s.testCtx, func(tx graph.Transaction) error {
		return setup(&s.Harness)
	})

	s.Graph.ReadTransaction(s.testCtx, func(tx graph.Transaction) error {
		txDelegate(s.Harness, tx)
		return nil
	})
}

func (s *GraphTestContext) DatabaseTransactionTestWithSetup(setup func(harness *HarnessDetails) error, dbDelegate func(harness HarnessDetails, db graph.Database, tx graph.Transaction)) {
	// Wipe the DB before executing the test
	s.Graph.WriteTransaction(s.testCtx, func(tx graph.Transaction) error {
		return tx.Nodes().Delete()
	})

	s.Graph.WriteTransaction(s.testCtx, func(tx graph.Transaction) error {
		return setup(&s.Harness)
	})

	s.Graph.ReadTransaction(s.testCtx, func(tx graph.Transaction) error {
		dbDelegate(s.Harness, s.Graph.Database, tx)
		return nil
	})
}

func (s *GraphTestContext) WriteTransactionTestWithSetup(setup func(harness *HarnessDetails) error, txDelegate func(harness HarnessDetails, tx graph.Transaction)) {
	s.Graph.WriteTransaction(s.testCtx, func(tx graph.Transaction) error {
		return setup(&s.Harness)
	})

	s.Graph.WriteTransaction(s.testCtx, func(tx graph.Transaction) error {
		txDelegate(s.Harness, tx)
		return nil
	})
}

func (s *GraphTestContext) NewNode(properties *graph.Properties, kinds ...graph.Kind) *graph.Node {
	var (
		node *graph.Node
		err  error
	)

	s.Graph.WriteTransaction(s.testCtx, func(tx graph.Transaction) error {
		node, err = tx.CreateNode(properties, kinds...)
		return err
	})

	return node
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

func (s *GraphTestContext) NewCustomAzureUser(properties *graph.Properties) *graph.Node {
	return s.NewNode(properties, azure.Entity, azure.User)
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

func (s *GraphTestContext) NewRelationship(startNode, endNode *graph.Node, kind graph.Kind, propertyBags ...*graph.Properties) *graph.Relationship {
	var (
		relationshipProperties = graph.NewPropertiesRed()
		relationship           *graph.Relationship
		err                    error
	)

	for _, additionalProperties := range propertyBags {
		relationshipProperties.Merge(additionalProperties)
	}

	s.Graph.WriteTransaction(s.testCtx, func(tx graph.Transaction) error {
		relationship, err = tx.CreateRelationshipByIDs(startNode.ID, endNode.ID, kind, relationshipProperties)
		return err
	})

	return relationship
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
		azure.TenantID:  tenantID,
		azure.License:   "license",
	}), azure.Entity, azure.Tenant)
}

// NewAzureBase creates a new AZBase (azure.Entity) node
func (s *GraphTestContext) NewAzureBase(name, objectID, tenantID string) *graph.Node {
	return s.NewNode(graph.AsProperties(graph.PropertyMap{
		common.Name:     name,
		common.ObjectID: objectID,
		azure.TenantID:  tenantID,
		azure.License:   "license",
	}), azure.Entity)
}

func (s *GraphTestContext) NewActiveDirectoryDomain(name, objectID string, blocksInheritance, collected bool, additionalKinds ...graph.Kind) *graph.Node {
	if collected {
		s.Harness.NumCollectedActiveDirectoryDomains++
	}

	return s.NewNode(graph.AsProperties(graph.PropertyMap{
		common.Name:          name,
		common.ObjectID:      objectID,
		ad.DomainSID:         objectID,
		common.Collected:     collected,
		ad.BlocksInheritance: blocksInheritance,
	}), append(additionalKinds, ad.Entity, ad.Domain)...)
}

func (s *GraphTestContext) NewActiveDirectoryComputer(name, domainSID string) *graph.Node {
	return s.NewNode(graph.AsProperties(graph.PropertyMap{
		common.Name:     name,
		common.ObjectID: must.NewUUIDv4().String(),
		ad.DomainSID:    domainSID,
	}), ad.Entity, ad.Computer)
}

func (s *GraphTestContext) NewActiveDirectoryContainer(name, domainSID string) *graph.Node {
	return s.NewNode(graph.AsProperties(graph.PropertyMap{
		common.Name:     name,
		common.ObjectID: must.NewUUIDv4().String(),
		ad.DomainSID:    domainSID,
	}), ad.Entity, ad.Container)
}

func (s *GraphTestContext) NewActiveDirectoryUser(name, domainSID string, isTierZero ...bool) *graph.Node {

	propertyMap := graph.PropertyMap{
		common.Name:     name,
		common.ObjectID: strings.ToUpper(must.NewUUIDv4().String()),
		ad.DomainSID:    domainSID,
	}

	if isTierZero != nil && isTierZero[0] {
		propertyMap[common.SystemTags] = ad.AdminTierZero
	}

	return s.NewNode(graph.AsProperties(propertyMap), ad.Entity, ad.User)
}

func (s *GraphTestContext) NewCustomActiveDirectoryUser(properties *graph.Properties) *graph.Node {
	return s.NewNode(properties, ad.Entity, ad.User)
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

func (s *GraphTestContext) NewActiveDirectoryAIACA(name, domainSID string, certThumbprint string, certChain []string) *graph.Node {
	return s.NewNode(graph.AsProperties(graph.PropertyMap{
		common.Name:       name,
		common.ObjectID:   must.NewUUIDv4().String(),
		ad.DomainSID:      domainSID,
		ad.CertThumbprint: certThumbprint,
		ad.CertChain:      certChain,
	}), ad.Entity, ad.AIACA)
}

func (s *GraphTestContext) NewActiveDirectoryCertTemplate(name, domainSID string, data CertTemplateData) *graph.Node {
	return s.NewNode(graph.AsProperties(graph.PropertyMap{
		common.Name:                      name,
		common.ObjectID:                  must.NewUUIDv4().String(),
		ad.DomainSID:                     domainSID,
		ad.RequiresManagerApproval:       data.RequiresManagerApproval,
		ad.AuthenticationEnabled:         data.AuthenticationEnabled,
		ad.SchannelAuthenticationEnabled: data.SchannelAuthenticationEnabled,
		ad.EnrolleeSuppliesSubject:       data.EnrolleeSuppliesSubject,
		ad.NoSecurityExtension:           data.NoSecurityExtension,
		ad.SchemaVersion:                 data.SchemaVersion,
		ad.AuthorizedSignatures:          data.AuthorizedSignatures,
		ad.EffectiveEKUs:                 data.EffectiveEKUs,
		ad.ApplicationPolicies:           data.ApplicationPolicies,
		ad.SubjectAltRequireUPN:          data.SubjectAltRequireUPN,
		ad.SubjectAltRequireSPN:          data.SubjectAltRequireSPN,
		ad.SubjectAltRequireDNS:          data.SubjectAltRequireDNS,
		ad.SubjectAltRequireDomainDNS:    data.SubjectAltRequireDomainDNS,
		ad.SubjectAltRequireEmail:        data.SubjectAltRequireEmail,
		ad.CertificatePolicy:             data.CertificatePolicy,
	}), ad.Entity, ad.CertTemplate)
}

func (s *GraphTestContext) NewActiveDirectoryCertTemplateWoutSchannelAuthEnabled(name, domainSID string, data CertTemplateData) *graph.Node {
	return s.NewNode(graph.AsProperties(graph.PropertyMap{
		common.Name:                   name,
		common.ObjectID:               must.NewUUIDv4().String(),
		ad.DomainSID:                  domainSID,
		ad.RequiresManagerApproval:    data.RequiresManagerApproval,
		ad.AuthenticationEnabled:      data.AuthenticationEnabled,
		ad.EnrolleeSuppliesSubject:    data.EnrolleeSuppliesSubject,
		ad.NoSecurityExtension:        data.NoSecurityExtension,
		ad.SchemaVersion:              data.SchemaVersion,
		ad.AuthorizedSignatures:       data.AuthorizedSignatures,
		ad.EffectiveEKUs:              data.EffectiveEKUs,
		ad.ApplicationPolicies:        data.ApplicationPolicies,
		ad.SubjectAltRequireUPN:       data.SubjectAltRequireUPN,
		ad.SubjectAltRequireSPN:       data.SubjectAltRequireSPN,
		ad.SubjectAltRequireDNS:       data.SubjectAltRequireDNS,
		ad.SubjectAltRequireDomainDNS: data.SubjectAltRequireDomainDNS,
		ad.SubjectAltRequireEmail:     data.SubjectAltRequireEmail,
		ad.CertificatePolicy:          data.CertificatePolicy,
	}), ad.Entity, ad.CertTemplate)
}

func (s *GraphTestContext) NewActiveDirectoryIssuancePolicy(name, domainSID string, certTemplateOID string) *graph.Node {
	return s.NewNode(graph.AsProperties(graph.PropertyMap{
		common.Name:        name,
		common.ObjectID:    must.NewUUIDv4().String(),
		ad.DomainSID:       domainSID,
		ad.CertTemplateOID: certTemplateOID,
	}), ad.Entity, ad.IssuancePolicy)
}

type CertTemplateData struct {
	RequiresManagerApproval       bool
	AuthenticationEnabled         bool
	SchannelAuthenticationEnabled bool
	EnrolleeSuppliesSubject       bool
	SubjectAltRequireUPN          bool
	SubjectAltRequireSPN          bool
	SubjectAltRequireDNS          bool
	SubjectAltRequireDomainDNS    bool
	SubjectAltRequireEmail        bool
	NoSecurityExtension           bool
	SchemaVersion                 float64
	AuthorizedSignatures          float64
	EffectiveEKUs                 []string
	ApplicationPolicies           []string
	CertificatePolicy             []string
}

func (s *GraphTestContext) SetupAzureAndActiveDirectory() {
	s.SetupAzure()
	s.SetupActiveDirectory()
}

func (s *GraphTestContext) SetupAzure() {
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
	s.Harness.AZManagementGroup.Setup(s)
}

func (s *GraphTestContext) SetupActiveDirectory() {
	// startServer a host of Tier Zero tagged assets
	s.Harness.RootADHarness.Setup(s)

	// startServer GPO Enforcement harness
	s.Harness.GPOEnforcement.Setup(s)

	// startServer CanRDP harness
	s.Harness.RDP.Setup(s)
	s.Harness.RDPB.Setup(s)

	// startServer Session Harness
	s.Harness.Session.Setup(s)

	// startServer localgroup harness
	s.Harness.LocalGroupSQL.Setup(s)

	// startServer CanBackup harness
	s.Harness.BackupOperators.Setup(s)
	s.Harness.BackupOperators2.Setup(s)

	// startServer control harnesses
	s.Harness.OutboundControl.Setup(s)
	s.Harness.InboundControl.Setup(s)

	s.Harness.OUHarness.Setup(s)
	s.Harness.MembershipHarness.Setup(s)
	s.Harness.ForeignHarness.Setup(s)
	s.Harness.TrustDCSync.Setup(s)
	s.Harness.ShortcutHarness.Setup(s)

	s.Harness.AssetGroupComboNodeHarness.Setup(s)
}
