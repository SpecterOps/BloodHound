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
import { Link, Box } from '@mui/material';

const References: FC = () => {
    return (
        <Box sx={{ overflowX: 'auto' }}>
            <Link
                target='_blank'
                rel='noopener'
                href='https://github.com/PowerShellMafia/PowerSploit/blob/dev/Recon/PowerView.ps1'>
                https://github.com/PowerShellMafia/PowerSploit/blob/dev/Recon/PowerView.ps1
            </Link>
            <br />
            <Link target='_blank' rel='noopener' href='http://www.selfadsi.org/deep-inside/ad-security-descriptors.htm'>
                http://www.selfadsi.org/deep-inside/ad-security-descriptors.htm
            </Link>
            <br />
            <Link target='_blank' rel='noopener' href='https://eladshamir.com/2019/01/28/Wagging-the-Dog.html'>
                https://eladshamir.com/2019/01/28/Wagging-the-Dog.html
            </Link>
            <br />
            <Link target='_blank' rel='noopener' href='https://github.com/GhostPack/Rubeus#s4u'>
                https://github.com/GhostPack/Rubeus#s4u
            </Link>
            <br />
            <Link
                target='_blank'
                rel='noopener'
                href='https://gist.github.com/HarmJ0y/224dbfef83febdaf885a8451e40d52ff'>
                https://gist.github.com/HarmJ0y/224dbfef83febdaf885a8451e40d52ff
            </Link>
            <br />
            <Link target='_blank' rel='noopener' href='https://blog.harmj0y.net/redteaming/another-word-on-delegation/'>
                https://blog.harmj0y.net/redteaming/another-word-on-delegation/
            </Link>
            <br />
            <Link
                target='_blank'
                rel='noopener'
                href='https://github.com/PowerShellMafia/PowerSploit/blob/dev/Recon/PowerView.ps1'>
                https://github.com/PowerShellMafia/PowerSploit/blob/dev/Recon/PowerView.ps1
            </Link>
            <br />
            <Link target='_blank' rel='noopener' href='https://github.com/Kevin-Robertson/Powermad#new-machineaccount'>
                https://github.com/Kevin-Robertson/Powermad#new-machineaccount
            </Link>
            <br />
            <Link
                target='_blank'
                rel='noopener'
                href='https://docs.microsoft.com/en-us/dotnet/api/system.directoryservices.activedirectorysecurityinheritance?view=netframework-4.8'>
                https://docs.microsoft.com/en-us/dotnet/api/system.directoryservices.activedirectorysecurityinheritance?view=netframework-4.8
            </Link>
            <br />
            <Link target='_blank' rel='noopener' href='https://github.com/eladshamir/Whisker'>
                https://github.com/eladshamir/Whisker
            </Link>
            <br />
            <Link
                target='_blank'
                rel='noopener'
                href='https://posts.specterops.io/shadow-credentials-abusing-key-trust-account-mapping-for-takeover-8ee1a53566ab'>
                https://posts.specterops.io/shadow-credentials-abusing-key-trust-account-mapping-for-takeover-8ee1a53566ab
            </Link>
            <br />
            <Link
                target='_blank'
                rel='noopener'
                href='https://specterops.io/wp-content/uploads/sites/3/2022/06/Certified_Pre-Owned.pdf'>
                https://specterops.io/wp-content/uploads/sites/3/2022/06/Certified_Pre-Owned.pdf
            </Link>
            <br />
            <Link
                target='_blank'
                rel='noopener'
                href='https://decoder.cloud/2023/11/20/a-deep-dive-in-cert-publishers-group/'>
                https://decoder.cloud/2023/11/20/a-deep-dive-in-cert-publishers-group/
            </Link>
            <br />
            <Link
                target='_blank'
                rel='noopener'
                href='https://posts.specterops.io/adcs-esc13-abuse-technique-fda4272fbd53'>
                https://posts.specterops.io/adcs-esc13-abuse-technique-fda4272fbd53
            </Link>
            <br />
            <Link
                target='_blank'
                rel='noopener'
                href='https://support.microsoft.com/en-us/topic/kb5008383-active-directory-permissions-updates-cve-2021-42291-536d5555-ffba-4248-a60e-d6cbc849cde1'>
                https://support.microsoft.com/en-us/topic/kb5008383-active-directory-permissions-updates-cve-2021-42291-536d5555-ffba-4248-a60e-d6cbc849cde1
            </Link>
            <br />
            <Link
                target='_blank'
                rel='noopener'
                href='https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-adts/fb7c101d-ec8b-4fbf-bca8-7d7c2d747d0c'>
                https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-adts/fb7c101d-ec8b-4fbf-bca8-7d7c2d747d0c
            </Link>
            <br />
            <Link
                target='_blank'
                rel='noopener'
                href='https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-adts/e5899be4-862e-496f-9a38-33950617d2c5'>
                https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-adts/e5899be4-862e-496f-9a38-33950617d2c5
            </Link>
            <br />
            <Link
                target='_blank'
                rel='noopener'
                href='https://learn.microsoft.com/en-us/windows-server/identity/ad-ds/manage/understand-security-identifiers'>
                https://learn.microsoft.com/en-us/windows-server/identity/ad-ds/manage/understand-security-identifiers
            </Link>
        </Box>
    );
};

export default References;
