// Copyright 2025 Specter Ops, Inc.
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

import { Alert, AlertTitle, Box, Grid, Link } from '@mui/material';
import {
    ActiveDirectoryPlatformInfo,
    AzurePlatformInfo,
    DomainInfo,
    LoadingOverlay,
    PageWithTitle,
    TenantInfo,
    useDataQualityEnvironmentsQuery,
} from 'bh-shared-ui';
import { Typography } from 'doodle-ui';
import { useEffect, useMemo, useState } from 'react';
import DataQualityEnvironmentSelector, {
    DataQualitySelection,
    dataQualityTypeFromEnvironmentKind,
} from './DataQualityEnvironmentSelector';
import OpenGraphNodeKindCounts from './OpenGraphNodeKindCounts';
import { dataCollectionMessage } from './utils';

const isBuiltInDataQualityType = (type?: string) => type === 'active-directory' || type === 'azure';

const getStatsComponent = (selectedEnvironment: DataQualitySelection | null, dataErrorHandler: () => void) => {
    const contextType = selectedEnvironment
        ? dataQualityTypeFromEnvironmentKind(selectedEnvironment.environmentKind)
        : undefined;
    const contextId = selectedEnvironment?.environmentId;

    if (contextType === 'active-directory') {
        if (selectedEnvironment?.selectionType === 'aggregate') {
            return <ActiveDirectoryPlatformInfo onDataError={dataErrorHandler} />;
        }
        if (!contextId) return null;
        return <DomainInfo contextId={contextId} onDataError={dataErrorHandler} />;
    }

    if (contextType === 'azure') {
        if (selectedEnvironment?.selectionType === 'aggregate') {
            return <AzurePlatformInfo onDataError={dataErrorHandler} />;
        }
        if (!contextId) return null;
        return <TenantInfo contextId={contextId} onDataError={dataErrorHandler} />;
    }

    return null;
};

const DataQuality: React.FC = () => {
    const { data: environmentsResponse, isLoading, isError } = useDataQualityEnvironmentsQuery();
    const environments = environmentsResponse?.data ?? [];

    const initialEnvironment = useMemo(() => {
        return [...environments].sort((first, second) =>
            first.environment_name.localeCompare(second.environment_name)
        )[0];
    }, [environments]);

    const initialSelection = useMemo<DataQualitySelection | null>(() => {
        if (!initialEnvironment) return null;

        return {
            environmentId: initialEnvironment.environment_id,
            environmentKind: initialEnvironment.environment_kind,
            selectionType: 'environment',
        };
    }, [initialEnvironment]);

    const [selectedEnvironment, setSelectedEnvironment] = useState<DataQualitySelection | null>(null);
    const environment = selectedEnvironment ?? initialSelection;
    const environmentType = environment ? dataQualityTypeFromEnvironmentKind(environment.environmentKind) : undefined;

    const [dataError, setDataError] = useState(false);

    useEffect(() => {
        if (!selectedEnvironment && initialSelection) {
            setSelectedEnvironment(initialSelection);
        }
    }, [initialSelection, selectedEnvironment]);

    useEffect(() => {
        setDataError(false);
    }, [environment]);

    const dataErrorHandler = () => {
        setDataError(true);
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

    if (!environment || !environmentType) {
        return (
            <PageWithTitle
                title='Data Quality'
                data-testid='data-quality'
                pageDescription={<QualityAssuranceDescription />}>
                <Box display='flex' justifyContent='flex-end' alignItems='center' minHeight='24px' mb={2}>
                    <DataQualityEnvironmentSelector
                        align='end'
                        environments={environments}
                        errorMessage={environmentErrorMessage}
                        isError={isError}
                        isLoading={isLoading}
                        selected={environment}
                        onSelect={setSelectedEnvironment}
                    />
                </Box>
                <Alert severity='info'>
                    <AlertTitle>No Environment Selected</AlertTitle>
                    Select an environment to view data. If you are unable to select an environment, you may need to run
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
                <DataQualityEnvironmentSelector
                    align='end'
                    environments={environments}
                    errorMessage={environmentErrorMessage}
                    isError={isError}
                    isLoading={isLoading}
                    selected={environment}
                    onSelect={setSelectedEnvironment}
                />
            </Box>
            {dataError && (
                <Box paddingBottom={2}>
                    <Alert severity='warning'>
                        <AlertTitle>Data Quality Warning</AlertTitle>
                        It looks like data is incomplete or has not been collected yet. See the{' '}
                        <Link
                            target='_blank'
                            href={'https://bloodhound.specterops.io/collect-data/overview#bloodhound-ce-collection'}>
                            Data Collection
                        </Link>{' '}
                        page to view instructions on how to begin data collection.
                    </Alert>
                </Box>
            )}
            <Grid container spacing={2}>
                <Grid item xs={12} data-testid='data-quality_statistics'>
                    {getStatsComponent(environment, dataErrorHandler)}
                    <OpenGraphNodeKindCounts
                        hideEmptyState={isBuiltInDataQualityType(environmentType)}
                        selectedEnvironment={environment}
                    />
                </Grid>
            </Grid>
        </PageWithTitle>
    );
};

export default DataQuality;

const QualityAssuranceDescription = () => (
    <Typography variant='body2'>
        Understand the data collected within BloodHound broken down by environment and principal type.
    </Typography>
);
