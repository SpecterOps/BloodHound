// Copyright 2026 Specter Ops, Inc.
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

import { faStream } from '@fortawesome/free-solid-svg-icons';
import { FontAwesomeIcon } from '@fortawesome/react-fontawesome';
import { Box, Paper, Table, TableBody, TableCell, TableContainer, TableHead, TableRow } from '@mui/material';
import makeStyles from '@mui/styles/makeStyles';
import {
    DataQualityAggregation,
    DataQualityMetricType,
    DataQualityStat,
} from 'js-client-library';
import React, { useEffect } from 'react';
import { NodeIcon } from '../../components';
import { useOpenGraphDataQualityAggregationsQuery, useOpenGraphDataQualityStatsQuery } from '../../hooks';
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

type OpenGraphMetric = DataQualityStat | DataQualityAggregation;
type OpenGraphMetricRow = Pick<OpenGraphMetric, 'metric_type' | 'metric_name' | 'metric_value'>;

const relationshipMetricName = 'relationships';

const loadingRows: OpenGraphMetricRow[] = [
    { metric_type: 'node', metric_name: 'Nodes', metric_value: 0 },
    { metric_type: 'relationship', metric_name: relationshipMetricName, metric_value: 0 },
];

export const OpenGraphEnvironmentInfo: React.FC<{
    contextId: string;
    extensionId?: number | null;
    schemaEnvironmentId?: number | null;
    headers?: boolean;
    onDataError?: () => void;
}> = ({ contextId, extensionId, schemaEnvironmentId, headers = false, onDataError = () => {} }) => {
    const { data: environmentData, isLoading, isError } = useOpenGraphDataQualityStatsQuery(
        contextId,
        extensionId,
        schemaEnvironmentId
    );

    useEffect(() => {
        if (isError) onDataError();
    }, [isError, onDataError]);

    if (isLoading) {
        return <Layout stats={[]} headers={headers} loading={true} />;
    }

    if (isError || !environmentData || !environmentData.data.length) {
        return null;
    }

    return <Layout stats={getLatestRunMetrics(environmentData.data)} headers={headers} loading={false} />;
};

export const OpenGraphExtensionInfo: React.FC<{
    extensionId?: number | null;
    schemaEnvironmentId?: number | null;
    headers?: boolean;
    onDataError?: () => void;
}> = ({ extensionId, schemaEnvironmentId, headers = false, onDataError = () => {} }) => {
    const { data: extensionData, isLoading, isError } = useOpenGraphDataQualityAggregationsQuery(
        extensionId,
        schemaEnvironmentId
    );

    useEffect(() => {
        if (isError) onDataError();
    }, [isError, onDataError]);

    if (extensionId === undefined || extensionId === null) {
        return null;
    }

    if (isLoading) {
        return <Layout stats={[]} headers={headers} loading={true} />;
    }

    if (isError || !extensionData || !extensionData.data.length) {
        return null;
    }

    return <Layout stats={getLatestRunMetrics(extensionData.data)} headers={headers} loading={false} />;
};

function getLatestRunMetrics(stats: OpenGraphMetric[]): OpenGraphMetricRow[] {
    const latestRunID = stats[0]?.run_id;
    if (!latestRunID) {
        return [];
    }

    // The summary tables show the newest collection run; history stays in the charts.
    return stats.filter((stat) => stat.run_id === latestRunID);
}

function getMetricDisplayName(row: OpenGraphMetricRow): string {
    if (row.metric_type === 'relationship' && row.metric_name === relationshipMetricName) {
        return 'Relationships';
    }

    return row.metric_name;
}

const Layout: React.FC<{
    stats: OpenGraphMetricRow[];
    loading: boolean;
    headers?: boolean;
}> = ({ stats, loading, headers }) => {
    const rows = loading ? loadingRows : stats;
    const nodeRows = rows.filter((stat) => stat.metric_type === 'node');
    const relationshipRows = rows.filter((stat) => stat.metric_type === 'relationship');
    const metricSections = [nodeRows, relationshipRows].filter((sectionRows) => sectionRows.length > 0);

    if (!loading && metricSections.length === 0) {
        return null;
    }

    return (
        <Box position='relative'>
            {metricSections.map((sectionRows, index) => (
                <MetricTable
                    key={sectionRows[0].metric_type}
                    headers={headers}
                    marginTop={index > 0}
                    rows={sectionRows}
                    loading={loading}
                />
            ))}
        </Box>
    );
};

const MetricTable: React.FC<{
    headers?: boolean;
    marginTop: boolean;
    rows: OpenGraphMetricRow[];
    loading: boolean;
}> = ({ headers, marginTop, rows, loading }) => {
    const classes = useStyles();

    return (
        <TableContainer component={Paper} className={classes.container} style={marginTop ? { marginTop: '16px' } : {}}>
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
                    {rows.map((row) => (
                        <LoadContainer
                            key={`${row.metric_type}-${row.metric_name}`}
                            icon={<MetricIcon metricName={row.metric_name} metricType={row.metric_type} />}
                            display={getMetricDisplayName(row)}
                            value={row.metric_value}
                            loading={loading}
                        />
                    ))}
                </TableBody>
            </Table>
        </TableContainer>
    );
};

const MetricIcon: React.FC<{ metricName: string; metricType: DataQualityMetricType }> = ({
    metricName,
    metricType,
}) => {
    if (metricType === 'node') {
        return <NodeIcon nodeType={metricName} />;
    }

    return <FontAwesomeIcon icon={faStream} />;
};
