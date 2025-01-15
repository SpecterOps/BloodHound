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
            <Link target='_blank' rel='noopener' href='https://blog.harmj0y.net/redteaming/the-trustpocalypse/'>
                https://blog.harmj0y.net/redteaming/the-trustpocalypse/
            </Link>
            <br />
            <Link
                target='_blank'
                rel='noopener'
                href='https://blog.harmj0y.net/redteaming/a-guide-to-attacking-domain-trusts/'>
                https://blog.harmj0y.net/redteaming/a-guide-to-attacking-domain-trusts/
            </Link>
            <br />
            <Link target='_blank' rel='noopener' href='https://adsecurity.org/?p=1772'>
                https://adsecurity.org/?p=1772
            </Link>
            <br />
            <Link target='_blank' rel='noopener' href='https://adsecurity.org/?tag=sidhistory'>
                https://adsecurity.org/?tag=sidhistory
            </Link>
            <br />
            <Link target='_blank' rel='noopener' href='https://attack.mitre.org/techniques/T1178/'>
                https://attack.mitre.org/techniques/T1178/
            </Link>
            <br />
            <Link
                target='_blank'
                rel='noopener'
                href='https://dirkjanm.io/active-directory-forest-trusts-part-one-how-does-sid-filtering-work/'>
                https://dirkjanm.io/active-directory-forest-trusts-part-one-how-does-sid-filtering-work/
            </Link>
        </Box>
    );
};

export default References;
