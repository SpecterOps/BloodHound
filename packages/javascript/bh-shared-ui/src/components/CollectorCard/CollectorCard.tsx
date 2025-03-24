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

import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@bloodhoundenterprise/doodleui';
import { Box, Link, Typography } from '@mui/material';
import { CollectorType } from 'js-client-library';
import React from 'react';
import { cn } from '../../utils';

export type LabelType = 'latest' | 'prerelease';

export const COLLECTOR_TYPE_LABEL: { [key in CollectorType]: string } = {
    sharphound_enterprise: 'SharpHound Enterprise',
    sharphound: 'SharpHound Community',
    azurehound: 'AzureHound',
};

interface CollectorDownloadFile {
    displayName: string;
    os: string;
    arch: string;
    onClickDownload: () => void;
    onClickDownloadChecksum: () => void;
}

interface CollectorCardProps extends React.HTMLAttributes<HTMLDivElement> {
    collectorType: CollectorType;
    version: string;
    downloadArtifacts: CollectorDownloadFile[];
    timestamp?: number;
    label?: LabelType;
    isPrerelease?: boolean;
}

const CollectorCard: React.FC<CollectorCardProps> = ({
    collectorType,
    version,
    downloadArtifacts,
    timestamp = undefined,
    label = undefined,
    isPrerelease = false,
    ...rest
}) => {
    const date = timestamp ? new Date(timestamp) : null;

    return (
        <Card {...rest}>
            <CardHeader>
                <Box display='flex' flexDirection='row' alignItems='center' overflow='hidden' gap='1rem'>
                    <CardTitle>{`${version}`.trim().toUpperCase()}</CardTitle>
                    {isPrerelease && <Typography variant='caption'>(pre-release)</Typography>}
                    {label && <CollectorLabel label={label} isPrerelease={isPrerelease} />}
                    {date && (
                        <CardDescription>
                            {`${date.getFullYear()}.${date.getMonth() + 1}.${date.getDate()}`}
                        </CardDescription>
                    )}
                </Box>
            </CardHeader>
            <CardContent>
                <Typography variant='body1' component='div'>
                    <ul>
                        {downloadArtifacts.map((collector) => (
                            <li key={collector.displayName}>
                                <Box display='flex' flexDirection='row'>
                                    <Link
                                        component='button'
                                        variant='body1'
                                        onClick={collector.onClickDownload}
                                        title={`Download ${COLLECTOR_TYPE_LABEL[collectorType]} ${version} ${collector.os} ${collector.arch}`.trim()}
                                        textOverflow='ellipsis'
                                        noWrap>
                                        {collector.displayName}
                                    </Link>
                                    <DashedFlexConnector />
                                    <Link
                                        component='button'
                                        variant='body1'
                                        onClick={collector.onClickDownloadChecksum}
                                        title={`Download ${COLLECTOR_TYPE_LABEL[collectorType]} ${version} ${collector.os} ${collector.arch} Checksum`.trim()}>
                                        (checksum)
                                    </Link>
                                </Box>
                            </li>
                        ))}
                    </ul>
                </Typography>
            </CardContent>
        </Card>
    );
};

interface CollectorLabelProps {
    label: LabelType;
    isPrerelease?: boolean;
}

const CollectorLabel: React.FC<CollectorLabelProps> = ({ label, isPrerelease = false }) => {
    // RiskBadge isn't set up out the gate to accomodate dark mode
    // so we improvise
    return (
        <div
            className={cn('w-auto rounded-full px-6 py-2 border-none text-center', {
                'bg-[#F3603640] dark:bg-[#02C577]': isPrerelease,
                'bg-[#33318F26] dark:bg-[#33318F]': !isPrerelease,
            })}>
            {label}
        </div>
    );
};

// Sometimes you just need a dashed line connecting two underlined text elements together
const DashedFlexConnector: React.FC = () => {
    return (
        <div
            className={cn(
                'flex-1 min-w-0 decoration-dashed decoration-neutral-light-5 underline whitespace-nowrap text-clip overflow-hidden pointer-events-none select-none aria-hidden'
            )}>
            {'\u00A0'.repeat(500)}
        </div>
    );
};

export default CollectorCard;
