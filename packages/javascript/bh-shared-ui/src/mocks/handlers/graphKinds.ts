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

import { rest } from 'msw';
import { createGraphKinds } from '../factories/graphKinds';

export const mockKindsHandler = (
    nodeKinds?: string[],
    edgeKinds?: string[],
    nodeKindDisplayNames: Record<string, string> = {}
) =>
    rest.get('/api/v2/graphs/kinds', async (req, res, ctx) => {
        const filter = req.url.searchParams.get('type')?.split(':')[1];

        if (filter === 'node') {
            return res(ctx.json({ data: createGraphKinds(nodeKinds, [], nodeKindDisplayNames) }));
        }

        if (filter === 'edge' || filter === 'relationship') {
            return res(ctx.json({ data: createGraphKinds([], edgeKinds, nodeKindDisplayNames) }));
        }

        return res(ctx.json({ data: createGraphKinds(nodeKinds, edgeKinds, nodeKindDisplayNames) }));
    });

export const mockSourceKindsHandler = () =>
    rest.get('/api/v2/graphs/source-kinds', (_req, res, ctx) => {
        return res(ctx.json({ data: { kinds: [] } }));
    });
