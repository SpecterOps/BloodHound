// Copyright 2025 Specter Ops, Inc.
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

import { screen, waitFor } from '@testing-library/react';
import { render } from '../test-utils';
import PageWithTitle, { AppNameProvider } from './PageWithTitle';

describe('PageWithTitle', () => {
    it('sets the document title using the default app name', async () => {
        render(<PageWithTitle title='Bloodhound Page' />);
        await waitFor(() => expect(document.title).toBe('Bloodhound Page | BloodHound Enterprise'));
    });

    it('sets the document title using the app name from AppNameProvider', async () => {
        render(
            <AppNameProvider name='BloodHound Community Edition'>
                <PageWithTitle title='Bloodhound Page' />
            </AppNameProvider>
        );
        await waitFor(() => expect(document.title).toBe('Bloodhound Page | BloodHound Community Edition'));
    });

    it('does not set the document title when no title prop is provided', () => {
        document.title = 'Previous Title';
        render(<PageWithTitle />);
        expect(document.title).toBe('Previous Title');
    });

    it('renders the title as a visible h1 heading', () => {
        render(<PageWithTitle title='Bloodhound Page' />);
        expect(screen.getByRole('heading', { level: 1, name: 'Bloodhound Page' })).toBeInTheDocument();
    });
});
