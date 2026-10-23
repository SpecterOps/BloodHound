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

import { Alert, Typography } from 'doodle-ui';
import React, { PropsWithChildren } from 'react';
import { ActiveDirectoryKindProperties, AzureKindProperties, CommonKindProperties } from '../../graphSchema';
import { EntityField, cn, format } from '../../utils';
import { adaptClickHandlerToKeyDown } from '../../utils/adaptClickHandlerToKeyDown';
import { FieldDisplayProps } from './Field.types';
import InlineField from './InlineField';
import SpreadField from './SpreadField';

const accordionCountClassName =
    'flex h-[1.6rem] min-w-12 items-center justify-center rounded-lg bg-neutral-5 px-2 text-[0.9rem] font-bold leading-[1.6em]';

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
    'serverreferencecomputer',
    'serverreferencecomputername',
    'siteservernode',
    'siteservernodename',
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
    return (
        <div className='flex w-full items-center justify-between'>
            <Typography variant='h6' className='ml-4 text-sm leading-[3em]'>
                {label}
            </Typography>
            {isLoading ? (
                <div className={accordionCountClassName}>
                    <span
                        aria-label={`Loading ${label}`}
                        className='size-5 animate-spin rounded-full border-2 border-neutral-4 border-t-primary'
                        role='progressbar'
                    />
                </div>
            ) : isError ? (
                <Alert
                    className='h-[1.6rem] w-12 items-center justify-center p-0 [&>div:first-child]:m-0 [&>div:nth-child(2)]:sr-only'
                    variant='error'>
                    Error loading {label}
                </Alert>
            ) : (
                count !== undefined && <span className={accordionCountClassName}>{count.toLocaleString()}</span>
            )}
        </div>
    );
};

export type ObjectInfoFieldLayout = 'spread' | 'inline';

export const FieldsContainer: React.FC<PropsWithChildren<{ layout?: ObjectInfoFieldLayout }>> = ({
    children,
    layout = 'spread',
}) => {
    return (
        <div
            className={cn(
                '[&>:nth-child(even)]:bg-data-table-row-even-fill [&>:nth-child(odd)]:bg-data-table-row-odd-fill',
                layout === 'spread' && 'rounded-lg text-caption'
            )}>
            {children}
        </div>
    );
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

export const Field: React.FC<EntityField & { layout?: ObjectInfoFieldLayout }> = (entityField) => {
    const { label, value, layout = 'spread' } = entityField;

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

    const values = typeof formattedValue === 'string' ? [formattedValue] : formattedValue;
    const displayProps: FieldDisplayProps = {
        label,
        values: (values ?? []).map((formattedItem, index) => ({
            value: formattedItem,
            accessibleLabel:
                Array.isArray(value) && value.length > 0
                    ? getEmptyValueAccessibleLabel(value[index])
                    : emptyValueAccessibleLabel,
        })),
    };

    return layout === 'inline' ? (
        <InlineField {...displayProps} />
    ) : (
        <SpreadField {...displayProps} isList={Array.isArray(formattedValue)} />
    );
};

export const ObjectInfoFields: React.FC<{ fields: EntityField[]; layout?: ObjectInfoFieldLayout }> = ({
    fields,
    layout,
}): JSX.Element => {
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
                        layout={layout}
                    />
                );
            })}
        </>
    );
};
