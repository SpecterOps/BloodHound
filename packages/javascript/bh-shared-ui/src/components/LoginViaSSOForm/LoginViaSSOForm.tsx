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

import { Button } from '@bloodhoundenterprise/doodleui';
import { Box, FormControl, Grid, InputLabel, MenuItem, Select } from '@mui/material';
import { SSOProvider } from 'js-client-library';
import React from 'react';

interface LoginViaSSOFormProps {
    providers: SSOProvider[] | undefined;
    onSubmit: (providerSlug: string) => void;
    onCancel: () => void;
}

const LoginViaSSOForm: React.FC<LoginViaSSOFormProps> = ({ providers, onSubmit, onCancel }) => {
    /* Hooks */
    const [selectedProviderSlug, setSelectedProviderSlug] = React.useState('');

    /* Event Handlers */
    const handleSubmit: React.FormEventHandler<HTMLFormElement> = (e) => {
        e.preventDefault();

        if (selectedProviderSlug === null) {
            return;
        }

        onSubmit(selectedProviderSlug);
    };

    /* Implementation */
    return (
        <form onSubmit={handleSubmit}>
            <Grid container spacing={4} justifyContent='center'>
                <Grid item xs={12}>
                    <FormControl variant='outlined'>
                        <InputLabel id='selected-saml-provider-label'>Choose your SSO Provider</InputLabel>
                        <Select
                            labelId='selected-saml-provider-label'
                            id='selected-saml-provider'
                            value={selectedProviderSlug}
                            label='Choose your SSO Provider'
                            onChange={(e) => setSelectedProviderSlug(e.target.value as string)}
                            fullWidth>
                            {providers?.map((provider) => (
                                <MenuItem key={provider.id} value={provider.slug}>
                                    {provider.name}
                                </MenuItem>
                            ))}
                        </Select>
                    </FormControl>
                </Grid>
                <Grid item xs={8}>
                    <Button size='large' type='submit' className='w-full' disabled={selectedProviderSlug === ''}>
                        CONTINUE
                    </Button>
                    <Box mt={2}>
                        <Button size='large' type='button' onClick={onCancel} variant={'tertiary'} className='w-full'>
                            CANCEL
                        </Button>
                    </Box>
                </Grid>
            </Grid>
        </form>
    );
};

export default LoginViaSSOForm;
