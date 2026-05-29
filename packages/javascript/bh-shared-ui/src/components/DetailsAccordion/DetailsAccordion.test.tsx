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

import { render, screen } from '../../test-utils';
import { DetailsAccordion } from './DetailsAccordion';

// Need test ID for icon
vi.mock('@fortawesome/react-fontawesome', () => ({
    FontAwesomeIcon: ({ className }: any) => <span data-testid='fa' className={className} />,
}));

vi.mock('@fortawesome/free-solid-svg-icons', () => ({ faAngleDown: {} }));

type Item = { title: string; description: string; disabled?: boolean };

const Header: React.FC<Item> = (item) => <h3>{item.title}</h3>;
const Content: React.FC<Item> = (item) => <p>{item.description}</p>;
const Empty: React.FC = () => <div role='status'>No items</div>;

describe('<DetailsAccordion />', () => {
    it('renders Empty when items is undefined and Empty is provided', () => {
        render(<DetailsAccordion<Item> Header={Header} Content={Content} Empty={Empty} />);
        expect(screen.getByRole('status')).toHaveTextContent('No items');
    });

    it('renders a single item when items is an object', () => {
        render(
            <DetailsAccordion<Item>
                Header={Header}
                Content={Content}
                items={{ title: 'One', description: 'Only child' }}
            />
        );
        expect(screen.getByRole('button')).toHaveTextContent('One');
    });

    it('passes item props to Header and Content', async () => {
        const user = userEvent.setup();
        render(
            <DetailsAccordion<Item>
                Header={Header}
                Content={Content}
                items={[
                    { title: 'A', description: 'Alpha' },
                    { title: 'B', description: 'Beta' },
                ]}
                openIndex={0}
            />
        );

        // default open is index 0 -> content "Alpha" should be visible
        expect(screen.getByText('Alpha')).toBeInTheDocument();

        // Click header B, then "Beta" should appear
        await user.click(screen.getByRole('button', { name: 'B' }));
        expect(screen.getByText('Beta')).toBeInTheDocument();
    });

    it('calls getKey for each item (stable keys hook)', () => {
        const items = [
            { title: 'A', description: 'Alpha' },
            { title: 'B', description: 'Beta' },
        ];
        const getKey = vi.fn((i: Item) => i.title);

        render(<DetailsAccordion<Item> Header={Header} Content={Content} items={items} getKey={getKey} />);

        expect(getKey).toHaveBeenCalledTimes(2);
        expect(getKey).toHaveBeenNthCalledWith(1, items[0], 0);
        expect(getKey).toHaveBeenNthCalledWith(2, items[1], 1);
    });

    it('skips undefined items gracefully', () => {
        render(
            <DetailsAccordion<any>
                Header={({ label }: any) => <div>{label}</div>}
                Content={() => <div>ok</div>}
                // an undefined item
                items={[{ label: 'Valid' }, undefined]}
            />
        );
        expect(screen.getByRole('button')).toHaveTextContent('Valid');
        // only one header rendered
        expect(screen.getAllByRole('button')).toHaveLength(1);
    });

    it('applies accent styling to headers when accent is true', () => {
        render(
            <DetailsAccordion<Item>
                Header={Header}
                Content={Content}
                items={[{ title: 'Fancy', description: 'With accent' }]}
                accent
            />
        );

        const header = screen.getByRole('button');
        // spot-check invariant classes from implementation
        expect(header).toHaveClass('border-l-8');
        expect(header).toHaveClass('border-primary');
    });

    it('disables items via itemDisabled and hides the chevron icon visually', () => {
        render(
            <DetailsAccordion<Item>
                Header={Header}
                Content={Content}
                items={[
                    { title: 'Disabled', description: 'Nope', disabled: true },
                    { title: 'Enabled', description: 'Yep', disabled: false },
                ]}
                itemDisabled={(i) => !!i.disabled}
            />
        );

        const [disabledHeader, enabledHeader] = screen.getAllByRole('button');
        expect(disabledHeader).toHaveAttribute('data-disabled');

        // Icon for disabled row should have opacity-0
        const disabledIcon = disabledHeader.querySelector('[data-testid="fa"]');
        expect(disabledIcon).toHaveClass('opacity-0');

        const enabledIcon = enabledHeader.querySelector('[data-testid="fa"]');
        expect(enabledIcon).not.toHaveClass('opacity-0');
    });

    it('respects openIndex (default opened panel)', () => {
        render(
            <DetailsAccordion<Item>
                Header={Header}
                Content={Content}
                items={[
                    { title: 'First', description: 'one' },
                    { title: 'Second', description: 'two' },
                ]}
                openIndex={1}
            />
        );
        const firstHeader = screen.getByRole('button', { name: /First/i });
        const secondHeader = screen.getByRole('button', { name: /Second/i });
        expect(firstHeader).toHaveAttribute('aria-expanded', 'false');
        expect(secondHeader).toHaveAttribute('aria-expanded', 'true');
    });

    it('clicking a disabled header does not toggle content', async () => {
        const user = userEvent.setup();
        render(
            <DetailsAccordion<Item>
                Header={Header}
                Content={Content}
                items={[{ title: 'Locked', description: 'hidden', disabled: true }]}
                itemDisabled={(i) => !!i.disabled}
            />
        );

        const header = screen.getByRole('button', { name: /Locked/i });
        await user.click(header);

        // never opens
        expect(screen.queryByText('hidden')).not.toBeInTheDocument();
    });

    it('clicking an enabled header toggles content (single mode)', async () => {
        const user = userEvent.setup();
        render(
            <DetailsAccordion<Item>
                Header={Header}
                Content={Content}
                items={[
                    { title: 'A', description: 'Alpha' },
                    { title: 'B', description: 'Beta' },
                ]}
            />
        );

        // Initially nothing is open (no openIndex)
        expect(screen.queryByText('Alpha')).not.toBeInTheDocument();
        expect(screen.queryByText('Beta')).not.toBeInTheDocument();

        const a = screen.getByRole('button', { name: /A/i });
        const b = screen.getByRole('button', { name: /B/i });

        // Open A
        await user.click(a);
        expect(a).toHaveAttribute('aria-expanded', 'true');

        // Open B (single mode should close A)
        await user.click(b);
        expect(a).toHaveAttribute('aria-expanded', 'false');
        expect(b).toHaveAttribute('aria-expanded', 'true');
    });

    it('collapses an open item when its header is clicked again', async () => {
        const user = userEvent.setup();
        render(
            <DetailsAccordion<Item>
                Header={Header}
                Content={Content}
                items={[{ title: 'A', description: 'Alpha' }]}
                openIndex={0}
            />
        );
        const header = screen.getByRole('button', { name: /A/i });
        expect(screen.getByText('Alpha')).toBeInTheDocument();
        await user.click(header);
        // Prefer visibility or aria-expanded as noted above
        expect(header).toHaveAttribute('aria-expanded', 'false');
    });
});
