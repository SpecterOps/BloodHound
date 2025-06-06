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
        <>
            <Typography variant='body2'>
                There are several forensic artifacts generated by the techniques described above. For instance, lateral
                movement via PsExec will generate 4697 events on the target system. If the target organization is
                collecting and analyzing those events, they may very easily detect lateral movement via PsExec.
            </Typography>

            <Typography variant='body2'>
                Additionally, an EDR product may detect your attempt to inject into lsass and alert a SOC analyst. There
                are many more opsec considerations to keep in mind when abusing administrator privileges. For more
                information, see the References tab.
            </Typography>
        </>
    );
};

export default Opsec;
