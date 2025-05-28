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

import { render, screen } from '../../test-utils';
import GenericErrorBoundaryFallback from './GenericErrorBoundaryFallback';

describe('GenericErrorBoundaryFallback', () => {
    beforeEach(async () => {
        render(<GenericErrorBoundaryFallback />);
    });

    it('should display', () => {
        const elem = screen.getByRole('alert');

        expect(elem).toBeInTheDocument();
        expect(elem).toHaveTextContent('An unexpected error has occurred.');
        expect(screen.getByTestId('ErrorOutlineIcon')).toBeInTheDocument();
    });

    it('should be aligned to right of screen', () => {
        const elem = screen.getByTestId('error-boundary');
        const styles = getComputedStyle(elem);
        expect(styles.display).toEqual('flex');
        expect(styles.justifyContent).toEqual('flex-end');
    });
});
