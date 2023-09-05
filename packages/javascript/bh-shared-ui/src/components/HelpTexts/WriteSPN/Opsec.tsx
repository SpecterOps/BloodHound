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
import { Typography } from '@mui/material';

const Opsec: FC = () => {
    return (
        <>
            <Typography variant='body2'>
                Executing this abuse with the net binary will require command line execution. If your target
                organization has command line logging enabled, this is a detection opportunity for their analysts.
            </Typography>
            <Typography variant='body2'>
                Regardless of what execution procedure you use, this action will generate a 4728 event on the domain
                controller that handled the request. This event may be centrally collected and analyzed by security
                analysts, especially for groups that are obviously very high privilege groups (i.e.: Domain Admins).
                Also be mindful that Powershell 5 introduced several key security features such as script block logging
                and AMSI that provide security analysts another detection opportunity.
            </Typography>
            <Typography variant='body2'>
                You may be able to completely evade those features by downgrading to PowerShell v2.
            </Typography>
        </>
    );
};

export default Opsec;
