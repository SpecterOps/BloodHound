// Copyright 2026 Specter Ops, Inc.
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

import { FieldPath, FieldValues, RegisterOptions } from 'react-hook-form';
import { MAX_DESCRIPTION_LENGTH, MAX_NAME_LENGTH, MIN_NAME_LENGTH } from '../constants';

/**
 * @NOTE These rules are all exported as factories so the returned rule type matches
 * each form's field-values type inferred by react-hook-form at the call site.
 * Declared as functions (not const arrows) so the binding is hoisted and safe
 * against module circular-dependency initialization order.
 */

function rejectSurroundingWhitespace(label: string) {
    return (value: unknown) =>
        typeof value === 'string' && value !== value.trim()
            ? `${label} does not allow leading or trailing spaces`
            : true;
}

type RequiredRuleOptions = {
    /** Reject strings containing only whitespace. Defaults to false. */
    rejectSpaces?: boolean;
};

/** Shared required-field rule. See the note at the top of the file for info on why this is a function. */
export function requiredRule<TFieldValues extends FieldValues, TName extends FieldPath<TFieldValues>>(
    message: string,
    options?: RequiredRuleOptions
): RegisterOptions<TFieldValues, TName> {
    return {
        required: message,
        ...(options?.rejectSpaces
            ? { validate: (value: unknown) => typeof value !== 'string' || !!value.trim() || message }
            : {}),
    };
}

type NameRulesOptions = {
    /** Minimum length for the name field. Defaults to {@link MIN_NAME_LENGTH}. */
    minLength?: number;
    /** Maximum length for the name field. Defaults to {@link MAX_NAME_LENGTH}. */
    maxLength?: number;
    /** Whether to allow underscores and hyphens in the name field. Defaults to true. */
    allowUnderscoreAndHyphen?: boolean;
};

/** Shared name-field rules.  See the note at the top of the file for info on why this is a function.
 */
export function nameRules<TFieldValues extends FieldValues, TName extends FieldPath<TFieldValues>>(
    options?: NameRulesOptions
): RegisterOptions<TFieldValues, TName> {
    const minLength = options?.minLength ?? MIN_NAME_LENGTH;
    const maxLength = options?.maxLength ?? MAX_NAME_LENGTH;
    const shouldAllowUnderscoreAndHyphen = options?.allowUnderscoreAndHyphen ?? true;
    const pattern = shouldAllowUnderscoreAndHyphen ? /^[A-Za-z0-9 _-]+$/ : /^[A-Za-z0-9 ]+$/;
    return {
        required: 'Name is required',
        minLength: { value: minLength, message: `Name must be ${minLength} characters or more` },
        maxLength: { value: maxLength, message: `Name must be less than ${maxLength + 1} characters` },
        pattern: { value: pattern, message: 'Name must be alphanumeric.' },
        validate: rejectSurroundingWhitespace('Name'),
    };
}

type DescriptionRulesOptions = {
    /** Maximum length for the description field. Defaults to {@link MAX_DESCRIPTION_LENGTH}. */
    maxLength?: number;
};

/** Shared description-field rules. See the note at the top of the file for info on why this is a function. */
export function descriptionRules<TFieldValues extends FieldValues, TName extends FieldPath<TFieldValues>>(
    options?: DescriptionRulesOptions
): RegisterOptions<TFieldValues, TName> {
    const maxLength = options?.maxLength ?? MAX_DESCRIPTION_LENGTH;
    return {
        maxLength: { value: maxLength, message: `Description must be less than ${maxLength + 1} characters` },
        validate: rejectSurroundingWhitespace('Description'),
    };
}
