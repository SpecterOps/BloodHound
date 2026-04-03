// Copyright 2026 Specter Ops, Inc.
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

import {
    Alert,
    Box,
    CircularProgress,
    Paper,
    Table,
    TableBody,
    TableCell,
    TableContainer,
    TableHead,
    TableRow,
    TextField,
    Typography,
} from '@mui/material';
import { PageWithTitle, apiClient } from 'bh-shared-ui';
import { Button } from 'doodle-ui';
import { FC, useState } from 'react';
import { useQuery } from 'react-query';

const AnonymizeLookup: FC = () => {
    const [searchInput, setSearchInput] = useState('');
    const [searchQuery, setSearchQuery] = useState('');

    const {
        data: statusData,
    } = useQuery('anonymize-status', () => apiClient.getAnonymizeStatus());

    const status = statusData?.data?.data;

    const {
        data: lookupData,
        isLoading: lookupLoading,
        isError: lookupError,
    } = useQuery(
        ['anonymize-lookup', searchQuery],
        () => apiClient.lookupAnonymized(searchQuery),
        {
            enabled: searchQuery.length > 0,
            keepPreviousData: true,
        }
    );

    const results = lookupData?.data?.data?.results ?? [];

    const handleSearch = () => {
        setSearchQuery(searchInput.trim());
    };

    const handleKeyDown = (e: React.KeyboardEvent) => {
        if (e.key === 'Enter') {
            handleSearch();
        }
    };

    return (
        <PageWithTitle
            title='Anonymize Lookup'
            data-testid='anonymize-lookup'
            pageDescription={
                <Typography variant='body2' paragraph>
                    Search for objects by their original or anonymized name to find the corresponding mapping. This
                    allows you to identify which real object an anonymized name refers to during your engagement.
                </Typography>
            }>
            <Box>
                {status && !status.anonymized && !status.backup_available && (
                    <Alert severity='info' sx={{ mb: 2 }}>
                        No anonymization has been performed yet. Anonymize your data first to use the lookup feature.
                    </Alert>
                )}

                {(status?.anonymized || status?.backup_available) && (
                    <>
                        <Box display='flex' gap={2} alignItems='flex-end' mb={3} maxWidth={600}>
                            <TextField
                                label='Search by name'
                                placeholder='Enter original or anonymized name...'
                                value={searchInput}
                                onChange={(e) => setSearchInput(e.target.value)}
                                onKeyDown={handleKeyDown}
                                fullWidth
                                variant='outlined'
                                size='small'
                            />
                            <Button
                                onClick={handleSearch}
                                disabled={searchInput.trim().length === 0 || lookupLoading}>
                                Search
                            </Button>
                        </Box>

                        {lookupLoading && <CircularProgress size={24} />}

                        {lookupError && (
                            <Alert severity='error' sx={{ mb: 2 }}>
                                Failed to search the translation table. Please check the server logs for details.
                            </Alert>
                        )}

                        {searchQuery && !lookupLoading && results.length === 0 && (
                            <Alert severity='info' sx={{ mb: 2 }}>
                                No results found for &quot;{searchQuery}&quot;.
                            </Alert>
                        )}

                        {results.length > 0 && (
                            <TableContainer component={Paper} sx={{ maxWidth: 800 }}>
                                <Table size='small'>
                                    <TableHead>
                                        <TableRow>
                                            <TableCell>
                                                <strong>Original Name</strong>
                                            </TableCell>
                                            <TableCell>
                                                <strong>Anonymized Name</strong>
                                            </TableCell>
                                            <TableCell>
                                                <strong>Object Type</strong>
                                            </TableCell>
                                        </TableRow>
                                    </TableHead>
                                    <TableBody>
                                        {results.map((row) => (
                                            <TableRow key={`${row.original_name}-${row.object_type}`}>
                                                <TableCell>{row.original_name}</TableCell>
                                                <TableCell>{row.anonymized_name}</TableCell>
                                                <TableCell>{row.object_type}</TableCell>
                                            </TableRow>
                                        ))}
                                    </TableBody>
                                </Table>
                            </TableContainer>
                        )}
                    </>
                )}
            </Box>
        </PageWithTitle>
    );
};

export default AnonymizeLookup;
