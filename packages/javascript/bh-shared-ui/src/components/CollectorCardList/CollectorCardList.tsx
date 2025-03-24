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

import { Accordion, AccordionContent, AccordionHeader, AccordionItem } from '@bloodhoundenterprise/doodleui';
import { Box, Typography, useTheme } from '@mui/material';
import { useState } from 'react';
import { CaretDown, CaretUp } from '../AppIcon/Icons';
import CollectorCard, { COLLECTOR_TYPE_LABEL, LabelType } from '../CollectorCard';
import { CollectorType } from 'js-client-library/src/types';

const COLLECTOR_SHORT_LABEL: { [key in CollectorType]: string } = {
    sharphound_enterprise: 'SHE',
    sharphound: 'SH',
    azurehound: 'AH',
};

interface CollectorDownloadFile {
    os: string;
    arch: string;
    onClickDownload: () => void;
    onClickDownloadChecksum: () => void;
}

interface CollectorCardProps {
    collectorType: CollectorType;
    version: string;
    downloadArtifacts: CollectorDownloadFile[];
    timestamp?: number;
    label?: LabelType;
    isPrerelease?: boolean;
}

interface CollectorCardListProps {
    collectors: CollectorCardProps[];
    noLabels?: boolean;
}

const CollectorCardList: React.FC<CollectorCardListProps> = ({ collectors, noLabels = false }) => {
    const theme = useTheme();

    // Few enough collectors that this isn't worth memoizing
    // And the only stateful changes that should cause rerender
    // are collectors and theme
    const sortedCollectors = collectors.toSorted((a, b) =>
        b.version.toLowerCase().localeCompare(a.version.toLowerCase())
    );
    const latestStable = sortedCollectors.filter((c) => !c.isPrerelease)[0];
    const latestPrerelease = sortedCollectors.filter(
        // Is prerelease and version is greater than latest stable
        (c) => c.isPrerelease && c.version.toLowerCase().localeCompare(latestStable.version.toLowerCase()) > 0
    )[0];
    const olderVersions = sortedCollectors.filter(
        (c) => !c.isPrerelease && c !== latestStable && c !== latestPrerelease
    );

    return (
        <Box display='flex' flexDirection='column' gap={theme.spacing(2)}>
            <Typography variant='h6'>{COLLECTOR_TYPE_LABEL[latestStable.collectorType]}</Typography>
            <Box display='flex' flexDirection='row' gap={theme.spacing(2)}>
                <CollectorCard
                    className='flex-1 min-w-0 bg-neutral-light-3 dark:bg-neutral-dark-5'
                    collectorType={latestStable.collectorType}
                    version={latestStable.version}
                    timestamp={latestStable.timestamp}
                    label={noLabels ? undefined : 'latest'}
                    isPrerelease={latestStable.isPrerelease}
                    downloadArtifacts={latestStable.downloadArtifacts.map(artifact => {return {fileName: `${COLLECTOR_SHORT_LABEL[latestStable.collectorType]} ${latestStable.version} ${artifact.os} ${artifact.arch}`, ...artifact}})}
                />
                {latestPrerelease && (
                    <CollectorCard
                        className='flex-1 min-w-0'
                        collectorType={latestPrerelease.collectorType}
                        version={latestPrerelease.version}
                        timestamp={latestPrerelease.timestamp}
                        label={noLabels ? undefined : latestPrerelease.label}
                        isPrerelease={latestPrerelease.isPrerelease}
                        downloadArtifacts={latestPrerelease.downloadArtifacts.map(artifact => {return {fileName: `${COLLECTOR_SHORT_LABEL[latestPrerelease.collectorType]} ${latestPrerelease.version} ${artifact.os} ${artifact.arch}`, ...artifact}})}
                    />
                )}
            </Box>
            {olderVersions.length > 0 && <OlderVersionsList collectors={olderVersions} noLabels />}
        </Box>
    );
};

const OlderVersionsList: React.FC<CollectorCardListProps> = ({ collectors, noLabels = false }) => {
    const theme = useTheme();
    const [expanded, setExpanded] = useState(false);

    return (
        <Accordion type='single' onValueChange={() => setExpanded((x) => !x)} collapsible>
            <AccordionItem value='older-versions' className='bg-neutral-light-0 dark:bg-neutral-dark-1'>
                <AccordionHeader>
                    <Box display='flex' flexDirection='row' gap={1}>
                        {expanded ? <CaretUp /> : <CaretDown />}
                        Older Versions
                    </Box>
                </AccordionHeader>
                <AccordionContent className="p-0">
                    <Box display='flex' flexDirection='column' gap={theme.spacing(2)}>
                        {collectors.map((collector) => (
                            <CollectorCard
                                key={collector.version}
                                className='flex-1 min-w-0'
                                collectorType={collector.collectorType}
                                version={collector.version}
                                timestamp={collector.timestamp}
                                label={noLabels ? undefined : collector.label}
                                isPrerelease={collector.isPrerelease}
                                downloadArtifacts={collector.downloadArtifacts.map(artifact => {return {fileName: `${COLLECTOR_SHORT_LABEL[collector.collectorType]} ${collector.version} ${artifact.os} ${artifact.arch}`, ...artifact}})}
                                needsUnderlineOffset
                            />
                        ))}
                    </Box>
                </AccordionContent>
            </AccordionItem>
        </Accordion>
    );
};

export default CollectorCardList;
