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

import { Alert, CircularProgress } from '@mui/material';
import { Typography, VisuallyHidden } from 'doodle-ui';
import React, { PropsWithChildren } from 'react';
import { ActiveDirectoryKindProperties, AzureKindProperties, CommonKindProperties } from '../../graphSchema';
import { EntityField, format } from '../../utils';
import { adaptClickHandlerToKeyDown } from '../../utils/adaptClickHandlerToKeyDown';
import useCollapsibleSectionStyles from './InfoStyles/CollapsibleSection';

export const exclusionList = [
    'gid',
    'admin_rights_count',
    'admin_rights_risk_percent',
    ActiveDirectoryKindProperties.HasSPN,
    CommonKindProperties.SystemTags,
    'isTierZero',
    CommonKindProperties.UserTags,
    'neo4jImportId',
    CommonKindProperties.Name,
    CommonKindProperties.ObjectID,
    CommonKindProperties.DisplayName,
    AzureKindProperties.ServicePrincipalID,
    AzureKindProperties.FederatedIdentityCredentialAppID,
    'highvalue',
    'reconcile',
    ActiveDirectoryKindProperties.InheritanceHashes,
    ActiveDirectoryKindProperties.InheritanceHash,
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
                        role='button'
                        aria-label={label}
                        tabIndex={0}
                        className={'link'}
                        onClick={(e) => {
                            e.preventDefault();
                        }}
                        onKeyDown={adaptClickHandlerToKeyDown((e) => e.preventDefault())}>
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
        <div className='flex w-full items-center justify-between'>
            <Typography variant='h6' className={styles.title}>
                {label}
            </Typography>
            {isLoading ? (
                <div className={styles.accordionCount}>
                    <CircularProgress size={20} />
                </div>
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
        </div>
    );
};

export const FieldsContainer: React.FC<PropsWithChildren> = ({ children }) => {
    const styles = useCollapsibleSectionStyles();
    return <div className={styles.fieldsContainer}>{children}</div>;
};

const getEmptyValueAccessibleLabel = (value: EntityField['value']): string | undefined => {
    if (Array.isArray(value) && value.length === 0) {
        return 'Empty array, zero values';
    }

    if (value === '') {
        return 'Empty string';
    }

    if (value === null) {
        return 'Null value';
    }
};

export const Field: React.FC<EntityField> = (entityField) => {
    const { label, value, keyprop } = entityField;

    try {
        if (
            value === undefined ||
            (value !== null && !Array.isArray(value) && typeof value === 'object' && Object.keys(value).length === 0)
        )
            return null;
    } catch (e) {
        return null;
    }

    const emptyValueAccessibleLabel = getEmptyValueAccessibleLabel(value);
    const formattedValue = format(entityField);

    let content: React.ReactNode;
    if (typeof formattedValue === 'string') {
        content = (
            <div className='flex flex-row flex-wrap p-2'>
                <div className='mr-2 grow shrink-0 font-bold'>{label}</div>
                <div className='overflow-hidden text-ellipsis break-all' title={formattedValue}>
                    {emptyValueAccessibleLabel ? (
                        <>
                            <span aria-hidden='true'>{formattedValue}</span>
                            <VisuallyHidden>{emptyValueAccessibleLabel}</VisuallyHidden>
                        </>
                    ) : (
                        formattedValue
                    )}
                </div>
            </div>
        );
    } else {
        content = formattedValue!.map((value: string, index: number) => {
            return (
                <div className='flex flex-row flex-wrap justify-end p-2' key={`${keyprop}-${index}`}>
                    {index === 0 && <div className='mr-2 grow shrink-0 font-bold'>{label}</div>}
                    <div className='overflow-hidden text-ellipsis break-all' title={value}>
                        {emptyValueAccessibleLabel ? (
                            <>
                                <span aria-hidden='true'>{value}</span>
                                <VisuallyHidden>{emptyValueAccessibleLabel}</VisuallyHidden>
                            </>
                        ) : (
                            value
                        )}
                    </div>
                </div>
            );
        });
    }

    return <>{content}</>;
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
