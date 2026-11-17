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

import { createGraphImageExportActions } from './graphImageExport';

describe('graph image export actions', () => {
    it('uses only the capabilities the caller supplies', () => {
        const onExport = vi.fn();
        const actions = createGraphImageExportActions(
            [
                { format: 'png', scope: 'view' },
                { format: 'jpeg', scope: 'graph' },
            ],
            onExport,
            true
        );

        expect(actions.map(({ id, label, group, disabled }) => ({ id, label, group, disabled }))).toEqual([
            { id: 'png-view', label: 'PNG', group: 'Current View', disabled: true },
            { id: 'jpeg-graph', label: 'JPG', group: 'Full Graph', disabled: true },
        ]);

        actions[1].onSelect();
        expect(onExport).toHaveBeenCalledWith({ format: 'jpeg', scope: 'graph' });
    });
});
