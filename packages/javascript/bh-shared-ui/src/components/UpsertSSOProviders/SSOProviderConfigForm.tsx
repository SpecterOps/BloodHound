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

import { Alert, FormControlLabel, Grid, Typography } from '@mui/material';
import clsx from 'clsx';
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue, Switch } from 'doodle-ui';
import {
    Role,
    SSOProviderConfiguration,
    UpsertOIDCProviderRequest,
    UpsertSAMLProviderFormInputs,
} from 'js-client-library';
import { FC } from 'react';
import { Control, Controller, FieldErrors, UseFormResetField, UseFormWatch } from 'react-hook-form';

export const maybeBackfillSSOProviderConfig = (
    readOnlyRoleId?: number,
    oldSSOProviderConfig?: SSOProviderConfiguration['config']
) => ({
    auto_provision: oldSSOProviderConfig?.auto_provision.enabled
        ? oldSSOProviderConfig.auto_provision
        : { enabled: false, default_role: readOnlyRoleId, role_provision: false },
});

const SSOProviderConfigForm: FC<{
    control: Control<UpsertSAMLProviderFormInputs | UpsertOIDCProviderRequest, any>;
    errors: FieldErrors<UpsertSAMLProviderFormInputs | UpsertOIDCProviderRequest>;
    watch: UseFormWatch<UpsertOIDCProviderRequest | UpsertOIDCProviderRequest>;
    resetField: UseFormResetField<UpsertSAMLProviderFormInputs | UpsertOIDCProviderRequest>;
    roles?: Role[];
    readOnlyRoleId?: number;
}> = ({ control, errors, readOnlyRoleId, resetField, roles, watch }) => (
    <>
        <Grid item xs={12} className='mx-4 mb-4'>
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
                                data-testid='sso-provider-config-form_toggle-auto-provision'
                            />
                        }
                        label={
                            <Typography
                                className={clsx(
                                    `ml-4 ${!watch('config.auto_provision.enabled') && 'dark:text-white dark:text-opacity-75 text-black text-opacity-75'}`
                                )}>
                                Automatically create new users on login
                            </Typography>
                        }
                    />
                )}
            />
        </Grid>
        <Grid item xs={12} className='mx-4 mb-4'>
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
                                data-testid='sso-provider-config-form_toggle-role-provision'
                            />
                        }
                        label={
                            <Typography
                                className={clsx(
                                    `ml-4 ${!watch('config.auto_provision.role_provision') && 'dark:text-white dark:text-opacity-75 text-black text-opacity-75'}`
                                )}>
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
                    <>
                        <label
                            htmlFor='role'
                            className={clsx('text-sm', !watch('config.auto_provision.enabled') && 'opacity-50')}>
                            Default User Role
                        </label>
                        <Select
                            value={isNaN(field.value) ? undefined : field.value.toString()}
                            onValueChange={(value) => {
                                const output = parseInt(value, 10);
                                field.onChange(isNaN(output) ? 1 : output);
                            }}
                            disabled={!watch('config.auto_provision.enabled')}>
                            <SelectTrigger
                                id='role'
                                variant='underlined'
                                className='w-full mt-1'
                                data-testid='sso-provider-config-form_select-default-role'>
                                <SelectValue />
                            </SelectTrigger>
                            <SelectContent>
                                {roles?.map((role: Role) => (
                                    <SelectItem key={role.id} value={role.id.toString()}>
                                        {role.name}
                                    </SelectItem>
                                ))}
                            </SelectContent>
                        </Select>
                    </>
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

export default SSOProviderConfigForm;
