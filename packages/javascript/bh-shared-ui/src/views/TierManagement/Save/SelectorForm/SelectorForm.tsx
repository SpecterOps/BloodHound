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

import { Skeleton } from '@bloodhoundenterprise/doodleui';
import { AssetGroupTagSelector, GraphNode, SeedTypeObjectId, SeedTypes } from 'js-client-library';
import { CreateSelectorRequest, SelectorSeedRequest, UpdateSelectorRequest } from 'js-client-library/dist/requests';
import isEmpty from 'lodash/isEmpty';
import isEqual from 'lodash/isEqual';
import { FC, useCallback, useEffect, useReducer } from 'react';
import { FormProvider, SubmitHandler, useForm } from 'react-hook-form';
import { useNavigate, useParams } from 'react-router-dom';
import { useNotifications } from '../../../../providers';
import { apiClient } from '../../../../utils';
import { SearchValue } from '../../../Explore';
import BasicInfo from './BasicInfo';
import SeedSelection from './SeedSelection';
import SelectorFormContext from './SelectorFormContext';
import { useCreateSelector, usePatchSelector, useSelectorInfo } from './hooks';
import { SelectorFormInputs } from './types';
import { handleError } from './utils';

const diffValues = (
    data: AssetGroupTagSelector | undefined,
    formValues: UpdateSelectorRequest
): UpdateSelectorRequest => {
    if (data === undefined) return formValues;

    const workingCopy = { ...formValues };

    const diffed: UpdateSelectorRequest = {};
    const disabled = data.disabled_at !== null;

    // 'on' means the switch hasn't been touched yet which means default to current disabled state
    if (workingCopy.disabled === 'on') {
        workingCopy.disabled = disabled;
    }

    if (data.name !== workingCopy.name) diffed.name = workingCopy.name;
    if (data.description !== workingCopy.description) diffed.description = workingCopy.description;
    if (workingCopy.disabled !== disabled) diffed.disabled = workingCopy.disabled;
    if (!isEqual(workingCopy.seeds, data.seeds)) diffed.seeds = workingCopy.seeds;

    return diffed;
};

export type AssetGroupSelectedNode = SearchValue & { memberCount?: number };
export type AssetGroupSelectedNodes = AssetGroupSelectedNode[];

type SelectorFormState = {
    selectorType: SeedTypes;
    seeds: SelectorSeedRequest[];
    selectedObjects: AssetGroupSelectedNodes;
};

const initialState: SelectorFormState = {
    selectorType: SeedTypeObjectId,
    seeds: [],
    selectedObjects: [],
};

export type Action =
    | { type: 'add-selected-object'; node: SearchValue }
    | { type: 'remove-selected-object'; node: SearchValue }
    | { type: 'set-selected-objects'; nodes: AssetGroupSelectedNodes }
    | { type: 'set-selector-type'; selectorType: SeedTypes }
    | { type: 'set-seeds'; seeds: SelectorSeedRequest[] };

const reducer = (state: SelectorFormState, action: Action): SelectorFormState => {
    switch (action.type) {
        case 'add-selected-object':
            return {
                ...state,
                seeds: [...state.seeds, { type: SeedTypeObjectId, value: action.node.objectid }],
                selectedObjects: [
                    ...state.selectedObjects,
                    {
                        objectid: action.node.objectid,
                        name: action.node.name || action.node.objectid,
                        type: action.node.type,
                    },
                ],
            };
        case 'remove-selected-object':
            return {
                ...state,
                seeds: state.seeds.filter((seed) => seed.value !== action.node.objectid),
                selectedObjects: state.selectedObjects.filter((node) => node.objectid !== action.node.objectid),
            };
        case 'set-selected-objects':
            return { ...state, selectedObjects: action.nodes };
        case 'set-selector-type':
            return { ...state, selectorType: action.selectorType };
        case 'set-seeds':
            return { ...state, seeds: action.seeds };
        default:
            return state;
    }
};

const SelectorForm: FC = () => {
    const { tierId = '', labelId, selectorId = '' } = useParams();
    const tagId = labelId === undefined ? tierId : labelId;
    const navigate = useNavigate();

    const { addNotification } = useNotifications();

    const [{ selectorType, seeds, selectedObjects }, dispatch] = useReducer(reducer, initialState);

    const selectorQuery = useSelectorInfo(tagId, selectorId);

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

                navigate(-1);
            } catch (error) {
                handleError(error, 'updating', addNotification);
            }
        },
        [tagId, selectorId, patchSelectorMutation, addNotification, selectorQuery.data, navigate]
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
        },
        [selectorId, handleCreateSelector, handlePatchSelector]
    );

    useEffect(() => {
        const abortController = new AbortController();
        if (selectorQuery.data) {
            const seedsToSelectedObjects = async () => {
                const nodesByObjectId = new Map<string, GraphNode>();

                const seedsList = selectorQuery.data.seeds.map((seed) => {
                    return `"${seed.value}"`;
                });

                const query = `match(n) where n.objectid in [${seedsList?.join(',')}] return n`;

                await apiClient
                    .cypherSearch(query, { signal: abortController.signal })
                    .then((res) => {
                        Object.values(res.data.data.nodes).forEach((node) => {
                            nodesByObjectId.set(node.objectId, node);
                        });
                    })
                    .catch((err) => console.error('Failed to resolve seed nodes', err));

                const selectedObjects = selectorQuery.data.seeds.map((seed) => {
                    const node = nodesByObjectId.get(seed.value);
                    if (node !== undefined) {
                        return { objectid: node.objectId, name: node.label, type: node.kind };
                    }
                    return { objectid: seed.value };
                });

                dispatch({ type: 'set-selected-objects', nodes: selectedObjects });
            };

            if (selectorQuery.data.seeds.length > 0)
                dispatch({ type: 'set-selector-type', selectorType: selectorQuery.data.seeds[0].type });
            dispatch({ type: 'set-seeds', seeds: selectorQuery.data.seeds });
            seedsToSelectedObjects();
        }
        return () => abortController.abort();
    }, [selectorQuery.data]);

    if (selectorQuery.isLoading) return <Skeleton />;

    if (selectorQuery.isError) throw new Error();

    return (
        <SelectorFormContext.Provider value={{ dispatch, seeds, selectorType, selectedObjects, selectorQuery }}>
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
