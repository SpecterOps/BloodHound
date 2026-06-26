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

import { Alert, Box, Paper, Table, TableBody, TableCell, TableContainer, TableHead, TableRow } from '@mui/material';
import { useDataQualityNodeKindStatsQuery } from 'bh-shared-ui';
import { Typography } from 'doodle-ui';
import { DataQualityNodeKindStat } from 'js-client-library';
import { DateTime } from 'luxon';
import { useMemo } from 'react';
import { DataQualitySelection } from './DataQualityEnvironmentSelector';

type HistoryPoint = {
    createdAt: string;
    value: number;
};

type NodeKindCountRow = {
    key: string;
    kindName: string;
    sourceKind: string;
    extensionName: string;
    latestValue: number;
    history: HistoryPoint[];
};

const numberFormatter = new Intl.NumberFormat();

const isBuiltInDataQualityType = (type: string) => type === 'active-directory' || type === 'azure';

const OpenGraphNodeKindChart: React.FC<{ rows: NodeKindCountRow[] }> = ({ rows }) => {
    const width = 900;
    const height = 360;
    const padding = 48;
    const colors = ['#4f46e5', '#0f766e', '#b45309', '#be123c', '#7c3aed', '#0369a1', '#15803d', '#c2410c'];
    const points = rows.flatMap((row) => row.history);
    const timestamps = points.map((point) => DateTime.fromISO(point.createdAt).toMillis());
    const values = points.map((point) => point.value);
    const minTime = Math.min(...timestamps);
    const maxTime = Math.max(...timestamps);
    const maxValue = Math.max(...values, 1);
    const timeRange = maxTime - minTime || 1;

    const getPoint = (point: HistoryPoint) => {
        const timestamp = DateTime.fromISO(point.createdAt).toMillis();
        const x = padding + ((timestamp - minTime) / timeRange) * (width - padding * 2);
        const y = height - padding - (point.value / maxValue) * (height - padding * 2);
        return `${x},${y}`;
    };

    return (
        <Paper variant='outlined'>
            <Box p={2}>
                <Typography variant='h6'>OpenGraph Nodes Over Time</Typography>
                <svg width='100%' viewBox={`0 0 ${width} ${height}`} role='img' aria-label='OpenGraph nodes over time'>
                    <line x1={padding} x2={padding} y1={padding} y2={height - padding} stroke='currentColor' />
                    <line
                        x1={padding}
                        x2={width - padding}
                        y1={height - padding}
                        y2={height - padding}
                        stroke='currentColor'
                    />
                    {rows.map((row, index) => {
                        const stroke = colors[index % colors.length];
                        return (
                            <polyline
                                key={row.key}
                                fill='none'
                                stroke={stroke}
                                strokeWidth='2'
                                points={row.history.map(getPoint).join(' ')}
                            />
                        );
                    })}
                </svg>
                <Box display='flex' flexWrap='wrap' gap={1}>
                    {rows.map((row, index) => (
                        <Box key={row.key} display='flex' alignItems='center' gap={0.5}>
                            <Box
                                width={10}
                                height={10}
                                borderRadius='50%'
                                bgcolor={colors[index % colors.length]}
                            />
                            <Typography variant='caption'>
                                {row.kindName} ({row.extensionName} / {row.sourceKind})
                            </Typography>
                        </Box>
                    ))}
                </Box>
            </Box>
        </Paper>
    );
};

const Sparkline: React.FC<{ points: HistoryPoint[] }> = ({ points }) => {
    const width = 140;
    const height = 36;
    const padding = 4;
    const values = points.map((point) => point.value);
    const min = Math.min(...values);
    const max = Math.max(...values);
    const range = max - min;

    const pathPoints = points
        .map((point, index) => {
            const x = points.length === 1 ? width / 2 : padding + (index / (points.length - 1)) * (width - padding * 2);
            const normalized = range === 0 ? 0.5 : (point.value - min) / range;
            const y = height - padding - normalized * (height - padding * 2);
            return `${x},${y}`;
        })
        .join(' ');

    return (
        <svg width={width} height={height} role='img' aria-label='Node count trend'>
            {points.length === 1 ? (
                <circle cx={width / 2} cy={height / 2} r='3' fill='currentColor' />
            ) : (
                <polyline fill='none' stroke='currentColor' strokeWidth='2' points={pathPoints} />
            )}
        </svg>
    );
};

