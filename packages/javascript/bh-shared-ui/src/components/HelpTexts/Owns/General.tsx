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
                {groupSpecialFormat(sourceType, sourceName)} ownership of the {typeFormat(targetType)} {targetName}.
            </Typography>

            <Typography variant='body2'>
                The owner of an object is implicitly granted the ability to modify object security descriptors, including the DACL,
                when the following conditions are met:
                <br />
                <ul>
                    <li>
                        The OWNER RIGHTS SID (S-1-3-4) is not explicitly granted privileges on the object
                    </li>
                    <Typography component={'pre'}>
                    {
                        "(Get-ACL -Path (\"AD:\" + \"CN=Object,DC=example,DC=com\")).Access | Where-Object { $_.IdentityReference -eq \"OWNER RIGHTS\" }"
                    }
                    </Typography>
                    <br />
                    OR
                    <br />
                    <br />
                    <li>
                        Implicit owner rights are not blocked
                    </li>
                </ul>
            </Typography>
            
            <Typography variant='body2'>
                Implicit owner rights are not blocked and are therefore abusable when the following conditions are met:

                <ul>
                    <li>
                        The domain's BlockOwnerImplicitRights setting is not in enforcement mode. This setting is defined in  
                        the 29th character in the domain's dSHeuristics attribute. When set to 0 or 2, implicit owner rights are not blocked.
                    </li>
                    <Typography component={'pre'}>
                    {
                        "$searcher = [adsisearcher]\"\"\n$searcher.SearchRoot = \"LDAP://CN=Directory Service,CN=Windows NT,CN=Services,CN=Configuration,DC=EXAMPLE,DC=LOCAL\"\n$searcher.SearchScope = [System.DirectoryServices.SearchScope]::Base\n$searcher.Filter = \"(objectClass=*)\"\n$searcher.PropertiesToLoad.Add(\"DSHeuristics\") | Out-Null\n$result = $searcher.FindOne()\nWrite-Output \"DSHeuristics: $($result.Properties['DSHeuristics'])\""
                    }
                    </Typography>                
                    <br />
                    AND EITHER:
                    <br />
                    <br />
                    <li>
                        The object is not a computer or derivative of a computer object (e.g., MSA, GMSA)
                    </li>
                    <br />
                    OR
                    <br />
                    <br />
                    <li>
                        The object is a computer or derivative of a computer object and the owner is a member of the Domain Admins or Enterprise Admins group (or is the SID of either group)
                    </li>
                </ul>
            </Typography>
        </>
    );
};

export default General;
