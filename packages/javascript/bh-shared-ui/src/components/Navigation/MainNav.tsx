import { FC, useState, ReactNode } from 'react';
import { useLocation, Link as RouterLink } from 'react-router-dom';
import { useApiVersion } from '../../hooks';
//To do: Put this whole file in bh-shared-ui and here only have the data and MainNav component imported

const MainNavLogoTextImage: FC<{ MainNavLogoData: any }> = ({ MainNavLogoData }) => {
    return (
        <img
            src={MainNavLogoData.image.imageUrl}
            alt={MainNavLogoData.image.altText}
            height={MainNavLogoData.image.dimensions.height}
            width={MainNavLogoData.image.dimensions.width}
            className={MainNavLogoData.image.classes}
        />
    );
};

const MainNavListItem: FC<{ children: ReactNode; route?: string }> = ({ children, route }) => {
    const location = useLocation();
    const isActiveRoute = route ? location.pathname.includes(route.replace(/\*/g, '')) : false;

    return (
        <li
            className={`h-10 px-2 mx-2 flex items-center ${isActiveRoute ? 'text-primary bg-neutral-light-4' : 'hover:text-secondary hover:underline'} cursor-pointer rounded`}>
            {children}
        </li>
    );
};

const MainNavItemAction: FC<{ onClick: () => void; children: ReactNode; isMenuExpanded: boolean }> = ({
    onClick,
    children,
    isMenuExpanded,
}) => {
    return (
        // Note: The w-full terniary is to avoid the hover area to overflow out of the nav when its collapsed
        <button
            onClick={onClick}
            className={`h-10 ${isMenuExpanded ? 'w-full' : 'w-auto'} absolute left-5 flex items-center gap-x-2`}>
            {children}
        </button>
    );
};

const MainNavItemLink: FC<{ route: string; children: ReactNode; isMenuExpanded: boolean }> = ({
    route,
    children,
    isMenuExpanded,
}) => {
    return (
        // Note: The w-full terniary is to avoid the hover area to overflow out of the nav when its collapsed
        <RouterLink
            to={route as string}
            className={`h-10 ${isMenuExpanded ? 'w-full' : 'w-auto'} absolute left-5 flex items-center gap-x-2`}>
            {children}
        </RouterLink>
    );
};

const MainNavItemLabel: FC<{ icon: ReactNode; label: ReactNode | string; isMenuExpanded: boolean }> = ({
    icon,
    label,
    isMenuExpanded,
}) => {
    return (
        // Note: The min-h here is to keep spacing between the logo and the list below.
        <>
            <span className='flex'>{icon}</span>
            <span
                className={`whitespace-nowrap flex min-h-10 items-center gap-x-5 font-medium text-xl ${isMenuExpanded ? 'opacity-100 block' : 'opacity-0 hidden'} duration-200 ease-in`}>
                {label}
            </span>
        </>
    );
};

const MainNavVersionNumber: FC<{ isMenuExpanded: boolean }> = ({ isMenuExpanded }) => {
    const { data: apiVersionResponse, isSuccess } = useApiVersion();
    const apiVersion = isSuccess && apiVersionResponse?.server_version;

    return (
        // Note: The min-h allows for the version number to keep its position when the nav is scrollable
        <div className='relative w-full flex min-h-10 h-10 overflow-x-hidden'>
            <div
                className={`w-full flex absolute bottom-3 ${isMenuExpanded ? 'left-16' : 'left-4'} duration-300 ease-in-out text-xs whitespace-nowrap font-medium text-neutral-dark-0 dark:text-neutral-light-1`}>
                <span
                    className={`${isMenuExpanded ? 'opacity-100 block' : 'opacity-0 hidden'} duration-300 ease-in-out`}>
                    BloodHound:&nbsp;
                </span>
                <span className={`${!isMenuExpanded && 'max-w-9 overflow-x-hidden'}`}>{apiVersion}</span>
            </div>
        </div>
    );
};

const MainNav: FC<{ MainNavLogoData: any; MainNavListData: any }> = ({ MainNavLogoData, MainNavListData }) => {
    const [isMenuExpanded, setIsMenuExpanded] = useState(false);

    return (
        // Note: z-index needs to be higher than sub-nav
        <nav
            className={`z-[1201] fixed top-0 left-0 h-full ${isMenuExpanded ? 'w-72 overflow-y-auto overflow-x-hidden' : 'w-16'} duration-300 ease-in flex flex-col items-center pt-4  bg-neutral-light-2 text-neutral-dark-0 dark:bg-neutral-dark-2 dark:text-neutral-light-1 print:hidden shadow-sm`}
            onMouseEnter={() => setIsMenuExpanded(true)}
            onMouseLeave={() => setIsMenuExpanded(false)}>
            <MainNavItemLink route={MainNavLogoData.route} isMenuExpanded={isMenuExpanded}>
                <MainNavItemLabel
                    icon={MainNavLogoData.icon}
                    label={<MainNavLogoTextImage MainNavLogoData={MainNavLogoData} />}
                    isMenuExpanded={isMenuExpanded}
                />
            </MainNavItemLink>
            {/* Note: min height here is to keep the version number in bottom of nav */}
            <div className='h-full min-h-[700px] w-full flex flex-col justify-between mt-6'>
                <ul className='flex flex-col gap-6 mt-8'>
                    {MainNavListData.primary.map((listDataItem: any) => (
                        <MainNavListItem key={listDataItem.testid} route={listDataItem.route as string}>
                            <MainNavItemLink route={listDataItem.route as string} isMenuExpanded={isMenuExpanded}>
                                <MainNavItemLabel
                                    icon={listDataItem.icon}
                                    label={listDataItem.label}
                                    isMenuExpanded={isMenuExpanded}
                                />
                            </MainNavItemLink>
                        </MainNavListItem>
                    ))}
                </ul>
                <ul className='flex flex-col gap-6 mt-16'>
                    {MainNavListData.secondary.map((listDataItem: any) =>
                        listDataItem.route ? (
                            <MainNavListItem key={listDataItem.testid} route={listDataItem.route as string}>
                                <MainNavItemLink route={listDataItem.route as string} isMenuExpanded={isMenuExpanded}>
                                    <MainNavItemLabel
                                        icon={listDataItem.icon}
                                        label={listDataItem.label}
                                        isMenuExpanded={isMenuExpanded}
                                    />
                                </MainNavItemLink>
                            </MainNavListItem>
                        ) : (
                            <MainNavListItem key={listDataItem.testid}>
                                <MainNavItemAction
                                    onClick={(() => listDataItem.functionHandler as () => void)()}
                                    isMenuExpanded={isMenuExpanded}>
                                    <MainNavItemLabel
                                        icon={listDataItem.icon}
                                        label={listDataItem.label}
                                        isMenuExpanded={isMenuExpanded}
                                    />
                                </MainNavItemAction>
                            </MainNavListItem>
                        )
                    )}
                </ul>
            </div>
            <MainNavVersionNumber isMenuExpanded={isMenuExpanded} />
        </nav>
    );
};

export default MainNav;
