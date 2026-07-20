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

import { faStream, faUsers } from '@fortawesome/free-solid-svg-icons';
import { FontAwesomeIcon } from '@fortawesome/react-fontawesome';
import { Box, Paper, Table, TableBody, TableContainer } from '@mui/material';
import makeStyles from '@mui/styles/makeStyles';
import { Environment, OpenGraphDataQualityStat } from 'js-client-library';
import React, { useEffect } from 'react';
import { NodeIcon } from '../../components';
import { useOpenGraphDataQualityStatsQuery, useOpenGraphPlatformsDataQualityStatsQuery } from '../../hooks';
import { cn } from '../../utils';
import LoadContainer from './LoadContainer';

const useStyles = makeStyles((theme) => ({
    print: {
        '@media print': {
            display: 'none',
        },
    },
    container: {
        backgroundColor: theme.palette.neutral.secondary,
    },
}));

// getLatestMetricStats keeps only the latest stat per kind_id (by created_at) and
// splits the result into node stats and the latest relationship stat.
export const getLatestMetricStats = (
    data: OpenGraphDataQualityStat[]
): { nodeStats: OpenGraphDataQualityStat[]; relationshipStat: OpenGraphDataQualityStat } => {
    const latestStatsByMetricKind = new Map<number, any>();
    for (const stat of data) {
        const existing = latestStatsByMetricKind.get(stat.kind_id);
        const isLatest = new Date(stat?.created_at).getTime() > new Date(existing?.created_at).getTime();
        if (!existing || isLatest) {
            latestStatsByMetricKind.set(stat.kind_id, stat);
        }
    }

    const stats = Array.from(latestStatsByMetricKind.values());
    const nodeStats = stats.filter((stat) => stat.metric_type === 'node');
    const relationshipStat = stats.find((stat) => stat.metric_type === 'relationship');

    return { nodeStats, relationshipStat };
};

export const OpenGraphInfo: React.FC<{ contextId: string; onDataError?: () => void }> = ({
    contextId,
    onDataError = () => {},
}) => {
    const { data: openGraphData, isLoading, isError } = useOpenGraphDataQualityStatsQuery(contextId);

    useEffect(() => {
        if (isError) onDataError();
    }, [isError, onDataError]);

    if (isLoading) {
        return <Layout nodeStats={null} relationshipStat={null} isLoading={true} />;
    }

    if (isError || !openGraphData || !openGraphData.data.length) {
        return null;
    }

    const { nodeStats, relationshipStat } = getLatestMetricStats(openGraphData.data);

    return <Layout nodeStats={nodeStats} relationshipStat={relationshipStat} isLoading={false} />;
};

export const OpenGraphPlatformInfo: React.FC<{
    contextKindId: Environment['environment_kind_id'];
    onDataError?: () => void;
}> = ({ contextKindId, onDataError = () => {} }) => {
    const { data: platformData, isLoading, isError } = useOpenGraphPlatformsDataQualityStatsQuery(contextKindId);

    useEffect(() => {
        if (isError) onDataError();
    }, [isError, onDataError]);

    if (isLoading) {
        return <Layout nodeStats={null} relationshipStat={null} isLoading={true} />;
    }

    if (isError || !platformData || !platformData.data.length) {
        return null;
    }

    const { nodeStats, relationshipStat } = getLatestMetricStats(platformData.data);

    return <Layout nodeStats={nodeStats} relationshipStat={relationshipStat} isLoading={false} />;
};

const MetricIcon: React.FC<{ metricName: string; metricType: any }> = ({ metricName, metricType }) => {
    if (metricType === 'node') {
        return <NodeIcon nodeType={metricName} />;
    }

    return <FontAwesomeIcon icon={faStream} />;
};

const Layout: React.FC<{
    nodeStats: OpenGraphDataQualityStat[] | null;
    relationshipStat: OpenGraphDataQualityStat | null;
    isLoading: boolean;
    headers?: boolean;
}> = ({ nodeStats, relationshipStat, isLoading }) => {
    const classes = useStyles();

    return (
        <Box position='relative'>
            <TableContainer className={classes.container}>
                <Table>
                    <TableBody>
                        {nodeStats?.map((key: any) => {
                            return (
                                <LoadContainer
                                    key={key.id}
                                    icon={<MetricIcon metricName={key.metric_name} metricType={key.metric_type} />}
                                    display={key.metric_name}
                                    value={key.metric_value}
                                    isLoading={isLoading}
                                />
                            );
                        })}
                    </TableBody>
                </Table>
            </TableContainer>
            <TableContainer component={Paper} className={cn(classes.container, { 'mt-4': !isLoading })}>
                <Table>
                    <TableBody>
                        <LoadContainer
                            icon={<FontAwesomeIcon icon={faUsers} />}
                            display='Relationships'
                            value={relationshipStat?.metric_value ?? 0}
                            isLoading={isLoading}
                        />
                    </TableBody>
                </Table>
            </TableContainer>
        </Box>
    );
};

export default OpenGraphInfo;
