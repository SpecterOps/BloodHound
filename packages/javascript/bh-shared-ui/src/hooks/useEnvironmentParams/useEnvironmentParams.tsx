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

import { Environment } from 'js-client-library';
import { useCallback } from 'react';
import { NavigateOptions, useSearchParams } from 'react-router-dom';
import { MappedStringLiteral } from '../../types';
import { setParamsFactory } from '../../utils/searchParams/searchParams';

export type EnvironmentAggregation = Environment['type'] | 'all';

export type EnvironmentQueryParams = {
    environmentId: Environment['id'] | null;
    environmentAggregation: EnvironmentAggregation | null;
};

export const environmentAggregationMap = {
    ['active-directory']: 'active-directory',
    azure: 'azure',
    all: 'all',
} as const satisfies MappedStringLiteral<EnvironmentAggregation, EnvironmentAggregation>;

export const parseEnvironmentAggregation = (paramValue: string | null): EnvironmentAggregation | null => {
    if (paramValue && paramValue in environmentAggregationMap) {
        return paramValue as EnvironmentAggregation;
    }
    return null;
};

interface UseEnvironmentParamsReturn extends EnvironmentQueryParams {
    setEnvironmentParams: (params: Partial<EnvironmentQueryParams>, navigateOpts?: NavigateOptions) => void;
}

export const useEnvironmentParams = (): UseEnvironmentParamsReturn => {
    const [searchParams, setSearchParams] = useSearchParams();

    return {
        environmentId: searchParams.get('environmentId'),
        environmentAggregation: parseEnvironmentAggregation(searchParams.get('environmentAggregation')),
        // react doesnt like this because it doesnt know the params needed for the function factory return function.
        // but the params needed are not needed in the deps array
        // eslint-disable-next-line react-hooks/exhaustive-deps
        setEnvironmentParams: useCallback(
            (updatedParams: Partial<EnvironmentQueryParams>, navigateOpts?: NavigateOptions) =>
                setParamsFactory(
                    setSearchParams,
                    ['environmentId', 'environmentAggregation'],
                    navigateOpts
                )(updatedParams),
            [setSearchParams]
        ),
    };
};
