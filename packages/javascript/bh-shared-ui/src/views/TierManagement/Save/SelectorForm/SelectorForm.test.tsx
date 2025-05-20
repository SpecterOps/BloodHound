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
import { createMemoryHistory } from 'history';
import { SeedTypeCypher, SeedTypeObjectId } from 'js-client-library';
import { rest } from 'msw';
import { setupServer } from 'msw/node';
import { Route, Routes } from 'react-router-dom';
import SelectorForm from '.';
import { tierHandlers } from '../../../../mocks';
import { act, longWait, render, screen, waitFor } from '../../../../test-utils';
import { apiClient, mockCodemirrorLayoutMethods } from '../../../../utils';

const testSelector = {
    id: 777,
    asset_group_tag_id: 1,
    name: 'foo',
    allow_disable: true,
    description: 'bar',
    is_default: false,
    auto_certify: true,
    created_at: '2024-10-05T17:54:32.245Z',
    created_by: 'Stephen64@gmail.com',
    updated_at: '2024-07-20T11:22:18.219Z',
    updated_by: 'Donna13@yahoo.com',
    disabled_at: '2024-09-15T09:55:04.177Z',
    disabled_by: 'Roberta_Morar72@hotmail.com',
    count: 3821,
    seeds: [{ selector_id: 777, type: SeedTypeCypher, value: 'match(n) return n limit 5' }],
};

const testObjectIdSelector = {
    ...testSelector,
    seeds: [
        { selector_id: 777, type: SeedTypeObjectId, value: '1' },
        { selector_id: 777, type: SeedTypeObjectId, value: '2' },
        { selector_id: 777, type: SeedTypeObjectId, value: '3' },
    ],
};

const testNodes = [
    {
        name: 'bar',
        objectid: '777',
        type: 'Bat',
    },
];
const testSearchResults = {
    data: testNodes,
};

beforeAll(() => server.listen());
afterEach(() => server.resetHandlers());
afterAll(() => server.close());

const handlers = [
    ...tierHandlers,
    rest.get('/api/v2/asset-group-tags/:tagId/selectors/777', async (_, res, ctx) => {
        return res(
            ctx.json({
                data: testSelector,
            })
        );
    }),
    rest.post(`/api/v2/asset-group-tags/preview-selectors`, (_, res, ctx) => {
        return res(ctx.json({ data: { members: [] } }));
    }),
    rest.post(`/api/v2/graphs/cypher`, (_, res, ctx) => {
        return res(ctx.json({ data: { nodes: {}, edges: [] } }));
    }),
    rest.get(`/api/v2/search`, (_, res, ctx) => {
        return res(ctx.json(testSearchResults));
    }),
];

const server = setupServer(...handlers);

beforeAll(() => server.listen());
afterEach(() => server.resetHandlers());
afterAll(() => server.close());

mockCodemirrorLayoutMethods();

