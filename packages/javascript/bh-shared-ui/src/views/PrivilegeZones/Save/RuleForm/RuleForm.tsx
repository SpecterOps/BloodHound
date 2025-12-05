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
import { useCreateRule, usePatchRule, useRuleInfo } from '../../../../hooks/useAssetGroupTags';
import { useNotifications } from '../../../../providers';
import { apiClient, useAppNavigate } from '../../../../utils';
import { SearchValue } from '../../../Explore';
import { RulesLink } from '../../fragments';
import { handleError } from '../utils';
import BasicInfo from './BasicInfo';
import RuleFormContext from './RuleFormContext';
import SeedSelection from './SeedSelection';
import { RuleFormInputs } from './types';

const diffValues = (data: AssetGroupTagSelector | undefined, formValues: RuleFormInputs): Partial<RuleFormInputs> => {
    if (data === undefined) return formValues;

    const workingCopy = { ...formValues };

    const diffed: Partial<RuleFormInputs> = {};
    const disabled = data.disabled_at !== null;

    if (data.name !== workingCopy.name) diffed.name = workingCopy.name;
    if (data.description !== workingCopy.description) diffed.description = workingCopy.description;
    if (data.auto_certify.toString() != workingCopy.auto_certify) diffed.auto_certify = workingCopy.auto_certify;
    if (workingCopy.disabled !== disabled) diffed.disabled = workingCopy.disabled;
    if (!isEqual(workingCopy.seeds, data.seeds)) diffed.seeds = workingCopy.seeds;

    return diffed;
};

/**
 * ruleStatus takes in the ruleId from the path param in the url and the rule's data.
 * It returns a boolean value associated with whether the rule is enabled or not.
 */
const ruleStatus = (id: string, data: AssetGroupTagSelector | undefined) => {
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

type RuleFormState = {
    ruleType: SeedTypes;
    seeds: SelectorSeedRequest[];
    selectedObjects: AssetGroupSelectedNodes;
    autoCertify: AssetGroupTagSelectorAutoCertifyType;
};

const initialState: RuleFormState = {
    ruleType: SeedTypeObjectId,
    seeds: [],
    selectedObjects: [],
    autoCertify: AssetGroupTagSelectorAutoCertifyDisabled,
};

export type Action =
    | { type: 'add-selected-object'; node: SearchValue }
    | { type: 'remove-selected-object'; node: SearchValue }
    | { type: 'set-selected-objects'; nodes: AssetGroupSelectedNodes }
    | { type: 'set-rule-type'; ruleType: SeedTypes }
    | { type: 'set-seeds'; seeds: SelectorSeedRequest[] };

const reducer = (state: RuleFormState, action: Action): RuleFormState => {
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
        case 'set-rule-type':
            return { ...state, ruleType: action.ruleType, seeds: [], selectedObjects: [] };
        case 'set-seeds':
            return { ...state, seeds: action.seeds };
        default:
            return state;
    }
};

const RuleForm: FC = () => {
    const { tagId, ruleId = '', tagDetailsLink, isLabelPage, tagTypeDisplay } = usePZPathParams();
    const navigate = useAppNavigate();
    const { addNotification } = useNotifications();

    const [{ ruleType, seeds, selectedObjects, autoCertify }, dispatch] = useReducer(reducer, initialState);

    const ruleQuery = useRuleInfo(tagId, ruleId);
    const form = useForm<RuleFormInputs>();
    const patchRuleMutation = usePatchRule(tagId);
    const createRuleMutation = useCreateRule(tagId);

    const handlePatchRule = useCallback(async () => {
        try {
            if (!tagId || !ruleId) throw new Error(`Missing required entity IDs; tagId: ${tagId}, ruleId: ${ruleId}`);

            const diffedValues = diffValues(ruleQuery.data, { ...form.getValues(), seeds });

            if (isEmpty(diffedValues)) {
                addNotification(
                    'No changes to rule detected',
                    `privilege-zones_update-rule_no-changes-warn_${ruleId}`,
                    {
                        anchorOrigin: { vertical: 'top', horizontal: 'right' },
                    }
                );
                return;
            }
            await patchRuleMutation.mutateAsync({
                tagId,
                ruleId,
                updatedValues: {
                    ...diffedValues,
                    id: parseInt(ruleId),
                    auto_certify:
                        diffedValues.auto_certify !== undefined
                            ? parseAutoCertifyValue(diffedValues.auto_certify) ?? undefined
                            : undefined,
                },
            });

            addNotification('Rule was updated successfully!', `privilege-zones_update-rule_success_${ruleId}`, {
                anchorOrigin: { vertical: 'top', horizontal: 'right' },
            });

            navigate(-1);
        } catch (error) {
            handleError(error, 'updating', 'rule', addNotification);
        }
    }, [tagId, ruleId, patchRuleMutation, addNotification, navigate, ruleQuery.data, form, seeds]);

    const handleCreateRule = useCallback(async () => {
        try {
            if (!tagId) throw new Error(`Missing required ID. tagId: ${tagId}`);

            const values = {
                ...form.getValues(),
                auto_certify: parseAutoCertifyValue(form.getValues().auto_certify),
                seeds,
            };

            await createRuleMutation.mutateAsync({ tagId, values });

            addNotification('Rule was created successfully!', undefined, {
                anchorOrigin: { vertical: 'top', horizontal: 'right' },
            });

            navigate(tagDetailsLink(tagId));
        } catch (error) {
            handleError(error, 'creating', 'rule', addNotification);
        }
    }, [tagId, form, seeds, createRuleMutation, addNotification, navigate, tagDetailsLink]);

    const onSubmit: SubmitHandler<RuleFormInputs> = useCallback(() => {
        if (ruleId !== '') {
            handlePatchRule();
        } else {
            handleCreateRule();
        }
    }, [ruleId, handleCreateRule, handlePatchRule]);

    useEffect(() => {
        const abortController = new AbortController();

        if (ruleQuery.data) {
            const { name, description, auto_certify, seeds } = ruleQuery.data;
            form.reset({
                name,
                description,
                auto_certify: auto_certify.toString(),
                seeds,
                disabled: !ruleStatus(ruleId, ruleQuery.data),
            });

            const seedsToSelectedObjects = async () => {
                const nodesByObjectId = new Map<string, GraphNode>();

                const seedsList = ruleQuery.data.seeds.map((seed) => {
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

                const selectedObjects = ruleQuery.data.seeds.map((seed) => {
                    const node = nodesByObjectId.get(seed.value);
                    if (node !== undefined) {
                        return { objectid: node.objectId, name: node.label, type: node.kind };
                    }
                    return { objectid: seed.value };
                });

                dispatch({ type: 'set-selected-objects', nodes: selectedObjects });
            };

            if (ruleQuery.data.seeds.length > 0) {
                dispatch({ type: 'set-rule-type', ruleType: ruleQuery.data.seeds[0].type });
            }

            dispatch({ type: 'set-seeds', seeds: ruleQuery.data.seeds });

            seedsToSelectedObjects();
        }

        return () => abortController.abort();
    }, [ruleQuery.data, form, ruleId]);

    if (ruleQuery.isLoading) return <Skeleton />;

    if (ruleQuery.isError) return <div>There was an error fetching the rule information.</div>;

    return (
        <RuleFormContext.Provider value={{ dispatch, seeds, ruleType, selectedObjects, ruleQuery, autoCertify }}>
            {ruleId !== '' ? (
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
                    data-testid='rule-form'>
                    <BasicInfo control={form.control} />
                    <SeedSelection control={form.control} />
                </form>
            </Form>
        </RuleFormContext.Provider>
    );
};

export default RuleForm;
