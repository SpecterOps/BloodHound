import { ComponentType, ReactNode, createContext } from 'react';

export interface BloodHoundUIRoute {
    exact?: boolean;
    path: string;
    component: ComponentType;
    authenticationRequired: boolean;
}

type BloodHoundUIContextValue = {
    routes: (baseRoutes: BloodHoundUIRoute[]) => BloodHoundUIRoute[];
};

export const BloodHoundUIContext = createContext<BloodHoundUIContextValue | undefined>(undefined);

export const BloodHoundUIContextProvider = ({
    children,
    value,
}: {
    children: ReactNode;
    value: BloodHoundUIContextValue;
}) => {
    return <BloodHoundUIContext.Provider value={value}>{children}</BloodHoundUIContext.Provider>;
};
