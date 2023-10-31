import { createContext } from 'react';
import { BloodHoundUIProps } from './types';

export const BloodHoundUIContext = createContext<BloodHoundUIProps | undefined>(undefined);

export const BloodHoundUIContextProvider = ({
    children,
    value,
}: {
    children: React.ReactNode;
    value: BloodHoundUIProps;
}) => {
    return <BloodHoundUIContext.Provider value={value}>{children}</BloodHoundUIContext.Provider>;
};
