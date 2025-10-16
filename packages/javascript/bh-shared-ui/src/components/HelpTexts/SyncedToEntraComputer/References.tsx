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

import { Box, Link } from '@mui/material';
import React, { FC } from 'react';

const References: FC = () => {
    const references = [
        {
            label: 'How it works: Device Registration',
            link: 'https://learn.microsoft.com/en-us/entra/identity/devices/device-registration-how-it-works',
        },
        {
            label: "An Operator's Guide to Device-Joined Hosts and the PRT Cookie",
            link: 'https://posts.specterops.io/an-operators-guide-to-device-joined-hosts-and-the-prt-cookie-bcd0db2812c4',
        },
        {
            label: 'Entra Connect Attacker Tradecraft: Part 3',
            link: 'https://specterops.io/blog/2025/07/30/entra-connect-attacker-tradecraft-part-3/',
        },
    ];
    return (
        <Box sx={{ overflowX: 'auto' }}>
            {references.map((reference) => {
                return (
                    <React.Fragment key={reference.link}>
                        <Link target='_blank' rel='noopener noreferrer' href={reference.link}>
                            {reference.label}
                        </Link>
                        <br />
                    </React.Fragment>
                );
            })}
        </Box>
    );
};

export default References;
