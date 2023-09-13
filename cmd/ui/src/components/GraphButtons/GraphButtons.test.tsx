// Copyright 2023 Specter Ops, Inc.
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

import { render, screen } from 'src/test-utils';
import GraphButtons from 'src/components/GraphButtons';
import { SigmaContainer } from '@react-sigma/core';
import userEvent from '@testing-library/user-event';

describe('GraphLayoutButtons', () => {
    const user = userEvent.setup();

    it('should render', () => {
        render(
            <SigmaContainer>
                <GraphButtons options={{ standard: false, sequential: true }} />
            </SigmaContainer>
        );

        const layoutButton = screen.getByRole('button', { name: /layout/i });
        expect(layoutButton).toBeInTheDocument();

        const exportButton = screen.getByRole('button', { name: /export/i });
        expect(exportButton).toBeInTheDocument();
    });

    it('should render only the layout button options specified', async () => {
        render(
            <SigmaContainer>
                <GraphButtons options={{ standard: false, sequential: true }} />
            </SigmaContainer>
        );

        const layoutButton = screen.getByRole('button', { name: /layout/i });
        expect(layoutButton).toBeInTheDocument();

        await user.click(layoutButton);

        const sequentialMenuItem = screen.getByRole('menuitem', { name: /sequential/i });
        expect(sequentialMenuItem).toBeInTheDocument();

        const standardMenuItem = screen.queryByRole('menuitem', { name: /standard/i });
        expect(standardMenuItem).not.toBeInTheDocument();
    });

    it('interacting with any menu item closes the menu', async () => {
        render(
            <SigmaContainer>
                <GraphButtons options={{ standard: false, sequential: true }} />
            </SigmaContainer>
        );

        const layoutButton = screen.getByRole('button', { name: /layout/i });
        expect(layoutButton).toBeInTheDocument();

        await user.click(layoutButton);

        const menuItem = screen.getByRole('menuitem', { name: /sequential/i });
        expect(menuItem).toBeInTheDocument();

        await userEvent.click(menuItem);
        expect(menuItem).not.toBeInTheDocument();
    });

    it('export action is disabled if the canvas is empty', async () => {
        render(
            <SigmaContainer>
                <GraphButtons />
            </SigmaContainer>
        );

        const exportButton = screen.getByRole('button', { name: /export/i });
        expect(exportButton).toBeInTheDocument();

        await user.click(exportButton);

        const jsonMenuItem = screen.getByRole('menuitem', { name: /json/i });
        expect(jsonMenuItem).toHaveAttribute('aria-disabled');
    });

    it('export action is enabled if the there is graph data saved in redux', async () => {
        render(
            <SigmaContainer>
                <GraphButtons />
            </SigmaContainer>,
            {
                initialState: {
                    explore: {
                        export: {
                            hello: 'world',
                        },
                    },
                },
            }
        );

        const exportButton = screen.getByRole('button', { name: /export/i });
        expect(exportButton).toBeInTheDocument();

        await user.click(exportButton);

        const jsonMenuItem = screen.getByRole('menuitem', { name: /json/i });
        expect(jsonMenuItem).not.toHaveAttribute('aria-disabled');
    });
});
