import { Switch } from '@bloodhoundenterprise/doodleui';
import { ReactNode } from 'react';
import { logout } from 'src/ducks/auth/authSlice';
import { setDarkMode } from 'src/ducks/global/actions.ts';
import * as routes from 'src/ducks/global/routes';
import { useAppDispatch, useAppSelector } from 'src/store';
import { AppIcon } from '../AppIcon';

export const useMainNavLogoData = () => {
    const darkMode = useAppSelector((state) => state.global.view.darkMode);
    const imageUrlDarkMode = '/img/logo-secondary-transparent-banner.svg';
    const imageUrlLightMode = '/img/logo-transparent-banner.svg';
    return {
        route: routes.ROUTE_HOME,
        icon: <AppIcon.BHCELogo size={24} className='scale-150 text-[#e61616]' />, // Note: size 24 icon looked too small in comparison so had to scale it up a bit because upping the size misaligns it
        image: {
            imageUrl: `${import.meta.env.BASE_URL}${darkMode ? imageUrlDarkMode : imageUrlLightMode}`,
            dimensions: { height: '40px', width: '165px' },
            classes: 'ml-4',
            altText: 'BHE Text Logo',
        },
    };
};

export type NavListDataItem = {
    label: string | ReactNode;
    icon: ReactNode;
    route?: string;
    functionHandler?: () => void;
    testid: string;
};

export const MainNavPrimaryListData: NavListDataItem[] = [
    {
        label: 'Explore',
        icon: <AppIcon.LineChart size={24} />,
        route: routes.ROUTE_EXPLORE,
        testid: 'global_header_nav-explore',
    },
    {
        label: 'Tier Management',
        icon: <AppIcon.Diamond size={24} />,
        route: routes.ROUTE_GROUP_MANAGEMENT,
        testid: 'global_header_nav-group-management',
    },
];

export const useMainNavSecondaryListData = (): NavListDataItem[] => {
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
            testid: 'global_header_nav-profile',
        },
        {
            label: 'Docs and Support',
            icon: <AppIcon.FileMagnifyingGlass size={24} />,
            functionHandler: handleGoToSupport,
            testid: 'global_header_nav-support',
        },
        {
            label: 'Administration',
            icon: <AppIcon.UserCog size={24} />,
            route: routes.ROUTE_ADMINISTRATION_ROOT,
            testid: 'global_header_nav-administration',
        },
        {
            label: 'API Explorer',
            icon: <AppIcon.Compass size={24} />,
            route: routes.ROUTE_API_EXPLORER,
            testid: 'global_header_nav-api-explorer',
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
            testid: 'global_header_nav-dark-mode',
        },
        {
            label: 'Log Out',
            icon: <AppIcon.Logout size={24} />,
            functionHandler: handleLogout,
            testid: 'global_header_nav-logout',
        },
    ];
};
