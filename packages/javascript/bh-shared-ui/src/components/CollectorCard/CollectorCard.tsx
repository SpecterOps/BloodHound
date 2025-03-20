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
import { cn } from '../../utils';

export type LabelType = 'latest' | 'prerelease';

export type CollectorType = 'sharphoundEnterprise' | 'sharphound' | 'azurehound';

export const COLLECTOR_TYPE_LABEL: { [key in CollectorType]: string } = {
    sharphoundEnterprise: 'Sharphound Enterprise',
    sharphound: 'Sharphound Community',
    azurehound: 'Azurehound',
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
    timestamp: number;
    downloadArtifacts: CollectorDownloadFile[];
    label?: LabelType;
    isLatest?: boolean;
    isPrerelease?: boolean;
}

const CollectorCard: React.FC<CollectorCardProps> = ({
    collectorType,
    version,
    timestamp,
    downloadArtifacts,
    label = undefined,
    isLatest = false,
    isPrerelease = false,
    ...props
}) => {
    const date = new Date(timestamp);

    return (
        <Card {...props} className={cn(props.className, { 'bg-neutral-light-3 dark:bg-neutral-dark-5': isLatest })}>
            <CardHeader>
                <Box display='flex' flexDirection='row' alignItems='center' overflow='hidden' gap='1rem'>
                    <CardTitle>{`${version}`.trim().toUpperCase()}</CardTitle>
                    {isPrerelease && <Typography variant='caption'>(pre-release)</Typography>}
                    {label && <CollectorLabel label={label} isPrerelease={isPrerelease} />}
                    <CardDescription>
                        {`${date.getFullYear()}.${date.getMonth() + 1}.${date.getDate()}`}
                    </CardDescription>
                </Box>
            </CardHeader>
            <CardContent>
                <ul>
                    {downloadArtifacts.map((collector) => (
                        <li key={collector.fileName}>
                            <Box display='flex' flexDirection='row' justifyContent='space-between' gap='2rem'>
                                <Link
                                    component='button'
                                    variant='body1'
                                    onClick={collector.onClickDownload}
                                    title={`Download ${COLLECTOR_TYPE_LABEL[collectorType]} ${version} ${collector.os} ${collector.arch}`.trim()}
                                    textOverflow='ellipsis'
                                    noWrap>
                                    {collector.fileName}
                                </Link>
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
    const color = isPrerelease ? 'rgba(243, 96, 54, 0.25)' : 'rgba(51, 49, 143, 0.15)';

    return <RiskBadge type='labeled' label={label} outlined={false} color={color} title={label} />;
};

export default CollectorCard;
