import { VisuallyHidden } from '@bloodhoundenterprise/doodleui';
import React from 'react';

export interface BaseSVGProps extends Omit<React.SVGProps<SVGSVGElement>, 'name'> {
    size?: number;
}

export const BaseSVG: React.FC<
    BaseSVGProps & {
        /**
         * PascalCase -> kebab-case
         */
        name: string;
    }
> = (props) => {
    const { size = 16, name, children, ...rest } = props;
    return (
        <svg {...rest} width={size} height={size}>
            {children}
            <VisuallyHidden>{`app-icon-${name}`}</VisuallyHidden>
        </svg>
    );
};

type BasePathProps = Omit<React.SVGProps<SVGPathElement>, 'fill'>;
export const BasePath: React.FC<BasePathProps> = (props) => {
    return <path {...props} fill='currentColor' />;
};
