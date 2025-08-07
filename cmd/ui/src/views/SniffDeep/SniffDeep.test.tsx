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
import SniffDeep from './SniffDeep';

describe('SniffDeep', () => {
    it('should render the page title', () => {
        const screen = render(<SniffDeep />);
        expect(screen.getByText('Sniff Deep')).toBeInTheDocument();
    });

    it('should render the placeholder content', () => {
        const screen = render(<SniffDeep />);
        expect(screen.getByText('This is an empty tab ready for your content.')).toBeInTheDocument();
    });

    it('should have the correct test id', () => {
        const screen = render(<SniffDeep />);
        expect(screen.getByTestId('sniff-deep-page')).toBeInTheDocument();
    });
});
