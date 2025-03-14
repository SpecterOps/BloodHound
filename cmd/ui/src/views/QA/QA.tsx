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
import makeStyles from '@mui/styles/makeStyles';
import {
    ActiveDirectoryPlatformInfo,
    AzurePlatformInfo,
    DataSelector,
    DomainInfo,
    LoadingOverlay,
    PageWithTitle,
    SelectedEnvironment,
    TenantInfo,
} from 'bh-shared-ui';
import { useEffect, useState } from 'react';
import { useInitialEnvironment } from 'src/hooks/useInitialEnvironment';
import { dataCollectionMessage } from './utils';

const useStyles = makeStyles((theme) => ({
    container: {
        '& div:first-child': {
            backgroundColor: theme.palette.neutral.tertiary,
        },
    },
}));

const QualityAssuranceV2: React.FC = () => {
    const { data: initialEnvironment, isLoading } = useInitialEnvironment();

    const [selectedEnvironment, setSelectedEnvironment] = useState<SelectedEnvironment | null>(
        initialEnvironment ?? null
    );
    const [dataError, setDataError] = useState(false);
    const classes = useStyles();

    const environment = selectedEnvironment ?? initialEnvironment;
    console.log(environment);

    useEffect(() => {
        setDataError(false);
    }, [environment?.type, environment?.id]);

    const dataErrorHandler = () => {
        setDataError(true);
    };

    const getStatsComponent = () => {
        const contextId = environment?.id;

        switch (environment?.type) {
            case 'active-directory':
                if (!contextId) return null;
                return <DomainInfo contextId={contextId} onDataError={dataErrorHandler} />;
            case 'active-directory-platform':
                return <ActiveDirectoryPlatformInfo onDataError={dataErrorHandler} />;
            case 'azure':
                if (!contextId) return null;
                return <TenantInfo contextId={contextId} onDataError={dataErrorHandler} />;
            case 'azure-platform':
                return <AzurePlatformInfo onDataError={dataErrorHandler} />;
            default:
                return null;
        }
    };

    const environmentErrorMessage = <>Environments unavailable. {dataCollectionMessage}</>;

    if (isLoading) {
        return (
            <PageWithTitle
                title='Data Quality'
                data-testid='data-quality'
                pageDescription={
                    <>
                        <QualityAssuranceDescription />
                        <LoadingOverlay loading />
                    </>
                }
            />
        );
    }

    if (
        !environment?.type ||
        (!environment?.id && (environment?.type === 'active-directory' || environment?.type === 'azure'))
    ) {
        return (
            <PageWithTitle
                title='Data Quality'
                data-testid='data-quality'
                pageDescription={<QualityAssuranceDescription />}>
                <Box display='flex' justifyContent='flex-end' alignItems='center' minHeight='24px' mb={2}>
                    <DataSelector
                        value={{
                            type: environment?.type ?? null,
                            id: environment?.id ?? null,
                        }}
                        errorMessage={environmentErrorMessage}
                        onChange={(selection) => setSelectedEnvironment(selection)}
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
            pageDescription={<QualityAssuranceDescription />}>
            <Box display='flex' justifyContent='flex-end' alignItems='center' minHeight='24px' mb={2}>
                <DataSelector
                    value={selectedEnvironment || initialEnvironment || { type: null, id: null }}
                    errorMessage={environmentErrorMessage}
                    onChange={(selection) => setSelectedEnvironment({ ...selection })}
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
                <Grid item xs={12} data-testid='data-quality_statistics' classes={classes.container}>
                    {getStatsComponent()}
                </Grid>
            </Grid>
        </PageWithTitle>
    );
};

export default QualityAssuranceV2;

const QualityAssuranceDescription = () => (
    <Typography variant='body2' paragraph>
        Understand the data collected within BloodHound broken down by environment and principal type.
    </Typography>
);
