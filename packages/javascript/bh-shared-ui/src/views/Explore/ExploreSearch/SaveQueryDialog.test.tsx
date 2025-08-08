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

import userEvent from '@testing-library/user-event';
import { render, screen } from '../../../test-utils';
import SaveQueryDialog from './SaveQueryDialog';

describe('SaveQueryDialog', () => {
    it('should render a SaveQueryDialog', () => {
        const testOnSave = vitest.fn();
        const testOnClose = vitest.fn();
        const testOnUpdate = vitest.fn();
        const testSetSharedIds = vitest.fn();
        const testSetIsPublic = vitest.fn();

        const testIsLoading = false;
        const testError = undefined;
        const testCypherSearchState = {
            cypherQuery: '',
            performSearch: vitest.fn(),
            setCypherQuery: vitest.fn(),
        };

        render(
            <SaveQueryDialog
                open
                onSave={testOnSave}
                onClose={testOnClose}
                isLoading={testIsLoading}
                error={testError}
                cypherSearchState={testCypherSearchState}
                selectedQuery={undefined}
                sharedIds={[]}
                isPublic={false}
                onUpdate={testOnUpdate}
                setSharedIds={testSetSharedIds}
                setIsPublic={testSetIsPublic}
            />
        );

        expect(screen.getByText(/save query/i)).toBeInTheDocument();
        expect(screen.getByLabelText(/query name/i)).toBeInTheDocument();
        expect(screen.getByRole('button', { name: /cancel/i })).toBeInTheDocument();
        expect(screen.getByRole('button', { name: /save/i })).toBeInTheDocument();
    });

    it('should disable save button while input is empty', async () => {
        const user = userEvent.setup();

        const testOnSave = vitest.fn();
        const testOnClose = vitest.fn();
        const testOnUpdate = vitest.fn();
        const testSetSharedIds = vitest.fn();
        const testSetIsPublic = vitest.fn();

        const testIsLoading = false;
        const testError = undefined;
        const testCypherSearchState = {
            cypherQuery: '',
            performSearch: vitest.fn(),
            setCypherQuery: vitest.fn(),
        };

        const testQueryName = 'query name';

        render(
            <SaveQueryDialog
                open
                onSave={testOnSave}
                onClose={testOnClose}
                isLoading={testIsLoading}
                error={testError}
                cypherSearchState={testCypherSearchState}
                selectedQuery={undefined}
                sharedIds={[]}
                isPublic={false}
                onUpdate={testOnUpdate}
                setSharedIds={testSetSharedIds}
                setIsPublic={testSetIsPublic}
            />
        );

        expect(screen.getByRole('button', { name: /save/i })).toBeDisabled();

        await user.type(screen.getByLabelText(/query name/i), testQueryName);

        expect(screen.getByRole('button', { name: /save/i })).not.toBeDisabled();
    });

    it('should call onClose when cancel button is clicked', async () => {
        const user = userEvent.setup();
        const testOnSave = vitest.fn();
        const testOnClose = vitest.fn();
        const testOnUpdate = vitest.fn();
        const testSetSharedIds = vitest.fn();
        const testSetIsPublic = vitest.fn();

        const testIsLoading = false;
        const testError = undefined;
        const testCypherSearchState = {
            cypherQuery: '',
            performSearch: vitest.fn(),
            setCypherQuery: vitest.fn(),
        };

        render(
            <SaveQueryDialog
                open
                onSave={testOnSave}
                onClose={testOnClose}
                isLoading={testIsLoading}
                error={testError}
                cypherSearchState={testCypherSearchState}
                selectedQuery={undefined}
                sharedIds={[]}
                isPublic={false}
                onUpdate={testOnUpdate}
                setSharedIds={testSetSharedIds}
                setIsPublic={testSetIsPublic}
            />
        );

        await user.click(screen.getByRole('button', { name: /cancel/i }));

        expect(testOnClose).toHaveBeenCalled();
    });

    it('should call onSave when save button is clicked', async () => {
        const user = userEvent.setup();
        const testOnSave = vitest.fn();
        const testOnClose = vitest.fn();
        const testOnUpdate = vitest.fn();
        const testSetSharedIds = vitest.fn();
        const testSetIsPublic = vitest.fn();

        const testIsLoading = false;
        const testError = undefined;
        const testCypherSearchState = {
            cypherQuery: '',
            performSearch: vitest.fn(),
            setCypherQuery: vitest.fn(),
        };
        const testQueryName = 'query name';

        render(
            <SaveQueryDialog
                open
                onSave={testOnSave}
                onClose={testOnClose}
                isLoading={testIsLoading}
                error={testError}
                cypherSearchState={testCypherSearchState}
                selectedQuery={undefined}
                sharedIds={[]}
                isPublic={false}
                onUpdate={testOnUpdate}
                setSharedIds={testSetSharedIds}
                setIsPublic={testSetIsPublic}
            />
        );

        await user.type(screen.getByLabelText(/query name/i), testQueryName);

        await user.click(screen.getByRole('button', { name: /save/i }));

        expect(testOnSave).toHaveBeenCalled();
    });

    it('should display an error when error prop is truthy', () => {
        const testOnSave = vitest.fn();
        const testOnClose = vitest.fn();
        const testOnUpdate = vitest.fn();
        const testSetSharedIds = vitest.fn();
        const testSetIsPublic = vitest.fn();

        const testIsLoading = false;
        const testError = true;
        const testCypherSearchState = {
            cypherQuery: '',
            performSearch: vitest.fn(),
            setCypherQuery: vitest.fn(),
        };

        render(
            <SaveQueryDialog
                open
                onSave={testOnSave}
                onClose={testOnClose}
                isLoading={testIsLoading}
                error={testError}
                cypherSearchState={testCypherSearchState}
                selectedQuery={undefined}
                sharedIds={[]}
                isPublic={false}
                onUpdate={testOnUpdate}
                setSharedIds={testSetSharedIds}
                setIsPublic={testSetIsPublic}
            />
        );

        expect(
            screen.getByText(/an error ocurred while attempting to save this query. please try again./i)
        ).toBeInTheDocument();
    });
});
