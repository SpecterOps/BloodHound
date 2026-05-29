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

import { render, screen } from '../test-utils';
import { StatusIndicator } from './StatusIndicator';

describe('StatusIndicator', () => {
    it('displays a status indicator', () => {
        const { container } = render(<StatusIndicator status='good' />);
        expect(container.querySelector('circle')).toHaveClass('fill-[#BCD3A8]');
    });

    it('displays a status label', async () => {
        const { container } = render(<StatusIndicator status='bad' label='Bad' />);
        expect(container.querySelector('circle')).toHaveClass('fill-[#D9442E]');
        expect(container.querySelector('div span')?.childNodes).toHaveLength(2);
        const badStatus = await screen.findByText('Bad', {});
        expect(badStatus).toBeInTheDocument();
    });

    it('renders without label when label is empty string', () => {
        const { container } = render(<StatusIndicator status='good' label='' />);
        expect(container.querySelector('div span')?.childNodes).toHaveLength(1);
    });

    it('displays a pulsing indicator', async () => {
        const { container } = render(<StatusIndicator status='pending' pulse />);
        expect(container.querySelector('circle')).toHaveClass('fill-[#5CC3AD] animate-pulse');
    });

    it('renders without pulse animation by default', () => {
        const { container } = render(<StatusIndicator status='pending' />);
        expect(container.querySelector('circle')).not.toHaveClass('animate-pulse');
    });
});
