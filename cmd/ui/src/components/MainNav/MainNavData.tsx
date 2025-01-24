import { Switch } from '@bloodhoundenterprise/doodleui';
import { AppIcon } from 'bh-shared-ui';
import { ReactNode } from 'react';
import { logout } from 'src/ducks/auth/authSlice';
import { setDarkMode } from 'src/ducks/global/actions.ts';
import * as routes from 'src/routes/constants';
import { useAppDispatch, useAppSelector } from 'src/store';

export const useMainNavLogoData = () => {
    const darkMode = useAppSelector((state) => state.global.view.darkMode);
    const imageUrlDarkMode = '/img/banner-ce-dark-mode.png';
    const imageUrlLightMode = '/img/banner-ce-light-mode.png';
    return {
        route: routes.ROUTE_HOME,
        icon: <AppIcon.BHCELogo size={24} className='scale-150 text-[#e61616]' />, // Note: size 24 icon looked too small in comparison so had to scale it up a bit because upping the size misaligns it
        image: {
            imageUrl: `${import.meta.env.BASE_URL}${darkMode ? imageUrlDarkMode : imageUrlLightMode}`,
            dimensions: { height: '40px', width: '165px' },
            classes: 'ml-4',
            altText: 'BHCE Text Logo',
        },
    };
};

export type NavListDataItem = {
    label: string | ReactNode;
    icon: ReactNode;
    route?: string;
    functionHandler?: () => void;
};

export const MainNavPrimaryListData: NavListDataItem[] = [
    {
        label: 'Explore',
        icon: <AppIcon.LineChart size={24} />,
        route: routes.ROUTE_EXPLORE,
    },
    {
        label: 'Tier Management',
        icon: <AppIcon.Diamond size={24} />,
        route: routes.ROUTE_GROUP_MANAGEMENT,
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
