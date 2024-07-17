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

import React, { FC } from 'react';
import { Link, Box } from '@mui/material';

const References: FC = () => {
    const references = [
        {
            label: 'Concept SSPR WriteBack',
            link: 'https://learn.microsoft.com/en-us/entra/identity/authentication/concept-sspr-writeback',
        },
    ];
    return (
        <Box sx={{ overflowX: 'auto' }}>
            {references.map((reference) => {
                return (
                    <React.Fragment key={reference.link}>
                        <Link target='_blank' rel='noopener' href={reference.link}>
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
