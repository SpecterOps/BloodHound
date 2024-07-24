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

import { Button } from '@bloodhoundenterprise/doodleui';
import {
    Box,
    Checkbox,
    Collapse,
    FormControl,
    FormControlLabel,
    Grid,
    InputLabel,
    MenuItem,
    Paper,
    Select,
    Typography,
    useTheme,
} from '@mui/material';
import clsx from 'clsx';
import { AssetGroupMemberCounts } from 'js-client-library';
import { AssetGroupMemberParams } from 'js-client-library/dist/types';
import { FC, useState } from 'react';
import NodeIcon from '../NodeIcon';

export const FILTERABLE_PARAMS: Array<keyof Pick<AssetGroupMemberParams, 'primary_kind' | 'custom_member'>> = [
    'primary_kind',
    'custom_member',
];

interface Props {
    filterParams: AssetGroupMemberParams;
    handleFilterChange: (key: (typeof FILTERABLE_PARAMS)[number], value: string) => void;
    memberCounts: AssetGroupMemberCounts | undefined;
}

const AssetGroupFilters: FC<Props> = ({ filterParams, handleFilterChange, memberCounts = { counts: {} } }) => {
    const [displayFilters, setDisplayFilters] = useState(false);
    const theme = useTheme();

    const handleClearFilters = () => {
        for (const filter of FILTERABLE_PARAMS) {
            handleFilterChange(filter, '');
        }
    };

    const active = !!filterParams.primary_kind || !!filterParams.custom_member;

    return (
        <Box
            p={1}
            component={Paper}
            bgcolor={theme.palette.neutral.secondary}
            elevation={0}
            marginBottom={1}
            data-testid='asset-group-filters-container'>
            <Button
                onClick={() => setDisplayFilters((prev) => !prev)}
                data-testid='display-filters-button'
                className={clsx('w-full', active && 'font-bold')}>
                FILTERS
                <Typography
                    component={'span'}
                    data-testid={'active-filters-dot'}
                    style={active ? { visibility: 'visible' } : {}}
                    sx={{
                        width: '6px',
                        height: '6px',
                        borderRadius: '100%',
                        backgroundColor: 'white',
                        alignSelf: 'baseline',
                        visibility: 'hidden',
                    }}
                />
            </Button>
            <Collapse in={displayFilters} data-testid='asset-group-filter-collapsible-section' sx={{ mt: '12px' }}>
                <Grid container spacing={2}>
                    <Grid item xs={12}>
                        <FormControl sx={{ display: 'block' }}>
                            <InputLabel id='nodeTypeFilter-label'>Node Type</InputLabel>
                            <Select
                                id='nodeType'
                                labelId='nodeTypeFilter-label'
                                value={filterParams.primary_kind ?? ''}
                                onChange={(e) => handleFilterChange('primary_kind', e.target.value)}
                                variant='standard'
                                fullWidth
                                data-testid='asset-groups-node-type-filter'>
                                <MenuItem value=''>
                                    <em>None</em>
                                </MenuItem>
                                {Object.keys(memberCounts.counts).map((value) => {
                                    return (
                                        <MenuItem value={`eq:${value}`} key={value}>
                                            <NodeIcon nodeType={value} />
                                            {value}
                                        </MenuItem>
                                    );
                                })}
                            </Select>
                        </FormControl>
                    </Grid>
                    <Grid item xs={12}>
                        <FormControlLabel
                            label='Custom Members'
                            control={
                                <Checkbox
                                    checked={!!filterParams.custom_member}
                                    onChange={(e) => {
                                        handleFilterChange('custom_member', `eq:${e.target.checked}`);
                                    }}
                                    data-testid='asset-groups-custom-member-filter'
                                />
                            }
                        />
                    </Grid>
                    <Grid item xs={12} p={1}>
                        <Box sx={{ width: '100%', display: 'flex', justifyContent: 'flex-end' }}>
                            <Button onClick={handleClearFilters} disabled={!active}>
                                Clear Filters
                            </Button>
                        </Box>
                    </Grid>
                </Grid>
            </Collapse>
        </Box>
    );
};

export default AssetGroupFilters;
