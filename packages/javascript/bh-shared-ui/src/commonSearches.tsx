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

// Join all elements with "|" but the last element, then append last element
// produces element1|element2|element3
const azureTransitEdgeTypes = AzurePathfindingEdges().slice(0, -1).join('|') + AzurePathfindingEdges().slice(-1);
const adTransitEdgeTypes =
    ActiveDirectoryPathfindingEdges().slice(0, -1).join('|') + '|' + ActiveDirectoryPathfindingEdges().slice(-1);

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
                cypher: `MATCH p=(n:Group)<-[:MemberOf*1..]-(m)\nWHERE n.objectid ENDS WITH "-512"\nRETURN p`,
            },
            {
                description: 'Map domain trusts',
                cypher: `MATCH p=(n:Domain)-[]->(m:Domain)\nRETURN p`,
            },
            {
                description: 'Computers with unsupported operating systems',
                cypher: `MATCH (n:Computer)\nWHERE n.operatingsystem =~ "(?i).*Windows.* (2000|2003|2008|2012|xp|vista|7|8|me|nt).*"\nRETURN n`,
            },
            {
                description: 'Locations of high value/Tier Zero objects',
                cypher: `MATCH p = (:Domain)-[:Contains*1..]->(n:Base)\nWHERE n.system_tags="admin_tier_0"\nRETURN p`,
            },
        ],
    },
    {
        subheader: 'Dangerous Privileges',
        category: categoryAD,
        queries: [
            {
                description: 'Principals with DCSync privileges',
                cypher: `MATCH p=()-[:DCSync|AllExtendedRights|GenericAll]->(:Domain)\nRETURN p`,
            },
            {
                description: 'Users with foreign domain group membership',
                cypher: `MATCH p=(n:User)-[:MemberOf]->(m:Group)\nWHERE m.domainsid<>n.domainsid\nRETURN p`,
            },
            {
                description: 'Groups with foreign domain group membership',
                cypher: `MATCH p=(n:Group)-[:MemberOf]->(m:Group)\nWHERE m.domainsid<>n.domainsid AND n.name<>m.name\nRETURN p`,
            },
            {
                description: 'Computers where Domain Users are local administrators',
                cypher: `MATCH p=(m:Group)-[:AdminTo]->(n:Computer)\nWHERE m.objectid ENDS WITH "-513"\nRETURN p`,
            },
            {
                description: 'Computers where Domain Users can read LAPS passwords',
                cypher: `MATCH p=(m:Group)-[:AllExtendedRights|ReadLAPSPassword]->(n:Computer)\nWHERE m.objectid ENDS WITH "-513"\nRETURN p`,
            },
            {
                description: 'Paths from Domain Users to high value/Tier Zero targets',
                cypher: `MATCH p=shortestPath((m:Group)-[:${adTransitEdgeTypes}*1..]->(n))\nWHERE n.system_tags="admin_tier_0" AND m.objectid ENDS WITH "-513" AND m<>n\nRETURN p`,
            },
            {
                description: 'Workstations where Domain Users can RDP',
                cypher: `MATCH p=(m:Group)-[:CanRDP]->(c:Computer)\nWHERE m.objectid ENDS WITH "-513" AND NOT c.operatingsystem CONTAINS "Server"\nRETURN p`,
            },
            {
                description: 'Servers where Domain Users can RDP',
                cypher: `MATCH p=(m:Group)-[:CanRDP]->(c:Computer)\nWHERE m.objectid ENDS WITH "-513" AND c.operatingsystem CONTAINS "Server"\nRETURN p`,
            },
            {
                description: 'Dangerous privileges for Domain Users groups',
                cypher: `MATCH p=(m:Group)-[:Owns|WriteDacl|GenericAll|WriteOwner|ExecuteDCOM|GenericWrite|AllowedToDelegate|ForceChangePassword]->(n:Computer)\nWHERE m.objectid ENDS WITH "-513"\nRETURN p`,
            },
            {
                description: 'Domain Admins logons to non-Domain Controllers',
                cypher: `MATCH (dc)-[r:MemberOf*0..]->(g:Group)\nWHERE g.objectid ENDS WITH '-516'\nWITH COLLECT(dc) AS exclude\nMATCH p = (c:Computer)-[n:HasSession]->(u:User)-[r2:MemberOf*1..]->(g:Group)\nWHERE g.objectid ENDS WITH '-512' AND NOT c IN exclude\nRETURN p`,
            },
        ],
    },
    {
        subheader: 'Kerberos Interaction',
        category: categoryAD,
        queries: [
            {
                description: 'Kerberoastable members of high value/Tier Zero groups',
                cypher: `MATCH p=shortestPath((n:User)-[:MemberOf]->(g:Group))\nWHERE g.system_tags = "admin_tier_0" AND n.hasspn=true\nRETURN p`,
            },
            {
                description: 'All Kerberoastable users',
                cypher: 'MATCH (n:User)\nWHERE n.hasspn=true\nRETURN n',
            },
            {
                description: 'Kerberoastable users with most privileges',
                cypher: `MATCH (u:User {hasspn:true})\nOPTIONAL MATCH (u)-[:AdminTo]->(c1:Computer)\nOPTIONAL MATCH (u)-[:MemberOf*1..]->(:Group)-[:AdminTo]->(c2:Computer)\nWITH u,COLLECT(c1) + COLLECT(c2) AS tempVar\nUNWIND tempVar AS comps\nRETURN u`,
            },
            {
                description: 'AS-REP Roastable users (DontReqPreAuth)',
                cypher: `MATCH (u:User)\nWHERE u.dontreqpreauth = true\nRETURN u`,
            },
        ],
    },
    {
        subheader: 'Shortest Paths',
        category: categoryAD,
        queries: [
            {
                description: 'Shortest paths to systems trusted for unconstrained delegation',
                cypher: `MATCH p=shortestPath((n)-[:${adTransitEdgeTypes}*1..]->(m:Computer))\nWHERE m.unconstraineddelegation = true AND n<>m\nRETURN p`,
            },
            {
                description: 'Shortest paths from Kerberoastable users',
                cypher: `MATCH p=shortestPath((n)-[:${adTransitEdgeTypes}*1..]->(m:Computer))\nWHERE m.unconstraineddelegation = true AND n<>m\nRETURN p`,
            },
            {
                description: 'Shortest paths to Domain Admins from Kerberoastable users',
                cypher: `MATCH p=shortestPath((n:User)-[:${adTransitEdgeTypes}*1..]->(m:Group))\nWHERE n.hasspn = true AND m.objectid ENDS WITH "-512"\nRETURN p`,
            },
            {
                description: 'Shortest paths to high value/Tier Zero targets',
                cypher: `MATCH p=shortestPath((n)-[:${adTransitEdgeTypes}*1..]->(m))\nWHERE m.system_tags = "admin_tier_0" AND n<>m\nRETURN p`,
            },
            {
                description: 'Shortest paths from Domain Users to high value/Tier Zero targets',
                cypher: `MATCH p=shortestPath((n:Group)-[:${adTransitEdgeTypes}*1..]->(m))\nWHERE m.system_tags = "admin_tier_0" AND n.objectid ENDS WITH "-513" AND n<>m\nRETURN p`,
            },
            {
                description: 'Shortest paths to Domain Admins',
                cypher: `MATCH p=shortestPath((n)-[:${adTransitEdgeTypes}*1..]->(g:Group))\nWHERE g.objectid ENDS WITH "-512" AND n<>g\nRETURN p`,
            },
        ],
    },
    {
        subheader: 'General',
        category: categoryAzure,
        queries: [
            {
                description: 'All Global Administrators',
                cypher: 'MATCH p = (n)-[r:AZGlobalAdmin*1..]->(m)\nRETURN p',
            },
            {
                description: 'All members of high privileged roles',
                cypher: `MATCH p=(n)-[:AZHasRole|AZMemberOf*1..2]->(r:AZRole)\nWHERE r.name =~ '(?i)${highPrivilegedRoleDisplayNameRegex}'\nRETURN p`,
            },
        ],
    },
    {
        subheader: 'Shortest Paths',
        category: categoryAzure,
        queries: [
            {
                description: 'Shortest paths to high value/Tier Zero targets',
                cypher: `MATCH p=shortestPath((m:AZUser)-[r:${azureTransitEdgeTypes}*1..]->(n))\nWHERE n.system_tags = "admin_tier_0" AND n.name =~ '(?i)${highPrivilegedRoleDisplayNameRegex}' AND m<>n\nRETURN p`,
            },
            {
                description: 'Shortest paths to privileged roles',
                cypher: `MATCH p=shortestPath((m)-[r:${azureTransitEdgeTypes}*1..]->(n:AZRole))\nWHERE n.name =~ '(?i)${highPrivilegedRoleDisplayNameRegex}' AND m<>n\nRETURN p`,
            },
            {
                description: 'Shortest paths from Azure Applications to high value/Tier Zero targets',
                cypher: `MATCH p=shortestPath((m:AZApp)-[r:${azureTransitEdgeTypes}*1..]->(n))\nWHERE n.system_tags = "admin_tier_0" AND m<>n\nRETURN p`,
            },
            {
                description: 'Shortest paths to Azure Subscriptions',
                cypher: `MATCH p=shortestPath((m)-[r:${azureTransitEdgeTypes}*1..]->(n:AZSubscription))\nWHERE m<>n\nRETURN p`,
            },
        ],
    },
    {
        subheader: 'Microsoft Graph',
        category: categoryAzure,
        queries: [
            {
                description: 'All service principals with Microsoft Graph privilege to grant arbitrary App Roles',
                cypher: 'MATCH p=(n)-[r:AZMGGrantAppRoles]->(o:AZTenant)\nRETURN p',
            },
            {
                description: 'All service principals with Microsoft Graph App Role assignments',
                cypher: 'MATCH p=(m:AZServicePrincipal)-[r:AZMGAppRoleAssignment_ReadWrite_All|AZMGApplication_ReadWrite_All|AZMGDirectory_ReadWrite_All|AZMGGroupMember_ReadWrite_All|AZMGGroup_ReadWrite_All|AZMGRoleManagement_ReadWrite_Directory|AZMGServicePrincipalEndpoint_ReadWrite_All]->(n:AZServicePrincipal)\nRETURN p',
            },
        ],
    },
];
