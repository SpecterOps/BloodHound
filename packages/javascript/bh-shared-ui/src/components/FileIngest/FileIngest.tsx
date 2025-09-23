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

import { Typography } from '@mui/material';
import { FC } from 'react';

import DocumentationLinks from '../DocumentationLinks';
import FeatureFlag from '../FeatureFlag';
import { FileIngestTable } from '../FileIngestTable';
import LegacyFileIngestTable from '../LegacyFileIngestTable/LegacyFileIngestTable';
import LoadingOverlay from '../LoadingOverlay';
import PageWithTitle from '../PageWithTitle';

const FileIngest: FC = () => {
    return (
        <PageWithTitle
            title='File Ingest'
            data-testid='manual-file-ingest'
            pageDescription={
                <Typography variant='body2'>
                    Upload data from SharpHound or AzureHound offline collectors. Check out our{' '}
                    {DocumentationLinks.fileIngestLink} documentation for more information.
                </Typography>
            }>
            <FeatureFlag
                flagKey='open_graph_phase_2'
                loadingFallback={<LoadingOverlay loading />}
                enabled={<FileIngestTable />}
                disabled={<LegacyFileIngestTable />}
            />
        </PageWithTitle>
    );
};

export default FileIngest;
