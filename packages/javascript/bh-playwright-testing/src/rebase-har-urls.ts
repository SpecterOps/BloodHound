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

type HarDocument = {
    log: {
        entries: { request: { url: string } }[];
    };
};

export function makeRebasedHar<Document extends HarDocument>(har: Document, baseUrl: string): Document {
    const rebasedHar = structuredClone(har);

    for (const entry of rebasedHar.log.entries) {
        const recordedUrl = new URL(entry.request.url);
        const relativeUrl = recordedUrl.pathname + recordedUrl.search;
        const rebasedUrl = new URL(relativeUrl, baseUrl).href;

        entry.request.url = rebasedUrl;
    }

    return rebasedHar;
}
