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
import {
    AssetGroupTagSelector,
    AssetGroupTagTypeLabel,
    CustomRulesKey,
    DefaultRulesKey,
    DisabledRulesKey,
    RulesKey,
} from 'js-client-library';
import { rest } from 'msw';
import { setupServer } from 'msw/node';
import * as useAssetGroupTags from '../../../hooks/useAssetGroupTags';
import { usePZPathParams } from '../../../hooks/usePZParams/usePZPathParams';
import { useSelectedTagPathParams } from '../../../hooks/useSelectedTag';
import { zoneHandlers } from '../../../mocks';
import { createAssetGroupTag, mockPZPathParams } from '../../../mocks/factories/privilegeZones';
import { render, screen, waitFor, within } from '../../../test-utils';
import { disabledStylesOverride } from './constants';
import { RulesAccordion } from './RulesAccordion';

const mockNavigate = vi.fn();
const mockSelectedRule = vi.hoisted(() => ({
    disabled_at: 0,
    is_default: false,
}));

const defaultPathParams = {
    ...mockPZPathParams,
    tagId: '1',
    zoneId: '1',
    labelId: undefined,
    hasZoneId: true,
    hasLabelId: false,
    tagType: 'zones' as const,
    tagTypeDisplay: 'Zone' as const,
    tagTypeDisplayPlural: 'Zones' as const,
    tagDetailsLink: (id: number | string) => `/privilege-zones/zones/${id}/details`,
    ruleDetailsLink: (tagId: number | string, ruleId: number | string) =>
        `/privilege-zones/zones/${tagId}/rules/${ruleId}/details`,
};

const labelPathParams = {
    ...defaultPathParams,
    zoneId: '',
    labelId: '1',
    hasZoneId: false,
    hasLabelId: true,
    isZonePage: false,
    isLabelPage: true,
    tagType: 'labels' as const,
    tagTypeDisplay: 'Label' as const,
    tagTypeDisplayPlural: 'Labels' as const,
    tagDetailsLink: (id: number | string) => `/privilege-zones/labels/${id}/details`,
    ruleDetailsLink: (tagId: number | string, ruleId: number | string) =>
        `/privilege-zones/labels/${tagId}/rules/${ruleId}/details`,
};

const defaultSelectedTag = {
    ...createAssetGroupTag(1),
    counts: {
        [RulesKey]: 6,
        [CustomRulesKey]: 2,
        [DefaultRulesKey]: 2,
        [DisabledRulesKey]: 2,
        members: 0,
    },
};

const labelSelectedTag = { ...defaultSelectedTag, type: AssetGroupTagTypeLabel };

vi.mock('../../../hooks/useSelectedTag', () => ({
    useSelectedTagPathParams: vi.fn(),
}));

vi.mock('../../../hooks/useAssetGroupTags', async (importOriginal) => {
    const original: Record<string, any> = await importOriginal();

    return {
        ...original,
        useRuleInfo: () => ({
            data: mockSelectedRule,
        }),
    };
});

vi.mock('../../../hooks/usePZParams/usePZQueryParams', () => ({
    usePZQueryParams: () => ({ assetGroupTagId: 1 }),
}));

