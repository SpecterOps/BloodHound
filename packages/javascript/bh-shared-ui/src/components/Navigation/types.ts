import { ReactNode } from 'react';

export type MainNavLogoDataObject = {
    project: {
        route: string;
        icon: ReactNode;
        image: {
            imageUrl: string;
            dimensions: { height: string; width: string };
            classes?: string;
            altText: string;
        };
    };
    specterOps: {
        image: {
            imageUrl: string;
            dimensions: { height: string; width: string };
            altText: string;
        };
    };
};

export type MainNavDataItem = {
    label: string | ReactNode;
    icon: ReactNode;
    route?: string;
    functionHandler?: () => void;
};
