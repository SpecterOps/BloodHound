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
            <Typography variant='body2' className={classes.containsCodeEl}>
                {groupSpecialFormat(sourceType, sourceName)} has the privileges to perform the ADCS ESC9 Scenario B
                attack against the target domain.
                <br />
                <br />
                The principal has control over a victim computer with permission to enroll on one or more certificate
                templates, configured to: 1) enable certificate authentication, 2) require the <code>dNSHostName</code>
                &nbsp;of the enrollee included in the Subject Alternative Name (SAN), and 3) not have the security
                extension enabled. The victim computer also has enrollment permission for an enterprise CA with the
                necessary templates published. This enterprise CA is trusted for NT authentication in the forest, and
                chains up to a root CA for the forest. There is an affected Domain Controller (DC) configured to allow
                weak certificate binding enforcement. This setup lets the principal impersonate any AD forest computer
                without their credentials.
                <br />
                <br />
                The attacker principal can abuse their control over the victim computer to modify the victim computer's{' '}
                <code>dNSHostName</code> attribute to match the <code>dNSHostName</code> of a targeted computer. The
                attacker principal will then abuse their control over the victim computer to obtain the credentials of
                the victim computer, or a session as the victim computer, and enroll a certificate as the victim in one
                of the affected certificate templates. The <code>dNSHostName</code> of the victim will be included in
                the issued certificate under SAN DNS name. As the certificate template does not have the security
                extension, the issued certificate will NOT include the SID of the victim computer. DCs with strong
                certificate binding configuration will require a SID to be present in a certificate used for Kerberos
                authentication, but the affected DCs with weak certificate binding configuration will not. The affected
                DCs will split the SAN DNS name into a computer name and a domain name, confirm that the domain name is
                correct, and use the computer name appended a $ to identify principals with a matching{' '}
                <code>sAMAccountName</code>. At last, the DC issues a Kerberos TGT as the targeted computer to the
                attacker, which means the attacker now has a session as the targeted computer.
            </Typography>
        </>
    );
};

export default General;
