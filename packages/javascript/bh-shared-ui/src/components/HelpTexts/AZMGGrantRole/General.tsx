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
import { Typography, List, ListItem, ListItemText } from '@mui/material';

const General: FC = () => {
    return (
        <>
            <Typography variant='body2'>
                This edge is created during post-processing. It is created against all AzureAD admin roles when a
                Service Principal has the following MS Graph app role assignment:
            </Typography>

            <Typography variant='body2'>
                <List>
                    <ListItem>
                        <ListItemText>RoleManagement.ReadWrite.Directory</ListItemText>
                    </ListItem>
                </List>
            </Typography>

            <Typography variant='body2'>
                This privilege allows the Service Principal to promote itself or any other principal to any AzureAD
                admin role, including Global Administrator.
            </Typography>
        </>
    );
};

export default General;
