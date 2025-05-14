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

import { AssetGroupTagNode, SeedTypeObjectId, SeedTypes } from 'js-client-library';
import { CreateSelectorRequest, RequestOptions, UpdateSelectorRequest } from 'js-client-library/dist/requests';
import { FC, useCallback, useState } from 'react';
import { FormProvider, SubmitHandler, useForm } from 'react-hook-form';
import { useMutation, useQueryClient } from 'react-query';
import { useNavigate, useParams } from 'react-router-dom';
import { ZERO_VALUE_API_DATE } from '../../../../constants';
import { useNotifications } from '../../../../providers';
import { apiClient } from '../../../../utils';
import BasicInfo from './BasicInfo';
import SeedSelection from './SeedSelection';
import { CreateSelectorParams, PatchSelectorParams, SelectorFormInputs } from './types';
import { handleError } from './utils';

const patchSelector = async (params: PatchSelectorParams, options?: RequestOptions) => {
    const { tagId, selectorId, updatedValues } = params;

    const res = await apiClient.updateAssetGroupTagSelector(tagId, selectorId, updatedValues, options);

    return res.data.data;
};

const usePatchSelector = (tagId: string | number | undefined) => {
    const queryClient = useQueryClient();
    return useMutation(patchSelector, {
        onSettled: () => {
            queryClient.invalidateQueries(['tier-management', 'tags', tagId, 'selectors']);
        },
    });
};

const createSelector = async (params: CreateSelectorParams, options?: RequestOptions) => {
    const { tagId, values } = params;

    const res = await apiClient.createAssetGroupTagSelector(tagId, values, options);

    return res.data.data;
};

const useCreateSelector = (tagId: string | number | undefined) => {
    const queryClient = useQueryClient();
    return useMutation(createSelector, {
        onSettled: () => {
            queryClient.invalidateQueries(['tier-management', 'tags', tagId, 'selectors']);
        },
    });
};

const SelectorForm: FC = () => {
    const { tierId, labelId, selectorId } = useParams();
    const tagId = labelId === undefined ? tierId : labelId;
    const navigate = useNavigate();

    const { addNotification } = useNotifications();

    const [selectorType, setSelectorType] = useState<SeedTypes>(SeedTypeObjectId);
    const [results, setResults] = useState<AssetGroupTagNode[] | null>(null);

    const formMethods = useForm<SelectorFormInputs>();

    const patchSelectorMutation = usePatchSelector(tagId);
    const createSelectorMutation = useCreateSelector(tagId);

    const handlePatchSelector = useCallback(
        async (updatedValues: UpdateSelectorRequest) => {
            try {
                if (!tagId || !selectorId)
                    throw new Error(`Missing required entity IDs; tagId: ${tagId}, selectorId: ${selectorId}`);

                if (updatedValues.disabled_at === 'on') {
                    updatedValues.disabled_at = ZERO_VALUE_API_DATE;
                }

                await patchSelectorMutation.mutateAsync({ tagId, selectorId, updatedValues });

                navigate(`/tier-management/details/tier/${tagId}`);
            } catch (error) {
                handleError(error, 'updating', addNotification);
            }
        },
        [tagId, selectorId, navigate, patchSelectorMutation, addNotification]
    );

    const handleCreateSelector = useCallback(
        async (values: CreateSelectorRequest) => {
            try {
                if (!tagId) throw new Error(`Missing required ID. tagId: ${tagId}`);

                if (values.disabled_at === 'on') {
                    values.disabled_at = ZERO_VALUE_API_DATE;
                }

                await createSelectorMutation.mutateAsync({ tagId, values });

                navigate(`/tier-management/details/tier/${tagId}`);
            } catch (error) {
                handleError(error, 'creating', addNotification);
            }
        },
        [tagId, navigate, createSelectorMutation, addNotification]
    );

    const onSubmit: SubmitHandler<SelectorFormInputs> = useCallback(
        (data) => {
            if (selectorId !== undefined) {
                handlePatchSelector(data);
            } else {
                handleCreateSelector(data);
            }
            setResults([]);
        },
        [selectorId, handleCreateSelector, handlePatchSelector, setResults]
    );

    return (
        <FormProvider {...formMethods}>
            <form
                onSubmit={formMethods.handleSubmit(onSubmit)}
                className='flex gap-6 mt-6 w-full max-w-[120rem] justify-between pointer-events-auto'>
                <BasicInfo setSelectorType={setSelectorType} selectorType={selectorType} />
                <SeedSelection
                    selectorType={selectorType}
                    results={results}
                    setResults={setResults}
                    onSubmit={onSubmit}
                />
            </form>
        </FormProvider>
    );
};

export default SelectorForm;
