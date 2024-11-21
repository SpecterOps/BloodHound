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
import { EdgeInfoProps } from '../index';
import { groupSpecialFormat, typeFormat } from '../utils';

const General: FC<EdgeInfoProps> = ({ sourceName, sourceType, targetName, targetType }) => {
    return (
        <>
            <Typography variant='body2'>
                {groupSpecialFormat(sourceType, sourceName)} the ability to modify the owner of the{' '}
                {typeFormat(targetType)} {targetName}.
            </Typography>

            <Typography variant='body2'>
                Implicit owner rights are not blocked and are therefore abusable via change in ownership when the following 
                conditions are met:
                <ul>
                    <li>
                        Inheritance is not configured for any privileges explicitly granted to the OWNER RIGHTS SID (S-1-3-4).
                        Non-inherited privileges granted to OWNER RIGHTS are removed when the owner is changed, allowing the 
                        new owner to have the full set of implicit owner rights.
                    </li>
                    <li>
                        The domain's BlockOwnerImplicitRights setting is not in enforcement mode. This setting is defined in  
                        the 29th character in the domain's dSHeuristics attribute. When set to 0 or 2, implicit owner rights are not blocked.
                    </li>
                    <Typography component={'pre'}>
                    {
                        "$searcher = [adsisearcher]\"\"\n$searcher.SearchRoot = \"LDAP://CN=Directory Service,CN=Windows NT,CN=Services,CN=Configuration,DC=EXAMPLE,DC=LOCAL\"\n$searcher.SearchScope = [System.DirectoryServices.SearchScope]::Base\n$searcher.Filter = \"(objectClass=*)\"\n$searcher.PropertiesToLoad.Add(\"DSHeuristics\") | Out-Null\n$result = $searcher.FindOne()\nWrite-Output \"DSHeuristics: $($result.Properties['DSHeuristics'])\""
                    }
                    </Typography>
                    <li>
                        The object is not a computer or a derivative of a computer object (e.g., MSA, GMSA).
                    </li>
                </ul>
            </Typography>
        </>
    );
};

export default General;
