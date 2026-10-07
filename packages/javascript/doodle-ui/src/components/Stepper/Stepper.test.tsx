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
import '@testing-library/jest-dom';
import matchers from '@testing-library/jest-dom/matchers';
import { fireEvent, render, screen, within } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { useEffect, useState } from 'react';
import {
    BasicStepper,
    Stepper,
    StepperContent,
    StepperDescription,
    StepperIndicator,
    StepperItem,
    StepperNav,
    StepperPanel,
    StepperSeparator,
    StepperTitle,
    StepperTrigger,
    useStepper,
    type Step,
} from './index';

expect.extend(matchers);

function NextStep() {
    const { activeStep, setActiveStep } = useStepper();
    return <button onClick={() => setActiveStep(activeStep + 1)}>Continue</button>;
}

const steps: Step[] = [
    { title: 'Profile', content: <NextStep /> },
    { title: 'Secret', content: 'Secret content' },
    { title: 'Schedule', content: 'Schedule content' },
];

const getTab = (oneBasedIndex: number, triggerTitle?: string) =>
    screen.getByRole('tab', { name: `${oneBasedIndex} ${triggerTitle ?? steps[oneBasedIndex - 1].title}` });

describe('BasicStepper', () => {
    it('supports navigation from step content in uncontrolled mode', async () => {
        const onValueChange = vi.fn();
        render(<BasicStepper steps={steps} onValueChange={onValueChange} />);

        await userEvent.click(screen.getByRole('button', { name: 'Continue' }));

        expect(getTab(2)).toHaveAttribute('aria-selected', 'true');
        expect(screen.getByText('Secret content')).toBeInTheDocument();
        expect(screen.queryByRole('button', { name: 'Continue' })).not.toBeInTheDocument();
        expect(onValueChange).toHaveBeenCalledExactlyOnceWith(2);
    });

    // Multiple steps and checks because we specifically want to test that the state is shared between all controls/triggers.
    it('shares controlled state with outside controls, inside controls, and step triggers', async () => {
        function ControlledStepper() {
            const [value, setValue] = useState(1);
            return (
                <>
                    <button onClick={() => setValue(3)}>Go to schedule</button>
                    <BasicStepper steps={steps} value={value} onValueChange={setValue} />
                </>
            );
        }
        render(<ControlledStepper />);

        expect(getTab(1)).toHaveAttribute('aria-selected', 'true');
        await userEvent.click(screen.getByRole('button', { name: 'Go to schedule' }));
        expect(screen.getByText('Schedule content')).toBeInTheDocument();
        expect(getTab(3)).toHaveAttribute('aria-selected', 'true');
        expect(getTab(1)).toHaveAttribute('aria-selected', 'false');

        await userEvent.click(getTab(1));
        expect(screen.queryByText('Schedule content')).not.toBeInTheDocument();

        await userEvent.click(screen.getByRole('button', { name: 'Continue' }));
        expect(screen.getByText('Secret content')).toBeInTheDocument();
    });

    it('waits for the parent to accept a controlled navigation request', async () => {
        const onValueChange = vi.fn();
        const { rerender } = render(<BasicStepper steps={steps} value={1} onValueChange={onValueChange} />);

        await userEvent.click(screen.getByRole('button', { name: 'Continue' }));

        expect(onValueChange).toHaveBeenCalledExactlyOnceWith(2);
        // Since we're not actually updating the value prop (just triggering onValueChange), the content shouldn't change yet.
        expect(getTab(1)).toHaveAttribute('aria-selected', 'true');
        expect(screen.queryByText('Secret content')).not.toBeInTheDocument();

        // Simulate what would be a state change in a parent component, passing `value={2}` to the BasicStepper.
        rerender(<BasicStepper steps={steps} value={2} onValueChange={onValueChange} />);
        // Content should now change, and onValueChange should not be called again.
        expect(screen.getByText('Secret content')).toBeInTheDocument();
        expect(onValueChange).toHaveBeenCalledTimes(1);
    });

    it('uses defaultValue only for initial selection', async () => {
        const { rerender } = render(<BasicStepper steps={steps} defaultValue={2} />);
        expect(screen.getByText('Secret content')).toBeInTheDocument();
        expect(getTab(2)).toHaveAttribute('aria-selected', 'true');

        await userEvent.click(getTab(3));
        rerender(<BasicStepper steps={steps} defaultValue={1} />);

        expect(screen.getByText('Schedule content')).toBeInTheDocument();
        expect(getTab(2)).toHaveAttribute('aria-selected', 'false');
    });

    it.each([undefined, 1])('rejects disabled navigation requests with value=%s until enabled', async (value) => {
        const onValueChange = vi.fn();
        const disabledSteps = steps.map((step) => ({ ...step, isDisabled: step.title === 'Secret' }));
        const { rerender } = render(<BasicStepper steps={disabledSteps} value={value} onValueChange={onValueChange} />);
        const disabledTrigger = getTab(2);

        expect(disabledTrigger).toBeDisabled();
        expect(disabledTrigger).toHaveAttribute('aria-disabled', 'true');
        expect(disabledTrigger).toHaveAttribute('tabindex', '-1');
        await userEvent.click(disabledTrigger);
        fireEvent.keyDown(disabledTrigger, { key: 'Enter' });
        fireEvent.keyDown(disabledTrigger, { key: ' ' });
        expect(onValueChange).not.toHaveBeenCalled();
        expect(screen.queryByText('Secret content')).not.toBeInTheDocument();

        await userEvent.click(screen.getByRole('button', { name: 'Continue' }));
        expect(getTab(1)).toHaveAttribute('aria-selected', 'true');
        expect(screen.queryByText('Secret content')).not.toBeInTheDocument();
        expect(onValueChange).not.toHaveBeenCalled();

        // replace `disabledSteps` with `steps` to enable the second step on rerender
        rerender(<BasicStepper steps={steps} value={value} onValueChange={onValueChange} />);
        expect(getTab(2)).toBeEnabled();
        await userEvent.click(screen.getByRole('button', { name: 'Continue' }));
        expect(onValueChange).toHaveBeenCalledExactlyOnceWith(2);
    });

    it('allows parent controls to enable and select a step together', async () => {
        function EnableAndContinue() {
            const [value, setValue] = useState(1);
            const [isEnabled, setIsEnabled] = useState(false);
            return (
                <>
                    <BasicStepper
                        steps={steps.map((step) => ({ ...step, isDisabled: step.title === 'Secret' && !isEnabled }))}
                        value={value}
                        onValueChange={setValue}
                    />
                    <button
                        onClick={() => {
                            setIsEnabled(true);
                            setValue(2);
                        }}>
                        Complete profile
                    </button>
                </>
            );
        }
        render(<EnableAndContinue />);

        await userEvent.click(screen.getByRole('button', { name: 'Continue' }));
        expect(screen.queryByText('Secret content')).not.toBeInTheDocument();
        expect(getTab(2)).toBeDisabled();

        await userEvent.click(screen.getByRole('button', { name: 'Complete profile' }));

        expect(getTab(2)).toBeEnabled();
        expect(screen.getByText('Secret content')).toBeInTheDocument();
    });

    it('skips disabled triggers during keyboard navigation and wraps focus', async () => {
        const user = userEvent.setup();
        render(<BasicStepper steps={steps.map((step) => ({ ...step, isDisabled: step.title === 'Secret' }))} />);
        const first = getTab(1);
        const last = getTab(3);
        first.focus();

        await user.keyboard('{ArrowRight}');
        expect(last).toHaveFocus();
        await user.keyboard('{Enter}');
        expect(screen.getByText('Schedule content')).toBeInTheDocument();
    });

    it('wraps focus during keyboard navigation', async () => {
        const user = userEvent.setup();
        render(<BasicStepper steps={steps.slice(0, 2)} />);
        const first = getTab(1);
        const second = getTab(2);
        first.focus();

        await user.keyboard('{ArrowRight}');
        expect(second).toHaveFocus();
        await user.keyboard('{ArrowRight}');
        expect(first).toHaveFocus();
        await user.keyboard('{ArrowLeft}');
        expect(second).toHaveFocus();
        await user.keyboard('{Enter}');
        expect(screen.getByText('Secret content')).toBeInTheDocument();
    });

    it('goes to the first and last steps with Home and End keys', async () => {
        const user = userEvent.setup();
        const fourSteps = [...steps, { title: 'Review', content: 'Review content' }];
        render(<BasicStepper steps={fourSteps} defaultValue={2} />);
        const second = getTab(2);
        expect(second).toHaveAttribute('aria-selected', 'true');
        second.focus();
        await user.keyboard('{Home}');
        expect(getTab(1)).toHaveFocus();
        await user.keyboard('{End}');
        expect(getTab(4, 'Review')).toHaveFocus();
    });

    it('uses enabled endpoints for Home and End and responds to disabled prop changes', async () => {
        const user = userEvent.setup();
        const fourSteps = [...steps, { title: 'Review', content: 'Review content' }];
        render(
            <BasicStepper
                steps={fourSteps.map((step, index) => ({ ...step, isDisabled: index === 0 || index === 3 }))}
                defaultValue={2}
            />
        );
        const second = getTab(2);
        const third = getTab(3);
        second.focus();

        await user.keyboard('{End}');
        expect(third).toHaveFocus(); // Not fourth because it's disabled
        await user.keyboard('{Home}');
        expect(second).toHaveFocus(); // Not first because it's disabled
        await user.keyboard('{ArrowDown}');
        expect(third).toHaveFocus(); // Next value
        await user.keyboard('{ArrowUp}');
        expect(second).toHaveFocus(); // Previous value
    });

    it('removes unmounted triggers from keyboard navigation', async () => {
        const user = userEvent.setup();
        const { rerender } = render(<BasicStepper steps={steps} />);
        const [profileTrigger, , scheduleTrigger] = steps;
        rerender(<BasicStepper steps={[profileTrigger, scheduleTrigger]} />);
        getTab(1).focus();

        await user.keyboard('{ArrowRight}');
        // Tab 2 should now be 'Schedule' because 'Secret' was removed from the DOM.
        expect(getTab(2, 'Schedule')).toHaveFocus();
    });

    it('uses current DOM order after keyed insertion and reordering', async () => {
        const user = userEvent.setup();
        const { rerender } = render(<BasicStepper steps={[steps[0], steps[2]]} />);
        rerender(<BasicStepper steps={steps} />);
        getTab(1).focus();
        await user.keyboard('{ArrowRight}');
        expect(getTab(2)).toHaveFocus();
        await user.keyboard('{End}');
        expect(getTab(3)).toHaveFocus();

        rerender(<BasicStepper steps={[steps[2], steps[0], steps[1]]} />);
        getTab(2, 'Profile').focus();
        await user.keyboard('{ArrowRight}');
        expect(getTab(3, 'Secret')).toHaveFocus();
        await user.keyboard('{Home}');
        expect(getTab(1, 'Schedule')).toHaveFocus();
    });

    it('counts registered descendants in BasicStepper as steps are added and removed', () => {
        function StepCount() {
            const { stepsCount } = useStepper();
            return <p role='status'>{stepsCount} steps</p>;
        }
        const countedSteps = steps.map((step, index) => ({
            ...step,
            content: index === 0 ? <StepCount /> : step.content,
        }));
        const { rerender } = render(<BasicStepper steps={countedSteps.slice(0, 2)} />);
        expect(screen.getByRole('status')).toHaveTextContent('2 steps');
        rerender(<BasicStepper steps={countedSteps} />);
        expect(screen.getByRole('status')).toHaveTextContent('3 steps');
        rerender(<BasicStepper steps={countedSteps.slice(0, 1)} />);
        expect(screen.getByRole('status')).toHaveTextContent('1 steps');
    });

    it('scopes tab relationships to each instance and keeps inactive panel targets mounted', async () => {
        const user = userEvent.setup();
        render(
            <>
                <BasicStepper steps={steps} />
                <BasicStepper steps={steps} />
            </>
        );
        const tabLists = screen.getAllByRole('tablist');
        expect(screen.getAllByRole('tabpanel')).toHaveLength(2);
        expect(screen.getAllByRole('tabpanel', { hidden: true })).toHaveLength(6);
        const ids = new Set<string>();
        for (const tablist of tabLists) {
            expect(tablist.tagName).toBe('DIV');
            expect(tablist).toHaveAttribute('aria-orientation', 'horizontal');
            expect(within(tablist).queryByRole('tabpanel', { hidden: true })).not.toBeInTheDocument();
            for (const tab of within(tablist).getAllByRole('tab')) {
                expect(ids.has(tab.id)).toBe(false);
                ids.add(tab.id);
                const panel = document.getElementById(tab.getAttribute('aria-controls')!);
                expect(panel).toHaveAttribute('role', 'tabpanel');
                expect(panel).toHaveAttribute('aria-labelledby', tab.id);
                expect(panel).toHaveAttribute('tabindex', '0');
                if (tab.getAttribute('aria-selected') === 'false') {
                    expect(panel).not.toBeVisible();
                    expect(panel).toBeEmptyDOMElement();
                }
            }
        }
        const secretTab = within(tabLists[0]).getByRole('tab', { name: '2 Secret' });
        await user.click(secretTab);
        expect(secretTab).toHaveAttribute('aria-selected', 'true');
        expect(within(tabLists[1]).getByRole('tab', { name: '1 Profile' })).toHaveAttribute('aria-selected', 'true');
        expect(within(tabLists[1]).getByRole('tab', { name: '2 Secret' })).toHaveAttribute('aria-selected', 'false');
        await user.tab();
        expect(document.getElementById(secretTab.getAttribute('aria-controls')!)).toHaveFocus();
    });

    it('provides an enabled Tab stop when the initial or updated active step is disabled', async () => {
        const user = userEvent.setup();
        const firstDisabled = steps.map((step, index) => ({ ...step, isDisabled: index === 0 }));
        const { rerender } = render(<BasicStepper steps={firstDisabled} />);
        expect(getTab(1)).toHaveAttribute('tabindex', '-1');
        expect(getTab(2)).toHaveAttribute('tabindex', '0');
        await user.tab();
        expect(getTab(2)).toHaveFocus();

        rerender(<BasicStepper steps={steps} value={3} />);
        expect(getTab(3)).toHaveAttribute('tabindex', '0');
        rerender(<BasicStepper steps={steps.map((step, index) => ({ ...step, isDisabled: index === 2 }))} value={3} />);
        expect(getTab(3)).toHaveAttribute('tabindex', '-1');
        expect(getTab(1)).toHaveAttribute('tabindex', '0');
        // While tab 2 is still accessible via arrow keys when inside the stepper focus, it is not a tab stop; Only the first enabled tab is a tab stop.
        expect(getTab(2)).toHaveAttribute('tabindex', '-1');
    });
});

