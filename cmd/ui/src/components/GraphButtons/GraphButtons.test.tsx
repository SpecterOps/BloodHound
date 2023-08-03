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

describe('GraphLayoutButtons', () => {
    it('should render only the button options specified', () => {
        render(
            <SigmaContainer>
                <GraphButtons options={{ standard: false, sequential: true }} />
            </SigmaContainer>
        );

        expect(screen.getByText('sequential')).toBeInTheDocument();
        expect(screen.queryByText('standard')).not.toBeInTheDocument();
        expect(screen.queryByText('expand all')).not.toBeInTheDocument();
        expect(screen.queryByText('collapse all')).not.toBeInTheDocument();
    });
});
