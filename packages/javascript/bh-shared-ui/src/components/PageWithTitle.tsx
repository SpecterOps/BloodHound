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

import React, { DataHTMLAttributes } from 'react';
import { Box, Container, Typography } from '@mui/material';

interface PageWithTitleProps extends DataHTMLAttributes<HTMLDivElement> {
    title?: string;
    children?: React.ReactNode;
}

const PageWithTitle: React.FC<PageWithTitleProps> = ({ title, children, ...rest }) => {
    return (
        <Container maxWidth='xl' {...rest}>
            <Box py={2}>
                {title && (
                    <Box mb={2}>
                        <Typography variant='h1'>{title}</Typography>
                    </Box>
                )}
                {children}
            </Box>
        </Container>
    );
};

export default PageWithTitle;
