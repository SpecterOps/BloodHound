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

import { Box, Grid } from '@mui/material';
import { Button, Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from 'doodle-ui';
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
                    <label htmlFor='selected-saml-provider' className='text-sm'>
                        Choose your SSO Provider
                    </label>
                    <Select value={selectedProviderSlug || undefined} onValueChange={setSelectedProviderSlug}>
                        <SelectTrigger id='selected-saml-provider' className='w-full mt-1'>
                            <SelectValue placeholder='Choose your SSO Provider' />
                        </SelectTrigger>
                        <SelectContent>
                            {providers?.map((provider) => (
                                <SelectItem key={provider.id} value={provider.slug}>
                                    {provider.name}
                                </SelectItem>
                            ))}
                        </SelectContent>
                    </Select>
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
