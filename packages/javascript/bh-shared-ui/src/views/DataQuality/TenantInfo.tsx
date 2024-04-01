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

import React, { useEffect, useState } from 'react';
import { Box, Table, TableBody, TableContainer, Paper, TableHead, TableCell, TableRow } from '@mui/material';
import makeStyles from '@mui/styles/makeStyles';
import {
    useAzureDataQualityStatsQuery,
    useAzurePlatformsDataQualityStatsQuery,
    AzureDataQualityStat,
} from '../../hooks';
import LoadContainer from './LoadContainer';
import { faUsers } from '@fortawesome/free-solid-svg-icons';
import { FontAwesomeIcon } from '@fortawesome/react-fontawesome';
import { NodeIcon } from '../../components';
import { AzureNodeKind } from '../../graphSchema';

const useStyles = makeStyles({
    print: {
        '@media print': {
            display: 'none',
        },
    },
});

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

    const stats = tenantData.data[0] as AzureDataQualityStat;

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

    const stats = platformData.data[0] as AzureDataQualityStat;

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
                        <LoadContainer
                            icon={<NodeIcon nodeType={AzureNodeKind.User} />}
                            display='Users'
                            value={stats?.users}
                            loading={loading}
                        />
                        <LoadContainer
                            icon={<NodeIcon nodeType={AzureNodeKind.Group} />}
                            display='Groups'
                            value={stats?.groups}
                            loading={loading}
                        />
                        <LoadContainer
                            icon={<NodeIcon nodeType={AzureNodeKind.App} />}
                            display='Apps'
                            value={stats?.apps}
                            loading={loading}
                        />
                        <LoadContainer
                            icon={<NodeIcon nodeType={AzureNodeKind.ServicePrincipal} />}
                            display='Service Principals'
                            value={stats?.service_principals}
                            loading={loading}
                        />
                        <LoadContainer
                            icon={<NodeIcon nodeType={AzureNodeKind.Device} />}
                            display='Devices'
                            value={stats?.devices}
                            loading={loading}
                        />
                        <LoadContainer
                            icon={<NodeIcon nodeType={AzureNodeKind.ManagementGroup} />}
                            display='Management Groups'
                            value={stats?.management_groups}
                            loading={loading}
                        />
                        <LoadContainer
                            icon={<NodeIcon nodeType={AzureNodeKind.Subscription} />}
                            display='Subscriptions'
                            value={stats?.subscriptions}
                            loading={loading}
                        />
                        <LoadContainer
                            icon={<NodeIcon nodeType={AzureNodeKind.ResourceGroup} />}
                            display='Resource Groups'
                            value={stats?.resource_groups}
                            loading={loading}
                        />
                        <LoadContainer
                            icon={<NodeIcon nodeType={AzureNodeKind.VM} />}
                            display='VMs'
                            value={stats?.vms}
                            loading={loading}
                        />
                        <LoadContainer
                            icon={<NodeIcon nodeType={AzureNodeKind.KeyVault} />}
                            display='Key Vaults'
                            value={stats?.key_vaults}
                            loading={loading}
                        />
                        <LoadContainer
                            icon={<NodeIcon nodeType={AzureNodeKind.AutomationAccount} />}
                            display='Automation Accounts'
                            value={stats?.automation_accounts}
                            loading={loading}
                        />
                        <LoadContainer
                            icon={<NodeIcon nodeType={AzureNodeKind.ContainerRegistry} />}
                            display='Container Registries'
                            value={stats?.container_registries}
                            loading={loading}
                        />
                        <LoadContainer
                            icon={<NodeIcon nodeType={AzureNodeKind.FunctionApp} />}
                            display='Function Apps'
                            value={stats?.function_apps}
                            loading={loading}
                        />
                        <LoadContainer
                            icon={<NodeIcon nodeType={AzureNodeKind.LogicApp} />}
                            display='Logic Apps'
                            value={stats?.logic_apps}
                            loading={loading}
                        />
                        <LoadContainer
                            icon={<NodeIcon nodeType={AzureNodeKind.ManagedCluster} />}
                            display='Managed Clusters'
                            value={stats?.managed_clusters}
                            loading={loading}
                        />
                        <LoadContainer
                            icon={<NodeIcon nodeType={AzureNodeKind.VMScaleSet} />}
                            display='VM Scale Sets'
                            value={stats?.vm_scale_sets}
                            loading={loading}
                        />
                        <LoadContainer
                            icon={<NodeIcon nodeType={AzureNodeKind.WebApp} />}
                            display='Web Apps'
                            value={stats?.web_apps}
                            loading={loading}
                        />
                        {stats?.tenants !== undefined && (
                            <LoadContainer
                                icon={<NodeIcon nodeType={AzureNodeKind.Tenant} />}
                                display='Tenants'
                                value={stats?.tenants}
                                loading={loading}
                            />
                        )}
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
