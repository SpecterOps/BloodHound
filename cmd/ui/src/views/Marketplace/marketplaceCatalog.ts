// Copyright 2026 Specter Ops, Inc.
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

import type { ProductCardProps } from './ProductCard';

export type MarketplacePublisher = 'specterops' | 'community' | 'partner';
export type MarketplaceAvailability = 'general' | 'early-access';

export type CatalogItem = Omit<ProductCardProps, 'categoryLabel' | 'linkLabel'> & {
    publisher: MarketplacePublisher;
    availability: MarketplaceAvailability;
};

type GenerallyAvailableCatalogItem = Omit<CatalogItem, 'availability'>;

const generallyAvailable = (items: readonly GenerallyAvailableCatalogItem[]): readonly CatalogItem[] =>
    items.map((item) => ({ ...item, availability: 'general' }));

export const COMMUNITY_EXTENSIONS_DISCLAIMER =
    'All code linked via this library is provided “as is,” without review, approval, or endorsement by SpecterOps, regardless of authorship. It has not been audited for accuracy, security, or fitness for any purpose. Use at your own risk. You are solely responsible for testing, validating, and ensuring the code meets your requirements before use in any environment. SpecterOps is not responsible for any damages, losses, or security issues arising from the use of any linked code.';

export const enterpriseExtensions = [
    {
        name: 'AWS',
        author: 'SpecterOps',
        publisher: 'specterops',
        availability: 'early-access',
        badge: 'Beta',
        logoPath: '/img/product-logos/aws.svg',
        description:
            'Surface traversable AWS attack paths, identify escalation chokepoints, and define Privilege Zones around vital data stores, roles, and assets.',
        href: 'https://pages.specterops.io/WEB---2026---AWS-Launch_AWS-Beta-Sign-Up.html',
    },
    {
        name: 'Microsoft Entra Agent ID',
        author: 'SpecterOps',
        publisher: 'specterops',
        availability: 'early-access',
        badge: 'Beta',
        logoPath: '/img/product-logos/entra-id.png',
        description:
            'Map agent identity blueprints, agent identities, Copilot Studio agents, Power Automate flows, and their relationships to existing Entra and Azure identities and resources.',
        href: 'https://pages.specterops.io/WEB---2026---AWS-Launch_AWS-Beta-Sign-Up.html',
    },
    {
        name: 'GitHub',
        author: 'SpecterOps',
        publisher: 'specterops',
        availability: 'general',
        logoPath: '/img/product-logos/github.svg',
        description:
            'Model GitHub organizations, identities, repositories, workflows, secrets, roles, and their relationships as graph data.',
        href: 'https://bloodhound.specterops.io/opengraph/extensions/github/overview',
    },
    {
        name: 'Okta',
        author: 'SpecterOps',
        publisher: 'specterops',
        availability: 'general',
        logoPath: '/img/product-logos/okta.png',
        description:
            'Model Okta users, groups, applications, roles, policies, and related relationships to reveal identity-provider attack paths.',
        href: 'https://bloodhound.specterops.io/opengraph/extensions/okta/overview',
    },
    {
        name: 'Jamf',
        author: 'SpecterOps',
        publisher: 'specterops',
        availability: 'general',
        logoPath: '/img/product-logos/jamf.svg',
        description:
            'Model Jamf Pro users, groups, sites, scripts, API integrations, and related relationships across managed Mac environments.',
        href: 'https://bloodhound.specterops.io/opengraph/extensions/jamf/overview',
    },
] satisfies readonly CatalogItem[];

