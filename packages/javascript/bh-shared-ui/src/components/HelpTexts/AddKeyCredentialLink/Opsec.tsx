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
                Executing the attack will generate a 5136 (A directory object was modified) event at the domain
                controller if an appropriate SACL is in place on the target object.
            </Typography>

            <Typography variant='body2'>
                If PKINIT is not common in the environment, a 4768 (Kerberos authentication ticket (TGT) was requested)
                ticket can also expose the attacker.
            </Typography>
        </>
    );
};

export default Opsec;
