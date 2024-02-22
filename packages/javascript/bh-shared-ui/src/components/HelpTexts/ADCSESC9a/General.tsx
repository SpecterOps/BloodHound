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
import { useHelpTextStyles, groupSpecialFormat } from '../utils';
import { EdgeInfoProps } from '../index';
import { Typography } from '@mui/material';

const General: FC<EdgeInfoProps> = ({ sourceName, sourceType, targetName }) => {
    const classes = useHelpTextStyles();
    return (
        <>
            <Typography variant='body2'>
                {groupSpecialFormat(sourceType, sourceName)} the privileges to perform the ADCS ESC9 Scenario A attack
                against the target domain.
            </Typography>
            <Typography variant='body2' className={classes.containsCodeEl}>
                The principal has control over a victim principal with permission to enroll on one or more certificate
                templates, configured to: 1) enable certificate authentication, 2) require the{' '}
                <code>userPrincipalName</code> (UPN) of the enrollee included in the Subject Alternative Name (SAN), and
                3) do not have the security extension enabled. The victim also has enrollment permission for an
                enterprise CA with the necessary templates published. This enterprise CA is trusted for NT
                authentication in the forest, and chains up to a root CA for the forest. There is an affected Domain
                Controller (DC) configured to allow weak certificate binding enforcement. This setup lets the principal
                impersonate any AD forest principal (user or computer) without their credentials.
            </Typography>
            <Typography variant='body2' className={classes.containsCodeEl}>
                The attacker principal can abuse their control over the victim principal to modify the victim’s UPN to
                match the <code>sAMAccountName</code> of a targeted principal. Example: If the targeted principal is
                Administrator@corp.local user, the victim's UPN will be populated with "Administrator" (without the
                @corp.local ending). The attacker principal will then abuse their control over the victim principal to
                obtain the credentials of the victim principal, or a session as the victim principal, and enroll a
                certificate as the victim in one of the affected certificate templates. The UPN of the victim
                ("Administrator") will be included in the issued certificate under the SAN. As the certificate template
                does not have the security extension, it will NOT include the SID of the victim user in the issued
                certificate. Next, the attacker principal will again set the UPN of the victim, this time to an
                arbitrary string (e.g. the original value). The issued certificate can now be used for authentication
                against an affected DC. The weak certificate binding configuration on the DC will make the DC accept
                that the SID of the victim user is not present in the issued certificate when performing Kerberos
                authentication, and it will use the SAN value to map the certificate to a principal. The DC will attempt
                to find a principal with a UPN matching the SAN value (“Administrator”) but as the victim’s UPN has been
                changed after the enrollment, there will be no principals with this UPN. The DC will then attempt to
                find a principal with a <code>sAMAccountName</code> matching the SAN value and find the targeted user.
                At last, the DC issues a Kerberos TGT as the targeted user to the attacker, which means the attacker now
                has a session as the targeted user. In case the target is a computer, the DC will find it as well as the
                DC will attempt <code>sAMAccountName</code> matching with a $ at the end of the SAN value as last
                resort.
            </Typography>
        </>
    );
};

export default General;
