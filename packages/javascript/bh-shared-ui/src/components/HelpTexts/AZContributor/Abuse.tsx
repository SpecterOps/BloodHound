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

import { FC } from 'react';
import { Link, Typography } from '@mui/material';

const Abuse: FC = () => {
    return (
        <>
            <Typography variant='body2'>This depends on what the target object is:</Typography>
            <Typography variant='body2'>
                <Typography fontWeight='fontWeightBold'>Key Vault</Typography>: You can read secrets and alter access
                policies (grant yourself access to read secrets)
            </Typography>
            <Typography variant='body2'>
                <Typography fontWeight='fontWeightBold'>Automation Account</Typography>: You can create a new runbook
                that runs as the Automation Account, and edit existing runbooks. Runbooks can be used to authenticate as
                the Automation Account and abuse privileges held by the Automation Account. If the Automation Account is
                using a 'RunAs' account, you can gather the certificate used to login and impersonate that account.
            </Typography>
            <Typography variant='body2'>
                <Typography fontWeight='fontWeightBold'>Virtual Machine</Typography>: Run SYSTEM commands on the VM
            </Typography>
            <Typography variant='body2'>
                <Typography fontWeight='fontWeightBold'>Resource Group</Typography>: NOT abusable, and not collected by
                AzureHound
            </Typography>

            <Typography variant='body2'>Via PowerZure:</Typography>
            <Link
                target='_blank'
                rel='noopener'
                href='https://powerzure.readthedocs.io/en/latest/Functions/operational.html#get-azurekeyvaultcontent'>
                Get-AzureKeyVaultContent
            </Link>
            <br />
            <Link
                target='_blank'
                rel='noopener'
                href='https://powerzure.readthedocs.io/en/latest/Functions/operational.html#export-azurekeyvaultcontent'>
                Export-AzureKeyVaultContent
            </Link>
            <br />
            <Link
                target='_blank'
                rel='noopener'
                href='https://powerzure.readthedocs.io/en/latest/Functions/operational.html#get-azurerunascertificate'>
                Get-AzureRunAsCertificate
            </Link>
            <br />
            <Link
                target='_blank'
                rel='noopener'
                href='https://powerzure.readthedocs.io/en/latest/Functions/operational.html#get-azurerunbookcontent'>
                Get-AzureRunbookContent
            </Link>
            <br />
            <Link target='_blank' rel='noopener' href='http://Invoke-AzureRunCommand'>
                Invoke-AzureRunCommand
            </Link>
            <br />
            <Link target='_blank' rel='noopener' href='http://Invoke-AzureRunMSBuild'>
                Invoke-AzureRunMSBuild
            </Link>
            <br />
            <Link
                target='_blank'
                rel='noopener'
                href='https://powerzure.readthedocs.io/en/latest/Functions/operational.html#invoke-azurerunprogram'>
                Invoke-AzureRunProgram
            </Link>
        </>
    );
};

export default Abuse;
