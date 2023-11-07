import { Reducer } from '@reduxjs/toolkit';

export interface BloodHoundUIRoute {
    exact?: boolean;
    path: string;
    component: React.ComponentType;
    authenticationRequired: boolean;
}

export interface BloodHoundUIProps {
    routes?: (baseRoutes: BloodHoundUIRoute[]) => BloodHoundUIRoute[];
    components?: Record<string, React.FC>;
    reducers?: Record<string, Reducer>;
}
