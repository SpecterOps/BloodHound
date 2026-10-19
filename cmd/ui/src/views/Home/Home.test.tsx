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

import { render, screen } from '@testing-library/react';
import { StrictMode } from 'react';
import { MemoryRouter, Route, Routes, useLocation } from 'react-router-dom';
import { LOGIN_REDIRECT_STORAGE_KEY } from 'src/routes/loginRedirect';
import Home from './Home';

const Destination = (): JSX.Element => {
    const { pathname, search, hash } = useLocation();
    return <div>{pathname + search + hash}</div>;
};

describe('Home', () => {
    afterEach(() => sessionStorage.clear());

    it('retains the saved destination when effects run twice in StrictMode', async () => {
        const destination = '/explore?searchType=pathfinding#graph';
        sessionStorage.setItem(LOGIN_REDIRECT_STORAGE_KEY, destination);

        render(
            <StrictMode>
                <MemoryRouter initialEntries={['/']} future={{ v7_startTransition: true, v7_relativeSplatPath: true }}>
                    <Routes>
                        <Route path='/' element={<Home />} />
                        <Route path='/explore' element={<Destination />} />
                    </Routes>
                </MemoryRouter>
            </StrictMode>
        );

        expect(await screen.findByText(destination)).toBeInTheDocument();
        expect(sessionStorage.getItem(LOGIN_REDIRECT_STORAGE_KEY)).toBeNull();
    });
});
