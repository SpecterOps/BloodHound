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

import { Form, Skeleton } from '@bloodhoundenterprise/doodleui';
import {
    AssetGroupTagSelector,
    AssetGroupTagSelectorAutoCertifyAllMembers,
    AssetGroupTagSelectorAutoCertifyDisabled,
    AssetGroupTagSelectorAutoCertifySeedsOnly,
    AssetGroupTagSelectorAutoCertifyType,
    GraphNode,
    SeedTypeObjectId,
    SeedTypes,
} from 'js-client-library';
import { SelectorSeedRequest } from 'js-client-library/dist/requests';
import isEmpty from 'lodash/isEmpty';
import isEqual from 'lodash/isEqual';
import { FC, useCallback, useEffect, useReducer } from 'react';
import { SubmitHandler, useForm } from 'react-hook-form';
import { usePZPathParams } from '../../../../hooks';
import { useCreateSelector, usePatchSelector, useSelectorInfo } from '../../../../hooks/useAssetGroupTags';
import { useNotifications } from '../../../../providers';
import { apiClient, useAppNavigate } from '../../../../utils';
import { SearchValue } from '../../../Explore';
import { RulesLink } from '../../fragments';
import { handleError } from '../utils';
import BasicInfo from './BasicInfo';
import SeedSelection from './SeedSelection';
import SelectorFormContext from './SelectorFormContext';
import { SelectorFormInputs } from './types';

const diffValues = (
    data: AssetGroupTagSelector | undefined,
    formValues: SelectorFormInputs
): Partial<SelectorFormInputs> => {
    if (data === undefined) return formValues;

    const workingCopy = { ...formValues };

    const diffed: Partial<SelectorFormInputs> = {};
    const disabled = data.disabled_at !== null;

    if (data.name !== workingCopy.name) diffed.name = workingCopy.name;
    if (data.description !== workingCopy.description) diffed.description = workingCopy.description;
    if (data.auto_certify.toString() != workingCopy.auto_certify) diffed.auto_certify = workingCopy.auto_certify;
    if (workingCopy.disabled !== disabled) diffed.disabled = workingCopy.disabled;
    if (!isEqual(workingCopy.seeds, data.seeds)) diffed.seeds = workingCopy.seeds;

    return diffed;
};

/**
 * selectorStatus takes in the selectorId from the path param in the url and the selector's data.
 * It returns a boolean value associated with whether the selector is enabled or not.
 */
const selectorStatus = (id: string, data: AssetGroupTagSelector | undefined) => {
    if (id === '') return false;
    if (data === undefined) return false;
    if (typeof data.disabled_at === 'string') return false;
    if (data.disabled_at === null) return true;

    return true;
};

const parseAutoCertifyValue = (stringValue: string | undefined): AssetGroupTagSelectorAutoCertifyType | null => {
    switch (stringValue) {
        case AssetGroupTagSelectorAutoCertifyDisabled.toString():
            return AssetGroupTagSelectorAutoCertifyDisabled;
        case AssetGroupTagSelectorAutoCertifySeedsOnly.toString():
            return AssetGroupTagSelectorAutoCertifySeedsOnly;
        case AssetGroupTagSelectorAutoCertifyAllMembers.toString():
            return AssetGroupTagSelectorAutoCertifyAllMembers;
        default:
            return null;
    }
};

export type AssetGroupSelectedNode = SearchValue & { memberCount?: number };
export type AssetGroupSelectedNodes = AssetGroupSelectedNode[];

type SelectorFormState = {
    selectorType: SeedTypes;
    seeds: SelectorSeedRequest[];
    selectedObjects: AssetGroupSelectedNodes;
    autoCertify: AssetGroupTagSelectorAutoCertifyType;
};

const initialState: SelectorFormState = {
    selectorType: SeedTypeObjectId,
    seeds: [],
    selectedObjects: [],
    autoCertify: AssetGroupTagSelectorAutoCertifyDisabled,
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
            return { ...state, selectorType: action.selectorType, seeds: [], selectedObjects: [] };
        case 'set-seeds':
            return { ...state, seeds: action.seeds };
        default:
            return state;
    }
};

