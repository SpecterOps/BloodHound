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

function respond(request, response, state) {
    if (request.url.startsWith('/api/')) {
        state.liveApiRequests++;
        if (!state.recordApiResponses) return response.writeHead(500).end('Backend must not be used');
        response.setHeader('Content-Type', 'application/json');
        return response.end(JSON.stringify({ title: request.method === 'POST' ? 'Mocked POST' : 'Mocked GET' }));
    }
    response.setHeader('Content-Type', 'text/html');
    response.end(
        `<button onclick="fetch('/api/v2/items', { method: 'POST', headers: {'Content-Type':'application/json'}, body: JSON.stringify({query:'items'}) }).then(r=>r.json()).then(d=>document.querySelector('h1').textContent=d.title)">Load</button><h1>Loading</h1><script>fetch('/api/v2/items', {headers: {Authorization: 'Bearer ' + JSON.parse(localStorage.getItem('persistedState')).auth.sessionToken}}).then(r=>r.json()).then(d=>document.querySelector('h1').textContent=d.title)</script>`
    );
}

export async function startTestServer() {
    const state = { liveApiRequests: 0, recordApiResponses: false };
    const server = createServer((request, response) => respond(request, response, state));
    await new Promise((resolve) => server.listen(0, '127.0.0.1', resolve));
    return {
        baseURL: `http://127.0.0.1:${server.address().port}`,
        state,
        close: async () => {
            server.closeAllConnections();
            await new Promise((resolve) => server.close(resolve));
        },
    };
}
