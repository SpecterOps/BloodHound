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

import { useLocalStorage } from '../hooks/useLocalStorage';
import { themeCatalog } from './catalog';
import type { ThemeEffect, ThemePackage } from './types';
import { parseThemePackage } from './validate';

const installedStorageKey = 'bloodhound.installedThemes';
const effectsStorageKey = 'bloodhound.themeEffects';
const emptyPackages: ThemePackage[] = [];
const emptyEffects: Record<string, ThemeEffect[]> = {};

const validInstalledThemes = (value: unknown): ThemePackage[] => {
    if (!Array.isArray(value)) return [];

    return value.flatMap((candidate) => {
        const catalogEntry = themeCatalog.find((entry) => entry.id === candidate?.id);
        if (!catalogEntry) return [];

        try {
            return [parseThemePackage(candidate, catalogEntry.id)];
        } catch {
            return [];
        }
    });
};

export const useThemePackages = () => {
    const [storedPackages, setStoredPackages] = useLocalStorage<ThemePackage[]>(installedStorageKey, emptyPackages);
    const [storedEffects, setStoredEffects] = useLocalStorage<Record<string, ThemeEffect[]>>(
        effectsStorageKey,
        emptyEffects
    );
    const installedThemes = validInstalledThemes(storedPackages);

    const getInstalledTheme = (id: string) => installedThemes.find((theme) => theme.id === id);

    const installTheme = async (id: string): Promise<ThemePackage> => {
        const catalogEntry = themeCatalog.find((entry) => entry.id === id);
        if (!catalogEntry) throw new Error('Theme is not available in the marketplace');

        const base = import.meta.env.BASE_URL.replace(/\/$/, '');
        const packageUrl = new URL(`${base}/${catalogEntry.packagePath}`, window.location.href);
        const response = await fetch(packageUrl.href);
        if (!response.ok) throw new Error('Could not download theme');

        const theme = parseThemePackage(await response.json(), id);
        setStoredPackages((current) => [...validInstalledThemes(current).filter((entry) => entry.id !== id), theme]);
        return theme;
    };

    const uninstallTheme = (id: string) => {
        setStoredPackages((current) => validInstalledThemes(current).filter((theme) => theme.id !== id));
        setStoredEffects((current) => {
            const next = { ...current };
            delete next[id];
            return next;
        });
    };

    const enabledEffectsFor = (theme: ThemePackage): ThemeEffect[] => {
        const selected = storedEffects?.[theme.id];
        return Array.isArray(selected) ? selected.filter((effect) => theme.availableEffects?.includes(effect)) : [];
    };

    const toggleEffect = (theme: ThemePackage, effect: ThemeEffect) => {
        if (!theme.availableEffects?.includes(effect)) return;
        setStoredEffects((current) => {
            const selected = Array.isArray(current?.[theme.id]) ? current[theme.id] : [];
            const updated = selected.includes(effect)
                ? selected.filter((selectedEffect) => selectedEffect !== effect)
                : [...selected, effect];
            return { ...current, [theme.id]: updated };
        });
    };

    return { installedThemes, getInstalledTheme, installTheme, uninstallTheme, enabledEffectsFor, toggleEffect };
};
