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
        return <Layout dbInfo={null} isPlatform={false} headers={headers} loading={true} />;
    }

    if (isError || !tenantData) {
        return <Layout dbInfo={null} isPlatform={false} headers={headers} loading={false} />;
    }

    const dbInfo = tenantData.data[0] as AzureDataQualityStat;

    return <Layout dbInfo={dbInfo} headers={headers} loading={false} />;
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
        return <Layout dbInfo={null} isPlatform={true} loading={true} />;
    }

    if (isError) {
        return <Layout dbInfo={null} isPlatform={true} loading={false} />;
    }

    const dbInfo = platformData.data[0] as AzureDataQualityStat;

    return <Layout dbInfo={dbInfo} isPlatform={true} loading={false} />;
};

const Layout: React.FC<{
    dbInfo: AzureDataQualityStat | null;
    loading: boolean;
    isPlatform?: boolean;
    headers?: boolean;
}> = ({ dbInfo, loading, isPlatform, headers }) => {
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
                            icon={<NodeIcon nodeType={'AZUser'} />}
                            display='Users'
                            value={dbInfo?.users || 0}
                            loading={loading}
                        />
                        <LoadContainer
                            icon={<NodeIcon nodeType={'AZGroup'} />}
                            display='Groups'
                            value={dbInfo?.groups || 0}
                            loading={loading}
                        />
                        <LoadContainer
                            icon={<NodeIcon nodeType={'AZApp'} />}
                            display='Apps'
                            value={dbInfo?.apps || 0}
                            loading={loading}
                        />
                        <LoadContainer
                            icon={<NodeIcon nodeType={'AZServicePrincipal'} />}
                            display='Service Principals'
                            value={dbInfo?.service_principals || 0}
                            loading={loading}
                        />
                        <LoadContainer
                            icon={<NodeIcon nodeType={'AZDevice'} />}
                            display='Devices'
                            value={dbInfo?.devices || 0}
                            loading={loading}
                        />
                        <LoadContainer
                            icon={<NodeIcon nodeType={'AZManagementGroup'} />}
                            display='Management Groups'
                            value={dbInfo?.management_groups || 0}
                            loading={loading}
                        />
                        <LoadContainer
                            icon={<NodeIcon nodeType={'AZSubscription'} />}
                            display='Subscriptions'
                            value={dbInfo?.subscriptions || 0}
                            loading={loading}
                        />
                        <LoadContainer
                            icon={<NodeIcon nodeType={'AZResourceGroup'} />}
                            display='Resource Groups'
                            value={dbInfo?.resource_groups || 0}
                            loading={loading}
                        />
                        <LoadContainer
                            icon={<NodeIcon nodeType={'AZVM'} />}
                            display='VMs'
                            value={dbInfo?.vms || 0}
                            loading={loading}
                        />
                        <LoadContainer
                            icon={<NodeIcon nodeType={'AZKeyVault'} />}
                            display='Key Vaults'
                            value={dbInfo?.key_vaults || 0}
                            loading={loading}
                        />
                        {isPlatform && (
                            <LoadContainer
                                icon={<NodeIcon nodeType={'AZTenant'} />}
                                display='Tenants'
                                value={dbInfo?.tenants || 0}
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
                            value={dbInfo?.relationships || 0}
                            loading={loading}
                        />
                    </TableBody>
                </Table>
            </TableContainer>
        </Box>
    );
};

export default TenantInfo;
