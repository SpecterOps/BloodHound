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

import { createBrowserHistory } from 'history';
import { AssetGroupTagNode, AssetGroupTagSelector, SeedTypeObjectId, SeedTypes } from 'js-client-library';
import {
    CreateSelectorRequest,
    RequestOptions,
    SelectorSeedRequest,
    UpdateSelectorRequest,
} from 'js-client-library/dist/requests';
import isEmpty from 'lodash/isEmpty';
import isEqual from 'lodash/isEqual';
import { FC, useCallback, useState } from 'react';
import { FormProvider, SubmitHandler, useForm } from 'react-hook-form';
import { useMutation, useQuery, useQueryClient } from 'react-query';
import { useNavigate, useParams } from 'react-router-dom';
import { useNotifications } from '../../../../providers';
import { apiClient } from '../../../../utils';
import BasicInfo from './BasicInfo';
import SeedSelection from './SeedSelection';
import SelectorFormContext from './SelectorFormContext';
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

const diffValues = (
    data: AssetGroupTagSelector | undefined,
    formValues: UpdateSelectorRequest
): UpdateSelectorRequest => {
    if (data === undefined) return formValues;

    const diffed: UpdateSelectorRequest = {};
    const disabled = data.disabled_at !== null;

    // 'on' means the switch hasn't been touched yet which means default to current disabled state
    if (formValues.disabled === 'on') {
        formValues.disabled = disabled;
    }

    if (data.name !== formValues.name) diffed.name = formValues.name;
    if (data.description !== formValues.description) diffed.description = formValues.description;
    if (formValues.disabled !== disabled) diffed.disabled = formValues.disabled;
    if (!isEqual(formValues.seeds, data.seeds)) diffed.seeds = formValues.seeds;

    return diffed;
};

const SelectorForm: FC = () => {
    const { tierId = '', labelId, selectorId = '' } = useParams();
    const tagId = labelId === undefined ? tierId : labelId;
    const history = createBrowserHistory();
    const navigate = useNavigate();

    const { addNotification } = useNotifications();

    const selectorQuery = useQuery({
        queryKey: ['tier-management', 'tags', tagId, 'selectors', selectorId],
        queryFn: async () => {
            const response = await apiClient.getAssetGroupTagSelector(tagId, selectorId);
            return response.data.data['selector'];
        },
        enabled: selectorId !== '',
    });

    const [selectorType, setSelectorType] = useState<SeedTypes>(SeedTypeObjectId);
    const [results, setResults] = useState<AssetGroupTagNode[] | null>(null);
    const [seeds, setSeeds] = useState<SelectorSeedRequest[]>(selectorQuery.data?.seeds || []);

    const formMethods = useForm<SelectorFormInputs>();

    const patchSelectorMutation = usePatchSelector(tagId);
    const createSelectorMutation = useCreateSelector(tagId);

    const handlePatchSelector = useCallback(
        async (updatedValues: UpdateSelectorRequest) => {
            try {
                if (!tagId || !selectorId)
                    throw new Error(`Missing required entity IDs; tagId: ${tagId}, selectorId: ${selectorId}`);

                const diffedValues = diffValues(selectorQuery.data, updatedValues);

                if (isEmpty(diffedValues)) {
                    addNotification(
                        'No changes to selector detected',
                        `tier-management_update-selector_no-changes-warn_${selectorId}`,
                        {
                            anchorOrigin: { vertical: 'top', horizontal: 'right' },
                        }
                    );
                    return;
                }

                await patchSelectorMutation.mutateAsync({ tagId, selectorId, updatedValues: diffedValues });

                addNotification(
                    'Selector was updated successfully!',
                    `tier-management_update-selector_success_${selectorId}`,
                    {
                        anchorOrigin: { vertical: 'top', horizontal: 'right' },
                    }
                );

                history.back();
            } catch (error) {
                handleError(error, 'updating', addNotification);
            }
        },
        [tagId, selectorId, patchSelectorMutation, addNotification, history, selectorQuery.data]
    );

    const handleCreateSelector = useCallback(
        async (values: CreateSelectorRequest) => {
            try {
                if (!tagId) throw new Error(`Missing required ID. tagId: ${tagId}`);

                await createSelectorMutation.mutateAsync({ tagId, values });

                addNotification('Selector was created successfully!', undefined, {
                    anchorOrigin: { vertical: 'top', horizontal: 'right' },
                });

                navigate(`/tier-management/details/tier/${tagId}`);
            } catch (error) {
                handleError(error, 'creating', addNotification);
            }
        },
        [tagId, navigate, createSelectorMutation, addNotification]
    );

    const onSubmit: SubmitHandler<SelectorFormInputs> = useCallback(
        (data) => {
            if (selectorId !== '') {
                handlePatchSelector(data);
            } else {
                handleCreateSelector(data);
            }
            setResults([]);
        },
        [selectorId, handleCreateSelector, handlePatchSelector, setResults]
    );

    return (
        <SelectorFormContext.Provider
            value={{ seeds, setSeeds, results, setResults, selectorType, setSelectorType, selectorQuery }}>
            <FormProvider {...formMethods}>
                <form
                    onSubmit={formMethods.handleSubmit(onSubmit)}
                    className='flex gap-6 mt-6 w-full max-w-[120rem] justify-between pointer-events-auto'>
                    <BasicInfo />
                    <SeedSelection onSubmit={onSubmit} />
                </form>
            </FormProvider>
        </SelectorFormContext.Provider>
    );
};

export default SelectorForm;
