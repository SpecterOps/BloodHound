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
import {
    Alert,
    Box,
    DialogActions,
    DialogContent,
    FormHelperText,
    Grid,
    TextField,
    Typography,
    useTheme,
} from '@mui/material';
import { useState, FC } from 'react';
import { Controller, useForm } from 'react-hook-form';
import { SSOProvider, UpsertSAMLProviderFormInputs } from 'js-client-library';

const UpsertSAMLProviderForm: FC<{
    error?: string;
    oldSSOProvider?: SSOProvider;
    onClose: () => void;
    onSubmit: (data: UpsertSAMLProviderFormInputs) => void;
}> = ({ error, onClose, oldSSOProvider, onSubmit }) => {
    const theme = useTheme();
    const {
        control,
        handleSubmit,
        reset,
        formState: { errors },
    } = useForm<UpsertSAMLProviderFormInputs>({
        defaultValues: {
            name: oldSSOProvider?.name ?? '',
            metadata: undefined,
        },
    });
    const [fileValue, setFileValue] = useState(''); // small workaround to use the file input

    const handleClose = () => {
        onClose();
        setFileValue('');
        reset();
    };

    return (
        <form autoComplete='off' onSubmit={handleSubmit(onSubmit)}>
            <DialogContent>
                <Grid container spacing={2}>
                    <Grid item xs={12}>
                        <Controller
                            control={control}
                            name='name'
                            rules={{
                                required: 'SAML Provider Name is required',
                                pattern: {
                                    value: /^[a-z0-9]+(?:-[a-z0-9]+)*$/,
                                    message:
                                        'SAML Provider Name must be a valid URL slug (e.g., "saml-provider", "test-idp-01", "any-old-slug")',
                                },
                            }}
                            render={({ field }) => (
                                <TextField
                                    {...field}
                                    id={'name'}
                                    variant='standard'
                                    fullWidth
                                    name='name'
                                    label='SAML Provider Name'
                                    error={!!errors.name}
                                    helperText={
                                        errors.name?.message || 'Choose a name for your SAML Provider configuration'
                                    }
                                />
                            )}
                        />
                    </Grid>
                    <Grid item xs={12}>
                        <Controller
                            control={control}
                            name='metadata'
                            rules={{ required: !oldSSOProvider && 'Metadata is required' }}
                            render={({ field }) => (
                                <Box p={1} borderRadius={4} bgcolor={theme.palette.neutral.tertiary}>
                                    <Box display='flex' flexDirection='row' alignItems='center'>
                                        <Button variant='secondary'>
                                            <label htmlFor='saml-provider-input'>Choose File</label>
                                            <input
                                                id='saml-provider-input'
                                                hidden
                                                type='file'
                                                accept='.xml'
                                                value={fileValue}
                                                onChange={(e) => {
                                                    setFileValue(e.target.value);
                                                    field.onChange(e.target.files as FileList);
                                                }}
                                                onBlur={field.onBlur}
                                            />
                                        </Button>
                                        <Box ml={1}>
                                            <Typography variant='body1'>
                                                {field.value?.[0] ? field.value[0].name : 'No file chosen'}
                                            </Typography>
                                        </Box>
                                    </Box>
                                </Box>
                            )}
                        />
                        <FormHelperText error={!!errors.metadata}>
                            {errors.metadata
                                ? errors.metadata.message
                                : 'Upload the Metadata file provided by your SAML Provider'}
                        </FormHelperText>
                    </Grid>
                    {error && (
                        <Grid item xs={12}>
                            <Alert severity='error'>{error}</Alert>
                        </Grid>
                    )}
                </Grid>
            </DialogContent>
            <DialogActions>
                <Button
                    type='button'
                    variant='tertiary'
                    onClick={handleClose}
                    data-testid='create-saml-provider-dialog_button-close'>
                    Cancel
                </Button>
                <Button data-testid='create-saml-provider-dialog_button-save' type='submit'>
                    {oldSSOProvider ? 'Confirm Edits' : 'Submit'}
                </Button>
            </DialogActions>
        </form>
    );
};

export default UpsertSAMLProviderForm;