const SelectorForm: FC = () => {
    const { tagId, selectorId = '', tagDetailsLink, isLabelPage, tagTypeDisplay } = usePZPathParams();
    const navigate = useAppNavigate();
    const { addNotification } = useNotifications();

    const [{ selectorType, seeds, selectedObjects, autoCertify }, dispatch] = useReducer(reducer, initialState);

    const selectorQuery = useSelectorInfo(tagId, selectorId);
    const form = useForm<SelectorFormInputs>();
    const patchSelectorMutation = usePatchSelector(tagId);
    const createSelectorMutation = useCreateSelector(tagId);

    const handlePatchSelector = useCallback(async () => {
        try {
            if (!tagId || !selectorId)
                throw new Error(`Missing required entity IDs; tagId: ${tagId}, selectorId: ${selectorId}`);

            const diffedValues = diffValues(selectorQuery.data, { ...form.getValues(), seeds });

            if (isEmpty(diffedValues)) {
                addNotification(
                    'No changes to rule detected',
                    `privilege-zones_update-selector_no-changes-warn_${selectorId}`,
                    {
                        anchorOrigin: { vertical: 'top', horizontal: 'right' },
                    }
                );
                return;
            }
            await patchSelectorMutation.mutateAsync({
                tagId,
                selectorId,
                updatedValues: {
                    ...diffedValues,
                    id: parseInt(selectorId),
                    auto_certify:
                        diffedValues.auto_certify !== undefined
                            ? parseAutoCertifyValue(diffedValues.auto_certify) ?? undefined
                            : undefined,
                },
            });

            addNotification('Rule was updated successfully!', `privilege-zones_update-selector_success_${selectorId}`, {
                anchorOrigin: { vertical: 'top', horizontal: 'right' },
            });

            navigate(-1);
        } catch (error) {
            handleError(error, 'updating', 'rule', addNotification);
        }
    }, [tagId, selectorId, patchSelectorMutation, addNotification, navigate, selectorQuery.data, form, seeds]);

    const handleCreateSelector = useCallback(async () => {
        try {
            if (!tagId) throw new Error(`Missing required ID. tagId: ${tagId}`);

            const values = {
                ...form.getValues(),
                auto_certify: parseAutoCertifyValue(form.getValues().auto_certify),
                seeds,
            };

            await createSelectorMutation.mutateAsync({ tagId, values });

            addNotification('Rule was created successfully!', undefined, {
                anchorOrigin: { vertical: 'top', horizontal: 'right' },
            });

            navigate(tagDetailsLink(tagId));
        } catch (error) {
            handleError(error, 'creating', 'rule', addNotification);
        }
    }, [tagId, form, seeds, createSelectorMutation, addNotification, navigate, tagDetailsLink]);

    const onSubmit: SubmitHandler<SelectorFormInputs> = useCallback(() => {
        if (selectorId !== '') {
            handlePatchSelector();
        } else {
            handleCreateSelector();
        }
    }, [selectorId, handleCreateSelector, handlePatchSelector]);

    useEffect(() => {
        const abortController = new AbortController();

        if (selectorQuery.data) {
            const { name, description, auto_certify, seeds } = selectorQuery.data;
            form.reset({
                name,
                description,
                auto_certify: auto_certify.toString(),
                seeds,
                disabled: !selectorStatus(selectorId, selectorQuery.data),
            });

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

            if (selectorQuery.data.seeds.length > 0) {
                dispatch({ type: 'set-selector-type', selectorType: selectorQuery.data.seeds[0].type });
            }

            dispatch({ type: 'set-seeds', seeds: selectorQuery.data.seeds });

            seedsToSelectedObjects();
        }

        return () => abortController.abort();
    }, [selectorQuery.data, form, selectorId]);

    if (selectorQuery.isLoading) return <Skeleton />;

    if (selectorQuery.isError) return <div>There was an error fetching the rule information.</div>;

    return (
        <SelectorFormContext.Provider
            value={{ dispatch, seeds, selectorType, selectedObjects, selectorQuery, autoCertify }}>
            {selectorId !== '' ? (
                <p className='mt-6'>
                    {`Update this Rule's details. ${!isLabelPage ? 'Adjust criteria, analysis, or certification settings.' : ''} Changes apply to
                    the ${tagTypeDisplay} after the next analysis completes.`}
                    <br />
                    <RulesLink />.
                </p>
            ) : (
                <p className='mt-6'>
                    {`Create a new Rule to define which Objects belong to this ${tagTypeDisplay}. Use Object based Rules to choose
                    directly or Cypher based Rules to query dynamically.`}{' '}
                    <br />{' '}
                    {`You can also enable/disable the Rule${!isLabelPage ? ' and configure certification settings' : ''}.`}{' '}
                    <RulesLink />.
                </p>
            )}
            <Form {...form}>
                <form
                    onSubmit={form.handleSubmit(onSubmit)}
                    className='flex max-xl:flex-wrap gap-6 mb-6 mt-6 max-w-[120rem] justify-between pointer-events-auto'
                    data-testid='selector-form'>
                    <BasicInfo control={form.control} />
                    <SeedSelection control={form.control} />
                </form>
            </Form>
        </SelectorFormContext.Provider>
    );
};

export default SelectorForm;
