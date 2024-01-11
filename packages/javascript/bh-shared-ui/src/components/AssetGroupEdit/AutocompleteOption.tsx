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

import { Box, Fade, ListItem, Tooltip, Typography } from '@mui/material';
import { FC, HTMLAttributes } from 'react';
import NodeIcon from '../NodeIcon';

const AutocompleteOption: FC<{
    props: HTMLAttributes<HTMLLIElement>;
    id: string;
    name?: string;
    type: string;
    actionLabel?: string;
}> = ({ props, id, name, type, actionLabel }) => {
    return (
        <ListItem
            {...props}
            key={id}
            role='option'
            style={{
                display: 'block',
                maxWidth: '100%',
            }}>
            <Typography variant='body2'> {actionLabel}</Typography>
            <Box style={{ display: 'flex', justifyContent: 'flex-start' }}>
                <NodeIcon nodeType={type}></NodeIcon>
                <Tooltip title={name || id} placement='top-start' TransitionComponent={Fade}>
                    <Typography
                        variant='body1'
                        style={{
                            display: 'block',
                            maxWidth: '100%',
                            whiteSpace: 'nowrap',
                            overflow: 'hidden',
                            textOverflow: 'ellipsis',
                        }}>
                        {name || id}
                    </Typography>
                </Tooltip>
            </Box>
        </ListItem>
    );
};

export default AutocompleteOption;
