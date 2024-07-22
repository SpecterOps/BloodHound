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

import { Alert, AlertTitle, Box, Grid, Link, Typography } from '@mui/material';
import {
    ActiveDirectoryPlatformInfo,
    AzurePlatformInfo,
    DataSelector,
    DomainInfo,
    PageWithTitle,
    TenantInfo,
} from 'bh-shared-ui';
import { useEffect, useState } from 'react';
import { dataCollectionMessage } from './utils';
import { useAppSelector } from 'src/store';

const QualityAssurance: React.FC = () => {
    const domain = useAppSelector((state) => state.global.options.domain);
    const [contextType, setContextType] = useState(domain?.type || null);
    const [contextId, setContextId] = useState(domain?.id || null);
    const [dataError, setDataError] = useState(false);

    useEffect(() => {
        setDataError(false);
    }, [contextId, contextType]);

    const dataErrorHandler = () => {
        setDataError(true);
    };

    const getStatsComponent = () => {
        switch (contextType) {
            case 'active-directory':
                return <DomainInfo contextId={contextId!} onDataError={dataErrorHandler} />;
            case 'active-directory-platform':
                return <ActiveDirectoryPlatformInfo onDataError={dataErrorHandler} />;
            case 'azure':
                return <TenantInfo contextId={contextId!} onDataError={dataErrorHandler} />;
            case 'azure-platform':
                return <AzurePlatformInfo onDataError={dataErrorHandler} />;
            default:
                return null;
        }
    };

    const domainErrorMessage = <>Domains unavailable. {dataCollectionMessage}</>;

    if (!contextType || (!contextId && (contextType === 'active-directory' || contextType === 'azure'))) {
        return (
            <PageWithTitle
                title='Data Quality'
                data-testid='data-quality'
                pageDescription={
                    <Typography variant='body2' paragraph>
                        Understand the data collected within BloodHound broken down by environment and principal type.
                    </Typography>
                }>
                <Box display='flex' justifyContent='flex-end' alignItems='center' minHeight='24px' mb={2}>
                    <DataSelector
                        value={{
                            type: contextType,
                            id: contextId,
                        }}
                        errorMessage={domainErrorMessage}
                        onChange={({ type, id }) => {
                            setContextType(type);
                            setContextId(id);
                        }}
                    />
                </Box>
                <Alert severity='info'>
                    <AlertTitle>No Domain or Tenant Selected</AlertTitle>
                    Select a domain or tenant to view data. If you are unable to select a domain, you may need to run
                    data collection first. {dataCollectionMessage}
                </Alert>
            </PageWithTitle>
        );
    }

    return (
        <PageWithTitle
            title='Data Quality'
            data-testid='data-quality'
            pageDescription={
                <Typography variant='body2' paragraph>
                    Understand the data collected within BloodHound broken down by environment and principal type.
                </Typography>
            }>
            <Box display='flex' justifyContent='flex-end' alignItems='center' minHeight='24px' mb={2}>
                <DataSelector
                    value={{
                        type: contextType,
                        id: contextId,
                    }}
                    errorMessage={domainErrorMessage}
                    onChange={({ type, id }) => {
                        setContextType(type);
                        setContextId(id);
                    }}
                />
            </Box>
            {dataError && (
                <Box paddingBottom={2}>
                    <Alert severity='warning'>
                        <AlertTitle>Data Quality Warning</AlertTitle>
                        It looks like data is incomplete or has not been collected yet. See the{' '}
                        <Link
                            target='_blank'
                            href={
                                'https://support.bloodhoundenterprise.io/hc/en-us/categories/9270370014875-Data-Collection'
                            }>
                            Data Collection
                        </Link>{' '}
                        page to view instructions on how to begin data collection.
                    </Alert>
                </Box>
            )}
            <Grid container spacing={2}>
                <Grid item xs={12} data-testid='data-quality_statistics'>
                    {getStatsComponent()}
                </Grid>
            </Grid>
        </PageWithTitle>
    );
};

export default QualityAssurance;
