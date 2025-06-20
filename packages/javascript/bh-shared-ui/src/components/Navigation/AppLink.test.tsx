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

import userEvent from '@testing-library/user-event';
import { Route, Routes } from 'react-router-dom';
import { GloballySupportedSearchParams } from '../..';
import { render } from '../../test-utils';
import { AppLink } from './AppLink';

const TEST_ROUTE_HOME = '/home';
const TEST_ROUTE_RANDOM = '/random';
const TEST_ROUTE_SPLAT = `/admin/*`;
const TEST_ROUTE_SPLAT_CHILD_ROUTE = 'page';

const HOME_CONTENT = 'home';
const CHILD_TEST_ROUTE_CONTENT = 'testing_splat';
const MOCK_GLOBAL_PARAMS = GloballySupportedSearchParams.map((p, i) => `${p}=${i}`).join('&');
const LINK_TEXT = 'CLICK ME';

export const TestRoutes = ({ children }: React.PropsWithChildren) => {
    return (
        <Routes>
            <Route path={TEST_ROUTE_HOME} element={HOME_CONTENT} />
            <Route path={TEST_ROUTE_SPLAT}>
                <Route path={TEST_ROUTE_SPLAT_CHILD_ROUTE} element={CHILD_TEST_ROUTE_CONTENT} />
                <Route path='*' element={<AppLink to={'/admin/page'}>{LINK_TEXT}</AppLink>} />
            </Route>
            {children}
        </Routes>
    );
};

describe('AppLink', () => {
    it('links to the specified route with global params intact', async () => {
        const initialRoute = `${TEST_ROUTE_RANDOM}?${MOCK_GLOBAL_PARAMS}`;
        const screen = render(
            <TestRoutes>
                <Route path={TEST_ROUTE_RANDOM} element={<AppLink to={TEST_ROUTE_HOME}>{LINK_TEXT}</AppLink>} />
            </TestRoutes>,
            { route: initialRoute }
        );

        const user = userEvent.setup();
        await user.click(screen.getByText(LINK_TEXT));

        expect(screen.queryByText(HOME_CONTENT)).toBeInTheDocument();
        expect(window.location.search).toContain(MOCK_GLOBAL_PARAMS);
    });

    it('links to the specified route with global params intact when triggered by a splat route', async () => {
        const initialRoute = `/admin/*?${MOCK_GLOBAL_PARAMS}`;
        const screen = render(<TestRoutes />, { route: initialRoute });

        const user = userEvent.setup();
        await user.click(screen.getByText(LINK_TEXT));

        expect(screen.queryByText(CHILD_TEST_ROUTE_CONTENT)).toBeInTheDocument();
        expect(window.location.search).toContain(MOCK_GLOBAL_PARAMS);
    });

    it('uses the discardQueryParams prop to discard all query params on the link', async () => {
        const initialRoute = `${TEST_ROUTE_RANDOM}?${MOCK_GLOBAL_PARAMS}`;
        const screen = render(
            <TestRoutes>
                <Route
                    path={TEST_ROUTE_RANDOM}
                    element={
                        <AppLink to={TEST_ROUTE_HOME} discardQueryParams>
                            {LINK_TEXT}
                        </AppLink>
                    }
                />
            </TestRoutes>,
            { route: initialRoute }
        );

        const user = userEvent.setup();
        await user.click(screen.getByText(LINK_TEXT));

        expect(screen.queryByText(HOME_CONTENT)).toBeInTheDocument();
        expect(window.location.search).not.toContain(MOCK_GLOBAL_PARAMS);
    });

    it('composes search params from the search prop and use them in the link', async () => {
        const ADDED_SEARCH_PARAM = 'test=param';
        const initialRoute = `${TEST_ROUTE_RANDOM}?${MOCK_GLOBAL_PARAMS}`;
        const screen = render(
            <TestRoutes>
                <Route
                    path={TEST_ROUTE_RANDOM}
                    element={
                        <AppLink to={{ pathname: TEST_ROUTE_HOME, search: ADDED_SEARCH_PARAM }}>{LINK_TEXT}</AppLink>
                    }
                />
            </TestRoutes>,
            { route: initialRoute }
        );

        const user = userEvent.setup();
        await user.click(screen.getByText(LINK_TEXT));

        expect(screen.queryByText(HOME_CONTENT)).toBeInTheDocument();
        expect(window.location.search).toContain(MOCK_GLOBAL_PARAMS);
        expect(window.location.search).toContain(ADDED_SEARCH_PARAM);
    });
});
