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

import { Link, Typography } from '@mui/material';
import { FC } from 'react';
import CodeController from '../CodeController/CodeController';
import { EdgeInfoProps } from '../index';

const WindowsAbuse: FC<EdgeInfoProps> = ({ sourceName, sourceType, targetName, targetType }) => {
    switch (targetType) {
        case 'Group':
            return (
                <>
                    <Typography variant='body2'>
                        GenericWrite to a group allows you to directly modify group membership of the group.
                    </Typography>
                    <Typography variant='body2'>
                        There are at least two ways to execute this attack. The first and most obvious is by using the
                        built-in net.exe binary in Windows (e.g.: net group "Domain Admins" harmj0y /add /domain). See
                        the opsec considerations tab for why this may be a bad idea. The second, and highly recommended
                        method, is by using the Add-DomainGroupMember function in PowerView. This function is superior
                        to using the net.exe binary in several ways. For instance, you can supply alternate credentials,
                        instead of needing to run a process as or logon as the user with the AddMember permission.
                        Additionally, you have much safer execution options than you do with spawning net.exe (see the
                        opsec tab).
                    </Typography>
                    <Typography variant='body2'>
                        To abuse this permission with PowerView's Add-DomainGroupMember, first import PowerView into
                        your agent session or into a PowerShell instance at the console. You may need to authenticate to
                        the Domain Controller as
                        {sourceType === 'User'
                            ? `${sourceName} if you are not running a process as that user`
                            : `a member of ${sourceName} if you are not running a process as a member`}
                        . To do this in conjunction with Add-DomainGroupMember, first create a PSCredential object
                        (these examples comes from the PowerView help documentation):
                    </Typography>
                    <Typography component={'pre'}>
                        {"$SecPassword = ConvertTo-SecureString 'Password123!' -AsPlainText -Force\n" +
                            "$Cred = New-Object System.Management.Automation.PSCredential('TESTLAB\\dfm.a', $SecPassword)"}
                    </Typography>
                    <Typography variant='body2'>
                        Then, use Add-DomainGroupMember, optionally specifying $Cred if you are not already running a
                        process as {sourceName}:
                    </Typography>
                    <Typography component={'pre'}>
                        {"Add-DomainGroupMember -Identity 'Domain Admins' -Members 'harmj0y' -Credential $Cred"}
                    </Typography>
                    <Typography variant='body2'>
                        Finally, verify that the user was successfully added to the group with PowerView's
                        Get-DomainGroupMember:
                    </Typography>
                    <Typography component={'pre'}>{"Get-DomainGroupMember -Identity 'Domain Admins'"}</Typography>
                </>
            );
        case 'User':
            return (
                <>
                    <Typography variant='body2'>
                        GenericWrite grants {sourceName} the permission to write to the "msds-KeyCredentialLink"
                        attribute of target. Writing to this property allows an attacker to create "Shadow Credentials"
                        on the object and authenticate as the principal using kerberos PKINIT. This is equivalent to the
                        "AddKeyCredentialLink" edge.
                    </Typography>
                    <Typography variant='body2'>
                        The permission also grants write access to the "altSecurityIdentities" attribute, which enables
                        an ADCS ESC14 Scenario A attack.
                    </Typography>
                    <Typography variant='body2'>
                        Alternatively, GenericWrite enables {sourceName} to set a ServicePrincipalName (SPN) on the
                        targeted user, which may be abused in a Targeted Kerberoast attack.
                    </Typography>

                    <Typography variant='body1'> Shadow Credentials attack </Typography>

                    <Typography variant='body2'>To abuse the permission, use Whisker. </Typography>

                    <Typography variant='body2'>
                        You may need to authenticate to the Domain Controller as{' '}
                        {sourceType === 'User' || sourceType === 'Computer'
                            ? `${sourceName} if you are not running a process as that user/computer`
                            : `a member of ${sourceName} if you are not running a process as a member`}
                    </Typography>

                    <Typography component={'pre'}>{'Whisker.exe add /target:<TargetPrincipal>'}</Typography>

                    <Typography variant='body2'>
                        For other optional parameters, view the Whisker documentation.
                    </Typography>

                    <Typography variant='body1'> ADCS ESC14 Scenario A </Typography>
                    <Typography variant='body2'>
                        An attacker can add an explicit certificate mapping in the altSecurityIdentities of the target
                        referring to a certificate in the attacker's possession, and then use this certificate to
                        authenticate as the target.
                    </Typography>
                    <Typography variant='body2'>
                        The certificate must meet the following requirements:
                        <ol style={{ listStyleType: 'decimal', paddingLeft: '1.5em' }}>
                            <li>Chain up to trusted root CA on the DC</li>
                            <li>Enhanced Key Usage extension contains an EKU that enables domain authentication</li>
                            <li>
                                Subject Alternative Name (SAN) does NOT contain a "Other Name/Principal Name" entry
                                (UPN)
                            </li>
                        </ol>
                        <p className='my-4'>
                            The EKUs that enable domain authentication over Kerberos:
                            <ul style={{ paddingLeft: '1.5em' }}>
                                <li>Client Authentication (1.3.6.1.5.5.7.3.2)</li>
                                <li>PKINIT Client Authentication (1.3.6.1.5.2.3.4)</li>
                                <li>Smart Card Logon (1.3.6.1.4.1.311.20.2.2)</li>
                                <li>Any Purpose (2.5.29.37.0)</li>
                                <li>SubCA (no EKUs)</li>
                            </ul>
                        </p>
                        <p className='my-4'>
                            The last certificate requirement means that user certificates will not work, so the
                            certificate typically must be of a computer. By default, the ADCS certificate template{' '}
                            <i>Computer (Machine)</i> meets these requirements and grants Domain Computers enrollment
                            rights. The target can still be a user.
                        </p>
                        The last requirement does not have to be met if a DC has UPN mapping disabled (see{' '}
                        <Link
                            target='_blank'
                            rel='noopener'
                            href='https://learn.microsoft.com/en-us/previous-versions/windows/it-pro/windows-server-2008-R2-and-2008/ff520074(v=ws.10)'>
                            How to disable the Subject Alternative Name for UPN mapping
                        </Link>
                        ).
                    </Typography>
                    <Typography variant='body2'>
                        Obtain a certificate meeting the above requirements for example by dumping a certificate from a
                        computer, or enrolling a new certificate as a computer:
                    </Typography>
                    <Typography component={'pre'}>
                        {
                            'Certify.exe request /ca:rootdomaindc.forestroot.com\\forestroot-RootDomainDC-CA /template:Machine /machine'
                        }
                    </Typography>
                    <Typography variant='body2'>
                        If the enrollment fails with an error message stating that the Email or DNS name is unavailable
                        and cannot be added to the Subject or Subject Alternate name, then it is because the enrollee
                        principal does not have their mail or dNSHostName attribute set, which is required by the
                        certificate template. The mail attribute can be set on both user and computer objects but the
                        dNSHostName attribute can only be set on computer objects. Computers have validated write
                        permission to their own dNSHostName attribute by default, but neither users nor computers can
                        write to their own mail attribute by default.
                    </Typography>
                    <Typography variant='body2'>
                        Save the certificate as cert.pem and the private key as cert.key. Use certutil to obtain the
                        certificate as a PFX file:
                    </Typography>
                    <Typography component={'pre'}>{'certutil.exe -MergePFX .\\cert.pem .\\cert.pfx'}</Typography>
                    <Typography variant='body2'>
                        The abuse is possible with the strong explicit certificate mappings X509IssuerSerialNumber or
                        X509SHA1PublicKey. In this example, we use X509SHA1PublicKey.
                    </Typography>
                    <Typography variant='body2'>
                        Get the SHA1 hash of the certificate public key using certutil:
                    </Typography>
                    <CodeController>
                        {`certutil.exe -dump -v .\\cert.pfx
…
Cert Hash(sha1): ef9375785421d3ad286d8bdeb166f0f697266992
…`}
                    </CodeController>
                    <Typography variant='body2'>
                        Use Add-AltSecIDMapping to add the explicit certificate mapping string to the
                        'altSecurityIdentities' attribute of the target principal:
                    </Typography>
                    <Typography component={'pre'}>
                        {
                            'Add-AltSecIDMapping -DistinguishedName "CN=Target,CN=Users,DC=forestroot,DC=com" -MappingString "X509:<SHA1-PUKEY>ef9375785421d3ad286d8bdeb166f0f697266992"'
                        }
                    </Typography>
                    <Typography variant='body2'>
                        Verify the that the mapping was added using Get-AltSecIDMapping:
                    </Typography>
                    <Typography component={'pre'}>
                        {'Get-AltSecIDMapping -SearchBase "CN=Target,CN=Users,DC=forestroot,DC=com"'}
                    </Typography>
                    <Typography variant='body2'>
                        Use Rubeus to request a ticket granting ticket (TGT) from the domain, specifying the target
                        identity, and the PFX-formatted certificate:
                    </Typography>
                    <Typography component={'pre'}>
                        {'Rubeus asktgt /user:"forestroot\\target" /certificate:cert.pfx /ptt'}
                    </Typography>
                    <Typography variant='body2'>
                        After the execution of the abuse, use Remove-AltSecIDMapping to remove the explicit certificate
                        mapping string from the 'altSecurityIdentities' attribute of the target principal:
                    </Typography>
                    <Typography component={'pre'}>
                        {
                            'Remove-AltSecIDMapping -DistinguishedName "CN=Target,CN=Users,DC=forestroot,DC=com" -MappingString "X509:<SHA1-PUKEY>ef9375785421d3ad286d8bdeb166f0f697266992"'
                        }
                    </Typography>

                    <Typography variant='body1'> Targeted Kerberoast attack </Typography>

                    <Typography variant='body2'>
                        A targeted kerberoast attack can be performed using PowerView's Set-DomainObject along with
                        Get-DomainSPNTicket.
                    </Typography>
                    <Typography variant='body2'>
                        You may need to authenticate to the Domain Controller as{' '}
                        {sourceType === 'User'
                            ? `${sourceName} if you are not running a process as that user`
                            : `a member of ${sourceName} if you are not running a process as a member`}
                        . To do this in conjunction with Set-DomainObject, first create a PSCredential object (these
                        examples comes from the PowerView help documentation):
                    </Typography>
                    <Typography component={'pre'}>
                        {"$SecPassword = ConvertTo-SecureString 'Password123!' -AsPlainText -Force\n" +
                            "$Cred = New-Object System.Management.Automation.PSCredential('TESTLAB\\dfm.a', $SecPassword)"}
                    </Typography>
                    <Typography variant='body2'>
                        Then, use Set-DomainObject, optionally specifying $Cred if you are not already running a process
                        as {sourceName}:
                    </Typography>
                    <Typography component={'pre'}>
                        {
                            "Set-DomainObject -Credential $Cred -Identity harmj0y -SET @{serviceprincipalname='nonexistent/BLAHBLAH'}"
                        }
                    </Typography>
                    <Typography variant='body2'>
                        After running this, you can use Get-DomainSPNTicket as follows:
                    </Typography>
                    <Typography component={'pre'}>{'Get-DomainSPNTicket -Credential $Cred harmj0y | fl'}</Typography>
                    <Typography variant='body2'>
                        The recovered hash can be cracked offline using the tool of your choice. Cleanup of the
                        ServicePrincipalName can be done with the Set-DomainObject command:
                    </Typography>
                    <Typography component={'pre'}>
                        {'Set-DomainObject -Credential $Cred -Identity harmj0y -Clear serviceprincipalname'}
                    </Typography>
                </>
            );
        case 'GPO':
            return (
                <>
                    <Typography variant='body2'>
                        With GenericWrite on a GPO, you may make modifications to that GPO which will then apply to the
                        users and computers affected by the GPO. Select the target object you wish to push an evil
                        policy down to, then use the gpedit GUI to modify the GPO, using an evil policy that allows
                        item-level targeting, such as a new immediate scheduled task. Then wait for the group policy
                        client to pick up and execute the new evil policy.
                    </Typography>
                    <Typography variant='body2'>
                        Refer to{' '}
                        <Link target='_blank' rel='noopener' href='https://wald0.com/?p=179'>
                            A Red Teamer's Guide to GPOs and OUs
                        </Link>
                        for details about the abuse technique, and check out{' '}
                        <Link target='_blank' rel='noopener' href='https://github.com/FSecureLABS/SharpGPOAbuse'>
                            SharpGPOAbuse
                        </Link>{' '}
                        for practical exploitation.
                    </Typography>
                    <Typography variant='body2'>
                        This edge can be a false positive in rare scenarios. If you have GenericWrite on the GPO with
                        'This object only' (no inheritance) and no other permissions in the ACL, it is not possible to
                        add or modify settings of the GPO. The GPO's settings are stored in SYSVOL under a folder for
                        the given GPO. Therefore, you need write access to child objects of this folder or create child
                        objects permission. The security descriptor of the GPO is reflected on the folder, meaning
                        permissions to write child items on the GPO are required.
                    </Typography>
                </>
            );
        case 'Computer':
            return (
                <>
                    <Typography variant='body2'>
                        GenericWrite grants {sourceName} the permission to write to the "msds-KeyCredentialLink"
                        attribute of {targetName}. Writing to this property allows an attacker to create "Shadow
                        Credentials" on the object and authenticate as the principal using kerberos PKINIT. This is
                        equivalent to the "AddKeyCredentialLink" edge.
                    </Typography>

                    <Typography variant='body2'>
                        Alternatively, GenericWrite on a computer object can be used to perform a Resource-Based
                        Constrained Delegation attack.
                    </Typography>

                    <Typography variant='body2'>
                        The permission also grants write access to the "altSecurityIdentities" attribute, which enables
                        an ADCS ESC14 Scenario A attack.
                    </Typography>

                    <Typography variant='body1'> Shadow Credentials attack </Typography>

                    <Typography variant='body2'>To abuse the permission, use Whisker. </Typography>

                    <Typography variant='body2'>
                        You may need to authenticate to the Domain Controller as{' '}
                        {sourceType === 'User' || sourceType === 'Computer'
                            ? `${sourceName} if you are not running a process as that user/computer`
                            : `a member of ${sourceName} if you are not running a process as a member`}
                    </Typography>

                    <Typography component={'pre'}>{'Whisker.exe add /target:<TargetPrincipal>'}</Typography>

                    <Typography variant='body2'>
                        For other optional parameters, view the Whisker documentation.
                    </Typography>

                    <Typography variant='body1'> Resource-Based Constrained Delegation attack </Typography>

                    <Typography variant='body2'>
                        Abusing this primitive is possible through the Rubeus project.
                    </Typography>

                    <Typography variant='body2'>
                        First, if an attacker does not control an account with an SPN set, Kevin Robertson's Powermad
                        project can be used to add a new attacker-controlled computer account:
                    </Typography>

                    <Typography component={'pre'}>
                        {
                            "New-MachineAccount -MachineAccount attackersystem -Password $(ConvertTo-SecureString 'Summer2018!' -AsPlainText -Force)"
                        }
                    </Typography>

                    <Typography variant='body2'>
                        PowerView can be used to then retrieve the security identifier (SID) of the newly created
                        computer account:
                    </Typography>

                    <Typography component={'pre'}>
                        $ComputerSid = Get-DomainComputer attackersystem -Properties objectsid | Select -Expand
                        objectsid
                    </Typography>

                    <Typography variant='body2'>
                        We now need to build a generic ACE with the attacker-added computer SID as the principal, and
                        get the binary bytes for the new DACL/ACE:
                    </Typography>

                    <Typography component={'pre'}>
                        {'$SD = New-Object Security.AccessControl.RawSecurityDescriptor -ArgumentList "O:BAD:(A;;CCDCLCSWRPWPDTLOCRSDRCWDWO;;;$($ComputerSid))"\n' +
                            '$SDBytes = New-Object byte[] ($SD.BinaryLength)\n' +
                            '$SD.GetBinaryForm($SDBytes, 0)'}
                    </Typography>

                    <Typography variant='body2'>
                        Next, we need to set this newly created security descriptor in the
                        msDS-AllowedToActOnBehalfOfOtherIdentity field of the computer account we're taking over, again
                        using PowerView in this case:
                    </Typography>

                    <Typography component={'pre'}>
                        {
                            "Get-DomainComputer $TargetComputer | Set-DomainObject -Set @{'msds-allowedtoactonbehalfofotheridentity'=$SDBytes}"
                        }
                    </Typography>

                    <Typography variant='body2'>
                        We can then use Rubeus to hash the plaintext password into its RC4_HMAC form:
                    </Typography>

                    <Typography component={'pre'}>{'Rubeus.exe hash /password:Summer2018!'}</Typography>

                    <Typography variant='body2'>
                        And finally we can use Rubeus' *s4u* module to get a service ticket for the service name (sname)
                        we want to "pretend" to be "admin" for. This ticket is injected (thanks to /ptt), and in this
                        case grants us access to the file system of the TARGETCOMPUTER:
                    </Typography>

                    <Typography component={'pre'}>
                        {
                            'Rubeus.exe s4u /user:attackersystem$ /rc4:EF266C6B963C0BB683941032008AD47F /impersonateuser:admin /msdsspn:cifs/TARGETCOMPUTER.testlab.local /ptt'
                        }
                    </Typography>

                    <Typography variant='body1'> ADCS ESC14 Scenario A </Typography>
                    <Typography variant='body2'>
                        An attacker can add an explicit certificate mapping in the altSecurityIdentities of the target
                        referring to a certificate in the attacker's possession, and then use this certificate to
                        authenticate as the target.
                    </Typography>
                    <Typography variant='body2'>
                        The certificate must meet the following requirements:
                        <ol style={{ listStyleType: 'decimal', paddingLeft: '1.5em' }}>
                            <li>Chain up to trusted root CA on the DC</li>
                            <li>Enhanced Key Usage extension contains an EKU that enables domain authentication</li>
                            <li>
                                Subject Alternative Name (SAN) does NOT contain a "Other Name/Principal Name" entry
                                (UPN)
                            </li>
                        </ol>
                        <p className='my-4'>
                            The EKUs that enable domain authentication over Kerberos:
                            <ul style={{ paddingLeft: '1.5em' }}>
                                <li>Client Authentication (1.3.6.1.5.5.7.3.2)</li>
                                <li>PKINIT Client Authentication (1.3.6.1.5.2.3.4)</li>
                                <li>Smart Card Logon (1.3.6.1.4.1.311.20.2.2)</li>
                                <li>Any Purpose (2.5.29.37.0)</li>
                                <li>SubCA (no EKUs)</li>
                            </ul>
                        </p>
                        <p className='my-4'>
                            The last certificate requirement means that user certificates will not work, so the
                            certificate typically must be of a computer. By default, the ADCS certificate template{' '}
                            <i>Computer (Machine)</i> meets these requirements and grants Domain Computers enrollment
                            rights. The target can still be a user.
                        </p>
                        The last requirement does not have to be met if a DC has UPN mapping disabled (see{' '}
                        <Link
                            target='_blank'
                            rel='noopener'
                            href='https://learn.microsoft.com/en-us/previous-versions/windows/it-pro/windows-server-2008-R2-and-2008/ff520074(v=ws.10)'>
                            How to disable the Subject Alternative Name for UPN mapping
                        </Link>
                        ).
                    </Typography>
                    <Typography variant='body2'>
                        Obtain a certificate meeting the above requirements for example by dumping a certificate from a
                        computer, or enrolling a new certificate as a computer:
                    </Typography>
                    <Typography component={'pre'}>
                        {
                            'Certify.exe request /ca:rootdomaindc.forestroot.com\\forestroot-RootDomainDC-CA /template:Machine /machine'
                        }
                    </Typography>
                    <Typography variant='body2'>
                        If the enrollment fails with an error message stating that the Email or DNS name is unavailable
                        and cannot be added to the Subject or Subject Alternate name, then it is because the enrollee
                        principal does not have their mail or dNSHostName attribute set, which is required by the
                        certificate template. The mail attribute can be set on both user and computer objects but the
                        dNSHostName attribute can only be set on computer objects. Computers have validated write
                        permission to their own dNSHostName attribute by default, but neither users nor computers can
                        write to their own mail attribute by default.
                    </Typography>
                    <Typography variant='body2'>
                        Save the certificate as cert.pem and the private key as cert.key. Use certutil to obtain the
                        certificate as a PFX file:
                    </Typography>
                    <Typography component={'pre'}>{'certutil.exe -MergePFX .\\cert.pem .\\cert.pfx'}</Typography>
                    <Typography variant='body2'>
                        The abuse is possible with the strong explicit certificate mappings X509IssuerSerialNumber or
                        X509SHA1PublicKey. In this example, we use X509SHA1PublicKey.
                    </Typography>
                    <Typography variant='body2'>
                        Get the SHA1 hash of the certificate public key using certutil:
                    </Typography>
                    <CodeController>
                        {`certutil.exe -dump -v .\\cert.pfx
…
Cert Hash(sha1): ef9375785421d3ad286d8bdeb166f0f697266992
…`}
                    </CodeController>
                    <Typography variant='body2'>
                        Use Add-AltSecIDMapping to add the explicit certificate mapping string to the
                        'altSecurityIdentities' attribute of the target principal:
                    </Typography>
                    <Typography component={'pre'}>
                        {
                            'Add-AltSecIDMapping -DistinguishedName "CN=Target,CN=Users,DC=forestroot,DC=com" -MappingString "X509:<SHA1-PUKEY>ef9375785421d3ad286d8bdeb166f0f697266992"'
                        }
                    </Typography>
                    <Typography variant='body2'>
                        Verify the that the mapping was added using Get-AltSecIDMapping:
                    </Typography>
                    <Typography component={'pre'}>
                        {'Get-AltSecIDMapping -SearchBase "CN=Target,CN=Users,DC=forestroot,DC=com"'}
                    </Typography>
                    <Typography variant='body2'>
                        Use Rubeus to request a ticket granting ticket (TGT) from the domain, specifying the target
                        identity, and the PFX-formatted certificate:
                    </Typography>
                    <Typography component={'pre'}>
                        {'Rubeus asktgt /user:"forestroot\\target" /certificate:cert.pfx /ptt'}
                    </Typography>
                    <Typography variant='body2'>
                        After the execution of the abuse, use Remove-AltSecIDMapping to remove the explicit certificate
                        mapping string from the 'altSecurityIdentities' attribute of the target principal:
                    </Typography>
                    <Typography component={'pre'}>
                        {
                            'Remove-AltSecIDMapping -DistinguishedName "CN=Target,CN=Users,DC=forestroot,DC=com" -MappingString "X509:<SHA1-PUKEY>ef9375785421d3ad286d8bdeb166f0f697266992"'
                        }
                    </Typography>
                </>
            );
        case 'OU':
        case 'Domain':
            return (
                <>
                    <Typography variant='body2'>
                        An attacker with permission to modify the gPLink attribute can link GPOs to the object,
                        affecting all contained users and computers. The GPO can be weaponized by injecting a malicious
                        configuration, such as a scheduled task executing a malicious script.
                    </Typography>
                    <Typography variant='body2'>
                        The GPO can be linked as enforced to bypass blocked GPO inheritance. WMI or security filtering
                        can be used to limit the impact to specific accounts, which is important in environments with
                        many users or computers under the affected scope.
                    </Typography>
                    <Typography variant='body2'>
                        Refer to{' '}
                        <Link target='_blank' rel='noopener' href='https://wald0.com/?p=179'>
                            A Red Teamer's Guide to GPOs and OUs
                        </Link>
                        for details about the abuse technique, and check out{' '}
                        <Link target='_blank' rel='noopener' href='https://github.com/FSecureLABS/SharpGPOAbuse'>
                            SharpGPOAbuse
                        </Link>{' '}
                        for practical exploitation.
                    </Typography>
                    <Typography variant='body2'>
                        <b>Without control over a GPO</b>
                        <br />
                        An attacker can still execute the attack without control over a GPO by setting up a fake LDAP
                        server to host a GPO. This approach requires the ability to add non-existent DNS records and
                        create machine accounts, or access to a compromised domain-joined machine. However, this method
                        is complex and requires significant setup.
                        <br />
                        <br />
                        From a domain-joined compromised Windows machine, the write access to the gPLink attribute may
                        be abused through Powermad, PowerView and native Windows functionalities. For a detailed outline
                        of exploit requirements and implementation, you can refer to this article:{' '}
                        <Link
                            target='_blank'
                            rel='noopener'
                            href='https://labs.withsecure.com/publications/ou-having-a-laugh'>
                            OU having a laugh?
                        </Link>
                    </Typography>
                </>
            );
        case 'CertTemplate':
            return (
                <>
                    <Typography variant='body2'>
                        With GenericWrite permission over a certificate template, you may be able to perform an ESC4
                        attack by modifying the template's attributes. BloodHound will in that case create an ADCSESC4
                        edge from the principal to the forest domain node.
                    </Typography>
                </>
            );
        case 'EnterpriseCA':
            return (
                <>
                    <Typography variant='body2'>
                        With GenericWrite permission over an enterprise CA, you can publish certificate templates to the
                        enterprise CA by adding the CN name of the template in the enterprise CA object's
                        certificateTemplates attribute. This action may enable you to perform an ADCS domain escalation.
                    </Typography>
                </>
            );
        case 'RootCA':
            return (
                <>
                    <Typography variant='body2'>
                        With GenericWrite permission over a root CA, you can make a rogue certificate trusted as a root
                        CA in the AD forest by adding the certificate in the root CA object's cACertificate attribute.
                        This action may enable you to perform an ADCS domain escalation.
                    </Typography>
                </>
            );
        case 'NTAuthStore':
            return (
                <>
                    <Typography variant='body2'>
                        With GenericWrite permission over a NTAuth store, you can make an enterprise CA certificate
                        trusted for NT (domain) authentication in the AD forest by adding the certificate in the root CA
                        object's cACertificate attribute. This action may enable you to perform an ADCS domain
                        escalation.
                    </Typography>
                </>
            );
        case 'IssuancePolicy':
            return (
                <>
                    <Typography variant='body2'>
                        With GenericWrite permission over an issuance policy object, you create a OID group link to a
                        targeted group by adding the group's distinguishedName in the msDS-OIDToGroupLink attribute of
                        the issuance policy object. This action may enable you to gain membership of the group through
                        an ADCS ESC13 attack.
                    </Typography>
                </>
            );
        default:
            return (
                <>
                    <Typography variant='body2'>No abuse information available for this node type.</Typography>
                </>
            );
    }
};

export default WindowsAbuse;
