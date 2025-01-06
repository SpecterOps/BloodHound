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

import { Alert, AlertTitle, Box, useTheme } from '@mui/material';

const WebGLDisabledAlert = () => {
    const theme = useTheme();
    return (
        <Box display={'flex'} justifyContent={'center'} mt={theme.spacing(8)} mx={theme.spacing(4)}>
            <Alert severity={'error'}>
                <AlertTitle>WebGL Not Supported</AlertTitle>
                <p>
                    This page requires WebGL to render. Please enable WebGL in your browser settings or switch to a
                    browser that supports this feature.
                </p>
            </Alert>
        </Box>
    );
};

export default WebGLDisabledAlert;
