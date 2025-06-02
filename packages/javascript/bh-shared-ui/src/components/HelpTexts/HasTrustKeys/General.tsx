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

import { Typography } from '@mui/material';
import { FC } from 'react';
import { EdgeInfoProps } from '../index';

const General: FC<EdgeInfoProps> = ({ sourceName, targetName }) => {
    return (
        <>
            <Typography variant='body2'>
                The domain {sourceName} has the trust keys of the trust account {targetName}. The trust account exists as
                the domain has an outbound trust to the domain of the trust account.
            </Typography>
            <Typography variant='body2'>
                The trust keys can be leveraged to authenticate as the trust account and gain access to the trust
                account's domain.
            </Typography>
        </>
    );
};

export default General;
