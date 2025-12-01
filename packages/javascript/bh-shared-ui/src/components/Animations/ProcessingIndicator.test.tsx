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
import ProcessingIndicator from './ProcessingIndicator';

describe('ProcessingIndicator', () => {
    it('renders the title', () => {
        render(<ProcessingIndicator title='Analyzing' />);
        expect(screen.getByText('Analyzing')).toBeInTheDocument();
    });

    it('renders three animated dots', () => {
        render(<ProcessingIndicator title='Loading' />);

        const dots = screen.getAllByText('.');
        expect(dots).toHaveLength(3);

        dots.forEach((dot) => {
            expect(dot).toHaveClass('animate-pulse');
        });

        // Check animation delays
        expect(dots[0]).toHaveStyle({ 'animation-delay': undefined });
        expect(dots[1]).toHaveStyle({ 'animation-delay': '0.2s' });
        expect(dots[2]).toHaveStyle({ 'animation-delay': '0.4s' });
    });

    it('applies animation class to the title', () => {
        render(<ProcessingIndicator title='Analyzing' />);
        const title = screen.getByText('Analyzing');
        expect(title).toHaveClass('animate-pulse');
    });
});
