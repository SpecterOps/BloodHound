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
import { createServer } from 'node:http';
import type { AddressInfo } from 'node:net';
import { test as mockedTest, expect } from '../src/mocked';

const test = mockedTest.extend<{ liveServer: { url: string; apiRequests: number } }>({
    liveServer: async ({ browserName }, use) => {
        void browserName;
        const state = { url: '', apiRequests: 0 };
        const server = createServer((request, response) => {
            if (request.url?.startsWith('/api/')) {
                state.apiRequests++;
                response.writeHead(500).end('Backend must not be used');
                return;
            }
            response.setHeader('Content-Type', 'text/html');
            response.end('<h1>Replay test</h1>');
        });
        await new Promise<void>((resolve) => server.listen(0, '127.0.0.1', resolve));
        state.url = `http://127.0.0.1:${(server.address() as AddressInfo).port}`;
        try {
            await use(state);
        } finally {
            server.closeAllConnections();
            await new Promise<void>((resolve, reject) => server.close((error) => (error ? reject(error) : resolve())));
        }
    },
    baseURL: async ({ liveServer }, use) => use(liveServer.url),
});

test.use({ replayHar: new URL('./mocked.har', import.meta.url) });

test('replays authenticated GET and POST requests on another host and blocks missing requests', async ({
    page,
    liveServer,
}) => {
    await page.goto('/ui/');
    const result = await page.evaluate(async () => {
        const token = JSON.parse(localStorage.getItem('persistedState')!).auth.sessionToken;
        const get = await fetch('/api/v2/items', { headers: { Authorization: `Bearer ${token}` } }).then((r) =>
            r.json()
        );
        const post = await fetch('/api/v2/items', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ query: 'items' }),
        }).then((r) => r.json());
        const missing = await fetch('/api/v2/missing').then(
            () => false,
            () => true
        );
        return { token, get, post, missing };
    });
    expect(result).toEqual({
        token: 'playwright-mock-session',
        get: { title: 'Mocked GET' },
        post: { title: 'Mocked POST' },
        missing: true,
    });
    expect(liveServer.apiRequests).toBe(0);
});
