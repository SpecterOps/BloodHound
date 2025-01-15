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

import NoDataDialog from '.';
import { render, screen } from '../../test-utils';

const gettingStartedLinkText = 'Getting Started guide';
const fileIngestLinkText = 'start by uploading your data';

describe('NoDataDialog', () => {
    it('should render', () => {
        render(
            <NoDataDialog
                gettingStartedLink={<>{gettingStartedLinkText}</>}
                fileIngestLink={<>{fileIngestLinkText}</>}
                open={true}
            />
        );

        expect(screen.getByText('No Data Available')).toBeInTheDocument();
        expect(screen.getByText(/Getting Started guide/)).toBeInTheDocument();
        expect(screen.getByText(/start by uploading your data/)).toBeInTheDocument();
    });
    it('should not render when data is present', () => {
        render(
            <NoDataDialog
                gettingStartedLink={<>{gettingStartedLinkText}</>}
                fileIngestLink={<>{fileIngestLinkText}</>}
                open={false}
            />
        );

        expect(screen.queryByText('No Data Available')).not.toBeInTheDocument();
        expect(screen.queryByText(/Getting Started guide/)).not.toBeInTheDocument();
        expect(screen.queryByText(/start by uploading your data/)).not.toBeInTheDocument();
    });
});
