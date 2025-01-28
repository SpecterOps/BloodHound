// Copyright 2025 Specter Ops, Inc.
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
import { AppIcon } from 'bh-shared-ui';
import { logout } from 'src/ducks/auth/authSlice';
import { setDarkMode } from 'src/ducks/global/actions.ts';
import * as routes from 'src/routes/constants';
import { useAppDispatch, useAppSelector } from 'src/store';

export const useMainNavLogoData = () => {
    const darkMode = useAppSelector((state) => state.global.view.darkMode);
    const bhceImageUrlDarkMode = '/img/banner-ce-dark-mode.png';
    const bhceImageUrlLightMode = '/img/banner-ce-light-mode.png';
    const soImageUrlDarkMode = '/img/banner-so-dark-mode.png';
    const soImageUrlLightMode = '/img/banner-so-light-mode.png';
    return {
        project: {
            route: routes.ROUTE_HOME,
            icon: <AppIcon.BHCELogo size={24} className='scale-150 text-[#e61616]' />, // Note: size 24 icon looked too small in comparison so had to scale it up a bit because upping the size misaligns it
            image: {
                imageUrl: `${import.meta.env.BASE_URL}${darkMode ? bhceImageUrlDarkMode : bhceImageUrlLightMode}`,
                dimensions: { height: '40px', width: '165px' },
                classes: 'ml-4 mt-2',
                altText: 'BHCE Text Logo',
            },
        },
        specterOps: {
            image: {
                imageUrl: `${import.meta.env.BASE_URL}${darkMode ? soImageUrlDarkMode : soImageUrlLightMode}`,
                dimensions: { height: '25px', width: '110px' },
                altText: 'SpecterOps Text Logo',
            },
        },
    };
};

export const MainNavPrimaryListData = [
    {
        label: 'Explore',
        icon: <AppIcon.LineChart size={24} />,
        route: routes.ROUTE_EXPLORE,
    },
    {
        label: 'Group Management',
        icon: <AppIcon.Diamond size={24} />,
        route: routes.ROUTE_GROUP_MANAGEMENT,
    },
];

export const useMainNavSecondaryListData = () => {
    const dispatch = useAppDispatch();
    const darkMode = useAppSelector((state) => state.global.view.darkMode);

    const handleLogout = () => {
        dispatch(logout());
    };

    const handleToggleDarkMode = () => {
        dispatch(setDarkMode(!darkMode));
    };

    const handleGoToSupport = () => {
        window.open('https://support.bloodhoundenterprise.io/hc/en-us', '_blank');
    };

    return [
        {
            label: 'Profile',
            icon: <AppIcon.User size={24} />,
            route: routes.ROUTE_MY_PROFILE,
        },
        {
            label: 'Docs and Support',
            icon: <AppIcon.FileMagnifyingGlass size={24} />,
            functionHandler: handleGoToSupport,
        },
        {
            label: 'Administration',
            icon: <AppIcon.UserCog size={24} />,
            route: routes.ROUTE_ADMINISTRATION_ROOT,
        },
        {
            label: 'API Explorer',
            icon: <AppIcon.Compass size={24} />,
            route: routes.ROUTE_API_EXPLORER,
        },
        {
            label: (
                <>
                    {'Dark Mode'}
                    <Switch checked={darkMode} />
                </>
            ),
            icon: <AppIcon.EclipseCircle size={24} />,
            functionHandler: handleToggleDarkMode,
        },
        {
            label: 'Log Out',
            icon: <AppIcon.Logout size={24} />,
            functionHandler: handleLogout,
        },
    ];
};
