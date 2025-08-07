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

import { render } from 'src/test-utils';
import { vi } from 'vitest';
import SniffDeep from './SniffDeep';

// Mock the API client
vi.mock('bh-shared-ui', async () => {
    const actual = await vi.importActual('bh-shared-ui');
    return {
        ...actual,
        apiClient: {
            cypherSearch: vi.fn().mockResolvedValue({
                data: {
                    data: {
                        nodes: {
                            'node1': { kind: 'User', label: 'Test User', properties: { name: 'Test User' } },
                            'node2': { kind: 'Domain', label: 'Test Domain', properties: { name: 'Test Domain' } }
                        }
                    }
                }
            })
        }
    };
});

describe('SniffDeep', () => {
    it('should render the page title', () => {
        const screen = render(<SniffDeep />);
        expect(screen.getByText('Sniff Deep Analysis')).toBeInTheDocument();
    });

    it('should render the description', () => {
        const screen = render(<SniffDeep />);
        expect(screen.getByText('Execute Cypher query to find paths with DCSync-like privileges')).toBeInTheDocument();
    });

    it('should have the correct test id', () => {
        const screen = render(<SniffDeep />);
        expect(screen.getByTestId('sniff-deep-page')).toBeInTheDocument();
    });

    it('should render the play button', () => {
        const screen = render(<SniffDeep />);
        expect(screen.getByTestId('sniff-deep-play-button')).toBeInTheDocument();
        expect(screen.getByText('Start Sniffing')).toBeInTheDocument();
    });
});
