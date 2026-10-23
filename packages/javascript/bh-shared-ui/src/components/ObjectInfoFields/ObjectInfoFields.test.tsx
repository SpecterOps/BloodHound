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

import { render, screen } from '../../test-utils';
import { EntityField } from '../../utils';
import { ObjectInfoFields, exclusionList } from './ObjectInfoFields';

describe('ObjectInfoFields', () => {
    it('should filter properties that we do not want to display in the entity panel', () => {
        const properties: EntityField[] = exclusionList.map((key) => {
            return { label: key, value: 'test', keyprop: key };
        });
        const validProperty: EntityField = { label: 'Object ID', value: 'OBJECT_ID' };
        properties.push(validProperty);

        render(<ObjectInfoFields fields={properties} />);

        expect(screen.getByText('Object ID')).toBeInTheDocument();

        exclusionList.forEach((property) => {
            expect(screen.queryByText(property)).not.toBeInTheDocument();
        });
    });
});
