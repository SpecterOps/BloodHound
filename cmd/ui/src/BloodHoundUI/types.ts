import { Reducer } from '@reduxjs/toolkit';
import { OVERRIDABLE_COMPONENTS } from '.';

export interface BloodHoundUIRoute {
    exact?: boolean;
    path: string;
    component: React.ComponentType;
    authenticationRequired: boolean;
}

export interface BloodHoundUIProps {
    routes?: (baseRoutes: BloodHoundUIRoute[]) => BloodHoundUIRoute[];
    components?: Record<keyof typeof OVERRIDABLE_COMPONENTS, React.ReactNode>;
    reducers?: Record<string, Reducer>;
}
