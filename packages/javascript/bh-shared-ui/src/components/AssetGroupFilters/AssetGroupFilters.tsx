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
    Grid,
    InputLabel,
    MenuItem,
    Paper,
    Select,
} from '@mui/material';
import createStyles from '@mui/styles/createStyles';
import makeStyles from '@mui/styles/makeStyles';
import { Theme } from '@mui/material/styles';

const useStyles = makeStyles((theme: Theme) =>
    createStyles({
        formControl: {
            display: 'block',
        },
        active: {
            '& button': {
                fontWeight: 'bolder',

                '& span': {
                    visibility: 'visible',
                },
            },
        },
        activeFilters: {
            width: '6px',
            height: '6px',
            borderRadius: '100%',
            backgroundColor: theme.palette.primary.main,
            alignSelf: 'baseline',
            visibility: 'hidden',
        },
    })
);

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

    const classes = useStyles();

    const active = !!filterParams.primary_kind || !!filterParams.custom_member;
    const activeStyles = active ? classes.active : '';

    return (
        <Box p={1} className={activeStyles} component={Paper} elevation={0} marginBottom={1}>
            <Button fullWidth onClick={() => setDisplayFilters((prev) => !prev)}>
                Filters
                <span className={classes.activeFilters} />
            </Button>
            <Collapse in={displayFilters}>
                <Grid container spacing={2}>
                    <Grid item xs={12} xl={6}>
                        <FormControl className={classes.formControl}>
                            <InputLabel id='nodeTypeFilter-label'>Node Type</InputLabel>
                            <Select
                                id='nodeType'
                                labelId='nodeTypeFilter-label'
                                value={filterParams.primary_kind ?? ''}
                                onChange={(e) => handleFilterChange('primary_kind', e.target.value)}
                                label='Node Type'
                                variant='standard'
                                fullWidth>
                                <MenuItem value=''>
                                    <em>None</em>
                                </MenuItem>
                                {availableNodeKinds.map((value) => {
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
                    <Grid item xs={12} xl={6}>
                        <FormControlLabel
                            label='Custom Members'
                            control={
                                <Checkbox
                                    value={filterParams.custom_member}
                                    onChange={(e) => {
                                        handleFilterChange('custom_member', `eq:${e.target.checked}`);
                                    }}
                                />
                            }
                        />
                    </Grid>
                </Grid>
            </Collapse>
        </Box>
    );
};

export default AssetGroupFilters;
