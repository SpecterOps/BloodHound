// Copyright 2026 Specter Ops, Inc.
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
import { Alert, CircularProgress } from '@mui/material';
import { Button, Switch, Typography } from 'doodle-ui';
import { PageWithTitle } from '../../components';
import { useUserPreferences } from '../../providers/UserPreferencesProvider';

const UserPreferences = () => {
    const { preferences, isLoading, isReady, isSaving, loadError, saveError, updatePreferences, retry } =
        useUserPreferences();
    const settings = [
        {
            key: 'rememberDomain' as const,
            title: 'Preferred Domain',
            description: 'Remember the last domain you select and use it by default when you return.',
        },
        { key: 'darkMode' as const, title: 'Dark Mode', description: 'Use a dark appearance throughout BloodHound.' },
    ];

    return (
        <PageWithTitle
            title='User Preferenecs'
            data-testid='user-preferences'
            pageDescription={
                <Typography variant='body2'>
                    Choose your personal display and domain settings. Changes are saved automatically.
                </Typography>
            }>
            {isLoading && (
                <div className='flex items-center gap-3 py-4' role='status'>
                    <CircularProgress size={20} />
                    Loading your preferences…
                </div>
            )}
            {loadError && (
                <Alert
                    severity='error'
                    action={
                        <Button variant='secondary' onClick={retry}>
                            Retry
                        </Button>
                    }>
                    Unable to load your preferences. Your saved settings have not been changed.
                </Alert>
            )}
            {saveError && <Alert severity='error'>Your change could not be saved. Please try again.</Alert>}
            <div className='overflow-x-auto rounded border border-neutral-3 mt-4'>
                <table className='w-full text-left' aria-label='User preferences' aria-busy={isLoading || isSaving}>
                    <thead className='bg-neutral-2'>
                        <tr>
                            <th scope='col' className='px-6 py-4 font-semibold'>
                                Setting
                            </th>
                            <th scope='col' className='px-6 py-4 font-semibold'>
                                Description
                            </th>
                            <th scope='col' className='px-6 py-4 font-semibold'>
                                Status
                            </th>
                        </tr>
                    </thead>
                    <tbody>
                        {settings.map(({ key, title, description }) => (
                            <tr key={key} className='border-t border-neutral-3'>
                                <th scope='row' className='px-6 py-5 font-medium whitespace-nowrap'>
                                    {title}
                                </th>
                                <td className='px-6 py-5' id={`${key}-description`}>
                                    {description}
                                </td>
                                <td className='px-6 py-5'>
                                    <Switch
                                        id={`user-preferences-${key}`}
                                        aria-label={title}
                                        aria-describedby={`${key}-description`}
                                        checked={preferences[key]}
                                        label={preferences[key] ? 'Enabled' : 'Disabled'}
                                        disabled={!isReady || isSaving}
                                        onCheckedChange={(checked) => {
                                            void updatePreferences({ [key]: checked }).catch(() => {});
                                        }}
                                    />
                                </td>
                            </tr>
                        ))}
                    </tbody>
                </table>
            </div>
            <div role='status' aria-live='polite' className='mt-3 min-h-6 text-sm'>
                {isSaving ? 'Saving preferences…' : ''}
            </div>
        </PageWithTitle>
    );
};

export default UserPreferences;
