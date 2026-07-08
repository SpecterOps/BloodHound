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
import { Box, Paper, Table, TableBody, TableCell, TableContainer, TableHead, TableRow } from '@mui/material';
import makeStyles from '@mui/styles/makeStyles';
import React, { useEffect } from 'react';
import { NodeIcon } from '../../components';
import { useOpenGraphDataQualityStatsQuery, useOpenGraphPlatformsDataQualityStatsQuery } from '../../hooks';
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

// getLatestMetricStats keeps only the latest stat per metric_kind_id (by created_at) and
// splits the result into node stats and the latest relationship stat.
const getLatestMetricStats = (data: any[]) => {
    const latestStatsByMetricKind = new Map<number, any>();
    for (const stat of data) {
        const existing = latestStatsByMetricKind.get(stat.metric_kind_id);
        const isLatest = new Date(stat?.created_at).getTime() > new Date(existing?.created_at).getTime();
        if (!existing || isLatest) {
            latestStatsByMetricKind.set(stat.metric_kind_id, stat);
        }
    }
    const stats = Array.from(latestStatsByMetricKind.values());
    const nodeStats = stats.filter((stat) => stat.metric_type === 'node');
    const relationshipStats = stats.find((stat) => stat.metric_type === 'relationship');

    return { nodeStats, relationshipStats };
};

export const OpenGraphInfo: React.FC<{ contextId: string; headers?: boolean; onDataError?: () => void }> = ({
    contextId,
    headers = false,
    onDataError = () => {},
}) => {
    const { data: tenantData, isLoading, isError } = useOpenGraphDataQualityStatsQuery(contextId);

    console.log('TD', tenantData);

    useEffect(() => {
        if (isError) onDataError();
    }, [isError, onDataError]);

    if (isLoading) {
        return <Layout nodeStats={null} relationshipStats={null} headers={headers} loading={true} />;
    }

    if (isError || !tenantData || !tenantData.data.length) {
        return null;
    }

    const { nodeStats, relationshipStats } = getLatestMetricStats(tenantData.data);

    return <Layout nodeStats={nodeStats} relationshipStats={relationshipStats} headers={headers} loading={false} />;
};

export const OpenGraphPlatformInfo: React.FC<{ contextType: string; onDataError?: () => void }> = ({
    contextType,
    onDataError = () => {},
}) => {
    const { data: platformData, isLoading, isError } = useOpenGraphPlatformsDataQualityStatsQuery(contextType);

    useEffect(() => {
        if (isError) onDataError();
    }, [isError, onDataError]);

    if (isLoading) {
        return <Layout nodeStats={null} relationshipStats={null} loading={true} />;
    }

    if (isError || !platformData || !platformData.data.length) {
        return null;
    }

    const { nodeStats, relationshipStats } = getLatestMetricStats(platformData.data);

    return <Layout nodeStats={nodeStats} relationshipStats={relationshipStats} loading={false} />;
};

const MetricIcon: React.FC<{ metricName: string; metricType: any }> = ({ metricName, metricType }) => {
    if (metricType === 'node') {
        return <NodeIcon nodeType={metricName} />;
    }

    return <FontAwesomeIcon icon={faStream} />;
};

const Layout: React.FC<{
    nodeStats: any;
    relationshipStats: any;
    loading: boolean;
    headers?: boolean;
}> = ({ nodeStats, relationshipStats, loading, headers }) => {
    const classes = useStyles();
    return (
        <Box position='relative'>
            <TableContainer className={classes.container}>
                <Table>
                    {headers && (
                        <TableHead className={classes.print}>
                            <TableRow>
                                <TableCell align={'left'}>Item</TableCell>
                                <TableCell align={'right'}>Result</TableCell>
                            </TableRow>
                        </TableHead>
                    )}
                    <TableBody>
                        {nodeStats?.map((key: any) => {
                            if (key.metric_kind_id === key.environment_kind_id) return null;

                            return (
                                <LoadContainer
                                    key={key.id}
                                    icon={<MetricIcon metricName={key.metric_name} metricType={key.metric_type} />}
                                    display={key.metric_name}
                                    value={key.metric_value}
                                    loading={loading}
                                />
                            );
                        })}
                    </TableBody>
                </Table>
            </TableContainer>
            <TableContainer style={{ marginTop: '16px' }} component={Paper} className={classes.container}>
                <Table>
                    <TableBody>
                        <LoadContainer
                            icon={<FontAwesomeIcon icon={faUsers} />}
                            display='Relationships'
                            value={relationshipStats?.metric_value ?? 0}
                            loading={loading}
                        />
                    </TableBody>
                </Table>
            </TableContainer>
        </Box>
    );
};

export default OpenGraphInfo;
