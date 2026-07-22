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

import { Typography } from 'doodle-ui';
import { FC } from 'react';
import { EdgeInfoProps } from '../index';
import { groupSpecialFormat } from '../utils';

const General: FC<EdgeInfoProps> = ({ sourceName, sourceType, targetType, targetName }) => {
    return (
        <>
            <Typography variant='body2'>
                {groupSpecialFormat(sourceType, sourceName)} has permission to modify the gPLink attribute of{' '}
                {targetType} {targetName}.
            </Typography>

            <Typography variant='body2'>
                Modifying an object's gPLink attribute can allow an attacker to link a malicious Group Policy Object
                (GPO) so it applies to affected users and computers. That GPO can force those objects to execute
                arbitrary commands, for example through an immediate scheduled task.
            </Typography>
            <Typography variant='body2'>
                For domain and OU objects, affected child users and computers include those contained directly within
                the domain or OU, as well as those in nested OUs. However, unless the GPO link is enforced, some users
                and computers may not be affected if GPO inheritance is blocked on a containing OU.
            </Typography>
            <Typography variant='body2'>
                For site objects, affected computers include the site's domain controllers, and also computers whose IP
                addresses fall within one of the site's subnets. If the site is the default site, affected computers
                also include computers that do not map to any other site. Affected users are those who sign in to the
                affected computers.
            </Typography>
        </>
    );
};

export default General;
