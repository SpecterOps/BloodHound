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

import { faUsers } from '@fortawesome/free-solid-svg-icons';
import { FontAwesomeIcon } from '@fortawesome/react-fontawesome';
import { Box, Paper, Table, TableBody, TableCell, TableContainer, TableHead, TableRow } from '@mui/material';
import makeStyles from '@mui/styles/makeStyles';
import { AzureDataQualityStat } from 'js-client-library';
import React, { useEffect, useState } from 'react';
import { NodeIcon } from '../../components';
import { AzureNodeKind } from '../../graphSchema';
import { useAzureDataQualityStatsQuery, useAzurePlatformsDataQualityStatsQuery } from '../../hooks';
import LoadContainer from './LoadContainer';

const useStyles = makeStyles({
    print: {
        '@media print': {
            display: 'none',
        },
    },
});

export const TenantMap = {
    users: { displayText: 'Users', kind: AzureNodeKind.User },
    groups: { displayText: 'Groups', kind: AzureNodeKind.Group },
    apps: { displayText: 'Apps', kind: AzureNodeKind.App },
    service_principals: {
        displayText: 'Service Principals',
        kind: AzureNodeKind.ServicePrincipal,
    },
    devices: { displayText: 'Devices', kind: AzureNodeKind.Device },
    management_groups: {
        displayText: 'Management Groups',
        kind: AzureNodeKind.ManagementGroup,
    },
    subscriptions: { displayText: 'Subscriptions', kind: AzureNodeKind.Subscription },
    resource_groups: { displayText: 'Resource Groups', kind: AzureNodeKind.ResourceGroup },
    vms: { displayText: 'VMs', kind: AzureNodeKind.VM },
    key_vaults: { displayText: 'Key Vaults', kind: AzureNodeKind.KeyVault },
    automation_accounts: {
        displayText: 'Automation Accounts',
        kind: AzureNodeKind.AutomationAccount,
    },
    container_registries: {
        displayText: 'Container Registries',
        kind: AzureNodeKind.ContainerRegistry,
    },
    function_apps: { displayText: 'Function Apps', kind: AzureNodeKind.FunctionApp },
    logic_apps: { displayText: 'Logic Apps', kind: AzureNodeKind.LogicApp },
    managed_clusters: { displayText: 'Managed Clusters', kind: AzureNodeKind.ManagedCluster },
    vm_scale_sets: { displayText: 'VM Scale Sets', kind: AzureNodeKind.VMScaleSet },
    web_apps: { displayText: 'Web Apps', kind: AzureNodeKind.WebApp },
    tenants: { displayText: 'Tenants', kind: AzureNodeKind.Tenant },
};

export const TenantInfo: React.FC<{ contextId: string; headers?: boolean; onDataError?: () => void }> = ({
    contextId,
    headers = false,
    onDataError = () => {},
}) => {
    const { data, isLoading, isError } = useAzureDataQualityStatsQuery(contextId);
    const [tenantData, setTenantData] = useState(data || null);

    useEffect(() => {
        if (data && data.data) setTenantData(data);
    }, [data, contextId]);

    useEffect(() => {
        if (isError) onDataError();
    }, [isError, onDataError]);

    if (isLoading) {
        return <Layout stats={null} headers={headers} loading={true} />;
    }

    if (isError || !tenantData) {
        return <Layout stats={null} headers={headers} loading={false} />;
    }

    const stats = tenantData.data[0];

    return <Layout stats={stats} headers={headers} loading={false} />;
};

export const AzurePlatformInfo: React.FC<{ onDataError?: () => void }> = ({ onDataError = () => {} }) => {
    const { data, isLoading, isError } = useAzurePlatformsDataQualityStatsQuery();
    const [platformData, setPlatformData] = useState(data || null);

    useEffect(() => {
        if (data && data.data) setPlatformData(data);
    }, [data]);

    useEffect(() => {
        if (isError) onDataError();
    }, [isError, onDataError]);

    if (isLoading || !platformData) {
        return <Layout stats={null} loading={true} />;
    }

    if (isError) {
        return <Layout stats={null} loading={false} />;
    }

    const stats = platformData.data[0];

    return <Layout stats={stats} loading={false} />;
};

const Layout: React.FC<{
    stats: AzureDataQualityStat | null;
    loading: boolean;
    headers?: boolean;
}> = ({ stats, loading, headers }) => {
    const classes = useStyles();
    return (
        <Box position='relative'>
            <TableContainer component={Paper}>
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
                        {Object.keys(TenantMap).map((key) => {
                            if (key === 'tenants' && stats?.tenants === undefined) return null;

                            const mapValue = TenantMap[key as keyof typeof TenantMap];
                            const value = stats?.[key as keyof AzureDataQualityStat] as number;

                            return (
                                <LoadContainer
                                    key={key}
                                    icon={<NodeIcon nodeType={mapValue.kind} />}
                                    display={mapValue.displayText}
                                    value={value}
                                    loading={loading}
                                />
                            );
                        })}
                    </TableBody>
                </Table>
            </TableContainer>
            <TableContainer style={{ marginTop: '16px' }} component={Paper}>
                <Table>
                    <TableBody>
                        <LoadContainer
                            icon={<FontAwesomeIcon icon={faUsers} />}
                            display='Relationships'
                            value={stats?.relationships}
                            loading={loading}
                        />
                    </TableBody>
                </Table>
            </TableContainer>
        </Box>
    );
};

export default TenantInfo;
