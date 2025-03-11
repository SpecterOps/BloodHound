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

import { Typography } from '@mui/material';
import { FC } from 'react';

const Abuse: FC = () => {
    return (
        <>
            <Typography variant='body2'>
                This edge should be taken into consideration when abusing control of an app. Apps authenticate with
                service principals to the tenant, so if you have control of an app, what you are abusing is that control
                plus the fact that the app runs as a privileged service principal
            </Typography>

            <Typography variant={'body2'}>
                <ol className='list-inside list-decimal flex flex-col gap-2'>
                    <li>
                        Use the{' '}
                        <a
                            className='text-secondary dark:text-secondary-variant-2'
                            href={
                                'https://learn.microsoft.com/en-us/graph/api/serviceprincipal-addpassword?view=graph-rest-1.0&tabs=http'
                            }>
                            Microsoft Graph API
                        </a>{' '}
                        to add a new client secret to the Azure Application.
                    </li>
                    <li>
                        Use the{' '}
                        <a
                            className='text-secondary dark:text-secondary-variant-2'
                            href={'https://learn.microsoft.com/en-us/cli/azure/'}>
                            Azure CLI
                        </a>{' '}
                        to authenticate as the Service Principal.
                    </li>
                    <li>Proceed to access additional Azure resources under the control of the Service Principal.</li>
                </ol>
            </Typography>
        </>
    );
};

export default Abuse;
