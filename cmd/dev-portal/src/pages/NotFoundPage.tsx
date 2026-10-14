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

import { Link } from 'react-router-dom';
import { PortalHeader } from '../components/PortalHeader';

// An edge that leads to an empty node: the path the reader followed goes nowhere
const DanglingEdge = () => (
    <svg viewBox='0 0 96 24' aria-hidden='true' className='h-6 w-24'>
        <circle cx='8' cy='12' r='6' fill='var(--dp-color-marker)' />
        <path d='M14 12 H72' stroke='var(--dp-color-border-strong)' strokeWidth='2' />
        <circle
            cx='84'
            cy='12'
            r='9'
            fill='none'
            stroke='var(--dp-color-border-strong)'
            strokeWidth='2'
            strokeDasharray='4 3'
        />
    </svg>
);

export const NotFoundPage = () => (
    <div className='flex min-h-screen flex-col'>
        <PortalHeader />
        <main className='mx-auto w-full max-w-5xl flex-1 px-6 py-16 sm:py-24'>
            <DanglingEdge />
            <h1 className='mt-6 font-dp-display text-4xl font-extrabold leading-tight tracking-tight sm:text-5xl'>
                Page not found
            </h1>
            <p className='mt-4 text-lg text-dp-text-muted'>There is nothing at this address.</p>
            <Link
                to='/'
                className='mt-8 inline-block font-semibold text-dp-link underline underline-offset-2 hover:text-dp-link-hover'>
                Back to the main page
            </Link>
        </main>
    </div>
);
