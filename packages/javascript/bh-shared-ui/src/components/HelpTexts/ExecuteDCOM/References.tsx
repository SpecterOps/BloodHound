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
import { Link, Typography, Box } from '@mui/material';

const References: FC = () => {
    return (
        <Box sx={{ overflowX: 'auto' }}>
            <Link
                target='_blank'
                rel='noopener'
                href='https://enigma0x3.net/2017/01/05/lateral-movement-using-the-mmc20-application-com-object/'>
                https://enigma0x3.net/2017/01/05/lateral-movement-using-the-mmc20-application-com-object/
            </Link>
            <br />
            <Link
                target='_blank'
                rel='noopener'
                href='https://enigma0x3.net/2017/01/23/lateral-movement-via-dcom-round-2/'>
                https://enigma0x3.net/2017/01/23/lateral-movement-via-dcom-round-2/
            </Link>
            <br />
            <Link
                target='_blank'
                rel='noopener'
                href='https://enigma0x3.net/2017/09/11/lateral-movement-using-excel-application-and-dcom/'>
                https://enigma0x3.net/2017/09/11/lateral-movement-using-excel-application-and-dcom/
            </Link>
            <br />
            <Link
                target='_blank'
                rel='noopener'
                href='https://enigma0x3.net/2017/11/16/lateral-movement-using-outlooks-createobject-method-and-dotnettojscript/'>
                https://enigma0x3.net/2017/11/16/lateral-movement-using-outlooks-createobject-method-and-dotnettojscript/
            </Link>
            <br />
            <Link
                target='_blank'
                rel='noopener'
                href='https://www.cybereason.com/blog/leveraging-excel-dde-for-lateral-movement-via-dcom '>
                https://www.cybereason.com/blog/leveraging-excel-dde-for-lateral-movement-via-dcom
            </Link>
            <br />
            <Link
                target='_blank'
                rel='noopener'
                href='https://www.cybereason.com/blog/dcom-lateral-movement-techniques'>
                https://www.cybereason.com/blog/dcom-lateral-movement-techniques
            </Link>
            <br />
            <Link
                target='_blank'
                rel='noopener'
                href='https://bohops.com/2018/04/28/abusing-dcom-for-yet-another-lateral-movement-technique/'>
                https://bohops.com/2018/04/28/abusing-dcom-for-yet-another-lateral-movement-technique/
            </Link>
            <br />
            <Link target='_blank' rel='noopener' href='https://attack.mitre.org/wiki/Technique/T1175'>
                https://attack.mitre.org/wiki/Technique/T1175
            </Link>

            <Typography variant='body1'>Invoke-DCOM</Typography>
            <Link
                target='_blank'
                rel='noopener'
                href='https://github.com/rvrsh3ll/Misc-Powershell-Scripts/blob/master/Invoke-DCOM.ps1'>
                https://github.com/rvrsh3ll/Misc-Powershell-Scripts/blob/master/Invoke-DCOM.ps1
            </Link>

            <Typography variant='body1'>LethalHTA</Typography>
            <Link target='_blank' rel='noopener' href='https://codewhitesec.blogspot.com/2018/07/lethalhta.html'>
                https://codewhitesec.blogspot.com/2018/07/lethalhta.html
            </Link>
            <br />
            <Link target='_blank' rel='noopener' href='https://github.com/codewhitesec/LethalHTA/ '>
                https://github.com/codewhitesec/LethalHTA/
            </Link>
        </Box>
    );
};

export default References;
