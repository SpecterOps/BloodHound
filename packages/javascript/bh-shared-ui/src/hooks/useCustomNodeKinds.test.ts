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

import { CustomNodeKindType } from 'js-client-library';
import { RACF_NODE_ICONS, RACF_NODE_KINDS } from '../utils';
import { createCustomIconDictionary } from './useCustomNodeKinds';

const createCustomNodeKind = (kindName: string, name: string, color: string): CustomNodeKindType => ({
    id: 1,
    kindName,
    config: {
        icon: {
            type: 'font-awesome',
            name,
            color,
        },
    },
});

describe('createCustomIconDictionary', () => {
    it('uses RACF defaults when the API has no definitions', () => {
        const icons = createCustomIconDictionary(undefined);

        expect(icons[RACF_NODE_KINDS.User]).toEqual(RACF_NODE_ICONS[RACF_NODE_KINDS.User]);
        expect(icons[`racf_${RACF_NODE_KINDS.User}`]).toEqual(RACF_NODE_ICONS[RACF_NODE_KINDS.User]);
    });

    it('does not let an automatically generated stub replace a RACF default', () => {
        const icons = createCustomIconDictionary([
            createCustomNodeKind(`racf_${RACF_NODE_KINDS.Certificate}`, 'question', '#FFFFFF'),
        ]);

        expect(icons[`racf_${RACF_NODE_KINDS.Certificate}`]).toEqual(RACF_NODE_ICONS[RACF_NODE_KINDS.Certificate]);
    });

    it('allows a provisioned custom type to override a RACF default', () => {
        const icons = createCustomIconDictionary([createCustomNodeKind(RACF_NODE_KINDS.User, 'house', '#123456')]);

        expect(icons[RACF_NODE_KINDS.User].icon.iconName).toBe('house');
        expect(icons[RACF_NODE_KINDS.User].color).toBe('#123456');
    });
});
