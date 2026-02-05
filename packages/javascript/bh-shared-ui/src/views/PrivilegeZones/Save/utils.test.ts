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

import { cloneDeep } from 'lodash';
import { errorSilencer } from '../../../mocks/stderr';
import { handleError } from './utils';

const mockAxiosError = {
    isAxiosError: true,
    name: 'AxiosError',
    message: 'Request failed with status code 404',
    toJSON: () => ({}),
    config: {},
    response: {
        data: {
            errors: [{ message: 'Not Found' }],
        },
        status: 404,
        statusText: 'Not Found',
        headers: {},
        config: {},
    },
};

const mockAxiosCypherError = cloneDeep(mockAxiosError);
const mockAxiosNameError = cloneDeep(mockAxiosError);

// 'any' type used as an easy way to make errors optional
const mockAxiosStatusTextOnly: any = cloneDeep(mockAxiosError);

mockAxiosNameError.response.data.errors[0].message = 'name must be unique';
mockAxiosCypherError.response.data.errors[0].message = 'seeds are required';
delete mockAxiosStatusTextOnly.response.data.errors;

const notificationDefault = 'An unexpected error occurred while creating the rule. Please try again.';
const notificationApiDefault =
    'An unexpected error occurred while creating the rule. Message: Not Found. Please try again.';

const notificationOptions = {
    anchorOrigin: {
        horizontal: 'right',
        vertical: 'top',
    },
};

describe('handleError', () => {
    const silencer = errorSilencer();
    beforeAll(() => silencer.silence());
    afterAll(() => silencer.restore());

    it('calls the provided notification method', () => {
        const handleErrorSpy = vi.fn();
        handleError({}, 'creating', 'rule', handleErrorSpy);
        expect(handleErrorSpy).toHaveBeenCalledWith(
            notificationDefault,
            'privilege-zones_creating-rule',
            notificationOptions
        );
    });

    it('reports when API error response has errors array', () => {
        const handleErrorSpy = vi.fn();
        handleError(mockAxiosError, 'creating', 'rule', handleErrorSpy);
        expect(handleErrorSpy).toHaveBeenCalledWith(
            notificationApiDefault,
            'privilege-zones_creating-rule',
            notificationOptions
        );
    });

    it('reports an API error when the name for the rule is not unique', () => {
        const notificationNameNotUnique =
            'Error creating rule: rule names must be unique. Please provide a unique name for your new rule and try again.';

        const handleErrorSpy = vi.fn();
        handleError(mockAxiosNameError, 'creating', 'rule', handleErrorSpy);
        expect(handleErrorSpy).toHaveBeenCalledWith(
            notificationNameNotUnique,
            'privilege-zones_creating-rule',
            notificationOptions
        );
    });

    it('reports an API error when Cypher is not run first', () => {
        const notificationCypherMustBeRanFirst =
            'To save a rule created using Cypher, the Cypher must be run first. Click "Run" to continue';

        const handleErrorSpy = vi.fn();
        handleError(mockAxiosCypherError, 'creating', 'rule', handleErrorSpy);
        expect(handleErrorSpy).toHaveBeenCalledWith(
            notificationCypherMustBeRanFirst,
            'privilege-zones_creating-rule',
            notificationOptions
        );
    });

    it('reports an API error when there is only statusText', () => {
        const handleErrorSpy = vi.fn();
        handleError(mockAxiosStatusTextOnly, 'creating', 'rule', handleErrorSpy);
        expect(handleErrorSpy).toHaveBeenCalledWith(
            notificationApiDefault,
            'privilege-zones_creating-rule',
            notificationOptions
        );
    });
});
