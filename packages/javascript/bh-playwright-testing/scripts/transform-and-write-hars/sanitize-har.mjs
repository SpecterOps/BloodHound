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

function sessionTokens(storageState) {
    return storageState.origins.flatMap(({ localStorage }) =>
        localStorage.flatMap(({ name, value }) => {
            if (name !== 'persistedState') return [];
            const token = JSON.parse(value).auth?.sessionToken;
            return token ? [token] : [];
        })
    );
}

function stripCredentials(entry) {
    entry.request.headers = (entry.request.headers ?? []).map((header) =>
        /^authorization$/i.test(header.name) ? { ...header, value: 'Bearer playwright-mock-session' } : header
    );
    for (const message of [entry.request, entry.response]) {
        message.headers = (message.headers ?? []).filter(
            ({ name }) =>
                !/^(cookie|set-cookie)$/i.test(name) && (message === entry.request || !/^authorization$/i.test(name))
        );
        message.cookies = [];
    }
}

export function sanitizeHar(har, storageState) {
    const document = structuredClone(har);
    for (const entry of document.log.entries) stripCredentials(entry);
    let json = JSON.stringify(document);
    for (const secret of sessionTokens(storageState)) json = json.replaceAll(secret, 'playwright-mock-session');
    return JSON.parse(json);
}
