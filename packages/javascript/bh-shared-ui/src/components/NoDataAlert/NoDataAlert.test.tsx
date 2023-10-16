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

import { render, screen } from '../../test-utils';
import NoDataAlert from '.';

const dataCollectionLinkText = 'data collection';
const fileIngestLinkText = 'file ingest';

describe('NoDataAlert', () => {
    it('should render', () => {
        render(<NoDataAlert dataCollectionLink={<>{dataCollectionLinkText}</>} />);

        expect(screen.getByText('No Data Available')).toBeInTheDocument();
        expect(screen.getByText(/It appears that no data has been uploaded yet./)).toBeInTheDocument();
        expect(screen.getByText(/data collection/)).toBeInTheDocument();

        //This text only displays if file ingest link prop is passed
        expect(screen.queryByText(/file ingest/)).toBeNull();
    });

    it('should show the file ingest text if the prop is passed', () => {
        render(
            <NoDataAlert
                dataCollectionLink={<>{dataCollectionLinkText}</>}
                fileIngestLink={<>{fileIngestLinkText}</>}
            />
        );

        expect(screen.getByText('No Data Available')).toBeInTheDocument();
        expect(screen.getByText(/It appears that no data has been uploaded yet./)).toBeInTheDocument();
        expect(screen.getByText(/data collection/)).toBeInTheDocument();
        expect(screen.getByText(/file ingest/)).toBeInTheDocument();
    });
});
