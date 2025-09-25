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
import { AssetGroupTagSelector, AssetGroupTagSelectorAutoCertifyAllMembers, SeedTypeCypher } from 'js-client-library';
import { rest } from 'msw';
import { setupServer } from 'msw/node';
import { useParams } from 'react-router-dom';
import SelectorForm from '.';
import { privilegeZonesKeys } from '../../../../hooks';
import { zoneHandlers } from '../../../../mocks';
import { act, render, screen, waitFor } from '../../../../test-utils';
import { apiClient, mockCodemirrorLayoutMethods, setUpQueryClient } from '../../../../utils';
import * as utils from '../utils';

const testSelector: AssetGroupTagSelector = {
    id: 777,
    asset_group_tag_id: 1,
    name: 'foo',
    allow_disable: true,
    description: 'bar',
    is_default: false,
    auto_certify: AssetGroupTagSelectorAutoCertifyAllMembers,
    created_at: '2024-10-05T17:54:32.245Z',
    created_by: 'Stephen64@gmail.com',
    updated_at: '2024-07-20T11:22:18.219Z',
    updated_by: 'Donna13@yahoo.com',
    disabled_at: '2024-09-15T09:55:04.177Z',
    disabled_by: 'Roberta_Morar72@hotmail.com',
    counts: {
        members: 3821,
    },
    seeds: [{ selector_id: 777, type: SeedTypeCypher, value: 'match(n) return n limit 5' }],
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

const handlers = [
    ...zoneHandlers,
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

const mockNavigate = vi.fn();
vi.mock('react-router-dom', async () => {
    const actual = await vi.importActual('react-router-dom');
    return {
        ...actual,
        useParams: vi.fn(),
        useNavigate: () => mockNavigate,
    };
});

const handleErrorSpy = vi.spyOn(utils, 'handleError');

mockCodemirrorLayoutMethods();

describe('Selector Form', () => {
    const user = userEvent.setup({ pointerEventsCheck: 0 });

    it('renders the form for creating a new selector', async () => {
        // Because there is no selector id path parameter in the url, the form is a create form
        // This means that none of the input fields should have any value aside from default values
        vi.mocked(useParams).mockReturnValue({ zoneId: '1', labelId: undefined });

        const mockState = [{ key: [privilegeZonesKeys.selectorDetail('1', '')], data: null }];

        const queryClient = setUpQueryClient(mockState);

        render(<SelectorForm />, { queryClient });

        expect(await screen.findByText('Defining Selector')).toBeInTheDocument();

        const nameInput = screen.getByLabelText('Name');
        expect(nameInput).toBeInTheDocument();
        expect(nameInput).toHaveValue('');

        const descriptionInput = screen.getByLabelText('Description');
        expect(descriptionInput).toBeInTheDocument();
        expect(descriptionInput).toHaveValue('');

        const autoCertifyDropdownDefault = await screen.findByTestId(
            'privilege-zones_save_selector-form_default-certify'
        );
        expect(autoCertifyDropdownDefault).toBeInTheDocument();

        await waitFor(() => {
            expect(autoCertifyDropdownDefault).toHaveTextContent('Off');
        });
        expect(screen.getByText('Selector Type')).toBeInTheDocument();

        // Object Selector component renders by default
        expect(screen.getByText('Object Selector')).toBeInTheDocument();
        // The delete button should not render when creating a new selector because it doesn't exist yet
        expect(screen.queryByRole('button', { name: /Delete Selector/ })).not.toBeInTheDocument();
        expect(screen.getByRole('button', { name: /Cancel/ })).toBeInTheDocument();
        // The save edits button should not render when creating a new selector
        expect(screen.queryByRole('button', { name: /Save Edits/ })).not.toBeInTheDocument();
        expect(screen.getByRole('button', { name: /Save/ })).toBeInTheDocument();

        expect(screen.getByText('Sample Results')).toBeInTheDocument();
    });

    it('renders the form for editing an existing selector', async () => {
        server.use(
            rest.get('/api/v2/asset-group-tags/:tagId/selectors/:selectorId', async (_, res, ctx) => {
                return res(
                    ctx.json({
                        data: { selector: testSelector },
                    })
                );
            })
        );
        // This url has the selector id of 777 in the path
        // and so this selector's data is filled into the form for the user to edit
        vi.mocked(useParams).mockReturnValue({ zoneId: '1', selectorId: '777' });

        const mockState = [{ key: [privilegeZonesKeys.selectorDetail('1', '777')], data: testSelector }];

        const queryClient = setUpQueryClient(mockState);

        render(<SelectorForm />, { queryClient });

        expect(await screen.findByText('Defining Selector')).toBeInTheDocument();

        await waitFor(
            async () => {
                const selectorStatusSwitch = await screen.findByLabelText('Selector Status');
                expect(selectorStatusSwitch).toBeInTheDocument();
                expect(selectorStatusSwitch).toHaveValue('');
                expect(screen.getByText('Disabled')).toBeInTheDocument();
            },
            { timeout: 5000 }
        );

        const nameInput = screen.getByLabelText('Name');
        expect(nameInput).toBeInTheDocument();
        await waitFor(
            () => {
                expect(nameInput).toHaveValue('foo');
            },
            { timeout: 5000 }
        );

        const descriptionInput = screen.getByLabelText('Description');
        expect(descriptionInput).toBeInTheDocument();
        await waitFor(() => {
            expect(descriptionInput).toHaveValue('bar');
        });

        const autoCertifyDropdownDefault = await screen.findByTestId(
            'privilege-zones_save_selector-form_default-certify'
        );
        expect(autoCertifyDropdownDefault).toBeInTheDocument();

        expect(autoCertifyDropdownDefault).toHaveTextContent('All members');
        expect(screen.getByText('Selector Type')).toBeInTheDocument();

        // Cypher Search renders because that is the seed type of the first seed of this selector
        await waitFor(() => {
            expect(screen.getByText('Cypher Search')).toBeInTheDocument();
        });

        await waitFor(() => {
            // The delete button should render because this selector exists and can be deleted
            expect(screen.getByRole('button', { name: /Delete Selector/ })).toBeInTheDocument();
            expect(screen.getByRole('button', { name: /Cancel/ })).toBeInTheDocument();
            expect(screen.getByRole('button', { name: /Save Edits/ })).toBeInTheDocument();
        });

        expect(screen.getByText('Sample Results')).toBeInTheDocument();
    });

    it('changes the text from "Disabled" to "Enabled" when the Selector Status switch is toggled', async () => {
        vi.mocked(useParams).mockReturnValue({ zoneId: '1', selectorId: '777' });
        server.use(
            rest.get('/api/v2/asset-group-tags/:tagId/selectors/:selectorId', async (_, res, ctx) => {
                return res(
                    ctx.json({
                        data: { selector: testSelector },
                    })
                );
            })
        );

        render(<SelectorForm />);

        expect(await screen.findByText('Defining Selector')).toBeInTheDocument();

        await waitFor(async () => {
            const selectorStatusSwitch = screen.getByLabelText('Selector Status');
            expect(selectorStatusSwitch).toBeInTheDocument();
            expect(selectorStatusSwitch).toHaveValue('');
            expect(screen.getByText('Disabled')).toBeInTheDocument();
            await user.click(selectorStatusSwitch);
            expect(screen.getByText('Enabled')).toBeInTheDocument();
        });
    });

    it('shows an error message when unable to delete a selector', async () => {
        console.error = vi.fn();
        vi.mocked(useParams).mockReturnValue({ zoneId: '1', selectorId: '777' });
        server.use(
            rest.get('/api/v2/asset-group-tags/:tagId/selectors/:selectorId', async (_, res, ctx) => {
                return res(
                    ctx.json({
                        data: { selector: testSelector },
                    })
                );
            })
        );
        render(<SelectorForm />);

        await waitFor(async () => {
            expect(await screen.findByRole('button', { name: /Delete Selector/ })).toBeInTheDocument();
        });

        await act(async () => {
            user.click(screen.getByRole('button', { name: /Delete Selector/ }));
        });

        await waitFor(async () => {
            expect(screen.getByText('Delete foo?')).toBeInTheDocument();
        });

        await act(async () => {
            await user.type(screen.getByTestId('confirmation-dialog_challenge-text'), 'delete this selector');
            await user.click(screen.getByRole('button', { name: /Confirm/ }));
        });

        await waitFor(() => {
            expect(handleErrorSpy).toBeCalled();
        });
    });

    test('clicking cancel on the form takes the user back to the details page the user was on previously', async () => {
        render(<SelectorForm />);

        await user.click(await screen.findByRole('button', { name: /Cancel/ }));

        await waitFor(() => {
            expect(mockNavigate).toBeCalledWith(-1);
        });
    });

    test('a name value is required to submit the form', async () => {
        vi.mocked(useParams).mockReturnValue({ zoneId: '1', selectorId: '' });
        render(<SelectorForm />);

        await waitFor(() => {
            expect(screen.getByRole('button', { name: /Save/ })).toBeInTheDocument();
        });

        await act(async () => {
            await user.click(screen.getByRole('button', { name: /Save/ }));
        });

        await waitFor(() => {
            expect(screen.getByText('Please provide a name for the Selector')).toBeInTheDocument();
        });
    });

    test('filling in the name value allows updating the selector and navigates back to the details page', async () => {
        vi.mocked(useParams).mockReturnValue({ zoneId: '1', selectorId: '777' });

        render(<SelectorForm />);

        const nameInput = await screen.findByLabelText('Name');

        await user.click(nameInput);
        await user.paste('foo');

        await waitFor(async () => {
            expect(screen.getByRole('button', { name: /Save Edits/ })).toBeInTheDocument();
            await user.click(screen.getByRole('button', { name: /Save Edits/ }));
        });

        expect(screen.queryByText('Please provide a name for the Selector')).not.toBeInTheDocument();

        await waitFor(() => {
            expect(mockNavigate).toBeCalled();
        });
    });

    it('handles creating a new selector', async () => {
        vi.mocked(useParams).mockReturnValue({ zoneId: '1', selectorId: undefined });
        // Because there is no selector id path parameter in the url, the form is a create form
        // This means that none of the input fields should have any value aside from default values
        render(<SelectorForm />);

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

        await waitFor(() => {
            expect(createSelectorSpy).toBeCalled();
        });
    });

    it('shows a warning for using labels associated with tags in zone forms', async () => {
        vi.mocked(useParams).mockReturnValue({ zoneId: '1', labelId: undefined });
        render(<SelectorForm />);

        const seedTypeSelect = await screen.findByLabelText('Selector Type');

        await user.click(seedTypeSelect);

        const cypherOption = await screen.findByRole('option', { name: /Cypher/ });

        await act(async () => {
            await user.click(cypherOption);
        });

        const textBoxes = screen.getAllByRole('textbox');
        const cypherTextBox = textBoxes.find((box) => box.className === 'flex-1');

        await user.click(cypherTextBox!);

        await act(async () => {
            await user.paste('match(n:Tag_foo) return n');
        });

        await waitFor(() => {
            expect(
                screen.getByText(
                    'Privilege Zone labels should only be used in cypher within the Explore page. Utilizing Privilege Zone labels in a cypher based Selector seed may result in incomplete data.'
                )
            ).toBeInTheDocument();
        });
    });

    it('does not show a warning for using labels associated with tags in label forms', async () => {
        vi.mocked(useParams).mockReturnValue({ zoneId: '', labelId: '1' });
        render(<SelectorForm />);

        const seedTypeSelect = await screen.findByLabelText('Selector Type');

        await user.click(seedTypeSelect);

        const cypherOption = await screen.findByRole('option', { name: /Cypher/ });

        await act(async () => {
            await user.click(cypherOption);
        });

        const textBoxes = screen.getAllByRole('textbox');
        const cypherTextBox = textBoxes.find((box) => box.className === 'flex-1');

        await user.click(cypherTextBox!);

        await act(async () => {
            await user.paste('match(n:Tag_foo) return n');
        });

        expect(
            screen.queryByText(
                'Privilege Zone labels should only be used in cypher within the Explore page. Utilizing Privilege Zone labels in a cypher based Selector seed may result in incomplete data.'
            )
        ).not.toBeInTheDocument();
    });
});