describe('Stepper navigation', () => {
    it('rejects disabled requests from mount effects without depending on a trigger', () => {
        function NavigateOnMount() {
            const { setActiveStep } = useStepper();
            useEffect(() => setActiveStep(2), [setActiveStep]);
            return null;
        }
        const onValueChange = vi.fn();
        render(
            <Stepper onValueChange={onValueChange}>
                <NavigateOnMount />
                <StepperItem step={1}>
                    <StepperContent value={1}>Profile content</StepperContent>
                </StepperItem>
                <StepperItem step={2} isDisabled>
                    <StepperContent value={2}>Secret content</StepperContent>
                </StepperItem>
            </Stepper>
        );

        expect(screen.getByText('Profile content')).toBeInTheDocument();
        expect(screen.queryByText('Secret content')).not.toBeInTheDocument();
        expect(onValueChange).not.toHaveBeenCalled();
    });

    it('provides a fallback when a selected trigger is removed or disabled directly', () => {
        function Triggers({ showFirst = true, disabled = false }) {
            return (
                <Stepper>
                    <StepperNav>
                        {showFirst && (
                            <StepperItem step={1}>
                                <StepperTrigger disabled={disabled}>Profile</StepperTrigger>
                            </StepperItem>
                        )}
                        <StepperItem step={2}>
                            <StepperTrigger>Secret</StepperTrigger>
                        </StepperItem>
                    </StepperNav>
                </Stepper>
            );
        }
        const { rerender } = render(<Triggers />);
        expect(screen.getByRole('tab', { name: 'Profile' })).toHaveAttribute('tabindex', '0');
        expect(screen.getByRole('tab', { name: 'Secret' })).toHaveAttribute('tabindex', '-1');
        rerender(<Triggers disabled />);
        expect(screen.getByRole('tab', { name: 'Secret' })).toHaveAttribute('tabindex', '0');
        rerender(<Triggers showFirst={false} />);
        expect(screen.getByRole('tab', { name: 'Secret' })).toHaveAttribute('tabindex', '0');
    });
});

