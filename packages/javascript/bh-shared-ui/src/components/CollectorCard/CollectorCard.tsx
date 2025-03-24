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

import { Card, CardContent, CardDescription, CardHeader, CardTitle, RiskBadge } from '@bloodhoundenterprise/doodleui';
import { Box, Link, Typography } from '@mui/material';
import React from 'react';
import { CollectorType } from 'js-client-library/src/types';
import { cn } from '../../utils';

export type LabelType = 'latest' | 'prerelease';

export const COLLECTOR_TYPE_LABEL: { [key in CollectorType]: string } = {
    sharphound_enterprise: 'SharpHound Enterprise',
    sharphound: 'SharpHound Community',
    azurehound: 'AzureHound',
};

interface CollectorDownloadFile {
    fileName: string;
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
    needsUnderlineOffset?: boolean;
}

const CollectorCard: React.FC<CollectorCardProps> = ({
    collectorType,
    version,
    downloadArtifacts,
    timestamp = undefined,
    label = undefined,
    isPrerelease = false,
    needsUnderlineOffset = false,
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
                    {date && <CardDescription>
                        {`${date.getFullYear()}.${date.getMonth() + 1}.${date.getDate()}`}
                    </CardDescription>}
                </Box>
            </CardHeader>
            <CardContent>
                <ul>
                    {downloadArtifacts.map((collector) => (
                        <li key={collector.fileName}>
                            <Box display='flex' flexDirection='row'>
                                <Link
                                    component='button'
                                    variant='body1'
                                    onClick={collector.onClickDownload}
                                    title={`Download ${COLLECTOR_TYPE_LABEL[collectorType]} ${version} ${collector.os} ${collector.arch}`.trim()}
                                    textOverflow='ellipsis'
                                    noWrap>
                                    {collector.fileName}
                                </Link>
                                <DashedFlexConnector offset={needsUnderlineOffset} />
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
    const color = isPrerelease ? 'rgba(243, 96, 54, 0.25)' : 'rgba(51, 49, 143, 0.15)';
    const darkColor = isPrerelease ? 'dark:bg-[#02C577]' : 'dark:bg-[#33318F]';

    return (
        <RiskBadge type='labeled' label={label} outlined={false} color={color} title={label} className={darkColor} />
    );
};

// Sometimes you just need a dashed line connecting two underlined text elements together
// Sometimes you need that dashed line to be slightly offset
const DashedFlexConnector: React.FC<{ offset?: boolean }> = ({ offset }) => {
    return <div className={cn("flex-1 min-w-0 decoration-dashed decoration-neutral-light-5 underline whitespace-nowrap text-clip overflow-hidden pointer-events-none select-none aria-hidden", { 'underline-offset-4': offset })}>{"\u00A0".repeat(500)}</div>;
}

export default CollectorCard;
