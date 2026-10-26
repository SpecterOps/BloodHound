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

import { Typography, VisuallyHidden } from 'doodle-ui';
import React from 'react';
import { EntityField, cn, format } from '../../utils';

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

export const Field: React.FC<EntityField & { className?: string }> = (entityField) => {
    const { label, value, className } = entityField;

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

    return (
        <>
            {values.map((formattedItem, index) => {
                const accessibleLabel =
                    Array.isArray(value) && value.length > 0
                        ? getEmptyValueAccessibleLabel(value[index])
                        : emptyValueAccessibleLabel;

                return (
                    <div
                        className={cn('flex flex-wrap items-baseline justify-between p-2 text-left', className)}
                        key={index}>
                        {index === 0 ? (
                            <Typography component='span' variant='body2' className='mr-2 font-semibold text-text-main'>
                                {typeof label === 'string' ? label.replace(/:$/, '') : label}
                            </Typography>
                        ) : (
                            <span aria-hidden='true' />
                        )}
                        <Typography
                            component='span'
                            variant='body2'
                            className='min-w-0 max-w-full break-words'
                            title={formattedItem}>
                            {accessibleLabel ? (
                                <>
                                    <span aria-hidden='true'>{formattedItem}</span>
                                    <VisuallyHidden>{accessibleLabel}</VisuallyHidden>
                                </>
                            ) : (
                                formattedItem
                            )}
                        </Typography>
                    </div>
                );
            })}
        </>
    );
};