describe('StepperIndicator', () => {
    it('prefers the loading indicator over the active indicator', () => {
        render(
            <Stepper indicators={{ loading: 'Loading', active: 'Active' }}>
                <StepperItem step={1} isLoading>
                    <StepperIndicator>1</StepperIndicator>
                </StepperItem>
            </Stepper>
        );

        expect(screen.getByText('Loading')).toBeInTheDocument();
        expect(screen.queryByText('Active')).not.toBeInTheDocument();
        expect(screen.queryByText('1')).not.toBeInTheDocument();
    });

    it('falls back to the state indicator and then children when loading has no indicator', () => {
        render(
            <Stepper indicators={{ active: 'Active' }}>
                <StepperItem step={1} isLoading>
                    <StepperIndicator>1</StepperIndicator>
                </StepperItem>
                <StepperItem step={2}>
                    <StepperIndicator>2</StepperIndicator>
                </StepperItem>
                <StepperItem step={3} isLoading>
                    <StepperIndicator>3</StepperIndicator>
                </StepperItem>
            </Stepper>
        );

        expect(screen.getByText('Active')).toBeInTheDocument();
        expect(screen.getByText('2')).toBeInTheDocument();
        expect(screen.getByText('3')).toBeInTheDocument();
    });

    it('updates indicators when only the indicators prop changes', () => {
        const children = (
            <StepperItem step={1}>
                <StepperIndicator>1</StepperIndicator>
            </StepperItem>
        );
        const { rerender } = render(<Stepper indicators={{ active: 'Original' }}>{children}</Stepper>);

        rerender(<Stepper indicators={{ active: 'Updated' }}>{children}</Stepper>);

        expect(screen.getByText('Updated')).toBeInTheDocument();
        expect(screen.queryByText('Original')).not.toBeInTheDocument();
    });
});

