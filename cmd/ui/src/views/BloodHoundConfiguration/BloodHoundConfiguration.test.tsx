// Copyright 2024 Specter Ops, Inc.
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
import BloodHoundConfiguration from './BloodHoundConfiguration';

describe('BloodHoundConfiguration', () => {
    beforeEach(() => {
        render(<BloodHoundConfiguration />);
    });

    it('renders with a title', () => {
        const title = screen.getByRole('heading', { name: /BloodHound Configuration/i });
        expect(title).toBeInTheDocument();
    });

    it('renders a link to the documentation', () => {
        const link = screen.getByRole('link');
        expect(link).toHaveTextContent('documentation');
    });

    it('renders citrix config controls', () => {
        const title = screen.getByRole('heading', { name: /Citrix RDP Support/i });
        expect(title).toBeInTheDocument();
    });
});
