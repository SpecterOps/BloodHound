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

import { FileUploadDialog } from 'bh-shared-ui';
import { useState } from 'react';
import { Link as RouterLink } from 'react-router-dom';

type NoDataFileUploadDialogWithLinksProps = {
    open: boolean;
};

const linkStyles = 'text-secondary dark:text-secondary-variant-2 hover:underline';

export const NoDataFileUploadDialogWithLinks: React.FC<NoDataFileUploadDialogWithLinksProps> = ({ open }) => {
    const [showDialog, shouldShowDialog] = useState(open);

    if (!showDialog) return null;

    return (
        <FileUploadDialog
            open={showDialog}
            onClose={() => shouldShowDialog(false)}
            headerText='Upload Data to Start Mapping Your Environment'
            description={
                <div>
                    <p className='pb-3'>
                        Easily upload data by dragging and dropping files anywhere in the interface, or use the upload
                        button in the main navigation.
                    </p>
                    <p className='pb-3'>
                        If you&apos;re just exploring, you can use the{' '}
                        <a
                            className={linkStyles}
                            href='https://bloodhound.specterops.io/get-started/quickstart/ce-ingest-sample-data'
                            target='_blank'
                            rel='noreferrer'>
                            sample dataset
                        </a>{' '}
                        to get a quick sense of how the platform works.
                    </p>
                    <p className='pb-3'>
                        To get started with collecting data,{' '}
                        <RouterLink className={linkStyles} to='/download-collectors'>
                            download a collector
                        </RouterLink>{' '}
                        then{' '}
                        <RouterLink className={linkStyles} to='/administration/clients'>
                            schedule your collector clients
                        </RouterLink>
                        .
                    </p>
                    <p className='pb-3'>
                        If you&apos;re having any difficulty, we have a{' '}
                        <a
                            className={linkStyles}
                            href='https://bloodhound.specterops.io/collect-data/ce-collection/overview'
                            target='_blank'
                            rel='noreferrer'>
                            Getting Started Guide
                        </a>
                    </p>
                </div>
            }
        />
    );
};

// export const NoDataFileUploadDialogWithLinks: React.FC<NoDataFileUploadDialogWithLinksProps> = ({ open }) => {
//     return (
//         <NoDataDialog open={open}>
//             To explore your environment, start by uploading your data on the{' '}
//             <Link {...fileIngestLinkProps}>file ingest</Link> page.
//             <br className='mb-4' />
//             Need help? Check out the <a {...gettingStartedLinkProps}>Getting Started</a> guide for instructions.
//             <br className='mb-4' />
//             If you want to test BloodHound with sample data, you may download some from our{' '}
//             <a {...sampleCollectionLinkProps}>Sample Collection</a> GitHub page.
//         </NoDataDialog>
//     );
// };
