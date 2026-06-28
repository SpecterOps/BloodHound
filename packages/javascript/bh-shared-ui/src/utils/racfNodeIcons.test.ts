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

import { GetIconInfo } from './icons';
import { RACF_NODE_ICONS, RACF_NODE_KINDS } from './racfNodeIcons';

describe('RACF node icons', () => {
    it.each(Object.values(RACF_NODE_KINDS))('defines matching legacy and namespaced icons for %s', (kind) => {
        expect(RACF_NODE_ICONS[kind]).toBeDefined();
        expect(RACF_NODE_ICONS[`racf_${kind}`]).toEqual(RACF_NODE_ICONS[kind]);
    });

    it('allows a server-provided custom type to override the default', () => {
        const customIcon = {
            icon: RACF_NODE_ICONS[RACF_NODE_KINDS.Group].icon,
            color: '#123456',
        };

        expect(GetIconInfo(RACF_NODE_KINDS.User, { [RACF_NODE_KINDS.User]: customIcon })).toEqual(customIcon);
    });
});
