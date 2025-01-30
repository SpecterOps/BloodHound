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
import { NoDataDialog } from 'bh-shared-ui';
import { Link } from 'react-router-dom';
import { ROUTE_ADMINISTRATION_FILE_INGEST } from 'src/routes/constants';

type NoDataDialogWithLinksProps = {
    open: boolean;
};

const linkStyles = 'text-secondary dark:text-secondary-variant-2 hover:underline';

const fileIngestLinkProps = {
    className: linkStyles,
    to: ROUTE_ADMINISTRATION_FILE_INGEST,
};

const gettingStartedLinkProps = {
    className: linkStyles,
    target: '_blank',
    rel: 'noreferrer',
    href: 'https://support.bloodhoundenterprise.io/hc/en-us/sections/17274904083483-BloodHound-CE-Collection',
};

const sampleCollectionLinkProps = {
    className: linkStyles,
    target: '_blank',
    rel: 'noreferrer',
    href: 'https://github.com/SpecterOps/BloodHound/wiki/Example-Data',
};

export const NoDataDialogWithLinks: React.FC<NoDataDialogWithLinksProps> = ({ open }) => {
    return (
        <NoDataDialog open={open}>
            To explore your environment, <Link {...fileIngestLinkProps}>start by uploading your data</Link> on the file
            ingest page.
            <br className='mb-4' />
            Need help? Check out the <a {...gettingStartedLinkProps}>Getting Started guide</a> for instructions.
            <br className='mb-4' />
            If you want to test BloodHound with sample data, you may download some from our{' '}
            <a {...sampleCollectionLinkProps}>Sample Collection</a> GitHub page.
        </NoDataDialog>
    );
};
