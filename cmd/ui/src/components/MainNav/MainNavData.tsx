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
import { AppIcon, MainNavData, useFeatureFlags } from 'bh-shared-ui';
import { fullyAuthenticatedSelector, logout } from 'src/ducks/auth/authSlice';
import { setDarkMode } from 'src/ducks/global/actions.ts';
import * as routes from 'src/routes/constants';
import { useAppDispatch, useAppSelector } from 'src/store';

export const useMainNavLogoData = (): MainNavData['logo'] => {
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

export const useMainNavPrimaryListData = (): MainNavData['primaryList'] => {
    const authState = useAppSelector((state) => state.auth);
    const fullyAuthenticated = useAppSelector(fullyAuthenticatedSelector);
    const enableFeatureFlagRequests = !!authState.isInitialized && fullyAuthenticated;
    const featureFlags = useFeatureFlags({ enabled: enableFeatureFlagRequests });
    const tierFlag = featureFlags?.data?.find((flag) => {
        return flag.key === 'tier_management_engine';
    });

    const primaryList = [
        {
            label: 'Explore',
            icon: <AppIcon.LineChart size={24} />,
            route: routes.ROUTE_EXPLORE,
            testId: 'global_nav-explore',
        },
        {
            label: tierFlag?.enabled ? 'Tier Management' : 'Group Management',
            icon: <AppIcon.Diamond size={24} />,
            route: tierFlag?.enabled ? routes.ROUTE_TIER_MANAGEMENT : routes.ROUTE_GROUP_MANAGEMENT,
            testId: tierFlag?.enabled ? 'global_nav-tier-management' : 'global_nav-group-management',
        },
    ];

    return primaryList;
};

export const useMainNavSecondaryListData = (): MainNavData['secondaryList'] => {
    const dispatch = useAppDispatch();
    const darkMode = useAppSelector((state) => state.global.view.darkMode);

    const handleLogout = () => {
        dispatch(logout());
    };

    const handleToggleDarkMode = () => {
        dispatch(setDarkMode(!darkMode));
    };

    const handleGoToSupport = () => {
        window.open('https://bloodhound.specterops.io', '_blank');
    };

    return [
        {
            label: 'Profile',
            icon: <AppIcon.User size={24} />,
            route: routes.ROUTE_MY_PROFILE,
            testId: 'global_nav-my-profile',
        },
        {
            label: 'Download Collectors',
            icon: <AppIcon.Download size={24} />,
            route: routes.ROUTE_DOWNLOAD_COLLECTORS,
            testId: 'global_nav-download-collectors',
        },
        {
            label: 'Administration',
            icon: <AppIcon.UserCog size={24} />,
            route: routes.DEFAULT_ADMINISTRATION_ROUTE,
            testId: 'global_nav-administration',
        },
        {
            label: 'API Explorer',
            icon: <AppIcon.Compass size={24} />,
            route: routes.ROUTE_API_EXPLORER,
            testId: 'global_nav-api-explorer',
        },
        {
            label: 'Docs and Support',
            icon: <AppIcon.FileMagnifyingGlass size={24} />,
            functionHandler: handleGoToSupport,
            testId: 'global_nav-support',
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
            testId: 'global_nav-dark-mode',
        },
        {
            label: 'Log Out',
            icon: <AppIcon.Logout size={24} />,
            functionHandler: handleLogout,
            testId: 'global_nav-logout',
        },
    ];
};