describe('StepperContent', () => {
    it('retains force-mounted inactive children inside a hidden panel', () => {
        render(
            <Stepper>
                <StepperNav>
                    <StepperItem step={1}>
                        <StepperTrigger>Profile</StepperTrigger>
                    </StepperItem>
                    <StepperItem step={2}>
                        <StepperTrigger>Secret</StepperTrigger>
                    </StepperItem>
                </StepperNav>
                <StepperPanel>
                    <StepperContent value={1}>Profile content</StepperContent>
                    <StepperContent value={2} forceMount>
                        Secret content
                    </StepperContent>
                </StepperPanel>
            </Stepper>
        );
        expect(screen.getByText('Secret content')).toBeInTheDocument();
        expect(screen.getByText('Secret content')).not.toBeVisible();
        expect(screen.getAllByRole('tabpanel')).toHaveLength(1);
    });
});

describe('StepperTrigger', () => {
    it('does not submit a form when changing steps', async () => {
        const onSubmit = vi.fn((event) => event.preventDefault());
        render(
            <form onSubmit={onSubmit}>
                <BasicStepper steps={steps} />
            </form>
        );
        await userEvent.click(getTab(2));
        expect(getTab(2)).toHaveAttribute('type', 'button');
        expect(screen.getByText('Secret content')).toBeInTheDocument();
        expect(onSubmit).not.toHaveBeenCalled();
    });

    it('allows an explicit submit type on a trigger', async () => {
        const onSubmit = vi.fn((event) => event.preventDefault());
        render(
            <form onSubmit={onSubmit}>
                <Stepper>
                    <StepperItem step={1}>
                        <StepperTrigger type='submit'>Submit</StepperTrigger>
                    </StepperItem>
                </Stepper>
            </form>
        );
        await userEvent.click(screen.getByRole('tab', { name: 'Submit' }));
        expect(onSubmit).toHaveBeenCalledOnce();
    });
});

