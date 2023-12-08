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

import { render, screen } from '../../../test-utils';
import { Table, TableBody } from '@mui/material';
import LoadContainer from '.';

describe('LoadContainer', () => {
    it('should display locally formatted numbers if they are larger than 999', () => {
        render(
            <Table>
                <TableBody>
                    <LoadContainer icon='globe' display='Test' value={1000} />
                </TableBody>
            </Table>
        );
        expect(screen.getByText('1,000')).toBeInTheDocument();
    });
});
