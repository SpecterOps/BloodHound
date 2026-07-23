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

import { describe, expect, it } from 'vitest';
import { createNodeKindDisplayLabelMap, getNodeKindDisplayLabel } from './nodeKindDisplay';

describe('getNodeKindDisplayLabel', () => {
    it('returns display labels from graph kind metadata', () => {
        const displayLabels = createNodeKindDisplayLabelMap([
            { name: 'AIACA', display_name: 'AIA Certificate Authority' },
            { name: 'AZApp', display_name: 'Azure Application' },
        ]);

        expect(getNodeKindDisplayLabel('AIACA', displayLabels)).toBe('AIA Certificate Authority');
        expect(getNodeKindDisplayLabel('AZApp', displayLabels)).toBe('Azure Application');
    });

    it('falls back to the raw kind for custom or unknown kinds', () => {
        expect(getNodeKindDisplayLabel('CustomNodeKind')).toBe('CustomNodeKind');
    });

    it('falls back to the raw kind when metadata has no display name', () => {
        const displayLabels = createNodeKindDisplayLabelMap([{ name: 'CustomNodeKind' }]);

        expect(getNodeKindDisplayLabel('CustomNodeKind', displayLabels)).toBe('CustomNodeKind');
    });
});
