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

import {
    ActiveDirectoryKindProperties,
    ActiveDirectoryNodeKind,
    AzureKindProperties,
    AzureNodeKind,
} from '../../graphSchema';

type PropertyHelp = {
    nodeTypes: readonly (ActiveDirectoryNodeKind | AzureNodeKind)[];
    description: string;
    location: string;
};

export const entityPropertyHelp: Partial<Record<ActiveDirectoryKindProperties | AzureKindProperties, PropertyHelp>> = {
    [ActiveDirectoryKindProperties.AdminCount]: {
        nodeTypes: [ActiveDirectoryNodeKind.User],
        description:
            'Indicates whether Active Directory has marked this account for protected-account handling. The flag can remain after privileged group membership is removed.',
        location:
            'Find this in Active Directory Users and Computers: Enable View → Advanced Features, then open the user’s Properties → Attribute Editor → adminCount.',
    },
    [ActiveDirectoryKindProperties.Sensitive]: {
        nodeTypes: [ActiveDirectoryNodeKind.User],
        description: 'Marks the account as sensitive and prevents delegation of its security context.',
        location:
            'Find this in Active Directory Users and Computers: Open the user’s Properties → Account → Account options → “Account is sensitive and cannot be delegated.”',
    },
    [ActiveDirectoryKindProperties.PasswordNotRequired]: {
        nodeTypes: [ActiveDirectoryNodeKind.User],
        description: 'Indicates a password-length policy exemption; it does not establish that the password is blank.',
        location:
            'Find this in Active Directory Users and Computers: Enable View → Advanced Features, then open the user’s Properties → Attribute Editor → userAccountControl. Inspect the PASSWD_NOTREQD flag (0x20 / 32).',
    },
    [ActiveDirectoryKindProperties.TrustedToAuth]: {
        nodeTypes: [ActiveDirectoryNodeKind.User, ActiveDirectoryNodeKind.Computer],
        description:
            'Reports the TRUSTED_TO_AUTH_FOR_DELEGATION flag, which enables constrained delegation with protocol transition. This flag does not cover Kerberos-only constrained delegation.',
        location:
            'Find this in Active Directory Users and Computers: View → Advanced Features → account Properties → Attribute Editor → userAccountControl; inspect flag 0x1000000.',
    },
    [ActiveDirectoryKindProperties.GMSA]: {
        nodeTypes: [ActiveDirectoryNodeKind.User],
        description:
            'Identifies a group managed service account, whose password is managed by Active Directory and can be used by authorized hosts.',
        location:
            'Find this in Windows PowerShell with the Active Directory module: Get-ADServiceAccount -Identity <account>; ObjectClass is msDS-GroupManagedServiceAccount.',
    },
    [ActiveDirectoryKindProperties.MSA]: {
        nodeTypes: [ActiveDirectoryNodeKind.User],
        description:
            'Identifies a standalone managed service account, whose password is managed by Active Directory for use on one computer.',
        location:
            'Find this in Windows PowerShell with the Active Directory module: Get-ADServiceAccount -Identity <account>; ObjectClass is msDS-ManagedServiceAccount.',
    },
    [ActiveDirectoryKindProperties.SIDHistory]: {
        nodeTypes: [ActiveDirectoryNodeKind.User, ActiveDirectoryNodeKind.Computer, ActiveDirectoryNodeKind.Group],
        description:
            'Additional security identifiers retained on this account, commonly after migration from another domain.',
        location:
            'Find this in Active Directory Users and Computers: View → Advanced Features → account Properties → Attribute Editor → sIDHistory.',
    },
    [ActiveDirectoryKindProperties.SFUPassword]: {
        nodeTypes: [ActiveDirectoryNodeKind.User],
        description: 'Legacy Services for UNIX password attribute that may contain a UNIX-format password hash.',
        location:
            'Find this in Active Directory Users and Computers: View → Advanced Features → user Properties → Attribute Editor → msSFU30Password, when that schema attribute exists.',
    },
    [ActiveDirectoryKindProperties.UnixPassword]: {
        nodeTypes: [ActiveDirectoryNodeKind.User],
        description:
            'Legacy UNIX authentication password attribute; its contents depend on the software that populated it.',
        location:
            'Find this in Active Directory Users and Computers: View → Advanced Features → user Properties → Attribute Editor → unixUserPassword, when that schema attribute exists.',
    },
    [ActiveDirectoryKindProperties.UserPassword]: {
        nodeTypes: [ActiveDirectoryNodeKind.User],
        description:
            'Collected values from the userPassword directory attribute, whose behavior depends on directory configuration. An empty field does not mean the account has no password.',
        location:
            'Find this in Active Directory Users and Computers: View → Advanced Features → user Properties → Attribute Editor → userPassword.',
    },
    [ActiveDirectoryKindProperties.DontRequirePreAuth]: {
        nodeTypes: [ActiveDirectoryNodeKind.User],
        description:
            'Allows the account to request a Kerberos ticket-granting ticket without Kerberos pre-authentication.',
        location:
            'Find this in Active Directory Users and Computers: user Properties → Account → Account options → Do not require Kerberos preauthentication.',
    },
    [ActiveDirectoryKindProperties.SupportedKerberosEncryptionTypes]: {
        nodeTypes: [ActiveDirectoryNodeKind.User, ActiveDirectoryNodeKind.Computer],
        description: 'The Kerberos encryption types advertised by this account for service-ticket encryption.',
        location:
            'Find this in Active Directory Users and Computers: View → Advanced Features → account Properties → Attribute Editor → msDS-SupportedEncryptionTypes.',
    },
    [ActiveDirectoryKindProperties.NTLMMinServerSec]: {
        nodeTypes: [ActiveDirectoryNodeKind.Computer],
        description:
            'The collected bitmask of minimum NTLM session-security requirements for servers, such as NTLMv2 session security and 128-bit encryption.',
        location:
            'Find this in Local Security Policy on the computer: Security Settings → Local Policies → Security Options → Network security: Minimum session security for NTLM SSP based (including secure RPC) servers.',
    },
    [ActiveDirectoryKindProperties.NTLMMinClientSec]: {
        nodeTypes: [ActiveDirectoryNodeKind.Computer],
        description:
            'The collected bitmask of minimum NTLM session-security requirements for clients, such as NTLMv2 session security and 128-bit encryption.',
        location:
            'Find this in Local Security Policy on the computer: Security Settings → Local Policies → Security Options → Network security: Minimum session security for NTLM SSP based (including secure RPC) clients.',
    },
    [ActiveDirectoryKindProperties.LMCompatibilityLevel]: {
        nodeTypes: [ActiveDirectoryNodeKind.Computer],
        description:
            'The collected LAN Manager authentication level, which controls LM, NTLM, and NTLMv2 authentication behavior.',
        location:
            'Find this in Local Security Policy on the computer: Security Settings → Local Policies → Security Options → Network security: LAN Manager authentication level.',
    },
    [ActiveDirectoryKindProperties.LDAPSEPA]: {
        nodeTypes: [ActiveDirectoryNodeKind.Computer],
        description:
            'Reports the LDAPS channel-binding test result. TRUE means authentication with invalid channel bindings did not succeed; it does not prove that channel binding is required for every client.',
        location:
            'Find this in Local Security Policy on the domain controller: Security Settings → Local Policies → Security Options → Domain controller: LDAP server channel binding token requirements.',
    },
    [ActiveDirectoryKindProperties.GroupScope]: {
        nodeTypes: [ActiveDirectoryNodeKind.Group],
        description:
            'The group scope, such as Global, Domain Local, or Universal, which controls membership and where the group can be granted permissions.',
        location: 'Find this in Active Directory Users and Computers: group Properties → General → Group scope.',
    },
    [ActiveDirectoryKindProperties.MachineAccountQuota]: {
        nodeTypes: [ActiveDirectoryNodeKind.Domain],
        description:
            'The domain quota for computer accounts that a user can create without separately delegated computer-creation permissions.',
        location:
            'Find this in ADSI Edit: connect to the Default naming context → domain object Properties → ms-DS-MachineAccountQuota.',
    },
    [ActiveDirectoryKindProperties.DSHeuristics]: {
        nodeTypes: [ActiveDirectoryNodeKind.Domain],
        description:
            'A forest-wide string of Directory Service settings, with individual character positions controlling different behavior.',
        location:
            'Find this in ADSI Edit: connect to the Configuration naming context → CN=Services → CN=Windows NT → CN=Directory Service → Properties → dSHeuristics.',
    },
    [ActiveDirectoryKindProperties.BlocksInheritance]: {
        nodeTypes: [ActiveDirectoryNodeKind.OU],
        description: 'Blocks inherited Group Policy links from parent containers. Enforced links can still apply.',
        location:
            'Find this in Group Policy Management: Forest → Domains → domain → right-click the organizational unit → Block Inheritance.',
    },
    [ActiveDirectoryKindProperties.GPOStatus]: {
        nodeTypes: [ActiveDirectoryNodeKind.GPO],
        description:
            'Whether the GPO computer settings, user settings, or both are enabled. This is separate from whether an individual GPO link is enabled.',
        location:
            'Find this in Group Policy Management: Forest → Domains → domain → Group Policy Objects → right-click the GPO → GPO Status.',
    },
    [ActiveDirectoryKindProperties.EnrolleeSuppliesSubject]: {
        nodeTypes: [ActiveDirectoryNodeKind.CertTemplate],
        description:
            'Allows the certificate requester to supply subject information in the request instead of building it from Active Directory.',
        location: 'Find this in Certificate Templates: template Properties → Subject Name → Supply in the request.',
    },
    [ActiveDirectoryKindProperties.RequiresManagerApproval]: {
        nodeTypes: [ActiveDirectoryNodeKind.CertTemplate],
        description: 'Requires approval by a CA certificate manager before a request using this template is issued.',
        location:
            'Find this in Certificate Templates: template Properties → Issuance Requirements → CA certificate manager approval.',
    },
    [ActiveDirectoryKindProperties.EffectiveEKUs]: {
        nodeTypes: [ActiveDirectoryNodeKind.CertTemplate],
        description:
            'The certificate purposes selected by the collector from the template EKU or application-policy attributes, according to the template schema version.',
        location:
            'Find this in Certificate Templates: template Properties → Extensions → Application Policies; raw values are pKIExtendedKeyUsage and msPKI-Certificate-Application-Policy.',
    },
    [ActiveDirectoryKindProperties.IsUserSpecifiesSanEnabled]: {
        nodeTypes: [ActiveDirectoryNodeKind.EnterpriseCA],
        description:
            'Reports the CA policy flag that accepts subject alternative names supplied as certificate-request attributes.',
        location:
            'Find this in Command Prompt on the CA: certutil -getreg policy\\EditFlags; inspect EDITF_ATTRIBUTESUBJECTALTNAME2.',
    },
    [AzureKindProperties.UserType]: {
        nodeTypes: [AzureNodeKind.User],
        description:
            'The user’s directory classification, such as Member or Guest. This classification does not by itself identify where the user authenticates.',
        location:
            'Find this in Microsoft Entra admin center: Entra ID → Users → select the user → Overview → User type.',
    },
    [AzureKindProperties.LastSuccessfulSignInDateTime]: {
        nodeTypes: [AzureNodeKind.User],
        description: 'The most recent successful interactive or non-interactive sign-in reported for this user.',
        location:
            'Source attribute in Microsoft Graph: signInActivity.lastSuccessfulSignInDateTime on the user object.',
    },
    [AzureKindProperties.OnPremID]: {
        nodeTypes: [AzureNodeKind.User, AzureNodeKind.Group],
        description: 'The security identifier (SID) of the corresponding on-premises Active Directory object.',
        location: 'Source attribute in Microsoft Graph: onPremisesSecurityIdentifier on the user or group object.',
    },
    [AzureKindProperties.OnPremSyncEnabled]: {
        nodeTypes: [AzureNodeKind.User, AzureNodeKind.Group],
        description: 'Reports whether this object is synchronized from an on-premises directory.',
        location: 'Source attribute in Microsoft Graph: onPremisesSyncEnabled on the user or group object.',
    },
    [AzureKindProperties.IsAssignableToRole]: {
        nodeTypes: [AzureNodeKind.Group],
        description:
            'Whether this group can receive Microsoft Entra role assignments. This does not mean that a role is currently assigned.',
        location:
            'Find this in Microsoft Entra admin center: Entra ID → Groups → All groups → Add filters → Role assignable.',
    },
    [AzureKindProperties.SignInAudience]: {
        nodeTypes: [AzureNodeKind.App],
        description:
            'The account types the application is configured to support, such as one organization, multiple organizations, or personal Microsoft accounts.',
        location:
            'Find this in Microsoft Entra admin center: Entra ID → App registrations → select the application → Manifest → signInAudience.',
    },
    [AzureKindProperties.ServicePrincipalType]: {
        nodeTypes: [AzureNodeKind.ServicePrincipal],
        description:
            'Whether this service principal represents an application, a managed identity, or a legacy application.',
        location: 'Source attribute in Microsoft Graph: servicePrincipalType on the service principal object.',
    },
    [AzureKindProperties.AppOwnerOrganizationID]: {
        nodeTypes: [AzureNodeKind.ServicePrincipal],
        description:
            'The tenant ID of the organization where the associated application is registered. Applies to service principals backed by an application.',
        location: 'Source attribute in Microsoft Graph: appOwnerOrganizationId on the service principal object.',
    },
    [AzureKindProperties.TrustType]: {
        nodeTypes: [AzureNodeKind.Device],
        description:
            'The device’s join relationship with Microsoft Entra ID: AzureAd means Entra joined, Workplace means Entra registered, and ServerAd means Entra hybrid joined.',
        location: 'Find this in Microsoft Entra admin center: Entra ID → Devices → All devices → Join type.',
    },
    [AzureKindProperties.IsBuiltIn]: {
        nodeTypes: [AzureNodeKind.Role],
        description: 'Whether this role definition is included with Microsoft Entra ID or is a custom role definition.',
        location: 'Source attribute in Microsoft Graph: isBuiltIn on the unifiedRoleDefinition object.',
    },
    [AzureKindProperties.EnableRBACAuthorization]: {
        nodeTypes: [AzureNodeKind.KeyVault],
        description:
            'Whether access to the vault’s keys, secrets, and certificates uses Azure role-based access control rather than vault access policies.',
        location:
            'Find this in Microsoft Azure portal: Key vaults → select the vault → Settings → Access configuration → Permission model.',
    },
};
