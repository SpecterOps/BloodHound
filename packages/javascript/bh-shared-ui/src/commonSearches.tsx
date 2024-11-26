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

import { ActiveDirectoryPathfindingEdges, AzurePathfindingEdges } from './graphSchema';

const categoryAD = 'Active Directory';
const categoryAzure = 'Azure';

const azureTransitEdgeTypes = AzurePathfindingEdges().join('|');
const adTransitEdgeTypes = ActiveDirectoryPathfindingEdges().join('|');

const highPrivilegedRoleDisplayNameRegex =
    'Global Administrator.*|User Administrator.*|Cloud Application Administrator.*|Authentication Policy Administrator.*|Exchange Administrator.*|Helpdesk Administrator.*|Privileged Authentication Administrator.*';

export type CommonSearchType = {
    subheader: string;
    category: string;
    queries: {
        description: string;
        cypher: string;
    }[];
};

export const CommonSearches: CommonSearchType[] = [
    {
        subheader: 'Domain Information',
        category: categoryAD,
        queries: [
            {
                description: 'All Domain Admins',
                cypher: `MATCH p=(n:Group)<-[:MemberOf*1..]-(m:Base)\nWHERE n.objectid ENDS WITH '-512'\nRETURN p\nLIMIT 1000`,
            },
            {
                description: 'Map domain trusts',
                cypher: `MATCH p=(n:Domain)-[:TrustedBy]->(m:Domain)\nRETURN p\nLIMIT 1000`,
            },
            {
                description: 'Locations of Tier Zero / High Value objects',
                cypher: `MATCH p = (:Domain)-[:Contains*1..]->(n:Base)\nWHERE 'admin_tier_0' IN split(n.system_tags, ' ')\nRETURN p\nLIMIT 1000`,
            },
            {
                description: 'Map OU structure',
                cypher: `MATCH p = (:Domain)-[:Contains*1..]->(n:OU)\nRETURN p\nLIMIT 1000`,
            },
        ],
    },
    {
        subheader: 'Dangerous Privileges',
        category: categoryAD,
        queries: [
            {
                description: 'Principals with DCSync privileges',
                cypher: `MATCH p=(:Base)-[:DCSync|AllExtendedRights|GenericAll]->(:Domain)\nRETURN p\nLIMIT 1000`,
            },
            {
                description: 'Principals with foreign domain group membership',
                cypher: `MATCH p=(n:Base)-[:MemberOf]->(m:Group)\nWHERE m.domainsid<>n.domainsid\nRETURN p\nLIMIT 1000`,
            },
            {
                description: 'Computers where Domain Users are local administrators',
                cypher: `MATCH p=(m:Group)-[:AdminTo]->(n:Computer)\nWHERE m.objectid ENDS WITH '-513'\nRETURN p\nLIMIT 1000`,
            },
            {
                description: 'Computers where Domain Users can read LAPS passwords',
                cypher: `MATCH p=(m:Group)-[:AllExtendedRights|ReadLAPSPassword]->(n:Computer)\nWHERE m.objectid ENDS WITH '-513'\nRETURN p\nLIMIT 1000`,
            },
            {
                description: 'Paths from Domain Users to Tier Zero / High Value targets',
                cypher: `MATCH p=shortestPath((m:Group)-[:${adTransitEdgeTypes}*1..]->(n))\nWHERE 'admin_tier_0' IN split(n.system_tags, ' ') AND m.objectid ENDS WITH '-513' AND m<>n\nRETURN p\nLIMIT 1000`,
            },
            {
                description: 'Workstations where Domain Users can RDP',
                cypher: `MATCH p=(m:Group)-[:CanRDP]->(c:Computer)\nWHERE m.objectid ENDS WITH '-513' AND NOT toUpper(c.operatingsystem) CONTAINS 'SERVER'\nRETURN p\nLIMIT 1000`,
            },
            {
                description: 'Servers where Domain Users can RDP',
                cypher: `MATCH p=(m:Group)-[:CanRDP]->(c:Computer)\nWHERE m.objectid ENDS WITH '-513' AND toUpper(c.operatingsystem) CONTAINS 'SERVER'\nRETURN p\nLIMIT 1000`,
            },
            {
                description: 'Dangerous privileges for Domain Users groups',
                cypher: `MATCH p=(m:Group)-[:${adTransitEdgeTypes}]->(n:Base)\nWHERE m.objectid ENDS WITH '-513'\nRETURN p\nLIMIT 1000`,
            },
            {
                description: 'Domain Admins logons to non-Domain Controllers',
                cypher: `MATCH (dc)-[r:MemberOf*0..]->(g:Group)\nWHERE g.objectid ENDS WITH '-516'\nWITH COLLECT(dc) AS exclude\nMATCH p = (c:Computer)-[n:HasSession]->(u:User)-[r2:MemberOf*1..]->(g:Group)\nWHERE g.objectid ENDS WITH '-512' AND NOT c IN exclude\nRETURN p\nLIMIT 1000`,
            },
        ],
    },
    {
        subheader: 'Kerberos Interaction',
        category: categoryAD,
        queries: [
            {
                description: 'Kerberoastable members of Tier Zero / High Value groups',
                cypher: `MATCH p=shortestPath((n:User)-[:MemberOf]->(g:Group))\nWHERE 'admin_tier_0' IN split(g.system_tags, ' ') AND n.hasspn=true\nAND n.enabled = true\nAND NOT n.objectid ENDS WITH '-502'\nRETURN p\nLIMIT 1000`,
            },
            {
                description: 'All Kerberoastable users',
                cypher: `MATCH (u:User)\nWHERE u.hasspn=true\nAND u.enabled = true\nAND NOT u.objectid ENDS WITH '-502'\nAND NOT coalesce(u.gmsa, ' ') = true\nAND NOT coalesce(u.msa, ' ') = true\nRETURN u\nLIMIT 100`,
            },
            {
                description: 'Kerberoastable users with most privileges',
                cypher: `MATCH (u:User)\nWHERE u.hasspn = true\nAND u.enabled = true\nAND NOT u.objectid ENDS WITH '-502'\nAND NOT coalesce(u.gmsa, ' ') = true\nAND NOT coalesce(u.msa, ' ') = true\nOPTIONAL MATCH (u)-[:AdminTo]->(c1:Computer)\nOPTIONAL MATCH (u)-[:MemberOf*1..]->(:Group)-[:AdminTo]->(c2:Computer)\nWITH u,COLLECT(c1) + COLLECT(c2) AS tempVar\nUNWIND tempVar AS comps\nRETURN u\nLIMIT 100`,
            },
            {
                description: 'AS-REP Roastable users (DontReqPreAuth)',
                cypher: `MATCH (u:User)\nWHERE u.dontreqpreauth = true\nAND u.enabled = true\nRETURN u\nLIMIT 100`,
            },
        ],
    },
    {
        subheader: 'Shortest Paths',
        category: categoryAD,
        queries: [
            {
                description: 'Shortest paths to systems trusted for unconstrained delegation',
                cypher: `MATCH p=shortestPath((n)-[:${adTransitEdgeTypes}*1..]->(m:Computer))\nWHERE m.unconstraineddelegation = true AND n<>m\nRETURN p\nLIMIT 1000`,
            },
            {
                description: 'Shortest paths to Domain Admins from Kerberoastable users',
                cypher: `MATCH p=shortestPath((n:User)-[:${adTransitEdgeTypes}*1..]->(m:Group))\nWHERE n.hasspn = true AND m.objectid ENDS WITH '-512'\nRETURN p\nLIMIT 1000`,
            },
            {
                description: 'Shortest paths to Tier Zero / High Value targets',
                cypher: `MATCH p=shortestPath((n)-[:${adTransitEdgeTypes}*1..]->(m))\nWHERE 'admin_tier_0' IN split(m.system_tags, ' ') AND n<>m\nRETURN p\nLIMIT 1000`,
            },
            {
                description: 'Shortest paths from Domain Users to Tier Zero / High Value targets',
                cypher: `MATCH p=shortestPath((n:Group)-[:${adTransitEdgeTypes}*1..]->(m))\nWHERE 'admin_tier_0' IN split(m.system_tags, ' ') AND n.objectid ENDS WITH '-513' AND n<>m\nRETURN p\nLIMIT 1000`,
            },
            {
                description: 'Shortest paths to Domain Admins',
                cypher: `MATCH p=shortestPath((n:Base)-[:${adTransitEdgeTypes}*1..]->(g:Group))\nWHERE g.objectid ENDS WITH '-512' AND n<>g\nRETURN p\nLIMIT 1000`,
            },
            {
                description: 'Shortest paths from Owned objects to Tier Zero',
                cypher: `// MANY TO MANY SHORTEST PATH QUERIES USE EXCESSIVE SYSTEM RESOURCES AND TYPICALLY WILL NOT COMPLETE\n// UNCOMMENT THE FOLLOWING LINES BY REMOVING THE DOUBLE FORWARD SLASHES AT YOUR OWN RISK\n// MATCH p=shortestPath((n)-[:${adTransitEdgeTypes}*1..]->(m))\n// WHERE 'admin_tier_0' IN split(m.system_tags, ' ') AND n<>m\n// AND 'owned' IN split(n.system_tags,' ')\n// RETURN p\n// LIMIT 1000`,
            },
            {
                description: 'Shortest paths from Owned objects',
                cypher: `MATCH p=shortestPath((n:Base)-[:${adTransitEdgeTypes}*1..]->(g:Base))\nWHERE 'owned' IN split(n.system_tags,' ') AND n<>g\nRETURN p\nLIMIT 1000`,
            },
        ],
    },
    {
        subheader: 'Active Directory Certificate Services',
        category: categoryAD,
        queries: [
            {
                description: 'PKI hierarchy',
                cypher: `MATCH p=(:Domain)<-[:HostsCAService|IssuedSignedBy|EnterpriseCAFor|RootCAFor|TrustedForNTAuth|NTAuthStoreFor*..]-()\nRETURN p\nLIMIT 1000`,
            },
            {
                description: 'Public Key Services container',
                cypher: `MATCH p = (c:Container)-[:Contains*..]->(:Base)\nWHERE c.distinguishedname starts with 'CN=PUBLIC KEY SERVICES,CN=SERVICES,CN=CONFIGURATION,DC='\nRETURN p\nLIMIT 1000`,
            },
            {
                description: 'Enrollment rights on published certificate templates',
                cypher: `MATCH p = (:Base)-[:Enroll|GenericAll|AllExtendedRights]->(ct:CertTemplate)-[:PublishedTo]->(:EnterpriseCA)\nRETURN p\nLIMIT 1000`,
            },
            {
                description: 'Enrollment rights on published ESC1 certificate templates',
                cypher: `MATCH p = (:Base)-[:Enroll|GenericAll|AllExtendedRights]->(ct:CertTemplate)-[:PublishedTo]->(:EnterpriseCA)\nWHERE ct.enrolleesuppliessubject = True\nAND ct.authenticationenabled = True\nAND ct.requiresmanagerapproval = False\nAND (ct.authorizedsignatures = 0 OR ct.schemaversion = 1)\nRETURN p\nLIMIT 1000`,
            },
            {
                description: 'Enrollment rights on published ESC2 certificate templates',
                cypher: `MATCH p = (:Base)-[:Enroll|GenericAll|AllExtendedRights]->(ct:CertTemplate)-[:PublishedTo]->(:EnterpriseCA)\nWHERE ct.requiresmanagerapproval = False\nAND (ct.effectiveekus = [] OR '2.5.29.37.0' IN ct.effectiveekus)\nAND (ct.authorizedsignatures = 0 OR ct.schemaversion = 1)\nRETURN p\nLIMIT 1000`,
            },
            {
                description: 'Enrollment rights on published enrollment agent certificate templates',
                cypher: `MATCH p = (:Base)-[:Enroll|GenericAll|AllExtendedRights]->(ct:CertTemplate)-[:PublishedTo]->(:EnterpriseCA)\nWHERE '1.3.6.1.4.1.311.20.2.1' IN ct.effectiveekus\nOR '2.5.29.37.0' IN ct.effectiveekus\nOR SIZE(ct.effectiveekus) = 0\nRETURN p\nLIMIT 1000`,
            },
            {
                description: 'Enrollment rights on published certificate templates with no security extension',
                cypher: `MATCH p = (:Base)-[:Enroll|GenericAll|AllExtendedRights]->(ct:CertTemplate)-[:PublishedTo]->(:EnterpriseCA)\nWHERE ct.nosecurityextension = true\nRETURN p\nLIMIT 1000`,
            },
            {
                description:
                    'Enrollment rights on certificate templates published to Enterprise CA with User Specified SAN enabled',
                cypher: `MATCH p = (:Base)-[:Enroll|GenericAll|AllExtendedRights]->(ct:CertTemplate)-[:PublishedTo]->(eca:EnterpriseCA)\nWHERE eca.isuserspecifiessanenabled = True\nRETURN p\nLIMIT 1000`,
            },
            {
                description: 'CA administrators and CA managers',
                cypher: `MATCH p = (:Base)-[:ManageCertificates|ManageCA]->(:EnterpriseCA)\nRETURN p\nLIMIT 1000`,
            },
            {
                description: 'Domain controllers with weak certificate binding enabled',
                cypher: `MATCH p = (dc:Computer)-[:DCFor]->(d:Domain)\nWHERE dc.strongcertificatebindingenforcementraw = 0 OR dc.strongcertificatebindingenforcementraw = 1\nRETURN p\nLIMIT 1000`,
            },
            {
                description: 'Domain controllers with UPN certificate mapping enabled',
                cypher: `MATCH p = (dc:Computer)-[:DCFor]->(d:Domain)\nWHERE dc.certificatemappingmethodsraw IN [4, 5, 6, 7, 12, 13, 14, 15, 20, 21, 22, 23, 28, 29, 30, 31]\nRETURN p\nLIMIT 1000`,
            },
            {
                description: 'Non-default permissions on IssuancePolicy nodes',
                cypher: `MATCH p = (n:Base)-[:GenericAll|GenericWrite|Owns|WriteOwner|WriteDacl]->(:IssuancePolicy)\nWHERE NOT n.objectid ENDS WITH '-512' AND NOT n.objectid ENDS WITH '-519'\nRETURN p\nLIMIT 1000`,
            },
            {
                description: 'Enrollment rights on CertTemplates with OIDGroupLink',
                cypher: `MATCH p = (:Base)-[:Enroll|GenericAll|AllExtendedRights]->(ct:CertTemplate)-[:ExtendedByPolicy]->(:IssuancePolicy)-[:OIDGroupLink]->(g:Group)\nRETURN p\nLIMIT 1000`,
            },
        ],
    },
    {
        subheader: 'Active Directory Hygiene',
        category: categoryAD,
        queries: [
            {
                description: 'Enabled Tier Zero / High Value principals inactive for 60 days',
                cypher: `WITH 60 as inactive_days\nMATCH (n:Base)\nWHERE n.system_tags CONTAINS 'admin_tier_0'\nAND n.enabled = true\nAND n.lastlogontimestamp < (datetime().epochseconds - (inactive_days * 86400)) // Replicated value\nAND n.lastlogon < (datetime().epochseconds - (inactive_days * 86400)) // Non-replicated value\nAND n.whencreated < (datetime().epochseconds - (inactive_days * 86400)) // Exclude recently created principals\nAND NOT n.name STARTS WITH 'AZUREADKERBEROS.' // Removes false positive, Azure KRBTGT\nAND NOT n.objectid ENDS WITH '-500' // Removes false positive, built-in Administrator\nAND NOT n.name STARTS WITH 'AZUREADSSOACC.' // Removes false positive, Entra Seamless SSO\nRETURN n`,
            },
            {
                description: 'Tier Zero / High Value enabled users not requiring smart card authentication',
                cypher: `MATCH (n:User)\nWHERE 'admin_tier_0' IN split(n.system_tags, ' ')\nAND n.enabled = true\nAND n.smartcardrequired = false\nAND NOT n.name STARTS WITH 'MSOL_' // Removes false positive, Entra sync\nAND NOT n.name STARTS WITH 'PROVAGENTGMSA' // Removes false positive, Entra sync\nAND NOT n.name STARTS WITH 'ADSYNCMSA_' // Removes false positive, Entra sync\nRETURN n`,
            },
            {
                description: 'Domains where any user can join a computer to the domain',
                cypher: `MATCH (n:Domain)\nWHERE n.machineaccountquota > 0\nRETURN n`,
            },
            {
                description: 'Domains with smart card accounts where smart account passwords do not expire',
                cypher: `MATCH (n:Domain)-[:Contains*1..]->(m:Base)\nWHERE n.expirepasswordsonsmartcardonlyaccounts = false\nAND m.enabled = true\nAND m.smartcardrequired = true\nRETURN n`,
            },
            {
                description: 'Two-way forest trusts enabled for delegation',
                cypher: `MATCH p=(n:Domain)-[r:TrustedBy]->(m:Domain)\nWHERE (n)<-[:TrustedBy]-(m)\nAND r.trusttype = 'Forest'\nAND r.tgtdelegationenabled = true\nRETURN p`,
            },
            {
                description: 'Computers with unsupported operating systems',
                cypher: `MATCH (n:Computer)\nWHERE n.operatingsystem =~ '(?i).*Windows.* (2000|2003|2008|2012|xp|vista|7|8|me|nt).*'\nRETURN n\nLIMIT 100`,
            },
            {
                description: 'Users which do not require password to authenticate',
                cypher: `MATCH (u:User)\nWHERE u.passwordnotreqd = true\nRETURN u\nLIMIT 100`,
            },
            {
                description: 'Users with passwords not rotated in over 1 year',
                cypher: `WITH 365 as days_since_change\nMATCH (u:User)\nWHERE u.pwdlastset < (datetime().epochseconds - (days_since_change * 86400))\nAND NOT u.pwdlastset IN [-1.0, 0.0]\nRETURN u\nLIMIT 100`,
            },
            {
                description: 'Nested groups within Tier Zero / High Value',
                cypher: `MATCH p=(n:Group)-[:MemberOf*..]->(t:Group)\nWHERE coalesce(t.system_tags,'') CONTAINS ('tier_0')\nAND NOT n.objectid ENDS WITH '-512' // Domain Admins\nAND NOT n.objectid ENDS WITH '-519' // Enterprise Admins\nRETURN p\nLIMIT 1000`,
            },
            {
                description: 'Disabled Tier Zero / High Value principals',
                cypher: `MATCH (n:Base)\nWHERE n.system_tags CONTAINS 'admin_tier_0'\nAND n.enabled = false\nAND NOT n.objectid ENDS WITH '-502' // Removes false positive, KRBTGT\nAND NOT n.objectid ENDS WITH '-500' // Removes false positive, built-in Administrator\nRETURN n\nLIMIT 100`,
            },
            {
                description: 'Principals with passwords stored using reversible encryption',
                cypher: `MATCH (n:Base)\nWHERE n.encryptedtextpwdallowed = true\nRETURN n`,
            },
            {
                description: 'Principals with DES-only Kerberos authentication',
                cypher: `MATCH (n:Base)\nWHERE n.enabled = true\nAND n.usedeskeyonly = true\nRETURN n`,
            },
            {
                description: 'Principals with weak supported Kerberos encryption types',
                cypher: `MATCH (n:Base)\nWHERE ANY(keyword IN n.supportedencryptiontypes WHERE keyword IN ['DES-CBC-CRC', 'DES-CBC-MD5', 'RC4-HMAC-MD5'])\nRETURN n`,
            },
            {
                description: 'Tier Zero / High Value users with non-expiring passwords',
                cypher: `MATCH (u:User)\nWHERE u.enabled = true\nAND u.pwdneverexpires = true\nand u.system_tags CONTAINS 'admin_tier_0'\nRETURN u\nLIMIT 100`,
            },
        ],
    },
    {
        subheader: 'General',
        category: categoryAzure,
        queries: [
            {
                description: 'All Global Administrators',
                cypher: 'MATCH p = (n:AZBase)-[r:AZGlobalAdmin*1..]->(m:AZTenant)\nRETURN p\nLIMIT 1000',
            },
            {
                description: 'All members of high privileged roles',
                cypher: `MATCH p=(n:AZBase)-[:AZHasRole|AZMemberOf*1..2]->(r:AZRole)\nWHERE r.name =~ '(?i)${highPrivilegedRoleDisplayNameRegex}'\nRETURN p\nLIMIT 1000`,
            },
        ],
    },
    {
        subheader: 'Shortest Paths',
        category: categoryAzure,
        queries: [
            {
                description: 'Shortest paths from Entra Users to Tier Zero / High Value targets',
                cypher: `MATCH p=shortestPath((m:AZUser)-[r:${azureTransitEdgeTypes}*1..]->(n:AZBase))\nWHERE 'admin_tier_0' IN split(n.system_tags, ' ') AND n.name =~ '(?i)${highPrivilegedRoleDisplayNameRegex}' AND m<>n\nRETURN p\nLIMIT 1000`,
            },
            {
                description: 'Shortest paths to privileged roles',
                cypher: `MATCH p=shortestPath((m:AZBase)-[r:${azureTransitEdgeTypes}*1..]->(n:AZRole))\nWHERE n.name =~ '(?i)${highPrivilegedRoleDisplayNameRegex}' AND m<>n\nRETURN p\nLIMIT 1000`,
            },
            {
                description: 'Shortest paths from Azure Applications to Tier Zero / High Value targets',
                cypher: `MATCH p=shortestPath((m:AZApp)-[r:${azureTransitEdgeTypes}*1..]->(n:AZBase))\nWHERE 'admin_tier_0' IN split(n.system_tags, ' ') AND m<>n\nRETURN p\nLIMIT 1000`,
            },
            {
                description: 'Shortest paths to Azure Subscriptions',
                cypher: `MATCH p=shortestPath((m:AZBase)-[r:${azureTransitEdgeTypes}*1..]->(n:AZSubscription))\nWHERE m<>n\nRETURN p\nLIMIT 1000`,
            },
        ],
    },
    {
        subheader: 'Microsoft Graph',
        category: categoryAzure,
        queries: [
            {
                description: 'All service principals with Microsoft Graph privilege to grant arbitrary App Roles',
                cypher: 'MATCH p=(n:AZServicePrincipal)-[r:AZMGGrantAppRoles]->(o:AZTenant)\nRETURN p\nLIMIT 1000',
            },
            {
                description: 'All service principals with Microsoft Graph App Role assignments',
                cypher: 'MATCH p=(m:AZServicePrincipal)-[r:AZMGAppRoleAssignment_ReadWrite_All|AZMGApplication_ReadWrite_All|AZMGDirectory_ReadWrite_All|AZMGGroupMember_ReadWrite_All|AZMGGroup_ReadWrite_All|AZMGRoleManagement_ReadWrite_Directory|AZMGServicePrincipalEndpoint_ReadWrite_All]->(n:AZServicePrincipal)\nRETURN p\nLIMIT 1000',
            },
        ],
    },
    {
        subheader: 'Azure Hygiene',
        category: categoryAzure,
        queries: [
            {
                description: 'Foreign principals in Tier Zero / High Value targets',
                cypher: `MATCH (n:AZServicePrincipal)\nWHERE n.system_tags contains 'admin_tier_0'\nAND NOT toUpper(n.appownerorganizationid) = toUpper(n.tenantid)\nAND n.appownerorganizationid CONTAINS '-'\nRETURN n\nLIMIT 100`,
            },
            {
                description: 'Tier Zero AD principals synchronized with Entra ID',
                cypher: `MATCH (ENTRA:AZBase)\nMATCH (AD:Base)\nWHERE ENTRA.onpremsyncenabled = true\nAND ENTRA.onpremid = AD.objectid\nAND AD.system_tags CONTAINS 'admin_tier_0'\nRETURN ENTRA\n// Replace 'RETURN ENTRA' with 'RETURN AD' to see the corresponding AD principals\nLIMIT 100`,
            },
            {
                description: 'Tier Zero / High Value external Entra ID users',
                cypher: `MATCH (n:AZUser)\nWHERE n.system_tags contains 'admin_tier_0'\nAND n.name CONTAINS '#EXT#@'\nRETURN n\nLIMIT 100`,
            },
            {
                description: 'Disabled Tier Zero / High Value principals',
                cypher: `MATCH (n:AZBase)\nWHERE n.system_tags CONTAINS 'admin_tier_0'\nAND n.enabled = false\nRETURN n\nLIMIT 100`,
            },
            {
                description: 'Devices with unsupported operating systems',
                cypher: `MATCH (n:AZDevice)\nWHERE n.operatingsystem CONTAINS 'WINDOWS'\nAND n.operatingsystemversion =~ '(10.0.19044|10.0.22000|10.0.19043|10.0.19042|10.0.19041|10.0.18363|10.0.18362|10.0.17763|10.0.17134|10.0.16299|10.0.15063|10.0.14393|10.0.10586|10.0.10240|6.3.9600|6.2.9200|6.1.7601|6.0.6200|5.1.2600|6.0.6003|5.2.3790|5.0.2195).?.*'\nRETURN n\nLIMIT 100`,
            },
        ],
    },
    {
        subheader: 'Cross Platform Attack Paths',
        category: categoryAzure,
        queries: [
            {
                description: 'Entra Users synced from On-Prem Users added to Domain Admins group',
                cypher: `MATCH p = (:AZUser)-[:SyncedToADUser]->(:User)-[:MemberOf]->(g:Group)\nWHERE g.objectid ENDS WITH '-512'\nRETURN p\nLIMIT 1000`,
            },
            {
                description: 'On-Prem Users synced to Entra Users with Entra Admin Roles (direct)',
                cypher: 'MATCH p = (:User)-[:SyncedToEntraUser]->(:AZUser)-[:AZHasRole]->(:AZRole)\nRETURN p\nLIMIT 1000',
            },
            {
                description: 'On-Prem Users synced to Entra Users with Entra Admin Roles (group delegated)',
                cypher: 'MATCH p = (:User)-[:SyncedToEntraUser]->(:AZUser)-[:AZMemberOf]->(:AZGroup)-[:AZHasRole]->(:AZRole)\nRETURN p\nLIMIT 1000',
            },
            {
                description: 'On-Prem Users synced to Entra Users with Azure RM Roles (direct)',
                cypher: 'MATCH p = (:User)-[:SyncedToEntraUser]->(:AZUser)-[:AZOwner|AZUserAccessAdministrator|AZGetCertificates|AZGetKeys|AZGetSecrets|AZAvereContributor|AZKeyVaultContributor|AZContributor|AZVMAdminLogin|AZVMContributor|AZAKSContributor|AZAutmomationContributor|AZLogicAppContributor|AZWebsiteContributor]->(:AZBase)\nRETURN p\nLIMIT 1000',
            },
            {
                description: 'On-Prem Users synced to Entra Users with Azure RM Roles (group delegated)',
                cypher: 'MATCH p = (:User)-[:SyncedToEntraUser]->(:AZUser)-[:AZMemberOf]->(:AZGroup)-[:AZOwner|AZUserAccessAdministrator|AZGetCertificates|AZGetKeys|AZGetSecrets|AZAvereContributor|AZKeyVaultContributor|AZContributor|AZVMAdminLogin|AZVMContributor|AZAKSContributor|AZAutmomationContributor|AZLogicAppContributor|AZWebsiteContributor]->(:AZBase)\nRETURN p\nLIMIT 1000',
            },
            {
                description: 'On-Prem Users synced to Entra Users that Own Entra Objects',
                cypher: 'MATCH p = (:User)-[:SyncedToEntraUser]->(:AZUser)-[:AZOwns]->(:AZBase)\nRETURN p\nLIMIT 1000',
            },
            {
                description: 'On-Prem Users synced to Entra Users with Entra Group Membership',
                cypher: 'MATCH p = (:User)-[:SyncedToEntraUser]->(:AZUser)-[:AZMemberOf]->(:AZGroup)\nRETURN p\nLIMIT 1000',
            },
        ],
    },
];
