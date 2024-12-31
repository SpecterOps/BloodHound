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
import NoDataDialog from '.';
import { UseQueryResult } from 'react-query';

const gettingStartedLinkText = 'Getting Started guide';
const fileIngestLinkText = 'start by uploading your data';
const mockUseAvailableDomainsQuery: UseQueryResult<any, Error> = {
    data: [],
    isLoading: false,
    isError: false,
    isSuccess: true,
    isIdle: false,
    isLoadingError: false,
    isRefetchError: false,
    status: 'success',
    failureCount: 0,
    dataUpdatedAt: 0,
    errorUpdatedAt: 0,
    errorUpdateCount: 0,
    isFetched: true,
    isFetchedAfterMount: true,
    isFetching: false,
    isPlaceholderData: false,
    isPreviousData: false,
    isRefetching: false,
    isStale: false,
    error: null,
    remove: vi.fn(),
    refetch: vi.fn(),
};

describe('NoDataDialog', () => {
    it('should render', () => {
        render(
            <NoDataDialog
                gettingStartedLink={<>{gettingStartedLinkText}</>}
                fileIngestLink={<>{fileIngestLinkText}</>}
                useAvailableDomainsQuery={mockUseAvailableDomainsQuery}
            />
        );

        expect(screen.getByText('No Data Available')).toBeInTheDocument();
        expect(screen.getByText(/Getting Started guide/)).toBeInTheDocument();
        expect(screen.getByText(/start by uploading your data/)).toBeInTheDocument();
    });
});
