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

import React, { useState, useEffect } from 'react';
import { Box, Table, TableBody, TableContainer, Paper, TableHead, TableRow, TableCell } from '@mui/material';
import makeStyles from '@mui/styles/makeStyles';
import LoadContainer from './LoadContainer';
import { faUsers, faChartPie, faSignInAlt, faStream } from '@fortawesome/free-solid-svg-icons';
import { FontAwesomeIcon } from '@fortawesome/react-fontawesome';
import { NodeIcon } from '../../components';
import {
    useActiveDirectoryDataQualityStatsQuery,
    useActiveDirectoryPlatformsDataQualityStatsQuery,
    ActiveDirectoryQualityStat,
} from '../../hooks';

const useStyles = makeStyles({
    print: {
        '@media print': {
            display: 'none',
        },
    },
});

export const DomainInfo: React.FC<{ contextId: string; headers?: boolean; onDataError?: () => void }> = ({
    contextId,
    headers = false,
    onDataError = () => {},
}) => {
    const { data, isLoading, isError } = useActiveDirectoryDataQualityStatsQuery(contextId);
    const [domainData, setDomainData] = useState(data || null);

    useEffect(() => {
        if (data && data.data) setDomainData(data);
    }, [data, contextId]);

    useEffect(() => {
        if (isError) onDataError();
    }, [isError, onDataError]);

    if (isLoading || !domainData) {
        return <Layout dbInfo={null} isPlatform={false} headers={headers} loading={true} />;
    }

    if (isError) {
        return <Layout dbInfo={null} isPlatform={false} headers={headers} loading={false} />;
    }

    const dbInfo = domainData.data[0] as ActiveDirectoryQualityStat;

    return <Layout dbInfo={dbInfo} headers={headers} loading={false} />;
};

export const ActiveDirectoryPlatformInfo: React.FC<{ onDataError?: () => void }> = ({ onDataError = () => {} }) => {
    const { data, isLoading, isError } = useActiveDirectoryPlatformsDataQualityStatsQuery();
    const [adPlatformData, setAdPlatformData] = useState(data || null);

    useEffect(() => {
        if (data && data.data) setAdPlatformData(data);
    }, [data]);

    useEffect(() => {
        if (isError) onDataError();
    }, [isError, onDataError]);

    if (isLoading) {
        return <Layout dbInfo={null} isPlatform={true} loading={true} />;
    }

    if (isError || !adPlatformData) {
        return <Layout dbInfo={null} isPlatform={true} loading={false} />;
    }

    const dbInfo = adPlatformData.data[0] as ActiveDirectoryQualityStat;

    return <Layout dbInfo={dbInfo} isPlatform={true} loading={false} />;
};

const Layout: React.FC<{
    dbInfo: ActiveDirectoryQualityStat | null;
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
                            icon={<NodeIcon nodeType={'User'} />}
                            display='Users'
                            value={dbInfo?.users || 0}
                            loading={loading}
                        />

                        <LoadContainer
                            icon={<NodeIcon nodeType={'Group'} />}
                            display='Groups'
                            value={dbInfo?.groups || 0}
                            loading={loading}
                        />

                        <LoadContainer
                            icon={<NodeIcon nodeType={'Computer'} />}
                            display='Computers'
                            value={dbInfo?.computers || 0}
                        />

                        <LoadContainer
                            icon={<NodeIcon nodeType={'OU'} />}
                            display='OUs'
                            value={dbInfo?.ous || 0}
                            loading={loading}
                        />

                        <LoadContainer
                            icon={<NodeIcon nodeType={'GPO'} />}
                            display='GPOs'
                            value={dbInfo?.gpos || 0}
                            loading={loading}
                        />

                        <LoadContainer
                            icon={<NodeIcon nodeType={'AIACA'} />}
                            display='AIACAs'
                            value={dbInfo?.aiacas || 0}
                            loading={loading}
                        />

                        <LoadContainer
                            icon={<NodeIcon nodeType={'RootCA'} />}
                            display='RootCAs'
                            value={dbInfo?.rootcas || 0}
                            loading={loading}
                        />

                        <LoadContainer
                            icon={<NodeIcon nodeType={'EnterpriseCA'} />}
                            display='EnterpriseCAs'
                            value={dbInfo?.enterprisecas || 0}
                            loading={loading}
                        />

                        <LoadContainer
                            icon={<NodeIcon nodeType={'NTAuthStore'} />}
                            display='NTAuthStores'
                            value={dbInfo?.ntauthstores || 0}
                            loading={loading}
                        />

                        <LoadContainer
                            icon={<NodeIcon nodeType={'CertTemplate'} />}
                            display='CertTemplates'
                            value={dbInfo?.certtemplates || 0}
                            loading={loading}
                        />

                        <LoadContainer
                            icon={<NodeIcon nodeType={'Container'} />}
                            display='Containers'
                            value={dbInfo?.containers || 0}
                            loading={loading}
                        />

                        {isPlatform && (
                            <LoadContainer
                                icon={<NodeIcon nodeType={'Domain'} />}
                                display='Domains'
                                value={dbInfo?.domains || 0}
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
                            icon={<FontAwesomeIcon icon={faSignInAlt} />}
                            display='Sessions'
                            value={dbInfo?.sessions || 0}
                            loading={loading}
                        />

                        <LoadContainer
                            icon={<FontAwesomeIcon icon={faStream} />}
                            display='ACLs'
                            value={dbInfo?.acls || 0}
                            loading={loading}
                        />

                        <LoadContainer
                            icon={<FontAwesomeIcon icon={faUsers} />}
                            display='Relationships'
                            value={dbInfo?.relationships || 0}
                            loading={loading}
                        />
                    </TableBody>
                </Table>
            </TableContainer>
            <TableContainer style={{ marginTop: '16px' }} component={Paper}>
                <Table>
                    <TableBody>
                        <LoadContainer
                            icon={<FontAwesomeIcon icon={faChartPie} />}
                            display='Group Completeness'
                            value={dbInfo?.local_group_completeness || 0}
                            loading={loading}
                            type='percent'
                        />

                        <LoadContainer
                            icon={<FontAwesomeIcon icon={faChartPie} />}
                            display='Session Completeness'
                            value={dbInfo?.session_completeness || 0}
                            loading={loading}
                            type='percent'
                        />
                    </TableBody>
                </Table>
            </TableContainer>
        </Box>
    );
};
