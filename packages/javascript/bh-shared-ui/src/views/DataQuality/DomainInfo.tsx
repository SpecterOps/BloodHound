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

import { faChartPie, faSignInAlt, faStream, faUsers } from '@fortawesome/free-solid-svg-icons';
import { FontAwesomeIcon } from '@fortawesome/react-fontawesome';
import { Box, Paper, Table, TableBody, TableCell, TableContainer, TableHead, TableRow } from '@mui/material';
import makeStyles from '@mui/styles/makeStyles';
import { ActiveDirectoryQualityStat } from 'js-client-library';
import React, { useEffect, useState } from 'react';
import { NodeIcon } from '../../components';
import { ActiveDirectoryNodeKind } from '../../graphSchema';
import { useActiveDirectoryDataQualityStatsQuery, useActiveDirectoryPlatformsDataQualityStatsQuery } from '../../hooks';
import LoadContainer from './LoadContainer';

const useStyles = makeStyles({
    print: {
        '@media print': {
            display: 'none',
        },
    },
});

export const DomainMap = {
    users: { displayText: 'Users', kind: ActiveDirectoryNodeKind.User },
    groups: { displayText: 'Groups', kind: ActiveDirectoryNodeKind.Group },
    computers: { displayText: 'Computers', kind: ActiveDirectoryNodeKind.Computer },
    ous: {
        displayText: 'OUs',
        kind: ActiveDirectoryNodeKind.OU,
    },
    gpos: { displayText: 'GPOs', kind: ActiveDirectoryNodeKind.GPO },
    aiacas: {
        displayText: 'AIACAs',
        kind: ActiveDirectoryNodeKind.AIACA,
    },
    rootcas: { displayText: 'RootCAs', kind: ActiveDirectoryNodeKind.RootCA },
    enterprisecas: { displayText: 'EnterpriseCAs', kind: ActiveDirectoryNodeKind.EnterpriseCA },
    ntauthstores: { displayText: 'NTAuthStores', kind: ActiveDirectoryNodeKind.NTAuthStore },
    certtemplates: { displayText: 'CertTemplates', kind: ActiveDirectoryNodeKind.CertTemplate },
    issuancepolicies: { displayText: 'IssuancePolicies', kind: ActiveDirectoryNodeKind.IssuancePolicy },
    containers: {
        displayText: 'Containers',
        kind: ActiveDirectoryNodeKind.Container,
    },
    domains: {
        displayText: 'Domains',
        kind: ActiveDirectoryNodeKind.Domain,
    },
};

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

    const stats = domainData.data[0];

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

    const stats = adPlatformData.data[0];

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
                        {Object.keys(DomainMap).map((key) => {
                            if (key === 'domains' && stats?.domains === undefined) return null;

                            const mapValue = DomainMap[key as keyof typeof DomainMap];
                            const value = stats?.[key as keyof ActiveDirectoryQualityStat] as number;

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
