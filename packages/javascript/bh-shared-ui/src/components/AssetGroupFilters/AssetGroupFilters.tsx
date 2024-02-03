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

import { AssetGroupMemberParams } from 'js-client-library/dist/types';
import { FC, useState } from 'react';
import { AzureNodeKind, ActiveDirectoryNodeKind, NodeIcon } from '../..';
import {
    Box,
    Button,
    Checkbox,
    Collapse,
    FormControl,
    FormControlLabel,
    InputLabel,
    MenuItem,
    Paper,
    Select,
} from '@mui/material';

interface Props {
    filterParams: AssetGroupMemberParams;
    handleFilterChange: (
        key: keyof Pick<AssetGroupMemberParams, 'primary_kind' | 'custom_member'>,
        value: string
    ) => void;
    availableNodeKinds: Array<ActiveDirectoryNodeKind | AzureNodeKind>;
}

const AssetGroupFilters: FC<Props> = (props) => {
    const { filterParams, handleFilterChange, availableNodeKinds } = props;

    const [displayFilters, setDisplayFilters] = useState(false);

    return (
        <Box sx={{ p: '10px' }} component={Paper} elevation={0} marginBottom={1}>
            <Button onClick={() => setDisplayFilters((prev) => !prev)}>Filters</Button>
            <Collapse in={displayFilters}>
                <FormControl sx={{ display: 'flex', flexDirection: 'row' }}>
                    <InputLabel id='testwa'>Node Type</InputLabel>
                    <Select
                        labelId='testwa'
                        value={filterParams.primary_kind}
                        onChange={(e) => handleFilterChange('primary_kind', e.target.value)}
                        sx={{ minWidth: 120 }}
                        label='Node Type'>
                        <MenuItem value=''>
                            <em>None</em>
                        </MenuItem>
                        {availableNodeKinds.map((value) => {
                            return (
                                <MenuItem value={`eq:${value}`}>
                                    <NodeIcon nodeType={value} />
                                    {value}
                                </MenuItem>
                            );
                        })}
                    </Select>
                    <FormControlLabel
                        label='Custom Members'
                        control={
                            <Checkbox
                                value={filterParams.custom_member}
                                onChange={(e) => {
                                    console.log(e.target.checked);
                                    handleFilterChange('custom_member', `eq:${e.target.checked}`);
                                }}
                            />
                        }
                    />
                </FormControl>
            </Collapse>
        </Box>
    );
};

export default AssetGroupFilters;