export const communityExtensions = generallyAvailable([
    {
        name: '1Password',
        author: 'SpecterOps',
        publisher: 'specterops',
        logoPath: '/img/product-logos/community-1password.png',
        description:
            'The 1Password for Business OpenGraph extension lets you bring your 1Password ACL data into BloodHound’s graph-analysis framework.',
        href: 'https://github.com/SpecterOps/1PassHound',
    },
    {
        name: 'Ansible',
        author: 'The Sleek Boy Company',
        publisher: 'community',
        logoPath: '/img/product-logos/community-ansible.png',
        description:
            'AnsibleHound collects Ansible AWX and Tower structure and permissions into a navigable OpenGraph attack-path graph.',
        href: 'https://github.com/TheSleekBoyCompany/AnsibleHound',
    },
    {
        name: 'Active Directory (AD)',
        author: 'martinsohn',
        publisher: 'community',
        logoPath: '/img/product-logos/community-active-directory.png',
        description:
            'A growing library of extensions supporting Active Directory analysis, including ADAttributeHound, ManagerOfHound, GhostHound, and ProfileHound.',
        href: 'https://github.com/martinsohn/ADAttributeHound',
    },
    {
        name: 'Atlassian',
        author: 'werdhaihai',
        publisher: 'community',
        logoPath: '/img/product-logos/community-atlassian.png',
        description:
            'AtlassianHound collects foundational Jira and Confluence access data and exports it in the BloodHound OpenGraph format.',
        href: 'https://github.com/werdhaihai/AtlassianHound',
    },
    {
        name: 'Cisco Duo Security',
        author: 'julian1j',
        publisher: 'community',
        logoPath: '/img/product-logos/community-cisco-duo.png',
        description:
            'DuoHound extracts data from Duo Security’s Admin API and converts it to BloodHound’s OpenGraph format.',
        href: 'https://github.com/julian1j/DuoHound',
    },
    {
        name: 'Credentials',
        author: 'Netwrix and C0KERNEL',
        publisher: 'community',
        logoPath: '/img/product-logos/community-specterops.png',
        description:
            'AIHound and SecretHound convert credential and secret-scanning results into OpenGraph data for attack-path analysis.',
        href: 'https://github.com/netwrix/AIHound',
    },
    {
        name: 'CyberArk',
        author: 'jazofra',
        publisher: 'community',
        logoPath: '/img/product-logos/cyberark.png',
        description:
            'CyberArkHound exports CyberArk PVWA users, groups, safes, accounts, and permissions into BloodHound-compatible OpenGraph JSON.',
        href: 'https://github.com/jazofra/CyberArkHound',
    },
    {
        name: 'DevOps',
        author: 'h4wkst3r',
        publisher: 'community',
        logoPath: '/img/product-logos/community-user-group.png',
        description:
            'Dop2Mop is a proof-of-concept OpenGraph collector that maps attack paths from DevOps to MLOps infrastructure.',
        href: 'https://github.com/h4wkst3r/Dop2Mop',
    },
    {
        name: 'Entra ID',
        author: 'MichaelGrafnetter',
        publisher: 'community',
        logoPath: '/img/product-logos/entra-id.png',
        description:
            'A growing library of extensions supporting Entra analysis, including EntraAuthPolicyHound and EntraSSSOHound.',
        href: 'https://github.com/MichaelGrafnetter/EntraAuthPolicyHound',
    },
    {
        name: 'FreeIPA',
        author: 'lvruibr',
        publisher: 'community',
        logoPath: '/img/product-logos/community-freeipa.png',
        description: 'IDMHound is a collector for FreeIPA and Red Hat Identity Management environments.',
        href: 'https://github.com/lvruibr/idmhound',
    },
    {
        name: 'GitLab',
        author: 'Compass Security',
        publisher: 'community',
        logoPath: '/img/product-logos/community-gitlab.png',
        description:
            'GitLabHound collects GitLab permissions and relationships into a navigable OpenGraph attack-path graph.',
        href: 'https://github.com/CompassSecurity/GitLabHound',
    },
    {
        name: 'Google Cloud Platform (GCP)',
        author: 'F41zK4r1m',
        publisher: 'community',
        logoPath: '/img/product-logos/community-gcp.png',
        description:
            'A growing library of extensions supporting Google Cloud Platform analysis, including GCP-Hound and GCPwn.',
        href: 'https://github.com/F41zK4r1m/GCP-Hound',
    },
    {
        name: 'Kubernetes',
        author: 'Dovetail',
        publisher: 'community',
        logoPath: '/img/product-logos/community-kubernetes.png',
        description: 'ClusterHound brings Kubernetes identities, resources, and relationships into BloodHound CE.',
        href: 'https://github.com/dovesec/ClusterHound',
    },
    {
        name: 'Microsoft Exchange',
        author: 'FilipPwn',
        publisher: 'community',
        logoPath: '/img/product-logos/community-microsoft-exchange.png',
        description: 'ExchangeHound is an OpenGraph extension for Microsoft Exchange on-premises environments.',
        href: 'https://github.com/FilipPwn/exchangehound',
    },
    {
        name: 'MSSQL',
        author: 'SpecterOps',
        publisher: 'specterops',
        logoPath: '/img/product-logos/community-mssql.png',
        description:
            'MSSQLHound collects SQL Server principals, databases, permissions, and relationships into OpenGraph data.',
        href: 'https://github.com/SpecterOps/MSSQLHound',
    },
    {
        name: 'Network',
        author: 'Mor David',
        publisher: 'community',
        logoPath: '/img/product-logos/community-network.png',
        description:
            'NetworkHound discovers domain hosts, resolves addresses, and scans network services to model additional infrastructure.',
        href: 'https://github.com/mordavid/NetworkHound',
    },
    {
        name: 'Oracle Cloud Infrastructure',
        author: 'NetSPI',
        publisher: 'community',
        logoPath: '/img/product-logos/community-oracle.png',
        description:
            'OCInferno is an OCI security assessment framework with an OpenGraph generator for graph-based attack-path analysis.',
        href: 'https://github.com/NetSPI/ocinferno',
    },
    {
        name: 'Ping',
        author: 'Andy Robbins',
        publisher: 'community',
        logoPath: '/img/product-logos/community-specterops.png',
        description: 'PingOneHound collects identity and access data from the PingOne identity platform.',
        href: 'https://github.com/andyrobbins/PingOneHound',
    },
    {
        name: 'Resource Access Control Facility (RACF)',
        author: '4-L3X',
        publisher: 'community',
        logoPath: '/img/product-logos/community-user-group.png',
        description: 'RacfHound is a BloodHound ingestor for the RACF database in z/OS mainframes.',
        href: 'https://github.com/4-L3X/racfhound',
    },
    {
        name: 'runZero',
        author: 'runZero',
        publisher: 'community',
        logoPath: '/img/product-logos/community-runzero.png',
        description: 'runZeroHound brings runZero Exposure Management data into BloodHound via OpenGraph.',
        href: 'https://github.com/runZeroInc/runZeroHound',
    },
    {
        name: 'Salesforce',
        author: 'NetSPI',
        publisher: 'community',
        logoPath: '/img/product-logos/community-salesforce.png',
        description: 'A growing library of Salesforce extensions, including ForceHound and SFHound.',
        href: 'https://github.com/NetSPI/ForceHound',
    },
    {
        name: 'Snowflake',
        author: 'SpecterOps',
        publisher: 'specterops',
        logoPath: '/img/product-logos/community-snowflake.png',
        description:
            'SnowHound maps key Snowflake tenant identities, resources, and relationships into OpenGraph data.',
        href: 'https://github.com/SpecterOps/SnowHound',
    },
    {
        name: 'System Center',
        author: 'SpecterOps and community',
        publisher: 'community',
        logoPath: '/img/product-logos/community-specterops.png',
        description:
            'A growing library of System Center extensions, including ConfigManBearPig, SCCM SQL Collector, and SCOMHound.',
        href: 'https://github.com/SpecterOps/ConfigManBearPig',
    },
    {
        name: 'Tailscale',
        author: 'KingOfTheNOPs',
        publisher: 'community',
        logoPath: '/img/product-logos/community-specterops.png',
        description: 'TailscaleHound is a BloodHound OpenGraph collector for Tailscale.',
        href: 'https://github.com/KingOfTheNOPs/TailscaleHound',
    },
    {
        name: 'vCenter',
        author: 'Mor David',
        publisher: 'community',
        logoPath: '/img/product-logos/community-vcenter.png',
        description:
            'vCenterHound collects vCenter infrastructure entities and permissions into a BloodHound-compatible OpenGraph.',
        href: 'https://github.com/MorDavid/vCenterHound',
    },
    {
        name: 'Windows',
        author: 'dazzyddos and community',
        publisher: 'community',
        logoPath: '/img/product-logos/community-windows.png',
        description:
            'A growing library of Microsoft Windows extensions, including PrivHound, ShareHound, and TaskHound.',
        href: 'https://github.com/dazzyddos/PrivHound',
    },
] satisfies readonly GenerallyAvailableCatalogItem[]);

