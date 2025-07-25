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
import { Button } from '@bloodhoundenterprise/doodleui';
import { RequestOptions } from 'js-client-library';
import { useState } from 'react';
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

type AnalyzeNowProps = {
    description: string;
    note?: string;
};

const AnalyzeNowConfiguration: React.FC<AnalyzeNowProps> = ({ description, note }) => {
    const [isOpenDialog, setIsOpenDialog] = useState(false);
    const requestAnalysis = useRequestAnalysis();

    const { addNotification } = useNotifications();

    const showDialog = () => setIsOpenDialog(true);
    const hideDialog = () => setIsOpenDialog(false);

    const { data, isLoading, isError } = useQuery(
        'datapipe-status',
        ({ signal }) => apiClient.getDatapipeStatus({ signal }).then((res) => res.data?.data.status),
        { refetchInterval: 2000 }
    );
    const buttonDisabled = isLoading || isError || data !== 'idle';

    const handleConfirm = () => {
        requestAnalysis.mutate(undefined, {
            onError: () => {
                hideDialog();
                addNotification('There was an error requesting analysis.');
            },
            onSuccess: () => {
                hideDialog();
                addNotification('Analysis requested successfully.');
            },
        });
    };

    return (
        <>
            <div>
                <div className='flex justify-between items-center'>
                    <h4 className='font-medium text-xl '>Run Analysis Now</h4>
                    <Button disabled={buttonDisabled} onClick={showDialog}>
                        Analyze Now
                    </Button>
                </div>
                <div className='w-[677px] flex flex-col items-start gap-2 shrink-0'>
                    <p>{description}</p>
                    {note && <p className='text-xs italic'>{note}</p>}
                </div>
            </div>

            <AnalyzeNowConfirmDialog open={isOpenDialog} onCancel={hideDialog} onConfirm={handleConfirm} />
        </>
    );
};
export default AnalyzeNowConfiguration;
