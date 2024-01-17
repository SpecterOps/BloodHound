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

import { ActiveDirectoryKindProperties, AzureKindProperties, CommonKindProperties } from 'bh-shared-ui';
import { validateProperty } from './utils';

describe('validating a node property against the shared generated schema', () => {
    it('should recognize active directory properties', () => {
        Object.values(ActiveDirectoryKindProperties).forEach((property: ActiveDirectoryKindProperties) => {
            expect(validateProperty(property)).toEqual({ isKnownProperty: true, kind: 'ad' });
        });
    });
    it('should recognize azure properties', () => {
        Object.values(AzureKindProperties).forEach((property: AzureKindProperties) => {
            expect(validateProperty(property)).toEqual({ isKnownProperty: true, kind: 'az' });
        });
    });
    it('should recognize a common properties', () => {
        Object.values(CommonKindProperties).forEach((property: CommonKindProperties) => {
            expect(validateProperty(property)).toEqual({ isKnownProperty: true, kind: 'cm' });
        });
    });
    it('should return an object denoting that the property is not in the schema when it is unrecognized', () => {
        expect(validateProperty('notInSchema')).toEqual({ isKnownProperty: false, kind: null });
    });
});
