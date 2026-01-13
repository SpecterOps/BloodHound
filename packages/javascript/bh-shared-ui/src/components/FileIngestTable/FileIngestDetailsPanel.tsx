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

import { Card, CardContent } from '@bloodhoundenterprise/doodleui';
import type { FileIngestCompletedTask, FileIngestJob } from 'js-client-library';
import { useFileUploadQuery } from '../../hooks';
import { DetailsAccordion } from '../DetailsAccordion';
import { StatusIndicator } from '../StatusIndicator';

/** Header for an individual file result */
const FileHeader: React.FC<FileIngestCompletedTask> = ({ file_name, errors }) => {
    const isSuccess = errors.length === 0;
    return (
        <div className='flex-grow text-left text-xs font-normal ml-4'>
            <div className='text-base font-bold'>{file_name}</div>
            <StatusIndicator status={isSuccess ? 'good' : 'bad'} label={isSuccess ? 'Success' : 'Failure'} />
        </div>
    );
};

/** Only displays content if ingest had errors or warnings  */
const FileContent: React.FC<FileIngestCompletedTask> = (ingest) =>
    ingest.errors.length > 0 || ingest.warnings.length > 0 ? <FileErrors {...ingest} /> : null;

/** Displays file ingest errors and warnings */
const FileErrors: React.FC<FileIngestCompletedTask> = ({ errors, warnings }) => {
    const hasErrors = errors.length > 0;
    const hasWarnings = warnings.length > 0;
    return <div className='p-3'>
        <div className='p-3 bg-neutral-3'>

            { hasErrors &&  (

            
            <div className='font-bold mb-2'>Error Message(s):</div>
            {errors.map((error, index) => (
                <div className='[&:not(:last-child)]:mb-2' key={index}>
                    {error}
                </div>
            ))}
        )

        }

        </div>
    </div>
};

const isErrorFree = (ingest: FileIngestCompletedTask | null) => ingest?.errors.length === 0;

/** Displays the ingest ID */
const IngestHeader: React.FC<FileIngestJob> = ({ id }) => <div className='ml-4'>ID {id}</div>;

/** Displays a list of all files in the ingest */
const IngestContent: React.FC<FileIngestJob> = (ingest) => {
    const { data, isSuccess } = useFileUploadQuery(ingest.id);

    const items = isSuccess ? data.data : [];

    return (
        <div className='max-h-[calc(100vh-16rem)] overflow-y-auto'>
            <DetailsAccordion Content={FileContent} Header={FileHeader} itemDisabled={isErrorFree} items={items} />
        </div>
    );
};

/** Displays a message to click an ingest ID */
export const NoIngest = () => (
    <Card>
        <CardContent className='px-4'>Click on the Ingest ID to reveal further information.</CardContent>
    </Card>
);

type Props = {
    ingest?: FileIngestJob;
};

/** Displays details for the selected ingest */
export const FileIngestDetailsPanel = ({ ingest }: Props) => {
    return (
        <DetailsAccordion
            accent
            Content={IngestContent}
            Empty={NoIngest}
            Header={IngestHeader}
            items={ingest}
            openIndex={0}
        />
    );
};
