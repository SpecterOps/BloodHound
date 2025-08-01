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

import { OWNED_OBJECT_TAG, TIER_ZERO_TAG } from './constants';
import { ActiveDirectoryPathfindingEdges, AzurePathfindingEdges } from './graphSchema';
import { CommonSearchType } from './types';

const categoryAD = 'Active Directory';
const categoryAzure = 'Azure';

const azureTransitEdgeTypes = AzurePathfindingEdges().join('|');
const adTransitEdgeTypes = ActiveDirectoryPathfindingEdges().join('|');

const highPrivilegedRoleDisplayNameRegex =
    '^(Global Administrator|User Administrator|Cloud Application Administrator|Authentication Policy Administrator|Exchange Administrator|Helpdesk Administrator|Privileged Authentication Administrator).*$';

/*
    NOTE: temporarily there exists 2 common searches files, edits here should be reflected in ./commonSearchesAGT.ts as well
 */

export const CommonSearches: CommonSearchType[] = [
    {
        subheader: 'Domain Information',
        category: categoryAD,
        queries: [
            {
                description: 'All Domain Admins',
                cypher: `MATCH p = (t:Group)<-[:MemberOf*1..]-(a)\nWHERE (a:User or a:Computer) and t.objectid ENDS WITH '-512'\nRETURN p\nLIMIT 1000`,
            },
            {
                description: 'Map domain trusts',
                cypher: `MATCH p = (:Domain)-[:SameForestTrust|CrossForestTrust]->(:Domain)\nRETURN p\nLIMIT 1000`,
            },
            {
                description: 'Locations of Tier Zero / High Value objects',
                cypher: `MATCH p = (t:Base)<-[:Contains*1..]-(:Domain)\nWHERE COALESCE(t.system_tags, '') CONTAINS '${TIER_ZERO_TAG}'\nRETURN p\nLIMIT 1000`,
            },
            {
                description: 'Map OU structure',
                cypher: `MATCH p = (:Domain)-[:Contains*1..]->(:OU)\nRETURN p\nLIMIT 1000`,
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
                cypher: `MATCH p=(s:Base)-[:MemberOf]->(t:Group)\nWHERE s.domainsid<>t.domainsid\nRETURN p\nLIMIT 1000`,
            },
            {
                description: 'Computers where Domain Users are local administrators',
                cypher: `MATCH p=(s:Group)-[:AdminTo]->(:Computer)\nWHERE s.objectid ENDS WITH '-513'\nRETURN p\nLIMIT 1000`,
            },
            {
                description: 'Computers where Domain Users can read LAPS passwords',
                cypher: `MATCH p=(s:Group)-[:AllExtendedRights|ReadLAPSPassword]->(:Computer)\nWHERE s.objectid ENDS WITH '-513'\nRETURN p\nLIMIT 1000`,
            },
            {
                description: 'Paths from Domain Users to Tier Zero / High Value targets',
                cypher: `MATCH p=shortestPath((s:Group)-[:${adTransitEdgeTypes}*1..]->(t))\nWHERE COALESCE(t.system_tags, '') CONTAINS '${TIER_ZERO_TAG}' AND s.objectid ENDS WITH '-513' AND s<>t\nRETURN p\nLIMIT 1000`,
            },
            {
                description: 'Workstations where Domain Users can RDP',
                cypher: `MATCH p=(s:Group)-[:CanRDP]->(t:Computer)\nWHERE s.objectid ENDS WITH '-513' AND NOT toUpper(t.operatingsystem) CONTAINS 'SERVER'\nRETURN p\nLIMIT 1000`,
            },
            {
                description: 'Servers where Domain Users can RDP',
                cypher: `MATCH p=(s:Group)-[:CanRDP]->(t:Computer)\nWHERE s.objectid ENDS WITH '-513' AND toUpper(t.operatingsystem) CONTAINS 'SERVER'\nRETURN p\nLIMIT 1000`,
            },
            {
                description: 'Dangerous privileges for Domain Users groups',
                cypher: `MATCH p=(s:Group)-[:${adTransitEdgeTypes}]->(:Base)\nWHERE s.objectid ENDS WITH '-513'\nRETURN p\nLIMIT 1000`,
            },
            {
                description: 'Domain Admins logons to non-Domain Controllers',
                cypher: `MATCH (s)-[:MemberOf*0..]->(g:Group)\nWHERE g.objectid ENDS WITH '-516'\nWITH COLLECT(s) AS exclude\nMATCH p = (c:Computer)-[:HasSession]->(:User)-[:MemberOf*1..]->(g:Group)\nWHERE g.objectid ENDS WITH '-512' AND NOT c IN exclude\nRETURN p\nLIMIT 1000`,
            },
        ],
    },
    {
        subheader: 'Kerberos Interaction',
        category: categoryAD,
        queries: [
            {
                description: 'Kerberoastable members of Tier Zero / High Value groups',
                cypher: `MATCH (u:User)\nWHERE u.hasspn=true\nAND u.enabled = true\nAND NOT u.objectid ENDS WITH '-502'\nAND NOT COALESCE(u.gmsa, false) = true\nAND NOT COALESCE(u.msa, false) = true\nAND COALESCE(u.system_tags, '') CONTAINS '${TIER_ZERO_TAG}'\nRETURN u\nLIMIT 100`,
            },
            {
                description: 'All Kerberoastable users',
                cypher: `MATCH (u:User)\nWHERE u.hasspn=true\nAND u.enabled = true\nAND NOT u.objectid ENDS WITH '-502'\nAND NOT COALESCE(u.gmsa, false) = true\nAND NOT COALESCE(u.msa, false) = true\nRETURN u\nLIMIT 100`,
            },
            {
                description: 'Kerberoastable users with most admin privileges',
                cypher: `MATCH (u:User)\nWHERE u.hasspn = true\n  AND u.enabled = true\n  AND NOT u.objectid ENDS WITH '-502'\n  AND NOT COALESCE(u.gmsa, false) = true\n  AND NOT COALESCE(u.msa, false) = true\nMATCH (u)-[:MemberOf|AdminTo*1..]->(c:Computer)\nWITH DISTINCT u, COUNT(c) AS adminCount\nRETURN u\nORDER BY adminCount DESC\nLIMIT 100`,
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
                cypher: `MATCH p=shortestPath((s)-[:${adTransitEdgeTypes}*1..]->(t:Computer))\nWHERE t.unconstraineddelegation = true AND s<>t\nRETURN p\nLIMIT 1000`,
            },
            {
                description: 'Shortest paths to Domain Admins from Kerberoastable users',
                cypher: `MATCH p=shortestPath((s:User)-[:${adTransitEdgeTypes}*1..]->(t:Group))\nWHERE s.hasspn=true\nAND s.enabled = true\nAND NOT s.objectid ENDS WITH '-502'\nAND NOT COALESCE(s.gmsa, false) = true\nAND NOT COALESCE(s.msa, false) = true\nAND t.objectid ENDS WITH '-512'\nRETURN p\nLIMIT 1000`,
            },
            {
                description: 'Shortest paths to Tier Zero / High Value targets',
                cypher: `MATCH p=shortestPath((s)-[:${adTransitEdgeTypes}*1..]->(t))\nWHERE COALESCE(t.system_tags, '') CONTAINS '${TIER_ZERO_TAG}' AND s<>t\nRETURN p\nLIMIT 1000`,
            },
            {
                description: 'Shortest paths from Domain Users to Tier Zero / High Value targets',
                cypher: `MATCH p=shortestPath((s:Group)-[:${adTransitEdgeTypes}*1..]->(t))\nWHERE COALESCE(t.system_tags, '') CONTAINS '${TIER_ZERO_TAG}' AND s.objectid ENDS WITH '-513' AND s<>t\nRETURN p\nLIMIT 1000`,
            },
            {
                description: 'Shortest paths to Domain Admins',
                cypher: `MATCH p=shortestPath((t:Group)<-[:${adTransitEdgeTypes}*1..]-(s:Base))\nWHERE t.objectid ENDS WITH '-512' AND s<>t\nRETURN p\nLIMIT 1000`,
            },
            {
                description: 'Shortest paths from Owned objects to Tier Zero',
                cypher: `// MANY TO MANY SHORTEST PATH QUERIES USE EXCESSIVE SYSTEM RESOURCES AND TYPICALLY WILL NOT COMPLETE\n// UNCOMMENT THE FOLLOWING LINES BY REMOVING THE DOUBLE FORWARD SLASHES AT YOUR OWN RISK\n// MATCH p=shortestPath((s)-[:${adTransitEdgeTypes}*1..]->(t))\n// WHERE COALESCE(t.system_tags, '') CONTAINS '${TIER_ZERO_TAG}' AND s<>t\n// AND COALESCE(s.system_tags, '') CONTAINS '${OWNED_OBJECT_TAG}'\n// RETURN p\n// LIMIT 1000`,
            },
            {
                description: 'Shortest paths from Owned objects',
                cypher: `MATCH p=shortestPath((s:Base)-[:${adTransitEdgeTypes}*1..]->(t:Base))\nWHERE COALESCE(s.system_tags, '') CONTAINS '${OWNED_OBJECT_TAG}' AND s<>t\nRETURN p\nLIMIT 1000`,
            },
        ],
    },
    {
        subheader: 'Active Directory Certificate Services',
        category: categoryAD,
        queries: [
            {
                description: 'PKI hierarchy',
                cypher: `MATCH p=()-[:HostsCAService|IssuedSignedBy|EnterpriseCAFor|RootCAFor|TrustedForNTAuth|NTAuthStoreFor*..]->(:Domain)\nRETURN p\nLIMIT 1000`,
            },
            {
                description: 'Public Key Services container',
                cypher: `MATCH p = (c:Container)-[:Contains*..]->(:Base)\nWHERE c.distinguishedname starts with 'CN=PUBLIC KEY SERVICES,CN=SERVICES,CN=CONFIGURATION,DC='\nRETURN p\nLIMIT 1000`,
            },
            {
                description: 'Enrollment rights on published certificate templates',
                cypher: `MATCH p = (:Base)-[:Enroll|GenericAll|AllExtendedRights]->(:CertTemplate)-[:PublishedTo]->(:EnterpriseCA)\nRETURN p\nLIMIT 1000`,
            },
            {
                description: 'Enrollment rights on published ESC1 certificate templates',
                cypher: `MATCH p = (:Base)-[:Enroll|GenericAll|AllExtendedRights]->(ct:CertTemplate)-[:PublishedTo]->(:EnterpriseCA)\nWHERE ct.enrolleesuppliessubject = True\nAND ct.authenticationenabled = True\nAND ct.requiresmanagerapproval = False\nAND (ct.authorizedsignatures = 0 OR ct.schemaversion = 1)\nRETURN p\nLIMIT 1000`,
            },
            {
                description: 'Enrollment rights on published ESC2 certificate templates',
                cypher: `MATCH p = (:Base)-[:Enroll|GenericAll|AllExtendedRights]->(c:CertTemplate)-[:PublishedTo]->(:EnterpriseCA)\nWHERE c.requiresmanagerapproval = false\nAND (c.effectiveekus = [''] OR '2.5.29.37.0' IN c.effectiveekus)\nAND (c.authorizedsignatures = 0 OR c.schemaversion = 1)\nRETURN p\nLIMIT 1000`,
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
                cypher: `MATCH p = (s:Computer)-[:DCFor]->(:Domain)\nWHERE s.strongcertificatebindingenforcementraw = 0 OR s.strongcertificatebindingenforcementraw = 1\nRETURN p\nLIMIT 1000`,
            },
            {
                description: 'Domain controllers with UPN certificate mapping enabled',
                cypher: `MATCH p = (s:Computer)-[:DCFor]->(:Domain)\nWHERE s.certificatemappingmethodsraw IN [4, 5, 6, 7, 12, 13, 14, 15, 20, 21, 22, 23, 28, 29, 30, 31]\nRETURN p\nLIMIT 1000`,
            },
            {
                description: 'Non-default permissions on IssuancePolicy nodes',
                cypher: `MATCH p = (s:Base)-[:GenericAll|GenericWrite|Owns|WriteOwner|WriteDacl]->(:IssuancePolicy)\nWHERE NOT s.objectid ENDS WITH '-512' AND NOT s.objectid ENDS WITH '-519'\nRETURN p\nLIMIT 1000`,
            },
            {
                description: 'Enrollment rights on CertTemplates with OIDGroupLink',
                cypher: `MATCH p = (:Base)-[:Enroll|GenericAll|AllExtendedRights]->(:CertTemplate)-[:ExtendedByPolicy]->(:IssuancePolicy)-[:OIDGroupLink]->(:Group)\nRETURN p\nLIMIT 1000`,
            },
        ],
    },
    {
        subheader: 'Active Directory Hygiene',
        category: categoryAD,
        queries: [
            {
                description: 'Enabled Tier Zero / High Value principals inactive for 60 days',
                cypher: `WITH 60 as inactive_days\nMATCH (n:Base)\nWHERE COALESCE(n.system_tags, '') CONTAINS '${TIER_ZERO_TAG}'\nAND n.enabled = true\nAND n.lastlogontimestamp < (datetime().epochseconds - (inactive_days * 86400)) // Replicated value\nAND n.lastlogon < (datetime().epochseconds - (inactive_days * 86400)) // Non-replicated value\nAND n.whencreated < (datetime().epochseconds - (inactive_days * 86400)) // Exclude recently created principals\nAND NOT n.name STARTS WITH 'AZUREADKERBEROS.' // Removes false positive, Azure KRBTGT\nAND NOT n.objectid ENDS WITH '-500' // Removes false positive, built-in Administrator\nAND NOT n.name STARTS WITH 'AZUREADSSOACC.' // Removes false positive, Entra Seamless SSO\nRETURN n`,
            },
            {
                description: 'Tier Zero / High Value enabled users not requiring smart card authentication',
                cypher: `MATCH (u:User)\nWHERE COALESCE(u.system_tags, '') CONTAINS '${TIER_ZERO_TAG}'\nAND u.enabled = true\nAND u.smartcardrequired = false\nAND NOT u.name STARTS WITH 'MSOL_' // Removes false positive, Entra sync\nAND NOT u.name STARTS WITH 'PROVAGENTGMSA' // Removes false positive, Entra sync\nAND NOT u.name STARTS WITH 'ADSYNCMSA_' // Removes false positive, Entra sync\nRETURN u`,
            },
            {
                description: 'Domains where any user can join a computer to the domain',
                cypher: `MATCH (d:Domain)\nWHERE d.machineaccountquota > 0\nRETURN d`,
            },
            {
                description: 'Domains with smart card accounts where smart account passwords do not expire',
                cypher: `MATCH (s:Domain)-[:Contains*1..]->(t:Base)\nWHERE s.expirepasswordsonsmartcardonlyaccounts = false\nAND t.enabled = true\nAND t.smartcardrequired = true\nRETURN s`,
            },
            {
                description: 'Cross-forest trusts with abusable configuration',
                cypher: `MATCH p=(n:Domain)-[:CrossForestTrust|SpoofSIDHistory|AbuseTGTDelegation]-(m:Domain)\nWHERE (n)-[:SpoofSIDHistory|AbuseTGTDelegation]-(m)\nRETURN p`,
            },
            {
                description: 'Computers with unsupported operating systems',
                cypher: `MATCH (c:Computer)\nWHERE c.operatingsystem =~ '(?i).*Windows.* (2000|2003|2008|2012|xp|vista|7|8|me|nt).*'\nRETURN c\nLIMIT 100`,
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
                cypher: `MATCH p=(t:Group)<-[:MemberOf*..]-(s:Group)\nWHERE COALESCE(t.system_tags, '') CONTAINS '${TIER_ZERO_TAG}'\nAND NOT s.objectid ENDS WITH '-512' // Domain Admins\nAND NOT s.objectid ENDS WITH '-519' // Enterprise Admins\nRETURN p\nLIMIT 1000`,
            },
            {
                description: 'Disabled Tier Zero / High Value principals',
                cypher: `MATCH (n:Base)\nWHERE COALESCE(n.system_tags, '') CONTAINS '${TIER_ZERO_TAG}'\nAND n.enabled = false\nAND NOT n.objectid ENDS WITH '-502' // Removes false positive, KRBTGT\nAND NOT n.objectid ENDS WITH '-500' // Removes false positive, built-in Administrator\nRETURN n\nLIMIT 100`,
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
                cypher: `MATCH (u:Base)\nWHERE 'DES-CBC-CRC' IN u.supportedencryptiontypes\nOR 'DES-CBC-MD5' IN u.supportedencryptiontypes\nOR 'RC4-HMAC-MD5' IN u.supportedencryptiontypes\nRETURN u`,
            },
            {
                description: 'Tier Zero / High Value users with non-expiring passwords',
                cypher: `MATCH (u:User)\nWHERE u.enabled = true\nAND u.pwdneverexpires = true\nand COALESCE(u.system_tags, '') CONTAINS '${TIER_ZERO_TAG}'\nRETURN u\nLIMIT 100`,
            },
            {
                description: 'Tier Zero principals without AdminSDHolder protection',
                cypher: `MATCH (n:Base)\nWHERE COALESCE(n.system_tags, '') CONTAINS '${TIER_ZERO_TAG}'\nAND n.adminsdholderprotected = false\nRETURN n\nLIMIT 500`,
            },
        ],
    },
    {
        subheader: 'General',
        category: categoryAzure,
        queries: [
            {
                description: 'All Global Administrators',
                cypher: `MATCH p=(:AZBase)-[:AZHasRole*1..]->(t:AZRole)\nWHERE t.name =~ '(?i)Global Administrator.*'\nRETURN p\nLIMIT 1000`,
            },
            {
                description: 'All members of high privileged roles',
                cypher: `MATCH p=(t:AZRole)<-[:AZHasRole|AZMemberOf*1..2]-(:AZBase)\nWHERE t.name =~ '(?i)${highPrivilegedRoleDisplayNameRegex}'\nRETURN p\nLIMIT 1000`,
            },
            {
                description: 'Entra Users with Entra Admin Role direct eligibility',
                cypher: `MATCH p = (:AZUser)-[:AZRoleEligible]->(:AZRole)\nRETURN p LIMIT 100`,
            },
            {
                description: 'Entra Users with Entra Admin Roles group delegated eligibility',
                cypher: `MATCH p = (:AZUser)-[:AZMemberOf]->(:AZGroup)-[:AZRoleEligible]->(:AZRole)\nRETURN p LIMIT 100`,
            },
            {
                description: 'Entra Users with Entra Admin Role approval (direct)',
                cypher: `MATCH p = (:AZUser)-[:AZRoleApprover]->(:AZRole)\nRETURN p LIMIT 100`,
            },
            {
                description: 'Entra Users with Entra Admin Role approval (group delegated)',
                cypher: `MATCH p = (:AZUser)-[:AZMemberOf]->(:AZGroup)-[:AZRoleApprover]->(:AZRole)\nRETURN p LIMIT 100`,
            },
        ],
    },
    {
        subheader: 'Shortest Paths',
        category: categoryAzure,
        queries: [
            {
                description: 'Shortest paths from Entra Users to Tier Zero / High Value targets',
                cypher: `MATCH p=shortestPath((s:AZUser)-[:${azureTransitEdgeTypes}*1..]->(t:AZBase))\nWHERE COALESCE(t.system_tags, '') CONTAINS '${TIER_ZERO_TAG}' AND t.name =~ '(?i)${highPrivilegedRoleDisplayNameRegex}' AND s<>t\nRETURN p\nLIMIT 1000`,
            },
            {
                description: 'Shortest paths to privileged roles',
                cypher: `MATCH p=shortestPath((s:AZBase)-[:${azureTransitEdgeTypes}*1..]->(t:AZRole))\nWHERE t.name =~ '(?i)${highPrivilegedRoleDisplayNameRegex}' AND s<>t\nRETURN p\nLIMIT 1000`,
            },
            {
                description: 'Shortest paths from Azure Applications to Tier Zero / High Value targets',
                cypher: `MATCH p=shortestPath((s:AZApp)-[:${azureTransitEdgeTypes}*1..]->(t:AZBase))\nWHERE COALESCE(t.system_tags, '') CONTAINS '${TIER_ZERO_TAG}' AND s<>t\nRETURN p\nLIMIT 1000`,
            },
            {
                description: 'Shortest paths to Azure Subscriptions',
                cypher: `MATCH p=shortestPath((s:AZBase)-[:${azureTransitEdgeTypes}*1..]->(t:AZSubscription))\nWHERE s<>t\nRETURN p\nLIMIT 1000`,
            },
        ],
    },
    {
        subheader: 'Microsoft Graph',
        category: categoryAzure,
        queries: [
            {
                description: 'All service principals with Microsoft Graph privilege to grant arbitrary App Roles',
                cypher: `MATCH p=(:AZServicePrincipal)-[:AZMGGrantAppRoles]->(:AZTenant)\nRETURN p\nLIMIT 1000`,
            },
            {
                description: 'All service principals with Microsoft Graph App Role assignments',
                cypher: `MATCH p=(:AZServicePrincipal)-[:AZMGAppRoleAssignment_ReadWrite_All|AZMGApplication_ReadWrite_All|AZMGDirectory_ReadWrite_All|AZMGGroupMember_ReadWrite_All|AZMGGroup_ReadWrite_All|AZMGRoleManagement_ReadWrite_Directory|AZMGServicePrincipalEndpoint_ReadWrite_All]->(:AZServicePrincipal)\nRETURN p\nLIMIT 1000`,
            },
        ],
    },
    {
        subheader: 'Azure Hygiene',
        category: categoryAzure,
        queries: [
            {
                description: 'Foreign principals in Tier Zero / High Value targets',
                cypher: `MATCH (n:AZServicePrincipal)\nWHERE COALESCE(n.system_tags, '') CONTAINS '${TIER_ZERO_TAG}'\nAND NOT toUpper(n.appownerorganizationid) = toUpper(n.tenantid)\nAND n.appownerorganizationid CONTAINS '-'\nRETURN n\nLIMIT 100`,
            },
            {
                description: 'Tier Zero AD principals synchronized with Entra ID',
                cypher: `MATCH (ENTRA:AZBase)\nMATCH (AD:Base)\nWHERE ENTRA.onpremsyncenabled = true\nAND ENTRA.onpremid = AD.objectid\nAND COALESCE(AD.system_tags, '') CONTAINS '${TIER_ZERO_TAG}'\nRETURN ENTRA\n// Replace 'RETURN ENTRA' with 'RETURN AD' to see the corresponding AD principals\nLIMIT 100`,
            },
            {
                description: 'Tier Zero / High Value external Entra ID users',
                cypher: `MATCH (n:AZUser)\nWHERE COALESCE(n.system_tags, '') CONTAINS '${TIER_ZERO_TAG}'\nAND n.name CONTAINS '#EXT#@'\nRETURN n\nLIMIT 100`,
            },
            {
                description: 'Disabled Tier Zero / High Value principals',
                cypher: `MATCH (n:AZBase)\nWHERE COALESCE(n.system_tags, '') CONTAINS '${TIER_ZERO_TAG}'\nAND n.enabled = false\nRETURN n\nLIMIT 100`,
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
                cypher: `MATCH p = (:AZUser)-[:SyncedToADUser]->(:User)-[:MemberOf]->(t:Group)\nWHERE t.objectid ENDS WITH '-512'\nRETURN p\nLIMIT 1000`,
            },
            {
                description: 'On-Prem Users synced to Entra Users with Entra Admin Roles (direct)',
                cypher: `MATCH p = (:User)-[:SyncedToEntraUser]->(:AZUser)-[:AZHasRole]->(:AZRole)\nRETURN p\nLIMIT 1000`,
            },
            {
                description: 'On-Prem Users synced to Entra Users with Entra Admin Roles (group delegated)',
                cypher: `MATCH p = (:User)-[:SyncedToEntraUser]->(:AZUser)-[:AZMemberOf]->(:AZGroup)-[:AZHasRole]->(:AZRole)\nRETURN p\nLIMIT 1000`,
            },
            {
                description: 'On-Prem Users synced to Entra Users with Azure RM Roles (direct)',
                cypher: `MATCH p = (:User)-[:SyncedToEntraUser]->(:AZUser)-[:AZOwner|AZUserAccessAdministrator|AZGetCertificates|AZGetKeys|AZGetSecrets|AZAvereContributor|AZKeyVaultContributor|AZContributor|AZVMAdminLogin|AZVMContributor|AZAKSContributor|AZAutomationContributor|AZLogicAppContributor|AZWebsiteContributor]->(:AZBase)\nRETURN p\nLIMIT 1000`,
            },
            {
                description: 'On-Prem Users synced to Entra Users with Azure RM Roles (group delegated)',
                cypher: `MATCH p = (:User)-[:SyncedToEntraUser]->(:AZUser)-[:AZMemberOf]->(:AZGroup)-[:AZOwner|AZUserAccessAdministrator|AZGetCertificates|AZGetKeys|AZGetSecrets|AZAvereContributor|AZKeyVaultContributor|AZContributor|AZVMAdminLogin|AZVMContributor|AZAKSContributor|AZAutomationContributor|AZLogicAppContributor|AZWebsiteContributor]->(:AZBase)\nRETURN p\nLIMIT 1000`,
            },
            {
                description: 'On-Prem Users synced to Entra Users that Own Entra Objects',
                cypher: `MATCH p = (:User)-[:SyncedToEntraUser]->(:AZUser)-[:AZOwns]->(:AZBase)\nRETURN p\nLIMIT 1000`,
            },
            {
                description: 'On-Prem Users synced to Entra Users with Entra Group Membership',
                cypher: `MATCH p = (:User)-[:SyncedToEntraUser]->(:AZUser)-[:AZMemberOf]->(:AZGroup)\nRETURN p\nLIMIT 1000`,
            },
            {
                description: 'Synced Entra Users with Entra Admin Role direct eligibility',
                cypher: `MATCH p = (:User)-[:SyncedToEntraUser]->(:AZUser)-[:AZRoleEligible]->(:AZRole)\nRETURN p LIMIT 100`,
            },
            {
                description: 'Synced Entra Users with Entra Admin Roles group delegated eligibility',
                cypher: `MATCH p = (:User)-[:SyncedToEntraUser]->(:AZUser)-[:AZMemberOf]->(:AZGroup)-[:AZRoleEligible]->(:AZRole)\nRETURN p LIMIT 100`,
            },
            {
                description: 'Synced Entra Users with Entra Admin Role approval (direct)',
                cypher: `MATCH p = (:User)-[:SyncedToEntraUser]->(:AZUser)-[:AZRoleApprover]->(:AZRole)\nRETURN p LIMIT 100`,
            },
            {
                description: 'Synced Entra Users with Entra Admin Role approval (group delegated)',
                cypher: `MATCH p = (:User)-[:SyncedToEntraUser]->(:AZUser)-[:AZMemberOf]->(:AZGroup)-[:AZRoleApprover]->(:AZRole)\nRETURN p LIMIT 100`,
            },
        ],
    },
    {
        subheader: 'NTLM Relay Attacks',
        category: categoryAD,
        queries: [
            {
                description: 'All coerce and NTLM relay edges',
                cypher: 'MATCH p = (n:Base)-[:CoerceAndRelayNTLMToLDAP|CoerceAndRelayNTLMToLDAPS|CoerceAndRelayNTLMToADCS|CoerceAndRelayNTLMToSMB]->(:Base)\nRETURN p LIMIT 500',
            },
            {
                description: 'ESC8-vulnerable Enterprise CAs',
                cypher: 'MATCH (n:EnterpriseCA)\nWHERE n.hasvulnerableendpoint=true\nRETURN n',
            },
            {
                description: 'Computers with the outgoing NTLM setting set to Deny all',
                cypher: 'MATCH (c:Computer)\nWHERE c.restrictoutboundntlm = True\nRETURN c LIMIT 1000',
            },
            {
                description: 'Computers with membership in Protected Users',
                cypher: 'MATCH p = (:Base)-[:MemberOf*1..]->(g:Group)\nWHERE g.objectid ENDS WITH "-525"\nRETURN p LIMIT 1000',
            },
            {
                description: 'DCs vulnerable to NTLM relay to LDAP attacks',
                cypher: 'MATCH p = (dc:Computer)-[:DCFor]->(:Domain)\nWHERE (dc.ldapavailable = True AND dc.ldapsigning = False)\nOR (dc.ldapsavailable = True AND dc.ldapsepa = False)\nOR (dc.ldapavailable = True AND dc.ldapsavailable = True AND dc.ldapsigning = False and dc.ldapsepa = True)\nRETURN p',
            },
            {
                description: 'Computers with the WebClient running',
                cypher: 'MATCH (c:Computer)\nWHERE c.webclientrunning = True\nRETURN c LIMIT 1000',
            },
            {
                description: 'Computers not requiring inbound SMB signing',
                cypher: 'MATCH (n:Computer)\nWHERE n.smbsigning = False\nRETURN n',
            },
        ],
    },
];
