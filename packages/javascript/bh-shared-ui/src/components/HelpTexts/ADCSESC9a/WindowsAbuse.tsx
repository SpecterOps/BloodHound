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
import { Box, Typography } from '@mui/material';

const WindowsAbuse: FC = () => {
    const steps = [
        {
            synopsis: 'Set UPN of victim to targeted principal’s sAMAccountName.',
            description: 'Set the UPN of the victim principal using PowerView:',
            command: "Set-DomainObject -Identity VICTIM -Set @{'userprincipalname'='Target'}",
        },
        {
            synopsis: 'Check if mail attribute of victim must be set and set it if required.',
            description: `If the certificate template is of schema version 2 or above and its attribute msPKI-CertificateNameFlag contains the flag SUBJECT_REQUIRE_EMAIL and/or SUBJECT_ALT_REQUIRE_EMAIL then the victim principal must have their mail attribute set for the certificate enrollment. The CertTemplate BloodHound node will have “Subject Require Email" or “Subject Alternative Name Require Email" set to true if any of the flags are present.

            If the certificate template is of schema version 1 or does not have any of the email flags, then continue to Step 3.
            
            If any of the two flags are present, you will need the victim’s mail attribute to be set. The value of the attribute will be included in the issues certificate but it is not used to identify the target principal why it can be set to any arbitrary string.
            
            Check if the victim has the mail attribute set using PowerView:`,
            command: "Set-DomainObject -Identity VICTIM -Set @{'userprincipalname'='Target'}",
        },
        {
            synopsis: 'Set UPN of victim to targeted principal’s sAMAccountName.',
            description: 'Set the UPN of the victim principal using PowerView:',
            command: "Set-DomainObject -Identity VICTIM -Set @{'userprincipalname'='Target'}",
        },
        {
            synopsis: 'Set UPN of victim to targeted principal’s sAMAccountName.',
            description: 'Set the UPN of the victim principal using PowerView:',
            command: "Set-DomainObject -Identity VICTIM -Set @{'userprincipalname'='Target'}",
        },
    ];
    return (
        <>
            <Typography variant='body2'>An attacker may perform this attack in the following steps:</Typography>
            {steps.map((step, i) => {
                return (
                    <>
                        <Typography variant='body2'>
                            <Box component='span' sx={{ fontWeight: 'bold' }}>
                                Step {i + 1}: {step.synopsis}
                            </Box>{' '}
                            {step.description}
                        </Typography>
                        <Typography component={'pre'}>{step.command}</Typography>
                    </>
                );
            })}
        </>
    );
};

export default WindowsAbuse;
