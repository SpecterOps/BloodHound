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

import { Box, Divider, Typography } from '@mui/material';
import React, { DataHTMLAttributes } from 'react';

interface ContentPageProps extends DataHTMLAttributes<HTMLDivElement> {
    title?: string;
    children?: React.ReactNode;
    actionButton?: React.ReactNode;
}

const ContentPage: React.FC<ContentPageProps> = ({ title, children, actionButton, ...rest }) => {
    return (
        <div {...rest}>
            {title && (
                <>
                    <Box display='flex' justifyContent='space-between'>
                        <Typography variant='h1'>{title}</Typography>
                        {actionButton}
                    </Box>
                    <Box mt={2} mb={4}>
                        <Divider />
                    </Box>
                </>
            )}
            {children}
        </div>
    );
};

export default ContentPage;
