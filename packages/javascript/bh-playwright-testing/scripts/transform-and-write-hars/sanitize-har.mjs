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

function extractSessionTokens(storageState) {
    return storageState.origins.flatMap(({ localStorage }) =>
        localStorage.flatMap(({ name, value }) => {
            if (name !== 'persistedState') {
                return [];
            }

            const token = JSON.parse(value).auth?.sessionToken;
            return token ? [token] : [];
        })
    );
}

function stripCredentials(entry) {
    const sanitizedEntry = structuredClone(entry);
    const requestHeaders = (entry.request.headers ?? []).map((header) =>
        /^authorization$/i.test(header.name) ? { ...header, value: 'Bearer playwright-mock-session' } : header
    );
    const sanitizedRequestHeaders = requestHeaders.filter(({ name }) => !/^(cookie|set-cookie)$/i.test(name));
    const sanitizedResponseHeaders = (entry.response.headers ?? []).filter(
        ({ name }) => !/^(cookie|set-cookie|authorization)$/i.test(name)
    );

    sanitizedEntry.request.headers = sanitizedRequestHeaders;
    sanitizedEntry.request.cookies = [];
    sanitizedEntry.response.headers = sanitizedResponseHeaders;
    sanitizedEntry.response.cookies = [];

    return sanitizedEntry;
}

function replaceSessionTokens(source, sessionTokens) {
    let sanitizedSource = source;

    for (const token of sessionTokens) {
        sanitizedSource = sanitizedSource.replaceAll(token, 'playwright-mock-session');
    }

    return sanitizedSource;
}

export function sanitizeHar(har, storageState) {
    const sanitizedEntries = har.log.entries.map(stripCredentials);
    const sanitizedHar = { ...har, log: { ...har.log, entries: sanitizedEntries } };
    const sessionTokens = extractSessionTokens(storageState);
    const harSource = JSON.stringify(sanitizedHar);
    const sanitizedSource = replaceSessionTokens(harSource, sessionTokens);
    const sanitizedDocument = JSON.parse(sanitizedSource);

    return sanitizedDocument;
}
