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

import { LinkCard } from '../components/LinkCard';
import { PortalHeader } from '../components/PortalHeader';
import { linkCards } from '../linkCards';

export const LandingPage = () => (
    <div className='flex min-h-screen flex-col bg-dp-hero-canvas text-dp-hero-text [--dp-color-focus:var(--dp-color-hero-edge)]'>
        <PortalHeader tone='hero' />
        <main className='mx-auto w-full max-w-5xl flex-1 px-6 pb-16 pt-12 sm:pt-24'>
            <div className='flex gap-5'>
                {/* The source node every Link Card's edge starts from, and the edge down to the first branch */}
                <div aria-hidden='true' className='flex w-4 shrink-0 flex-col items-center pt-4 sm:pt-6'>
                    <span className='graph-source-node h-4 w-4' />
                    <span className='graph-lead w-0.5 flex-1 bg-dp-hero-edge' />
                </div>
                <div className='min-w-0 pb-8'>
                    <h1 className='font-dp-display text-5xl font-extrabold leading-none tracking-tight sm:text-7xl'>
                        BloodHound Developer Portal
                    </h1>
                    <p className='mt-4 max-w-[52ch] text-lg text-dp-hero-text-muted'>
                        Component stories and repository docs for building BloodHound.
                    </p>
                </div>
            </div>
            {/* Aligns the trunk with the source node's center */}
            <ul className='graph-tree graph-tree-animated ml-[calc(0.5rem-1px)]'>
                {linkCards.map((linkCard) => (
                    <li key={linkCard.href}>
                        <LinkCard {...linkCard} />
                    </li>
                ))}
            </ul>
        </main>
    </div>
);
