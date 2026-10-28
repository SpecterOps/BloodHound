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
//

import type { ReactNode } from 'react';
import type { CardAccent } from '../config/portalConfig';

type PortalCardProps = {
    accent: CardAccent;
    description: string;
    href: string;
    icon: ReactNode;
    index: string;
    title: string;
};

export function PortalCard({ accent, description, href, icon, index, title }: PortalCardProps) {
    return (
        <a className='portal-card group' data-accent={accent} href={href} aria-label={`Open ${title}`}>
            <div className='flex items-start justify-between gap-6'>
                <div className='card-icon' aria-hidden='true'>
                    {icon}
                </div>
                <span className='card-index'>{index}</span>
            </div>

            <div className='mt-14'>
                <div className='card-title' role='heading' aria-level={2}>
                    {title}
                </div>
                <p className='card-description'>{description}</p>
            </div>

            <div className='card-action' aria-hidden='true'>
                <span>Open</span>
                <svg viewBox='0 0 24 24' fill='none'>
                    <path d='M5 12h13M13 7l5 5-5 5' />
                </svg>
            </div>
        </a>
    );
}
