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
                This edge is created during post-processing. It is created against all Azure App Registrations and
                Service Principals when a Service Principal has one of the following MS Graph app roles:
            </Typography>

            <Typography variant='body2'>
                <List>
                    <ListItem>
                        <ListItemText>Application.ReadWrite.All</ListItemText>
                    </ListItem>
                    <ListItem>
                        <ListItemText>RoleManagement.ReadWrite.Directory</ListItemText>
                    </ListItem>
                </List>
            </Typography>

            <Typography variant='body2'>
                You will not see this privilege when using just the Azure portal or any other Microsoft tooling. If you
                audit the roles and administrators affecting any particular Azure App or Service Principal, you will not
                see that the Service Principal can add secrets to the object, but it indeed can because of the parallel
                access management system used by MS Graph.
            </Typography>
        </>
    );
};

export default General;
