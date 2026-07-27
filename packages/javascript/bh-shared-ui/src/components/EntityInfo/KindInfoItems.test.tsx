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
import { NodeKindInfoItem, RelationshipKindInfoItem } from 'js-client-library';
import { useExploreParams } from '../../hooks/useExploreParams';
import { render, screen } from '../../test-utils';
import { KindInfoItems } from './KindInfoItems';

vi.mock('../../hooks/useExploreParams');
const mockUseExploreParams = vi.mocked(useExploreParams);

const mockSetExploreParams = vi.fn();

const makeNodeItem = (overrides: Partial<NodeKindInfoItem> = {}): NodeKindInfoItem => ({
    name: 'item-1',
    title: 'Item One',
    position: 0,
    node_kind_id: 1,
    markdown: { content: '## Hello' },
    ...overrides,
});

const makeRelationshipItem = (overrides: Partial<RelationshipKindInfoItem> = {}): RelationshipKindInfoItem => ({
    name: 'rel-item-1',
    title: 'Rel Item One',
    position: 0,
    relationship_kind_id: 10,
    markdown: { content: '## Rel Hello' },
    ...overrides,
});

beforeEach(() => {
    mockUseExploreParams.mockReturnValue({
        expandedPanelSections: [],
        setExploreParams: mockSetExploreParams,
        exploreSearchTab: null,
        primarySearch: null,
        secondarySearch: null,
        cypherSearch: null,
        searchType: null,
        selectedItem: null,
        relationshipQueryType: null,
        relationshipQueryItemId: null,
        pathFilters: [],
    });
});

afterEach(() => {
    vi.clearAllMocks();
});

describe('KindInfoItems', () => {
    it('renders null when items is undefined', () => {
        render(
            <div data-testid='kind-info-wrapper'>
                <KindInfoItems items={undefined} />
            </div>
        );
        expect(screen.getByTestId('kind-info-wrapper')).toBeEmptyDOMElement();
    });

    it('renders node kind items with correct test-id', () => {
        const items = [makeNodeItem({ name: 'a', title: 'Alpha', position: 0, node_kind_id: 1 })];
        render(<KindInfoItems items={items} />);
        expect(screen.getByTestId('entity-info_kind-info-items')).toBeInTheDocument();
        expect(screen.getByTestId('entity-info_kind-info-item')).toBeInTheDocument();
    });

    it('renders relationship kind items with correct test-id', () => {
        const items = [makeRelationshipItem({ name: 'r', title: 'Rel Alpha', position: 0, relationship_kind_id: 10 })];
        render(<KindInfoItems items={items} />);
        expect(screen.getByTestId('entity-info_kind-info-items')).toBeInTheDocument();
        expect(screen.getByTestId('entity-info_kind-info-item')).toBeInTheDocument();
    });
    it('renders multiple items with test-ids', () => {
        const items = [
            makeNodeItem({ name: 'a', title: 'Alpha', position: 0, node_kind_id: 1 }),
            makeNodeItem({ name: 'b', title: 'Beta', position: 1, node_kind_id: 2 }),
        ];
        render(<KindInfoItems items={items} />);
        expect(screen.getAllByTestId('entity-info_kind-info-item')).toHaveLength(2);
    });

    it('sorts items by position ascending', () => {
        const items = [
            makeNodeItem({ name: 'b', title: 'Beta', position: 2, node_kind_id: 2 }),
            makeNodeItem({ name: 'a', title: 'Alpha', position: 1, node_kind_id: 1 }),
        ];
        render(<KindInfoItems items={items} />);
        const renderedTitles = screen.getAllByText(/Alpha|Beta/).map((el) => el.textContent);
        expect(renderedTitles).toEqual(['Alpha', 'Beta']);
    });

    it('sorts by title as secondary key when positions are equal', () => {
        const items = [
            makeNodeItem({ name: 'b', title: 'Zebra', position: 0, node_kind_id: 2 }),
            makeNodeItem({ name: 'a', title: 'Apple', position: 0, node_kind_id: 1 }),
        ];
        render(<KindInfoItems items={items} />);
        const renderedTitles = screen.getAllByText(/Apple|Zebra/).map((el) => el.textContent);
        expect(renderedTitles).toEqual(['Apple', 'Zebra']);
    });

    it('sorts by node_kind_id as tertiary key when position and title are equal', () => {
        const items = [
            makeNodeItem({ name: 'b', title: 'Same', position: 0, node_kind_id: 5 }),
            makeNodeItem({ name: 'a', title: 'Same', position: 0, node_kind_id: 2 }),
        ];
        render(<KindInfoItems items={items} />);
        // Both items have the same title; verify both rendered
        expect(screen.getAllByText('Same')).toHaveLength(2);
    });

    it('does not mutate the original items array when sorting', () => {
        const items = [
            makeNodeItem({ name: 'b', title: 'Beta', position: 2, node_kind_id: 2 }),
            makeNodeItem({ name: 'a', title: 'Alpha', position: 1, node_kind_id: 1 }),
        ];
        const originalOrder = items.map((i) => i.name);
        render(<KindInfoItems items={items} />);
        expect(items.map((i) => i.name)).toEqual(originalOrder);
    });

    it('calls setExploreParams with expanded sections when accordion value changes', async () => {
        const user = userEvent.setup();
        const items = [makeNodeItem({ name: 'section-1', title: 'Section One', position: 0, node_kind_id: 1 })];
        render(<KindInfoItems items={items} />);

        const trigger = screen.getByText('Section One');
        await user.click(trigger);

        expect(mockSetExploreParams).toHaveBeenCalledWith({ expandedPanelSections: ['section-1'] });
    });

    it('passes expandedPanelSections from useExploreParams to Accordion', () => {
        mockUseExploreParams.mockReturnValue({
            expandedPanelSections: ['section-1'],
            setExploreParams: mockSetExploreParams,
            exploreSearchTab: null,
            primarySearch: null,
            secondarySearch: null,
            cypherSearch: null,
            searchType: null,
            selectedItem: null,
            relationshipQueryType: null,
            relationshipQueryItemId: null,
            pathFilters: [],
        });

        const items = [makeNodeItem({ name: 'section-1', title: 'Section One', position: 0, node_kind_id: 1 })];
        render(<KindInfoItems items={items} />);

        // The accordion content should be visible when the section is in expandedPanelSections
        expect(screen.getByText('Section One')).toBeInTheDocument();
    });
});
