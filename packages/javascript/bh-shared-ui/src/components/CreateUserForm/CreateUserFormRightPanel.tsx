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

import { DialogDescription, DialogTitle, VisuallyHidden } from '@bloodhoundenterprise/doodleui';
import { faSearch } from '@fortawesome/free-solid-svg-icons';
import { FontAwesomeIcon } from '@fortawesome/react-fontawesome';
import { Box, Card, Checkbox, FormControlLabel, TextField } from '@mui/material';
import { CreateUserRequest } from 'js-client-library';
import React from 'react';

export type CreateUserRequestForm = Omit<CreateUserRequest, 'SSOProviderId'> & { SSOProviderId: string | undefined };

const CreateUserFormRightPanel: React.FC<{
    checked?: any[];
    disabled?: boolean;
    onChange?: (checked: any[]) => void;
}> = ({ checked, onChange }) => {
    const environments = [
        { name: 'Environment 1', id: 'environment1' },
        { name: 'Environment 2', id: 'environment2' },
        { name: 'Environment 3', id: 'environment3' },
    ];

    let amountChecked = 'none';

    // If all boxes are checked, they are all unchecked; other wise all boxes are checked
    /*
    const toggleAllChecked = () => {
        if (environments) {
            onChange(['none', 'some'].includes(amountChecked) ? environments.map((item) => item.id) : []);
        }
    };
    */

    /*
    const toggleCheckedEnvironment = (id: string) => () => {
        const newChecked = checked.includes(id) ? checked.filter((item) => item !== id) : [...checked, id];
        onChange(newChecked);
    };
    */

    return (
        <Card className='flex-1 p-4 rounded shadow max-w-[400px]'>
            <DialogTitle>Environmental Access Control</DialogTitle>
            <VisuallyHidden>
                something that we want to hide visually but still want in the DOM for accessibility
            </VisuallyHidden>
            <DialogDescription className='flex flex-col' data-testid='environments-checkboxes'>
                <Box className={'ml-4 w-[90%] flex items-center uppercase'}>
                    <FontAwesomeIcon icon={faSearch} size='lg' color='inherit' />
                    <TextField
                        autoFocus
                        //onChange={handleEnvironmentSearch}
                        className={'w-full'}
                        variant='standard'
                        label='Search'
                    />
                </Box>
                <div className='flex flex-col' data-testid='source-kinds-checkboxes'>
                    <FormControlLabel
                        label='Select All Environments'
                        control={
                            <Checkbox
                                checked={amountChecked === 'all'}
                                //disabled={isDisabled}
                                indeterminate={amountChecked === 'some'}
                                name='All Environments Data'
                                //onChange={toggleAllChecked}
                            />
                        }
                    />

                    {/*isLoading && LOADING_CHECKBOXES*/}

                    {environments.map((item) => (
                        <FormControlLabel
                            className='pl-8'
                            control={
                                <Checkbox
                                    //checked={checked.includes(item.id)}
                                    //onChange={toggleCheckedEnvironment(item.id)}
                                    name={item.name}
                                    //disabled={isDisabled}
                                />
                            }
                            label={item.name}
                            key={item.id}
                        />
                    ))}
                </div>
            </DialogDescription>
        </Card>
    );
};

export default CreateUserFormRightPanel;
