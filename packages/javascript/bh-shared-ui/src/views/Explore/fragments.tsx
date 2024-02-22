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
import useCollapsibleSectionStyles from './InfoStyles/CollapsibleSection';
import React, { PropsWithChildren } from 'react';
import { EntityField, format } from '../../utils';

const exclusionList = [
    'gid',
    'admin_rights_count',
    'admin_rights_risk_percent',
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
        (typeof value === 'object' && Object.keys(value).length === 0)
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
