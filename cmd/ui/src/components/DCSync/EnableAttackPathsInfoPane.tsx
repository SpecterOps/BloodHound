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

import { Box, Link, Paper, Typography, SxProps } from '@mui/material';
import React from 'react';

interface EnableAttackPathsInfoPaneProps {
    sx?: SxProps;
}

const EnableAttackPathsInfoPane: React.FC<EnableAttackPathsInfoPaneProps> = ({ sx }) => {
    console.log('EnableAttackPathsInfoPane: Rendering');
    
    return (
        <Box sx={sx}>
            <Paper sx={{ height: '100%', overflow: 'auto', p: 2 }}>
                <Typography variant="h5" sx={{ mb: 2, fontWeight: 'bold', color: 'primary.main' }}>
                    Enable DCSync Paths
                </Typography>
                
                <Typography variant="body1" sx={{ mb: 2 }}>
                    The principal has attack paths that can grant it DCSync permissions on the target domain.
                </Typography>
                
                <Typography variant="h6" sx={{ mb: 1, fontWeight: 'bold' }}>
                    Abuse
                </Typography>
                
                <Typography variant="body1" sx={{ mb: 2 }}>
                    A DCSync attack requires both the GetChanges and GetChangesAll permissions. Execute the attack
                    paths that result in obtaining these permissions on the target domain.
                </Typography>
                
                <Typography variant="body1" sx={{ mb: 2 }}>
                    For details on performing a DCSync attack, refer to the{' '}
                    <Link 
                        href="https://bloodhound.readthedocs.io/en/latest/data-analysis/edges.html#dcsync" 
                        target="_blank" 
                        rel="noopener noreferrer"
                        color="primary">
                        DCSync edge documentation
                    </Link>
                    .
                </Typography>
                
                <Typography variant="h6" sx={{ mb: 1, fontWeight: 'bold' }}>
                    Detection
                </Typography>
                
                <Typography variant="body1">
                    Monitor for Event ID 4662 (An operation was performed on an object) with Object Access of 
                    "Replicating Directory Changes" and "Replicating Directory Changes All" on domain objects.
                </Typography>
            </Paper>
        </Box>
    );
};

export default EnableAttackPathsInfoPane;
