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

import { faCopy } from '@fortawesome/free-solid-svg-icons';
import { FontAwesomeIcon } from '@fortawesome/react-fontawesome';
import { IconButton, Tooltip } from '@mui/material';
import { FC, ReactNode, useState } from 'react';
import { cn, copyToClipboard } from '../utils';

const LabelWithCopy: FC<{
    label: ReactNode;
    valueToCopy: string | number;
    hoverOnly?: boolean;
    className?: string;
}> = ({ label, valueToCopy, hoverOnly = false, className }) => {
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
        <div
            onMouseEnter={handleMouseEnter}
            onMouseLeave={handleMouseLeave}
            className={cn('h-6 flex items-center', className)}>
            {label}
            <Tooltip title='Copied' open={copied} placement='right'>
                <IconButton
                    onClick={handleCopy}
                    sx={{ fontSize: 'inherit', display: !hoverOnly || hoverActive ? undefined : 'none' }}>
                    <FontAwesomeIcon icon={faCopy} />
                </IconButton>
            </Tooltip>
        </div>
    );
};

export default LabelWithCopy;