const aggregateStats = (stats: DataQualityNodeKindStat[], aggregate: boolean): NodeKindCountRow[] => {
    const groupedRows = new Map<string, Map<string, HistoryPoint>>();

    for (const stat of stats) {
        const rowKey = `${stat.schema_extension_id}:${stat.environment_kind}:${stat.source_kind}:${stat.kind_name}`;
        const pointKey = aggregate ? stat.run_id : `${stat.run_id}:${stat.environment_id}`;

        if (!groupedRows.has(rowKey)) {
            groupedRows.set(rowKey, new Map<string, HistoryPoint>());
        }

        const rowPoints = groupedRows.get(rowKey);
        if (!rowPoints) continue;

        const existingPoint = rowPoints.get(pointKey);
        rowPoints.set(pointKey, {
            createdAt: stat.created_at,
            value: (existingPoint?.value ?? 0) + stat.metric_value,
        });
    }

    return [...groupedRows.entries()]
        .map(([key, pointsByRun]) => {
            const [schemaExtensionID, , sourceKind, kindName] = key.split(':');
            const firstStat = stats.find(
                (stat) =>
                    String(stat.schema_extension_id) === schemaExtensionID &&
                    stat.source_kind === sourceKind &&
                    stat.kind_name === kindName
            );
            const history = [...pointsByRun.values()].sort(
                (first, second) => new Date(first.createdAt).getTime() - new Date(second.createdAt).getTime()
            );
            const latestPoint = history[history.length - 1];

            return {
                key,
                kindName,
                sourceKind,
                extensionName: firstStat?.schema_extension_display_name || 'OpenGraph',
                latestValue: latestPoint?.value ?? 0,
                history,
            };
        })
        .sort((first, second) => first.kindName.localeCompare(second.kindName));
};

const OpenGraphNodeKindCounts: React.FC<{ selectedEnvironment: DataQualitySelection | null }> = ({
    selectedEnvironment,
}) => {
    const includeBuiltin = selectedEnvironment ? !isBuiltInDataQualityType(selectedEnvironment.type) : true;
    const sourceKind =
        selectedEnvironment && isBuiltInDataQualityType(selectedEnvironment.type)
            ? selectedEnvironment.sourceKind
            : undefined;

    const { data, isLoading, isError } = useDataQualityNodeKindStatsQuery({
        environmentKind: selectedEnvironment?.environmentKind,
        environmentId: selectedEnvironment?.selectionType === 'environment' ? selectedEnvironment.id : null,
        sourceKind,
        includeBuiltin,
    });

    const rows = useMemo(
        () => aggregateStats(data?.data ?? [], selectedEnvironment?.selectionType === 'aggregate'),
        [data?.data, selectedEnvironment?.selectionType]
    );

    if (!selectedEnvironment || isLoading) return null;

    if (isError) {
        return <Alert severity='warning'>OpenGraph node count history is unavailable.</Alert>;
    }

    if (rows.length === 0) {
        return (
            <Alert severity='info'>
                No OpenGraph node kind count history is available for this environment selection.
            </Alert>
        );
    }

    return (
        <Box mt={2}>
            <Box mb={1}>
                <Typography variant='h5'>OpenGraph Node Counts</Typography>
            </Box>
            <OpenGraphNodeKindChart rows={rows} />
            <TableContainer component={Paper} variant='outlined' sx={{ mt: 2 }}>
                <Table size='small' aria-label='OpenGraph node counts'>
                    <TableHead>
                        <TableRow>
                            <TableCell>Node Kind</TableCell>
                            <TableCell>Extension</TableCell>
                            <TableCell>Source</TableCell>
                            <TableCell align='right'>Latest Count</TableCell>
                            <TableCell>30 Day Trend</TableCell>
                        </TableRow>
                    </TableHead>
                    <TableBody>
                        {rows.map((row) => (
                            <TableRow key={row.key}>
                                <TableCell>{row.kindName}</TableCell>
                                <TableCell>{row.extensionName}</TableCell>
                                <TableCell>{row.sourceKind}</TableCell>
                                <TableCell align='right'>{numberFormatter.format(row.latestValue)}</TableCell>
                                <TableCell>
                                    <Sparkline points={row.history} />
                                </TableCell>
                            </TableRow>
                        ))}
                    </TableBody>
                </Table>
            </TableContainer>
        </Box>
    );
};

export default OpenGraphNodeKindCounts;
