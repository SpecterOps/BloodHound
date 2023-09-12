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

import { Alert, Box, Paper, Skeleton, Typography, useTheme } from '@mui/material';
import fileDownload from 'js-file-download';
import { useDispatch } from 'react-redux';
import { apiClient } from 'bh-shared-ui';
import { CollectorCardList, PageWithTitle } from 'bh-shared-ui';
import { addSnackbar } from 'src/ducks/global/actions';
import { CollectorType, useGetCollectorsByType } from 'src/hooks/useCollectors';

const DownloadCollectors = () => {
    /* Hooks */
    const theme = useTheme();
    const dispatch = useDispatch();
    const sharpHoundCollectorsQuery = useGetCollectorsByType('sharphound');
    const azureHoundCollectorsQuery = useGetCollectorsByType('azurehound');

    /* Event Handlers */
    const downloadCollector = (collectorType: CollectorType, version: string) => {
        apiClient
            .downloadCollector(collectorType, version)
            .then((res) => {
                const filename =
                    res.headers['content-disposition']?.match(/^.*filename="(.*)"$/)?.[1] ||
                    `${collectorType}-${version}.zip`;
                fileDownload(res.data, filename);
            })
            .catch((err) => {
                console.error(err);
                dispatch(
                    addSnackbar('This file could not be downloaded. Please try again.', 'downloadCollectorFailure')
                );
            });
    };

    const downloadCollectorChecksum = (collectorType: CollectorType, version: string) => {
        apiClient
            .downloadCollectorChecksum(collectorType, version)
            .then((res) => {
                const filename =
                    res.headers['content-disposition']?.match(/^.*filename="(.*)"$/)?.[1] ||
                    `${collectorType}-${version}.zip.sha256`;
                fileDownload(res.data, filename);
            })
            .catch((err) => {
                console.error(err);
                dispatch(
                    addSnackbar(
                        'This file could not be downloaded. Please try again.',
                        'downloadCollectorChecksumFailure'
                    )
                );
            });
    };

    /* Implementation */
    return (
        <PageWithTitle title='Download Collectors' data-testid='download-collectors'>
            <Box display='grid' gap={theme.spacing(4)}>
                {(sharpHoundCollectorsQuery.isError ||
                    azureHoundCollectorsQuery.isError ||
                    sharpHoundCollectorsQuery.data?.data.versions.length === 0) && (
                    <Alert severity='warning'>
                        A browser extension (such as an ad blocker or other privacy extension) may prevent download
                        links on this page from being displayed. Pause or disable your browser extensions and then
                        refresh this page.
                    </Alert>
                )}
                <Box>
                    <Typography variant='h2'>SharpHound</Typography>
                    {sharpHoundCollectorsQuery.isLoading ? (
                        <Paper>
                            <Box p={2}>
                                <Typography variant='h6'>
                                    <Skeleton variant='text' />
                                </Typography>
                                <Typography variant='body1'>
                                    <Skeleton variant='text' />
                                </Typography>
                            </Box>
                        </Paper>
                    ) : sharpHoundCollectorsQuery.isError ||
                      sharpHoundCollectorsQuery.data?.data.versions.length === 0 ? (
                        <Typography variant='body1'>
                            There are currently no versions of SharpHound available for download
                        </Typography>
                    ) : (
                        <CollectorCardList
                            collectors={sharpHoundCollectorsQuery
                                .data!.data.versions.map((collector) => ({
                                    collectorType: 'sharphound' as const,
                                    version: collector.version,
                                    checksum: collector.sha256sum,
                                    isLatest: collector.version === sharpHoundCollectorsQuery.data!.data.latest,
                                    isDeprecated: collector.deprecated,
                                    onClickDownload: downloadCollector,
                                    onClickDownloadChecksum: downloadCollectorChecksum,
                                }))
                                .sort((a, b) => b.version.localeCompare(a.version))}
                        />
                    )}
                </Box>
                <Box>
                    <Typography variant='h2'>AzureHound</Typography>
                    {azureHoundCollectorsQuery.isLoading ? (
                        <Paper>
                            <Box p={2}>
                                <Typography variant='h6'>
                                    <Skeleton variant='text' />
                                </Typography>
                                <Typography variant='body1'>
                                    <Skeleton variant='text' />
                                </Typography>
                            </Box>
                        </Paper>
                    ) : azureHoundCollectorsQuery.isError ||
                      azureHoundCollectorsQuery.data!.data.versions.length === 0 ? (
                        <Typography variant='body1'>
                            There are currently no versions of AzureHound available for download
                        </Typography>
                    ) : (
                        <CollectorCardList
                            collectors={azureHoundCollectorsQuery
                                .data!.data.versions.map((collector) => ({
                                    collectorType: 'azurehound' as const,
                                    version: collector.version,
                                    checksum: collector.sha256sum,
                                    isLatest: collector.version === azureHoundCollectorsQuery.data!.data.latest,
                                    isDeprecated: collector.deprecated,
                                    onClickDownload: downloadCollector,
                                    onClickDownloadChecksum: downloadCollectorChecksum,
                                }))
                                .sort((a, b) => b.version.localeCompare(a.version))}
                        />
                    )}
                </Box>
            </Box>
        </PageWithTitle>
    );
};

export default DownloadCollectors;
