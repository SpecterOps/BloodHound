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

import { FORBIDDEN } from '../../hooks';
import { useDogTag } from '../../hooks/useDogTag';

export type DogTagValue = boolean | number | string;

type DogTagProps<T extends DogTagValue> = {
    dogTagKey: string;
    loadingFallback?: JSX.Element | null;
    errorFallback?: JSX.Element | null;
    expectedValue?: T | null;
    enabled?: JSX.Element | null;
    disabled?: JSX.Element | null;
};

/**
 * DogTag Component
 * A component that will be used to return a JSX element if a DogTag value equals the defined value
 *
 * @param {string} dogTagKey - The key to the DogTag that you wish to compare against
 * @param {JSX.Element | null} loadingFallback - The JSX element to return while the data is still loading
 * @param {JSX.Element | null} errorFallback - The JSX Element to return when an error occurred retrieving the DogTag
 * @param {boolean | string | number} expectedValue - The value in which you want to compare the value of the DogTag against
 * @param {JSX.Element | null} enabled - The JSX Element to return when the DogTag value that was retrieved equals the value defined above
 * @param {JSX.Element | null} disabled - The JSX Element to return when the DogTag value that was retrieved does not equal the value defined above
 */
const DogTag = ({
    dogTagKey,
    loadingFallback = <span>Loading...</span>,
    errorFallback = <span>Error</span>,
    expectedValue = null,
    enabled = null,
    disabled = null,
}: DogTagProps<DogTagValue>) => {
    const { data: dogTag, isLoading, isError, error } = useDogTag(dogTagKey);

    if (isLoading) {
        return loadingFallback;
    }

    if (isError) {
        console.error(error);

        // Forbidden status means user can't fetch features;
        // Treat forbidden like disabled
        return error.status === FORBIDDEN ? expectedValue : errorFallback;
    }

    if (dogTag === undefined) {
        console.error(`DogTag "${dogTagKey}" not found`);
        return errorFallback;
    }

    if (dogTag == expectedValue) {
        return enabled;
    } else {
        return disabled;
    }
};

export default DogTag;
