// Copyright 2023 Specter Ops, Inc.
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
import { EntityField, NodeIcon, format } from 'bh-shared-ui';
import isEmpty from 'lodash/isEmpty';
import React, { PropsWithChildren } from 'react';
import { TIER_ZERO_TAG } from 'src/constants';
import { GraphNodeTypes } from 'src/ducks/graph/types';
import { setSearchValue, startSearchSelected } from 'src/ducks/searchbar/actions';
import { PRIMARY_SEARCH, SEARCH_TYPE_EXACT } from 'src/ducks/searchbar/types';
import { useAppDispatch } from 'src/store';
import useCollapsibleSectionStyles from 'src/views/Explore/InfoStyles/CollapsibleSection';

const exclusionList = [
    'gid',
    'hasspn',
    'system_tags',
    'user_tags',
    'neo4jImportId',
    'name',
    'objectid',
    'displayname',
    'service_principal_id',
    'highvalue',
];

const filterNegatedFields = (fields: EntityField[]): EntityField[] =>
    fields.filter((field: EntityField) => !exclusionList.includes(field.keyprop || ''));

export const Section: React.FC<PropsWithChildren<{ label?: string | null; className?: string }>> = ({
    label,
    className = '',
    children,
}) => {
    return (
        <div className={className}>
            {label && (
                <Typography variant='h6'>
                    <span
                        className={'link'}
                        onClick={(e) => {
                            e.preventDefault();
                        }}>
                        {label}
                    </span>
                </Typography>
            )}
            {children}
        </div>
    );
};

export const SubHeader: React.FC<{ label: string; count?: number; isLoading?: boolean; isError?: boolean }> = ({
    label,
    count,
    isLoading = false,
    isError = false,
}) => {
    const styles = useCollapsibleSectionStyles();
    return (
        <Box display='flex' justifyContent='space-between' alignItems='center' width='100%'>
            <Typography variant='h6' className={styles.title}>
                {label}
            </Typography>
            {isLoading ? (
                <Box className={styles.accordionCount}>
                    <CircularProgress size={20} />
                </Box>
            ) : isError ? (
                <Alert
                    severity='error'
                    classes={{
                        root: styles.alertRoot,
                        icon: styles.alertIcon,
                    }}
                />
            ) : (
                count !== undefined && <span className={styles.accordionCount}>{count.toLocaleString()}</span>
            )}
        </Box>
    );
};

export const FieldsContainer: React.FC<PropsWithChildren> = ({ children }) => {
    const styles = useCollapsibleSectionStyles();
    return <div className={styles.fieldsContainer}>{children}</div>;
};

export const Field: React.FC<EntityField> = (entityField) => {
    const { label, value, keyprop } = entityField;

    if (
        value === undefined ||
        value === '' ||
        (Array.isArray(value) && value.length === 0) ||
        (typeof value === 'object' && isEmpty(value))
    )
        return null;

    const formattedValue = format(entityField);

    let content: React.ReactNode;
    if (typeof formattedValue === 'string') {
        content = (
            <Box display='flex' flexDirection='row' flexWrap='wrap' padding={1}>
                <Box flexShrink={0} flexGrow={1} fontWeight='bold' mr={1}>
                    {label}
                </Box>
                <Box overflow='hidden' textOverflow='ellipsis' title={formattedValue}>
                    {formattedValue}
                </Box>
            </Box>
        );
    } else {
        content = formattedValue!.map((value: string, index: number) => {
            return (
                <Box
                    display='flex'
                    flexDirection='row'
                    flexWrap='wrap'
                    padding={1}
                    justifyContent='flex-end'
                    key={`${keyprop}-${index}`}>
                    {index === 0 && (
                        <Box flexShrink={0} flexGrow={1} fontWeight='bold' mr={1}>
                            {label}
                        </Box>
                    )}
                    <Box overflow='hidden' textOverflow='ellipsis' title={value}>
                        {value}
                    </Box>
                </Box>
            );
        });
    }

    return <>{content}</>;
};

interface BasicObjectInfoFieldsProps {
    objectid: string;
    displayname?: string;
    system_tags?: string;
    service_principal_id?: string;
    noderesourcegroupid?: string;
    name?: string;
}

const RelatedKindField = (fieldLabel: string, relatedKind: GraphNodeTypes, id: string, name?: string) => {
    const dispatch = useAppDispatch();
    return (
        <Box padding={1}>
            <Box fontWeight='bold' mr={1}>
                {fieldLabel}
            </Box>
            <br />
            <Box display='flex' flexDirection='row' flexWrap='wrap' justifyContent='flex-start'>
                <NodeIcon nodeType={relatedKind} />
                <Box
                    onClick={() => {
                        dispatch(
                            setSearchValue(
                                {
                                    objectid: id,
                                    label: '',
                                    type: relatedKind,
                                    name: name || '',
                                },
                                PRIMARY_SEARCH,
                                SEARCH_TYPE_EXACT
                            )
                        );
                        dispatch(startSearchSelected(PRIMARY_SEARCH));
                    }}
                    style={{ cursor: 'pointer' }}
                    overflow='hidden'
                    textOverflow='ellipsis'
                    title={id}>
                    {id}
                </Box>
            </Box>
        </Box>
    );
};

export const BasicObjectInfoFields: React.FC<BasicObjectInfoFieldsProps> = (props): JSX.Element => {
    return (
        <>
            {props.system_tags?.includes(TIER_ZERO_TAG) && <Field label='Tier Zero:' value={true} />}
            {props.displayname && <Field label='Display Name:' value={props.displayname} />}
            <Field label='Object ID:' value={props.objectid} />
            {props.service_principal_id &&
                RelatedKindField(
                    'Service Principal ID:',
                    GraphNodeTypes.AZServicePrincipal,
                    props.service_principal_id,
                    props.name
                )}
            {props.noderesourcegroupid &&
                RelatedKindField(
                    'Node Resource Group ID:',
                    GraphNodeTypes.AZResourceGroup,
                    props.noderesourcegroupid,
                    props.name
                )}
        </>
    );
};

export const ObjectInfoFields: React.FC<{ fields: EntityField[] }> = ({ fields }): JSX.Element => {
    const filteredFields = filterNegatedFields(fields);

    return (
        <>
            {filteredFields.map((field: EntityField) => {
                return (
                    <Field
                        kind={field.kind}
                        label={field.label}
                        value={field.value}
                        keyprop={`${field.keyprop}`}
                        key={`${field.keyprop}-${field.label}`}
                    />
                );
            })}
        </>
    );
};
