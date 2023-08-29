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

import { Box, Button, FormControl, Grid, InputLabel, MenuItem, Select } from '@mui/material';
import React from 'react';

interface LoginViaSAMLFormProps {
    providers: { name: string; initiation_url: string }[];
    onSubmit: (redirectURL: string) => void;
    onCancel: () => void;
}

const LoginViaSAMLForm: React.FC<LoginViaSAMLFormProps> = ({ providers, onSubmit, onCancel }) => {
    /* Hooks */
    const [redirectURL, setRedirectURL] = React.useState('');

    /* Event Handlers */
    const handleSubmit: React.FormEventHandler<HTMLFormElement> = (e) => {
        e.preventDefault();

        if (redirectURL === null) {
            return;
        }

        onSubmit(redirectURL);
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
                            value={redirectURL}
                            label='Choose your SSO Provider'
                            onChange={(e) => setRedirectURL(e.target.value as string)}
                            fullWidth>
                            {providers.map((provider) => (
                                <MenuItem key={provider.name} value={provider.initiation_url}>
                                    {provider.name}
                                </MenuItem>
                            ))}
                        </Select>
                    </FormControl>
                </Grid>
                <Grid item xs={8}>
                    <Button
                        variant='contained'
                        color='primary'
                        size='large'
                        type='submit'
                        fullWidth
                        disableElevation
                        disabled={redirectURL === ''}>
                        CONTINUE
                    </Button>
                    <Box mt={2}>
                        <Button
                            color='inherit'
                            onClick={onCancel}
                            variant='contained'
                            size='large'
                            type='button'
                            fullWidth
                            disableElevation>
                            CANCEL
                        </Button>
                    </Box>
                </Grid>
            </Grid>
        </form>
    );
};

export default LoginViaSAMLForm;
