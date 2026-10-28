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
import { render, screen } from '@testing-library/react';
import { Pagination } from './Pagination';

describe('Pagination', () => {
    it('shows an empty result range and disables navigation when there are no rows', () => {
        render(<Pagination page={0} rowsPerPage={25} count={0} onPageChange={vi.fn()} onRowsPerPageChange={vi.fn()} />);

        expect(screen.getByText('0 - 0 of 0')).toBeInTheDocument();
        expect(screen.getByRole('button', { name: 'Go to previous page' })).toBeDisabled();
        expect(screen.getByRole('button', { name: 'Go to next page' })).toBeDisabled();
    });
});
