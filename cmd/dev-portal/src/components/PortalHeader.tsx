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

import { faChevronDown } from '@fortawesome/free-solid-svg-icons';
import { FontAwesomeIcon } from '@fortawesome/react-fontawesome';
import { Menu, MenuContent, MenuItem, MenuTrigger } from 'doodle-ui';
import { Link, useLocation, useNavigate } from 'react-router-dom';
import { isDocArtifactHref, linkCardForPath, linkCards, type LinkCardConfig } from '../linkCards';

type PortalHeaderProps = {
    // 'hero' sits on the landing page's graph canvas; 'surface' sits above reading pages
    tone?: 'hero' | 'surface';
};

const toneClassNames: Record<NonNullable<PortalHeaderProps['tone']>, string> = {
    hero: 'bg-dp-hero-canvas text-dp-hero-text',
    surface: 'border-b border-dp-border bg-dp-surface text-dp-text',
};

// Three linked nodes: a small graph standing in for a logo
const BrandMark = () => (
    <svg viewBox='0 0 24 24' aria-hidden='true' className='h-6 w-6 shrink-0'>
        <path d='M6 7 L18 7 M6 7 L12 18' stroke='var(--dp-color-hero-edge)' strokeWidth='2' fill='none' />
        <circle cx='6' cy='7' r='3.5' fill='var(--dp-color-marker)' />
        <circle cx='18' cy='7' r='3' fill='var(--dp-color-hero-edge)' />
        <circle cx='12' cy='18' r='3' fill='var(--dp-color-hero-edge)' />
    </svg>
);

// Switches between the Link Card destinations without returning to the landing page
const SectionMenu = ({ currentSection }: { currentSection: LinkCardConfig }) => {
    const navigate = useNavigate();
    const openSection = (href: string) => (isDocArtifactHref(href) ? window.location.assign(href) : navigate(href));

    return (
        <Menu>
            <MenuTrigger className='group flex items-center gap-2 rounded-dp-sm px-2 py-1 font-dp-display text-sm font-semibold text-dp-text-muted hover:text-dp-text data-[state=open]:text-dp-text'>
                {currentSection.title}
                <FontAwesomeIcon
                    icon={faChevronDown}
                    aria-hidden='true'
                    className='text-xs motion-safe:transition-transform group-data-[state=open]:rotate-180'
                />
            </MenuTrigger>
            <MenuContent
                align='start'
                className='min-w-64 border border-dp-border bg-dp-surface text-dp-text dark:border-dp-border dark:bg-dp-surface'>
                {linkCards.map((linkCard) => {
                    const isCurrent = linkCard.href === currentSection.href;
                    return (
                        <MenuItem
                            key={linkCard.href}
                            iconLeft
                            icon={<FontAwesomeIcon icon={linkCard.icon} fixedWidth aria-hidden='true' />}
                            aria-current={isCurrent ? 'page' : undefined}
                            onSelect={() => openSection(linkCard.href)}
                            className='gap-1 px-3 py-2.5'>
                            <span className='flex items-center justify-between gap-4'>
                                <span className='flex flex-col gap-1'>
                                    <span className='font-dp-display font-bold'>{linkCard.title}</span>
                                    <span className='text-xs opacity-80'>{linkCard.description}</span>
                                </span>
                                {isCurrent && (
                                    <span aria-hidden='true' className='h-2 w-2 shrink-0 rounded-full bg-dp-marker' />
                                )}
                            </span>
                        </MenuItem>
                    );
                })}
            </MenuContent>
        </Menu>
    );
};

export const PortalHeader = ({ tone = 'surface' }: PortalHeaderProps) => {
    const currentSection = linkCardForPath(useLocation().pathname);

    return (
        <header className={`flex h-14 shrink-0 items-center gap-4 px-6 ${toneClassNames[tone]}`}>
            <Link
                to='/'
                className='flex items-center gap-2.5 rounded-dp-sm font-dp-display text-lg font-extrabold tracking-tight hover:opacity-80'>
                <BrandMark />
                Dev Portal
            </Link>
            {currentSection && (
                <>
                    <span aria-hidden='true' className='h-5 border-l border-dp-border-strong' />
                    <SectionMenu currentSection={currentSection} />
                </>
            )}
        </header>
    );
};
