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
import { ActiveDirectoryNodeKind } from '../../graphSchema';

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
        return <Layout stats={null} headers={headers} loading={true} />;
    }

    if (isError) {
        return <Layout stats={null} headers={headers} loading={false} />;
    }

    const stats = domainData.data[0] as ActiveDirectoryQualityStat;

    return <Layout stats={stats} headers={headers} loading={false} />;
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
        return <Layout stats={null} loading={true} />;
    }

    if (isError || !adPlatformData) {
        return <Layout stats={null} loading={false} />;
    }

    const stats = adPlatformData.data[0] as ActiveDirectoryQualityStat;

    return <Layout stats={stats} loading={false} />;
};

const Layout: React.FC<{
    stats: ActiveDirectoryQualityStat | null;
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
                            icon={<NodeIcon nodeType={ActiveDirectoryNodeKind.User} />}
                            display='Users'
                            value={stats?.users}
                            loading={loading}
                        />

                        <LoadContainer
                            icon={<NodeIcon nodeType={ActiveDirectoryNodeKind.Group} />}
                            display='Groups'
                            value={stats?.groups}
                            loading={loading}
                        />

                        <LoadContainer
                            icon={<NodeIcon nodeType={ActiveDirectoryNodeKind.Computer} />}
                            display='Computers'
                            value={stats?.computers}
                            loading={loading}
                        />

                        <LoadContainer
                            icon={<NodeIcon nodeType={ActiveDirectoryNodeKind.OU} />}
                            display='OUs'
                            value={stats?.ous}
                            loading={loading}
                        />

                        <LoadContainer
                            icon={<NodeIcon nodeType={ActiveDirectoryNodeKind.GPO} />}
                            display='GPOs'
                            value={stats?.gpos}
                            loading={loading}
                        />

                        <LoadContainer
                            icon={<NodeIcon nodeType={ActiveDirectoryNodeKind.AIACA} />}
                            display='AIACAs'
                            value={stats?.aiacas}
                            loading={loading}
                        />

                        <LoadContainer
                            icon={<NodeIcon nodeType={ActiveDirectoryNodeKind.RootCA} />}
                            display='RootCAs'
                            value={stats?.rootcas}
                            loading={loading}
                        />

                        <LoadContainer
                            icon={<NodeIcon nodeType={ActiveDirectoryNodeKind.EnterpriseCA} />}
                            display='EnterpriseCAs'
                            value={stats?.enterprisecas}
                            loading={loading}
                        />

                        <LoadContainer
                            icon={<NodeIcon nodeType={ActiveDirectoryNodeKind.NTAuthStore} />}
                            display='NTAuthStores'
                            value={stats?.ntauthstores}
                            loading={loading}
                        />

                        <LoadContainer
                            icon={<NodeIcon nodeType={ActiveDirectoryNodeKind.CertTemplate} />}
                            display='CertTemplates'
                            value={stats?.certtemplates}
                            loading={loading}
                        />
                        <LoadContainer
                            icon={<NodeIcon nodeType={'IssuancePolicy'} />}
                            display='IssuancePolicies'
                            value={dbInfo?.issuancepolicies || 0}
                            loading={loading}
                        />

                        <LoadContainer
                            icon={<NodeIcon nodeType={ActiveDirectoryNodeKind.Container} />}
                            display='Containers'
                            value={stats?.containers}
                            loading={loading}
                        />

                        {stats?.domains !== undefined && (
                            <LoadContainer
                                icon={<NodeIcon nodeType={ActiveDirectoryNodeKind.Domain} />}
                                display='Domains'
                                value={stats.domains}
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
                            value={stats?.sessions}
                            loading={loading}
                        />

                        <LoadContainer
                            icon={<FontAwesomeIcon icon={faStream} />}
                            display='ACLs'
                            value={stats?.acls}
                            loading={loading}
                        />

                        <LoadContainer
                            icon={<FontAwesomeIcon icon={faUsers} />}
                            display='Relationships'
                            value={stats?.relationships}
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
                            value={stats?.local_group_completeness}
                            loading={loading}
                            type='percent'
                        />

                        <LoadContainer
                            icon={<FontAwesomeIcon icon={faChartPie} />}
                            display='Session Completeness'
                            value={stats?.session_completeness}
                            loading={loading}
                            type='percent'
                        />
                    </TableBody>
                </Table>
            </TableContainer>
        </Box>
    );
};
