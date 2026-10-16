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

import type { IconDefinition } from '@fortawesome/fontawesome-svg-core';

export type LinkCardConfig = {
    icon: IconDefinition;
    title: string;
    description: string;
    href: string;
};

// Doc Artifacts are static builds served beside the Dev Portal, so they open as full page loads rather than routes
export const isDocArtifactHref = (href: string) => href.startsWith('/artifacts/');

// The Link Card whose destination contains the given path, if any
export const linkCardForPath = (pathname: string) =>
    linkCards.find(({ href }) => {
        const destination = href.replace(/\/$/, '');
        return pathname === destination || pathname.startsWith(`${destination}/`);
    });

export const linkCards: LinkCardConfig[] = [];