describe('Stepper primitives', () => {
    it('forwards native props while preserving state and panel visibility', async () => {
        const onClick = vi.fn();
        render(
            <Stepper orientation='vertical'>
                <StepperNav aria-label='Setup steps'>
                    <StepperItem step={1}>
                        <StepperTrigger>
                            <StepperIndicator aria-hidden='true' data-testid='indicator'>
                                1
                            </StepperIndicator>
                            <StepperTitle id='profile-title'>Profile</StepperTitle>
                            <StepperDescription id='profile-description'>Details</StepperDescription>
                        </StepperTrigger>
                        <StepperSeparator aria-hidden='true' data-testid='separator' />
                    </StepperItem>
                </StepperNav>
                <StepperPanel aria-label='Step contents'>
                    <StepperContent value={1} tabIndex={-1} onClick={onClick} aria-describedby='profile-description'>
                        Profile content
                    </StepperContent>
                    <StepperContent value={2} hidden={false} forceMount>
                        Secret content
                    </StepperContent>
                </StepperPanel>
            </Stepper>
        );
        expect(screen.getByRole('tablist', { name: 'Setup steps' })).toHaveAttribute('aria-orientation', 'vertical');
        expect(screen.getByTestId('indicator')).toHaveAttribute('aria-hidden', 'true');
        expect(screen.getByTestId('separator')).toHaveAttribute('aria-hidden', 'true');
        expect(screen.getByText('Profile')).toHaveAttribute('id', 'profile-title');
        expect(screen.getByText('Details')).toHaveAttribute('id', 'profile-description');
        expect(screen.getByLabelText('Step contents')).toHaveAttribute('data-slot', 'stepper-panel');
        const panel = screen.getByRole('tabpanel');
        expect(panel).toHaveAttribute('tabindex', '-1');
        expect(panel).toHaveAttribute('aria-describedby', 'profile-description');
        await userEvent.click(panel);
        expect(onClick).toHaveBeenCalledOnce();
        expect(screen.getByText('Secret content')).not.toBeVisible();
    });
});
