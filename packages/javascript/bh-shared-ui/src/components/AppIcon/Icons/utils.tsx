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
