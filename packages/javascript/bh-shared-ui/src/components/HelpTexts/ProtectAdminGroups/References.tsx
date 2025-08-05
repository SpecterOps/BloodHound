// Copyright 2025 Specter Ops, Inc.
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

import { Box, Link } from '@mui/material';
import { FC } from 'react';

const References: FC = () => {
    return (
        <Box sx={{ overflowX: 'auto' }}>
            <Link target='_blank' rel='noopener' href='https://specterops.io/resources/adminsdholder'>
                SpecterOps: AdminSDHolder: Misconceptions and Myths
            </Link>
            <br />
            <Link target='_blank' rel='noopener' href='https://secureidentity.se/adminsdholder-pitfalls-and-misunderstandings/'>
                Secure Identity: AdminSDHolder - pitfalls and misunderstandings
            </Link>
            <br />
            <Link target='_blank' rel='noopener' href='https://secureidentity.se/adminsdholder-pt2/'>
                Secure Identity: Where the adminCount doesn't count and the SD isn't what you thought
            </Link>
        </Box>
    );
};

export default References;