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

const Opsec: FC = () => {
    return (
        <Typography variant='body2'>
            The present attack vector relies on the execution of a malicious Group Policy Object. In case some objects
            in the target Organizational Unit are unable to apply said Group Policy Object (for instance, because these
            objects cannot reach the attacker's machine in the internal network), events related to failed GPO
            application will be created. Furthermore, the execution of this attack will result in the modification of
            the gPLink property of the target Organizational Unit. The property should be reset to its original value
            after attack execution to avoid detection and ensure the OU child items can apply their legitimate Group
            Policy Objects again.
        </Typography>
    );
};

export default Opsec;
