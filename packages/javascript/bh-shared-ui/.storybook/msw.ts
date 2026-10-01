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
import type { Loader } from '@storybook/react';
import { setupWorker, type RestHandler } from 'msw';

export type StoryMocks = { handlers: () => RestHandler[] };

const worker = setupWorker();
let started: ReturnType<typeof worker.start> | undefined;

// Factories give every story load fresh mutable fixtures (including after a play-function rerun).
export const mswLoader: Loader = async ({ parameters }) => {
    started ??= worker.start({
        serviceWorker: { url: './mockServiceWorker.js' },
        quiet: true,
        onUnhandledRequest(request, print) {
            if (new URL(request.url).pathname.startsWith('/api/')) print.error();
        },
    });
    await started;
    const mocks = parameters.msw as StoryMocks | undefined;
    worker.resetHandlers(...(mocks?.handlers() ?? []));
    return {};
};
