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

import { SeedTypeCypher, SeedTypeObjectId } from 'js-client-library';
import { cloneDeep } from 'lodash';
import { errorSilencer } from '../../../mocks/stderr';
import { getErrorMessage, handleError } from './utils';

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

    it('reports an Object ID-specific error when ruleType is SeedTypeObjectId and seeds are required', () => {
        const expectedMessage = 'To create a rule using Object ID, add at least one object using the field below.';

        const handleErrorSpy = vi.fn();
        handleError(mockAxiosCypherError, 'creating', 'rule', handleErrorSpy, { ruleType: SeedTypeObjectId });
        expect(handleErrorSpy).toHaveBeenCalledWith(
            expectedMessage,
            'privilege-zones_creating-rule',
            notificationOptions
        );
    });

    it('reports a Cypher-specific error when ruleType is SeedTypeCypher and seeds are required', () => {
        const expectedMessage =
            'To save a rule created using Cypher, the Cypher must be run first. Click "Run" to continue';

        const handleErrorSpy = vi.fn();
        handleError(mockAxiosCypherError, 'creating', 'rule', handleErrorSpy, { ruleType: SeedTypeCypher });
        expect(handleErrorSpy).toHaveBeenCalledWith(
            expectedMessage,
            'privilege-zones_creating-rule',
            notificationOptions
        );
    });

    it('falls back to the default Cypher message when ruleType is undefined and seeds are required', () => {
        const expectedMessage =
            'To save a rule created using Cypher, the Cypher must be run first. Click "Run" to continue';

        const handleErrorSpy = vi.fn();
        handleError(mockAxiosCypherError, 'creating', 'rule', handleErrorSpy);
        expect(handleErrorSpy).toHaveBeenCalledWith(
            expectedMessage,
            'privilege-zones_creating-rule',
            notificationOptions
        );
    });

    it('passes ruleType through optionalParams with empty object', () => {
        const handleErrorSpy = vi.fn();
        handleError(mockAxiosCypherError, 'creating', 'rule', handleErrorSpy, {});
        // No ruleType provided, should fall back to default Cypher message
        expect(handleErrorSpy).toHaveBeenCalledWith(
            'To save a rule created using Cypher, the Cypher must be run first. Click "Run" to continue',
            'privilege-zones_creating-rule',
            notificationOptions
        );
    });
});

describe('getErrorMessage', () => {
    it('returns name uniqueness message for "name must be unique"', () => {
        const result = getErrorMessage('name must be unique', 'creating', 'rule');
        expect(result).toBe(
            'Error creating rule: rule names must be unique. Please provide a unique name for your new rule and try again.'
        );
    });

    it('returns Object ID message when ruleType is SeedTypeObjectId and seeds are required', () => {
        const result = getErrorMessage('seeds are required', 'creating', 'rule', SeedTypeObjectId);
        expect(result).toContain('Object ID');
        expect(result).toBe('To create a rule using Object ID, add at least one object using the field below.');
    });

    it('returns Cypher message when ruleType is SeedTypeCypher and seeds are required', () => {
        const result = getErrorMessage('seeds are required', 'creating', 'rule', SeedTypeCypher);
        expect(result).toContain('Cypher');
        expect(result).toContain('Click "Run" to continue');
    });

    it('returns fallback Cypher message when ruleType is undefined and seeds are required', () => {
        const result = getErrorMessage('seeds are required', 'creating', 'rule');
        expect(result).toBe(
            'To save a rule created using Cypher, the Cypher must be run first. Click "Run" to continue'
        );
    });

    it('returns default error message for unknown API messages', () => {
        const result = getErrorMessage('something unexpected', 'updating', 'zone');
        expect(result).toBe(
            'An unexpected error occurred while updating the zone. Message: something unexpected. Please try again.'
        );
    });

    it('uses the correct entity name in the Object ID message', () => {
        const result = getErrorMessage('seeds are required', 'updating', 'zone', SeedTypeObjectId);
        expect(result).toBe(
            'To save a zone created using Object ID, the Object ID must be run first. Click "Run" to continue'
        );
    });
});
