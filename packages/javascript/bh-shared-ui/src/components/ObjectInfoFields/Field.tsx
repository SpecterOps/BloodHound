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

import React from 'react';
import { EntityField, format } from '../../utils';
import InlineField from './InlineField';
import SpreadField from './SpreadField';
import { FieldDisplayProps, ObjectInfoFieldLayout } from './types';

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
    const isList = Array.isArray(formattedValue);

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
        <InlineField label={displayProps.label} values={displayProps.values} />
    ) : (
        <SpreadField label={displayProps.label} values={displayProps.values} isList={isList} />
    );
};
