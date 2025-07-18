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
import { Button, Card, CardContent, CardHeader, CardTitle } from '@bloodhoundenterprise/doodleui';
import { RequestOptions } from 'js-client-library';
import { FC, useState } from 'react';
import { useMutation, useQuery, useQueryClient } from 'react-query';
import { useNotifications } from '../../providers';
import { apiClient } from '../../utils';
import AnalyzeNowConfirmDialog from './AnalyzeNowConfirmDialog';

const requestAnalysis = (options?: RequestOptions) => {
    return apiClient.requestAnalysis(options).then((res) => res.data);
};

const useRequestAnalysis = () => {
    const queryClient = useQueryClient();
    return useMutation(requestAnalysis, {
        onSuccess: () => {
            queryClient.invalidateQueries('datapipe-status');
        },
    });
};

const AnalyzeNowConfiguration: FC = () => {
    const [isOpenDialog, setIsOpenDialog] = useState(false);
    const requestAnalysis = useRequestAnalysis();

    const { addNotification } = useNotifications();

    const showDialog = () => {
        setIsOpenDialog((prev) => !prev);
    };

    const { data, isLoading, isError } = useQuery(
        'datapipe-status',
        ({ signal }) => apiClient.getDatapipeStatus({ signal }).then((res) => res.data?.data.status),
        { refetchInterval: 5000 }
    );
    const buttonDisabled = isLoading || isError || data !== 'idle';

    const handleConfirm = () => {
        showDialog();
        requestAnalysis.mutate(undefined, {
            onError: () => {
                addNotification('There was an error requesting analysis.');
            },
            onSuccess: () => {
                addNotification('Analysis requested successfully.');
            },
        });
    };

    return (
        <>
            <Card>
                <CardHeader>
                    <div className='flex justify-between'>
                        <CardTitle className='font-medium'>Run Analysis Now</CardTitle>
                        <Button disabled={buttonDisabled} onClick={showDialog}>
                            Analyze Now
                        </Button>
                    </div>
                </CardHeader>
                <CardContent className='grid gap-4'>
                    This will re-run analysis in the BloodHound environment, recreating all Attack Paths that exist as a
                    result of complex configurations.
                </CardContent>
            </Card>

            <AnalyzeNowConfirmDialog open={isOpenDialog} onCancel={showDialog} onConfirm={handleConfirm} />
        </>
    );
};
export default AnalyzeNowConfiguration;