vi.mock('../../../hooks/usePZParams/usePZPathParams', () => ({
    usePZPathParams: vi.fn(),
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
        useRulesInfiniteQuerySpy.mockClear();
        vi.mocked(usePZPathParams).mockReturnValue(defaultPathParams);
        vi.mocked(useSelectedTagPathParams).mockReturnValue(defaultSelectedTag);
    });

    it('renders total rules and accordion sections', () => {
        render(<RulesAccordion />);

        expect(screen.getByText('Rules')).toBeInTheDocument();
        expect(screen.getByText(/Total Rules:/)).toBeInTheDocument();

        expect(screen.getByTestId(`privilege-zones_details_${CustomRulesKey}-accordion-item`)).toBeInTheDocument();

        expect(screen.getByTestId(`privilege-zones_details_${DefaultRulesKey}-accordion-item`)).toBeInTheDocument();

        expect(screen.getByTestId(`privilege-zones_details_${DisabledRulesKey}-accordion-item`)).toBeInTheDocument();
    });

    describe.each([
        ['zone', true],
        ['label', false],
    ] as const)('%s default rules visibility', (_tagType, isZonePage) => {
        const selectedTag = isZonePage ? defaultSelectedTag : labelSelectedTag;

        beforeEach(() => {
            vi.mocked(usePZPathParams).mockReturnValue(isZonePage ? defaultPathParams : labelPathParams);
        });

        it('shows the default rules accordion when at least one default is enabled', () => {
            vi.mocked(useSelectedTagPathParams).mockReturnValue({
                ...selectedTag,
                counts: { ...selectedTag.counts, [RulesKey]: 5, [DefaultRulesKey]: 1 },
            });
            render(<RulesAccordion />);

            const accordion = screen.getByTestId(`privilege-zones_details_${DefaultRulesKey}-accordion-item`);

            expect(accordion).toBeInTheDocument();
            expect(within(accordion).getByText('1')).toBeInTheDocument();
            expect(within(accordion).getByRole('button', { name: 'Expand' })).toBeEnabled();
        });

        it.each([0, 2])(
            'hides the default rules accordion with zero enabled defaults and %i disabled rules',
            (disabledRuleCount) => {
                vi.mocked(useSelectedTagPathParams).mockReturnValue({
                    ...selectedTag,
                    counts: {
                        ...selectedTag.counts,
                        [RulesKey]: 2 + disabledRuleCount,
                        [DefaultRulesKey]: 0,
                        [DisabledRulesKey]: disabledRuleCount,
                    },
                });
                render(<RulesAccordion />);

                expect(
                    screen.queryByTestId(`privilege-zones_details_${DefaultRulesKey}-accordion-item`)
                ).not.toBeInTheDocument();
                expect(
                    screen.getByTestId(`privilege-zones_details_${CustomRulesKey}-accordion-item`)
                ).toBeInTheDocument();
                expect(
                    screen.getByTestId(`privilege-zones_details_${DisabledRulesKey}-accordion-item`)
                ).toBeInTheDocument();
                if (disabledRuleCount > 0) {
                    expect(
                        screen.getByTestId(`privilege-zones_details_${DisabledRulesKey}-accordion_open-toggle-button`)
                    ).toBeEnabled();
                }
            }
        );
    });

    it('navigates to rule details when clicking a default rule on a label', async () => {
        const defaultRule: AssetGroupTagSelector = {
            id: 1,
            asset_group_tag_id: 1,
            name: 'Default label rule',
            description: 'A built-in rule for this label',
            is_default: true,
            allow_disable: true,
            auto_certify: 0,
            disabled_at: null,
            disabled_by: '',
            created_at: '2026-01-01T00:00:00Z',
            created_by: 'System',
            updated_at: '2026-01-01T00:00:00Z',
            updated_by: 'System',
            seeds: [],
            counts: { members: 0 },
        };

        vi.mocked(usePZPathParams).mockReturnValue(labelPathParams);
        vi.mocked(useSelectedTagPathParams).mockReturnValue({
            ...labelSelectedTag,
            counts: { ...labelSelectedTag.counts, [RulesKey]: 5, [DefaultRulesKey]: 1 },
        });
        server.use(
            rest.get('/api/v2/asset-group-tags/1/selectors', (request, response, context) => {
                const isEnabledDefaultQuery =
                    request.url.searchParams.get('is_default') === 'eq:true' &&
                    request.url.searchParams.get('disabled_at') === 'eq:null';

                return response(context.json({ data: { selectors: isEnabledDefaultQuery ? [defaultRule] : [] } }));
            })
        );
        render(<RulesAccordion />);

        const accordion = screen.getByTestId(`privilege-zones_details_${DefaultRulesKey}-accordion-item`);

        await userEvent.click(within(accordion).getByRole('button', { name: 'Expand' }));

        expect(useRulesInfiniteQuerySpy).toHaveBeenCalledWith(
            '1',
            expect.objectContaining({ isDefault: true, disabled: false }),
            true
        );

        await userEvent.click(await within(accordion).findByRole('button', { name: defaultRule.name }));

        expect(mockNavigate).toHaveBeenCalledWith('/privilege-zones/labels/1/rules/1/details');
    });

    it('updates the toggle when collapsing the open accordion', async () => {
        render(<RulesAccordion />);

        const toggle = screen.getByRole('button', { name: 'Collapse' });

        expect(toggle).toHaveClass('rotate-180');

        await userEvent.click(toggle);

        expect(toggle).toHaveAccessibleName('Expand');
        expect(toggle).not.toHaveClass('rotate-180');
    });

    it('navigates to all rules when clicking "All Rules"', async () => {
        render(<RulesAccordion />);

        await userEvent.click(screen.getByText('All Rules'));

        expect(mockNavigate).toHaveBeenCalledWith('/privilege-zones/zones/1/details');
    });

    it('toggles sort order when clicking sortable header', async () => {
        render(<RulesAccordion />);

        const header = screen.getByTestId('privilege-zones_details_custom_selectors-accordion-item');
        const sortButton = within(header).getByTestId('column-header_sort-button');

        await userEvent.click(sortButton);

        expect(useRulesInfiniteQuerySpy).toBeCalledWith(
            '1',
            expect.objectContaining({ sortOrder: 'asc', environments: ['env-1'], isDefault: false, disabled: false }),
            true
        );

        await userEvent.click(sortButton);

        expect(useRulesInfiniteQuerySpy).toBeCalledWith(
            '1',
            expect.objectContaining({ sortOrder: 'asc', environments: ['env-1'], isDefault: false, disabled: false }),
            true
        );
    });

    it('disables sortable headers only while their accordions are closed', () => {
        render(<RulesAccordion />);

        const customRulesSortButton = screen.getByRole('button', { name: 'Sort by Custom Rules' });
        const defaultRulesSortButton = screen.getByRole('button', { name: 'Sort by Default Rules' });

        expect(customRulesSortButton).toBeEnabled();
        expect(defaultRulesSortButton).toBeDisabled();
        expect(defaultRulesSortButton).toHaveClass(...disabledStylesOverride.split(' '));
    });

    it('navigates to rule details when clicking a rule row', async () => {
        render(<RulesAccordion />);

        await waitFor(() => screen.getAllByRole('listitem'));

        await userEvent.click(screen.getByText('tag-0-rule-1'));

        expect(mockNavigate).toHaveBeenCalledWith('/privilege-zones/zones/1/rules/1/details');
        ``;
    });
});
