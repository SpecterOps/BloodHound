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
import type { Meta, StoryObj } from '@storybook/react';
import { Button } from 'doodle-ui';
import { useId, useRef, useState } from 'react';
import { SkipLink } from './SkipLink';

function SkipLinkExample() {
    const exampleId = useId();
    const mainId = `skip-link-main-${exampleId}`;
    const startButton = useRef<HTMLButtonElement>(null);
    const [focusLocation, setFocusLocation] = useState('Ready to start');

    return (
        <div className='mx-auto max-w-2xl space-y-4 p-6 pt-20'>
            <header className='space-y-4'>
                <h1 className='text-xl font-bold'>Skip link</h1>
                <p>Click Start demo, then press Tab → Enter to skip navigation.</p>
                <p role='status' className='rounded border border-neutral-3 bg-neutral-2 p-3'>
                    <strong>Focus: </strong>
                    {focusLocation}
                </p>
                <Button
                    ref={startButton}
                    onFocus={() => setFocusLocation('Start button — press Tab to reveal the skip link')}
                    onClick={(event) => event.currentTarget.focus()}>
                    Start demo
                </Button>
            </header>

            <SkipLink
                href={`#${mainId}`}
                target='_self'
                onFocus={() => setFocusLocation('Skip link — press Enter to bypass navigation')}>
                Skip to main content
            </SkipLink>

            <nav aria-label='Example navigation' className='rounded border border-neutral-3 bg-neutral-2 p-4'>
                <div className='flex flex-wrap gap-6'>
                    {['Overview', 'Reports', 'Settings'].map((label) => (
                        <a
                            key={label}
                            href={`#${mainId}-${label.toLowerCase()}`}
                            target='_self'
                            className='rounded text-primary underline focus:focus-ring'
                            onFocus={() => setFocusLocation(`Navigation: ${label}`)}>
                            {label}
                        </a>
                    ))}
                </div>
            </nav>

            <main
                id={mainId}
                tabIndex={-1}
                onFocus={(event) => {
                    if (event.target === event.currentTarget) setFocusLocation('Main content — navigation bypassed');
                }}
                className='space-y-4 rounded border border-neutral-3 p-6 focus:focus-ring'>
                <h2 className='text-lg font-bold'>Main content</h2>
                <Button
                    variant='secondary'
                    onFocus={() => setFocusLocation('First button in main content')}
                    onClick={() => startButton.current?.focus()}>
                    Restart demo
                </Button>
                <div className='flex flex-wrap gap-6 text-sm'>
                    {['Overview', 'Reports', 'Settings'].map((label) => (
                        <p key={label} id={`${mainId}-${label.toLowerCase()}`}>
                            {label}
                        </p>
                    ))}
                </div>
            </main>
        </div>
    );
}
const meta: Meta<typeof SkipLink> = {
    title: 'Components/Navigation/SkipLink',
    component: SkipLink,
    tags: ['autodocs'],
    parameters: {
        layout: 'fullscreen',
        docs: {
            // Isolate keyboard focus, fixed positioning, and hash navigation from the Docs page.
            story: { inline: false, height: '560px' },
            description: {
                component: 'Hidden until focused. Press Tab to reveal the link, then Enter to skip navigation.',
            },
        },
    },
    render: () => <SkipLinkExample />,
};

export default meta;
type Story = StoryObj<typeof meta>;

export const KeyboardNavigation: Story = {};
