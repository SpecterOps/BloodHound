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

const Abuse: FC = () => {
    return (
        <>
            <Typography variant='body2'>
                No special actions are needed to abuse this, as the kerberos tickets created will have all SIDs in the
                object's SID history attribute added to them; however, if traversing a domain trust boundary, ensure
                that SID filtering is not enforced, as SID filtering will ignore any SIDs in the SID history portion of
                a kerberos ticket.
            </Typography>
            <Typography variant='body2'>
                By default, SID filtering is not enabled for all domain trust types.
            </Typography>
        </>
    );
};

export default Abuse;
