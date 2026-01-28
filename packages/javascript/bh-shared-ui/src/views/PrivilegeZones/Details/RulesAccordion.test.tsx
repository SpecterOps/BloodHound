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
import userEvent from '@testing-library/user-event';
import { CustomRulesKey, DefaultRulesKey, DisabledRulesKey, RulesKey } from 'js-client-library';
import { setupServer } from 'msw/node';
import * as useAssetGroupTags from '../../../hooks/useAssetGroupTags';
import { zoneHandlers } from '../../../mocks';
import { render, screen, waitFor, within } from '../../../test-utils';
import { RulesAccordion } from './RulesAccordion';

const mockNavigate = vi.fn();

vi.mock('../../../hooks/useSelectedTag', () => ({
    useSelectedTagPathParams: () => ({
        counts: {
            [RulesKey]: 6,
            [CustomRulesKey]: 2,
            [DefaultRulesKey]: 2,
            [DisabledRulesKey]: 2,
        },
    }),
}));

vi.mock('../../../hooks/useAssetGroupTags', async (importOriginal) => {
    const original: Record<string, any> = await importOriginal();

    return {
        ...original,
        useRuleInfo: () => ({
            data: {
                disabled_at: 0,
                is_default: false,
            },
        }),
    };
});

vi.mock('../../../hooks/usePZParams/usePZQueryParams', () => ({
    usePZQueryParams: () => ({ assetGroupTagId: 1 }),
}));

vi.mock('../../../hooks/usePZParams/usePZPathParams', () => ({
    usePZPathParams: () => ({
        tagId: 1,
        ruleId: undefined,
        isZonePage: true,
        tagDetailsLink: (id: number) => `/tags/${id}`,
        ruleDetailsLink: (tagId: number, ruleId: number) => `/tags/${tagId}/rules/${ruleId}`,
    }),
}));

vi.mock('../../../hooks/useEnvironmentIdList', () => ({
    useEnvironmentIdList: () => ['env-1'],
}));

vi.mock('../../../utils', async () => {
    const actual = await vi.importActual<any>('../../../utils');
    return {
        ...actual,
        useAppNavigate: () => mockNavigate,
    };
});

vi.mock('../../../hooks/useMeasure', () => ({
    useMeasure: () => [600, 600],
}));

const server = setupServer(...zoneHandlers);

beforeAll(() => server.listen());
afterEach(() => server.resetHandlers());
afterAll(() => server.close());

const useRulesInfiniteQuerySpy = vi.spyOn(useAssetGroupTags, 'useRulesInfiniteQuery');

describe('RulesAccordion', () => {
    beforeEach(() => {
        mockNavigate.mockClear();
    });

    it('renders total rules and accordion sections', () => {
        render(<RulesAccordion />);

        expect(screen.getByText('Rules')).toBeInTheDocument();
        expect(screen.getByText(/Total Rules:/)).toBeInTheDocument();

        expect(screen.getByTestId(`privilege-zones_details_${CustomRulesKey}-accordion-item`)).toBeInTheDocument();

        expect(screen.getByTestId(`privilege-zones_details_${DefaultRulesKey}-accordion-item`)).toBeInTheDocument();

        expect(screen.getByTestId(`privilege-zones_details_${DisabledRulesKey}-accordion-item`)).toBeInTheDocument();
    });

    it('navigates to all rules when clicking "All Rules"', async () => {
        render(<RulesAccordion />);

        await userEvent.click(screen.getByText('All Rules'));

        expect(mockNavigate).toHaveBeenCalledWith('/tags/1');
    });

    it('toggles sort order when clicking sortable header', async () => {
        render(<RulesAccordion />);

        const header = screen.getByTestId('privilege-zones_details_custom_selectors-accordion-item');
        const sortButton = within(header).getByTestId('sort-button');

        await userEvent.click(sortButton);

        expect(useRulesInfiniteQuerySpy).toBeCalledWith(
            1,
            expect.objectContaining({ sortOrder: 'desc', environments: ['env-1'], isDefault: false, disabled: false }),
            true
        );

        await userEvent.click(sortButton);

        expect(useRulesInfiniteQuerySpy).toBeCalledWith(
            1,
            expect.objectContaining({ sortOrder: 'asc', environments: ['env-1'], isDefault: false, disabled: false }),
            true
        );
    });

    it('navigates to rule details when clicking a rule row', async () => {
        render(<RulesAccordion />);

        await waitFor(() => screen.getAllByRole('listitem'));

        await userEvent.click(screen.getByText('tag-0-rule-1'));

        expect(mockNavigate).toHaveBeenCalledWith('/tags/1/rules/1');
    });
});
