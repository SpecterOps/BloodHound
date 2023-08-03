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

import { icon } from '@fortawesome/fontawesome-svg-core';
import {
    faBox,
    faBoxOpen,
    faBolt,
    faClipboardList,
    faCloud,
    faCube,
    faCubes,
    faDesktop,
    faCog,
    faGlobe,
    faKey,
    faList,
    faLock,
    faObjectGroup,
    faRobot,
    faServer,
    faSitemap,
    faUser,
    faUsers,
    faWindowRestore,
    IconDefinition,
    faGem,
    faPlus,
    faQuestion,
    faMinus,
} from '@fortawesome/free-solid-svg-icons';
import { ActiveDirectoryKind, AzureKind } from 'bh-shared-ui';

export type IconInfo = {
    icon: IconDefinition;
    color: string;
    url?: string;
};

export type IconDictionary = {
    [index: string]: IconInfo;
};

export type GlyphDictionary = {
    [index: string]: IconInfo & { iconColor: string };
};

export enum GlyphKind {
    TIER_ZERO,
    EXPAND,
    COLLAPSE,
}

const NODE_SCALE = '0.6';
const GLYPH_SCALE = '0.5';
const DEFAULT_ICON_COLOR = '#000000';

export const NODE_ICON: IconDictionary = {
    [ActiveDirectoryKind.User]: {
        icon: faUser,
        color: '#17E625',
    },

    [ActiveDirectoryKind.Group]: {
        icon: faUsers,
        color: '#DBE617',
    },

    [ActiveDirectoryKind.Computer]: {
        icon: faDesktop,
        color: '#E67873',
    },

    [ActiveDirectoryKind.Domain]: {
        icon: faGlobe,
        color: '#17E6B9',
    },

    [ActiveDirectoryKind.GPO]: {
        icon: faList,
        color: '#998EFD',
    },

    [ActiveDirectoryKind.OU]: {
        icon: faSitemap,
        color: '#FFAA00',
    },

    [ActiveDirectoryKind.Container]: {
        icon: faBox,
        color: '#F79A78',
    },

    [AzureKind.User]: {
        icon: faUser,
        color: '#34D2EB',
    },

    [AzureKind.Group]: {
        icon: faUsers,
        color: '#F57C9B',
    },

    [AzureKind.Tenant]: {
        icon: faCloud,
        color: '#54F2F2',
    },

    [AzureKind.Subscription]: {
        icon: faKey,
        color: '#D2CCA1',
    },

    [AzureKind.ResourceGroup]: {
        icon: faCube,
        color: '#89BD9E',
    },

    [AzureKind.VM]: {
        icon: faDesktop,
        color: '#F9ADA0',
    },
    [AzureKind.WebApp]: {
        icon: faObjectGroup,
        color: '#4696E9',
    },
    [AzureKind.LogicApp]: {
        icon: faSitemap,
        color: '#9EE047',
    },

    [AzureKind.AutomationAccount]: {
        icon: faCog,
        color: '#F4BA44',
    },

    [AzureKind.FunctionApp]: {
        icon: faBolt,
        color: '#F4BA44',
    },

    [AzureKind.ContainerRegistry]: {
        icon: faBoxOpen,
        color: '#0885D7',
    },

    [AzureKind.ManagedCluster]: {
        icon: faCubes,
        color: '#326CE5',
    },

    [AzureKind.Device]: {
        icon: faDesktop,
        color: '#B18FCF',
    },

    [AzureKind.KeyVault]: {
        icon: faLock,
        color: '#ED658C',
    },

    [AzureKind.App]: {
        icon: faWindowRestore,
        color: '#03FC84',
    },

    [AzureKind.VMScaleSet]: {
        icon: faServer,
        color: '#007CD0',
    },

    [AzureKind.ServicePrincipal]: {
        icon: faRobot,
        color: '#C1D6D6',
    },

    [AzureKind.Role]: {
        icon: faClipboardList,
        color: '#ED8537',
    },

    [AzureKind.ManagementGroup]: {
        icon: faSitemap,
        color: '#BD93D8',
    },
};

export const GLYPHS: GlyphDictionary = {
    [GlyphKind.TIER_ZERO]: {
        icon: faGem,
        color: '#000000',
        iconColor: '#FFFFFF',
    },
    [GlyphKind.EXPAND]: {
        icon: faPlus,
        color: '#FFFFFF',
        iconColor: '#000000',
    },
    [GlyphKind.COLLAPSE]: {
        icon: faMinus,
        color: '#FFFFFF',
        iconColor: '#000000',
    },
};

export const UNKNOWN_ICON: IconInfo = {
    icon: faQuestion,
    color: '#FFFFFF',
};

// Adds object URLs to all icon and glyph definitions so that our fontawesome icons can be used by sigmajs node programs
const appendSvgUrls = (icons: IconDictionary | GlyphDictionary, scale: string): void => {
    Object.entries(icons).forEach(([type, value]) => {
        if (value.url) return;

        const color = (icons[type] as any).iconColor || DEFAULT_ICON_COLOR;
        icons[type].url = getModifiedSvgUrlFromIcon(value.icon, scale, color);
    });
};

const getModifiedSvgUrlFromIcon = (iconDefinition: IconDefinition, scale: string, color: string): string => {
    const modifiedIcon = icon(iconDefinition, {
        styles: { 'transform-origin': 'center', scale, color },
    });

    const svgString = modifiedIcon.html[0].replace(/<svg/, '<svg width="200" height="200"');
    const svg = new Blob([svgString], { type: 'image/svg+xml' });
    return URL.createObjectURL(svg);
};

// Append URLs for nodes, glyphs, and any additional utility icons
appendSvgUrls(NODE_ICON, NODE_SCALE);
appendSvgUrls(GLYPHS, GLYPH_SCALE);
UNKNOWN_ICON.url = getModifiedSvgUrlFromIcon(UNKNOWN_ICON.icon, NODE_SCALE, DEFAULT_ICON_COLOR);
