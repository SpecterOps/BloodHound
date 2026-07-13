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

import type { Page } from '@playwright/test';

type StubAuthToken = {
    id: string;
    name: string;
    user_id: string;
    hmac_method: string;
    created_at: string;
    updated_at: string;
    last_access: string;
    deleted_at: { Time: string; Valid: boolean };
    expires_at: { Time: string; Valid: boolean } | null;
};

const DEFAULT_TOKENS: StubAuthToken[] = [
    {
        id: 'playwright-token-id',
        name: 'Playwright Token',
        user_id: 'playwright-user-id',
        hmac_method: 'hmac-sha2-256',
        created_at: '2026-01-01T00:00:00Z',
        updated_at: '2026-01-01T00:00:00Z',
        last_access: '2026-01-02T00:00:00Z',
        deleted_at: { Time: '0001-01-01T00:00:00Z', Valid: false },
        expires_at: null,
    },
];

export type UserTokensStubOptions = {
    tokens?: StubAuthToken[];
};

/**
 * Stubs the user tokens list so the API Key Management dialog renders at least one token row
 * without touching the real user's tokens. Install before opening the dialog so the "Revoke"
 * action is available. Non-GET traffic falls through to any lower-priority route handlers.
 * Pass an empty `tokens` array to render the "No tokens available" empty state.
 */
export async function installUserTokensStub(page: Page, opts: UserTokensStubOptions = {}): Promise<void> {
    const tokens = opts.tokens ?? DEFAULT_TOKENS;

    await page.route(/\/api\/v2\/tokens(\?|$)/, async (route) => {
        if (route.request().method() !== 'GET') {
            return route.fallback();
        }

        return route.fulfill({
            json: {
                data: {
                    tokens,
                },
            },
        });
    });
}

type StubNewAuthToken = StubAuthToken & {
    key: string;
};

const DEFAULT_NEW_TOKEN: StubNewAuthToken = {
    id: 'playwright-token-id',
    name: 'Playwright Token',
    user_id: 'playwright-user-id',
    hmac_method: 'hmac-sha2-256',
    created_at: '2026-01-01T00:00:00Z',
    updated_at: '2026-01-01T00:00:00Z',
    last_access: '2026-01-02T00:00:00Z',
    deleted_at: { Time: '0001-01-01T00:00:00Z', Valid: false },
    expires_at: null,
    key: 'playwright-token-key',
};

/**
 * Stubs the delete token endpoint so the API Key Management dialog can complete the revoke token
 * flow without deleting a real token. Install before confirming the revoke action. Non-DELETE
 * traffic falls through to any lower-priority route handlers (such as installUserTokensStub).
 */
export async function installDeleteUserTokenStub(page: Page): Promise<void> {
    await page.route(/\/api\/v2\/tokens\/[^/?]+(\?|$)/, async (route) => {
        if (route.request().method() !== 'DELETE') {
            return route.fallback();
        }

        return route.fulfill({ status: 200 });
    });
}

export type CreateUserTokenStubOptions = {
    token?: StubNewAuthToken;
};

/**
 * Stubs the create token endpoint so the API Key Management dialog can complete the create token
 * flow without minting a real token. Install before submitting the create token form. Non-POST
 * traffic falls through to any lower-priority route handlers (such as installUserTokensStub).
 */
export async function installCreateUserTokenStub(page: Page, opts: CreateUserTokenStubOptions = {}): Promise<void> {
    const token = opts.token ?? DEFAULT_NEW_TOKEN;

    await page.route(/\/api\/v2\/tokens(\?|$)/, async (route) => {
        if (route.request().method() !== 'POST') {
            return route.fallback();
        }

        return route.fulfill({
            json: {
                data: token,
            },
        });
    });
}
