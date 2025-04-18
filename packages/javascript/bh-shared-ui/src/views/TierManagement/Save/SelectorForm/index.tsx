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

import { isAxiosError } from 'axios';
import { SeedTypeObjectId, SeedTypes } from 'js-client-library';
import { RequestOptions, UpdateSelectorRequest } from 'js-client-library/dist/requests';
import { FC, useCallback, useState } from 'react';
import { FormProvider, SubmitHandler, useForm } from 'react-hook-form';
import { useMutation, useQueryClient } from 'react-query';
import { useNavigate, useParams } from 'react-router-dom';
import { useNotifications } from '../../../../providers';
import { apiClient } from '../../../../utils';
import BasicInfo from './BasicInfo';
import SeedSelection from './SeedSelection';
import { PatchSelectorParams, SelectorFormInputs } from './types';

const patchSelector = async (params: PatchSelectorParams, options?: RequestOptions) => {
    const { tagId, selectorId, updatedValues } = params;

    const res = await apiClient.updateAssetGroupTagSelector(tagId, selectorId, updatedValues, options);

    return res.data.data;
};

const usePatchSelector = (tagId: string | number) => {
    const queryClient = useQueryClient();
    return useMutation(patchSelector, {
        onSettled: () => {
            queryClient.invalidateQueries(['tier-management', 'tags', tagId, 'selectors']);
        },
    });
};

const SelectorForm: FC = () => {
    const { tagId = '', selectorId = '' } = useParams();

    const [selectorType, setSelectorType] = useState<SeedTypes>(SeedTypeObjectId);
    const formMethods = useForm<SelectorFormInputs>();

    const patchSelectorMutation = usePatchSelector(tagId);
    const navigate = useNavigate();
    const { addNotification } = useNotifications();

    const onSubmit: SubmitHandler<SelectorFormInputs> = (data) => handlePatchSelector(data);

    const handlePatchSelector = useCallback(
        async (updatedValues: UpdateSelectorRequest) => {
            try {
                await patchSelectorMutation.mutateAsync({ tagId, selectorId, updatedValues });

                navigate(`/tier-management/details/tags/${tagId}`);
            } catch (error) {
                console.error(error);

                if (isAxiosError(error)) {
                    addNotification(
                        `An unexpected error occurred while updating the selector. Message: ${error.response?.statusText}. Please try again.`,
                        'tier-management_patch-selector',
                        { anchorOrigin: { vertical: 'top', horizontal: 'right' } }
                    );
                }
            }
        },
        [tagId, selectorId, navigate, patchSelectorMutation, addNotification]
    );

    return (
        <FormProvider {...formMethods}>
            <form onSubmit={formMethods.handleSubmit(onSubmit)} className='flex gap-6 mt-6 w-full justify-between'>
                <BasicInfo setSelectorType={setSelectorType} selectorType={selectorType} />
                <SeedSelection selectorType={selectorType} onSubmit={onSubmit} />
            </form>
        </FormProvider>
    );
};

export default SelectorForm;
