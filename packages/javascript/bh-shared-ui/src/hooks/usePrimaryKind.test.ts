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

import { NodeKindRef } from 'js-client-library';
import { rest } from 'msw';
import { setupServer } from 'msw/node';
import { renderHook, waitFor } from '../test-utils';
import { usePrimaryKind } from './usePrimaryKind';

// custom_node_kinds rows stand in for extension node kinds flagged is_display_kind, so a kind listed here
// is treated as a display kind by the hook. The icon name must resolve to a real fas icon or
// useCustomNodeKinds drops the entry.
const displayKind = (kindName: string) => ({
    id: 1,
    kindName,
    config: { icon: { type: 'font-awesome', name: 'user', color: '#fff' } },
});

const mockCustomNodeKinds = (kindNames: string[]) =>
    rest.get('/api/v2/custom-nodes', (_req, res, ctx) => res(ctx.json({ data: kindNames.map(displayKind) })));

const mockSourceKinds = (kindNames: string[]) =>
    rest.get('/api/v2/graphs/source-kinds', (_req, res, ctx) =>
        res(ctx.json({ data: { kinds: kindNames.map((name, id) => ({ id, name })) } }))
    );

const server = setupServer(mockCustomNodeKinds([]), mockSourceKinds([]));

beforeAll(() => server.listen());
afterEach(() => server.resetHandlers());
afterAll(() => server.close());

// Renders the hook and waits until the custom-nodes/source-kinds queries have resolved so that display-kind
// prioritization has been applied before we assert on the result.
const renderPrimaryKind = async (kinds: string[] | NodeKindRef[], expected: string) => {
    const { result } = renderHook(() => usePrimaryKind(kinds));
    await waitFor(() => expect(result.current).toBe(expected));
    return result;
};

describe('usePrimaryKind', () => {
    it('returns the display kind when it appears after a non-display kind', async () => {
        server.use(mockCustomNodeKinds(['DisplayKind']));

        await renderPrimaryKind(['NonDisplayKind', 'DisplayKind'], 'DisplayKind');
    });

    it('returns the display kind when it appears before a non-display kind', async () => {
        server.use(mockCustomNodeKinds(['DisplayKind']));

        await renderPrimaryKind(['DisplayKind', 'NonDisplayKind'], 'DisplayKind');
    });

    it('returns the first display kind when multiple display kinds are present', async () => {
        server.use(mockCustomNodeKinds(['FirstDisplayKind', 'SecondDisplayKind']));

        await renderPrimaryKind(['NonDisplayKind', 'FirstDisplayKind', 'SecondDisplayKind'], 'FirstDisplayKind');
    });

    it('falls back to the first non-source, non-tag kind when no display kind is present', async () => {
        server.use(mockCustomNodeKinds([]));

        await renderPrimaryKind(['FirstKind', 'SecondKind'], 'FirstKind');
    });

    it('prioritizes a display kind over source and tag kinds regardless of order', async () => {
        server.use(mockCustomNodeKinds(['DisplayKind']), mockSourceKinds(['SourceKind']));

        await renderPrimaryKind(['SourceKind', 'Tag_MyTag', 'DisplayKind'], 'DisplayKind');
    });

    it('filters out source and tag kinds before selecting a fallback primary kind', async () => {
        server.use(mockSourceKinds(['SourceKind']));

        await renderPrimaryKind(['SourceKind', 'Tag_MyTag', 'RemainingKind'], 'RemainingKind');
    });

    it('falls back to the first raw kind when every kind is filtered out', async () => {
        server.use(mockSourceKinds(['SourceKind']));

        await renderPrimaryKind(['SourceKind'], 'SourceKind');
    });

    it('accepts NodeKindRef objects and prioritizes the display kind', async () => {
        server.use(mockCustomNodeKinds(['DisplayKind']));

        const kinds: NodeKindRef[] = [
            { node_kind_id: 1, name: 'NonDisplayKind' },
            { node_kind_id: 2, name: 'DisplayKind' },
        ];

        await renderPrimaryKind(kinds, 'DisplayKind');
    });
});
