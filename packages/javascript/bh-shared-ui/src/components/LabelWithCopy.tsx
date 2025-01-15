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

import { Box, IconButton, Tooltip, useTheme } from '@mui/material';
import { faCopy } from '@fortawesome/free-solid-svg-icons';
import { FontAwesomeIcon } from '@fortawesome/react-fontawesome';
import { FC, useState } from 'react';
import { copyToClipboard } from '../utils';

const LabelWithCopy: FC<{
    label: string;
    valueToCopy: string | number;
    hoverOnly?: boolean;
}> = ({ label, valueToCopy, hoverOnly = false }) => {
    const theme = useTheme();
    const [copied, setCopied] = useState(false);
    const [hoverActive, setHoverActive] = useState(false);

    const handleMouseEnter = () => setHoverActive(true);
    const handleMouseLeave = () => setHoverActive(false);

    const handleCopy = async () => {
        setCopied(true);

        await copyToClipboard(valueToCopy);

        setTimeout(() => {
            setCopied(false);
        }, 1000);
    };

    return (
        <Box
            display='flex'
            alignItems='center'
            onMouseEnter={handleMouseEnter}
            onMouseLeave={handleMouseLeave}
            sx={{ height: theme.spacing(3) }}>
            {label}
            <Tooltip title='Copied' open={copied} placement='right'>
                <IconButton
                    onClick={handleCopy}
                    sx={{ fontSize: 'inherit', display: !hoverOnly || hoverActive ? undefined : 'none' }}>
                    <FontAwesomeIcon icon={faCopy} />
                </IconButton>
            </Tooltip>
        </Box>
    );
};

export default LabelWithCopy;
