// Copyright 2024 Specter Ops, Inc.
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
        <Typography variant='body2'>
            When the affected certificate authority issues the certificate to the attacker, it will retain a local copy
            of that certificate in its issued certificates store. Defenders may analyze those issued certificates to
            identify illegitimately issued certificates and identify the principal that requested the certificate, as
            well as the target identity the attacker is attempting to impersonate.
        </Typography>
    );
};

export default Opsec;