export const enterpriseIntegrations = generallyAvailable([
    {
        name: 'Jira',
        author: 'Atlassian',
        publisher: 'specterops',
        logoPath: '/img/product-logos/jira.svg',
        description:
            'Automatically sync BloodHound Enterprise attack path findings to Jira issues for remediation tracking in existing team workflows.',
        href: 'https://bloodhound.specterops.io/integrations/atlassian/jira/configure',
    },
    {
        name: 'Splunk SIEM',
        author: 'Splunk',
        publisher: 'specterops',
        logoPath: '/img/product-logos/splunk.svg',
        description:
            'Ingest attack path, posture, and impacted-principal data into Splunk with pre-built dashboards and alerts.',
        href: 'https://bloodhound.specterops.io/integrations/splunk/siem/install',
    },
    {
        name: 'Splunk SOAR',
        author: 'Splunk',
        publisher: 'specterops',
        logoPath: '/img/product-logos/splunk.svg',
        description:
            'Pull BloodHound findings into Splunk SOAR, enrich alerts with graph context, and automate response playbooks.',
        href: 'https://bloodhound.specterops.io/integrations/splunk/soar/configure',
    },
    {
        name: 'Google SecOps',
        author: 'Google',
        publisher: 'specterops',
        logoPath: '/img/product-logos/google-secops.svg',
        description:
            'Automatically sync attack path findings to Google SecOps cases so teams can investigate, track, and coordinate remediation.',
        href: 'https://bloodhound.specterops.io/integrations/google-secops/configure',
    },
    {
        name: 'ServiceNow Security Incident Response',
        author: 'ServiceNow',
        publisher: 'specterops',
        logoPath: '/img/product-logos/servicenow.svg',
        description:
            'Generate Security Incident Response tickets to track and monitor attack path findings from BloodHound Enterprise.',
        href: 'https://bloodhound.specterops.io/integrations/service-now/security-incident-response/configure',
    },
    {
        name: 'ServiceNow Vulnerability Response',
        author: 'ServiceNow',
        publisher: 'specterops',
        logoPath: '/img/product-logos/servicenow.svg',
        description:
            'Automate vulnerable-item creation and remediation workflows in ServiceNow based on BloodHound attack path findings.',
        href: 'https://bloodhound.specterops.io/integrations/service-now/vulnerability-response/configure',
    },
    {
        name: 'Cortex XSOAR',
        author: 'Palo Alto Networks',
        publisher: 'specterops',
        logoPath: '/img/product-logos/cortex-xsoar.svg',
        description:
            'Ingest and manage attack path findings as Cortex XSOAR incidents with remediation guidance and posture context.',
        href: 'https://bloodhound.specterops.io/integrations/cortex-xsoar/configure',
    },
    {
        name: 'Tines',
        author: 'Tines',
        publisher: 'partner',
        logoPath: '/img/product-logos/tines.svg',
        description:
            'Use BloodHound findings and graph context in Tines workflows to automate investigation, enrichment, notification, and remediation tasks.',
        href: 'https://www.tines.com/library/tools/bloodhound/',
    },
    {
        name: 'Axonius',
        author: 'Axonius',
        publisher: 'partner',
        logoPath: '/img/product-logos/axonius.svg',
        description:
            'Fetch and catalog BloodHound Enterprise users, devices, groups, Tier Zero assets, privileged access, and assets on attack paths in Axonius.',
        href: 'https://docs.axonius.com/docs/bloodhound',
    },
    {
        name: 'Quest On Demand Audit',
        author: 'Quest',
        publisher: 'partner',
        logoPath: '/img/product-logos/quest-monogram.svg',
        description:
            'Ingest BloodHound Enterprise Tier Zero assets and attack path edge data into Quest On Demand Audit to monitor critical changes and inform remediation.',
        href: 'https://support.quest.com/kb/4375854/how-to-integrate-on-demand-audit-with-specterops-bloodhound-enterprise',
    },
    {
        name: 'Cisco Duo Single Sign-On',
        author: 'Cisco',
        publisher: 'partner',
        logoPath: '/img/product-logos/cisco-duo.svg',
        description:
            'Protect BloodHound Enterprise SAML logins with Duo Single Sign-On, multi-factor authentication, endpoint verification, and flexible access policies.',
        href: 'https://duo.com/docs/sso-bloodhound-enterprise',
    },
] satisfies readonly GenerallyAvailableCatalogItem[]);

export const communityIntegrations = generallyAvailable([
    {
        name: 'BloodHound MCP',
        author: 'mwnickerson',
        publisher: 'community',
        logoPath: '/img/logo-red-transparent-logo-only.svg',
        description:
            'Connect BloodHound CE to MCP-aware assistants for signed API queries, attack-path analysis, collection uploads, and OpenGraph management.',
        href: 'https://github.com/mwnickerson/bloodhound_mcp',
    },
    {
        name: 'CypherHound',
        author: 'fin3ss3g0d',
        publisher: 'community',
        logoPath: '/img/logo-red-transparent-logo-only.svg',
        description:
            'A terminal companion for BloodHound datasets with tools to import saved queries and run query-library content in BloodHound CE.',
        href: 'https://github.com/fin3ss3g0d/cypherhound',
    },
    {
        name: 'BloodHound CLI',
        author: 'dadevel',
        publisher: 'community',
        logoPath: '/img/logo-red-transparent-logo-only.svg',
        description:
            'Utilities for local BHCE projects, data collection, Cypher workflows, enrichment, and interoperability with other tools.',
        href: 'https://github.com/dadevel/bloodhoundcli',
    },
] satisfies readonly GenerallyAvailableCatalogItem[]);
