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

import { Alert, Button, Checkbox, FormControlLabel } from '@mui/material';
import { useState } from 'react';
import { defaultThemePreset } from '../constants';
import { themeCatalog } from './catalog';
import type { ThemeEffect } from './types';
import { useThemePackages } from './useThemePackages';

type Props = {
    activeThemeId: string;
    onActivate: (id: string) => void;
    search?: string;
};

const effectLabels: Record<ThemeEffect, string> = {
    clippy: 'Desktop assistant',
    'retro-cursor': 'Classic cursor',
};

export const ThemeMarketplaceSection = ({ activeThemeId, onActivate, search = '' }: Props) => {
    const { getInstalledTheme, installTheme, uninstallTheme, enabledEffectsFor, toggleEffect } = useThemePackages();
    const [pendingId, setPendingId] = useState<string>();
    const [error, setError] = useState<string>();
    const normalizedSearch = search.trim().toLowerCase();
    const visibleThemes = themeCatalog.filter((entry) =>
        `${entry.name} ${entry.author} ${entry.description}`.toLowerCase().includes(normalizedSearch)
    );
    const showDefault = 'bloodhound default the original bloodhound appearance'.includes(normalizedSearch);

    const handleInstall = async (id: string) => {
        setPendingId(id);
        setError(undefined);
        try {
            await installTheme(id);
        } catch (error) {
            setError(error instanceof Error ? error.message : 'Could not install theme');
        } finally {
            setPendingId(undefined);
        }
    };

    const handleUninstall = (id: string) => {
        if (activeThemeId === id) onActivate(defaultThemePreset);
        uninstallTheme(id);
    };

    if (!showDefault && visibleThemes.length === 0) return null;

    return (
        <section aria-label='Themes' className='mt-8 border-t border-neutral-4 pt-8'>
            <h2 className='mb-2 text-2xl font-bold'>Themes</h2>
            <p className='mb-5 text-text-light'>Download and install themes for this browser.</p>
            {error && (
                <Alert className='mb-4' severity='error' role='alert'>
                    {error}
                </Alert>
            )}
            <div className='grid grid-cols-1 gap-5 md:grid-cols-2 xl:grid-cols-3'>
                {showDefault && (
                    <div className='flex flex-col gap-3 rounded-lg border border-neutral-4 bg-neutral-2 p-5'>
                        <h3 className='text-lg font-bold'>BloodHound Default</h3>
                        <p className='grow'>The original BloodHound appearance.</p>
                        <Button
                            disabled={activeThemeId === defaultThemePreset}
                            onClick={() => onActivate(defaultThemePreset)}
                            variant='outlined'>
                            {activeThemeId === defaultThemePreset ? 'Active' : 'Apply'}
                        </Button>
                    </div>
                )}
                {visibleThemes.map((entry) => {
                    const installed = getInstalledTheme(entry.id);
                    const active = activeThemeId === entry.id && !!installed;
                    const enabledEffects = installed ? enabledEffectsFor(installed) : [];

                    return (
                        <div
                            className='flex flex-col gap-3 rounded-lg border border-neutral-4 bg-neutral-2 p-5'
                            key={entry.id}>
                            <div className='flex gap-1' aria-hidden='true'>
                                {entry.preview.map((color) => (
                                    <span
                                        className='h-7 flex-1 border border-neutral-5'
                                        key={color}
                                        style={{ backgroundColor: color }}
                                    />
                                ))}
                            </div>
                            <h3 className='text-lg font-bold'>{entry.name}</h3>
                            <p className='text-sm text-text-light'>{entry.author}</p>
                            <p className='grow'>{entry.description}</p>
                            <div className='flex flex-wrap gap-2'>
                                {installed ? (
                                    <>
                                        <Button
                                            disabled={active}
                                            onClick={() => onActivate(entry.id)}
                                            variant='contained'>
                                            {active ? 'Active' : 'Apply'}
                                        </Button>
                                        <Button onClick={() => handleUninstall(entry.id)} variant='outlined'>
                                            Uninstall
                                        </Button>
                                    </>
                                ) : (
                                    <Button
                                        disabled={pendingId === entry.id}
                                        onClick={() => handleInstall(entry.id)}
                                        variant='contained'>
                                        {pendingId === entry.id ? 'Installing…' : 'Install'}
                                    </Button>
                                )}
                            </div>
                            {active &&
                                installed?.availableEffects?.map((effect) => (
                                    <FormControlLabel
                                        control={
                                            <Checkbox
                                                checked={enabledEffects.includes(effect)}
                                                onChange={() => toggleEffect(installed, effect)}
                                                size='small'
                                            />
                                        }
                                        key={effect}
                                        label={effectLabels[effect]}
                                    />
                                ))}
                        </div>
                    );
                })}
            </div>
        </section>
    );
};
