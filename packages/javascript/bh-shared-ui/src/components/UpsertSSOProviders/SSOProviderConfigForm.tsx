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

import { Switch } from '@bloodhoundenterprise/doodleui';
import {
    Alert,
    Grid,
    Typography,
    FormControl,
    InputLabel,
    Select,
    MenuItem,
    FormControlLabel,
    useTheme,
} from '@mui/material';
import { FC } from 'react';
import { Control, Controller, FieldErrors, UseFormResetField, UseFormWatch } from 'react-hook-form';
import { Role, UpsertOIDCProviderRequest, UpsertSAMLProviderFormInputs } from 'js-client-library';

export const backfillSSOProviderConfig = (readOnlyRoleId?: number) => ({
    auto_provision: { enabled: false, default_role: readOnlyRoleId, role_provision: false },
});

const SSOProviderConfigForm: FC<{
    control: Control<UpsertSAMLProviderFormInputs | UpsertOIDCProviderRequest, any>;
    errors: FieldErrors<UpsertSAMLProviderFormInputs | UpsertOIDCProviderRequest>;
    watch: UseFormWatch<UpsertOIDCProviderRequest | UpsertOIDCProviderRequest>;
    resetField: UseFormResetField<UpsertSAMLProviderFormInputs | UpsertOIDCProviderRequest>;
    roles?: Role[];
    readOnlyRoleId?: number;
}> = ({ control, errors, readOnlyRoleId, resetField, roles, watch }) => {
    const theme = useTheme();

    return (
        <>
            <Grid item xs={12} sx={{ m: theme.spacing(2) }}>
                <Controller
                    name='config.auto_provision.enabled'
                    control={control}
                    defaultValue={false}
                    render={({ field }) => (
                        <FormControlLabel
                            control={
                                <Switch
                                    checked={field.value}
                                    onCheckedChange={(checked) => {
                                        field.onChange(checked);
                                        if (!checked) {
                                            resetField('config.auto_provision.role_provision');
                                            resetField('config.auto_provision.default_role_id');
                                        }
                                    }}
                                    color='primary'
                                    data-testid='sso-provider-config-form_toggle-auto-provision'
                                />
                            }
                            label={
                                <Typography sx={{ ml: theme.spacing(2) }}>
                                    Automatically create new users on login
                                </Typography>
                            }
                        />
                    )}
                />
            </Grid>
            <Grid item xs={12} sx={{ mx: theme.spacing(2), mb: theme.spacing(2) }}>
                <Controller
                    name='config.auto_provision.role_provision'
                    control={control}
                    defaultValue={false}
                    render={({ field }) => (
                        <FormControlLabel
                            disabled={!watch('config.auto_provision.enabled')}
                            control={
                                <Switch
                                    checked={field.value}
                                    onCheckedChange={(checked) => field.onChange(checked)}
                                    color='primary'
                                    data-testid='sso-provider-config-form_toggle-role-provision'
                                />
                            }
                            label={
                                <Typography sx={{ ml: theme.spacing(2) }}>
                                    Allow SSO Provider to modify roles
                                </Typography>
                            }
                        />
                    )}
                />
            </Grid>
            <Grid item xs={3}>
                <Controller
                    name='config.auto_provision.default_role_id'
                    control={control}
                    defaultValue={readOnlyRoleId}
                    rules={{
                        required: 'Default role is required',
                        validate: (value) => value != 0 || 'Default role is required',
                    }}
                    render={({ field }) => (
                        <FormControl>
                            <InputLabel
                                id='role-label'
                                sx={{ ml: '-14px', mt: '8px' }}
                                disabled={!watch('config.auto_provision.enabled')}>
                                Default User Role
                            </InputLabel>
                            <Select
                                disabled={!watch('config.auto_provision.enabled')}
                                labelId='role-label'
                                id='role'
                                name='role'
                                onChange={(e) => {
                                    const output = parseInt(e.target.value as string, 10);
                                    field.onChange(isNaN(output) ? 1 : output);
                                }}
                                value={isNaN(field.value) ? '' : field.value.toString()}
                                variant='standard'
                                fullWidth
                                data-testid='sso-provider-config-form_select-default-role'>
                                {roles?.map((role: Role) => (
                                    <MenuItem key={role.id} value={role.id.toString()}>
                                        {role.name}
                                    </MenuItem>
                                ))}
                            </Select>
                        </FormControl>
                    )}
                />
            </Grid>
            {!!errors.config?.auto_provision?.default_role_id && (
                <Grid item xs={5}>
                    <Alert severity='error'>{errors.config?.auto_provision?.default_role_id?.message}</Alert>
                </Grid>
            )}
        </>
    );
};

export default SSOProviderConfigForm;
