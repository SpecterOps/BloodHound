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

import { Box, Typography } from '@mui/material';
import { PageWithTitle, CitrixRDPConfiguration } from 'bh-shared-ui';

const BloodHoundConfiguration = () => {
    return (
        <PageWithTitle
            title='Bloodhound Configuration'
            pageDescription={
                <Typography variant='body2' paragraph>
                    Brief Description of the Feature Lorem ipsum dolor sit amet, consectetur adipiscing elit, sed do
                    eiusmod tempor incididunt ut labore et dolore magna aliqua. Ut enim ad minim veniam, quis nostrud
                    exercitation ullamco laboris nisi ut aliquip ex ea commodo consequat.
                </Typography>
            }>
            <Box sx={{ display: 'flex', flexDirection: 'column', gap: '24px' }}>
                <CitrixRDPConfiguration />
            </Box>
        </PageWithTitle>
    );
};

export default BloodHoundConfiguration;
