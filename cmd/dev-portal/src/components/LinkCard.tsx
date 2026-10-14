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

import { FontAwesomeIcon } from '@fortawesome/react-fontawesome';
import { Link } from 'react-router-dom';
import { isDocArtifactHref, type LinkCardConfig } from '../linkCards';

// Drawn as a node on the landing page's graph; the node and title react together on hover
const linkClassName = 'group flex items-center gap-4 rounded-dp-md py-1 sm:gap-5 sm:pr-4';

export const LinkCard = ({ icon, title, description, href }: LinkCardConfig) => {
    const card = (
        <>
            <span className='graph-node h-14 w-14 text-xl transition-colors group-hover:border-dp-marker group-hover:bg-dp-marker'>
                <FontAwesomeIcon icon={icon} aria-hidden='true' />
            </span>
            <span className='flex min-w-0 flex-col gap-0.5'>
                <span className='font-dp-display text-xl font-bold text-dp-hero-text underline-offset-4 group-hover:underline'>
                    {title}
                </span>
                <span className='text-base text-dp-hero-text-muted'>{description}</span>
            </span>
        </>
    );

    return isDocArtifactHref(href) ? (
        <a href={href} className={linkClassName}>
            {card}
        </a>
    ) : (
        <Link to={href} className={linkClassName}>
            {card}
        </Link>
    );
};
