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
import { getNodeKindDisplayLabel } from './nodeKindDisplay';

describe('getNodeKindDisplayLabel', () => {
    it('returns human-readable labels for built-in node kinds', () => {
        expect(getNodeKindDisplayLabel('AIACA')).toBe('AIA Certificate Authority');
        expect(getNodeKindDisplayLabel('AZApp')).toBe('Azure Application');
        expect(getNodeKindDisplayLabel('AZServicePrincipal')).toBe('Azure Service Principal');
        expect(getNodeKindDisplayLabel('CertTemplate')).toBe('Certificate Template');
        expect(getNodeKindDisplayLabel('EnterpriseCA')).toBe('Enterprise Certificate Authority');
        expect(getNodeKindDisplayLabel('GPO')).toBe('Group Policy Object');
    });

    it('falls back to the raw kind for custom or unknown kinds', () => {
        expect(getNodeKindDisplayLabel('CustomNodeKind')).toBe('CustomNodeKind');
    });
});
