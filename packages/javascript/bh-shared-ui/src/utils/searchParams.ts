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

import { SetURLSearchParams } from 'react-router-dom';
import { nullish } from './types';

export const isNotNullish = <T>(value: T | nullish): value is T => {
    return value !== undefined && value !== null;
};

export const setSingleParamFactory = <T>(updatedParams: T, urlParams: URLSearchParams, deleteFalsy = true) => {
    return (param: keyof T) => {
        const key = param as string;
        const value = (updatedParams as Record<string, string>)[key];

        if (isNotNullish(value)) {
            urlParams.set(key, value);
        } else if (deleteFalsy) {
            urlParams.delete(key);
        }
    };
};

export const setParamsFactory = <T>(setSearchParams: SetURLSearchParams, availableParams: Array<keyof T>) => {
    return (updatedParams: T) => {
        setSearchParams((params) => {
            const setParams = setSingleParamFactory(updatedParams, params);

            availableParams.forEach((param) => setParams(param));

            return params;
        });
    };
};
