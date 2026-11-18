// Copyright 2026 Specter Ops, Inc.
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

import { cn } from '../utils';
import { AdornmentVariants, comboBadgeLabelClasses, ComboBadgeVariants } from './ComboBadge.styles';

interface ComboBadgeProps extends Omit<React.HTMLAttributes<HTMLDivElement>, 'aria-label'> {
    label: string | JSX.Element;
    adornment?: string | JSX.Element;
    /** An arial label is required because of the hidden adornment. Please provide context as to what this badge is conveying to a visual user  */
    ariaLabel: string;
    /** slideLeft by default */
    type?: 'slideLeft' | 'slideRight' | 'inlineSlideLeft' | 'inlineSlideRight';
    /** false by default, displays the adornment without hovering over its peer */
    displayAdornment?: boolean;
    disableAdornment?: boolean;
}

const ComboBadge = (props: ComboBadgeProps) => {
    const {
        label,
        adornment,
        ariaLabel,
        type = 'slideLeft',
        displayAdornment = false,
        disableAdornment = false,
        className,
        ...rest
    } = props;

    return (
        <div role='status' aria-label={ariaLabel} className={cn(ComboBadgeVariants({ type }), className)} {...rest}>
            <div className={cn(comboBadgeLabelClasses, { peer: !disableAdornment })}>{label}</div>
            {adornment ? <div className={cn(AdornmentVariants({ type, displayAdornment }))}>{adornment}</div> : null}
        </div>
    );
};

export { ComboBadge };
