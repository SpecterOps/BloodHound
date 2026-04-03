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

import { Alert, Box, CircularProgress, Typography } from '@mui/material';
import { PageWithTitle, apiClient } from 'bh-shared-ui';
import { Button } from 'doodle-ui';
import { FC, useCallback, useEffect, useState } from 'react';
import { useMutation, useQuery, useQueryClient } from 'react-query';

const AnonymizeData: FC = () => {
    const queryClient = useQueryClient();
    const [confirmAnonymize, setConfirmAnonymize] = useState(false);
    const [confirmRestore, setConfirmRestore] = useState(false);

    const {
        data: statusData,
        isLoading: statusLoading,
        refetch: refetchStatus,
    } = useQuery('anonymize-status', () => apiClient.getAnonymizeStatus(), {
        refetchInterval: 10000,
    });

    const status = statusData?.data?.data;

    const anonymizeMutation = useMutation({
        mutationFn: () => apiClient.anonymizeData(),
        onSuccess: () => {
            setConfirmAnonymize(false);
            queryClient.invalidateQueries('anonymize-status');
        },
    });

    const restoreMutation = useMutation({
        mutationFn: () => apiClient.restoreAnonymizedData(),
        onSuccess: () => {
            setConfirmRestore(false);
            queryClient.invalidateQueries('anonymize-status');
        },
    });

    const { mutate: anonymizeMutate } = anonymizeMutation;
    const { mutate: restoreMutate } = restoreMutation;

    const handleAnonymize = useCallback(() => {
        if (!confirmAnonymize) {
            setConfirmAnonymize(true);
            return;
        }
        anonymizeMutate();
    }, [confirmAnonymize, anonymizeMutate]);

    const handleRestore = useCallback(() => {
        if (!confirmRestore) {
            setConfirmRestore(true);
            return;
        }
        restoreMutate();
    }, [confirmRestore, restoreMutate]);

    useEffect(() => {
        if (confirmAnonymize) {
            const timer = setTimeout(() => setConfirmAnonymize(false), 10000);
            return () => clearTimeout(timer);
        }
    }, [confirmAnonymize]);

    useEffect(() => {
        if (confirmRestore) {
            const timer = setTimeout(() => setConfirmRestore(false), 10000);
            return () => clearTimeout(timer);
        }
    }, [confirmRestore]);

    const isProcessing = anonymizeMutation.isLoading || restoreMutation.isLoading;

    return (
        <PageWithTitle
            title='Anonymize Data'
            data-testid='anonymize-data'
            pageDescription={
                <Typography variant='body2' paragraph>
                    Anonymize all identifiable data in the graph database to protect client confidentiality. This
                    replaces domain names, usernames, computer names, GPO names, OU names, and other identifying
                    information with generated pseudonyms. A backup is created automatically so you can restore the
                    original data at any time. Use the Anonymize Lookup view to translate between anonymized and
                    original names.
                </Typography>
            }>
            <Box>
                {statusLoading && <CircularProgress size={24} />}

                {status && (
                    <Alert severity={status.anonymized ? 'info' : 'success'} sx={{ mb: 2 }}>
                        {status.anonymized
                            ? 'Data is currently anonymized. A backup of the original data is available.'
                            : 'Data is in its original (non-anonymized) state.'}
                    </Alert>
                )}

                {anonymizeMutation.isError && (
                    <Alert severity='error' sx={{ mb: 2 }}>
                        Failed to anonymize data. Please check the server logs for details.
                    </Alert>
                )}

                {anonymizeMutation.isSuccess && (
                    <Alert severity='success' sx={{ mb: 2 }}>
                        Data has been anonymized successfully. A translation table has been created for lookups.
                    </Alert>
                )}

                {restoreMutation.isError && (
                    <Alert severity='error' sx={{ mb: 2 }}>
                        Failed to restore data. Please check the server logs for details.
                    </Alert>
                )}

                {restoreMutation.isSuccess && (
                    <Alert severity='success' sx={{ mb: 2 }}>
                        Original data has been restored successfully.
                    </Alert>
                )}

                <Box display='flex' flexDirection='column' gap={3} mt={2} maxWidth={600}>
                    <Box>
                        <Typography variant='subtitle1' fontWeight='bold' gutterBottom>
                            Anonymize
                        </Typography>
                        <Typography variant='body2' sx={{ mb: 1 }}>
                            Creates a backup of the current database, generates a translation table mapping original
                            names to anonymized names, then replaces all identifiable data with pseudonyms. This
                            includes: domain names, usernames, computer names, GPO names, OU names, group names, and
                            other AD object names.
                        </Typography>
                        <Alert severity='warning' sx={{ mb: 1 }}>
                            This operation modifies graph data in place. Ensure you are ready before proceeding.
                        </Alert>
                        <Button
                            disabled={isProcessing || (status?.anonymized ?? false)}
                            onClick={handleAnonymize}>
                            {anonymizeMutation.isLoading ? (
                                <CircularProgress size={20} sx={{ mr: 1 }} />
                            ) : null}
                            {confirmAnonymize ? 'Click again to confirm anonymization' : 'Anonymize Data'}
                        </Button>
                    </Box>

                    <Box>
                        <Typography variant='subtitle1' fontWeight='bold' gutterBottom>
                            Restore Original Data
                        </Typography>
                        <Typography variant='body2' sx={{ mb: 1 }}>
                            Restores the original graph data from the backup created during anonymization. The
                            translation table will be removed.
                        </Typography>
                        <Button
                            disabled={isProcessing || !(status?.backup_available ?? false)}
                            onClick={handleRestore}>
                            {restoreMutation.isLoading ? (
                                <CircularProgress size={20} sx={{ mr: 1 }} />
                            ) : null}
                            {confirmRestore ? 'Click again to confirm restore' : 'Restore Original Data'}
                        </Button>
                    </Box>
                </Box>
            </Box>
        </PageWithTitle>
    );
};

export default AnonymizeData;
