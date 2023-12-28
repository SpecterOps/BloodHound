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

import { faGlobe, faCloud } from '@fortawesome/free-solid-svg-icons';
import { FontAwesomeIcon } from '@fortawesome/react-fontawesome';
import {
    Alert,
    Box,
    Button,
    Divider,
    MenuItem,
    Popover,
    Skeleton,
    TextField,
    Tooltip,
    Typography,
} from '@mui/material';
import { useAvailableDomains, Domain } from '../../../hooks';
import React, { ReactNode, useState } from 'react';

const DataSelector: React.FC<{
    value: { type: string | null; id: string | null };
    errorMessage: ReactNode;
    onChange?: (newValue: { type: string; id: string | null }) => void;
    fullWidth?: boolean;
}> = ({ value, errorMessage, onChange = () => {}, fullWidth = false }) => {
    const [anchorEl, setAnchorEl] = useState(null);
    const [searchInput, setSearchInput] = useState<string>('');
    const { data, isLoading, isError } = useAvailableDomains();

    if (isLoading) return <Skeleton variant='rounded' height={36} width={256} />;

    if (isError) return <Alert severity='error'>{errorMessage}</Alert>;

    const handleClick = (event: any) => {
        setAnchorEl(event.currentTarget);
    };
    const handleClose = () => {
        setAnchorEl(null);
    };
    const open = Boolean(anchorEl);

    const filteredDomains = data.filter((domain: Domain) =>
        domain.name.toLowerCase().includes(searchInput.toLowerCase())
    );

    let selectedDomainName: string;

    if (value.type === 'active-directory-platform') {
        selectedDomainName = 'All Active Directory Domains';
    } else if (value.type === 'azure-platform') {
        selectedDomainName = 'All Azure Tenants';
    } else {
        const selectedDomain: Domain | undefined = data?.find((domain: Domain) => domain.id === value.id);
        if (selectedDomain) {
            selectedDomainName = selectedDomain.name;
        } else {
            selectedDomainName = 'Unknown Domain or Tenant';
        }
    }

    return (
        <Box data-testid='data-selector' p={1}>
            <Button
                sx={{
                    whiteSpace: 'nowrap',
                    overflow: 'hidden',
                    textOverflow: 'ellipsis',
                    display: 'block',
                }}
                fullWidth={fullWidth}
                variant='contained'
                disableElevation
                color='primary'
                onClick={handleClick}
                data-testid='data-quality_context-selector'>
                {selectedDomainName !== null ? selectedDomainName : 'Select Context'}
            </Button>
            <Popover
                open={open}
                anchorEl={anchorEl}
                onClose={handleClose}
                anchorOrigin={{
                    vertical: 'bottom',
                    horizontal: 'center',
                }}
                transformOrigin={{
                    vertical: 'top',
                    horizontal: 'center',
                }}
                data-testid='data-quality_context-selector-popover'>
                <Box display='flex' padding='0px 16px'>
                    <TextField
                        autoFocus={true}
                        value={searchInput}
                        onChange={(e) => {
                            setSearchInput(e.target.value);
                        }}
                        variant='standard'
                        fullWidth
                        label='Search'
                        data-testid={'data-quality_context-selector-search'}></TextField>
                </Box>
                {filteredDomains &&
                    filteredDomains
                        .sort((a: Domain, b: Domain) => {
                            return a.name.localeCompare(b.name);
                        })
                        .map((item: Domain) => {
                            return item.collected ? (
                                <MenuItem
                                    style={{
                                        display: 'flex',
                                        justifyContent: 'space-between',
                                        width: 450,
                                        maxWidth: 450,
                                    }}
                                    key={item.id}
                                    onClick={() => {
                                        onChange({ type: item.type, id: item.id });
                                        handleClose();
                                    }}>
                                    <Tooltip title={item.name}>
                                        <Typography
                                            style={{
                                                overflow: 'hidden',
                                                textTransform: 'uppercase',
                                                display: 'inline-block',
                                                textOverflow: 'ellipsis',
                                                maxWidth: 350,
                                            }}>
                                            {item.name}
                                        </Typography>
                                    </Tooltip>
                                    <FontAwesomeIcon
                                        style={{ width: '10%', alignSelf: 'center' }}
                                        icon={item.type === 'azure' ? faCloud : faGlobe}
                                        size='sm'
                                    />
                                </MenuItem>
                            ) : null;
                        })}
                <Divider />
                <MenuItem
                    style={{
                        display: 'flex',
                        justifyContent: 'space-between',
                    }}
                    onClick={() => {
                        onChange({ type: 'active-directory-platform', id: null });
                        handleClose();
                    }}>
                    All Active Directory Domains
                    <FontAwesomeIcon
                        style={{ width: '10%', alignSelf: 'center', marginLeft: '8px' }}
                        icon={faGlobe}
                        size='sm'
                    />
                </MenuItem>

                <MenuItem
                    style={{
                        display: 'flex',
                        justifyContent: 'space-between',
                    }}
                    onClick={() => {
                        onChange({ type: 'azure-platform', id: null });
                        handleClose();
                    }}>
                    All Azure Tenants
                    <FontAwesomeIcon
                        style={{ width: '10%', alignSelf: 'center', marginLeft: '8px' }}
                        icon={faCloud}
                        size='sm'
                    />
                </MenuItem>
            </Popover>
        </Box>
    );
};

export default DataSelector;
