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

import { Checkbox, DialogDescription, DialogTitle } from '@bloodhoundenterprise/doodleui';
import { faSearch } from '@fortawesome/free-solid-svg-icons';
import { FontAwesomeIcon } from '@fortawesome/react-fontawesome';
import { Box, Card, TextField } from '@mui/material';
import { Environment } from 'js-client-library';
import React, { useState } from 'react';
import { useAvailableEnvironments } from '../../hooks/useAvailableEnvironments/useAvailableEnvironments';

const UserFormEnvironmentSelector: React.FC<{
    //checked: string[];
    disabled?: boolean;
    //onChange: (checked: any[]) => void;
}> = ({}) => {
    const { data: availableEnvironments, isLoading, isError } = useAvailableEnvironments();
    const [searchInput, setSearchInput] = useState<string>('');
    const [selectedItems, setSelectedItems] = useState<string[]>([]);

    const filteredEnvironments = availableEnvironments?.filter((environment: Environment) =>
        environment.name.toLowerCase().includes(searchInput.toLowerCase())
    );

    const handleSelectAllChange = (checked: any) => {
        if (checked) {
            const returnMappedEnvironments: string[] | undefined = availableEnvironments?.map((item) => item.id);
            setSelectedItems(returnMappedEnvironments || []);
        } else {
            setSelectedItems([]);
        }
    };

    const handleItemChange = (itemId: any, checked: any) => {
        if (checked) {
            setSelectedItems((prevSelected) => [...prevSelected, itemId]);
        } else {
            setSelectedItems((prevSelected) => prevSelected.filter((id) => id !== itemId));
        }
    };

    const isAllSelected = selectedItems.length === availableEnvironments?.length && availableEnvironments.length > 0;

    return (
        <Card className='flex-1 p-4 rounded shadow max-w-[400px]'>
            <DialogTitle>Environmental Access Control</DialogTitle>
            <DialogDescription
                className='flex flex-col'
                data-testid='create-user-dialog_environments-checkboxes-dialog'>
                <Box className={'ml-4 w-[90%] flex items-center uppercase'}>
                    <FontAwesomeIcon icon={faSearch} size='lg' color='inherit' />
                    <TextField
                        autoFocus
                        className={'w-full ml-3'}
                        label='Search'
                        onChange={(e) => {
                            setSearchInput(e.target.value);
                        }}
                        variant='standard'
                    />
                </Box>
                <div
                    className='flex flex-row mt-6 mb-2 items-center'
                    data-testid='create-user-dialog_environments-checkboxes-select-all'>
                    <Checkbox checked={isAllSelected} id='selectAll' onCheckedChange={handleSelectAllChange} />
                    <label htmlFor={''} className='ml-3 w-full cursor-pointer'>
                        Select All Environments
                    </label>
                </div>
                <div className='flex flex-col' data-testid='create-user-dialog_environments-checkboxes'>
                    {filteredEnvironments &&
                        filteredEnvironments?.map((item) => {
                            return (
                                <div
                                    className='flex justify-start items-center'
                                    data-testid='create-user-dialog_environments-checkbox'>
                                    <Checkbox
                                        checked={selectedItems.includes(item.id)}
                                        className='m-3'
                                        id={item.id}
                                        onCheckedChange={(checked) => handleItemChange(item.id, checked)}
                                        value={item.name}
                                    />
                                    <label
                                        htmlFor={``}
                                        className='mr-3 w-full cursor-pointer'
                                        data-testid='create-user-dialog_environments-checkbox'>
                                        {item.name}
                                    </label>
                                </div>
                            );
                        })}
                </div>
            </DialogDescription>
        </Card>
    );
};

export default UserFormEnvironmentSelector;
