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

import { Box, Typography } from '@mui/material';
import { PageWithTitle } from 'bh-shared-ui';

const SniffDeep = () => {
    return (
        <PageWithTitle title="Sniff Deep" data-testid="sniff-deep-page">
            <Box
                sx={{
                    display: 'flex',
                    flexDirection: 'column',
                    alignItems: 'center',
                    justifyContent: 'center',
                    height: '60vh',
                    textAlign: 'center',
                }}
            >
                <Typography variant="h4" component="h1" gutterBottom>
                    Sniff Deep
                </Typography>
                <Typography variant="body1" color="text.secondary">
                    This is an empty tab ready for your content.
                </Typography>
            </Box>
        </PageWithTitle>
    );
};

export default SniffDeep;
