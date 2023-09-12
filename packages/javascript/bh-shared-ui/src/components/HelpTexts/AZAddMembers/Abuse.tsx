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

const Abuse: FC = () => {
    return (
        <>
            <Typography variant='body2'>Via the Azure portal:</Typography>
            <List>
                <ListItem>
                    <ListItemText>
                        Find the group in your tenant (Azure Active Directory -&gt; Groups -&gt; Find Group in list)
                    </ListItemText>
                </ListItem>
                <ListItem>
                    <ListItemText>Click the group from the list</ListItemText>
                </ListItem>
                <ListItem>
                    <ListItemText>In the left pane, click "Members"</ListItemText>
                </ListItem>
                <ListItem>
                    <ListItemText>At the top, click "Add members"</ListItemText>
                </ListItem>
                <ListItem>
                    <ListItemText>
                        Find the principals you want to add to the group and click them, then click "select" at the
                        bottom
                    </ListItemText>
                </ListItem>
                <ListItem>
                    <ListItemText>
                        You should see a message in the top right saying "Member successfully added"
                    </ListItemText>
                </ListItem>
            </List>
            <Typography variant='body2'>Via PowerZure: Add-AzureADGroup -User [UPN] -Group [Group name]</Typography>
        </>
    );
};

export default Abuse;
