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

import { Link, Route, Routes } from 'react-router-dom';
import { DocumentationIcon, StorybookIcon } from './components/Icons';
import { PortalCard } from './components/PortalCard';
import { NotFoundPage } from './pages/NotFoundPage';
import { StorybookPage } from './pages/StorybookPage';

export default function App() {
    return (
        <Routes>
            <Route path='/' element={<LandingPage />} />
            <Route path='/storybook' element={<StorybookPage />} />
            <Route path='*' element={<NotFoundPage />} />
        </Routes>
    );
}

function LandingPage() {
    return (
        <main className='site-shell'>
            <div className='ambient-grid' aria-hidden='true' />

            <div className='content-wrap'>
                <header className='site-header'>
                    <Link className='brand' to='/' aria-label='BloodHound Developer Portal home'>
                        <span className='brand-mark' aria-hidden='true'>
                            BH
                        </span>
                        <span>BloodHound</span>
                        <span className='brand-divider' aria-hidden='true' />
                        <span className='brand-context'>Developer Portal</span>
                    </Link>
                    <span className='header-label'>Internal resources</span>
                </header>

                <section className='hero' aria-labelledby='portal-title'>
                    <div className='eyebrow'>
                        <span className='eyebrow-dot' />
                        Engineering enablement
                    </div>
                    <div className='hero-title' id='portal-title' role='heading' aria-level={1}>
                        BloodHound
                        <br />
                        Developer Portal
                    </div>
                    <p className='hero-copy'>
                        A central home for the tools, standards, and knowledge that help us build better software.
                    </p>
                </section>

                <section className='portal-grid' aria-label='Developer portal destinations'>
                    <PortalCard
                        accent='blue'
                        index='01'
                        title='Doodle UI Storybook'
                        description='Explore our shared components, patterns, and design foundations.'
                        href='#/storybook?path=%2Fdocs%2Fabout--docs'
                        icon={<StorybookIcon />}
                    />
                    <PortalCard
                        accent='orange'
                        index='02'
                        title='Documentation'
                        description='Find technical guides, standards, and product knowledge.'
                        href='#documentation'
                        icon={<DocumentationIcon />}
                    />
                </section>

                <footer>
                    <span>© 2026 Specter Ops, Inc.</span>
                </footer>
            </div>
        </main>
    );
}
