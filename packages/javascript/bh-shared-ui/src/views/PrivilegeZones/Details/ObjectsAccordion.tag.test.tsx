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
import { setupServer } from 'msw/node';
import { beforeEach, describe, expect, it } from 'vitest';
import * as useAssetGroupTags from '../../../hooks/useAssetGroupTags';
import zoneHandlers from '../../../mocks/handlers/zoneHandlers';
import { render, screen, waitFor, within } from '../../../test-utils';
import { ObjectsAccordion } from './ObjectsAccordion';

const mockNavigate = vi.fn();

vi.mock('../../../hooks/useEnvironmentIdList', () => ({
    useEnvironmentIdList: () => ['env-1'],
}));

vi.mock('../../../hooks/useMeasure', () => ({
    useMeasure: () => [600, 600],
}));

vi.mock('../../../utils', async () => {
    const actual = await vi.importActual<any>('../../../utils');
    return {
        ...actual,
        useAppNavigate: () => mockNavigate,
    };
});

const server = setupServer(...zoneHandlers);

beforeEach(() => mockNavigate.mockClear());
beforeAll(() => server.listen());
afterEach(() => server.resetHandlers());
afterAll(() => server.close());

const useTagMembersInfiniteQuerySpy = vi.spyOn(useAssetGroupTags, 'useTagMembersInfiniteQuery');

describe('ObjectsAccordion', () => {
    const testKindCounts = {
        User: 4,
        Computer: 5,
        Group: 6,
    };

    it('renders', async () => {
        render(<ObjectsAccordion tagId={'42'} onObjectClick={vi.fn()} kindCounts={testKindCounts} totalCount={15} />);

        expect(screen.getByText('Objects')).toBeInTheDocument();
        expect(screen.getByText('Total Objects:')).toBeInTheDocument();
        expect(screen.getByText('15')).toBeInTheDocument();

        expect(screen.getByText('User')).toBeInTheDocument();
        expect(screen.getByText('4')).toBeInTheDocument();

        expect(screen.getByText('Computer')).toBeInTheDocument();
        expect(screen.getByText('5')).toBeInTheDocument();

        expect(screen.getByText('Group')).toBeInTheDocument();
        expect(screen.getByText('6')).toBeInTheDocument();

        // Accessing these child nodes at 0, 1, and 2 indices tests the render order of the elements
        // The elements should be rendered in alphabetical order even though the testKindCounts is not ordered that way
        const accordion = screen.getByTestId('privilege-zones_details_objects-accordion');
        const computerAccordionItem = accordion.childNodes[0];
        const groupAccordionItem = accordion.childNodes[1];
        const userAccordionItem = accordion.childNodes[2];

        expect(computerAccordionItem).not.toBeUndefined();
        expect(within(computerAccordionItem as HTMLElement).getByText('Computer')).toBeInTheDocument();

        expect(groupAccordionItem).not.toBeUndefined();
        expect(within(groupAccordionItem as HTMLElement).getByText('Group')).toBeInTheDocument();

        expect(userAccordionItem).not.toBeUndefined();
        expect(within(userAccordionItem as HTMLElement).getByText('User')).toBeInTheDocument();
    });

    it('toggles sort order when clicking sortable header', async () => {
        render(<ObjectsAccordion tagId={'42'} onObjectClick={vi.fn()} kindCounts={testKindCounts} totalCount={15} />);

        const accordionHeader = screen.getByTestId('privilege-zones_details_User-accordion-item');

        const sortButton = within(accordionHeader).getByTestId('sort-button');

        await userEvent.click(sortButton);

        expect(useTagMembersInfiniteQuerySpy).toBeCalledWith('42', 'desc', ['env-1'], 'User', false);

        await userEvent.click(sortButton);

        expect(useTagMembersInfiniteQuerySpy).toBeCalledWith('42', 'asc', ['env-1'], 'User', false);
    });

    it('navigates to object details when clicking an object row', async () => {
        render(<ObjectsAccordion tagId={'42'} onObjectClick={mockNavigate} kindCounts={{ User: 1 }} totalCount={1} />);

        const accordionOpenButton = screen.getByTestId('privilege-zones_details_User-accordion_open-toggle-button');

        await userEvent.click(accordionOpenButton);

        await waitFor(() => screen.getAllByRole('listitem'));

        await userEvent.click(screen.getByText('tag-41-object-1'));

        expect(mockNavigate).toHaveBeenCalled();
    });
});