describe('Selector Form', () => {
    const user = userEvent.setup({ pointerEventsCheck: 0 });
    const detailsPath = '/tier-management/details/tier/1/selector/777';
    const createNewPath = '/tier-management/save/tier/1/selector';
    const editExistingPath = '/tier-management/save/tier/1/selector/777';

    it('renders the form for creating a new selector', async () => {
        // Because there is no selector id path parameter in the url, the form is a create form
        // This means that none of the input fields should have any value aside from default values
        const history = createMemoryHistory({ initialEntries: [createNewPath] });

        await act(async () => {
            render(<SelectorForm />, { history });
        });

        expect(await screen.findByText('Defining Selector')).toBeInTheDocument();

        const nameInput = screen.getByLabelText('Name');
        expect(nameInput).toBeInTheDocument();
        expect(nameInput).toHaveValue('');

        const descriptionInput = screen.getByLabelText('Description');
        expect(descriptionInput).toBeInTheDocument();
        expect(descriptionInput).toHaveValue('');

        expect(screen.getByText('Selector Type')).toBeInTheDocument();

        // Object Selector component renders by default
        expect(screen.getByText('Object Selector')).toBeInTheDocument();
        // The delete button should not render when creating a new selector because it doesn't exist yet
        expect(screen.queryByRole('button', { name: /Delete Selector/ })).not.toBeInTheDocument();
        expect(screen.getByRole('button', { name: /Cancel/ })).toBeInTheDocument();
        expect(screen.getByRole('button', { name: /Save/ })).toBeInTheDocument();

        expect(screen.getByText('Sample Results')).toBeInTheDocument();
    });

    it('renders the form for editing an existing selector', async () => {
        // This url has the selector id of 777 in the path
        // and so this selector's data is filled into the form for the user to edit
        const history = createMemoryHistory({ initialEntries: [editExistingPath] });

        await act(async () => {
            render(<SelectorForm />, { history });
        });

        expect(await screen.findByText('Defining Selector')).toBeInTheDocument();

        longWait(async () => {
            const selectorStatusSwitch = await screen.findByLabelText('Selector Status');
            expect(selectorStatusSwitch).toBeInTheDocument();
            expect(selectorStatusSwitch).toHaveValue('on');
            expect(screen.getByText('Enabled')).toBeInTheDocument();
        });

        const nameInput = screen.getByLabelText('Name');
        expect(nameInput).toBeInTheDocument();
        longWait(() => {
            expect(nameInput).toHaveValue('foo');
        });

        const descriptionInput = screen.getByLabelText('Description');
        expect(descriptionInput).toBeInTheDocument();
        longWait(() => {
            expect(descriptionInput).toHaveValue('bar');
        });

        expect(screen.getByText('Selector Type')).toBeInTheDocument();

        // Cypher Search renders because that is the seed type of the first seed of this selector
        longWait(() => {
            expect(screen.getByText('Cypher Search')).toBeInTheDocument();
        });
        // The delete button should render because this selector exists and can be deleted
        longWait(() => {
            expect(screen.getByRole('button', { name: /Delete Selector/ })).toBeInTheDocument();
        });
        expect(screen.getByRole('button', { name: /Cancel/ })).toBeInTheDocument();
        expect(screen.getByRole('button', { name: /Save/ })).toBeInTheDocument();

        expect(screen.getByText('Sample Results')).toBeInTheDocument();
    });

    it('changes the text from "Enabled" to "Disabled" when the Selector Status switch is toggled', async () => {
        const history = createMemoryHistory({ initialEntries: [editExistingPath] });

        await act(async () => {
            render(<SelectorForm />, { history });
        });

        expect(await screen.findByText('Defining Selector')).toBeInTheDocument();

        longWait(async () => {
            const selectorStatusSwitch = screen.getByLabelText('Selector Status');
            expect(selectorStatusSwitch).toBeInTheDocument();
            expect(selectorStatusSwitch).toHaveValue('on');
            expect(screen.getByText('Enabled')).toBeInTheDocument();
            await user.click(selectorStatusSwitch);
            expect(screen.getByText('Disabled')).toBeInTheDocument();
        });
    });

    it('shows an error message when unable to delete a selector', async () => {
        const history = createMemoryHistory({ initialEntries: ['/tier-management/save/tier/1/selector/777'] });

        console.error = vi.fn();

        render(
            <Routes>
                <Route path={'/'} element={<SelectorForm />} />
                <Route path={'/tier-management/save/tier/:tierId/selector/:selectorId'} element={<SelectorForm />} />
            </Routes>,
            { history }
        );

        longWait(async () => {
            expect(await screen.findByRole('button', { name: /Delete Selector/ })).toBeInTheDocument();
            await act(async () => {
                user.click(screen.getByRole('button', { name: /Delete Selector/ }));
            });
        });

        longWait(async () => {
            expect(screen.getByText('Delete foo?')).toBeInTheDocument();

            await user.type(screen.getByTestId('confirmation-dialog_challenge-text'), 'delete this selector');
            await user.click(screen.getByRole('button', { name: /Confirm/ }));

            expect(await screen.findByText('get rekt')).toBeInTheDocument();
        });
    });

    test('clicking cancel on the form takes the user back to the details page the user was on previously', async () => {
        const initialEntries = [detailsPath, editExistingPath];
        const history = createMemoryHistory({
            initialEntries: initialEntries,
        });

        render(<SelectorForm />, { history });

        await user.click(screen.getByRole('button', { name: /Cancel/ }));

        longWait(() => {
            expect(history.location.pathname).toBe(detailsPath);
        });
    });

    test('a name value is required to submit the form', async () => {
        const history = createMemoryHistory({
            initialEntries: [createNewPath],
        });

        render(<SelectorForm />, { history });

        await user.click(screen.getByRole('button', { name: /Save/ }));

        longWait(() => {
            expect(screen.getByText('Please provide a name for the selector')).toBeInTheDocument();
        });
    });

    test('filling in the name value allows updating the selector and navigates back to the details page', async () => {
        const history = createMemoryHistory({
            initialEntries: [detailsPath, editExistingPath],
        });

        render(
            <Routes>
                <Route path={'/'} element={<SelectorForm />} />;
                <Route path={'/tier-management/save/tier/:tierId/selector/:selectorId'} element={<SelectorForm />} />;
            </Routes>,
            { history }
        );

        const nameInput = await screen.findByLabelText('Name');

        await user.click(nameInput);
        await user.paste('foo');

        longWait(async () => {
            expect(screen.getByRole('button', { name: /Save/ })).toBeInTheDocument();
            await user.click(screen.getByRole('button', { name: /Save/ }));
        });

        expect(screen.queryByText('Please provide a name for the selector')).not.toBeInTheDocument();

        longWait(() => {
            expect(history.location.pathname).toBe(detailsPath);
        });
    });

    it('handles creating a new selector', async () => {
        // Because there is no selector id path parameter in the url, the form is a create form
        // This means that none of the input fields should have any value aside from default values
        const history = createMemoryHistory({ initialEntries: [createNewPath] });

        await act(async () => {
            render(<SelectorForm />, { history });
        });

        const nameInput = await screen.findByLabelText('Name');

        await user.click(nameInput);
        await user.paste('foo');

        const createSelectorSpy = vi.spyOn(apiClient, 'createAssetGroupTagSelector');

        const input = screen.getByLabelText('Search Objects To Add');

        await user.click(input);
        await user.paste('bar');

        const options = await screen.findAllByRole('option');

        await user.click(
            options.find((option) => {
                return option.innerText === 'bar';
            })!
        );

        await user.click(await screen.findByRole('button', { name: /Save/ }));

        waitFor(() => {
            expect(createSelectorSpy).toBeCalled();
        });
    });

    test('the object selector list is prepopluated correctly according to the selector seeds', async () => {
        const history = createMemoryHistory({
            initialEntries: [detailsPath, editExistingPath],
        });

        await act(async () => {
            render(<SelectorForm />, { history });
        });

        testObjectIdSelector;
    });
});
