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
                This edge is created during post-processing. It is created against all App Registrations and Service
                Principals within the same tenant when an Azure principal has one of the following Azure Active
                Directory roles:
            </Typography>

            <Typography variant='body2'>
                <List>
                    <ListItem>
                        <ListItemText>Hybrid Identity Administrator</ListItemText>
                    </ListItem>
                    <ListItem>
                        <ListItemText>Partner Tier1 Support</ListItemText>
                    </ListItem>
                    <ListItem>
                        <ListItemText>Partner Tier2 Support</ListItemText>
                    </ListItem>
                    <ListItem>
                        <ListItemText>Directory Synchronization Accounts</ListItemText>
                    </ListItem>
                </List>
            </Typography>

            <Typography variant='body2'>
                You will not see these privileges when auditing permissions against any of the mentioned objects when
                you use Microsoft tooling, including the Azure portal or any API.
            </Typography>
        </>
    );
};

export default General;
